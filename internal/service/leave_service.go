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

const (
	fullDayLeaveMinutes int32 = 480
)

type LeaveService struct {
	repository          domain.LeaveRepository
	employeeRepo        domain.EmployeeRepository
	notificationService domain.NotificationService
	logger              domain.Logger
}

func NewLeaveService(
	repository domain.LeaveRepository,
	employeeRepo domain.EmployeeRepository,
	notificationService domain.NotificationService,
	logger domain.Logger,
) domain.LeaveService {
	return &LeaveService{
		repository:          repository,
		employeeRepo:        employeeRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (s *LeaveService) CreateLeaveRequest(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.EmployeeID = actorEmployeeID
	params.CreatedByEmployeeID = actorEmployeeID
	req, err := s.createLeaveRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// Trigger notification to administrators
	if s.notificationService != nil {
		employeeName := "An employee"
		if s.employeeRepo != nil {
			emp, err := s.employeeRepo.GetEmployeeByID(ctx, actorEmployeeID)
			if err == nil && emp != nil {
				employeeName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		reason := ""
		if req.Reason != nil {
			reason = *req.Reason
		}

		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				Roles: []string{"admin"},
			},
			Message: fmt.Sprintf("%s has requested leave (%s) from %s to %s.", employeeName, req.LeaveType, req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
			Data: domain.LeaveRequestCreatedNotificationData{
				LeaveRequestID:   req.ID,
				EmployeeID:       req.EmployeeID,
				EmployeeName:     employeeName,
				LeaveType:        req.LeaveType,
				StartDate:        req.StartDate,
				EndDate:          req.EndDate,
				RequestedMinutes: req.RequestedMinutes,
				Reason:           reason,
			},
		})
	}

	return req, nil
}

func (s *LeaveService) CreateLeaveRequestByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.CreateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	if adminEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.CreatedByEmployeeID = adminEmployeeID
	return s.createLeaveRequest(ctx, params)
}

func (s *LeaveService) createLeaveRequest(
	ctx context.Context,
	params domain.CreateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	if params.EmployeeID == uuid.Nil || params.CreatedByEmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	leaveType := strings.TrimSpace(params.LeaveType)
	if !isValidLeaveType(leaveType) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.LeaveType = leaveType

	durationType := strings.TrimSpace(params.DurationType)
	if !isValidLeaveDurationType(durationType) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	params.DurationType = durationType

	if params.StartDate.IsZero() || params.EndDate.IsZero() {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.EndDate.Before(params.StartDate) {
		return nil, fmt.Errorf(
			"%w: end date must be on or after start date",
			domain.ErrLeaveRequestInvalidRequest,
		)
	}

	policy, err := s.repository.GetActiveLeavePolicyByType(ctx, leaveType)
	if err != nil {
		return nil, err
	}
	if policy.DeductsBalance && params.StartDate.Year() != params.EndDate.Year() {
		return nil, fmt.Errorf(
			"%w: leave date range must be within one year for deductible leave types",
			domain.ErrLeaveRequestInvalidRequest,
		)
	}

	requestedMinutes, err := s.calculateRequestedMinutes(ctx, params.EmployeeID, durationType, params.StartDate, params.EndDate, params.StartTime, params.EndTime)
	if err != nil {
		return nil, err
	}
	params.RequestedMinutes = requestedMinutes

	item, err := s.repository.CreateLeaveRequest(ctx, params)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *LeaveService) UpdateLeaveRequest(
	ctx context.Context,
	actorEmployeeID, leaveRequestID uuid.UUID,
	params domain.UpdateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
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
			return fmt.Errorf(
				"%w: leave date range must be within one year for deductible leave types",
				domain.ErrLeaveRequestInvalidRequest,
			)
		}

		requestedMinutes, err := s.calculateRequestedMinutesForUpdate(ctx, tx, current.EmployeeID, next)
		if err != nil {
			return err
		}
		next.updateParams.RequestedMinutes = &requestedMinutes

		updated, err = tx.UpdateLeaveRequestEditableFields(ctx, leaveRequestID, next.updateParams)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *LeaveService) UpdateLeaveRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, leaveRequestID uuid.UUID,
	params domain.UpdateLeaveRequestParams,
	adminUpdateNote string,
) (*domain.LeaveRequest, error) {
	if adminEmployeeID == uuid.Nil || leaveRequestID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if strings.TrimSpace(adminUpdateNote) == "" {
		return nil, fmt.Errorf(
			"%w: admin_update_note is required",
			domain.ErrLeaveRequestInvalidRequest,
		)
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
			return fmt.Errorf(
				"%w: leave date range must be within one year for deductible leave types",
				domain.ErrLeaveRequestInvalidRequest,
			)
		}

		requestedMinutes, err := s.calculateRequestedMinutesForUpdate(ctx, tx, current.EmployeeID, next)
		if err != nil {
			return err
		}
		next.updateParams.RequestedMinutes = &requestedMinutes

		updated, err = tx.UpdateLeaveRequestEditableFields(ctx, leaveRequestID, next.updateParams)
		return err
	})
	if err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.LogInfo(
			ctx,
			"LeaveService.UpdateLeaveRequestByAdmin",
			"admin updated leave request",
			zap.String("leave_request_id", leaveRequestID.String()),
			zap.String("admin_employee_id", adminEmployeeID.String()),
		)
	}

	return updated, nil
}

