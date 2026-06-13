package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type fakeEmployeeDashboardRepository struct {
	params               domain.GetEmployeeDashboardKPIsParams
	pendingRequestParams domain.ListEmployeeDashboardPendingRequestsParams
	result               *domain.EmployeeDashboardRepositoryKPI
	pendingRequestResult []domain.EmployeeDashboardPendingRequest
	err                  error
	pendingRequestErr    error
}

func (f *fakeEmployeeDashboardRepository) GetKPIs(
	_ context.Context,
	params domain.GetEmployeeDashboardKPIsParams,
) (*domain.EmployeeDashboardRepositoryKPI, error) {
	f.params = params
	return f.result, f.err
}

func (f *fakeEmployeeDashboardRepository) ListPendingRequests(
	_ context.Context,
	params domain.ListEmployeeDashboardPendingRequestsParams,
) ([]domain.EmployeeDashboardPendingRequest, error) {
	f.pendingRequestParams = params
	return f.pendingRequestResult, f.pendingRequestErr
}

type fakeEmployeeDashboardSalaryReader struct {
	employeeID  uuid.UUID
	periodStart time.Time
	periodEnd   time.Time
	result      *domain.SalaryPageData
	err         error
}

func (f *fakeEmployeeDashboardSalaryReader) GetMySalaryPage(
	_ context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) (*domain.SalaryPageData, error) {
	f.employeeID = employeeID
	f.periodStart = periodStart
	f.periodEnd = periodEnd
	return f.result, f.err
}

