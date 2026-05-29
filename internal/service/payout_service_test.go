package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestDecidePayoutRequestByAdminApproveDeductsExtraLeaveBalance(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	requestID := uuid.New()
	employeeID := uuid.New()
	balanceID := uuid.New()
	salaryMonth := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			currentRequest: &domain.PayoutRequest{
				ID:             requestID,
				EmployeeID:     employeeID,
				RequestedHours: 8,
				BalanceYear:    2026,
				Status:         domain.PayoutRequestStatusPending,
			},
			balance: &domain.PayoutBalanceSnapshot{
				LeaveBalanceID: balanceID,
				ExtraRemaining: 16,
			},
		},
	}
	service := NewPayoutService(repo, nil)

	updated, err := service.DecidePayoutRequestByAdmin(ctx, adminID, requestID, domain.DecidePayoutRequestParams{
		Decision:    "approve",
		SalaryMonth: &salaryMonth,
	})
	if err != nil {
		t.Fatalf("expected approve to succeed, got error: %v", err)
	}
	if updated.Status != domain.PayoutRequestStatusApproved {
		t.Fatalf("expected approved status, got %q", updated.Status)
	}
	if repo.tx.deductBalanceID != balanceID {
		t.Fatalf("expected deduction from balance %s, got %s", balanceID, repo.tx.deductBalanceID)
	}
	if repo.tx.deductExtraMinutes != 480 {
		t.Fatalf("expected 480 deducted extra minutes, got %d", repo.tx.deductExtraMinutes)
	}
	if repo.tx.deductLegalMinutes != 0 {
		t.Fatalf("expected no legal minutes deducted, got %d", repo.tx.deductLegalMinutes)
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
	if repo.tx.deductCalls != 0 {
		t.Fatalf("expected no balance deduction on mark-paid, got %d calls", repo.tx.deductCalls)
	}
}

type fakePayoutRepository struct {
	tx *fakePayoutTxRepository
}

func (r *fakePayoutRepository) WithTx(ctx context.Context, fn func(tx domain.PayoutTxRepository) error) error {
	return fn(r.tx)
}

