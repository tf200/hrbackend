package handler

import (
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
	BaseAmount            float64   `json:"base_amount"`
	PremiumAmount         float64   `json:"premium_amount"`
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