func (s *LeaveService) DecideLeaveRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, leaveRequestID uuid.UUID,
	params domain.DecideLeaveRequestParams,
) (*domain.LeaveRequest, error) {
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
					return fmt.Errorf(
						"%w: leave date range must be within one year",
						domain.ErrLeaveRequestInvalidRequest,
					)
				}

				requestedMinutes := current.RequestedMinutes
				if requestedMinutes <= 0 {
					return fmt.Errorf(
						"%w: invalid leave duration",
						domain.ErrLeaveRequestInvalidRequest,
					)
				}

				year := int32(start.Year())
				if err := tx.LockEmployeeForLeaveBalance(ctx, current.EmployeeID); err != nil {
					return err
				}
				legalCalculatedMinutes, err := tx.ComputeLegalLeaveTotalForYear(ctx, current.EmployeeID, year, time.Now().UTC())
				if err != nil {
					return err
				}
				legalUsedMinutes, err := tx.ComputeLegalLeaveUsedForYear(ctx, current.EmployeeID, year)
				if err != nil {
					return err
				}

				if legalCalculatedMinutes-legalUsedMinutes < requestedMinutes {
					return domain.ErrLeaveBalanceInsufficient
				}
			}
		}

		updated, err = tx.UpdateLeaveRequestDecision(
			ctx,
			leaveRequestID,
			nextStatus,
			params.DecisionNote,
			adminEmployeeID,
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Trigger notification to the employee who made the request
	if s.notificationService != nil && updated.EmployeeID != uuid.Nil {
		decidedByName := "An administrator"
		if s.employeeRepo != nil {
			emp, err := s.employeeRepo.GetEmployeeByID(ctx, adminEmployeeID)
			if err == nil && emp != nil {
				decidedByName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		decisionNote := ""
		if updated.DecisionNote != nil {
			reason := strings.TrimSpace(*updated.DecisionNote)
			if reason != "" {
				decisionNote = reason
			}
		}

		var message string
		if decisionNote != "" {
			message = fmt.Sprintf("Your leave request for %s has been %s by %s. Note: %s", updated.LeaveType, updated.Status, decidedByName, decisionNote)
		} else {
			message = fmt.Sprintf("Your leave request for %s has been %s by %s.", updated.LeaveType, updated.Status, decidedByName)
		}

		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				EmployeeIDs: []uuid.UUID{updated.EmployeeID},
			},
			Message: message,
			Data: domain.LeaveRequestDecidedNotificationData{
				LeaveRequestID:      updated.ID,
				EmployeeID:          updated.EmployeeID,
				Status:              updated.Status,
				LeaveType:           updated.LeaveType,
				StartDate:           updated.StartDate,
				EndDate:             updated.EndDate,
				DecidedByEmployeeID: adminEmployeeID,
				DecidedByName:       decidedByName,
				DecisionNote:        decisionNote,
			},
		})
	}

	return updated, nil
}

func (s *LeaveService) ListMyLeaveRequests(
	ctx context.Context,
	params domain.ListMyLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.Status != nil && !isValidLeaveStatus(*params.Status) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListMyLeaveRequests(ctx, params)
}

