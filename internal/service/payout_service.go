package service

import (
	"context"
	"fmt"
	"strings"

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
	return nil, fmt.Errorf("%w: leave payout is unavailable", domain.ErrPayoutRequestInvalidRequest)
}

func (s *PayoutService) CreateApprovedPayoutRequestByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.CreatePayoutRequestByAdminParams,
) (*domain.PayoutRequest, error) {
	return nil, fmt.Errorf("%w: leave payout is unavailable", domain.ErrPayoutRequestInvalidRequest)
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
	if decision == "approve" {
		return nil, fmt.Errorf("%w: leave payout is unavailable", domain.ErrPayoutRequestInvalidRequest)
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
