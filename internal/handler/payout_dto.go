package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const payoutMonthLayout = "2006-01"

type createPayoutRequestRequest struct {
	RequestedHours int32   `json:"requested_hours" binding:"required,min=1"`
	BalanceYear    int32   `json:"balance_year"    binding:"required,min=2000,max=2100"`
	RequestNote    *string `json:"request_note"`
}

type decidePayoutRequestByAdminRequest struct {
	Decision     string  `json:"decision"      binding:"required,oneof=approve reject"`
	DecisionNote *string `json:"decision_note"`
	SalaryMonth  *string `json:"salary_month"  binding:"omitempty,datetime=2006-01"`
}

type listMyPayoutRequestsRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected paid"`
}

type listPayoutRequestsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status"          binding:"omitempty,oneof=pending approved rejected paid"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type closePayPeriodRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id"  binding:"required"`
	PeriodStart string    `json:"period_start" binding:"required,datetime=2006-01-02"`
	PeriodEnd   string    `json:"period_end"   binding:"required,datetime=2006-01-02"`
}

type listPayPeriodsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status"          binding:"omitempty,oneof=draft paid"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type payrollMonthSummaryRequest struct {
	httpapi.PageRequest
	Month          string  `form:"month"           binding:"required,datetime=2006-01"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type payrollMonthDetailRequest struct {
	EmployeeID   string  `form:"employee_id"   binding:"required"`
	Month        string  `form:"month"         binding:"required,datetime=2006-01"`
	ContractType *string `form:"contract_type" binding:"omitempty,oneof=loondienst ZZP"`
}

type previewPayrollRequest struct {
	EmployeeID  uuid.UUID `form:"employee_id,parser=encoding.TextUnmarshaler" binding:"required"`
	PeriodStart string    `form:"period_start"                                 binding:"required,datetime=2006-01-02"`
	PeriodEnd   string    `form:"period_end"                                   binding:"required,datetime=2006-01-02"`
}

type previewMyPayrollRequest struct {
	PeriodStart string `form:"period_start" binding:"required,datetime=2006-01-02"`
	PeriodEnd   string `form:"period_end"   binding:"required,datetime=2006-01-02"`
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
	ShiftCount           int32                               `json:"shift_count"`
	PendingEntryCount    int32                               `json:"pending_entry_count"`
	PendingWorkedMinutes int32                               `json:"pending_worked_minutes"`
	PayPeriodID          *uuid.UUID                          `json:"pay_period_id,omitempty"`
	PayPeriodStatus      *string                             `json:"pay_period_status,omitempty"`
	PaidAt               *time.Time                          `json:"paid_at,omitempty"`
	MultiplierSummaries  []payrollMonthMultiplierSummaryItem `json:"multiplier_summaries"`
}

type payrollMonthDetailResponse struct {
	EmployeeID   uuid.UUID               `json:"employee_id"`
	EmployeeName string                  `json:"employee_name"`
	Month        string                  `json:"month"`
	DataSource   string                  `json:"data_source"`
	PayPeriod    *payPeriodResponse      `json:"pay_period,omitempty"`
	Preview      *payrollPreviewResponse `json:"preview,omitempty"`
}

type payrollMonthORTOverviewEmployeeResponse struct {
	EmployeeID        uuid.UUID                           `json:"employee_id"`
	EmployeeName      string                              `json:"employee_name"`
	Month             string                              `json:"month"`
	IsCurrentMonth    bool                                `json:"is_current_month"`
	IsLocked          bool                                `json:"is_locked"`
	HasLockedSnapshot bool                                `json:"has_locked_snapshot"`
	DataSource        string                              `json:"data_source"`
	WorkedMinutes     float64                             `json:"worked_minutes"`
	PaidMinutes       float64                             `json:"paid_minutes"`
	BaseAmount        float64                             `json:"base_amount"`
	PremiumAmount     float64                             `json:"premium_amount"`
	PayPeriodID       *uuid.UUID                          `json:"pay_period_id,omitempty"`
	PayPeriodStatus   *string                             `json:"pay_period_status,omitempty"`
	PaidAt            *time.Time                          `json:"paid_at,omitempty"`
	Distribution      []payrollMonthMultiplierSummaryItem `json:"distribution"`
}

