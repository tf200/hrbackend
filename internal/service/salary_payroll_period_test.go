package service

import (
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestBuildFixedPayrollPeriodContractSegmentsByEmployee(t *testing.T) {
	periodStart := time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 0, 27)
	employeeID := uuid.New()

	tests := []struct {
		name            string
		hoursPerWeek    float64
		activeFrom      time.Time
		activeUntil     time.Time
		wantBaseAmount  float64
		wantPaidMinutes float64
		wantProration   float64
	}{
		{
			name:            "36 hour full period",
			hoursPerWeek:    36,
			activeFrom:      periodStart,
			activeUntil:     periodEnd,
			wantBaseAmount:  3600,
			wantPaidMinutes: 8640,
			wantProration:   1,
		},
		{
			name:            "38 hour full period",
			hoursPerWeek:    38,
			activeFrom:      periodStart,
			activeUntil:     periodEnd,
			wantBaseAmount:  3800,
			wantPaidMinutes: 9120,
			wantProration:   1,
		},
		{
			name:            "40 hour full period",
			hoursPerWeek:    40,
			activeFrom:      periodStart,
			activeUntil:     periodEnd,
			wantBaseAmount:  4000,
			wantPaidMinutes: 9600,
			wantProration:   1,
		},
		{
			name:            "40 hour half period",
			hoursPerWeek:    40,
			activeFrom:      periodStart,
			activeUntil:     periodStart.AddDate(0, 0, 13),
			wantBaseAmount:  2000,
			wantPaidMinutes: 4800,
			wantProration:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segmentsByEmployee := buildFixedPayrollPeriodContractSegmentsByEmployee(
				[]domain.FixedPayrollContractSegmentSource{
					{
						EmployeeID:           employeeID,
						ContractID:           uuid.New(),
						ContractType:         "permanent",
						ActiveFrom:           tt.activeFrom,
						ActiveUntil:          tt.activeUntil,
						HoursPerWeek:         tt.hoursPerWeek,
						FullTimeHoursPerWeek: 36,
						MonthlySalary:        3900,
						HourlyRate:           25,
					},
				},
				periodStart,
				periodEnd,
			)

			segments := segmentsByEmployee[employeeID]
			if len(segments) != 1 {
				t.Fatalf("expected 1 segment, got %d", len(segments))
			}

			segment := segments[0]
			if segment.BaseAmount != tt.wantBaseAmount {
				t.Fatalf("base amount = %v, want %v", segment.BaseAmount, tt.wantBaseAmount)
			}
			if segment.ProrationRatio != tt.wantProration {
				t.Fatalf("proration ratio = %v, want %v", segment.ProrationRatio, tt.wantProration)
			}
			paidMinutes := contractSegmentPeriodPaidMinutes(segment, periodStart, periodEnd)
			if paidMinutes != tt.wantPaidMinutes {
				t.Fatalf("paid minutes = %v, want %v", paidMinutes, tt.wantPaidMinutes)
			}
		})
	}
}

func TestBuildFixedPayrollPeriodContractSegmentsByEmployeeFiveWeekPeriod(t *testing.T) {
	periodStart := time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2027, time.January, 3, 0, 0, 0, 0, time.UTC)
	employeeID := uuid.New()

	segmentsByEmployee := buildFixedPayrollPeriodContractSegmentsByEmployee(
		[]domain.FixedPayrollContractSegmentSource{
			{
				EmployeeID:           employeeID,
				ContractID:           uuid.New(),
				ContractType:         "permanent",
				ActiveFrom:           periodStart,
				ActiveUntil:          periodEnd,
				HoursPerWeek:         40,
				FullTimeHoursPerWeek: 36,
				MonthlySalary:        3900,
				HourlyRate:           25,
			},
		},
		periodStart,
		periodEnd,
	)

	segments := segmentsByEmployee[employeeID]
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	segment := segments[0]
	if segment.BaseAmount != 5000 {
		t.Fatalf("base amount = %v, want 5000", segment.BaseAmount)
	}
	paidMinutes := contractSegmentPeriodPaidMinutes(segment, periodStart, periodEnd)
	if paidMinutes != 12000 {
		t.Fatalf("paid minutes = %v, want 12000", paidMinutes)
	}
}
