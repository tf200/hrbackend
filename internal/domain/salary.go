package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ListSalaryScaleStepsParams struct {
	ActiveOnly bool
}

type SalaryScaleStepOption struct {
	ID            uuid.UUID
	SalaryTableID uuid.UUID
	Step          string
	IPNumber      *int32
	MonthlySalary float64
	HourlyRate    float64
	Label         string
}

type SalaryScaleGroup struct {
	Scale int32
	Steps []SalaryScaleStepOption
}

type SalaryScaleStepsMeta struct {
	SalaryTableID   uuid.UUID
	CaoCode         string
	SalaryTableName string
	EffectiveFrom   time.Time
	EffectiveTo     *time.Time
	ScaleCount      int
}

type SalaryScaleStepsResult struct {
	Groups []SalaryScaleGroup
	Meta   *SalaryScaleStepsMeta
}

var (
	ErrSalaryInvalidRequest   = errors.New("invalid salary request")
	ErrPayPeriodNotFound      = errors.New("pay period not found")
	ErrPayPeriodStateInvalid  = errors.New("pay period is not in a valid state for this operation")
	ErrPayPeriodAlreadyExists = errors.New("pay period already exists for this employee and date range")
	ErrPayPeriodNoEntries     = errors.New("no eligible time entries found for pay period")
)

const (
	PayPeriodStatusDraft = "draft"
	PayPeriodStatusPaid  = "paid"

	PayrollGroupFixed  = "fixed"
	PayrollGroupOnCall = "on_call"

	PayrollSourceSchedule    = "schedule"
	PayrollSourceOvertime    = "overtime"
	PayrollSourceLeavePayout = "leave_payout"
)

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
	PayrollGroup         string
	CutoffAt             *time.Time
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

type PayrollPeriodSummaryParams struct {
	PeriodStart    time.Time
	PeriodEnd      time.Time
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

type PayrollMonthStats struct {
	Month                       time.Time
	TotalBaseContractPay        float64
	TotalORTPay                 float64
	TotalOvertimePay            float64
	TotalRequestedLeaveHoursPay float64
	TotalRequestedLeaveHours    float64
	TotalGrossPayable           float64
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
	EmployeeID   uuid.UUID
	PeriodStart  time.Time
	PeriodEnd    time.Time
	PayrollGroup string
	CutoffAt     time.Time
}

type ClosePayrollMonthParams struct {
	PayrollGroup string
	Month        time.Time
	EmployeeIDs  []uuid.UUID
	CutoffAt     time.Time
}

type ClosePayrollPeriodParams struct {
	PayrollGroup string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	EmployeeIDs  []uuid.UUID
	CutoffAt     time.Time
}

type PayrollMonthCloseEmployeeResult struct {
	EmployeeID     uuid.UUID
	EmployeeName   string
	Status         string
	Reason         string
	PayPeriodID    *uuid.UUID
	GrossAmount    float64
	PendingEntries int32
}

type PayrollMonthCloseResult struct {
	Month        time.Time
	PayrollGroup string
	CutoffAt     time.Time
	ClosedCount  int32
	SkippedCount int32
	FailedCount  int32
	Items        []PayrollMonthCloseEmployeeResult
}

type PayrollPeriodCloseResult struct {
	PeriodStart  time.Time
	PeriodEnd    time.Time
	PayrollGroup string
	CutoffAt     time.Time
	ClosedCount  int32
	SkippedCount int32
	FailedCount  int32
	Items        []PayrollMonthCloseEmployeeResult
}

type PayrollPeriodOption struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	IsCurrent   bool
}

var PayrollPeriodAnchorStart = time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC)

func ResolvePayrollPeriod(date time.Time) (time.Time, time.Time) {
	day := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	daysSinceAnchor := int(day.Sub(PayrollPeriodAnchorStart).Hours() / 24)
	periodIndex := floorDiv(daysSinceAnchor, 28)
	periodStart := PayrollPeriodAnchorStart.AddDate(0, 0, periodIndex*28)
	return periodStart, periodStart.AddDate(0, 0, 27)
}

