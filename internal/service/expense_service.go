package service

import (
	"context"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ExpenseService struct {
	repository domain.ExpenseRepository
	logger     domain.Logger
}

func NewExpenseService(
	repository domain.ExpenseRepository,
	logger domain.Logger,
) domain.ExpenseService {
	return &ExpenseService{
		repository: repository,
		logger:     logger,
	}
}

func (s *ExpenseService) CreateExpenseRequestByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.CreateExpenseRequestParams,
) (*domain.ExpenseRequest, error) {
	if adminEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}
	params.CreatedByEmployeeID = adminEmployeeID

	if err := normalizeCreateExpenseParams(&params); err != nil {
		return nil, err
	}

	item, err := s.repository.CreateExpenseRequest(ctx, params)
	if err != nil {
		s.logError(ctx, "CreateExpenseRequestByAdmin", "failed to create expense request", err)
		return nil, err
	}
	return item, nil
}

func (s *ExpenseService) GetExpenseRequestByID(
	ctx context.Context,
	expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	if expenseRequestID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}
	item, err := s.repository.GetExpenseRequestByID(ctx, expenseRequestID)
	if err != nil {
		s.logError(ctx, "GetExpenseRequestByID", "failed to get expense request", err)
		return nil, err
	}
	return item, nil
}

func (s *ExpenseService) ListExpenseRequests(
	ctx context.Context,
	params domain.ListExpenseRequestsParams,
) (*domain.ExpenseRequestPage, error) {
	if params.Status != nil && !isValidExpenseStatus(*params.Status) {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}
	if params.Category != nil && !isValidExpenseCategory(*params.Category) {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	page, err := s.repository.ListExpenseRequests(ctx, params)
	if err != nil {
		s.logError(ctx, "ListExpenseRequests", "failed to list expense requests", err)
		return nil, err
	}
	return page, nil
}

func (s *ExpenseService) UpdateExpenseRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, expenseRequestID uuid.UUID,
	params domain.UpdateExpenseRequestParams,
) (*domain.ExpenseRequest, error) {
	if adminEmployeeID == uuid.Nil || expenseRequestID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	var updated *domain.ExpenseRequest
	err := s.repository.WithTx(ctx, func(tx domain.ExpenseTxRepository) error {
		current, err := tx.GetExpenseRequestForUpdate(ctx, expenseRequestID)
		if err != nil {
			return err
		}
		if current.Status == domain.ExpenseRequestStatusReimbursed ||
			current.Status == domain.ExpenseRequestStatusCancelled {
			return domain.ErrExpenseRequestStateInvalid
		}

		next, err := normalizeUpdateExpenseParams(*current, params)
		if err != nil {
			return err
		}

		updated, err = tx.UpdateExpenseRequestEditableFields(ctx, expenseRequestID, next)
		return err
	})
	if err != nil {
		s.logError(ctx, "UpdateExpenseRequestByAdmin", "failed to update expense request", err)
		return nil, err
	}
	return updated, nil
}

