package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const leaveDateLayout = "2006-01-02"

type createLeaveRequestRequest struct {
	LeaveType string  `json:"leave_type" binding:"required,oneof=vacation personal sick pregnancy unpaid other"`
	StartDate string  `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate   string  `json:"end_date" binding:"required,datetime=2006-01-02"`
	Reason    *string `json:"reason"`
}

type createLeaveRequestByAdminRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	LeaveType  string    `json:"leave_type" binding:"required,oneof=vacation personal sick pregnancy unpaid other"`
	StartDate  string    `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate    string    `json:"end_date" binding:"required,datetime=2006-01-02"`
	Reason     *string   `json:"reason"`
}

type updateLeaveRequestRequest struct {
	LeaveType *string `json:"leave_type" binding:"omitempty,oneof=vacation personal sick pregnancy unpaid other"`
	StartDate *string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate   *string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Reason    *string `json:"reason"`
}

type updateLeaveRequestByAdminRequest struct {
	LeaveType       *string `json:"leave_type" binding:"omitempty,oneof=vacation personal sick pregnancy unpaid other"`
	StartDate       *string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate         *string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Reason          *string `json:"reason"`
	AdminUpdateNote string  `json:"admin_update_note" binding:"required"`
}

type decideLeaveRequestByAdminRequest struct {
	Decision     string  `json:"decision" binding:"required,oneof=approve reject"`
	DecisionNote *string `json:"decision_note"`
}

type listMyLeaveRequestsRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected cancelled expired"`
}

type listLeaveRequestsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status" binding:"omitempty,oneof=pending approved rejected cancelled expired"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type listLeaveBalancesRequest struct {
	httpapi.PageRequest
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
	Year           *int32  `form:"year" binding:"omitempty,min=2000,max=2100"`
}

type listMyLeaveBalancesRequest struct {
	httpapi.PageRequest
	Year *int32 `form:"year" binding:"omitempty,min=2000,max=2100"`
}

type adjustLeaveBalanceRequest struct {
	EmployeeID      uuid.UUID `json:"employee_id" binding:"required"`
	Year            int32     `json:"year" binding:"required,min=2000,max=2100"`
	LegalHoursDelta int32     `json:"legal_hours_delta"`
	ExtraHoursDelta int32     `json:"extra_hours_delta"`
	Reason          string    `json:"reason" binding:"required"`
}

type leaveRequestResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	CreatedByEmployeeID *uuid.UUID `json:"created_by_employee_id,omitempty"`
	LeaveType           string     `json:"leave_type"`
	Status              string     `json:"status"`
	StartDate           time.Time  `json:"start_date"`
	EndDate             time.Time  `json:"end_date"`
	Reason              *string    `json:"reason,omitempty"`
	DecisionNote        *string    `json:"decision_note,omitempty"`
	DecidedByEmployeeID *uuid.UUID `json:"decided_by_employee_id,omitempty"`
	RequestedAt         time.Time  `json:"requested_at"`
	DecidedAt           *time.Time `json:"decided_at,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type leaveRequestListItemResponse struct {
	leaveRequestResponse
	EmployeeName string `json:"employee_name"`
}

type leaveRequestStatsResponse struct {
	OpenRequests     int64 `json:"open_requests"`
	ApprovedRequests int64 `json:"approved_requests"`
	RejectedRequests int64 `json:"rejected_requests"`
	SicknessAbsence  int64 `json:"sickness_absence"`
}

type leaveBalanceResponse struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	Year            int32     `json:"year"`
	LegalTotalHours int32     `json:"legal_total_hours"`
	ExtraTotalHours int32     `json:"extra_total_hours"`
	LegalUsedHours  int32     `json:"legal_used_hours"`
	ExtraUsedHours  int32     `json:"extra_used_hours"`
	LegalRemaining  int32     `json:"legal_remaining_hours"`
	ExtraRemaining  int32     `json:"extra_remaining_hours"`
	TotalRemaining  int32     `json:"total_remaining_hours"`
	DerivedDayHours int32     `json:"derived_day_hours"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type leaveBudgetTypeResponse struct {
	TotalHours     int32 `json:"total_hours"`
	UsedHours      int32 `json:"used_hours"`
	RemainingHours int32 `json:"remaining_hours"`
}