func TestEmployeeDashboardServiceGetKPIs(t *testing.T) {
	employeeID := uuid.New()
	currentYear := int32(time.Now().UTC().Year())
	repo := &fakeEmployeeDashboardRepository{
		result: &domain.EmployeeDashboardRepositoryKPI{
			LeaveBalance: domain.EmployeeDashboardLeaveBalanceKPI{
				Year:         currentYear,
				UsedMinutes:  480,
				TotalMinutes: 9600,
			},
			PendingLeaveRequests: 2,
			PendingSignatures:    3,
		},
	}
	salary := &fakeEmployeeDashboardSalaryReader{
		result: &domain.SalaryPageData{
			DataSource: "live",
			Preview: &domain.PayrollPreview{
				GrossAmount: 1840.50,
			},
		},
	}
	service := NewEmployeeDashboardService(repo, salary, nil)

	result, err := service.GetKPIs(context.Background(), employeeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.params.EmployeeID != employeeID {
		t.Fatalf("expected repo employee ID %s, got %s", employeeID, repo.params.EmployeeID)
	}
	if repo.params.Year != currentYear {
		t.Fatalf("expected repo year to be current year, got %d", repo.params.Year)
	}
	if salary.employeeID != employeeID {
		t.Fatalf("expected salary employee ID %s, got %s", employeeID, salary.employeeID)
	}
	if salary.periodStart.IsZero() || salary.periodEnd.IsZero() {
		t.Fatal("expected current payroll period to be passed to salary service")
	}
	if result.LeaveBalance.UsedMinutes != 480 || result.LeaveBalance.TotalMinutes != 9600 {
		t.Fatalf("unexpected leave balance: %+v", result.LeaveBalance)
	}
	if result.PendingLeaveRequests != 2 {
		t.Fatalf("expected 2 pending leave requests, got %d", result.PendingLeaveRequests)
	}
	if result.PendingSignatures != 3 {
		t.Fatalf("expected 3 pending signatures, got %d", result.PendingSignatures)
	}
	if result.EstimatedCurrentPay.GrossAmount != 1840.50 {
		t.Fatalf("expected gross amount 1840.50, got %f", result.EstimatedCurrentPay.GrossAmount)
	}
	if result.EstimatedCurrentPay.DataSource != "live" {
		t.Fatalf("expected live data source, got %s", result.EstimatedCurrentPay.DataSource)
	}
}

func TestEmployeeDashboardServiceGetKPIsInvalidEmployee(t *testing.T) {
	service := NewEmployeeDashboardService(
		&fakeEmployeeDashboardRepository{},
		&fakeEmployeeDashboardSalaryReader{},
		nil,
	)

	_, err := service.GetKPIs(context.Background(), uuid.Nil)
	if !errors.Is(err, domain.ErrEmployeeDashboardInvalidRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestEmployeeDashboardServiceListPendingRequests(t *testing.T) {
	employeeID := uuid.New()
	requestID := uuid.New()
	repo := &fakeEmployeeDashboardRepository{
		pendingRequestResult: []domain.EmployeeDashboardPendingRequest{
			{
				ID:          requestID,
				RequestType: "leave",
				Status:      "pending",
			},
		},
	}
	service := NewEmployeeDashboardService(repo, &fakeEmployeeDashboardSalaryReader{}, nil)

	items, err := service.ListPendingRequests(
		context.Background(),
		domain.ListEmployeeDashboardPendingRequestsParams{
			EmployeeID: employeeID,
			RecentDays: 30,
			Limit:      5,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.pendingRequestParams.EmployeeID != employeeID {
		t.Fatalf(
			"expected repo employee ID %s, got %s",
			employeeID,
			repo.pendingRequestParams.EmployeeID,
		)
	}
	if repo.pendingRequestParams.Limit != 5 {
		t.Fatalf("expected limit 5, got %d", repo.pendingRequestParams.Limit)
	}
	if repo.pendingRequestParams.Since.IsZero() {
		t.Fatal("expected service to set since timestamp")
	}
	if len(items) != 1 || items[0].ID != requestID {
		t.Fatalf("unexpected pending request items: %+v", items)
	}
}

func TestEmployeeDashboardServiceListPendingRequestsDefaultsAndCaps(t *testing.T) {
	employeeID := uuid.New()
	repo := &fakeEmployeeDashboardRepository{}
	service := NewEmployeeDashboardService(repo, &fakeEmployeeDashboardSalaryReader{}, nil)

	_, err := service.ListPendingRequests(
		context.Background(),
		domain.ListEmployeeDashboardPendingRequestsParams{
			EmployeeID: employeeID,
			RecentDays: 365,
			Limit:      100,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.pendingRequestParams.RecentDays != maxEmployeeDashboardPendingRequestDays {
		t.Fatalf(
			"expected days capped to %d, got %d",
			maxEmployeeDashboardPendingRequestDays,
			repo.pendingRequestParams.RecentDays,
		)
	}
	if repo.pendingRequestParams.Limit != maxEmployeeDashboardPendingRequestLimit {
		t.Fatalf(
			"expected limit capped to %d, got %d",
			maxEmployeeDashboardPendingRequestLimit,
			repo.pendingRequestParams.Limit,
		)
	}
}

func TestEmployeeDashboardServiceListPendingRequestsInvalidDays(t *testing.T) {
	service := NewEmployeeDashboardService(
		&fakeEmployeeDashboardRepository{},
		&fakeEmployeeDashboardSalaryReader{},
		nil,
	)

	_, err := service.ListPendingRequests(
		context.Background(),
		domain.ListEmployeeDashboardPendingRequestsParams{
			EmployeeID: uuid.New(),
			RecentDays: -1,
		},
	)
	if !errors.Is(err, domain.ErrEmployeeDashboardInvalidRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestEmployeeDashboardGrossAmountUsesLockedPayPeriod(t *testing.T) {
	amount := employeeDashboardGrossAmount(&domain.SalaryPageData{
		PayPeriod: &domain.PayPeriod{GrossAmount: 2100},
		Preview:   &domain.PayrollPreview{GrossAmount: 1900},
	})

	if amount != 2100 {
		t.Fatalf("expected locked pay period gross amount, got %f", amount)
	}
}