func (s *ExpenseService) DecideExpenseRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, expenseRequestID uuid.UUID,
	params domain.DecideExpenseRequestParams,
) (*domain.ExpenseRequest, error) {
	if adminEmployeeID == uuid.Nil || expenseRequestID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	var updated *domain.ExpenseRequest
	err := s.repository.WithTx(ctx, func(tx domain.ExpenseTxRepository) error {
		current, err := tx.GetExpenseRequestForUpdate(ctx, expenseRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.ExpenseRequestStatusPending {
			return domain.ErrExpenseRequestStateInvalid
		}

		if decision == "approve" {
			approvedAmount := current.ClaimedAmount
			if params.ApprovedAmount != nil {
				approvedAmount = *params.ApprovedAmount
			}
			if approvedAmount < 0 || approvedAmount > current.ClaimedAmount {
				return domain.ErrExpenseRequestInvalidRequest
			}

			updated, err = tx.ApproveExpenseRequest(
				ctx,
				expenseRequestID,
				adminEmployeeID,
				approvedAmount,
				trimmedPtr(params.DecisionNote),
			)
			return err
		}

		updated, err = tx.RejectExpenseRequest(
			ctx,
			expenseRequestID,
			adminEmployeeID,
			trimmedPtr(params.DecisionNote),
		)
		return err
	})
	if err != nil {
		s.logError(ctx, "DecideExpenseRequestByAdmin", "failed to decide expense request", err)
		return nil, err
	}
	return updated, nil
}

func (s *ExpenseService) CancelExpenseRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	if adminEmployeeID == uuid.Nil || expenseRequestID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	var updated *domain.ExpenseRequest
	err := s.repository.WithTx(ctx, func(tx domain.ExpenseTxRepository) error {
		current, err := tx.GetExpenseRequestForUpdate(ctx, expenseRequestID)
		if err != nil {
			return err
		}
		if current.Status == domain.ExpenseRequestStatusReimbursed ||
			current.Status == domain.ExpenseRequestStatusCancelled {
			return domain.ErrExpenseRequestStateInvalid
		}

		updated, err = tx.CancelExpenseRequest(ctx, expenseRequestID)
		return err
	})
	if err != nil {
		s.logError(ctx, "CancelExpenseRequestByAdmin", "failed to cancel expense request", err)
		return nil, err
	}
	return updated, nil
}

