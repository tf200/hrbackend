package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const expenseDateLayout = "2006-01-02"

type createExpenseRequestByAdminRequest struct {
	EmployeeID      uuid.UUID `json:"employee_id" binding:"required"`
	Category        string    `json:"category" binding:"required,oneof=travel meal accommodation office_supplies training client_entertainment other"`
	ExpenseDate     string    `json:"expense_date" binding:"required,datetime=2006-01-02"`
	MerchantName    *string   `json:"merchant_name"`
	Description     string    `json:"description" binding:"required"`
	BusinessPurpose string    `json:"business_purpose" binding:"required"`
	Currency        string    `json:"currency" binding:"required,len=3"`
	ClaimedAmount   float64   `json:"claimed_amount" binding:"required"`
	TravelMode      *string   `json:"travel_mode"`
	TravelFrom      *string   `json:"travel_from"`
	TravelTo        *string   `json:"travel_to"`
	DistanceKm      *float64  `json:"distance_km"`
	RequestNote     *string   `json:"request_note"`
}

type updateExpenseRequestByAdminRequest struct {
	Category        *string  `json:"category" binding:"omitempty,oneof=travel meal accommodation office_supplies training client_entertainment other"`
	ExpenseDate     *string  `json:"expense_date" binding:"omitempty,datetime=2006-01-02"`
	MerchantName    *string  `json:"merchant_name"`
	Description     *string  `json:"description"`
	BusinessPurpose *string  `json:"business_purpose"`
	Currency        *string  `json:"currency" binding:"omitempty,len=3"`
	ClaimedAmount   *float64 `json:"claimed_amount"`
	TravelMode      *string  `json:"travel_mode"`
	TravelFrom      *string  `json:"travel_from"`
	TravelTo        *string  `json:"travel_to"`
	DistanceKm      *float64 `json:"distance_km"`
	RequestNote     *string  `json:"request_note"`
}

type decideExpenseRequestByAdminRequest struct {
	Decision       string   `json:"decision" binding:"required,oneof=approve reject"`
	ApprovedAmount *float64 `json:"approved_amount"`
	DecisionNote   *string  `json:"decision_note"`
}

type listExpenseRequestsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status" binding:"omitempty,oneof=pending approved rejected reimbursed cancelled"`
	Category       *string `form:"category" binding:"omitempty,oneof=travel meal accommodation office_supplies training client_entertainment other"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type expenseRequestResponse struct {
	ID                     uuid.UUID  `json:"id"`
	EmployeeID             uuid.UUID  `json:"employee_id"`
	EmployeeName           string     `json:"employee_name,omitempty"`
	CreatedByEmployeeID    uuid.UUID  `json:"created_by_employee_id"`
	Category               string     `json:"category"`
	ExpenseDate            time.Time  `json:"expense_date"`
	MerchantName           *string    `json:"merchant_name,omitempty"`
	Description            string     `json:"description"`
	BusinessPurpose        string     `json:"business_purpose"`
	Currency               string     `json:"currency"`
	ClaimedAmount          float64    `json:"claimed_amount"`
	ApprovedAmount         *float64   `json:"approved_amount,omitempty"`
	TravelMode             *string    `json:"travel_mode,omitempty"`
	TravelFrom             *string    `json:"travel_from,omitempty"`
	TravelTo               *string    `json:"travel_to,omitempty"`
	DistanceKm             *float64   `json:"distance_km,omitempty"`
	Status                 string     `json:"status"`
	RequestNote            *string    `json:"request_note,omitempty"`
	DecisionNote           *string    `json:"decision_note,omitempty"`
	DecidedByEmployeeID    *uuid.UUID `json:"decided_by_employee_id,omitempty"`
	ReimbursedByEmployeeID *uuid.UUID `json:"reimbursed_by_employee_id,omitempty"`
	RequestedAt            time.Time  `json:"requested_at"`
	DecidedAt              *time.Time `json:"decided_at,omitempty"`
	ReimbursedAt           *time.Time `json:"reimbursed_at,omitempty"`
	CancelledAt            *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func toCreateExpenseRequestByAdminParams(
	req createExpenseRequestByAdminRequest,
) (domain.CreateExpenseRequestParams, error) {
	expenseDate, err := time.Parse(expenseDateLayout, req.ExpenseDate)
	if err != nil {
		return domain.CreateExpenseRequestParams{}, err
	}

	return domain.CreateExpenseRequestParams{
		EmployeeID:      req.EmployeeID,
		Category:        strings.TrimSpace(req.Category),
		ExpenseDate:     expenseDate.UTC(),
		MerchantName:    req.MerchantName,
		Description:     req.Description,
		BusinessPurpose: req.BusinessPurpose,
		Currency:        strings.TrimSpace(req.Currency),
		ClaimedAmount:   req.ClaimedAmount,
		TravelMode:      req.TravelMode,
		TravelFrom:      req.TravelFrom,
		TravelTo:        req.TravelTo,
		DistanceKm:      req.DistanceKm,
		RequestNote:     req.RequestNote,
	}, nil
}

func toUpdateExpenseRequestByAdminParams(
	req updateExpenseRequestByAdminRequest,
) (domain.UpdateExpenseRequestParams, error) {
	expenseDate, err := parseExpenseDatePtr(req.ExpenseDate)
	if err != nil {
		return domain.UpdateExpenseRequestParams{}, err
	}

	return domain.UpdateExpenseRequestParams{
		Category:        req.Category,
		ExpenseDate:     expenseDate,
		MerchantName:    req.MerchantName,
		Description:     req.Description,
		BusinessPurpose: req.BusinessPurpose,
		Currency:        req.Currency,
		ClaimedAmount:   req.ClaimedAmount,
		TravelMode:      req.TravelMode,
		TravelFrom:      req.TravelFrom,
		TravelTo:        req.TravelTo,
		DistanceKm:      req.DistanceKm,
		RequestNote:     req.RequestNote,
	}, nil
}

func toDecideExpenseRequestParams(
	req decideExpenseRequestByAdminRequest,
) domain.DecideExpenseRequestParams {
	return domain.DecideExpenseRequestParams{
		Decision:       req.Decision,
		ApprovedAmount: req.ApprovedAmount,
		DecisionNote:   req.DecisionNote,
	}
}

func toListExpenseRequestsParams(req listExpenseRequestsRequest) domain.ListExpenseRequestsParams {
	page := req.PageRequest.Params()
	return domain.ListExpenseRequestsParams{
		Limit:          page.Limit,
		Offset:         page.Offset,
		Status:         req.Status,
		Category:       req.Category,
		EmployeeSearch: req.EmployeeSearch,
	}
}

func toExpenseRequestResponse(item domain.ExpenseRequest) expenseRequestResponse {
	return expenseRequestResponse{
		ID:                     item.ID,
		EmployeeID:             item.EmployeeID,
		EmployeeName:           item.EmployeeName,
		CreatedByEmployeeID:    item.CreatedByEmployeeID,
		Category:               item.Category,
		ExpenseDate:            item.ExpenseDate,
		MerchantName:           item.MerchantName,
		Description:            item.Description,
		BusinessPurpose:        item.BusinessPurpose,
		Currency:               item.Currency,
		ClaimedAmount:          item.ClaimedAmount,
		ApprovedAmount:         item.ApprovedAmount,
		TravelMode:             item.TravelMode,
		TravelFrom:             item.TravelFrom,
		TravelTo:               item.TravelTo,
		DistanceKm:             item.DistanceKm,
		Status:                 item.Status,
		RequestNote:            item.RequestNote,
		DecisionNote:           item.DecisionNote,
		DecidedByEmployeeID:    item.DecidedByEmployeeID,
		ReimbursedByEmployeeID: item.ReimbursedByEmployeeID,
		RequestedAt:            item.RequestedAt,
		DecidedAt:              item.DecidedAt,
		ReimbursedAt:           item.ReimbursedAt,
		CancelledAt:            item.CancelledAt,
		CreatedAt:              item.CreatedAt,
		UpdatedAt:              item.UpdatedAt,
	}
}

func toExpenseRequestResponses(items []domain.ExpenseRequest) []expenseRequestResponse {
	results := make([]expenseRequestResponse, len(items))
	for i, item := range items {
		results[i] = toExpenseRequestResponse(item)
	}
	return results
}

func parseExpenseDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(expenseDateLayout, *value)
	if err != nil {
		return nil, err
	}
	t := parsed.UTC()
	return &t, nil
}
