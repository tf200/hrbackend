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
	ErrPayPeriodNotFound              = errors.New("pay period not found")
	ErrPayPeriodStateInvalid          = errors.New(
		"pay period is not in a valid state for this operation",
	)
	ErrPayPeriodAlreadyExists = errors.New(
		"pay period already exists for this employee and date range",
	)
	ErrPayPeriodNoEntries = errors.New("no eligible time entries found for pay period")
)

const (
	PayoutRequestStatusPending  = "pending"
	PayoutRequestStatusApproved = "approved"
	PayoutRequestStatusRejected = "rejected"
	PayoutRequestStatusPaid     = "paid"

	PayPeriodStatusDraft = "draft"
	PayPeriodStatusPaid  = "paid"

	PayrollSourceSchedule    = "schedule"
	PayrollSourceOvertime    = "overtime"
	PayrollSourceLeavePayout = "leave_payout"
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

type CreatePayoutRequestByAdminParams struct {
	EmployeeID     uuid.UUID
	RequestedHours int32
	BalanceYear    int32
	SalaryMonth    time.Time
	RequestNote    *string
	DecisionNote   *string
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
	ScheduleID            *uuid.UUID
	OvertimeEntryID       *uuid.UUID
	LeavePayoutRequestID  *uuid.UUID
	SourceType            string
	Label                 string
	ContractType          string
	WorkDate              time.Time
	StartTime             string
	EndTime               string
	BreakMinutes          int32
	IrregularHoursProfile string
	AppliedRatePercent    float64
	MinutesWorked         int32
	PaidMinutes           float64
	BaseAmount            float64
	PremiumAmount         float64
}

type PayrollWorkItem struct {
	ID                    uuid.UUID
	EmployeeID            uuid.UUID
	EmployeeName          string
	Label                 string
	WorkDate              time.Time
	StartTime             string
	EndTime               string
	BreakMinutes          int32
	MinutesWorked         float64
	SourceType            string
	ScheduleID            *uuid.UUID
	OvertimeEntryID       *uuid.UUID
	LeavePayoutRequestID  *uuid.UUID
	ContractType          string
	ContractRate          *float64
	GrossAmountOverride   *float64
	IrregularHoursProfile string
}

type NationalHoliday struct {
	Date time.Time
	Name string
}

type PayPeriod struct {
	ID                   uuid.UUID
	EmployeeID           uuid.UUID
	EmployeeName         string
	PeriodStart          time.Time
	PeriodEnd            time.Time
	Status               string
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	PaidAt               *time.Time
	CreatedByEmployeeID  *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LineItems            []PayPeriodLineItem
}

type PayPeriodLineItem struct {
	ID                    uuid.UUID
	PayPeriodID           uuid.UUID
	ScheduleID            *uuid.UUID
	OvertimeEntryID       *uuid.UUID
	LeavePayoutRequestID  *uuid.UUID
	ContractType          string
	WorkDate              time.Time
	LineType              string
	IrregularHoursProfile string
	AppliedRatePercent    float64
	MinutesWorked         float64
	BaseAmount            float64
	PremiumAmount         float64
	Metadata              []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type PayPeriodPage struct {
	Items      []PayPeriod
	TotalCount int64
}

type PayrollMonthSummaryParams struct {
	Month          time.Time
	Limit          int32
	Offset         int32
	EmployeeSearch *string
	ContractType   *string
}

type PayrollMonthORTOverviewParams struct {
	Month          time.Time
	Limit          int32
	Offset         int32
	EmployeeSearch *string
}

type PayrollMonthSummaryPage struct {
	Items      []PayrollMonthSummaryRow
	TotalCount int64
}

type FixedPayrollMonthSummaryPage struct {
	Items      []FixedPayrollMonthSummaryRow
	TotalCount int64
}

type OnCallPayrollMonthSummaryPage struct {
	Items      []OnCallPayrollMonthSummaryRow
	TotalCount int64
}

type PayrollMonthORTOverviewPage struct {
	Month        time.Time
	Distribution []PayrollMultiplierSummary
	Items        []PayrollMonthORTOverviewRow
	TotalCount   int64
}

type ORTRulesResponse struct {
	Rules []ORTRule
}

type PayrollMonthSummaryRow struct {
	EmployeeID           uuid.UUID
	EmployeeName         string
	Month                time.Time
	IsCurrentMonth       bool
	IsLocked             bool
	HasLockedSnapshot    bool
	DataSource           string
	WorkedMinutes        int32
	PaidMinutes          float64
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	ShiftCount           int32
	PendingEntryCount    int32
	PendingWorkedMinutes int32
	PayPeriodID          *uuid.UUID
	PayPeriodStatus      *string
	PaidAt               *time.Time
	MultiplierSummaries  []PayrollMultiplierSummary
}

type FixedPayrollMonthSummaryRow struct {
	EmployeeID              uuid.UUID
	EmployeeName            string
	Month                   time.Time
	IsCurrentMonth          bool
	CalculationMode         string
	DataSource              string
	ContractBaseAmount      float64
	ActualORTAmount         float64
	ForecastORTAmount       float64
	ApprovedOvertimeAmount  float64
	LeavePayoutAmount       float64
	PayableGrossAmount      float64
	ProjectedGrossAmount    float64
	ContractPaidMinutes     float64
	ScheduledActualMinutes  int32
	ScheduledFutureMinutes  int32
	ApprovedOvertimeMinutes int32
	LeavePayoutMinutes      int32
	PendingEntryCount       int32
	PendingWorkedMinutes    int32
	ContractSegments        []FixedPayrollContractSegment
	ORTBreakdown            []FixedPayrollORTBreakdown
	OvertimeBreakdown       []FixedPayrollOvertimeBreakdown
	LeavePayoutBreakdown    []FixedPayrollLeavePayoutBreakdown
}

type FixedPayrollContractSegment struct {
	ContractID           uuid.UUID
	ContractType         string
	ActiveFrom           time.Time
	ActiveUntil          time.Time
	HoursPerWeek         float64
	FullTimeHoursPerWeek float64
	MonthlySalary        float64
	HourlyRate           float64
	ProrationRatio       float64
	BaseAmount           float64
}

type FixedPayrollORTBreakdown struct {
	ScheduleID    *uuid.UUID
	WorkDate      time.Time
	StartTime     string
	EndTime       string
	Status        string
	RatePercent   float64
	Minutes       int32
	BasisAmount   float64
	PremiumAmount float64
}

type FixedPayrollOvertimeBreakdown struct {
	OvertimeEntryID *uuid.UUID
	WorkDate        time.Time
	Minutes         int32
	Amount          float64
	Status          string
}

type FixedPayrollLeavePayoutBreakdown struct {
	LeavePayoutRequestID *uuid.UUID
	SalaryMonth          time.Time
	RequestedHours       int32
	Minutes              int32
	Amount               float64
	Status               string
}

type OnCallPayrollMonthSummaryRow struct {
	EmployeeID              uuid.UUID
	EmployeeName            string
	Month                   time.Time
	IsCurrentMonth          bool
	CalculationMode         string
	DataSource              string
	WorkedMinutes           int32
	WorkedHoursAmount       float64
	ApprovedOvertimeMinutes int32
	ApprovedOvertimeAmount  float64
	LeavePayoutMinutes      int32
	LeavePayoutAmount       float64
	PayableGrossAmount      float64
	PendingEntryCount       int32
	PendingWorkedMinutes    int32
	WorkedHoursBreakdown    []OnCallPayrollWorkedHoursBreakdown
	OvertimeBreakdown       []OnCallPayrollOvertimeBreakdown
	LeavePayoutBreakdown    []OnCallPayrollLeavePayoutBreakdown
}

type OnCallPayrollWorkedHoursBreakdown struct {
	ScheduleID  *uuid.UUID
	WorkDate    time.Time
	StartTime   string
	EndTime     string
	Minutes     int32
	HourlyRate  float64
	BaseAmount  float64
	TotalAmount float64
}

type OnCallPayrollOvertimeBreakdown struct {
	OvertimeEntryID *uuid.UUID
	WorkDate        time.Time
	Minutes         int32
	HourlyRate      float64
	Amount          float64
	Status          string
}

type OnCallPayrollLeavePayoutBreakdown struct {
	LeavePayoutRequestID *uuid.UUID
	SalaryMonth          time.Time
	RequestedHours       int32
	Minutes              int32
	Amount               float64
	Status               string
}

type PayrollMonthORTOverviewRow struct {
	EmployeeID        uuid.UUID
	EmployeeName      string
	Month             time.Time
	IsCurrentMonth    bool
	IsLocked          bool
	HasLockedSnapshot bool
	DataSource        string
	WorkedMinutes     float64
	PaidMinutes       float64
	BaseAmount        float64
	PremiumAmount     float64
	PayPeriodID       *uuid.UUID
	PayPeriodStatus   *string
	PaidAt            *time.Time
	Distribution      []PayrollMultiplierSummary
}

type PayrollMonthDetail struct {
	EmployeeID   uuid.UUID
	EmployeeName string
	Month        time.Time
	DataSource   string
	PayPeriod    *PayPeriod
	Preview      *PayrollPreview
}

type SalaryPageData struct {
	EmployeeID   uuid.UUID
	EmployeeName string
	Month        time.Time

	ContractType          string
	ContractRate          *float64
	ContractHours         *float64
	IrregularHoursProfile string
	ContractStartDate     *time.Time
	ContractEndDate       *time.Time

	DataSource string
	PayPeriod  *PayPeriod
	Preview    *PayrollPreview

	PendingEntries      []PayrollPendingEntryDetail
	LeavePayoutRequests []PayoutRequest
	ExtraLeaveRemaining int32
}

type PayrollMultiplierSummary struct {
	RatePercent   float64
	WorkedMinutes float64
	PaidMinutes   float64
	BaseAmount    float64
	PremiumAmount float64
}

type ORTRule struct {
	Order                 int32
	RatePercent           float64
	Label                 string
	Description           string
	ContractType          string
	IrregularHoursProfile *string
	DayType               string
	TimeFrom              *string
	TimeTo                *string
}

type PayrollMonthEmployee struct {
	EmployeeID   uuid.UUID
	EmployeeName string
}

type FixedPayrollContractSegmentSource struct {
	EmployeeID           uuid.UUID
	ContractID           uuid.UUID
	ContractType         string
	ActiveFrom           time.Time
	ActiveUntil          time.Time
	HoursPerWeek         float64
	FullTimeHoursPerWeek float64
	MonthlySalary        float64
	HourlyRate           float64
}

type PayrollMonthPendingSummary struct {
	EmployeeID           uuid.UUID
	PendingEntryCount    int32
	PendingWorkedMinutes int32
}

type PayrollMonthPendingEntry struct {
	EmployeeID    uuid.UUID
	WorkedMinutes int32
	ContractType  string
}

type PayrollPendingEntryDetail struct {
	ID            uuid.UUID
	WorkDate      time.Time
	StartTime     string
	EndTime       string
	BreakMinutes  int32
	WorkedMinutes int32
	Status        string
}

type PayrollLockedMultiplierSummary struct {
	PayPeriodID   uuid.UUID
	RatePercent   float64
	WorkedMinutes float64
	PaidMinutes   float64
	BaseAmount    float64
	PremiumAmount float64
}

type ClosePayPeriodParams struct {
	EmployeeID  uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type ListPayPeriodsParams struct {
	Limit          int32
	Offset         int32
	Status         *string
	EmployeeSearch *string
}

type PayoutTxRepository interface {
	GetEmployeePayoutContract(ctx context.Context, employeeID uuid.UUID) (*PayoutContract, error)
	EnsureLeaveBalanceForYear(ctx context.Context, employeeID uuid.UUID, year int32) error
	GetPayoutBalanceForUpdate(
		ctx context.Context,
		employeeID uuid.UUID,
		year int32,
	) (*PayoutBalanceSnapshot, error)
	CreatePayoutRequest(
		ctx context.Context,
		params CreatePayoutRequestTxParams,
	) (*PayoutRequest, error)
	GetPayoutRequestForUpdate(
		ctx context.Context,
		payoutRequestID uuid.UUID,
	) (*PayoutRequest, error)
	ApprovePayoutRequest(
		ctx context.Context,
		payoutRequestID, decidedByEmployeeID uuid.UUID,
		salaryMonth time.Time,
		decisionNote *string,
	) (*PayoutRequest, error)
	RejectPayoutRequest(
		ctx context.Context,
		payoutRequestID, decidedByEmployeeID uuid.UUID,
		decisionNote *string,
	) (*PayoutRequest, error)
	MarkPayoutRequestPaid(
		ctx context.Context,
		payoutRequestID, paidByEmployeeID uuid.UUID,
	) (*PayoutRequest, error)
	ApplyLeaveBalanceDeduction(
		ctx context.Context,
		balanceID uuid.UUID,
		extraHours, legalHours int32,
	) (*LeaveBalance, error)
	GetPayPeriodByEmployeePeriod(
		ctx context.Context,
		employeeID uuid.UUID,
		periodStart, periodEnd time.Time,
	) (*PayPeriod, error)
	LockPayrollOvertimeEntries(
		ctx context.Context,
		params PayrollPreviewParams,
	) ([]uuid.UUID, error)
	LockPayrollPreviewWorkItems(
		ctx context.Context,
		params PayrollPreviewParams,
	) ([]PayrollWorkItem, error)
	CreatePayPeriod(
		ctx context.Context,
		params ClosePayPeriodParams,
		createdByEmployeeID uuid.UUID,
		preview PayrollPreview,
	) (*PayPeriod, error)
	CreatePayPeriodLineItem(
		ctx context.Context,
		payPeriodID uuid.UUID,
		item PayPeriodLineItem,
	) (*PayPeriodLineItem, error)
	AssignOvertimeEntriesToPayPeriod(
		ctx context.Context,
		payPeriodID uuid.UUID,
		overtimeEntryIDs []uuid.UUID,
	) error
	AssignLeavePayoutRequestsToPayPeriod(
		ctx context.Context,
		payPeriodID uuid.UUID,
		leavePayoutRequestIDs []uuid.UUID,
	) error
	GetPayPeriodForUpdate(ctx context.Context, payPeriodID uuid.UUID) (*PayPeriod, error)
	MarkPayPeriodPaid(ctx context.Context, payPeriodID uuid.UUID) (*PayPeriod, error)
	MarkLeavePayoutRequestsPaidByPayPeriod(
		ctx context.Context,
		payPeriodID, paidByEmployeeID uuid.UUID,
	) error
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
	ListMyPayoutRequests(
		ctx context.Context,
		params ListMyPayoutRequestsParams,
	) (*PayoutRequestPage, error)
	ListPayoutRequests(
		ctx context.Context,
		params ListPayoutRequestsParams,
	) (*PayoutRequestPage, error)
	GetPayrollPreviewEmployee(ctx context.Context, employeeID uuid.UUID) (*EmployeeDetail, error)
	ListPayrollPreviewWorkItems(
		ctx context.Context,
		params PayrollPreviewParams,
	) ([]PayrollWorkItem, error)
	ListNationalHolidays(
		ctx context.Context,
		countryCode string,
		startDate, endDate time.Time,
	) ([]NationalHoliday, error)
	GetPayPeriodByID(ctx context.Context, payPeriodID uuid.UUID) (*PayPeriod, error)
	ListPayPeriods(ctx context.Context, params ListPayPeriodsParams) (*PayPeriodPage, error)
	ListPayPeriodLineItems(ctx context.Context, payPeriodID uuid.UUID) ([]PayPeriodLineItem, error)
	ListPayrollMonthEmployees(
		ctx context.Context,
		params PayrollMonthSummaryParams,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthEmployee, int64, error)
	ListPayrollMonthEmployeesAll(
		ctx context.Context,
		params PayrollMonthORTOverviewParams,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthEmployee, error)
	ListFixedPayrollMonthEmployees(
		ctx context.Context,
		params PayrollMonthSummaryParams,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthEmployee, int64, error)
	ListOnCallPayrollMonthEmployees(
		ctx context.Context,
		params PayrollMonthSummaryParams,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthEmployee, int64, error)
	ListFixedPayrollContractSegments(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]FixedPayrollContractSegmentSource, error)
	ListPayPeriodsByEmployeesAndRange(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]PayPeriod, error)
	ListPayrollMonthLockedMultiplierSummaries(
		ctx context.Context,
		payPeriodIDs []uuid.UUID,
	) ([]PayrollLockedMultiplierSummary, error)
	ListPayrollMonthApprovedWorkItems(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]PayrollWorkItem, error)
	ListPayrollMonthPendingSummaries(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthPendingSummary, error)
	ListPayrollMonthPendingEntries(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]PayrollMonthPendingEntry, error)

	// Salary page queries (single employee, mine endpoint)
	ListPendingOvertimeEntriesDetail(
		ctx context.Context,
		employeeID uuid.UUID,
		monthStart, monthEnd time.Time,
	) ([]PayrollPendingEntryDetail, error)
	ListPayoutRequestsByEmployeeAndMonth(
		ctx context.Context,
		employeeID uuid.UUID,
		salaryMonth time.Time,
	) ([]PayoutRequest, error)
	GetLeaveBalanceExtraRemaining(
		ctx context.Context,
		employeeID uuid.UUID,
		year int32,
	) (int32, error)
}

