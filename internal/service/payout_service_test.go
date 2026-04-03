package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestPreviewPayrollRosterShiftSplitsAcrossEveningAndNight(t *testing.T) {
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("11111111-1111-1111-1111-111111111111"),
			FirstName: "Sara",
			LastName:  "Jansen",
		},
		entries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("22222222-2222-2222-2222-222222222222"),
				EmployeeID:            mustUUID("11111111-1111-1111-1111-111111111111"),
				EmployeeName:          "Sara Jansen",
				EntryDate:             dateUTC(2026, 4, 1),
				StartTime:             "19:00",
				EndTime:               "23:00",
				BreakMinutes:          30,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(20),
				IrregularHoursProfile: domain.IrregularHoursProfileRoster,
			},
		},
	}

	service := &PayoutService{repository: repo}
	preview, err := service.PreviewPayroll(context.Background(), domain.PayrollPreviewParams{
		EmployeeID:  repo.employee.ID,
		PeriodStart: dateUTC(2026, 4, 1),
		PeriodEnd:   dateUTC(2026, 4, 30),
	})
	if err != nil {
		t.Fatalf("PreviewPayroll returned error: %v", err)
	}

	if preview.TotalWorkedMinutes != 210 {
		t.Fatalf("expected 210 worked minutes, got %d", preview.TotalWorkedMinutes)
	}
	if preview.BaseGrossAmount != 70 {
		t.Fatalf("expected base gross 70, got %.2f", preview.BaseGrossAmount)
	}
	if preview.IrregularGrossAmount != 21.01 {
		t.Fatalf("expected irregular gross 21.01, got %.2f", preview.IrregularGrossAmount)
	}
	if len(preview.LineItems) != 2 {
		t.Fatalf("expected 2 line items, got %d", len(preview.LineItems))
	}
	if preview.LineItems[0].AppliedRatePercent != 25 ||
		preview.LineItems[1].AppliedRatePercent != 45 {
		t.Fatalf(
			"unexpected rates: %.2f and %.2f",
			preview.LineItems[0].AppliedRatePercent,
			preview.LineItems[1].AppliedRatePercent,
		)
	}
}

func TestPreviewPayrollNonRosterLeavesNineteenToTwentyAtZero(t *testing.T) {
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("33333333-3333-3333-3333-333333333333"),
			FirstName: "Mila",
			LastName:  "de Boer",
		},
		entries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("44444444-4444-4444-4444-444444444444"),
				EmployeeID:            mustUUID("33333333-3333-3333-3333-333333333333"),
				EmployeeName:          "Mila de Boer",
				EntryDate:             dateUTC(2026, 4, 1),
				StartTime:             "19:00",
				EndTime:               "21:00",
				BreakMinutes:          0,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(10),
				IrregularHoursProfile: domain.IrregularHoursProfileNonRoster,
			},
		},
	}

	service := &PayoutService{repository: repo}
	preview, err := service.PreviewPayroll(context.Background(), domain.PayrollPreviewParams{
		EmployeeID:  repo.employee.ID,
		PeriodStart: dateUTC(2026, 4, 1),
		PeriodEnd:   dateUTC(2026, 4, 30),
	})
	if err != nil {
		t.Fatalf("PreviewPayroll returned error: %v", err)
	}

	if len(preview.LineItems) != 2 {
		t.Fatalf("expected 2 line items, got %d", len(preview.LineItems))
	}
	if preview.LineItems[0].AppliedRatePercent != 0 {
		t.Fatalf("expected first segment at 0%%, got %.2f", preview.LineItems[0].AppliedRatePercent)
	}
	if preview.LineItems[1].AppliedRatePercent != 25 {
		t.Fatalf(
			"expected second segment at 25%%, got %.2f",
			preview.LineItems[1].AppliedRatePercent,
		)
	}
	if preview.IrregularGrossAmount != 2.5 {
		t.Fatalf("expected irregular gross 2.50, got %.2f", preview.IrregularGrossAmount)
	}
}

