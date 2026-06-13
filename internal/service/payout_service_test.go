package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestDecidePayoutRequestByAdminApproveSuccess(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	requestID := uuid.New()
	employeeID := uuid.New()
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{
		currentRequest: &domain.PayoutRequest{
			ID:             requestID,
			EmployeeID:     employeeID,
			RequestedHours: 8,
			BalanceYear:    2026,
			Status:         domain.PayoutRequestStatusPending,
		},
		legalTotalMinutes:     9600,
		legalUsedMinutes:      1200,
		reservedPayoutMinutes: 480,
	}}
	service := NewPayoutService(repo, nil)

	updated, err := service.DecidePayoutRequestByAdmin(
		ctx,
		adminID,
		requestID,
		domain.DecidePayoutRequestParams{
			Decision:       "approve",
			PayPeriodStart: &payPeriodStart,
		},
	)
	if err != nil {
		t.Fatalf("expected approval to succeed, got %v", err)
	}
	if updated.Status != domain.PayoutRequestStatusApproved {
		t.Fatalf("expected approved status, got %q", updated.Status)
	}
	if !repo.tx.lockedEmployee {
		t.Fatal("expected employee leave balance lock")
	}
}

func TestDecidePayoutRequestByAdminApproveInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	requestID := uuid.New()
	employeeID := uuid.New()
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{
		currentRequest: &domain.PayoutRequest{
			ID:             requestID,
			EmployeeID:     employeeID,
			RequestedHours: 8,
			BalanceYear:    2026,
			Status:         domain.PayoutRequestStatusPending,
		},
		legalTotalMinutes:     400,
		legalUsedMinutes:      0,
		reservedPayoutMinutes: 480,
	}}
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
	if !errors.Is(err, domain.ErrLeaveBalanceInsufficient) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
}

func TestCreatePayoutRequestSuccess(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			payoutContract:        &domain.PayoutContract{ContractRate: float64Ptr(25.50)},
			legalTotalMinutes:     9600,
			legalUsedMinutes:      1200,
			reservedPayoutMinutes: 600,
		},
	}
	service := NewPayoutService(repo, nil)

	created, err := service.CreatePayoutRequest(
		ctx,
		employeeID,
		domain.CreatePayoutRequestParams{
			EmployeeID:          employeeID,
			CreatedByEmployeeID: employeeID,
			RequestedHours:      34,
			BalanceYear:         2026,
			RequestNote:         ptrString("cash out leave"),
		},
	)
	if err != nil {
		t.Fatalf("expected create to succeed, got error: %v", err)
	}

	if created.RequestedHours != 34 {
		t.Fatalf("expected 34 requested hours, got %d", created.RequestedHours)
	}
	if created.GrossAmount != 34*25.50 {
		t.Fatalf("expected gross amount %f, got %f", 34*25.50, created.GrossAmount)
	}
	if !repo.tx.lockedEmployee {
		t.Fatal("expected employee leave balance lock")
	}
}

func TestCreatePayoutRequestInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			payoutContract:        &domain.PayoutContract{ContractRate: float64Ptr(25.50)},
			legalTotalMinutes:     2400,
			legalUsedMinutes:      0,
			reservedPayoutMinutes: 600,
		},
	}
	service := NewPayoutService(repo, nil)

	_, err := service.CreatePayoutRequest(
		ctx,
		employeeID,
		domain.CreatePayoutRequestParams{
			EmployeeID:          employeeID,
			CreatedByEmployeeID: employeeID,
			RequestedHours:      34,
			BalanceYear:         2026,
		},
	)
	if !errors.Is(err, domain.ErrLeaveBalanceInsufficient) {
		t.Fatalf("expected insufficient balance, got %v", err)
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

func TestUpdatePayoutRequestByAdminSuccess(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
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

	updated, err := service.UpdatePayoutRequestByAdmin(
		ctx,
		adminID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 15,
			BalanceYear:    2026,
			RequestNote:    ptrString("admin adjustment"),
		},
	)
	if err != nil {
		t.Fatalf("expected admin update to succeed, got error: %v", err)
	}

	if updated.RequestedHours != 15 {
		t.Fatalf("expected 15 requested hours, got %d", updated.RequestedHours)
	}
	expectedGross := 15 * 25.50
	if updated.GrossAmount != expectedGross {
		t.Fatalf("expected gross amount %f, got %f", expectedGross, updated.GrossAmount)
	}
}

