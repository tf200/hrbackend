package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type fakeSalaryRepository struct {
	employee          *domain.EmployeeDetail
	payPeriods        []domain.PayPeriod
	lineItems         map[uuid.UUID][]domain.PayPeriodLineItem
	workItems         []domain.PayrollWorkItem
	contractSegments  []domain.FixedPayrollContractSegmentSource
	pendingEntries    []domain.PayrollPendingEntryDetail
	payoutRequests    []domain.PayoutRequest
	holidays          []domain.NationalHoliday
	err               error
}

func (f *fakeSalaryRepository) WithTxSalary(ctx context.Context, fn func(tx domain.SalaryTxRepository) error) error {
	return nil
}

func (f *fakeSalaryRepository) GetPayrollPreviewEmployee(ctx context.Context, employeeID uuid.UUID) (*domain.EmployeeDetail, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.employee, nil
}

func (f *fakeSalaryRepository) ListPayrollPreviewWorkItems(ctx context.Context, params domain.PayrollPreviewParams) ([]domain.PayrollWorkItem, error) {
	return f.workItems, nil
}

func (f *fakeSalaryRepository) ListNationalHolidays(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]domain.NationalHoliday, error) {
	return f.holidays, nil
}

func (f *fakeSalaryRepository) GetPayPeriodByID(ctx context.Context, payPeriodID uuid.UUID) (*domain.PayPeriod, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListPayPeriods(ctx context.Context, params domain.ListPayPeriodsParams) (*domain.PayPeriodPage, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListPayPeriodLineItems(ctx context.Context, payPeriodID uuid.UUID) ([]domain.PayPeriodLineItem, error) {
	return f.lineItems[payPeriodID], nil
}

func (f *fakeSalaryRepository) ListPayrollMonthEmployees(ctx context.Context, params domain.PayrollMonthSummaryParams, monthStart, monthEnd time.Time) ([]domain.PayrollMonthEmployee, int64, error) {
	return nil, 0, nil
}

func (f *fakeSalaryRepository) ListPayrollMonthEmployeesAll(ctx context.Context, params domain.PayrollMonthORTOverviewParams, monthStart, monthEnd time.Time) ([]domain.PayrollMonthEmployee, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListFixedPayrollMonthEmployees(ctx context.Context, params domain.PayrollMonthSummaryParams, monthStart, monthEnd time.Time) ([]domain.PayrollMonthEmployee, int64, error) {
	return nil, 0, nil
}

func (f *fakeSalaryRepository) ListOnCallPayrollMonthEmployees(ctx context.Context, params domain.PayrollMonthSummaryParams, monthStart, monthEnd time.Time) ([]domain.PayrollMonthEmployee, int64, error) {
	return nil, 0, nil
}

func (f *fakeSalaryRepository) ListFixedPayrollContractSegments(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.FixedPayrollContractSegmentSource, error) {
	return f.contractSegments, nil
}

func (f *fakeSalaryRepository) ListPayPeriodsByEmployeesAndRange(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayPeriod, error) {
	return f.payPeriods, nil
}

func (f *fakeSalaryRepository) ListPayrollMonthLockedMultiplierSummaries(ctx context.Context, payPeriodIDs []uuid.UUID) ([]domain.PayrollLockedMultiplierSummary, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListPayrollMonthApprovedWorkItems(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollWorkItem, error) {
	return f.workItems, nil
}

func (f *fakeSalaryRepository) ListPayrollMonthPendingSummaries(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollMonthPendingSummary, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListPayrollMonthPendingEntries(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollMonthPendingEntry, error) {
	return nil, nil
}

func (f *fakeSalaryRepository) ListPendingOvertimeEntriesDetail(ctx context.Context, employeeID uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollPendingEntryDetail, error) {
	return f.pendingEntries, nil
}

func (f *fakeSalaryRepository) ListPayoutRequestsByEmployeeAndMonth(ctx context.Context, employeeID uuid.UUID, salaryMonth time.Time) ([]domain.PayoutRequest, error) {
	return f.payoutRequests, nil
}

func (f *fakeSalaryRepository) ListSalaryScaleSteps(ctx context.Context, params domain.ListSalaryScaleStepsParams) (*domain.SalaryScaleStepsResult, error) {
	return nil, nil
}

func TestGetMySalaryPageLiveFixedEmployee(t *testing.T) {
	ctx := context.Background()
	employeeID := uuid.New()
	rate := 25.0
	hours := 40.0

	repo := &fakeSalaryRepository{
		employee: &domain.EmployeeDetail{
			ID:            employeeID,
			FirstName:     "John",
			LastName:      "Doe",
			ContractType:  "permanent",
			ContractRate:  &rate,
			ContractHours: &hours,
		},
		contractSegments: []domain.FixedPayrollContractSegmentSource{
			{
				EmployeeID:           employeeID,
				ContractID:           uuid.New(),
				ContractType:         "permanent",
				ActiveFrom:           time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
				ActiveUntil:          time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC),
				HoursPerWeek:         40,
				FullTimeHoursPerWeek: 40,
				HourlyRate:           25,
				MonthlySalary:        4000,
			},
		},
		workItems: []domain.PayrollWorkItem{
			{
				EmployeeID:            employeeID,
				SourceType:            domain.PayrollSourceSchedule,
				ContractType:          "permanent",
				ContractRate:          &rate,
				WorkDate:              time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
				StartTime:             "09:00",
				EndTime:               "17:00",
				IrregularHoursProfile: "none",
			},
		},
	}
	salaryService := NewSalaryService(repo, nil)

	periodStart := time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 0, 27)

	data, err := salaryService.GetMySalaryPage(ctx, employeeID, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if data.DataSource != "live" {
		t.Fatalf("expected live data source, got %s", data.DataSource)
	}
	if data.Preview == nil {
		t.Fatalf("expected preview to be populated")
	}

	if data.Preview.BaseGrossAmount != 4000 {
		t.Fatalf("expected BaseGrossAmount to be 4000 (fixed contract base), got %f", data.Preview.BaseGrossAmount)
	}

	if len(data.Preview.LineItems) != 2 {
		t.Fatalf("expected 2 line items (fixed_base and schedule), got %d", len(data.Preview.LineItems))
	}

	var hasFixedBase, hasSchedule bool
	for _, item := range data.Preview.LineItems {
		if item.SourceType == "fixed_base" {
			hasFixedBase = true
			if item.BaseAmount != 4000 {
				t.Fatalf("expected fixed base amount 4000, got %f", item.BaseAmount)
			}
		}
		if item.SourceType == "schedule" {
			hasSchedule = true
			if item.BaseAmount != 0 {
				t.Fatalf("expected schedule base amount to be zeroed out for fixed contract, got %f", item.BaseAmount)
			}
		}
	}

	if !hasFixedBase || !hasSchedule {
		t.Fatalf("expected both fixed_base and schedule line items")
	}
}