func TestPreviewPayrollSundayOverridesAllWindows(t *testing.T) {
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("55555555-5555-5555-5555-555555555555"),
			FirstName: "Noor",
			LastName:  "Visser",
		},
		entries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("66666666-6666-6666-6666-666666666666"),
				EmployeeID:            mustUUID("55555555-5555-5555-5555-555555555555"),
				EmployeeName:          "Noor Visser",
				EntryDate:             dateUTC(2026, 4, 5),
				StartTime:             "12:00",
				EndTime:               "14:00",
				BreakMinutes:          0,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(18),
				IrregularHoursProfile: domain.IrregularHoursProfileNonRoster,
			},
		},
	}

	service := &PayoutService{repository: repo}
	preview, err := service.PreviewPayroll(context.Background(), domain.PayrollPreviewParams{
		EmployeeID:  repo.employee.ID,
		PeriodStart: dateUTC(2026, 4, 1),
		PeriodEnd:   dateUTC(2026, 4, 30),
	})
	if err != nil {
		t.Fatalf("PreviewPayroll returned error: %v", err)
	}

	if len(preview.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(preview.LineItems))
	}
	if preview.LineItems[0].AppliedRatePercent != 45 {
		t.Fatalf("expected Sunday rate 45%%, got %.2f", preview.LineItems[0].AppliedRatePercent)
	}
	if preview.IrregularGrossAmount != 16.2 {
		t.Fatalf("expected irregular gross 16.20, got %.2f", preview.IrregularGrossAmount)
	}
}

func TestClosePayPeriodCreatesDraftAndAssignsEntries(t *testing.T) {
	txRepo := &fakePayoutTxRepository{
		entries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("77777777-7777-7777-7777-777777777777"),
				EmployeeID:            mustUUID("11111111-1111-1111-1111-111111111111"),
				EmployeeName:          "Sara Jansen",
				EntryDate:             dateUTC(2026, 4, 1),
				StartTime:             "19:00",
				EndTime:               "21:00",
				BreakMinutes:          15,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(10),
				IrregularHoursProfile: domain.IrregularHoursProfileNonRoster,
			},
		},
	}
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("11111111-1111-1111-1111-111111111111"),
			FirstName: "Sara",
			LastName:  "Jansen",
		},
		tx: txRepo,
	}

	service := &PayoutService{repository: repo}
	period, err := service.ClosePayPeriod(
		context.Background(),
		mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		domain.ClosePayPeriodParams{
			EmployeeID:  repo.employee.ID,
			PeriodStart: dateUTC(2026, 4, 1),
			PeriodEnd:   dateUTC(2026, 4, 30),
		},
	)
	if err != nil {
		t.Fatalf("ClosePayPeriod returned error: %v", err)
	}

	if period.Status != domain.PayPeriodStatusDraft {
		t.Fatalf("expected draft status, got %s", period.Status)
	}
	if len(period.LineItems) != 2 {
		t.Fatalf("expected 2 persisted line items, got %d", len(period.LineItems))
	}
	if period.LineItems[0].MinutesWorked != 52.5 || period.LineItems[1].MinutesWorked != 52.5 {
		t.Fatalf(
			"expected paid minutes 52.5/52.5, got %.2f/%.2f",
			period.LineItems[0].MinutesWorked,
			period.LineItems[1].MinutesWorked,
		)
	}
	if len(txRepo.assignedTimeEntryIDs) != 1 ||
		txRepo.assignedTimeEntryIDs[0] != txRepo.entries[0].ID {
		t.Fatalf("expected one assigned time entry id, got %#v", txRepo.assignedTimeEntryIDs)
	}

	var metadata map[string]any
	if err := json.Unmarshal(period.LineItems[0].Metadata, &metadata); err != nil {
		t.Fatalf("expected valid metadata json: %v", err)
	}
	if metadata["break_minutes"] != float64(15) {
		t.Fatalf("expected break_minutes metadata, got %#v", metadata["break_minutes"])
	}
}

func TestClosePayPeriodRejectsDuplicatePeriod(t *testing.T) {
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("11111111-1111-1111-1111-111111111111"),
			FirstName: "Sara",
			LastName:  "Jansen",
		},
		tx: &fakePayoutTxRepository{
			existingPayPeriod: &domain.PayPeriod{
				ID: mustUUID("88888888-8888-8888-8888-888888888888"),
			},
		},
	}

	service := &PayoutService{repository: repo}
	_, err := service.ClosePayPeriod(
		context.Background(),
		mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		domain.ClosePayPeriodParams{
			EmployeeID:  repo.employee.ID,
			PeriodStart: dateUTC(2026, 4, 1),
			PeriodEnd:   dateUTC(2026, 4, 30),
		},
	)
	if !errors.Is(err, domain.ErrPayPeriodAlreadyExists) {
		t.Fatalf("expected ErrPayPeriodAlreadyExists, got %v", err)
	}
}

