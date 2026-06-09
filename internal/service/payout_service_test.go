package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestDecidePayoutRequestByAdminApproveUnavailable(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	requestID := uuid.New()
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{}}
	service := NewPayoutService(repo, nil)

	_, err := service.DecidePayoutRequestByAdmin(
		ctx,
		adminID,
		requestID,
		domain.DecidePayoutRequestParams{
			Decision:       "approve",
			PayPeriodStart: &payPeriodStart,
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestInvalidRequest) {
		t.Fatalf("expected unavailable invalid request, got %v", err)
	}
}

func TestUpdatePayoutRequestSuccess(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	requestID := uuid.New()

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     employeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				HourlyRate:     25.50,
				Status:         domain.PayoutRequestStatusPending,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	updated, err := service.UpdatePayoutRequest(
		ctx,
		employeeID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 12,
			BalanceYear:    2026,
			RequestNote:    ptrString("updated note"),
		},
	)
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v", err)
	}

	if updated.RequestedHours != 12 {
		t.Fatalf("expected 12 requested hours, got %d", updated.RequestedHours)
	}
	expectedGross := 12 * 25.50
	if updated.GrossAmount != expectedGross {
		t.Fatalf("expected gross amount %f, got %f", expectedGross, updated.GrossAmount)
	}
	if updated.RequestNote == nil || *updated.RequestNote != "updated note" {
		t.Fatalf("expected request note 'updated note', got %v", updated.RequestNote)
	}
}

func TestUpdatePayoutRequestForbidden(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	anotherEmployeeID := uuid.New()
	requestID := uuid.New()

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     anotherEmployeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				HourlyRate:     25.50,
				Status:         domain.PayoutRequestStatusPending,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	_, err := service.UpdatePayoutRequest(
		ctx,
		employeeID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 12,
			BalanceYear:    2026,
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestForbidden) {
		t.Fatalf("expected forbidden error, got: %v", err)
	}
}

func TestUpdatePayoutRequestApprovedStateInvalid(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	requestID := uuid.New()

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     employeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				HourlyRate:     25.50,
				Status:         domain.PayoutRequestStatusApproved,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	_, err := service.UpdatePayoutRequest(
		ctx,
		employeeID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 12,
			BalanceYear:    2026,
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestStateInvalid) {
		t.Fatalf("expected invalid state error, got: %v", err)
	}
}

func TestUpdatePayoutRequestPaidStateInvalid(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	requestID := uuid.New()

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     employeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				HourlyRate:     25.50,
				Status:         domain.PayoutRequestStatusPaid,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	_, err := service.UpdatePayoutRequest(
		ctx,
		employeeID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 12,
			BalanceYear:    2026,
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestStateInvalid) {
		t.Fatalf("expected invalid state error, got: %v", err)
	}
}

func TestMarkPayoutRequestPaidByAdminDoesNotDeductLeaveBalanceAgain(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	requestID := uuid.New()
	employeeID := uuid.New()

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     employeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				Status:         domain.PayoutRequestStatusApproved,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	updated, err := service.MarkPayoutRequestPaidByAdmin(ctx, adminID, requestID)
	if err != nil {
		t.Fatalf("expected mark paid to succeed, got error: %v", err)
	}
	if updated.Status != domain.PayoutRequestStatusPaid {
		t.Fatalf("expected paid status, got %q", updated.Status)
	}

}

type fakePayoutRepository struct {
	tx *fakePayoutTxRepository
}

func TestCreateApprovedPayoutRequestByAdminUnavailable(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	employeeID := uuid.New()
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	requestedHours := int32(8)
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{}}
	service := NewPayoutService(repo, nil)

	_, err := service.CreateApprovedPayoutRequestByAdmin(
		ctx,
		adminID,
		domain.CreatePayoutRequestByAdminParams{
			EmployeeID:     employeeID,
			RequestedHours: requestedHours,
			BalanceYear:    2026,
			PayPeriodStart: payPeriodStart,
			RequestNote:    ptrString("admin-initiated"),
			DecisionNote:   ptrString("approved by admin"),
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestInvalidRequest) {
		t.Fatalf("expected unavailable invalid request, got %v", err)
	}
}