func PayrollPeriodOptionsThrough(date time.Time) []PayrollPeriodOption {
	currentStart, currentEnd := ResolvePayrollPeriod(date)
	if currentStart.Before(PayrollPeriodAnchorStart) {
		return []PayrollPeriodOption{}
	}

	options := []PayrollPeriodOption{}
	for periodStart := PayrollPeriodAnchorStart; !periodStart.After(currentStart); periodStart = periodStart.AddDate(0, 0, 28) {
		periodEnd := periodStart.AddDate(0, 0, 27)
		options = append(options, PayrollPeriodOption{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			IsCurrent:   periodStart.Equal(currentStart) && periodEnd.Equal(currentEnd),
		})
	}
	return options
}

func IsPayrollPeriodStart(date time.Time) bool {
	periodStart, _ := ResolvePayrollPeriod(date)
	day := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return day.Equal(periodStart)
}

func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && ((r < 0) != (b < 0)) {
		q--
	}
	return q
}

type ListPayPeriodsParams struct {
	Limit          int32
	Offset         int32
	Status         *string
	EmployeeSearch *string
}

type SalaryTxRepository interface {
	GetPayPeriodByEmployeePeriod(
		ctx context.Context,
		employeeID uuid.UUID,
		periodStart, periodEnd time.Time,
		payrollGroup string,
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
	AssignSchedulesToPayPeriod(
		ctx context.Context,
		payPeriodID uuid.UUID,
		scheduleIDs []uuid.UUID,
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

type SalaryRepository interface {
	WithTxSalary(ctx context.Context, fn func(tx SalaryTxRepository) error) error
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
	ListSalaryScaleSteps(ctx context.Context, params ListSalaryScaleStepsParams) (*SalaryScaleStepsResult, error)
}

type SalaryService interface {
	ListSalaryScaleSteps(ctx context.Context, params ListSalaryScaleStepsParams) (*SalaryScaleStepsResult, error)
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
	PreviewPayrollMonthClose(
		ctx context.Context,
		params ClosePayrollMonthParams,
	) (*PayrollMonthCloseResult, error)
	ClosePayrollMonthByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params ClosePayrollMonthParams,
	) (*PayrollMonthCloseResult, error)
	PreviewPayrollPeriodClose(
		ctx context.Context,
		params ClosePayrollPeriodParams,
	) (*PayrollPeriodCloseResult, error)
	ClosePayrollPeriodByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params ClosePayrollPeriodParams,
	) (*PayrollPeriodCloseResult, error)
	GetPayPeriodByID(ctx context.Context, payPeriodID uuid.UUID) (*PayPeriod, error)
	ListPayPeriods(ctx context.Context, params ListPayPeriodsParams) (*PayPeriodPage, error)
	MarkPayPeriodPaidByAdmin(
		ctx context.Context,
		adminEmployeeID, payPeriodID uuid.UUID,
	) (*PayPeriod, error)
	GetFixedPayrollMonthSummary(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*FixedPayrollMonthSummaryPage, error)
	GetOnCallPayrollMonthSummary(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*OnCallPayrollMonthSummaryPage, error)
	GetFixedPayrollPeriodSummary(
		ctx context.Context,
		params PayrollPeriodSummaryParams,
	) (*FixedPayrollMonthSummaryPage, error)
	GetOnCallPayrollPeriodSummary(
		ctx context.Context,
		params PayrollPeriodSummaryParams,
	) (*OnCallPayrollMonthSummaryPage, error)
	GetFixedPayrollMonthStats(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*PayrollMonthStats, error)
	GetOnCallPayrollMonthStats(
		ctx context.Context,
		params PayrollMonthSummaryParams,
	) (*PayrollMonthStats, error)
	GetFixedPayrollPeriodStats(
		ctx context.Context,
		params PayrollPeriodSummaryParams,
	) (*PayrollMonthStats, error)
	GetOnCallPayrollPeriodStats(
		ctx context.Context,
		params PayrollPeriodSummaryParams,
	) (*PayrollMonthStats, error)
	GetPayrollPeriodOptions(ctx context.Context) ([]PayrollPeriodOption, error)
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
