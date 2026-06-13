package handler

import (
	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type employeeDashboardPendingRequestsRequest struct {
	Days  int32 `form:"days"  binding:"omitempty,min=1,max=180"`
	Limit int32 `form:"limit" binding:"omitempty,min=1,max=20"`
}

type employeeDashboardKPIsResponse struct {
	LeaveBalance         employeeDashboardLeaveBalanceResponse `json:"leave_balance"`
	PendingLeaveRequests int64                                 `json:"pending_leave_requests"`
	EstimatedCurrentPay  employeeDashboardPayResponse          `json:"estimated_current_pay"`
	PendingSignatures    int64                                 `json:"pending_signatures"`
}

type employeeDashboardLeaveBalanceResponse struct {
	Year         int32   `json:"year"`
	UsedMinutes  int32   `json:"used_minutes"`
	TotalMinutes int32   `json:"total_minutes"`
	UsedHours    float64 `json:"used_hours"`
	TotalHours   float64 `json:"total_hours"`
}

type employeeDashboardPayResponse struct {
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	GrossAmount float64 `json:"gross_amount"`
	DataSource  string  `json:"data_source"`
}

type employeeDashboardPendingRequestsResponse struct {
	Results []employeeDashboardPendingRequestResponse `json:"results"`
}

type employeeDashboardPendingRequestResponse struct {
	ID              uuid.UUID `json:"id"`
	RequestType     string    `json:"request_type"`
	Status          string    `json:"status"`
	SubmittedAt     string    `json:"submitted_at"`
	RequestDate     string    `json:"request_date"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	DurationMinutes *int32    `json:"duration_minutes,omitempty"`
	Amount          *float64  `json:"amount,omitempty"`
	Currency        *string   `json:"currency,omitempty"`
}

func toEmployeeDashboardKPIsResponse(
	kpi *domain.EmployeeDashboardKPI,
) employeeDashboardKPIsResponse {
	return employeeDashboardKPIsResponse{
		LeaveBalance: employeeDashboardLeaveBalanceResponse{
			Year:         kpi.LeaveBalance.Year,
			UsedMinutes:  kpi.LeaveBalance.UsedMinutes,
			TotalMinutes: kpi.LeaveBalance.TotalMinutes,
			UsedHours:    roundCurrency(float64(kpi.LeaveBalance.UsedMinutes) / 60),
			TotalHours:   roundCurrency(float64(kpi.LeaveBalance.TotalMinutes) / 60),
		},
		PendingLeaveRequests: kpi.PendingLeaveRequests,
		EstimatedCurrentPay: employeeDashboardPayResponse{
			PeriodStart: kpi.EstimatedCurrentPay.PeriodStart.Format(leaveDateLayout),
			PeriodEnd:   kpi.EstimatedCurrentPay.PeriodEnd.Format(leaveDateLayout),
			GrossAmount: roundCurrency(kpi.EstimatedCurrentPay.GrossAmount),
			DataSource:  kpi.EstimatedCurrentPay.DataSource,
		},
		PendingSignatures: kpi.PendingSignatures,
	}
}

func toEmployeeDashboardPendingRequestsResponse(
	items []domain.EmployeeDashboardPendingRequest,
) employeeDashboardPendingRequestsResponse {
	results := make([]employeeDashboardPendingRequestResponse, len(items))
	for i, item := range items {
		results[i] = employeeDashboardPendingRequestResponse{
			ID:              item.ID,
			RequestType:     item.RequestType,
			Status:          item.Status,
			SubmittedAt:     item.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
			RequestDate:     item.RequestDate.Format(leaveDateLayout),
			Title:           item.Title,
			Description:     item.Description,
			DurationMinutes: item.DurationMinutes,
			Amount:          item.Amount,
			Currency:        item.Currency,
		}
	}
	return employeeDashboardPendingRequestsResponse{Results: results}
}