type PayoutService interface {
	CreatePayoutRequest(
		ctx context.Context,
		actorEmployeeID uuid.UUID,
		params CreatePayoutRequestParams,
	) (*PayoutRequest, error)
	DecidePayoutRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, payoutRequestID uuid.UUID,
		params DecidePayoutRequestParams,
	) (*PayoutRequest, error)
	MarkPayoutRequestPaidByAdmin(
		ctx context.Context,
		adminEmployeeID, payoutRequestID uuid.UUID,
	) (*PayoutRequest, error)
	CreateApprovedPayoutRequestByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params CreatePayoutRequestByAdminParams,
	) (*PayoutRequest, error)
	ListMyPayoutRequests(
		ctx context.Context,
		params ListMyPayoutRequestsParams,
	) (*PayoutRequestPage, error)
	ListPayoutRequests(
		ctx context.Context,
		params ListPayoutRequestsParams,
	) (*PayoutRequestPage, error)
	PreviewPayroll(ctx context.Context, params PayrollPreviewParams) (*PayrollPreview, error)
	PreviewMyPayroll(
		ctx context.Context,
		actorEmployeeID uuid.UUID,
		periodStart, periodEnd time.Time,
	) (*PayrollPreview, error)
	ClosePayPeriod(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params ClosePayPeriodParams,
	) (*PayPeriod, error)
	GetPayPeriodByID(ctx context.Context, payPeriodID uuid.UUID) (*PayPeriod, error)
	ListPayPeriods(ctx context.Context, params ListPayPeriodsParams) (*PayPeriodPage, error)
	MarkPayPeriodPaidByAdmin(
		ctx context.Context,
		adminEmployeeID, payPeriodID uuid.UUID,
	) (*PayPeriod, error)
	GetPayrollMonthSummary(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*PayrollMonthSummaryPage, error)
	GetFixedPayrollMonthSummary(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*FixedPayrollMonthSummaryPage, error)
	GetOnCallPayrollMonthSummary(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*OnCallPayrollMonthSummaryPage, error)
	GetPayrollMonthORTOverview(
		ctx context.Context,
		params PayrollMonthORTOverviewParams,
	) (*PayrollMonthORTOverviewPage, error)
	GetORTRules(ctx context.Context) (*ORTRulesResponse, error)
	GetPayrollMonthDetail(
		ctx context.Context,
		employeeID uuid.UUID,
		month time.Time,
		contractType *string,
	) (*PayrollMonthDetail, error)
	GetMySalaryPage(
		ctx context.Context,
		employeeID uuid.UUID,
		month time.Time,
	) (*SalaryPageData, error)
	ExportPayrollMonthPDF(
		ctx context.Context,
		employeeID uuid.UUID,
		month time.Time,
		contractType *string,
	) ([]byte, string, error)
}
