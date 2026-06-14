package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLeaveRequestInvalidRequest = errors.New("invalid leave request")
	ErrLeaveRequestNotFound       = errors.New("leave request not found")
	ErrLeaveRequestStateInvalid   = errors.New("leave request is not in an editable state")
	ErrLeaveRequestForbidden      = errors.New("leave request is not accessible by the actor")
	ErrLeaveBalanceInsufficient   = errors.New("insufficient leave balance")
	ErrLeaveBalanceInvalidAdjust  = errors.New("invalid leave balance adjustment")
	ErrLeaveDurationInvalid       = errors.New("invalid leave duration")
)

type LeaveRequest struct {
	ID                  uuid.UUID
	EmployeeID          uuid.UUID
	CreatedByEmployeeID *uuid.UUID
	LeaveType           string
	Status              string
	DurationType        string
	RequestedMinutes    int32
	StartTime           *time.Time
	EndTime             *time.Time
	StartDate           time.Time
	EndDate             time.Time
	Reason              *string
	DecisionNote        *string
	DecidedByEmployeeID *uuid.UUID
	RequestedAt         time.Time
	DecidedAt           *time.Time
	CancelledAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type LeaveRequestListItem struct {
	LeaveRequest
	EmployeeName string
}

type LeaveCalendarRecord struct {
	LeaveRequestID   uuid.UUID
	LeaveType        string
	Status           string
	DurationType     string
	RequestedMinutes int32
	StartDate        time.Time
	EndDate          time.Time
	Reason           *string
}

type LeaveCalendarEmployee struct {
	EmployeeID     uuid.UUID
	EmployeeName   string
	DepartmentName *string
	LeaveRecords   []LeaveCalendarRecord
}

type LeaveRequestPage struct {
	Items      []LeaveRequestListItem
	TotalCount int64
}

type LeaveRequestStats struct {
	OpenRequests     int64
	ApprovedRequests int64
	RejectedRequests int64
	SicknessAbsence  int64
}

type DeductedLeaveSummary struct {
	ID              uuid.UUID
	LeaveType       string
	StartDate       time.Time
	EndDate         time.Time
	DurationMinutes int32
}

type DeductedPayoutSummary struct {
	ID              uuid.UUID
	RequestedHours  int32
	DurationMinutes int32
	BalanceYear     int32
	PayPeriodStart  *time.Time
	Status          string
	RequestedAt     time.Time
	DecidedAt       *time.Time
	PaidAt          *time.Time
}

type LeaveBalance struct {
	EmployeeID            uuid.UUID
	EmployeeName          string
	Year                  int32
	LegalTotalMinutes     int32
	LegalUsedMinutes      int32
	LegalRemainingMinutes int32
	TotalRemainingMinutes int32
	ContractHours         *float64
	ContractType          *string
	ContractStartDate     *time.Time
	ContractEndDate       *time.Time
	EffectiveEndDate      *time.Time
	DeductedLeaves        []DeductedLeaveSummary
	DeductedPayouts       []DeductedPayoutSummary
}

type LeaveBalancePage struct {
	Items      []LeaveBalance
	TotalCount int64
}

type LeaveBalanceDetails struct {
	Balance          LeaveBalance
	ContractAccruals []LeaveContractAccrual
}

type LeaveContractAccrual struct {
	ContractID        uuid.UUID
	ContractType      string
	ContractHours     *float64
	ContractStartDate time.Time
	ContractEndDate   *time.Time
	EffectiveEndDate  *time.Time
	SegmentStartDate  time.Time
	SegmentEndDate    time.Time
	YearDays          int32
	SegmentDays       int32
	FullYearMinutes   int32
	ScheduleMinutes   int32
	OvertimeMinutes   int32
	GainedMinutes     int32
	DeductedMinutes   int32
	RemainingMinutes  int32
}

type LeavePolicy struct {
	LeaveType      string
	DeductsBalance bool
}

type CreateLeaveRequestParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	LeaveType           string
	DurationType        string
	RequestedMinutes    int32
	StartDate           time.Time
	EndDate             time.Time
	StartTime           *time.Time
	EndTime             *time.Time
	Reason              *string
}

type UpdateLeaveRequestParams struct {
	LeaveType        *string
	DurationType     *string
	StartDate        *time.Time
	EndDate          *time.Time
	StartTime        *time.Time
	EndTime          *time.Time
	Reason           *string
	RequestedMinutes *int32
}

type DecideLeaveRequestParams struct {
	Decision     string
	DecisionNote *string
}

type ListMyLeaveRequestsParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Status     *string
}

type ListLeaveRequestsParams struct {
	Limit          int32
	Offset         int32
	Status         *string
	EmployeeSearch *string
}

type ListLeaveCalendarParams struct {
	Month          time.Time
	DepartmentID   *uuid.UUID
	LeaveTypes     []string
	EmployeeSearch *string
}

type ListLeaveBalancesParams struct {
	Limit          int32
	Offset         int32
	EmployeeSearch *string
	Year           *int32
}