type payrollMonthORTOverviewResponse struct {
	Month        string                                    `json:"month"`
	Distribution []payrollMonthMultiplierSummaryItem       `json:"distribution"`
	Next         *string                                   `json:"next"`
	Previous     *string                                   `json:"previous"`
	Count        int64                                     `json:"count"`
	PageSize     int32                                     `json:"page_size"`
	Results      []payrollMonthORTOverviewEmployeeResponse `json:"results"`
}

type ortRuleResponse struct {
	Order                 int32   `json:"order"`
	RatePercent           float64 `json:"rate_percent"`
	Label                 string  `json:"label"`
	Description           string  `json:"description"`
	ContractType          string  `json:"contract_type"`
	IrregularHoursProfile *string `json:"irregular_hours_profile,omitempty"`
	DayType               string  `json:"day_type"`
	TimeFrom              *string `json:"time_from,omitempty"`
	TimeTo                *string `json:"time_to,omitempty"`
}

type ortRulesResponse struct {
	Rules []ortRuleResponse `json:"rules"`
}

type payrollMonthMultiplierSummaryItem struct {
	RatePercent   float64 `json:"rate_percent"`
	WorkedMinutes float64 `json:"worked_minutes"`
	PaidMinutes   float64 `json:"paid_minutes"`
	BaseAmount    float64 `json:"base_amount"`
	PremiumAmount float64 `json:"premium_amount"`
}

func toCreatePayoutRequestParams(
	employeeID uuid.UUID,
	req createPayoutRequestRequest,
) domain.CreatePayoutRequestParams {
	return domain.CreatePayoutRequestParams{
		EmployeeID:          employeeID,
		CreatedByEmployeeID: employeeID,
		RequestedHours:      req.RequestedHours,
		BalanceYear:         req.BalanceYear,
		RequestNote:         req.RequestNote,
	}
}

func toDecidePayoutRequestParams(
	req decidePayoutRequestByAdminRequest,
) (domain.DecidePayoutRequestParams, error) {
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

func toListMyPayoutRequestsParams(
	employeeID uuid.UUID,
	req listMyPayoutRequestsRequest,
) domain.ListMyPayoutRequestsParams {
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

func toPayrollMonthSummaryParams(
	req payrollMonthSummaryRequest,
) (domain.PayrollMonthSummaryParams, error) {
	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		return domain.PayrollMonthSummaryParams{}, err
	}

	return domain.PayrollMonthSummaryParams{
		Month:          month.UTC(),
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		EmployeeSearch: req.EmployeeSearch,
		ContractType:   nil,
	}, nil
}

func toPayrollMonthORTOverviewParams(
	req payrollMonthSummaryRequest,
) (domain.PayrollMonthORTOverviewParams, error) {
	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		return domain.PayrollMonthORTOverviewParams{}, err
	}

	return domain.PayrollMonthORTOverviewParams{
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
		ShiftCount:           item.ShiftCount,
		PendingEntryCount:    item.PendingEntryCount,
		PendingWorkedMinutes: item.PendingWorkedMinutes,
		PayPeriodID:          item.PayPeriodID,
		PayPeriodStatus:      item.PayPeriodStatus,
		PaidAt:               item.PaidAt,
		MultiplierSummaries:  multipliers,
	}
}

func toPayrollMonthSummaryResponses(
	items []domain.PayrollMonthSummaryRow,
) []payrollMonthSummaryResponse {
	results := make([]payrollMonthSummaryResponse, len(items))
	for i, item := range items {
		results[i] = toPayrollMonthSummaryResponse(item)
	}
	return results
}

func toPayrollMonthMultiplierSummaryItems(
	items []domain.PayrollMultiplierSummary,
) []payrollMonthMultiplierSummaryItem {
	results := make([]payrollMonthMultiplierSummaryItem, len(items))
	for i, item := range items {
		results[i] = payrollMonthMultiplierSummaryItem{
			RatePercent:   item.RatePercent,
			WorkedMinutes: item.WorkedMinutes,
			PaidMinutes:   item.PaidMinutes,
			BaseAmount:    item.BaseAmount,
			PremiumAmount: item.PremiumAmount,
		}
	}
	return results
}

