package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LeaveService struct {
	repository domain.LeaveRepository
	logger     domain.Logger
}

func NewLeaveService(repository domain.LeaveRepository, logger domain.Logger) domain.LeaveService {
	return &LeaveService{
		repository: repository,
		logger:     logger,
	}
}

func (s *LeaveService) CreateLeaveRequest(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CreateLeaveRequestParams) (*domain.LeaveRequest, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.EmployeeID = actorEmployeeID
	params.CreatedByEmployeeID = actorEmployeeID
	return s.createLeaveRequest(ctx, params)
}

func (s *LeaveService) CreateLeaveRequestByAdmin(ctx context.Context, adminEmployeeID uuid.UUID, params domain.CreateLeaveRequestParams) (*domain.LeaveRequest, error) {
	if adminEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.CreatedByEmployeeID = adminEmployeeID
	return s.createLeaveRequest(ctx, params)
}

func (s *LeaveService) createLeaveRequest(ctx context.Context, params domain.CreateLeaveRequestParams) (*domain.LeaveRequest, error) {
	if params.EmployeeID == uuid.Nil || params.CreatedByEmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	leaveType := strings.TrimSpace(params.LeaveType)
	if !isValidLeaveType(leaveType) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.LeaveType = leaveType

	if params.StartDate.IsZero() || params.EndDate.IsZero() {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.EndDate.Before(params.StartDate) {
		return nil, fmt.Errorf("%w: end date must be on or after start date", domain.ErrLeaveRequestInvalidRequest)
	}

	policy, err := s.repository.GetActiveLeavePolicyByType(ctx, leaveType)
	if err != nil {
		return nil, err
	}
	if policy.DeductsBalance && params.StartDate.Year() != params.EndDate.Year() {
		return nil, fmt.Errorf("%w: leave date range must be within one year for deductible leave types", domain.ErrLeaveRequestInvalidRequest)
	}

	item, err := s.repository.CreateLeaveRequest(ctx, params)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *LeaveService) UpdateLeaveRequest(ctx context.Context, actorEmployeeID, leaveRequestID uuid.UUID, params domain.UpdateLeaveRequestParams) (*domain.LeaveRequest, error) {
	if actorEmployeeID == uuid.Nil || leaveRequestID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	var updated *domain.LeaveRequest
	err := s.repository.WithTx(ctx, func(tx domain.LeaveTxRepository) error {
		current, err := tx.GetLeaveRequestForUpdate(ctx, leaveRequestID)
		if err != nil {
			return err
		}
		if current.EmployeeID != actorEmployeeID {
			return domain.ErrLeaveRequestForbidden
		}
		if current.Status != "pending" {
			return domain.ErrLeaveRequestStateInvalid
		}
		if !dateOnlyUTC(current.StartDate).After(currentUTCDate()) {
			return domain.ErrLeaveRequestStateInvalid
		}

		next, err := normalizeUpdateParams(*current, params, true)
		if err != nil {
			return err
		}
		policy, err := tx.GetActiveLeavePolicyByType(ctx, next.effectiveLeaveType)
		if err != nil {
			return err
		}
		if policy.DeductsBalance && next.finalStartDate.Year() != next.finalEndDate.Year() {
			return fmt.Errorf("%w: leave date range must be within one year for deductible leave types", domain.ErrLeaveRequestInvalidRequest)
		}

		updated, err = tx.UpdateLeaveRequestEditableFields(ctx, leaveRequestID, next.updateParams)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *LeaveService) UpdateLeaveRequestByAdmin(ctx context.Context, adminEmployeeID, leaveRequestID uuid.UUID, params domain.UpdateLeaveRequestParams, adminUpdateNote string) (*domain.LeaveRequest, error) {
	if adminEmployeeID == uuid.Nil || leaveRequestID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if strings.TrimSpace(adminUpdateNote) == "" {
		return nil, fmt.Errorf("%w: admin_update_note is required", domain.ErrLeaveRequestInvalidRequest)
	}

	var updated *domain.LeaveRequest
	err := s.repository.WithTx(ctx, func(tx domain.LeaveTxRepository) error {
		current, err := tx.GetLeaveRequestForUpdate(ctx, leaveRequestID)
		if err != nil {
			return err
		}
		if current.Status != "pending" && current.Status != "rejected" {
			return domain.ErrLeaveRequestStateInvalid
		}

		next, err := normalizeUpdateParams(*current, params, false)
		if err != nil {
			return err
		}
		policy, err := tx.GetActiveLeavePolicyByType(ctx, next.effectiveLeaveType)
		if err != nil {
			return err
		}
		if policy.DeductsBalance && next.finalStartDate.Year() != next.finalEndDate.Year() {
			return fmt.Errorf("%w: leave date range must be within one year for deductible leave types", domain.ErrLeaveRequestInvalidRequest)
		}

		updated, err = tx.UpdateLeaveRequestEditableFields(ctx, leaveRequestID, next.updateParams)
		return err
	})
	if err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.LogInfo(ctx, "LeaveService.UpdateLeaveRequestByAdmin", "admin updated leave request",
			zap.String("leave_request_id", leaveRequestID.String()),
			zap.String("admin_employee_id", adminEmployeeID.String()),
		)
	}

	return updated, nil
}

func (s *LeaveService) DecideLeaveRequestByAdmin(ctx context.Context, adminEmployeeID, leaveRequestID uuid.UUID, params domain.DecideLeaveRequestParams) (*domain.LeaveRequest, error) {
	if adminEmployeeID == uuid.Nil || leaveRequestID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	decision := strings.TrimSpace(params.Decision)
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	var updated *domain.LeaveRequest
	err := s.repository.WithTx(ctx, func(tx domain.LeaveTxRepository) error {
		current, err := tx.GetLeaveRequestForUpdate(ctx, leaveRequestID)
		if err != nil {
			return err
		}
		if current.Status != "pending" {
			return domain.ErrLeaveRequestStateInvalid
		}

		nextStatus := "rejected"
		if decision == "approve" {
			nextStatus = "approved"
			policy, err := tx.GetActiveLeavePolicyByType(ctx, current.LeaveType)
			if err != nil {
				return err
			}
			if policy.DeductsBalance {
				start := dateOnlyUTC(current.StartDate)
				end := dateOnlyUTC(current.EndDate)
				if start.Year() != end.Year() {
					return fmt.Errorf("%w: leave date range must be within one year", domain.ErrLeaveRequestInvalidRequest)
				}

				requestedDays := int32(end.Sub(start).Hours()/24) + 1
				if requestedDays <= 0 {
					return fmt.Errorf("%w: invalid leave duration", domain.ErrLeaveRequestInvalidRequest)
				}
				hoursPerDay, err := tx.GetLeaveHoursPerDay(ctx, current.EmployeeID)
				if err != nil {
					return err
				}
				if hoursPerDay <= 0 {
					return fmt.Errorf("%w: invalid employee day-hour configuration", domain.ErrLeaveRequestInvalidRequest)
				}
				requestedHours := requestedDays * hoursPerDay

				year := int32(start.Year())
				if err := tx.EnsureLeaveBalanceForYear(ctx, current.EmployeeID, year); err != nil {
					return err
				}

				balance, err := tx.GetLeaveBalanceForUpdate(ctx, current.EmployeeID, year)
				if err != nil {
					return err
				}

				if balance.TotalRemaining < requestedHours {
					return domain.ErrLeaveBalanceInsufficient
				}

				extraToUse := minInt32(balance.ExtraRemaining, requestedHours)
				legalToUse := requestedHours - extraToUse
				if _, err := tx.ApplyLeaveBalanceDeduction(ctx, balance.ID, extraToUse, legalToUse); err != nil {
					return err
				}
			}
		}

		updated, err = tx.UpdateLeaveRequestDecision(ctx, leaveRequestID, nextStatus, params.DecisionNote, adminEmployeeID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *LeaveService) ListMyLeaveRequests(ctx context.Context, params domain.ListMyLeaveRequestsParams) (*domain.LeaveRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.Status != nil && !isValidLeaveStatus(*params.Status) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListMyLeaveRequests(ctx, params)
}

func (s *LeaveService) ListLeaveRequests(ctx context.Context, params domain.ListLeaveRequestsParams) (*domain.LeaveRequestPage, error) {
	if params.Status != nil && !isValidLeaveStatus(*params.Status) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListLeaveRequests(ctx, params)
}

func (s *LeaveService) GetMyLeaveRequestStats(ctx context.Context, employeeID uuid.UUID) (*domain.LeaveRequestStats, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.GetMyLeaveRequestStats(ctx, employeeID)
}

func (s *LeaveService) GetLeaveRequestStats(ctx context.Context) (*domain.LeaveRequestStats, error) {
	return s.repository.GetLeaveRequestStats(ctx)
}

func (s *LeaveService) ListLeaveBalances(ctx context.Context, params domain.ListLeaveBalancesParams) (*domain.LeaveBalancePage, error) {
	return s.repository.ListLeaveBalances(ctx, params)
}

func (s *LeaveService) ListMyLeaveBalances(ctx context.Context, params domain.ListMyLeaveBalancesParams) (*domain.LeaveBalancePage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListMyLeaveBalances(ctx, params)
}

func (s *LeaveService) AdjustLeaveBalance(ctx context.Context, params domain.AdjustLeaveBalanceParams) (*domain.LeaveBalance, error) {
	if params.AdminEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.LegalHoursDelta == 0 && params.ExtraHoursDelta == 0 {
		return nil, fmt.Errorf("%w: at least one delta is required", domain.ErrLeaveBalanceInvalidAdjust)
	}
	params.Reason = strings.TrimSpace(params.Reason)
	if params.Reason == "" {
		return nil, fmt.Errorf("%w: reason is required", domain.ErrLeaveBalanceInvalidAdjust)
	}

	var adjusted *domain.LeaveBalance
	err := s.repository.WithTx(ctx, func(tx domain.LeaveTxRepository) error {
		if err := tx.EnsureLeaveBalanceForYear(ctx, params.EmployeeID, params.Year); err != nil {
			return err
		}
		current, err := tx.GetLeaveBalanceForUpdate(ctx, params.EmployeeID, params.Year)
		if err != nil {
			return err
		}

		nextLegalTotal := current.LegalTotalHours + params.LegalHoursDelta
		nextExtraTotal := current.ExtraTotalHours + params.ExtraHoursDelta
		if nextLegalTotal < 0 || nextExtraTotal < 0 {
			return fmt.Errorf("%w: totals cannot be negative", domain.ErrLeaveBalanceInvalidAdjust)
		}
		if nextLegalTotal < current.LegalUsedHours || nextExtraTotal < current.ExtraUsedHours {
			return fmt.Errorf("%w: totals cannot be lower than already used hours", domain.ErrLeaveBalanceInvalidAdjust)
		}

		adjusted, err = tx.ApplyLeaveBalanceTotalAdjustment(ctx, current.ID, params.LegalHoursDelta, params.ExtraHoursDelta)
		if err != nil {
			return err
		}

		return tx.CreateLeaveBalanceAdjustmentAudit(ctx, domain.CreateLeaveBalanceAdjustmentAuditParams{
			LeaveBalanceID:        current.ID,
			EmployeeID:            params.EmployeeID,
			Year:                  params.Year,
			LegalHoursDelta:       params.LegalHoursDelta,
			ExtraHoursDelta:       params.ExtraHoursDelta,
			Reason:                params.Reason,
			AdjustedByEmployeeID:  params.AdminEmployeeID,
			LegalTotalHoursBefore: current.LegalTotalHours,
			ExtraTotalHoursBefore: current.ExtraTotalHours,
			LegalTotalHoursAfter:  adjusted.LegalTotalHours,
			ExtraTotalHoursAfter:  adjusted.ExtraTotalHours,
		})
	})
	if err != nil {
		return nil, err
	}
	return adjusted, nil
}

type normalizedUpdateParams struct {
	updateParams       domain.UpdateLeaveRequestParams
	effectiveLeaveType string
	finalStartDate     time.Time
	finalEndDate       time.Time
}

func normalizeUpdateParams(current domain.LeaveRequest, update domain.UpdateLeaveRequestParams, enforceFutureStart bool) (*normalizedUpdateParams, error) {
	var hasUpdates bool
	out := &normalizedUpdateParams{
		updateParams:       domain.UpdateLeaveRequestParams{},
		effectiveLeaveType: current.LeaveType,
		finalStartDate:     dateOnlyUTC(current.StartDate),
		finalEndDate:       dateOnlyUTC(current.EndDate),
	}

	if update.LeaveType != nil {
		trimmed := strings.TrimSpace(*update.LeaveType)
		if !isValidLeaveType(trimmed) {
			return nil, domain.ErrLeaveRequestInvalidRequest
		}
		out.updateParams.LeaveType = &trimmed
		out.effectiveLeaveType = trimmed
		hasUpdates = true
	}
	if update.StartDate != nil {
		d := dateOnlyUTC(*update.StartDate)
		out.updateParams.StartDate = &d
		out.finalStartDate = d
		hasUpdates = true
	}
	if update.EndDate != nil {
		d := dateOnlyUTC(*update.EndDate)
		out.updateParams.EndDate = &d
		out.finalEndDate = d
		hasUpdates = true
	}
	if update.Reason != nil {
		trimmed := strings.TrimSpace(*update.Reason)
		out.updateParams.Reason = &trimmed
		hasUpdates = true
	}
	if !hasUpdates {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if out.finalEndDate.Before(out.finalStartDate) {
		return nil, fmt.Errorf("%w: end date must be on or after start date", domain.ErrLeaveRequestInvalidRequest)
	}
	if enforceFutureStart && !out.finalStartDate.After(currentUTCDate()) {
		return nil, domain.ErrLeaveRequestStateInvalid
	}
	return out, nil
}

func isValidLeaveType(value string) bool {
	switch strings.TrimSpace(value) {
	case "vacation", "personal", "sick", "pregnancy", "unpaid", "other":
		return true
	default:
		return false
	}
}

func isValidLeaveStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "pending", "approved", "rejected", "cancelled", "expired":
		return true
	default:
		return false
	}
}

func currentUTCDate() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func dateOnlyUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

var _ domain.LeaveService = (*LeaveService)(nil)