type GetLeaveBalanceDetailsParams struct {
	EmployeeID uuid.UUID
	Year       int32
}

type LeaveContractAtDate struct {
	EmployeeID    uuid.UUID
	RosterFreeDay string
}

type LeaveTxRepository interface {
	GetLeaveRequestForUpdate(ctx context.Context, leaveRequestID uuid.UUID) (*LeaveRequest, error)
	UpdateLeaveRequestEditableFields(
		ctx context.Context,
		leaveRequestID uuid.UUID,
		params UpdateLeaveRequestParams,
	) (*LeaveRequest, error)
	UpdateLeaveRequestDecision(
		ctx context.Context,
		leaveRequestID uuid.UUID,
		status string,
		decisionNote *string,
		decidedByEmployeeID uuid.UUID,
	) (*LeaveRequest, error)
	GetActiveLeavePolicyByType(ctx context.Context, leaveType string) (*LeavePolicy, error)
	LockEmployeeForLeaveBalance(ctx context.Context, employeeID uuid.UUID) error
	GetLeaveHoursPerDay(ctx context.Context, employeeID uuid.UUID) (int32, error)
	GetEmployeeContractAtDate(
		ctx context.Context,
		employeeID uuid.UUID,
		date time.Time,
	) (*LeaveContractAtDate, error)
	ComputeLegalLeaveTotalForYear(
		ctx context.Context,
		employeeID uuid.UUID,
		year int32,
		asOf time.Time,
	) (int32, error)
	ComputeLegalLeaveUsedForYear(
		ctx context.Context,
		employeeID uuid.UUID,
		year int32,
	) (int32, error)
}

type LeaveRepository interface {
	WithTx(ctx context.Context, fn func(tx LeaveTxRepository) error) error
	CreateLeaveRequest(ctx context.Context, params CreateLeaveRequestParams) (*LeaveRequest, error)
	GetActiveLeavePolicyByType(ctx context.Context, leaveType string) (*LeavePolicy, error)
	GetEmployeeContractAtDate(
		ctx context.Context,
		employeeID uuid.UUID,
		date time.Time,
	) (*LeaveContractAtDate, error)
	ListMyLeaveRequests(
		ctx context.Context,
		params ListMyLeaveRequestsParams,
	) (*LeaveRequestPage, error)
	ListLeaveRequests(
		ctx context.Context,
		params ListLeaveRequestsParams,
	) (*LeaveRequestPage, error)
	ListLeaveCalendar(
		ctx context.Context,
		params ListLeaveCalendarParams,
	) ([]LeaveCalendarEmployee, error)
	GetMyLeaveRequestStats(ctx context.Context, employeeID uuid.UUID) (*LeaveRequestStats, error)
	GetLeaveRequestStats(ctx context.Context) (*LeaveRequestStats, error)
	ListLeaveBalances(
		ctx context.Context,
		params ListLeaveBalancesParams,
	) (*LeaveBalancePage, error)

	GetLeaveBalanceDetails(
		ctx context.Context,
		params GetLeaveBalanceDetailsParams,
	) (*LeaveBalanceDetails, error)
}

type LeaveService interface {
	CreateLeaveRequest(
		ctx context.Context,
		actorEmployeeID uuid.UUID,
		params CreateLeaveRequestParams,
	) (*LeaveRequest, error)
	CreateLeaveRequestByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params CreateLeaveRequestParams,
	) (*LeaveRequest, error)
	UpdateLeaveRequest(
		ctx context.Context,
		actorEmployeeID, leaveRequestID uuid.UUID,
		params UpdateLeaveRequestParams,
	) (*LeaveRequest, error)
	UpdateLeaveRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, leaveRequestID uuid.UUID,
		params UpdateLeaveRequestParams,
		adminUpdateNote string,
	) (*LeaveRequest, error)
	DecideLeaveRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, leaveRequestID uuid.UUID,
		params DecideLeaveRequestParams,
	) (*LeaveRequest, error)
	ListMyLeaveRequests(
		ctx context.Context,
		params ListMyLeaveRequestsParams,
	) (*LeaveRequestPage, error)
	ListLeaveRequests(
		ctx context.Context,
		params ListLeaveRequestsParams,
	) (*LeaveRequestPage, error)
	ListLeaveCalendar(
		ctx context.Context,
		params ListLeaveCalendarParams,
	) ([]LeaveCalendarEmployee, error)
	GetMyLeaveRequestStats(ctx context.Context, employeeID uuid.UUID) (*LeaveRequestStats, error)
	GetLeaveRequestStats(ctx context.Context) (*LeaveRequestStats, error)
	ListLeaveBalances(
		ctx context.Context,
		params ListLeaveBalancesParams,
	) (*LeaveBalancePage, error)

	GetLeaveBalanceDetails(
		ctx context.Context,
		params GetLeaveBalanceDetailsParams,
	) (*LeaveBalanceDetails, error)
}
