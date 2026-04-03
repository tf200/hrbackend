package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const payoutMonthLayout = "2006-01"

type createPayoutRequestRequest struct {
	RequestedHours int32   `json:"requested_hours" binding:"required,min=1"`
	BalanceYear    int32   `json:"balance_year" binding:"required,min=2000,max=2100"`
	RequestNote    *string `json:"request_note"`
}

type decidePayoutRequestByAdminRequest struct {
	Decision     string  `json:"decision" binding:"required,oneof=approve reject"`
	DecisionNote *string `json:"decision_note"`
	SalaryMonth  *string `json:"salary_month" binding:"omitempty,datetime=2006-01"`
}

type listMyPayoutRequestsRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected paid"`
}

type listPayoutRequestsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status" binding:"omitempty,oneof=pending approved rejected paid"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type closePayPeriodRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id" binding:"required"`
	PeriodStart string    `json:"period_start" binding:"required,datetime=2006-01-02"`
	PeriodEnd   string    `json:"period_end" binding:"required,datetime=2006-01-02"`
}

type listPayPeriodsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status" binding:"omitempty,oneof=draft paid"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type payrollMonthSummaryRequest struct {
	httpapi.PageRequest
	Month          string  `form:"month" binding:"required,datetime=2006-01"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type previewPayrollRequest struct {
	EmployeeID  uuid.UUID `form:"employee_id" binding:"required"`
	PeriodStart string    `form:"period_start" binding:"required,datetime=2006-01-02"`
	PeriodEnd   string    `form:"period_end" binding:"required,datetime=2006-01-02"`
}

type previewMyPayrollRequest struct {
	PeriodStart string `form:"period_start" binding:"required,datetime=2006-01-02"`
	PeriodEnd   string `form:"period_end" binding:"required,datetime=2006-01-02"`
}

type payoutRequestResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	EmployeeName        string     `json:"employee_name"`
	CreatedByEmployeeID uuid.UUID  `json:"created_by_employee_id"`
	RequestedHours      int32      `json:"requested_hours"`
	BalanceYear         int32      `json:"balance_year"`
	HourlyRate          float64    `json:"hourly_rate"`
	GrossAmount         float64    `json:"gross_amount"`
	SalaryMonth         *string    `json:"salary_month,omitempty"`
	Status              string     `json:"status"`
	RequestNote         *string    `json:"request_note,omitempty"`
	DecisionNote        *string    `json:"decision_note,omitempty"`
	DecidedByEmployeeID *uuid.UUID `json:"decided_by_employee_id,omitempty"`
	PaidByEmployeeID    *uuid.UUID `json:"paid_by_employee_id,omitempty"`
	RequestedAt         time.Time  `json:"requested_at"`
	DecidedAt           *time.Time `json:"decided_at,omitempty"`
	PaidAt              *time.Time `json:"paid_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type payrollPreviewResponse struct {
	EmployeeID           uuid.UUID                    `json:"employee_id"`
	EmployeeName         string                       `json:"employee_name"`
	PeriodStart          string                       `json:"period_start"`
	PeriodEnd            string                       `json:"period_end"`
	TotalWorkedMinutes   int32                        `json:"total_worked_minutes"`
	BaseGrossAmount      float64                      `json:"base_gross_amount"`
	IrregularGrossAmount float64                      `json:"irregular_gross_amount"`
	GrossAmount          float64                      `json:"gross_amount"`
	LineItems            []payrollPreviewLineResponse `json:"line_items"`
}

type payrollPreviewLineResponse struct {
	TimeEntryID           uuid.UUID `json:"time_entry_id"`
	WorkDate              string    `json:"work_date"`
	HourType              string    `json:"hour_type"`
	StartTime             string    `json:"start_time"`
	EndTime               string    `json:"end_time"`
	IrregularHoursProfile string    `json:"irregular_hours_profile"`
	AppliedRatePercent    float64   `json:"applied_rate_percent"`
	MinutesWorked         int32     `json:"minutes_worked"`
	PaidMinutes           float64   `json:"paid_minutes"`
	BaseAmount            float64   `json:"base_amount"`
	PremiumAmount         float64   `json:"premium_amount"`
}

type payPeriodResponse struct {
	ID                   uuid.UUID               `json:"id"`
	EmployeeID           uuid.UUID               `json:"employee_id"`
	EmployeeName         string                  `json:"employee_name"`
	PeriodStart          string                  `json:"period_start"`
	PeriodEnd            string                  `json:"period_end"`
	Status               string                  `json:"status"`
	BaseGrossAmount      float64                 `json:"base_gross_amount"`
	IrregularGrossAmount float64                 `json:"irregular_gross_amount"`
	GrossAmount          float64                 `json:"gross_amount"`
	PaidAt               *time.Time              `json:"paid_at,omitempty"`
	CreatedByEmployeeID  *uuid.UUID              `json:"created_by_employee_id,omitempty"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	LineItems            []payPeriodLineResponse `json:"line_items,omitempty"`
}