func TestClosePayPeriodRejectsWhenNoEligibleEntries(t *testing.T) {
	repo := &fakePayoutRepository{
		employee: &domain.EmployeeDetail{
			ID:        mustUUID("11111111-1111-1111-1111-111111111111"),
			FirstName: "Sara",
			LastName:  "Jansen",
		},
		tx: &fakePayoutTxRepository{},
	}

	service := &PayoutService{repository: repo}
	_, err := service.ClosePayPeriod(
		context.Background(),
		mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		domain.ClosePayPeriodParams{
			EmployeeID:  repo.employee.ID,
			PeriodStart: dateUTC(2026, 4, 1),
			PeriodEnd:   dateUTC(2026, 4, 30),
		},
	)
	if !errors.Is(err, domain.ErrPayPeriodNoEntries) {
		t.Fatalf("expected ErrPayPeriodNoEntries, got %v", err)
	}
}

func TestMarkPayPeriodPaidByAdminRequiresDraftState(t *testing.T) {
	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			payPeriodForUpdate: &domain.PayPeriod{
				ID:     mustUUID("99999999-9999-9999-9999-999999999999"),
				Status: domain.PayPeriodStatusPaid,
			},
		},
	}

	service := &PayoutService{repository: repo}
	_, err := service.MarkPayPeriodPaidByAdmin(
		context.Background(),
		mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		mustUUID("99999999-9999-9999-9999-999999999999"),
	)
	if !errors.Is(err, domain.ErrPayPeriodStateInvalid) {
		t.Fatalf("expected ErrPayPeriodStateInvalid, got %v", err)
	}
}

func TestMarkPayPeriodPaidByAdminMarksDraftPaid(t *testing.T) {
	repo := &fakePayoutRepository{
		tx: &fakePayoutTxRepository{
			payPeriodForUpdate: &domain.PayPeriod{
				ID:     mustUUID("99999999-9999-9999-9999-999999999999"),
				Status: domain.PayPeriodStatusDraft,
			},
		},
	}

	service := &PayoutService{repository: repo}
	period, err := service.MarkPayPeriodPaidByAdmin(
		context.Background(),
		mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		mustUUID("99999999-9999-9999-9999-999999999999"),
	)
	if err != nil {
		t.Fatalf("MarkPayPeriodPaidByAdmin returned error: %v", err)
	}
	if period.Status != domain.PayPeriodStatusPaid {
		t.Fatalf("expected paid status, got %s", period.Status)
	}
}

func TestGetPayrollMonthSummaryCurrentMonthUsesLiveTotalsEvenWithLockedSnapshot(t *testing.T) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	employeeID := mustUUID("12121212-1212-1212-1212-121212121212")
	payPeriodID := mustUUID("34343434-3434-3434-3434-343434343434")

	repo := &fakePayoutRepository{
		monthEmployees: []domain.PayrollMonthEmployee{
			{EmployeeID: employeeID, EmployeeName: "Sara Jansen"},
		},
		monthEmployeeTotalCount: 1,
		monthPayPeriods: []domain.PayPeriod{
			{
				ID:                   payPeriodID,
				EmployeeID:           employeeID,
				EmployeeName:         "Sara Jansen",
				PeriodStart:          monthStart,
				PeriodEnd:            monthEnd,
				Status:               domain.PayPeriodStatusDraft,
				BaseGrossAmount:      1,
				IrregularGrossAmount: 1,
				GrossAmount:          2,
			},
		},
		monthLockedMultipliers: []domain.PayrollLockedMultiplierSummary{
			{
				PayPeriodID:   payPeriodID,
				RatePercent:   25,
				WorkedMinutes: 60,
				PaidMinutes:   60,
				BaseAmount:    10,
				PremiumAmount: 2.5,
			},
		},
		monthApprovedEntries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("56565656-5656-5656-5656-565656565656"),
				EmployeeID:            employeeID,
				EmployeeName:          "Sara Jansen",
				EntryDate:             monthStart.AddDate(0, 0, 1),
				StartTime:             "20:00",
				EndTime:               "22:00",
				BreakMinutes:          0,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(10),
				IrregularHoursProfile: domain.IrregularHoursProfileNonRoster,
			},
		},
		monthPendingSummaries: []domain.PayrollMonthPendingSummary{
			{EmployeeID: employeeID, PendingEntryCount: 1, PendingWorkedMinutes: 30},
		},
	}

	service := &PayoutService{repository: repo}
	page, err := service.GetPayrollMonthSummary(
		context.Background(),
		domain.PayrollMonthSummaryParams{
			Month: monthStart,
			Limit: 20,
		},
	)
	if err != nil {
		t.Fatalf("GetPayrollMonthSummary returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 row, got %d", len(page.Items))
	}

	row := page.Items[0]
	if row.DataSource != "live" {
		t.Fatalf("expected live source, got %s", row.DataSource)
	}
	if row.IsLocked {
		t.Fatalf("expected current month row not to be marked locked")
	}
	if !row.HasLockedSnapshot {
		t.Fatalf("expected locked snapshot indicator")
	}
	if row.GrossAmount != 25 {
		t.Fatalf("expected live gross 25.00, got %.2f", row.GrossAmount)
	}
	if row.PendingEntryCount != 1 || row.PendingWorkedMinutes != 30 {
		t.Fatalf(
			"unexpected pending values: count=%d minutes=%d",
			row.PendingEntryCount,
			row.PendingWorkedMinutes,
		)
	}
	if len(row.MultiplierSummaries) != 1 || row.MultiplierSummaries[0].RatePercent != 25 {
		t.Fatalf("expected one 25%% multiplier summary, got %#v", row.MultiplierSummaries)
	}
}

