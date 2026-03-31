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
)

type LeaveRequest struct {
	ID                  uuid.UUID
	EmployeeID          uuid.UUID
	CreatedByEmployeeID *uuid.UUID
	LeaveType           string
	Status              string
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

type LeaveBalance struct {
	ID             uuid.UUID
	EmployeeID     uuid.UUID
	EmployeeName   string
	Year           int32
	LegalTotalDays int32
	ExtraTotalDays int32
	LegalUsedDays  int32
	ExtraUsedDays  int32
	LegalRemaining int32
	ExtraRemaining int32
	TotalRemaining int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type LeaveBalancePage struct {
	Items      []LeaveBalance
	TotalCount int64
}

type LeavePolicy struct {
	LeaveType      string
	DeductsBalance bool
}

type CreateLeaveRequestParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	LeaveType           string
	StartDate           time.Time
	EndDate             time.Time
	Reason              *string
}

type UpdateLeaveRequestParams struct {
	LeaveType *string
	StartDate *time.Time
	EndDate   *time.Time
	Reason    *string
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

type ListLeaveBalancesParams struct {
	Limit          int32
	Offset         int32
	EmployeeSearch *string
	Year           *int32
}

type ListMyLeaveBalancesParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Year       *int32
}

type AdjustLeaveBalanceParams struct {
	AdminEmployeeID uuid.UUID
	EmployeeID      uuid.UUID
	Year            int32
	LegalDaysDelta  int32
	ExtraDaysDelta  int32
	Reason          string
}

type LeaveTxRepository interface {
	GetLeaveRequestForUpdate(ctx context.Context, leaveRequestID uuid.UUID) (*LeaveRequest, error)
	UpdateLeaveRequestEditableFields(ctx context.Context, leaveRequestID uuid.UUID, params UpdateLeaveRequestParams) (*LeaveRequest, error)
	UpdateLeaveRequestDecision(ctx context.Context, leaveRequestID uuid.UUID, status string, decisionNote *string, decidedByEmployeeID uuid.UUID) (*LeaveRequest, error)
	GetActiveLeavePolicyByType(ctx context.Context, leaveType string) (*LeavePolicy, error)
	EnsureLeaveBalanceForYear(ctx context.Context, employeeID uuid.UUID, year int32) error
	GetLeaveBalanceForUpdate(ctx context.Context, employeeID uuid.UUID, year int32) (*LeaveBalance, error)
	ApplyLeaveBalanceDeduction(ctx context.Context, balanceID uuid.UUID, extraDays, legalDays int32) (*LeaveBalance, error)
	ApplyLeaveBalanceTotalAdjustment(ctx context.Context, balanceID uuid.UUID, legalDaysDelta, extraDaysDelta int32) (*LeaveBalance, error)
	CreateLeaveBalanceAdjustmentAudit(ctx context.Context, params CreateLeaveBalanceAdjustmentAuditParams) error
}

type CreateLeaveBalanceAdjustmentAuditParams struct {
	LeaveBalanceID       uuid.UUID
	EmployeeID           uuid.UUID
	Year                 int32
	LegalDaysDelta       int32
	ExtraDaysDelta       int32
	Reason               string
	AdjustedByEmployeeID uuid.UUID
	LegalTotalDaysBefore int32
	ExtraTotalDaysBefore int32
	LegalTotalDaysAfter  int32
	ExtraTotalDaysAfter  int32
}

type LeaveRepository interface {
	WithTx(ctx context.Context, fn func(tx LeaveTxRepository) error) error
	CreateLeaveRequest(ctx context.Context, params CreateLeaveRequestParams) (*LeaveRequest, error)
	GetActiveLeavePolicyByType(ctx context.Context, leaveType string) (*LeavePolicy, error)
	ListMyLeaveRequests(ctx context.Context, params ListMyLeaveRequestsParams) (*LeaveRequestPage, error)
	ListLeaveRequests(ctx context.Context, params ListLeaveRequestsParams) (*LeaveRequestPage, error)
	GetMyLeaveRequestStats(ctx context.Context, employeeID uuid.UUID) (*LeaveRequestStats, error)
	GetLeaveRequestStats(ctx context.Context) (*LeaveRequestStats, error)
	ListLeaveBalances(ctx context.Context, params ListLeaveBalancesParams) (*LeaveBalancePage, error)
	ListMyLeaveBalances(ctx context.Context, params ListMyLeaveBalancesParams) (*LeaveBalancePage, error)
}

type LeaveService interface {
	CreateLeaveRequest(ctx context.Context, actorEmployeeID uuid.UUID, params CreateLeaveRequestParams) (*LeaveRequest, error)
	CreateLeaveRequestByAdmin(ctx context.Context, adminEmployeeID uuid.UUID, params CreateLeaveRequestParams) (*LeaveRequest, error)
	UpdateLeaveRequest(ctx context.Context, actorEmployeeID, leaveRequestID uuid.UUID, params UpdateLeaveRequestParams) (*LeaveRequest, error)
	UpdateLeaveRequestByAdmin(ctx context.Context, adminEmployeeID, leaveRequestID uuid.UUID, params UpdateLeaveRequestParams, adminUpdateNote string) (*LeaveRequest, error)
	DecideLeaveRequestByAdmin(ctx context.Context, adminEmployeeID, leaveRequestID uuid.UUID, params DecideLeaveRequestParams) (*LeaveRequest, error)
	ListMyLeaveRequests(ctx context.Context, params ListMyLeaveRequestsParams) (*LeaveRequestPage, error)
	ListLeaveRequests(ctx context.Context, params ListLeaveRequestsParams) (*LeaveRequestPage, error)
	GetMyLeaveRequestStats(ctx context.Context, employeeID uuid.UUID) (*LeaveRequestStats, error)
	GetLeaveRequestStats(ctx context.Context) (*LeaveRequestStats, error)
	ListLeaveBalances(ctx context.Context, params ListLeaveBalancesParams) (*LeaveBalancePage, error)
	ListMyLeaveBalances(ctx context.Context, params ListMyLeaveBalancesParams) (*LeaveBalancePage, error)
	AdjustLeaveBalance(ctx context.Context, params AdjustLeaveBalanceParams) (*LeaveBalance, error)
}
