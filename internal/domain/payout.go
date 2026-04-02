package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPayoutRequestInvalidRequest    = errors.New("invalid payout request")
	ErrPayoutRequestNotFound          = errors.New("payout request not found")
	ErrPayoutRequestForbidden         = errors.New("payout request is not accessible by the actor")
	ErrPayoutRequestStateInvalid      = errors.New("payout request is not in an editable state")
	ErrPayoutRequestInsufficientHours = errors.New("insufficient extra leave balance for payout")
)

const (
	PayoutRequestStatusPending  = "pending"
	PayoutRequestStatusApproved = "approved"
	PayoutRequestStatusRejected = "rejected"
	PayoutRequestStatusPaid     = "paid"
)

type PayoutRequest struct {
	ID                  uuid.UUID
	EmployeeID          uuid.UUID
	EmployeeName        string
	CreatedByEmployeeID uuid.UUID
	RequestedHours      int32
	BalanceYear         int32
	HourlyRate          float64
	GrossAmount         float64
	SalaryMonth         *time.Time
	Status              string
	RequestNote         *string
	DecisionNote        *string
	DecidedByEmployeeID *uuid.UUID
	PaidByEmployeeID    *uuid.UUID
	RequestedAt         time.Time
	DecidedAt           *time.Time
	PaidAt              *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PayoutRequestPage struct {
	Items      []PayoutRequest
	TotalCount int64
}

type PayoutContract struct {
	ContractType string
	ContractRate *float64
}

type CreatePayoutRequestParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	RequestedHours      int32
	BalanceYear         int32
	RequestNote         *string
}

type DecidePayoutRequestParams struct {
	Decision     string
	DecisionNote *string
	SalaryMonth  *time.Time
}

type ListMyPayoutRequestsParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Status     *string
}

type ListPayoutRequestsParams struct {
	Limit          int32
	Offset         int32
	Status         *string
	EmployeeSearch *string
}

type PayoutBalanceSnapshot struct {
	LeaveBalanceID uuid.UUID
	ExtraRemaining int32
}

type PayrollPreviewParams struct {
	EmployeeID  uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type PayrollPreview struct {
	EmployeeID           uuid.UUID
	EmployeeName         string
	PeriodStart          time.Time
	PeriodEnd            time.Time
	TotalWorkedMinutes   int32
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	LineItems            []PayrollPreviewLineItem
}

type PayrollPreviewLineItem struct {
	TimeEntryID           uuid.UUID
	WorkDate              time.Time
	HourType              string
	StartTime             string
	EndTime               string
	IrregularHoursProfile string
	AppliedRatePercent    float64
	MinutesWorked         int32
	BaseAmount            float64
	PremiumAmount         float64
}

type PayrollPreviewTimeEntry struct {
	ID                    uuid.UUID
	EmployeeID            uuid.UUID
	EmployeeName          string
	EntryDate             time.Time
	StartTime             string
	EndTime               string
	BreakMinutes          int32
	HourType              string
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
}

type NationalHoliday struct {
	Date time.Time
	Name string
}

type PayoutTxRepository interface {
	GetEmployeePayoutContract(ctx context.Context, employeeID uuid.UUID) (*PayoutContract, error)
	EnsureLeaveBalanceForYear(ctx context.Context, employeeID uuid.UUID, year int32) error
	GetPayoutBalanceForUpdate(ctx context.Context, employeeID uuid.UUID, year int32) (*PayoutBalanceSnapshot, error)
	CreatePayoutRequest(ctx context.Context, params CreatePayoutRequestTxParams) (*PayoutRequest, error)
	GetPayoutRequestForUpdate(ctx context.Context, payoutRequestID uuid.UUID) (*PayoutRequest, error)
	ApprovePayoutRequest(ctx context.Context, payoutRequestID, decidedByEmployeeID uuid.UUID, salaryMonth time.Time, decisionNote *string) (*PayoutRequest, error)
	RejectPayoutRequest(ctx context.Context, payoutRequestID, decidedByEmployeeID uuid.UUID, decisionNote *string) (*PayoutRequest, error)
	MarkPayoutRequestPaid(ctx context.Context, payoutRequestID, paidByEmployeeID uuid.UUID) (*PayoutRequest, error)
	ApplyLeaveBalanceDeduction(ctx context.Context, balanceID uuid.UUID, extraHours, legalHours int32) (*LeaveBalance, error)
}

type CreatePayoutRequestTxParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	RequestedHours      int32
	BalanceYear         int32
	HourlyRate          float64
	GrossAmount         float64
	RequestNote         *string
}

type PayoutRepository interface {
	WithTx(ctx context.Context, fn func(tx PayoutTxRepository) error) error
	ListMyPayoutRequests(ctx context.Context, params ListMyPayoutRequestsParams) (*PayoutRequestPage, error)
	ListPayoutRequests(ctx context.Context, params ListPayoutRequestsParams) (*PayoutRequestPage, error)
	GetPayrollPreviewEmployee(ctx context.Context, employeeID uuid.UUID) (*EmployeeDetail, error)
	ListPayrollPreviewTimeEntries(ctx context.Context, params PayrollPreviewParams) ([]PayrollPreviewTimeEntry, error)
	ListNationalHolidays(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]NationalHoliday, error)
}

type PayoutService interface {
	CreatePayoutRequest(ctx context.Context, actorEmployeeID uuid.UUID, params CreatePayoutRequestParams) (*PayoutRequest, error)
	DecidePayoutRequestByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID, params DecidePayoutRequestParams) (*PayoutRequest, error)
	MarkPayoutRequestPaidByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID) (*PayoutRequest, error)
	ListMyPayoutRequests(ctx context.Context, params ListMyPayoutRequestsParams) (*PayoutRequestPage, error)
	ListPayoutRequests(ctx context.Context, params ListPayoutRequestsParams) (*PayoutRequestPage, error)
	PreviewPayroll(ctx context.Context, params PayrollPreviewParams) (*PayrollPreview, error)
	PreviewMyPayroll(ctx context.Context, actorEmployeeID uuid.UUID, periodStart, periodEnd time.Time) (*PayrollPreview, error)
}