func TestGetPayrollMonthSummaryPastMonthUsesLockedSnapshot(t *testing.T) {
	monthStart := dateUTC(2026, 3, 1)
	monthEnd := monthStart.AddDate(0, 1, -1)
	employeeID := mustUUID("78787878-7878-7878-7878-787878787878")
	payPeriodID := mustUUID("90909090-9090-9090-9090-909090909090")

	repo := &fakePayoutRepository{
		monthEmployees: []domain.PayrollMonthEmployee{
			{EmployeeID: employeeID, EmployeeName: "Noor Visser"},
		},
		monthEmployeeTotalCount: 1,
		monthPayPeriods: []domain.PayPeriod{
			{
				ID:                   payPeriodID,
				EmployeeID:           employeeID,
				EmployeeName:         "Noor Visser",
				PeriodStart:          monthStart,
				PeriodEnd:            monthEnd,
				Status:               domain.PayPeriodStatusPaid,
				BaseGrossAmount:      100,
				IrregularGrossAmount: 30,
				GrossAmount:          130,
			},
		},
		monthLockedMultipliers: []domain.PayrollLockedMultiplierSummary{
			{
				PayPeriodID:   payPeriodID,
				RatePercent:   25,
				WorkedMinutes: 120,
				PaidMinutes:   120,
				BaseAmount:    50,
				PremiumAmount: 12.5,
			},
			{
				PayPeriodID:   payPeriodID,
				RatePercent:   45,
				WorkedMinutes: 90,
				PaidMinutes:   90,
				BaseAmount:    50,
				PremiumAmount: 17.5,
			},
		},
		monthApprovedEntries: []domain.PayrollPreviewTimeEntry{
			{
				ID:                    mustUUID("78780000-7878-7878-7878-787878787878"),
				EmployeeID:            employeeID,
				EmployeeName:          "Noor Visser",
				EntryDate:             monthStart.AddDate(0, 0, 2),
				StartTime:             "20:00",
				EndTime:               "22:00",
				BreakMinutes:          0,
				HourType:              domain.TimeEntryHourTypeNormal,
				ContractType:          "loondienst",
				ContractRate:          ptrFloat(1),
				IrregularHoursProfile: domain.IrregularHoursProfileNonRoster,
			},
		},
	}

	service := &PayoutService{repository: repo}
	page, err := service.GetPayrollMonthSummary(
		context.Background(),
		domain.PayrollMonthSummaryParams{
			Month: monthStart,
			Limit: 20,
		},
	)
	if err != nil {
		t.Fatalf("GetPayrollMonthSummary returned error: %v", err)
	}

	row := page.Items[0]
	if row.DataSource != "locked" || !row.IsLocked {
		t.Fatalf("expected locked row, got source=%s locked=%v", row.DataSource, row.IsLocked)
	}
	if row.GrossAmount != 130 {
		t.Fatalf("expected locked gross 130, got %.2f", row.GrossAmount)
	}
	if row.WorkedMinutes != 210 {
		t.Fatalf("expected worked minutes 210, got %d", row.WorkedMinutes)
	}
	if len(row.MultiplierSummaries) != 2 {
		t.Fatalf("expected 2 multiplier buckets, got %d", len(row.MultiplierSummaries))
	}
}