type payPeriodLineResponse struct {
	ID                    uuid.UUID       `json:"id"`
	PayPeriodID           uuid.UUID       `json:"pay_period_id"`
	TimeEntryID           *uuid.UUID      `json:"time_entry_id,omitempty"`
	WorkDate              string          `json:"work_date"`
	LineType              string          `json:"line_type"`
	IrregularHoursProfile string          `json:"irregular_hours_profile"`
	AppliedRatePercent    float64         `json:"applied_rate_percent"`
	MinutesWorked         float64         `json:"minutes_worked"`
	BaseAmount            float64         `json:"base_amount"`
	PremiumAmount         float64         `json:"premium_amount"`
	Metadata              json.RawMessage `json:"metadata"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type payrollMonthSummaryResponse struct {
	EmployeeID           uuid.UUID                           `json:"employee_id"`
	EmployeeName         string                              `json:"employee_name"`
	Month                string                              `json:"month"`
	IsCurrentMonth       bool                                `json:"is_current_month"`
	IsLocked             bool                                `json:"is_locked"`
	HasLockedSnapshot    bool                                `json:"has_locked_snapshot"`
	DataSource           string                              `json:"data_source"`
	WorkedMinutes        int32                               `json:"worked_minutes"`
	PaidMinutes          float64                             `json:"paid_minutes"`
	BaseGrossAmount      float64                             `json:"base_gross_amount"`
	IrregularGrossAmount float64                             `json:"irregular_gross_amount"`
	GrossAmount          float64                             `json:"gross_amount"`
	PendingEntryCount    int32                               `json:"pending_entry_count"`
	PendingWorkedMinutes int32                               `json:"pending_worked_minutes"`
	PayPeriodID          *uuid.UUID                          `json:"pay_period_id,omitempty"`
	PayPeriodStatus      *string                             `json:"pay_period_status,omitempty"`
	PaidAt               *time.Time                          `json:"paid_at,omitempty"`
	MultiplierSummaries  []payrollMonthMultiplierSummaryItem `json:"multiplier_summaries"`
}

type payrollMonthMultiplierSummaryItem struct {
	RatePercent   float64 `json:"rate_percent"`
	WorkedMinutes float64 `json:"worked_minutes"`
	PaidMinutes   float64 `json:"paid_minutes"`
	BaseAmount    float64 `json:"base_amount"`
	PremiumAmount float64 `json:"premium_amount"`
}

func toCreatePayoutRequestParams(employeeID uuid.UUID, req createPayoutRequestRequest) domain.CreatePayoutRequestParams {
	return domain.CreatePayoutRequestParams{
		EmployeeID:          employeeID,
		CreatedByEmployeeID: employeeID,
		RequestedHours:      req.RequestedHours,
		BalanceYear:         req.BalanceYear,
		RequestNote:         req.RequestNote,
	}
}

func toDecidePayoutRequestParams(req decidePayoutRequestByAdminRequest) (domain.DecidePayoutRequestParams, error) {
	salaryMonth, err := parsePayoutSalaryMonth(req.SalaryMonth)
	if err != nil {
		return domain.DecidePayoutRequestParams{}, err
	}
	return domain.DecidePayoutRequestParams{
		Decision:     req.Decision,
		DecisionNote: req.DecisionNote,
		SalaryMonth:  salaryMonth,
	}, nil
}

func toListMyPayoutRequestsParams(employeeID uuid.UUID, req listMyPayoutRequestsRequest) domain.ListMyPayoutRequestsParams {
	return domain.ListMyPayoutRequestsParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Status:     req.Status,
	}
}

func toListPayoutRequestsParams(req listPayoutRequestsRequest) domain.ListPayoutRequestsParams {
	return domain.ListPayoutRequestsParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		Status:         req.Status,
		EmployeeSearch: req.EmployeeSearch,
	}
}

func toClosePayPeriodParams(req closePayPeriodRequest) (domain.ClosePayPeriodParams, error) {
	periodStart, err := time.Parse(timeEntryDateLayout, req.PeriodStart)
	if err != nil {
		return domain.ClosePayPeriodParams{}, err
	}
	periodEnd, err := time.Parse(timeEntryDateLayout, req.PeriodEnd)
	if err != nil {
		return domain.ClosePayPeriodParams{}, err
	}

	return domain.ClosePayPeriodParams{
		EmployeeID:  req.EmployeeID,
		PeriodStart: periodStart.UTC(),
		PeriodEnd:   periodEnd.UTC(),
	}, nil
}

func toListPayPeriodsParams(req listPayPeriodsRequest) domain.ListPayPeriodsParams {
	return domain.ListPayPeriodsParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		Status:         req.Status,
		EmployeeSearch: req.EmployeeSearch,
	}
}

func toPayrollMonthSummaryParams(req payrollMonthSummaryRequest) (domain.PayrollMonthSummaryParams, error) {
	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		return domain.PayrollMonthSummaryParams{}, err
	}

	return domain.PayrollMonthSummaryParams{
		Month:          month.UTC(),
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		EmployeeSearch: req.EmployeeSearch,
	}, nil
}

func toPreviewPayrollParams(req previewPayrollRequest) (domain.PayrollPreviewParams, error) {
	periodStart, err := time.Parse(timeEntryDateLayout, req.PeriodStart)
	if err != nil {
		return domain.PayrollPreviewParams{}, err
	}
	periodEnd, err := time.Parse(timeEntryDateLayout, req.PeriodEnd)
	if err != nil {
		return domain.PayrollPreviewParams{}, err
	}

	return domain.PayrollPreviewParams{
		EmployeeID:  req.EmployeeID,
		PeriodStart: periodStart.UTC(),
		PeriodEnd:   periodEnd.UTC(),
	}, nil
}

func toPreviewMyPayrollDates(req previewMyPayrollRequest) (time.Time, time.Time, error) {
	periodStart, err := time.Parse(timeEntryDateLayout, req.PeriodStart)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	periodEnd, err := time.Parse(timeEntryDateLayout, req.PeriodEnd)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return periodStart.UTC(), periodEnd.UTC(), nil
}

func toPayoutRequestResponse(item domain.PayoutRequest) payoutRequestResponse {
	return payoutRequestResponse{
		ID:                  item.ID,
		EmployeeID:          item.EmployeeID,
		EmployeeName:        item.EmployeeName,
		CreatedByEmployeeID: item.CreatedByEmployeeID,
		RequestedHours:      item.RequestedHours,
		BalanceYear:         item.BalanceYear,
		HourlyRate:          item.HourlyRate,
		GrossAmount:         item.GrossAmount,
		SalaryMonth:         formatPayoutSalaryMonth(item.SalaryMonth),
		Status:              item.Status,
		RequestNote:         item.RequestNote,
		DecisionNote:        item.DecisionNote,
		DecidedByEmployeeID: item.DecidedByEmployeeID,
		PaidByEmployeeID:    item.PaidByEmployeeID,
		RequestedAt:         item.RequestedAt,
		DecidedAt:           item.DecidedAt,
		PaidAt:              item.PaidAt,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func toPayoutRequestResponses(items []domain.PayoutRequest) []payoutRequestResponse {
	results := make([]payoutRequestResponse, len(items))
	for i, item := range items {
		results[i] = toPayoutRequestResponse(item)
	}
	return results
}

func toPayrollPreviewResponse(item *domain.PayrollPreview) payrollPreviewResponse {
	lines := make([]payrollPreviewLineResponse, len(item.LineItems))
	for i, line := range item.LineItems {
		lines[i] = payrollPreviewLineResponse{
			TimeEntryID:           line.TimeEntryID,
			WorkDate:              line.WorkDate.UTC().Format(timeEntryDateLayout),
			HourType:              line.HourType,
			StartTime:             line.StartTime,
			EndTime:               line.EndTime,
			IrregularHoursProfile: line.IrregularHoursProfile,
			AppliedRatePercent:    line.AppliedRatePercent,
			MinutesWorked:         line.MinutesWorked,
			PaidMinutes:           line.PaidMinutes,
			BaseAmount:            line.BaseAmount,
			PremiumAmount:         line.PremiumAmount,
		}
	}

	return payrollPreviewResponse{
		EmployeeID:           item.EmployeeID,
		EmployeeName:         item.EmployeeName,
		PeriodStart:          item.PeriodStart.UTC().Format(timeEntryDateLayout),
		PeriodEnd:            item.PeriodEnd.UTC().Format(timeEntryDateLayout),
		TotalWorkedMinutes:   item.TotalWorkedMinutes,
		BaseGrossAmount:      item.BaseGrossAmount,
		IrregularGrossAmount: item.IrregularGrossAmount,
		GrossAmount:          item.GrossAmount,
		LineItems:            lines,
	}
}

func toPayPeriodResponse(item *domain.PayPeriod) payPeriodResponse {
	lines := make([]payPeriodLineResponse, len(item.LineItems))
	for i, line := range item.LineItems {
		lines[i] = payPeriodLineResponse{
			ID:                    line.ID,
			PayPeriodID:           line.PayPeriodID,
			TimeEntryID:           line.TimeEntryID,
			WorkDate:              line.WorkDate.UTC().Format(timeEntryDateLayout),
			LineType:              line.LineType,
			IrregularHoursProfile: line.IrregularHoursProfile,
			AppliedRatePercent:    line.AppliedRatePercent,
			MinutesWorked:         line.MinutesWorked,
			BaseAmount:            line.BaseAmount,
			PremiumAmount:         line.PremiumAmount,
			Metadata:              json.RawMessage(line.Metadata),
			CreatedAt:             line.CreatedAt,
			UpdatedAt:             line.UpdatedAt,
		}
	}

	return payPeriodResponse{
		ID:                   item.ID,
		EmployeeID:           item.EmployeeID,
		EmployeeName:         item.EmployeeName,
		PeriodStart:          item.PeriodStart.UTC().Format(timeEntryDateLayout),
		PeriodEnd:            item.PeriodEnd.UTC().Format(timeEntryDateLayout),
		Status:               item.Status,
		BaseGrossAmount:      item.BaseGrossAmount,
		IrregularGrossAmount: item.IrregularGrossAmount,
		GrossAmount:          item.GrossAmount,
		PaidAt:               item.PaidAt,
		CreatedByEmployeeID:  item.CreatedByEmployeeID,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
		LineItems:            lines,
	}
}

func toPayPeriodResponses(items []domain.PayPeriod) []payPeriodResponse {
	results := make([]payPeriodResponse, len(items))
	for i, item := range items {
		results[i] = toPayPeriodResponse(&item)
	}
	return results
}

func toPayrollMonthSummaryResponse(item domain.PayrollMonthSummaryRow) payrollMonthSummaryResponse {
	multipliers := make([]payrollMonthMultiplierSummaryItem, len(item.MultiplierSummaries))
	for i, multiplier := range item.MultiplierSummaries {
		multipliers[i] = payrollMonthMultiplierSummaryItem{
			RatePercent:   multiplier.RatePercent,
			WorkedMinutes: multiplier.WorkedMinutes,
			PaidMinutes:   multiplier.PaidMinutes,
			BaseAmount:    multiplier.BaseAmount,
			PremiumAmount: multiplier.PremiumAmount,
		}
	}

	return payrollMonthSummaryResponse{
		EmployeeID:           item.EmployeeID,
		EmployeeName:         item.EmployeeName,
		Month:                item.Month.UTC().Format(payoutMonthLayout),
		IsCurrentMonth:       item.IsCurrentMonth,
		IsLocked:             item.IsLocked,
		HasLockedSnapshot:    item.HasLockedSnapshot,
		DataSource:           item.DataSource,
		WorkedMinutes:        item.WorkedMinutes,
		PaidMinutes:          item.PaidMinutes,
		BaseGrossAmount:      item.BaseGrossAmount,
		IrregularGrossAmount: item.IrregularGrossAmount,
		GrossAmount:          item.GrossAmount,
		PendingEntryCount:    item.PendingEntryCount,
		PendingWorkedMinutes: item.PendingWorkedMinutes,
		PayPeriodID:          item.PayPeriodID,
		PayPeriodStatus:      item.PayPeriodStatus,
		PaidAt:               item.PaidAt,
		MultiplierSummaries:  multipliers,
	}
}

func toPayrollMonthSummaryResponses(items []domain.PayrollMonthSummaryRow) []payrollMonthSummaryResponse {
	results := make([]payrollMonthSummaryResponse, len(items))
	for i, item := range items {
		results[i] = toPayrollMonthSummaryResponse(item)
	}
	return results
}

func parsePayoutSalaryMonth(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(payoutMonthLayout, *value)
	if err != nil {
		return nil, fmt.Errorf("invalid salary_month format, expected YYYY-MM")
	}
	firstDay := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	return &firstDay, nil
}

func formatPayoutSalaryMonth(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(payoutMonthLayout)
	return &formatted
}
