package service

import (
	"context"
	"math"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type PayoutService struct {
	repository domain.PayoutRepository
}

func NewPayoutService(repository domain.PayoutRepository) domain.PayoutService {
	return &PayoutService{repository: repository}
}

func (s *PayoutService) CreatePayoutRequest(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CreatePayoutRequestParams) (*domain.PayoutRequest, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.RequestedHours <= 0 || params.BalanceYear < 2000 || params.BalanceYear > 2100 {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	params.EmployeeID = actorEmployeeID
	params.CreatedByEmployeeID = actorEmployeeID

	var created *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		contract, err := tx.GetEmployeePayoutContract(ctx, params.EmployeeID)
		if err != nil {
			return err
		}
		if contract.ContractType != "loondienst" {
			return domain.ErrPayoutRequestInvalidRequest
		}
		if contract.ContractRate == nil || *contract.ContractRate <= 0 {
			return domain.ErrPayoutRequestInvalidRequest
		}

		if err := tx.EnsureLeaveBalanceForYear(ctx, params.EmployeeID, params.BalanceYear); err != nil {
			return err
		}
		balance, err := tx.GetPayoutBalanceForUpdate(ctx, params.EmployeeID, params.BalanceYear)
		if err != nil {
			return err
		}
		if balance.ExtraRemaining < params.RequestedHours {
			return domain.ErrPayoutRequestInsufficientHours
		}

		hourlyRate := roundCurrency(*contract.ContractRate)
		grossAmount := roundCurrency(float64(params.RequestedHours) * hourlyRate)

		created, err = tx.CreatePayoutRequest(ctx, domain.CreatePayoutRequestTxParams{
			EmployeeID:          params.EmployeeID,
			CreatedByEmployeeID: params.CreatedByEmployeeID,
			RequestedHours:      params.RequestedHours,
			BalanceYear:         params.BalanceYear,
			HourlyRate:          hourlyRate,
			GrossAmount:         grossAmount,
			RequestNote:         params.RequestNote,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *PayoutService) DecidePayoutRequestByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID, params domain.DecidePayoutRequestParams) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if decision == "approve" && (params.SalaryMonth == nil || params.SalaryMonth.IsZero()) {
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
			if err := tx.EnsureLeaveBalanceForYear(ctx, current.EmployeeID, current.BalanceYear); err != nil {
				return err
			}
			balance, err := tx.GetPayoutBalanceForUpdate(ctx, current.EmployeeID, current.BalanceYear)
			if err != nil {
				return err
			}
			if balance.ExtraRemaining < current.RequestedHours {
				return domain.ErrPayoutRequestInsufficientHours
			}
			if _, err := tx.ApplyLeaveBalanceDeduction(ctx, balance.LeaveBalanceID, current.RequestedHours, 0); err != nil {
				return err
			}

			updated, err = tx.ApprovePayoutRequest(ctx, payoutRequestID, adminEmployeeID, params.SalaryMonth.UTC(), params.DecisionNote)
			return err
		}

		updated, err = tx.RejectPayoutRequest(ctx, payoutRequestID, adminEmployeeID, params.DecisionNote)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) MarkPayoutRequestPaidByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID) (*domain.PayoutRequest, error) {
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

func (s *PayoutService) ListMyPayoutRequests(ctx context.Context, params domain.ListMyPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListMyPayoutRequests(ctx, params)
}

func (s *PayoutService) ListPayoutRequests(ctx context.Context, params domain.ListPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListPayoutRequests(ctx, params)
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

func roundCurrency(v float64) float64 {
	return math.Round(v*100) / 100
}
