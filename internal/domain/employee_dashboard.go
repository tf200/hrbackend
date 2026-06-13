package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrEmployeeDashboardInvalidRequest = errors.New("invalid employee dashboard request")

type EmployeeDashboardLeaveBalanceKPI struct {
	Year         int32
	UsedMinutes  int32
	TotalMinutes int32
}

type EmployeeDashboardPayKPI struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	GrossAmount float64
	DataSource  string
}

type EmployeeDashboardKPI struct {
	LeaveBalance         EmployeeDashboardLeaveBalanceKPI
	PendingLeaveRequests int64
	EstimatedCurrentPay  EmployeeDashboardPayKPI
	PendingSignatures    int64
}

type EmployeeDashboardPendingRequest struct {
	ID              uuid.UUID
	RequestType     string
	Status          string
	SubmittedAt     time.Time
	RequestDate     time.Time
	Title           string
	Description     *string
	DurationMinutes *int32
	Amount          *float64
	Currency        *string
}

type GetEmployeeDashboardKPIsParams struct {
	EmployeeID uuid.UUID
	Year       int32
}

type ListEmployeeDashboardPendingRequestsParams struct {
	EmployeeID uuid.UUID
	Since      time.Time
	RecentDays int32
	Limit      int32
}

type EmployeeDashboardRepositoryKPI struct {
	LeaveBalance         EmployeeDashboardLeaveBalanceKPI
	PendingLeaveRequests int64
	PendingSignatures    int64
}

type EmployeeDashboardService interface {
	GetKPIs(ctx context.Context, employeeID uuid.UUID) (*EmployeeDashboardKPI, error)
	ListPendingRequests(
		ctx context.Context,
		params ListEmployeeDashboardPendingRequestsParams,
	) ([]EmployeeDashboardPendingRequest, error)
}

type EmployeeDashboardRepository interface {
	GetKPIs(
		ctx context.Context,
		params GetEmployeeDashboardKPIsParams,
	) (*EmployeeDashboardRepositoryKPI, error)
	ListPendingRequests(
		ctx context.Context,
		params ListEmployeeDashboardPendingRequestsParams,
	) ([]EmployeeDashboardPendingRequest, error)
}