func (r *fakePayoutRepository) ListMyPayoutRequests(context.Context, domain.ListMyPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayoutRequests(context.Context, domain.ListPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	return nil, nil
}

func (r *fakePayoutRepository) GetPayrollPreviewEmployee(context.Context, uuid.UUID) (*domain.EmployeeDetail, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollPreviewWorkItems(context.Context, domain.PayrollPreviewParams) ([]domain.PayrollWorkItem, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListNationalHolidays(context.Context, string, time.Time, time.Time) ([]domain.NationalHoliday, error) {
	return nil, nil
}

func (r *fakePayoutRepository) GetPayPeriodByID(context.Context, uuid.UUID) (*domain.PayPeriod, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayPeriods(context.Context, domain.ListPayPeriodsParams) (*domain.PayPeriodPage, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayPeriodLineItems(context.Context, uuid.UUID) ([]domain.PayPeriodLineItem, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollMonthEmployees(context.Context, domain.PayrollMonthSummaryParams, time.Time, time.Time) ([]domain.PayrollMonthEmployee, int64, error) {
	return nil, 0, nil
}

func (r *fakePayoutRepository) ListPayrollMonthEmployeesAll(context.Context, domain.PayrollMonthORTOverviewParams, time.Time, time.Time) ([]domain.PayrollMonthEmployee, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayPeriodsByEmployeesAndRange(context.Context, []uuid.UUID, time.Time, time.Time) ([]domain.PayPeriod, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollMonthLockedMultiplierSummaries(context.Context, []uuid.UUID) ([]domain.PayrollLockedMultiplierSummary, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollMonthApprovedWorkItems(context.Context, []uuid.UUID, time.Time, time.Time) ([]domain.PayrollWorkItem, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollMonthPendingSummaries(context.Context, []uuid.UUID, time.Time, time.Time) ([]domain.PayrollMonthPendingSummary, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayrollMonthPendingEntries(context.Context, []uuid.UUID, time.Time, time.Time) ([]domain.PayrollMonthPendingEntry, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPendingOvertimeEntriesDetail(context.Context, uuid.UUID, time.Time, time.Time) ([]domain.PayrollPendingEntryDetail, error) {
	return nil, nil
}

func (r *fakePayoutRepository) ListPayoutRequestsByEmployeeAndMonth(context.Context, uuid.UUID, time.Time) ([]domain.PayoutRequest, error) {
	return nil, nil
}

func (r *fakePayoutRepository) GetLeaveBalanceExtraRemaining(context.Context, uuid.UUID, int32) (int32, error) {
	return 0, nil
}

type fakePayoutTxRepository struct {
	currentRequest     *domain.PayoutRequest
	balance            *domain.PayoutBalanceSnapshot
	deductCalls        int
	deductBalanceID    uuid.UUID
	deductExtraMinutes int32
	deductLegalMinutes int32
}

func (r *fakePayoutTxRepository) GetEmployeePayoutContract(context.Context, uuid.UUID) (*domain.PayoutContract, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) EnsureLeaveBalanceForYear(context.Context, uuid.UUID, int32) error {
	return nil
}

func (r *fakePayoutTxRepository) GetPayoutBalanceForUpdate(context.Context, uuid.UUID, int32) (*domain.PayoutBalanceSnapshot, error) {
	return r.balance, nil
}

func (r *fakePayoutTxRepository) CreatePayoutRequest(context.Context, domain.CreatePayoutRequestTxParams) (*domain.PayoutRequest, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) GetPayoutRequestForUpdate(context.Context, uuid.UUID) (*domain.PayoutRequest, error) {
	return r.currentRequest, nil
}

func (r *fakePayoutTxRepository) ApprovePayoutRequest(_ context.Context, _ uuid.UUID, decidedByEmployeeID uuid.UUID, salaryMonth time.Time, _ *string) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.Status = domain.PayoutRequestStatusApproved
	updated.DecidedByEmployeeID = &decidedByEmployeeID
	updated.SalaryMonth = &salaryMonth
	return &updated, nil
}

func (r *fakePayoutTxRepository) RejectPayoutRequest(_ context.Context, _ uuid.UUID, decidedByEmployeeID uuid.UUID, _ *string) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.Status = domain.PayoutRequestStatusRejected
	updated.DecidedByEmployeeID = &decidedByEmployeeID
	return &updated, nil
}

func (r *fakePayoutTxRepository) MarkPayoutRequestPaid(_ context.Context, _ uuid.UUID, paidByEmployeeID uuid.UUID) (*domain.PayoutRequest, error) {
	updated := *r.currentRequest
	updated.Status = domain.PayoutRequestStatusPaid
	updated.PaidByEmployeeID = &paidByEmployeeID
	return &updated, nil
}

func (r *fakePayoutTxRepository) ApplyLeaveBalanceDeduction(_ context.Context, balanceID uuid.UUID, extraHours, legalHours int32) (*domain.LeaveBalance, error) {
	r.deductCalls++
	r.deductBalanceID = balanceID
	r.deductExtraMinutes = extraHours
	r.deductLegalMinutes = legalHours
	return &domain.LeaveBalance{ID: balanceID}, nil
}

func (r *fakePayoutTxRepository) GetPayPeriodByEmployeePeriod(context.Context, uuid.UUID, time.Time, time.Time) (*domain.PayPeriod, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) LockPayrollOvertimeEntries(context.Context, domain.PayrollPreviewParams) ([]uuid.UUID, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) LockPayrollPreviewWorkItems(context.Context, domain.PayrollPreviewParams) ([]domain.PayrollWorkItem, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) CreatePayPeriod(context.Context, domain.ClosePayPeriodParams, uuid.UUID, domain.PayrollPreview) (*domain.PayPeriod, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) CreatePayPeriodLineItem(context.Context, uuid.UUID, domain.PayPeriodLineItem) (*domain.PayPeriodLineItem, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) AssignOvertimeEntriesToPayPeriod(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}

func (r *fakePayoutTxRepository) GetPayPeriodForUpdate(context.Context, uuid.UUID) (*domain.PayPeriod, error) {
	return nil, nil
}

func (r *fakePayoutTxRepository) MarkPayPeriodPaid(context.Context, uuid.UUID) (*domain.PayPeriod, error) {
	return nil, nil
}