func toPayrollMonthORTOverviewEmployeeResponse(
	item domain.PayrollMonthORTOverviewRow,
) payrollMonthORTOverviewEmployeeResponse {
	return payrollMonthORTOverviewEmployeeResponse{
		EmployeeID:        item.EmployeeID,
		EmployeeName:      item.EmployeeName,
		Month:             item.Month.UTC().Format(payoutMonthLayout),
		IsCurrentMonth:    item.IsCurrentMonth,
		IsLocked:          item.IsLocked,
		HasLockedSnapshot: item.HasLockedSnapshot,
		DataSource:        item.DataSource,
		WorkedMinutes:     item.WorkedMinutes,
		PaidMinutes:       item.PaidMinutes,
		BaseAmount:        item.BaseAmount,
		PremiumAmount:     item.PremiumAmount,
		PayPeriodID:       item.PayPeriodID,
		PayPeriodStatus:   item.PayPeriodStatus,
		PaidAt:            item.PaidAt,
		Distribution:      toPayrollMonthMultiplierSummaryItems(item.Distribution),
	}
}

func toPayrollMonthORTOverviewEmployeeResponses(
	items []domain.PayrollMonthORTOverviewRow,
) []payrollMonthORTOverviewEmployeeResponse {
	results := make([]payrollMonthORTOverviewEmployeeResponse, len(items))
	for i, item := range items {
		results[i] = toPayrollMonthORTOverviewEmployeeResponse(item)
	}
	return results
}

func toPayrollMonthORTOverviewResponse(
	page *domain.PayrollMonthORTOverviewPage,
	paged httpapi.PageResponse[payrollMonthORTOverviewEmployeeResponse],
) payrollMonthORTOverviewResponse {
	return payrollMonthORTOverviewResponse{
		Month:        page.Month.UTC().Format(payoutMonthLayout),
		Distribution: toPayrollMonthMultiplierSummaryItems(page.Distribution),
		Next:         paged.Next,
		Previous:     paged.Previous,
		Count:        paged.Count,
		PageSize:     paged.PageSize,
		Results:      paged.Results,
	}
}

func toORTRulesResponse(item *domain.ORTRulesResponse) ortRulesResponse {
	rules := make([]ortRuleResponse, len(item.Rules))
	for i, rule := range item.Rules {
		rules[i] = ortRuleResponse{
			Order:                 rule.Order,
			RatePercent:           rule.RatePercent,
			Label:                 rule.Label,
			Description:           rule.Description,
			ContractType:          rule.ContractType,
			IrregularHoursProfile: rule.IrregularHoursProfile,
			DayType:               rule.DayType,
			TimeFrom:              rule.TimeFrom,
			TimeTo:                rule.TimeTo,
		}
	}
	return ortRulesResponse{Rules: rules}
}

func toPayrollMonthDetailRequest(req payrollMonthDetailRequest) (uuid.UUID, time.Time, *string, error) {
	employeeRaw := strings.TrimSpace(req.EmployeeID)
	employeeRaw = strings.TrimPrefix(employeeRaw, "[")
	employeeRaw = strings.TrimSuffix(employeeRaw, "]")
	employeeRaw = strings.TrimSpace(employeeRaw)
	employeeRaw = strings.Trim(employeeRaw, "\"")

	employeeID, err := uuid.Parse(employeeRaw)
	if err != nil {
		return uuid.Nil, time.Time{}, nil, err
	}

	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		return uuid.Nil, time.Time{}, nil, err
	}
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	contractType, err := normalizePayrollContractType(req.ContractType)
	if err != nil {
		return uuid.Nil, time.Time{}, nil, err
	}

	return employeeID, monthStart, contractType, nil
}

