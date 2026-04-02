package service

import (
	"context"
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
	if preview.LineItems[0].AppliedRatePercent != 25 || preview.LineItems[1].AppliedRatePercent != 45 {
		t.Fatalf("unexpected rates: %.2f and %.2f", preview.LineItems[0].AppliedRatePercent, preview.LineItems[1].AppliedRatePercent)
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
		t.Fatalf("expected second segment at 25%%, got %.2f", preview.LineItems[1].AppliedRatePercent)
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

type fakePayoutRepository struct {
	employee *domain.EmployeeDetail
	entries  []domain.PayrollPreviewTimeEntry
	holidays []domain.NationalHoliday
}

func (f *fakePayoutRepository) WithTx(_ context.Context, _ func(tx domain.PayoutTxRepository) error) error {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListMyPayoutRequests(_ context.Context, _ domain.ListMyPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) ListPayoutRequests(_ context.Context, _ domain.ListPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutRepository) GetPayrollPreviewEmployee(_ context.Context, _ uuid.UUID) (*domain.EmployeeDetail, error) {
	return f.employee, nil
}

func (f *fakePayoutRepository) ListPayrollPreviewTimeEntries(_ context.Context, _ domain.PayrollPreviewParams) ([]domain.PayrollPreviewTimeEntry, error) {
	return f.entries, nil
}

func (f *fakePayoutRepository) ListNationalHolidays(_ context.Context, _ string, _, _ time.Time) ([]domain.NationalHoliday, error) {
	return f.holidays, nil
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
