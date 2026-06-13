package service

import (
	"context"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PayoutService struct {
	repository domain.PayoutRepository
	logger     domain.Logger
}

func NewPayoutService(
	repository domain.PayoutRepository,
	logger domain.Logger,
) domain.PayoutService {
	return &PayoutService{
		repository: repository,
		logger:     logger,
	}
}

func (s *PayoutService) CreatePayoutRequest(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreatePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if actorEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil ||
		params.CreatedByEmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.EmployeeID != actorEmployeeID || params.CreatedByEmployeeID != actorEmployeeID {
		return nil, domain.ErrPayoutRequestForbidden
	}
	if params.RequestedHours <= 0 || params.BalanceYear < 2000 || params.BalanceYear > 2100 {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var created *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		if err := tx.LockEmployeeForLeaveBalance(ctx, params.EmployeeID); err != nil {
			return err
		}

		legalTotalMinutes, err := tx.ComputeLegalLeaveTotalForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
			payoutBalanceAsOf(params.BalanceYear),
		)
		if err != nil {
			return err
		}

		legalUsedMinutes, err := tx.ComputeLegalLeaveUsedForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
		)
		if err != nil {
			return err
		}

		reservedPayoutMinutes, err := tx.ComputeReservedPayoutMinutesForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
		)
		if err != nil {
			return err
		}

		requestedMinutes := params.RequestedHours * 60
		availableMinutes := legalTotalMinutes - legalUsedMinutes - reservedPayoutMinutes
		if availableMinutes < requestedMinutes {
			return domain.ErrLeaveBalanceInsufficient
		}

		contract, err := tx.GetEmployeePayoutContract(ctx, params.EmployeeID)
		if err != nil {
			return err
		}
		if contract.ContractRate == nil || *contract.ContractRate <= 0 {
			return domain.ErrPayoutRequestInvalidRequest
		}

		created, err = tx.CreatePayoutRequest(ctx, domain.CreatePayoutRequestTxParams{
			EmployeeID:          params.EmployeeID,
			CreatedByEmployeeID: params.CreatedByEmployeeID,
			RequestedHours:      params.RequestedHours,
			BalanceYear:         params.BalanceYear,
			HourlyRate:          *contract.ContractRate,
			GrossAmount:         float64(params.RequestedHours) * *contract.ContractRate,
			RequestNote:         params.RequestNote,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *PayoutService) UpdatePayoutRequest(
	ctx context.Context,
	actorEmployeeID, payoutRequestID uuid.UUID,
	params domain.UpdatePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if actorEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}

		if current.EmployeeID != actorEmployeeID {
			return domain.ErrPayoutRequestForbidden
		}

		if current.Status == domain.PayoutRequestStatusApproved ||
			current.Status == domain.PayoutRequestStatusPaid {
			return domain.ErrPayoutRequestStateInvalid
		}

		newGrossAmount := float64(params.RequestedHours) * current.HourlyRate

		updated, err = tx.UpdatePayoutRequest(
			ctx,
			payoutRequestID,
			domain.UpdatePayoutRequestTxParams{
				RequestedHours: params.RequestedHours,
				BalanceYear:    params.BalanceYear,
				GrossAmount:    newGrossAmount,
				RequestNote:    params.RequestNote,
			},
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) UpdatePayoutRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, payoutRequestID uuid.UUID,
	params domain.UpdatePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}

		newGrossAmount := float64(params.RequestedHours) * current.HourlyRate

		updated, err = tx.UpdatePayoutRequest(
			ctx,
			payoutRequestID,
			domain.UpdatePayoutRequestTxParams{
				RequestedHours: params.RequestedHours,
				BalanceYear:    params.BalanceYear,
				GrossAmount:    newGrossAmount,
				RequestNote:    params.RequestNote,
			},
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) CreateApprovedPayoutRequestByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.CreatePayoutRequestByAdminParams,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.RequestedHours <= 0 || params.BalanceYear < 2000 || params.BalanceYear > 2100 ||
		params.PayPeriodStart.IsZero() {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var approved *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		if err := tx.LockEmployeeForLeaveBalance(ctx, params.EmployeeID); err != nil {
			return err
		}

		legalTotalMinutes, err := tx.ComputeLegalLeaveTotalForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
			payoutBalanceAsOf(params.BalanceYear),
		)
		if err != nil {
			return err
		}

		legalUsedMinutes, err := tx.ComputeLegalLeaveUsedForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
		)
		if err != nil {
			return err
		}

		reservedPayoutMinutes, err := tx.ComputeReservedPayoutMinutesForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
		)
		if err != nil {
			return err
		}

		requestedMinutes := params.RequestedHours * 60
		availableMinutes := legalTotalMinutes - legalUsedMinutes - reservedPayoutMinutes
		if availableMinutes < requestedMinutes {
			return domain.ErrLeaveBalanceInsufficient
		}

		contract, err := tx.GetEmployeePayoutContract(ctx, params.EmployeeID)
		if err != nil {
			return err
		}
		if contract.ContractRate == nil || *contract.ContractRate <= 0 {
			return domain.ErrPayoutRequestInvalidRequest
		}

		created, err := tx.CreatePayoutRequest(ctx, domain.CreatePayoutRequestTxParams{
			EmployeeID:          params.EmployeeID,
			CreatedByEmployeeID: adminEmployeeID,
			RequestedHours:      params.RequestedHours,
			BalanceYear:         params.BalanceYear,
			HourlyRate:          *contract.ContractRate,
			GrossAmount:         float64(params.RequestedHours) * *contract.ContractRate,
			RequestNote:         params.RequestNote,
		})
		if err != nil {
			return err
		}

		approved, err = tx.ApprovePayoutRequest(
			ctx,
			created.ID,
			adminEmployeeID,
			params.PayPeriodStart,
			params.DecisionNote,
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return approved, nil
}

