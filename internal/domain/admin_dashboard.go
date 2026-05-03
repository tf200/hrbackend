package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AvailabilityStats is the domain struct for the employee availability breakdown.
type AvailabilityStats struct {
	Active       int64
	OnLeave      int64
	Sick         int64
	OutOfService int64
	Unscheduled  int64
}

type OpenShiftCoverageItem struct {
	LocationID          uuid.UUID
	LocationName        string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	ShiftID             uuid.UUID
	ShiftName           string
	ShiftDate           time.Time
	OpenSlots           int64
}

type CriticalActionStats struct {
	PendingLeaveRequests int64
	PendingShiftSwaps    int64
	PendingExpenseClaims int64
	PendingTimeEntries   int64
}

type PayrollTotalStats struct {
	MonthStart  time.Time
	SalaryTotal float64
	ZZPTotal    float64
	ORTTotal    float64
}

type RiskRadarStats struct {
	MonthStart               time.Time
	ContractsEndingThisMonth int64
	OverdueTraining          int64
	LateArrivalsThisMonth    int64
}

type TeamHealthByDepartmentItem struct {
	DepartmentID    uuid.UUID
	DepartmentName  string
	StaffingPercent float64
	AbsencePercent  float64
	TrainingPercent float64
	Score           float64
	Risk            string
}

type AdminDashboardRepository interface {
	GetAvailabilityStats(ctx context.Context) (*AvailabilityStats, error)
	ListOpenShiftCoverage(ctx context.Context, days int32) ([]OpenShiftCoverageItem, error)
	GetCriticalActionStats(ctx context.Context) (*CriticalActionStats, error)
	GetPayrollTotalStats(ctx context.Context) (*PayrollTotalStats, error)
	GetRiskRadarStats(ctx context.Context) (*RiskRadarStats, error)
	ListTeamHealthByDepartment(ctx context.Context) ([]TeamHealthByDepartmentItem, error)
}

type AdminDashboardService interface {
	GetAvailabilityStats(ctx context.Context) (*AvailabilityStats, error)
	ListOpenShiftCoverage(ctx context.Context, days int32) ([]OpenShiftCoverageItem, error)
	GetCriticalActionStats(ctx context.Context) (*CriticalActionStats, error)
	GetPayrollTotalStats(ctx context.Context) (*PayrollTotalStats, error)
	GetRiskRadarStats(ctx context.Context) (*RiskRadarStats, error)
	ListTeamHealthByDepartment(ctx context.Context) ([]TeamHealthByDepartmentItem, error)
}