func TestGetPayrollMonthSummaryIncludesPendingOnlyEmployee(t *testing.T) {
	monthStart := dateUTC(2026, 2, 1)
	employeeID := mustUUID("abab1212-1212-1212-1212-121212121212")

	repo := &fakePayoutRepository{
		monthEmployees: []domain.PayrollMonthEmployee{
			{EmployeeID: employeeID, EmployeeName: "Mila de Boer"},
		},
		monthEmployeeTotalCount: 1,
		monthPendingSummaries: []domain.PayrollMonthPendingSummary{
			{EmployeeID: employeeID, PendingEntryCount: 2, PendingWorkedMinutes: 180},
		},
	}

	service := &PayoutService{repository: repo}
	page, err := service.GetPayrollMonthSummary(
		context.Background(),
		domain.PayrollMonthSummaryParams{
			Month: monthStart,
			Limit: 20,
		},
	)
	if err != nil {
		t.Fatalf("GetPayrollMonthSummary returned error: %v", err)
	}

	row := page.Items[0]
	if row.GrossAmount != 0 || row.PendingEntryCount != 2 || row.PendingWorkedMinutes != 180 {
		t.Fatalf("unexpected pending-only row: %#v", row)
	}
	if row.DataSource != "live" {
		t.Fatalf("expected live source for pending-only row, got %s", row.DataSource)
	}
}

type fakePayoutRepository struct {
	employee                *domain.EmployeeDetail
	entries                 []domain.PayrollPreviewTimeEntry
	holidays                []domain.NationalHoliday
	tx                      *fakePayoutTxRepository
	monthEmployees          []domain.PayrollMonthEmployee
	monthEmployeeTotalCount int64
	monthPayPeriods         []domain.PayPeriod
	monthLockedMultipliers  []domain.PayrollLockedMultiplierSummary
	monthApprovedEntries    []domain.PayrollPreviewTimeEntry
	monthPendingSummaries   []domain.PayrollMonthPendingSummary
}

func (f *fakePayoutRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.PayoutTxRepository) error,
) error {
	if f.tx == nil {
		panic("unexpected call")
	}
	return fn(f.tx)
}