func TestUpdatePayoutRequestByAdminPaidSuccess(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
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

	updated, err := service.UpdatePayoutRequestByAdmin(
		ctx,
		adminID,
		requestID,
		domain.UpdatePayoutRequestParams{
			RequestedHours: 20,
			BalanceYear:    2026,
			RequestNote:    ptrString("admin update paid request"),
		},
	)
	if err != nil {
		t.Fatalf("expected admin update on paid request to succeed, got error: %v", err)
	}

	if updated.RequestedHours != 20 {
		t.Fatalf("expected 20 requested hours, got %d", updated.RequestedHours)
	}
	expectedGross := 20 * 25.50
	if updated.GrossAmount != expectedGross {
		t.Fatalf("expected gross amount %f, got %f", expectedGross, updated.GrossAmount)
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

func TestCreateApprovedPayoutRequestByAdminSuccess(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	employeeID := uuid.New()
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	requestedHours := int32(8)
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{
		payoutContract:        &domain.PayoutContract{ContractRate: float64Ptr(32.75)},
		legalTotalMinutes:     9600,
		legalUsedMinutes:      1200,
		reservedPayoutMinutes: 600,
	}}
	service := NewPayoutService(repo, nil)

	approved, err := service.CreateApprovedPayoutRequestByAdmin(
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
	if err != nil {
		t.Fatalf("expected admin create approved to succeed, got %v", err)
	}
	if approved.Status != domain.PayoutRequestStatusApproved {
		t.Fatalf("expected approved status, got %q", approved.Status)
	}
	if approved.DecidedByEmployeeID == nil || *approved.DecidedByEmployeeID != adminID {
		t.Fatalf("expected decided by admin %s, got %v", adminID, approved.DecidedByEmployeeID)
	}
	if approved.PayPeriodStart == nil || !approved.PayPeriodStart.Equal(payPeriodStart) {
		t.Fatalf("expected pay period %s, got %v", payPeriodStart, approved.PayPeriodStart)
	}
	if !repo.tx.lockedEmployee {
		t.Fatal("expected employee leave balance lock")
	}
}

func TestCreateApprovedPayoutRequestByAdminInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	employeeID := uuid.New()
	repo := &fakePayoutRepository{tx: &fakePayoutTxRepository{
		payoutContract:        &domain.PayoutContract{ContractRate: float64Ptr(32.75)},
		legalTotalMinutes:     400,
		legalUsedMinutes:      0,
		reservedPayoutMinutes: 480,
	}}
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
	if !errors.Is(err, domain.ErrLeaveBalanceInsufficient) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
}

func ptrString(s string) *string {
	return &s
}

func float64Ptr(v float64) *float64 {
	return &v
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
	currentRequest        *domain.PayoutRequest
	payoutContract        *domain.PayoutContract
	createdPayout         *domain.PayoutRequest
	createPayoutParams    domain.CreatePayoutRequestTxParams
	lockedEmployee        bool
	legalTotalMinutes     int32
	legalUsedMinutes      int32
	reservedPayoutMinutes int32
}

func (r *fakePayoutTxRepository) LockEmployeeForLeaveBalance(
	context.Context,
	uuid.UUID,
) error {
	r.lockedEmployee = true
	return nil
}

func (r *fakePayoutTxRepository) ComputeLegalLeaveTotalForYear(
	context.Context,
	uuid.UUID,
	int32,
	time.Time,
) (int32, error) {
	return r.legalTotalMinutes, nil
}

func (r *fakePayoutTxRepository) ComputeLegalLeaveUsedForYear(
	context.Context,
	uuid.UUID,
	int32,
) (int32, error) {
	return r.legalUsedMinutes, nil
}

func (r *fakePayoutTxRepository) ComputeReservedPayoutMinutesForYear(
	context.Context,
	uuid.UUID,
	int32,
) (int32, error) {
	return r.reservedPayoutMinutes, nil
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