func TestCreateApprovedPayoutRequestByAdminNoLongerChecksExtraHours(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	employeeID := uuid.New()
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{}}
	service := NewPayoutService(repo, nil)

	_, err := service.CreateApprovedPayoutRequestByAdmin(
		ctx,
		adminID,
		domain.CreatePayoutRequestByAdminParams{
			EmployeeID:     employeeID,
			RequestedHours: 8,
			BalanceYear:    2026,
			PayPeriodStart: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	)
	if !errors.Is(err, domain.ErrPayoutRequestInvalidRequest) {
		t.Fatalf("expected unavailable invalid request, got %v", err)
	}
}

func ptrString(s string) *string {
	return &s
}

func (r *fakePayoutRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.PayoutTxRepository) error,
) error {
	return fn(r.tx)
}

func (r *fakePayoutRepository) ListMyPayoutRequests(
	context.Context,
	domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayoutRequests(
	context.Context,
	domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	return nil, nil
}

type fakePayoutTxRepository struct {
	currentRequest     *domain.PayoutRequest
	payoutContract     *domain.PayoutContract
	createdPayout      *domain.PayoutRequest
	createPayoutParams domain.CreatePayoutRequestTxParams
}

func (r *fakePayoutTxRepository) GetEmployeePayoutContract(
	context.Context,
	uuid.UUID,
) (*domain.PayoutContract, error) {
	return r.payoutContract, nil
}

func (r *fakePayoutTxRepository) CreatePayoutRequest(
	_ context.Context,
	params domain.CreatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	r.createPayoutParams = params
	created := &domain.PayoutRequest{
		ID:                  uuid.New(),
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: params.CreatedByEmployeeID,
		RequestedHours:      params.RequestedHours,
		BalanceYear:         params.BalanceYear,
		HourlyRate:          params.HourlyRate,
		GrossAmount:         params.GrossAmount,
		RequestNote:         params.RequestNote,
		Status:              domain.PayoutRequestStatusPending,
	}
	r.createdPayout = created
	return created, nil
}

func (r *fakePayoutTxRepository) GetPayoutRequestForUpdate(
	context.Context,
	uuid.UUID,
) (*domain.PayoutRequest, error) {
	return r.currentRequest, nil
}

func (r *fakePayoutTxRepository) UpdatePayoutRequest(
	_ context.Context,
	_ uuid.UUID,
	params domain.UpdatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.RequestedHours = params.RequestedHours
	updated.BalanceYear = params.BalanceYear
	updated.GrossAmount = params.GrossAmount
	updated.RequestNote = params.RequestNote
	return &updated, nil
}

func (r *fakePayoutTxRepository) ApprovePayoutRequest(
	_ context.Context,
	_ uuid.UUID,
	decidedByEmployeeID uuid.UUID,
	payPeriodStart time.Time,
	_ *string,
) (*domain.PayoutRequest, error) {
	base := r.currentRequest
	if base == nil {
		base = r.createdPayout
	}
	if base == nil {
		return nil, nil
	}
	updated := *base
	updated.Status = domain.PayoutRequestStatusApproved
	updated.DecidedByEmployeeID = &decidedByEmployeeID
	updated.PayPeriodStart = &payPeriodStart
	return &updated, nil
}

func (r *fakePayoutTxRepository) RejectPayoutRequest(
	_ context.Context,
	_ uuid.UUID,
	decidedByEmployeeID uuid.UUID,
	_ *string,
) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.Status = domain.PayoutRequestStatusRejected
	updated.DecidedByEmployeeID = &decidedByEmployeeID
	return &updated, nil
}

func (r *fakePayoutTxRepository) MarkPayoutRequestPaid(
	_ context.Context,
	_ uuid.UUID,
	paidByEmployeeID uuid.UUID,
) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.Status = domain.PayoutRequestStatusPaid
	updated.PaidByEmployeeID = &paidByEmployeeID
	return &updated, nil
}