func (f *fakePayoutRepository) ListMyPayoutRequests(
	_ context.Context,
	_ domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListPayoutRequests(
	_ context.Context,
	_ domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) GetPayrollPreviewEmployee(
	_ context.Context,
	_ uuid.UUID,
) (*domain.EmployeeDetail, error) {
	return f.employee, nil
}

func (f *fakePayoutRepository) ListPayrollPreviewTimeEntries(
	_ context.Context,
	_ domain.PayrollPreviewParams,
) ([]domain.PayrollPreviewTimeEntry, error) {
	return f.entries, nil
}

func (f *fakePayoutRepository) ListNationalHolidays(
	_ context.Context,
	_ string,
	_, _ time.Time,
) ([]domain.NationalHoliday, error) {
	return f.holidays, nil
}

func (f *fakePayoutRepository) GetPayPeriodByID(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListPayPeriods(
	_ context.Context,
	_ domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListPayPeriodLineItems(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.PayPeriodLineItem, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListPayrollMonthEmployees(
	_ context.Context,
	_ domain.PayrollMonthSummaryParams,
	_, _ time.Time,
) ([]domain.PayrollMonthEmployee, int64, error) {
	return f.monthEmployees, f.monthEmployeeTotalCount, nil
}

func (f *fakePayoutRepository) ListPayPeriodsByEmployeesAndRange(
	_ context.Context,
	_ []uuid.UUID,
	_, _ time.Time,
) ([]domain.PayPeriod, error) {
	return f.monthPayPeriods, nil
}

func (f *fakePayoutRepository) ListPayrollMonthLockedMultiplierSummaries(
	_ context.Context,
	_ []uuid.UUID,
) ([]domain.PayrollLockedMultiplierSummary, error) {
	return f.monthLockedMultipliers, nil
}

func (f *fakePayoutRepository) ListPayrollMonthApprovedTimeEntries(
	_ context.Context,
	_ []uuid.UUID,
	_, _ time.Time,
) ([]domain.PayrollPreviewTimeEntry, error) {
	return f.monthApprovedEntries, nil
}

func (f *fakePayoutRepository) ListPayrollMonthPendingSummaries(
	_ context.Context,
	_ []uuid.UUID,
	_, _ time.Time,
) ([]domain.PayrollMonthPendingSummary, error) {
	return f.monthPendingSummaries, nil
}

type fakePayoutTxRepository struct {
	existingPayPeriod    *domain.PayPeriod
	entries              []domain.PayrollPreviewTimeEntry
	createdPayPeriod     *domain.PayPeriod
	createdLineItems     []domain.PayPeriodLineItem
	assignedTimeEntryIDs []uuid.UUID
	payPeriodForUpdate   *domain.PayPeriod
}

func (f *fakePayoutTxRepository) GetEmployeePayoutContract(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayoutContract, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) EnsureLeaveBalanceForYear(
	_ context.Context,
	_ uuid.UUID,
	_ int32,
) error {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) GetPayoutBalanceForUpdate(
	_ context.Context,
	_ uuid.UUID,
	_ int32,
) (*domain.PayoutBalanceSnapshot, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) CreatePayoutRequest(
	_ context.Context,
	_ domain.CreatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) GetPayoutRequestForUpdate(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) ApprovePayoutRequest(
	_ context.Context,
	_, _ uuid.UUID,
	_ time.Time,
	_ *string,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) RejectPayoutRequest(
	_ context.Context,
	_, _ uuid.UUID,
	_ *string,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) MarkPayoutRequestPaid(
	_ context.Context,
	_, _ uuid.UUID,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) ApplyLeaveBalanceDeduction(
	_ context.Context,
	_ uuid.UUID,
	_, _ int32,
) (*domain.LeaveBalance, error) {
	panic("unexpected call")
}

func (f *fakePayoutTxRepository) GetPayPeriodByEmployeePeriod(
	_ context.Context,
	_ uuid.UUID,
	_, _ time.Time,
) (*domain.PayPeriod, error) {
	if f.existingPayPeriod == nil {
		return nil, domain.ErrPayPeriodNotFound
	}
	return f.existingPayPeriod, nil
}

func (f *fakePayoutTxRepository) LockPayrollPreviewTimeEntries(
	_ context.Context,
	_ domain.PayrollPreviewParams,
) ([]domain.PayrollPreviewTimeEntry, error) {
	return f.entries, nil
}

func (f *fakePayoutTxRepository) CreatePayPeriod(
	_ context.Context,
	params domain.ClosePayPeriodParams,
	createdByEmployeeID uuid.UUID,
	preview domain.PayrollPreview,
) (*domain.PayPeriod, error) {
	f.createdPayPeriod = &domain.PayPeriod{
		ID:                   mustUUID("abababab-abab-abab-abab-abababababab"),
		EmployeeID:           params.EmployeeID,
		EmployeeName:         preview.EmployeeName,
		PeriodStart:          params.PeriodStart,
		PeriodEnd:            params.PeriodEnd,
		Status:               domain.PayPeriodStatusDraft,
		BaseGrossAmount:      preview.BaseGrossAmount,
		IrregularGrossAmount: preview.IrregularGrossAmount,
		GrossAmount:          preview.GrossAmount,
		CreatedByEmployeeID:  &createdByEmployeeID,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	return f.createdPayPeriod, nil
}

func (f *fakePayoutTxRepository) CreatePayPeriodLineItem(
	_ context.Context,
	payPeriodID uuid.UUID,
	item domain.PayPeriodLineItem,
) (*domain.PayPeriodLineItem, error) {
	item.ID = uuid.New()
	item.PayPeriodID = payPeriodID
	item.CreatedAt = time.Now().UTC()
	item.UpdatedAt = item.CreatedAt
	f.createdLineItems = append(f.createdLineItems, item)
	return &item, nil
}

func (f *fakePayoutTxRepository) AssignTimeEntriesToPayPeriod(
	_ context.Context,
	_ uuid.UUID,
	timeEntryIDs []uuid.UUID,
) error {
	f.assignedTimeEntryIDs = append([]uuid.UUID(nil), timeEntryIDs...)
	return nil
}

func (f *fakePayoutTxRepository) GetPayPeriodForUpdate(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayPeriod, error) {
	if f.payPeriodForUpdate == nil {
		return nil, domain.ErrPayPeriodNotFound
	}
	return f.payPeriodForUpdate, nil
}

func (f *fakePayoutTxRepository) MarkPayPeriodPaid(
	_ context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	paidAt := time.Now().UTC()
	return &domain.PayPeriod{
		ID:     payPeriodID,
		Status: domain.PayPeriodStatusPaid,
		PaidAt: &paidAt,
	}, nil
}

func ptrFloat(v float64) *float64 {
	return &v
}

func mustUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return id
}

func dateUTC(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