func (s *ExpenseService) MarkExpenseRequestReimbursedByAdmin(
	ctx context.Context,
	adminEmployeeID, expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	if adminEmployeeID == uuid.Nil || expenseRequestID == uuid.Nil {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	var updated *domain.ExpenseRequest
	err := s.repository.WithTx(ctx, func(tx domain.ExpenseTxRepository) error {
		current, err := tx.GetExpenseRequestForUpdate(ctx, expenseRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.ExpenseRequestStatusApproved {
			return domain.ErrExpenseRequestStateInvalid
		}

		updated, err = tx.MarkExpenseRequestReimbursed(ctx, expenseRequestID, adminEmployeeID)
		return err
	})
	if err != nil {
		s.logError(
			ctx,
			"MarkExpenseRequestReimbursedByAdmin",
			"failed to mark expense request reimbursed",
			err,
		)
		return nil, err
	}
	return updated, nil
}

func normalizeCreateExpenseParams(params *domain.CreateExpenseRequestParams) error {
	if params == nil {
		return domain.ErrExpenseRequestInvalidRequest
	}
	if params.EmployeeID == uuid.Nil || params.CreatedByEmployeeID == uuid.Nil {
		return domain.ErrExpenseRequestInvalidRequest
	}

	params.Category = strings.TrimSpace(params.Category)
	if !isValidExpenseCategory(params.Category) {
		return domain.ErrExpenseRequestInvalidRequest
	}

	if params.ExpenseDate.IsZero() {
		return domain.ErrExpenseRequestInvalidRequest
	}
	params.ExpenseDate = dateOnlyExpenseUTC(params.ExpenseDate)
	if params.ExpenseDate.After(currentExpenseUTCDate()) {
		return domain.ErrExpenseRequestInvalidRequest
	}

	params.Description = strings.TrimSpace(params.Description)
	params.BusinessPurpose = strings.TrimSpace(params.BusinessPurpose)
	if params.Description == "" || params.BusinessPurpose == "" {
		return domain.ErrExpenseRequestInvalidRequest
	}

	params.Currency = strings.TrimSpace(strings.ToUpper(params.Currency))
	if len(params.Currency) != 3 {
		return domain.ErrExpenseRequestInvalidRequest
	}
	if params.ClaimedAmount <= 0 {
		return domain.ErrExpenseRequestInvalidRequest
	}

	params.MerchantName = trimmedPtr(params.MerchantName)
	params.RequestNote = trimmedPtr(params.RequestNote)
	params.TravelMode = trimmedPtr(params.TravelMode)
	params.TravelFrom = trimmedPtr(params.TravelFrom)
	params.TravelTo = trimmedPtr(params.TravelTo)

	if params.Category != "travel" {
		if params.TravelMode != nil || params.TravelFrom != nil ||
			params.TravelTo != nil || params.DistanceKm != nil {
			return domain.ErrExpenseRequestInvalidRequest
		}
	} else if params.DistanceKm != nil && *params.DistanceKm < 0 {
		return domain.ErrExpenseRequestInvalidRequest
	}

	return nil
}

func normalizeUpdateExpenseParams(
	current domain.ExpenseRequest,
	update domain.UpdateExpenseRequestParams,
) (domain.UpdateExpenseRequestParams, error) {
	out := update

	finalCategory := strings.TrimSpace(current.Category)
	if update.Category != nil {
		finalCategory = strings.TrimSpace(*update.Category)
		if !isValidExpenseCategory(finalCategory) {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		value := finalCategory
		out.Category = &value
	}

	finalExpenseDate := dateOnlyExpenseUTC(current.ExpenseDate)
	if update.ExpenseDate != nil {
		if update.ExpenseDate.IsZero() {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		finalExpenseDate = dateOnlyExpenseUTC(*update.ExpenseDate)
		if finalExpenseDate.After(currentExpenseUTCDate()) {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		out.ExpenseDate = &finalExpenseDate
	}

	finalDescription := current.Description
	if update.Description != nil {
		finalDescription = strings.TrimSpace(*update.Description)
		if finalDescription == "" {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		out.Description = &finalDescription
	}

	finalBusinessPurpose := current.BusinessPurpose
	if update.BusinessPurpose != nil {
		finalBusinessPurpose = strings.TrimSpace(*update.BusinessPurpose)
		if finalBusinessPurpose == "" {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		out.BusinessPurpose = &finalBusinessPurpose
	}

	if update.Currency != nil {
		currency := strings.TrimSpace(strings.ToUpper(*update.Currency))
		if len(currency) != 3 {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		out.Currency = &currency
	}

	finalClaimedAmount := current.ClaimedAmount
	if update.ClaimedAmount != nil {
		finalClaimedAmount = *update.ClaimedAmount
		if finalClaimedAmount <= 0 {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
	}
	if current.ApprovedAmount != nil && *current.ApprovedAmount > finalClaimedAmount {
		return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
	}

	out.MerchantName = trimmedPtr(update.MerchantName)
	out.RequestNote = trimmedPtr(update.RequestNote)
	out.TravelMode = trimmedPtr(update.TravelMode)
	out.TravelFrom = trimmedPtr(update.TravelFrom)
	out.TravelTo = trimmedPtr(update.TravelTo)

	finalTravelMode := valueOrCurrent(out.TravelMode, current.TravelMode)
	finalTravelFrom := valueOrCurrent(out.TravelFrom, current.TravelFrom)
	finalTravelTo := valueOrCurrent(out.TravelTo, current.TravelTo)
	finalDistanceKm := current.DistanceKm
	if update.DistanceKm != nil {
		if *update.DistanceKm < 0 {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
		finalDistanceKm = update.DistanceKm
	}

	if finalCategory != "travel" {
		if finalTravelMode != nil || finalTravelFrom != nil || finalTravelTo != nil || finalDistanceKm != nil {
			return domain.UpdateExpenseRequestParams{}, domain.ErrExpenseRequestInvalidRequest
		}
	}

	return out, nil
}

func valueOrCurrent(next, current *string) *string {
	if next != nil {
		return next
	}
	return current
}

func trimmedPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isValidExpenseCategory(value string) bool {
	switch strings.TrimSpace(value) {
	case "travel", "meal", "accommodation", "office_supplies", "training", "client_entertainment", "other":
		return true
	default:
		return false
	}
}

func isValidExpenseStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "pending", "approved", "rejected", "reimbursed", "cancelled":
		return true
	default:
		return false
	}
}

func currentExpenseUTCDate() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func dateOnlyExpenseUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *ExpenseService) logError(
	ctx context.Context,
	operation string,
	message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger == nil {
		return
	}

	attrs := []zap.Field{
		zap.String("operation", operation),
	}
	attrs = append(attrs, fields...)
	s.logger.LogError(ctx, "ExpenseService", message, err, attrs...)
}
