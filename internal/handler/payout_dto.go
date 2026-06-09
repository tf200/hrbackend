package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const (
	payoutMonthLayout   = "2006-01"
	timeEntryDateLayout = "2006-01-02"
)

type createPayoutRequestRequest struct {
	RequestedHours int32   `json:"requested_hours" binding:"required,min=1"`
	BalanceYear    int32   `json:"balance_year"    binding:"required,min=2000,max=2100"`
	RequestNote    *string `json:"request_note"`
}

type createPayoutRequestByAdminRequest struct {
	EmployeeID     uuid.UUID `json:"employee_id"     binding:"required"`
	RequestedHours int32     `json:"requested_hours" binding:"required,min=1"`
	BalanceYear    int32     `json:"balance_year"    binding:"required,min=2000,max=2100"`
	PayPeriodStart string    `json:"pay_period_start" binding:"required,datetime=2006-01-02"`
	RequestNote    *string   `json:"request_note"`
	DecisionNote   *string   `json:"decision_note"`
}

type decidePayoutRequestByAdminRequest struct {
	Decision       string  `json:"decision"      binding:"required,oneof=approve reject"`
	DecisionNote   *string `json:"decision_note"`
	PayPeriodStart *string `json:"pay_period_start" binding:"omitempty,datetime=2006-01-02"`
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

type payoutRequestResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	EmployeeName        string     `json:"employee_name"`
	CreatedByEmployeeID uuid.UUID  `json:"created_by_employee_id"`
	RequestedHours      int32      `json:"requested_hours"`
	BalanceYear         int32      `json:"balance_year"`
	HourlyRate          float64    `json:"hourly_rate"`
	GrossAmount         float64    `json:"gross_amount"`
	PayPeriodStart      *string    `json:"pay_period_start,omitempty"`
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

func toCreatePayoutRequestParams(employeeID uuid.UUID, req createPayoutRequestRequest) domain.CreatePayoutRequestParams {
	return domain.CreatePayoutRequestParams{
		EmployeeID:          employeeID,
		CreatedByEmployeeID: employeeID,
		RequestedHours:      req.RequestedHours,
		BalanceYear:         req.BalanceYear,
		RequestNote:         req.RequestNote,
	}
}

func toCreatePayoutRequestByAdminParams(req createPayoutRequestByAdminRequest) (domain.CreatePayoutRequestByAdminParams, error) {
	payPeriodStart, err := time.Parse(timeEntryDateLayout, req.PayPeriodStart)
	if err != nil {
		return domain.CreatePayoutRequestByAdminParams{}, err
	}

	return domain.CreatePayoutRequestByAdminParams{
		EmployeeID:     req.EmployeeID,
		RequestedHours: req.RequestedHours,
		BalanceYear:    req.BalanceYear,
		PayPeriodStart: payPeriodStart,
		RequestNote:    req.RequestNote,
		DecisionNote:   req.DecisionNote,
	}, nil
}

func toDecidePayoutRequestParams(req decidePayoutRequestByAdminRequest) (domain.DecidePayoutRequestParams, error) {
	params := domain.DecidePayoutRequestParams{
		Decision:     req.Decision,
		DecisionNote: req.DecisionNote,
	}
	if req.PayPeriodStart != nil {
		parsed, err := time.Parse(timeEntryDateLayout, *req.PayPeriodStart)
		if err != nil {
			return domain.DecidePayoutRequestParams{}, err
		}
		params.PayPeriodStart = &parsed
	}
	return params, nil
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
		PayPeriodStart:      formatPayoutPayPeriodStart(item.PayPeriodStart),
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
	if items == nil {
		return nil
	}
	resp := make([]payoutRequestResponse, len(items))
	for i, item := range items {
		resp[i] = toPayoutRequestResponse(item)
	}
	return resp
}