func (s *LeaveService) ListLeaveRequests(
	ctx context.Context,
	params domain.ListLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	if params.Status != nil && !isValidLeaveStatus(*params.Status) {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListLeaveRequests(ctx, params)
}

func (s *LeaveService) ListLeaveCalendar(
	ctx context.Context,
	params domain.ListLeaveCalendarParams,
) ([]domain.LeaveCalendarEmployee, error) {
	if params.Month.IsZero() {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	params.Month = time.Date(
		params.Month.UTC().Year(),
		params.Month.UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if len(params.LeaveTypes) > 0 {
		normalized := make([]string, 0, len(params.LeaveTypes))
		for _, leaveType := range params.LeaveTypes {
			trimmed := strings.TrimSpace(leaveType)
			if !isValidLeaveType(trimmed) {
				return nil, domain.ErrLeaveRequestInvalidRequest
			}
			normalized = append(normalized, trimmed)
		}
		params.LeaveTypes = normalized
	}

	return s.repository.ListLeaveCalendar(ctx, params)
}

func (s *LeaveService) GetMyLeaveRequestStats(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.LeaveRequestStats, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.GetMyLeaveRequestStats(ctx, employeeID)
}

func (s *LeaveService) GetLeaveRequestStats(
	ctx context.Context,
) (*domain.LeaveRequestStats, error) {
	return s.repository.GetLeaveRequestStats(ctx)
}

func (s *LeaveService) ListLeaveBalances(
	ctx context.Context,
	params domain.ListLeaveBalancesParams,
) (*domain.LeaveBalancePage, error) {
	return s.repository.ListLeaveBalances(ctx, params)
}

func (s *LeaveService) ListMyLeaveBalances(
	ctx context.Context,
	params domain.ListMyLeaveBalancesParams,
) (*domain.LeaveBalancePage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	return s.repository.ListMyLeaveBalances(ctx, params)
}

func (s *LeaveService) GetLeaveBalanceDetails(
	ctx context.Context,
	params domain.GetLeaveBalanceDetailsParams,
) (*domain.LeaveBalanceDetails, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if params.Year == 0 {
		params.Year = int32(time.Now().UTC().Year())
	}
	if params.Year < 2000 || params.Year > 2100 {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	details, err := s.repository.GetLeaveBalanceDetails(ctx, params)
	if err != nil {
		return nil, err
	}
	if details != nil {
		allocateLeaveUsageToContractAccruals(
			details.ContractAccruals,
			details.Balance.LegalUsedMinutes,
		)
	}
	return details, nil
}

// allocateLeaveUsageToContractAccruals splits the employee's annual used
// leave across contract segments, prioritising the oldest segment first.
//
// The slice is expected to be ordered from oldest to newest segment. The
// function mutates the slice in place, setting DeductedMinutes and
// RemainingMinutes on each entry. The total of deducted minutes never
// exceeds legalUsedMinutes, and per segment it never exceeds GainedMinutes.
func allocateLeaveUsageToContractAccruals(
	accruals []domain.LeaveContractAccrual,
	legalUsedMinutes int32,
) {
	if len(accruals) == 0 {
		return
	}
	remaining := legalUsedMinutes
	if remaining < 0 {
		remaining = 0
	}
	for i := range accruals {
		if remaining <= 0 {
			accruals[i].DeductedMinutes = 0
			accruals[i].RemainingMinutes = accruals[i].GainedMinutes
			continue
		}
		deducted := accruals[i].GainedMinutes
		if remaining < deducted {
			deducted = remaining
		}
		accruals[i].DeductedMinutes = deducted
		accruals[i].RemainingMinutes = accruals[i].GainedMinutes - deducted
		remaining -= deducted
	}
}

type normalizedUpdateParams struct {
	updateParams          domain.UpdateLeaveRequestParams
	effectiveLeaveType    string
	effectiveDurationType string
	effectiveStartTime    *time.Time
	effectiveEndTime      *time.Time
	finalStartDate        time.Time
	finalEndDate          time.Time
	effectiveReason       *string
}

func normalizeUpdateParams(
	current domain.LeaveRequest,
	update domain.UpdateLeaveRequestParams,
	enforceFutureStart bool,
) (*normalizedUpdateParams, error) {
	var hasUpdates bool
	out := &normalizedUpdateParams{
		effectiveLeaveType:    current.LeaveType,
		effectiveDurationType: current.DurationType,
		effectiveStartTime:    current.StartTime,
		effectiveEndTime:      current.EndTime,
		finalStartDate:        dateOnlyUTC(current.StartDate),
		finalEndDate:          dateOnlyUTC(current.EndDate),
		effectiveReason:       current.Reason,
	}

	if update.LeaveType != nil {
		trimmed := strings.TrimSpace(*update.LeaveType)
		if !isValidLeaveType(trimmed) {
			return nil, domain.ErrLeaveRequestInvalidRequest
		}
		out.effectiveLeaveType = trimmed
		hasUpdates = true
	}
	if update.DurationType != nil {
		trimmed := strings.TrimSpace(*update.DurationType)
		if !isValidLeaveDurationType(trimmed) {
			return nil, domain.ErrLeaveRequestInvalidRequest
		}
		out.effectiveDurationType = trimmed
		hasUpdates = true
	}
	if update.StartDate != nil {
		d := dateOnlyUTC(*update.StartDate)
		out.finalStartDate = d
		hasUpdates = true
	}
	if update.EndDate != nil {
		d := dateOnlyUTC(*update.EndDate)
		out.finalEndDate = d
		hasUpdates = true
	}
	if update.DurationType != nil {
		out.effectiveStartTime = update.StartTime
		out.effectiveEndTime = update.EndTime
	} else {
		if update.StartTime != nil {
			out.effectiveStartTime = update.StartTime
			hasUpdates = true
		}
		if update.EndTime != nil {
			out.effectiveEndTime = update.EndTime
			hasUpdates = true
		}
	}
	if update.Reason != nil {
		trimmed := strings.TrimSpace(*update.Reason)
		out.effectiveReason = &trimmed
		hasUpdates = true
	}
	if !hasUpdates {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}
	if out.finalEndDate.Before(out.finalStartDate) {
		return nil, fmt.Errorf(
			"%w: end date must be on or after start date",
			domain.ErrLeaveRequestInvalidRequest,
		)
	}
	if enforceFutureStart && !out.finalStartDate.After(currentUTCDate()) {
		return nil, domain.ErrLeaveRequestStateInvalid
	}

	out.updateParams.LeaveType = &out.effectiveLeaveType
	out.updateParams.DurationType = &out.effectiveDurationType
	out.updateParams.StartDate = &out.finalStartDate
	out.updateParams.EndDate = &out.finalEndDate
	out.updateParams.StartTime = out.effectiveStartTime
	out.updateParams.EndTime = out.effectiveEndTime
	out.updateParams.Reason = out.effectiveReason

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

func isValidLeaveDurationType(value string) bool {
	switch strings.TrimSpace(value) {
	case "full_day", "hours":
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

// [code_quality]: these need to be refactored into a shared utility package
func currentUTCDate() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func dateOnlyUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

type leaveContractLookup func(context.Context, uuid.UUID, time.Time) (*domain.LeaveContractAtDate, error)

func (s *LeaveService) calculateRequestedMinutes(
	ctx context.Context,
	employeeID uuid.UUID,
	durationType string,
	startDate, endDate time.Time,
	startTime, endTime *time.Time,
) (int32, error) {
	if durationType == "full_day" && (startTime != nil || endTime != nil) {
		return 0, fmt.Errorf(
			"%w: start_time and end_time must not be set for full-day leave",
			domain.ErrLeaveDurationInvalid,
		)
	}
	if durationType == "hours" && (startTime == nil || endTime == nil) {
		return 0, fmt.Errorf(
			"%w: start_time and end_time are required for hourly leave",
			domain.ErrLeaveDurationInvalid,
		)
	}
	lookup := s.repository.GetEmployeeContractAtDate
	switch durationType {
	case "full_day":
		return calculateFullDayMinutes(lookup, ctx, employeeID, startDate, endDate)
	case "hours":
		return calculateHoursMinutes(lookup, ctx, employeeID, startDate, endDate, startTime, endTime)
	default:
		return 0, domain.ErrLeaveDurationInvalid
	}
}

func (s *LeaveService) calculateRequestedMinutesForUpdate(
	ctx context.Context,
	tx domain.LeaveTxRepository,
	employeeID uuid.UUID,
	next *normalizedUpdateParams,
) (int32, error) {
	lookup := tx.GetEmployeeContractAtDate
	switch next.effectiveDurationType {
	case "full_day":
		if next.effectiveStartTime != nil || next.effectiveEndTime != nil {
			return 0, fmt.Errorf(
				"%w: start_time and end_time must not be set for full-day leave",
				domain.ErrLeaveDurationInvalid,
			)
		}
		return calculateFullDayMinutes(lookup, ctx, employeeID, next.finalStartDate, next.finalEndDate)
	case "hours":
		if next.effectiveStartTime == nil || next.effectiveEndTime == nil {
			return 0, fmt.Errorf(
				"%w: start_time and end_time are required for hourly leave",
				domain.ErrLeaveDurationInvalid,
			)
		}
		return calculateHoursMinutes(lookup, ctx, employeeID, next.finalStartDate, next.finalEndDate, next.effectiveStartTime, next.effectiveEndTime)
	default:
		return 0, domain.ErrLeaveDurationInvalid
	}
}

func calculateFullDayMinutes(
	lookup leaveContractLookup,
	ctx context.Context,
	employeeID uuid.UUID,
	startDate, endDate time.Time,
) (int32, error) {
	current := dateOnlyUTC(startDate)
	end := dateOnlyUTC(endDate)

	if end.Before(current) {
		return 0, fmt.Errorf("%w: end date must be on or after start date", domain.ErrLeaveDurationInvalid)
	}

	var totalMinutes int32
	for !current.After(end) {
		contract, err := lookup(ctx, employeeID, current)
		if err != nil {
			return 0, fmt.Errorf("%w: no active contract for date %s: %w", domain.ErrLeaveDurationInvalid, current.Format("2006-01-02"), err)
		}

		if contract.RosterFreeDay != "" && strings.EqualFold(current.Weekday().String(), contract.RosterFreeDay) {
			current = current.AddDate(0, 0, 1)
			continue
		}

		totalMinutes += fullDayLeaveMinutes
		current = current.AddDate(0, 0, 1)
	}

	if totalMinutes <= 0 {
		return 0, fmt.Errorf("%w: no chargeable days in the requested range", domain.ErrLeaveDurationInvalid)
	}

	return totalMinutes, nil
}

func calculateHoursMinutes(
	lookup leaveContractLookup,
	ctx context.Context,
	employeeID uuid.UUID,
	startDate, endDate time.Time,
	startTime, endTime *time.Time,
) (int32, error) {
	start := dateOnlyUTC(startDate)
	end := dateOnlyUTC(endDate)

	if !start.Equal(end) {
		return 0, fmt.Errorf("%w: hourly leave must be on a single date", domain.ErrLeaveDurationInvalid)
	}

	contract, err := lookup(ctx, employeeID, start)
	if err != nil {
		return 0, fmt.Errorf("%w: no active contract for date %s: %w", domain.ErrLeaveDurationInvalid, start.Format("2006-01-02"), err)
	}

	if contract.RosterFreeDay != "" && strings.EqualFold(start.Weekday().String(), contract.RosterFreeDay) {
		return 0, fmt.Errorf("%w: hourly leave is not allowed on roster-free day", domain.ErrLeaveDurationInvalid)
	}

	if startTime == nil || endTime == nil {
		return 0, fmt.Errorf("%w: start_time and end_time are required for hourly leave", domain.ErrLeaveDurationInvalid)
	}

	duration := endTime.Sub(*startTime)
	if duration <= 0 {
		return 0, fmt.Errorf("%w: end_time must be after start_time", domain.ErrLeaveDurationInvalid)
	}

	requestedMinutes := int32(duration.Minutes())

	if requestedMinutes > fullDayLeaveMinutes {
		return 0, fmt.Errorf("%w: hourly leave cannot exceed %d minutes", domain.ErrLeaveDurationInvalid, fullDayLeaveMinutes)
	}

	return requestedMinutes, nil
}

var _ domain.LeaveService = (*LeaveService)(nil)