type leaveBudgetByTypeResponse struct {
	Legal  leaveBudgetTypeResponse `json:"legal"`
	Budget leaveBudgetTypeResponse `json:"budget"`
}

type leaveContractDetailsResponse struct {
	ContractHours     *float64   `json:"contract_hours,omitempty"`
	ContractType      *string    `json:"contract_type,omitempty"`
	ContractStartDate *time.Time `json:"contract_start_date,omitempty"`
	ContractEndDate   *time.Time `json:"contract_end_date,omitempty"`
}

type managerLeaveBalanceResponse struct {
	ID           uuid.UUID                    `json:"id"`
	EmployeeID   uuid.UUID                    `json:"employee_id"`
	EmployeeName string                       `json:"employee_name"`
	Year         int32                        `json:"year"`
	LeaveBudget  leaveBudgetByTypeResponse    `json:"leave_budget"`
	Contract     leaveContractDetailsResponse `json:"contract"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
}

type adjustLeaveBalanceResponse struct {
	Balance leaveBalanceResponse `json:"balance"`
}

func toCreateLeaveRequestParams(req createLeaveRequestRequest) (domain.CreateLeaveRequestParams, error) {
	startDate, err := time.Parse(leaveDateLayout, req.StartDate)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	endDate, err := time.Parse(leaveDateLayout, req.EndDate)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	return domain.CreateLeaveRequestParams{
		LeaveType: strings.TrimSpace(req.LeaveType),
		StartDate: startDate.UTC(),
		EndDate:   endDate.UTC(),
		Reason:    req.Reason,
	}, nil
}

func toCreateLeaveRequestByAdminParams(req createLeaveRequestByAdminRequest) (domain.CreateLeaveRequestParams, error) {
	base, err := toCreateLeaveRequestParams(createLeaveRequestRequest{
		LeaveType: req.LeaveType,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	})
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	base.EmployeeID = req.EmployeeID
	return base, nil
}

func toUpdateLeaveRequestParams(req updateLeaveRequestRequest) (domain.UpdateLeaveRequestParams, error) {
	startDate, err := parseLeaveDatePtr(req.StartDate)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}
	endDate, err := parseLeaveDatePtr(req.EndDate)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}

	return domain.UpdateLeaveRequestParams{
		LeaveType: req.LeaveType,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
	}, nil
}

func toUpdateLeaveRequestByAdminParams(req updateLeaveRequestByAdminRequest) (domain.UpdateLeaveRequestParams, string, error) {
	updateParams, err := toUpdateLeaveRequestParams(updateLeaveRequestRequest{
		LeaveType: req.LeaveType,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	})
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, "", err
	}
	return updateParams, req.AdminUpdateNote, nil
}

func toDecideLeaveRequestParams(req decideLeaveRequestByAdminRequest) domain.DecideLeaveRequestParams {
	return domain.DecideLeaveRequestParams{
		Decision:     req.Decision,
		DecisionNote: req.DecisionNote,
	}
}

func toListMyLeaveRequestsParams(employeeID uuid.UUID, req listMyLeaveRequestsRequest) domain.ListMyLeaveRequestsParams {
	return domain.ListMyLeaveRequestsParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Status:     req.Status,
	}
}

func toListLeaveRequestsParams(req listLeaveRequestsRequest) domain.ListLeaveRequestsParams {
	return domain.ListLeaveRequestsParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		Status:         req.Status,
		EmployeeSearch: req.EmployeeSearch,
	}
}

func toListLeaveBalancesParams(req listLeaveBalancesRequest) domain.ListLeaveBalancesParams {
	return domain.ListLeaveBalancesParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		EmployeeSearch: req.EmployeeSearch,
		Year:           req.Year,
	}
}

func toListMyLeaveBalancesParams(employeeID uuid.UUID, req listMyLeaveBalancesRequest) domain.ListMyLeaveBalancesParams {
	return domain.ListMyLeaveBalancesParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Year:       req.Year,
	}
}

func toAdjustLeaveBalanceParams(adminEmployeeID uuid.UUID, req adjustLeaveBalanceRequest) domain.AdjustLeaveBalanceParams {
	return domain.AdjustLeaveBalanceParams{
		AdminEmployeeID: adminEmployeeID,
		EmployeeID:      req.EmployeeID,
		Year:            req.Year,
		LegalHoursDelta: req.LegalHoursDelta,
		ExtraHoursDelta: req.ExtraHoursDelta,
		Reason:          req.Reason,
	}
}

func toLeaveRequestResponse(item *domain.LeaveRequest) leaveRequestResponse {
	return leaveRequestResponse{
		ID:                  item.ID,
		EmployeeID:          item.EmployeeID,
		CreatedByEmployeeID: item.CreatedByEmployeeID,
		LeaveType:           item.LeaveType,
		Status:              item.Status,
		StartDate:           item.StartDate,
		EndDate:             item.EndDate,
		Reason:              item.Reason,
		DecisionNote:        item.DecisionNote,
		DecidedByEmployeeID: item.DecidedByEmployeeID,
		RequestedAt:         item.RequestedAt,
		DecidedAt:           item.DecidedAt,
		CancelledAt:         item.CancelledAt,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func toLeaveRequestListItemResponse(item domain.LeaveRequestListItem) leaveRequestListItemResponse {
	return leaveRequestListItemResponse{
		leaveRequestResponse: toLeaveRequestResponse(&item.LeaveRequest),
		EmployeeName:         item.EmployeeName,
	}
}

func toLeaveRequestStatsResponse(stats *domain.LeaveRequestStats) leaveRequestStatsResponse {
	return leaveRequestStatsResponse{
		OpenRequests:     stats.OpenRequests,
		ApprovedRequests: stats.ApprovedRequests,
		RejectedRequests: stats.RejectedRequests,
		SicknessAbsence:  stats.SicknessAbsence,
	}
}

func toLeaveBalanceResponse(item domain.LeaveBalance) leaveBalanceResponse {
	return leaveBalanceResponse{
		ID:              item.ID,
		EmployeeID:      item.EmployeeID,
		EmployeeName:    item.EmployeeName,
		Year:            item.Year,
		LegalTotalHours: item.LegalTotalHours,
		ExtraTotalHours: item.ExtraTotalHours,
		LegalUsedHours:  item.LegalUsedHours,
		ExtraUsedHours:  item.ExtraUsedHours,
		LegalRemaining:  item.LegalRemaining,
		ExtraRemaining:  item.ExtraRemaining,
		TotalRemaining:  item.TotalRemaining,
		DerivedDayHours: item.DerivedDayHours,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func toManagerLeaveBalanceResponse(item domain.LeaveBalance) managerLeaveBalanceResponse {
	return managerLeaveBalanceResponse{
		ID:           item.ID,
		EmployeeID:   item.EmployeeID,
		EmployeeName: item.EmployeeName,
		Year:         item.Year,
		LeaveBudget: leaveBudgetByTypeResponse{
			Legal: leaveBudgetTypeResponse{
				TotalHours:     item.LegalTotalHours,
				UsedHours:      item.LegalUsedHours,
				RemainingHours: item.LegalRemaining,
			},
			Budget: leaveBudgetTypeResponse{
				TotalHours:     item.ExtraTotalHours,
				UsedHours:      item.ExtraUsedHours,
				RemainingHours: item.ExtraRemaining,
			},
		},
		Contract: leaveContractDetailsResponse{
			ContractHours:     item.ContractHours,
			ContractType:      item.ContractType,
			ContractStartDate: item.ContractStartDate,
			ContractEndDate:   item.ContractEndDate,
		},
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func parseLeaveDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(leaveDateLayout, *value)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