func toPayrollMonthDetailResponse(item *domain.PayrollMonthDetail) payrollMonthDetailResponse {
	res := payrollMonthDetailResponse{
		EmployeeID:   item.EmployeeID,
		EmployeeName: item.EmployeeName,
		Month:        item.Month.Format(payoutMonthLayout),
		DataSource:   item.DataSource,
	}

	if item.PayPeriod != nil {
		payPeriod := toPayPeriodResponse(item.PayPeriod)
		res.PayPeriod = &payPeriod
	}
	if item.Preview != nil {
		preview := toPayrollPreviewResponse(item.Preview)
		res.Preview = &preview
	}

	return res
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

func normalizePayrollContractType(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	switch normalized {
	case "loondienst":
		value := "LOONDIENST"
		return &value, nil
	case "zzp":
		value := "ZZP"
		return &value, nil
	default:
		return nil, fmt.Errorf("invalid contract_type, expected loondienst or ZZP")
	}
}

func formatPayoutSalaryMonth(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(payoutMonthLayout)
	return &formatted
}

// ---------------------------------------------------------------------------
// Salary page DTOs
// ---------------------------------------------------------------------------

type salaryPageRequest struct {
	Month string `form:"month" binding:"required,datetime=2006-01"`
}

type salaryPageResponse struct {
	Employee    salaryPageEmployeeResponse       `json:"employee"`
	Month       string                           `json:"month"`
	Period      salaryPagePeriodResponse         `json:"period"`
	Contract    salaryPageContractResponse       `json:"contract"`
	Status      salaryPageStatusResponse         `json:"status"`
	Summary     salaryPageSummaryResponse        `json:"summary"`
	ORT         salaryPageORTResponse            `json:"ort"`
	LineItems   []salaryPageLineItemResponse     `json:"line_items"`
	Pending     []salaryPagePendingEntryResponse `json:"pending_entries"`
	LeavePayout salaryPageLeavePayoutResponse    `json:"leave_payout"`
	Actions     salaryPageActionsResponse        `json:"actions"`
}

type salaryPageEmployeeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type salaryPagePeriodResponse struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type salaryPageContractResponse struct {
	Type                  string   `json:"type"`
	Label                 string   `json:"label"`
	HourlyRate            *float64 `json:"hourly_rate"`
	ContractHoursPerWeek  *float64 `json:"contract_hours_per_week"`
	IrregularHoursProfile string   `json:"irregular_hours_profile"`
	StartDate             *string  `json:"start_date"`
	EndDate               *string  `json:"end_date"`
}

type salaryPageStatusResponse struct {
	DataSource      string     `json:"data_source"`
	PayPeriodStatus *string    `json:"pay_period_status"`
	IsLocked        bool       `json:"is_locked"`
	IsPaid          bool       `json:"is_paid"`
	PaidAt          *time.Time `json:"paid_at"`
	Label           string     `json:"label"`
}

type salaryPageSummaryResponse struct {
	WorkedMinutes          float64 `json:"worked_minutes"`
	PaidMinutes            float64 `json:"paid_minutes"`
	WorkedHours            float64 `json:"worked_hours"`
	PaidHours              float64 `json:"paid_hours"`
	ShiftCount             int32   `json:"shift_count"`
	PendingEntryCount      int32   `json:"pending_entry_count"`
	PendingWorkedMinutes   int32   `json:"pending_worked_minutes"`
	PendingHours           float64 `json:"pending_hours"`
	BaseGrossAmount        float64 `json:"base_gross_amount"`
	IrregularGrossAmount   float64 `json:"irregular_gross_amount"`
	LeavePayoutGrossAmount float64 `json:"leave_payout_gross_amount"`
	GrossAmount            float64 `json:"gross_amount"`
}

type salaryPageMultiplierSummaryItem struct {
	Label         string  `json:"label"`
	RatePercent   float64 `json:"rate_percent"`
	WorkedMinutes float64 `json:"worked_minutes"`
	PaidMinutes   float64 `json:"paid_minutes"`
	PaidHours     float64 `json:"paid_hours"`
	BaseAmount    float64 `json:"base_amount"`
	PremiumAmount float64 `json:"premium_amount"`
}

type salaryPageORTResponse struct {
	Applies bool                              `json:"applies"`
	Profile string                            `json:"profile"`
	Buckets []salaryPageMultiplierSummaryItem `json:"buckets"`
}

type salaryPageLineItemResponse struct {
	ID                 uuid.UUID `json:"id"`
	TimeEntryID        uuid.UUID `json:"time_entry_id"`
	WorkDate           string    `json:"work_date"`
	DisplayDate        string    `json:"display_date"`
	LineType           string    `json:"line_type"`
	Label              string    `json:"label"`
	StartTime          string    `json:"start_time"`
	EndTime            string    `json:"end_time"`
	BreakMinutes       int32     `json:"break_minutes"`
	WorkedMinutes      int32     `json:"worked_minutes"`
	PaidMinutes        float64   `json:"paid_minutes"`
	PaidHours          float64   `json:"paid_hours"`
	AppliedRatePercent float64   `json:"applied_rate_percent"`
	BaseAmount         float64   `json:"base_amount"`
	PremiumAmount      float64   `json:"premium_amount"`
	GrossAmount        float64   `json:"gross_amount"`
}

type salaryPagePendingEntryResponse struct {
	ID            uuid.UUID `json:"id"`
	WorkDate      string    `json:"work_date"`
	DisplayDate   string    `json:"display_date"`
	Status        string    `json:"status"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	WorkedMinutes int32     `json:"worked_minutes"`
}

type salaryPagePayoutRequestResponse struct {
	ID             uuid.UUID  `json:"id"`
	SalaryMonth    *string    `json:"salary_month"`
	BalanceYear    int32      `json:"balance_year"`
	RequestedHours int32      `json:"requested_hours"`
	HourlyRate     float64    `json:"hourly_rate"`
	GrossAmount    float64    `json:"gross_amount"`
	Status         string     `json:"status"`
	RequestedAt    time.Time  `json:"requested_at"`
	DecidedAt      *time.Time `json:"decided_at"`
	PaidAt         *time.Time `json:"paid_at"`
	RequestNote    *string    `json:"request_note"`
	DecisionNote   *string    `json:"decision_note"`
}

type salaryPageLeavePayoutResponse struct {
	Applies                  bool                              `json:"applies"`
	CanRequest               bool                              `json:"can_request"`
	AvailableExtraLeaveHours *int32                            `json:"available_extra_leave_hours"`
	Requests                 []salaryPagePayoutRequestResponse `json:"requests"`
}

type salaryPageActionsResponse struct {
	CanDownloadPDF        bool   `json:"can_download_pdf"`
	PDFURL                string `json:"pdf_url"`
	CanRequestLeavePayout bool   `json:"can_request_leave_payout"`
}

// ---------------------------------------------------------------------------
// Salary page mappers
// ---------------------------------------------------------------------------

func toSalaryPageResponse(data *domain.SalaryPageData) *salaryPageResponse {
	if data == nil {
		return nil
	}

	monthStr := data.Month.Format(payoutMonthLayout)
	monthStartStr := data.Month.Format("2006-01-02")
	monthEnd := data.Month.AddDate(0, 1, -1)
	monthEndStr := monthEnd.Format("2006-01-02")

	// Contract label
	contractLabel := contractTypeLabel(data.ContractType)

	// Status fields
	dataSource := data.DataSource
	var payPeriodStatus *string
	isLocked := data.PayPeriod != nil
	isPaid := false
	var paidAt *time.Time
	if data.PayPeriod != nil {
		s := data.PayPeriod.Status
		payPeriodStatus = &s
		isPaid = data.PayPeriod.Status == "paid"
		paidAt = data.PayPeriod.PaidAt
	}
	statusLabel := statusLabel(dataSource, payPeriodStatus)

	// Summary
	workedMinutes, paidMinutes := computeWorkedPaidMinutes(data)
	shiftCount := computeShiftCount(data)
	pendingEntryCount := int32(len(data.PendingEntries))
	var pendingWorkedMinutes int32
	for _, pe := range data.PendingEntries {
		pendingWorkedMinutes += pe.WorkedMinutes
	}
	baseGross := computeBaseGross(data)
	irregularGross := computeIrregularGross(data)
	leavePayoutGross := computeLeavePayoutGrossAmount(data.LeavePayoutRequests)
	totalGross := roundCurrency(baseGross + irregularGross + leavePayoutGross)

	// ORT buckets
	ortBuckets := buildSalaryPageORTBuckets(data)

	// Line items
	lineItems := buildSalaryPageLineItems(data)

	// Pending entries
	pendingEntries := buildSalaryPagePendingEntries(data.PendingEntries)

	// Leave payout
	leavePayout := buildSalaryPageLeavePayout(data)

	// Actions
	hasData := (data.PayPeriod != nil && len(data.PayPeriod.LineItems) > 0) ||
		(data.Preview != nil && len(data.Preview.LineItems) > 0)
	canRequest := data.ContractType == "loondienst" && data.ExtraLeaveRemaining > 0
	pdfURL := fmt.Sprintf("/api/payouts/detail/pdf?month=%s", monthStr)

	return &salaryPageResponse{
		Employee: salaryPageEmployeeResponse{
			ID:   data.EmployeeID,
			Name: data.EmployeeName,
		},
		Month: monthStr,
		Period: salaryPagePeriodResponse{
			Start: monthStartStr,
			End:   monthEndStr,
		},
		Contract: salaryPageContractResponse{
			Type:                  data.ContractType,
			Label:                 contractLabel,
			HourlyRate:            data.ContractRate,
			ContractHoursPerWeek:  data.ContractHours,
			IrregularHoursProfile: data.IrregularHoursProfile,
			StartDate:             formatOptionalDate(data.ContractStartDate),
			EndDate:               formatOptionalDate(data.ContractEndDate),
		},
		Status: salaryPageStatusResponse{
			DataSource:      dataSource,
			PayPeriodStatus: payPeriodStatus,
			IsLocked:        isLocked,
			IsPaid:          isPaid,
			PaidAt:          paidAt,
			Label:           statusLabel,
		},
		Summary: salaryPageSummaryResponse{
			WorkedMinutes:          workedMinutes,
			PaidMinutes:            paidMinutes,
			WorkedHours:            roundCurrency(workedMinutes / 60),
			PaidHours:              roundCurrency(paidMinutes / 60),
			ShiftCount:             shiftCount,
			PendingEntryCount:      pendingEntryCount,
			PendingWorkedMinutes:   pendingWorkedMinutes,
			PendingHours:           roundCurrency(float64(pendingWorkedMinutes) / 60),
			BaseGrossAmount:        baseGross,
			IrregularGrossAmount:   irregularGross,
			LeavePayoutGrossAmount: leavePayoutGross,
			GrossAmount:            totalGross,
		},
		ORT: salaryPageORTResponse{
			Applies: data.ContractType == "loondienst",
			Profile: data.IrregularHoursProfile,
			Buckets: ortBuckets,
		},
		LineItems:   lineItems,
		Pending:     pendingEntries,
		LeavePayout: leavePayout,
		Actions: salaryPageActionsResponse{
			CanDownloadPDF:        hasData,
			PDFURL:                pdfURL,
			CanRequestLeavePayout: canRequest,
		},
	}
}

// ---------------------------------------------------------------------------
// Salary page helper functions
// ---------------------------------------------------------------------------

func contractTypeLabel(ct string) string {
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "loondienst":
		return "Loondienst"
	case "zzp":
		return "ZZP"
	default:
		return ct
	}
}

func statusLabel(dataSource string, payPeriodStatus *string) string {
	if dataSource == "live" {
		return "Live estimate"
	}
	if payPeriodStatus != nil {
		switch *payPeriodStatus {
		case "draft":
			return "Finalized (awaiting payment)"
		case "paid":
			return "Paid"
		}
	}
	return dataSource
}

func computeWorkedPaidMinutes(data *domain.SalaryPageData) (float64, float64) {
	if data.PayPeriod != nil {
		var worked, paid float64
		for _, item := range data.PayPeriod.LineItems {
			worked += item.MinutesWorked
			paid += item.MinutesWorked
		}
		return roundCurrency(worked), roundCurrency(paid)
	}
	if data.Preview != nil {
		worked := float64(data.Preview.TotalWorkedMinutes)
		var paid float64
		for _, item := range data.Preview.LineItems {
			paid += item.PaidMinutes
		}
		return roundCurrency(worked), roundCurrency(paid)
	}
	return 0, 0
}

func computeShiftCount(data *domain.SalaryPageData) int32 {
	if data.PayPeriod != nil {
		seen := make(map[uuid.UUID]struct{})
		for _, item := range data.PayPeriod.LineItems {
			if item.TimeEntryID != nil {
				seen[*item.TimeEntryID] = struct{}{}
			}
		}
		return int32(len(seen))
	}
	if data.Preview != nil {
		seen := make(map[uuid.UUID]struct{})
		for _, item := range data.Preview.LineItems {
			seen[item.TimeEntryID] = struct{}{}
		}
		return int32(len(seen))
	}
	return 0
}

func computeBaseGross(data *domain.SalaryPageData) float64 {
	if data.PayPeriod != nil {
		return data.PayPeriod.BaseGrossAmount
	}
	if data.Preview != nil {
		return data.Preview.BaseGrossAmount
	}
	return 0
}

func computeIrregularGross(data *domain.SalaryPageData) float64 {
	if data.PayPeriod != nil {
		return data.PayPeriod.IrregularGrossAmount
	}
	if data.Preview != nil {
		return data.Preview.IrregularGrossAmount
	}
	return 0
}

func computeLeavePayoutGrossAmount(requests []domain.PayoutRequest) float64 {
	var total float64
	for _, r := range requests {
		if r.Status == "approved" || r.Status == "paid" {
			total = roundCurrency(total + r.GrossAmount)
		}
	}
	return total
}

func buildSalaryPageORTBuckets(data *domain.SalaryPageData) []salaryPageMultiplierSummaryItem {
	rateBuckets := make(map[float64]*payrollMonthMultiplierSummaryItem)
	accumulate := func(ratePercent float64, paidMinutes, baseAmount, premiumAmount float64) {
		b := rateBuckets[ratePercent]
		if b == nil {
			b = &payrollMonthMultiplierSummaryItem{RatePercent: ratePercent}
			rateBuckets[ratePercent] = b
		}
		b.WorkedMinutes = roundCurrency(b.WorkedMinutes + paidMinutes)
		b.PaidMinutes = roundCurrency(b.PaidMinutes + paidMinutes)
		b.BaseAmount = roundCurrency(b.BaseAmount + baseAmount)
		b.PremiumAmount = roundCurrency(b.PremiumAmount + premiumAmount)
	}

	if data.PayPeriod != nil {
		for _, item := range data.PayPeriod.LineItems {
			accumulate(item.AppliedRatePercent, item.MinutesWorked, item.BaseAmount, item.PremiumAmount)
		}
	} else if data.Preview != nil {
		for _, item := range data.Preview.LineItems {
			accumulate(item.AppliedRatePercent, item.PaidMinutes, item.BaseAmount, item.PremiumAmount)
		}
	}

	// Sort by rate percent
	keys := make([]float64, 0, len(rateBuckets))
	for k := range rateBuckets {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	items := make([]salaryPageMultiplierSummaryItem, 0, len(keys))
	for _, rate := range keys {
		b := rateBuckets[rate]
		items = append(items, salaryPageMultiplierSummaryItem{
			Label:         ortBucketLabel(rate, data.IrregularHoursProfile),
			RatePercent:   b.RatePercent,
			WorkedMinutes: b.WorkedMinutes,
			PaidMinutes:   b.PaidMinutes,
			PaidHours:     roundCurrency(b.PaidMinutes / 60),
			BaseAmount:    b.BaseAmount,
			PremiumAmount: b.PremiumAmount,
		})
	}
	return items
}

func ortBucketLabel(ratePercent float64, profile string) string {
	switch ratePercent {
	case 0:
		return "Regular hours"
	case 25:
		if strings.EqualFold(strings.TrimSpace(profile), "roster") {
			return "Evening roster"
		}
		if strings.EqualFold(strings.TrimSpace(profile), "non_roster") {
			return "Evening non-roster"
		}
		return "Evening"
	case 30:
		return "Saturday daytime"
	case 45:
		return "Sunday / night"
	default:
		return fmt.Sprintf("ORT %.0f%%", ratePercent)
	}
}

func buildSalaryPageLineItems(data *domain.SalaryPageData) []salaryPageLineItemResponse {
	if data.PayPeriod != nil {
		return lineItemsFromPayPeriod(data.PayPeriod.LineItems)
	}
	if data.Preview != nil {
		return lineItemsFromPreview(data.Preview.LineItems)
	}
	return []salaryPageLineItemResponse{}
}

func lineItemsFromPreview(items []domain.PayrollPreviewLineItem) []salaryPageLineItemResponse {
	res := make([]salaryPageLineItemResponse, 0, len(items))
	for _, item := range items {
		res = append(res, salaryPageLineItemResponse{
			ID:                 item.TimeEntryID, // use time_entry_id as id for live
			TimeEntryID:        item.TimeEntryID,
			WorkDate:           item.WorkDate.Format("2006-01-02"),
			DisplayDate:        item.WorkDate.Format("Mon 2 Jan"),
			LineType:           item.HourType,
			Label:              item.Label,
			StartTime:          item.StartTime,
			EndTime:            item.EndTime,
			BreakMinutes:       item.BreakMinutes,
			WorkedMinutes:      item.MinutesWorked,
			PaidMinutes:        item.PaidMinutes,
			PaidHours:          roundCurrency(item.PaidMinutes / 60),
			AppliedRatePercent: item.AppliedRatePercent,
			BaseAmount:         item.BaseAmount,
			PremiumAmount:      item.PremiumAmount,
			GrossAmount:        roundCurrency(item.BaseAmount + item.PremiumAmount),
		})
	}
	return res
}

type payPeriodLineItemMetadata struct {
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	PaidMinutes  float64 `json:"paid_minutes"`
	BreakMinutes int32   `json:"break_minutes"`
}

func lineItemsFromPayPeriod(items []domain.PayPeriodLineItem) []salaryPageLineItemResponse {
	res := make([]salaryPageLineItemResponse, 0, len(items))
	for _, item := range items {
		startTime := ""
		endTime := ""
		breakMinutes := int32(0)

		if len(item.Metadata) > 0 {
			var meta payPeriodLineItemMetadata
			if err := json.Unmarshal(item.Metadata, &meta); err == nil {
				startTime = meta.StartTime
				endTime = meta.EndTime
				breakMinutes = meta.BreakMinutes
			}
		}

		res = append(res, salaryPageLineItemResponse{
			ID:                 item.ID,
			TimeEntryID:        ptrOrDefault(item.TimeEntryID),
			WorkDate:           item.WorkDate.Format("2006-01-02"),
			DisplayDate:        item.WorkDate.Format("Mon 2 Jan"),
			LineType:           item.LineType,
			Label:              "",
			StartTime:          startTime,
			EndTime:            endTime,
			BreakMinutes:       breakMinutes,
			WorkedMinutes:      int32(item.MinutesWorked),
			PaidMinutes:        item.MinutesWorked,
			PaidHours:          roundCurrency(item.MinutesWorked / 60),
			AppliedRatePercent: item.AppliedRatePercent,
			BaseAmount:         item.BaseAmount,
			PremiumAmount:      item.PremiumAmount,
			GrossAmount:        roundCurrency(item.BaseAmount + item.PremiumAmount),
		})
	}
	return res
}

func ptrOrDefault(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}

func buildSalaryPagePendingEntries(items []domain.PayrollPendingEntryDetail) []salaryPagePendingEntryResponse {
	res := make([]salaryPagePendingEntryResponse, 0, len(items))
	for _, item := range items {
		res = append(res, salaryPagePendingEntryResponse{
			ID:            item.ID,
			WorkDate:      item.WorkDate.Format("2006-01-02"),
			DisplayDate:   item.WorkDate.Format("Mon 2 Jan"),
			Status:        item.Status,
			StartTime:     item.StartTime,
			EndTime:       item.EndTime,
			WorkedMinutes: item.WorkedMinutes,
		})
	}
	return res
}

func buildSalaryPageLeavePayout(data *domain.SalaryPageData) salaryPageLeavePayoutResponse {
	applies := data.ContractType == "loondienst"
	canRequest := applies && data.ExtraLeaveRemaining > 0

	var availPtr *int32
	if applies {
		availPtr = &data.ExtraLeaveRemaining
	}

	requests := make([]salaryPagePayoutRequestResponse, 0, len(data.LeavePayoutRequests))
	for _, r := range data.LeavePayoutRequests {
		requests = append(requests, salaryPagePayoutRequestResponse{
			ID:             r.ID,
			SalaryMonth:    formatPayoutSalaryMonth(r.SalaryMonth),
			BalanceYear:    r.BalanceYear,
			RequestedHours: r.RequestedHours,
			HourlyRate:     r.HourlyRate,
			GrossAmount:    r.GrossAmount,
			Status:         r.Status,
			RequestedAt:    r.RequestedAt,
			DecidedAt:      r.DecidedAt,
			PaidAt:         r.PaidAt,
			RequestNote:    r.RequestNote,
			DecisionNote:   r.DecisionNote,
		})
	}

	return salaryPageLeavePayoutResponse{
		Applies:                  applies,
		CanRequest:               canRequest,
		AvailableExtraLeaveHours: availPtr,
		Requests:                 requests,
	}
}

func formatOptionalDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func roundCurrency(v float64) float64 {
	return math.Round(v*100) / 100
}
