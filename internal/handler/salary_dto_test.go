package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPreviewPayrollRequestBindsUUIDQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/payroll/preview?employee_id=a5514673-7217-476b-bbe3-07db2a725e12&period_start=2026-04-01&period_end=2026-04-30",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	var got previewPayrollRequest
	if err := ctx.ShouldBindQuery(&got); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.EmployeeID.String() != "a5514673-7217-476b-bbe3-07db2a725e12" {
		t.Fatalf("unexpected employee id: %s", got.EmployeeID.String())
	}
}

func TestToSalaryPageResponseSeparatesLeavePayoutFromFixedBase(t *testing.T) {
	employeeID := uuid.New()
	payoutID := uuid.New()
	periodStart := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	nextPeriodStart := periodStart.AddDate(0, 0, 28)
	rate := 25.0
	hoursPerWeek := 32.0

	response := toSalaryPageResponse(&domain.SalaryPageData{
		EmployeeID:    employeeID,
		EmployeeName:  "Jane Doe",
		Month:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		ContractType:  "permanent",
		ContractRate:  &rate,
		ContractHours: &hoursPerWeek,
		DataSource:    "live",
		Preview: &domain.PayrollPreview{
			EmployeeID:         employeeID,
			EmployeeName:       "Jane Doe",
			PeriodStart:        periodStart,
			PeriodEnd:          periodEnd,
			BaseGrossAmount:    4200,
			GrossAmount:        4200,
			TotalWorkedMinutes: 0,
			LineItems: []domain.PayrollPreviewLineItem{
				{
					SourceType:            "fixed_base",
					Label:                 "Fixed 4-week base salary",
					ContractType:          "permanent",
					WorkDate:              periodStart,
					IrregularHoursProfile: "none",
					PaidMinutes:           7680,
					BaseAmount:            3200,
				},
				{
					LeavePayoutRequestID:  &payoutID,
					SourceType:            domain.PayrollSourceLeavePayout,
					Label:                 "Leave payout",
					ContractType:          "permanent",
					WorkDate:              periodStart,
					IrregularHoursProfile: "none",
					PaidMinutes:           2400,
					BaseAmount:            1000,
				},
			},
		},
		LeavePayoutRequests: []domain.PayoutRequest{
			{
				ID:             payoutID,
				EmployeeID:     employeeID,
				RequestedHours: 40,
				HourlyRate:     25,
				GrossAmount:    1000,
				PayPeriodStart: &periodStart,
				Status:         domain.PayoutRequestStatusApproved,
			},
			{
				ID:             uuid.New(),
				EmployeeID:     employeeID,
				RequestedHours: 20,
				HourlyRate:     25,
				GrossAmount:    500,
				PayPeriodStart: &nextPeriodStart,
				Status:         domain.PayoutRequestStatusApproved,
			},
			{
				ID:             uuid.New(),
				EmployeeID:     employeeID,
				RequestedHours: 12,
				HourlyRate:     25,
				GrossAmount:    300,
				PayPeriodStart: &periodStart,
				Status:         domain.PayoutRequestStatusPending,
			},
		},
	})

	if response.Summary.WorkedMinutes != 0 {
		t.Fatalf("worked minutes = %v, want 0", response.Summary.WorkedMinutes)
	}
	if response.Summary.PaidMinutes != 7680 {
		t.Fatalf("paid minutes = %v, want 7680", response.Summary.PaidMinutes)
	}
	if response.Summary.PaidHours != 128 {
		t.Fatalf("paid hours = %v, want 128", response.Summary.PaidHours)
	}
	if response.Summary.BaseGrossAmount != 3200 {
		t.Fatalf("base gross = %v, want 3200", response.Summary.BaseGrossAmount)
	}
	if response.Summary.LeavePayoutGrossAmount != 1000 {
		t.Fatalf(
			"leave payout gross = %v, want 1000",
			response.Summary.LeavePayoutGrossAmount,
		)
	}
	if response.Summary.GrossAmount != 4200 {
		t.Fatalf("gross = %v, want 4200", response.Summary.GrossAmount)
	}
	if response.BaseEarnings.Amount == nil || *response.BaseEarnings.Amount != 3200 {
		t.Fatalf("base earnings amount = %v, want 3200", response.BaseEarnings.Amount)
	}
}
