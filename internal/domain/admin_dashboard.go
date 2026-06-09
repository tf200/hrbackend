package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAdminDashboardInvalidRequest = errors.New("invalid admin dashboard request")

type AdminDashboardKPI struct {
	TotalEmployees   int64
	EmployeesPresent int64
	TotalDocuments   int64
	ProcessingDocs   int64
}

type RecentDashboardEmployee struct {
	ID                 uuid.UUID
	FirstName          string
	LastName           string
	OrganizationalRole *string
	DepartmentName     *string
	LocationName       *string
	CreatedAt          time.Time
}

type RecentDashboardEmployeePage struct {
	Items      []RecentDashboardEmployee
	TotalCount int64
}

type ListRecentEmployeesParams struct {
	Limit  int32
	Offset int32
}

type FullTimeEmployeeDeptBreakdownItem struct {
	DepartmentID   uuid.UUID
	DepartmentName string
	TotalEmployees int64
}

type FullTimeEmployeeLocBreakdownItem struct {
	LocationID     uuid.UUID
	LocationName   string
	TotalEmployees int64
}

type FullTimeEmployeeBreakdowns struct {
	ByDepartment []FullTimeEmployeeDeptBreakdownItem
	ByLocation   []FullTimeEmployeeLocBreakdownItem
}

type LeaveAbsenceTrendPoint struct {
	Month        time.Time
	EmployeesOut int64
}

type LeaveAbsenceTrends struct {
	View   string
	Year   *int32
	Points []LeaveAbsenceTrendPoint
}

type GetLeaveAbsenceTrendsParams struct {
	View string
	Year *int32
}

type ListLeaveAbsenceTrendPointsParams struct {
	FromDate time.Time
	ToDate   time.Time
}

type UpcomingDashboardAlerts struct {
	EndingContracts     []EndingContractAlert
	ExpiringCredentials []ExpiringCredentialAlert
	ReturningFromLeave  []ReturningFromLeaveAlert
}

type EndingContractAlert struct {
	EmployeeID      uuid.UUID
	EmployeeName    string
	ContractID      uuid.UUID
	ContractType    string
	ContractEndDate time.Time
	DaysRemaining   int32
	Department      string
	Location        string
}

type ExpiringCredentialAlert struct {
	EmployeeID     uuid.UUID
	EmployeeName   string
	CredentialID   uuid.UUID
	CredentialType string
	Name           string
	ExpiryDate     time.Time
	DaysRemaining  int32
}

type ReturningFromLeaveAlert struct {
	EmployeeID      uuid.UUID
	EmployeeName    string
	LeaveRequestID  uuid.UUID
	LeaveType       string
	LeaveEndDate    time.Time
	ReturnDate      time.Time
	DaysUntilReturn int32
}

type GetUpcomingDashboardAlertsParams struct {
	Days  int32
	Limit int32
}

type ListUpcomingDashboardAlertsParams struct {
	ToDate time.Time
	Limit  int32
}

type AdminDashboardService interface {
	GetKPIs(ctx context.Context) (*AdminDashboardKPI, error)
	ListRecentEmployees(
		ctx context.Context,
		params ListRecentEmployeesParams,
	) (*RecentDashboardEmployeePage, error)
	GetFullTimeEmployeeBreakdowns(ctx context.Context) (*FullTimeEmployeeBreakdowns, error)
	GetLeaveAbsenceTrends(
		ctx context.Context,
		params GetLeaveAbsenceTrendsParams,
	) (*LeaveAbsenceTrends, error)
	GetUpcomingDashboardAlerts(
		ctx context.Context,
		params GetUpcomingDashboardAlertsParams,
	) (*UpcomingDashboardAlerts, error)
}

type AdminDashboardRepository interface {
	GetKPIs(ctx context.Context) (*AdminDashboardKPI, error)
	ListRecentEmployees(
		ctx context.Context,
		params ListRecentEmployeesParams,
	) (*RecentDashboardEmployeePage, error)
	GetFullTimeEmployeeBreakdowns(ctx context.Context) (*FullTimeEmployeeBreakdowns, error)
	ListLeaveAbsenceTrendPoints(
		ctx context.Context,
		params ListLeaveAbsenceTrendPointsParams,
	) ([]LeaveAbsenceTrendPoint, error)
	ListEndingContractAlerts(
		ctx context.Context,
		params ListUpcomingDashboardAlertsParams,
	) ([]EndingContractAlert, error)
	ListExpiringCredentialAlerts(
		ctx context.Context,
		params ListUpcomingDashboardAlertsParams,
	) ([]ExpiringCredentialAlert, error)
	ListReturningFromLeaveAlerts(
		ctx context.Context,
		params ListUpcomingDashboardAlertsParams,
	) ([]ReturningFromLeaveAlert, error)
}