func (s *PayoutService) DecidePayoutRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, payoutRequestID uuid.UUID,
	params domain.DecidePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if decision == "approve" && params.PayPeriodStart == nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayoutRequestStatusPending {
			return domain.ErrPayoutRequestStateInvalid
		}

		if decision == "approve" {
			if err := tx.LockEmployeeForLeaveBalance(ctx, current.EmployeeID); err != nil {
				return err
			}

			legalTotalMinutes, err := tx.ComputeLegalLeaveTotalForYear(
				ctx,
				current.EmployeeID,
				current.BalanceYear,
				payoutBalanceAsOf(current.BalanceYear),
			)
			if err != nil {
				return err
			}

			legalUsedMinutes, err := tx.ComputeLegalLeaveUsedForYear(
				ctx,
				current.EmployeeID,
				current.BalanceYear,
			)
			if err != nil {
				return err
			}

			reservedPayoutMinutes, err := tx.ComputeReservedPayoutMinutesForYear(
				ctx,
				current.EmployeeID,
				current.BalanceYear,
			)
			if err != nil {
				return err
			}

			if legalTotalMinutes-legalUsedMinutes-reservedPayoutMinutes < 0 {
				return domain.ErrLeaveBalanceInsufficient
			}

			updated, err = tx.ApprovePayoutRequest(
				ctx,
				payoutRequestID,
				adminEmployeeID,
				*params.PayPeriodStart,
				params.DecisionNote,
			)
			return err
		}

		updated, err = tx.RejectPayoutRequest(
			ctx,
			payoutRequestID,
			adminEmployeeID,
			params.DecisionNote,
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) MarkPayoutRequestPaidByAdmin(
	ctx context.Context,
	adminEmployeeID, payoutRequestID uuid.UUID,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayoutRequestStatusApproved {
			return domain.ErrPayoutRequestStateInvalid
		}

		updated, err = tx.MarkPayoutRequestPaid(ctx, payoutRequestID, adminEmployeeID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) ListMyPayoutRequests(
	ctx context.Context,
	params domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListMyPayoutRequests(ctx, params)
}

func (s *PayoutService) ListPayoutRequests(
	ctx context.Context,
	params domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListPayoutRequests(ctx, params)
}

func (s *PayoutService) logError(
	ctx context.Context,
	operation, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger == nil {
		return
	}
	s.logger.LogError(ctx, "PayoutService."+operation, message, err, fields...)
}

func isValidPayoutStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.PayoutRequestStatusPending,
		domain.PayoutRequestStatusApproved,
		domain.PayoutRequestStatusRejected,
		domain.PayoutRequestStatusPaid:
		return true
	default:
		return false
	}
}

func payoutBalanceAsOf(year int32) time.Time {
	now := time.Now().UTC()
	currentYear := int32(now.Year())
	if year < currentYear {
		return time.Date(int(year), 12, 31, 23, 59, 59, 0, time.UTC)
	}
	if year > currentYear {
		return time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return now
}
