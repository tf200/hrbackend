package handler

import (
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const leaveDateLayout = "2006-01-02"
const leaveMonthLayout = "2006-01"

type createLeaveRequestRequest struct {
	LeaveType    string  `json:"leave_type"    binding:"required,oneof=vacation personal sick pregnancy unpaid other"`
	DurationType string  `json:"duration_type" binding:"required,oneof=full_day hours"`
	StartDate    string  `json:"start_date"    binding:"required,datetime=2006-01-02"`
	EndDate      string  `json:"end_date"      binding:"required,datetime=2006-01-02"`
	StartTime    *string `json:"start_time"    binding:"omitempty,datetime=15:04"`
	EndTime      *string `json:"end_time"      binding:"omitempty,datetime=15:04"`
	Reason       *string `json:"reason"`
}

type createLeaveRequestByAdminRequest struct {
	EmployeeID   uuid.UUID `json:"employee_id"  binding:"required"`
	LeaveType    string    `json:"leave_type"    binding:"required,oneof=vacation personal sick pregnancy unpaid other"`
	DurationType string    `json:"duration_type" binding:"required,oneof=full_day hours"`
	StartDate    string    `json:"start_date"    binding:"required,datetime=2006-01-02"`
	EndDate      string    `json:"end_date"      binding:"required,datetime=2006-01-02"`
	StartTime    *string   `json:"start_time"    binding:"omitempty,datetime=15:04"`
	EndTime      *string   `json:"end_time"      binding:"omitempty,datetime=15:04"`
	Reason       *string   `json:"reason"`
}

type updateLeaveRequestRequest struct {
	LeaveType    *string `json:"leave_type"    binding:"omitempty,oneof=vacation personal sick pregnancy unpaid other"`
	DurationType *string `json:"duration_type" binding:"omitempty,oneof=full_day hours"`
	StartDate    *string `json:"start_date"    binding:"omitempty,datetime=2006-01-02"`
	EndDate      *string `json:"end_date"      binding:"omitempty,datetime=2006-01-02"`
	StartTime    *string `json:"start_time"    binding:"omitempty,datetime=15:04"`
	EndTime      *string `json:"end_time"      binding:"omitempty,datetime=15:04"`
	Reason       *string `json:"reason"`
}

type updateLeaveRequestByAdminRequest struct {
	LeaveType       *string `json:"leave_type"        binding:"omitempty,oneof=vacation personal sick pregnancy unpaid other"`
	DurationType    *string `json:"duration_type"     binding:"omitempty,oneof=full_day hours"`
	StartDate       *string `json:"start_date"        binding:"omitempty,datetime=2006-01-02"`
	EndDate         *string `json:"end_date"          binding:"omitempty,datetime=2006-01-02"`
	StartTime       *string `json:"start_time"        binding:"omitempty,datetime=15:04"`
	EndTime         *string `json:"end_time"          binding:"omitempty,datetime=15:04"`
	Reason          *string `json:"reason"`
	AdminUpdateNote string  `json:"admin_update_note" binding:"required"`
}

type decideLeaveRequestByAdminRequest struct {
	Decision     string  `json:"decision"      binding:"required,oneof=approve reject"`
	DecisionNote *string `json:"decision_note"`
}

type listMyLeaveRequestsRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected cancelled expired"`
}

type listLeaveRequestsRequest struct {
	httpapi.PageRequest
	Status         *string `form:"status"          binding:"omitempty,oneof=pending approved rejected cancelled expired"`
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
}

type listLeaveCalendarRequest struct {
	Month          string     `form:"month"                                               binding:"required,datetime=2006-01"`
	DepartmentID   *uuid.UUID `form:"department_id,parser=encoding.TextUnmarshaler"`
	LeaveTypes     []string   `form:"leave_types"`
	EmployeeSearch *string    `form:"employee_search"                                     binding:"omitempty,max=120"`
}

type listLeaveBalancesRequest struct {
	httpapi.PageRequest
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
	Year           *int32  `form:"year"            binding:"omitempty,min=2000,max=2100"`
}

type getLeaveBalanceDetailsRequest struct {
	Year int32 `form:"year" binding:"omitempty,min=2000,max=2100"`
}

type leaveRequestResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	CreatedByEmployeeID *uuid.UUID `json:"created_by_employee_id,omitempty"`
	LeaveType           string     `json:"leave_type"`
	Status              string     `json:"status"`
	DurationType        string     `json:"duration_type"`
	RequestedMinutes    int32      `json:"requested_minutes"`
	StartTime           *string    `json:"start_time,omitempty"`
	EndTime             *string    `json:"end_time,omitempty"`
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

type leaveCalendarRecordResponse struct {
	LeaveRequestID   uuid.UUID `json:"leave_request_id"`
	LeaveType        string    `json:"leave_type"`
	Status           string    `json:"status"`
	DurationType     string    `json:"duration_type"`
	RequestedMinutes int32     `json:"requested_minutes"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Reason           *string   `json:"reason,omitempty"`
}

type leaveCalendarEmployeeResponse struct {
	EmployeeID     uuid.UUID                     `json:"employee_id"`
	EmployeeName   string                        `json:"employee_name"`
	DepartmentName *string                       `json:"department_name,omitempty"`
	LeaveRecords   []leaveCalendarRecordResponse `json:"leave_records"`
}

type leaveRequestStatsResponse struct {
	OpenRequests     int64 `json:"open_requests"`
	ApprovedRequests int64 `json:"approved_requests"`
	RejectedRequests int64 `json:"rejected_requests"`
	SicknessAbsence  int64 `json:"sickness_absence"`
}

type deductedLeaveResponse struct {
	ID              uuid.UUID `json:"id"`
	LeaveType       string    `json:"leave_type"`
	StartDate       string    `json:"start_date"`
	EndDate         string    `json:"end_date"`
	DurationMinutes int32     `json:"duration_minutes"`
}

type leaveBalanceResponse struct {
	EmployeeID            uuid.UUID               `json:"employee_id"`
	EmployeeName          string                  `json:"employee_name"`
	Year                  int32                   `json:"year"`
	LegalTotalMinutes     int32                   `json:"legal_total_minutes"`
	LegalUsedMinutes      int32                   `json:"legal_used_minutes"`
	LegalRemainingMinutes int32                   `json:"legal_remaining_minutes"`
	TotalRemainingMinutes int32                   `json:"total_remaining_minutes"`
	DeductedLeaves        []deductedLeaveResponse `json:"deducted_leaves"`
}

type leaveBudgetTypeResponse struct {
	TotalMinutes     int32 `json:"total_minutes"`
	UsedMinutes      int32 `json:"used_minutes"`
	RemainingMinutes int32 `json:"remaining_minutes"`
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
	EffectiveEndDate  *time.Time `json:"effective_end_date,omitempty"`
}

type managerLeaveBalanceResponse struct {
	EmployeeID   uuid.UUID                    `json:"employee_id"`
	EmployeeName string                       `json:"employee_name"`
	Year         int32                        `json:"year"`
	LeaveBudget  leaveBudgetByTypeResponse    `json:"leave_budget"`
	Contract     leaveContractDetailsResponse `json:"contract"`
}

type leaveContractAccrualResponse struct {
	ContractID        uuid.UUID  `json:"contract_id"`
	ContractType      string     `json:"contract_type"`
	ContractHours     *float64   `json:"contract_hours,omitempty"`
	ContractStartDate time.Time  `json:"contract_start_date"`
	ContractEndDate   *time.Time `json:"contract_end_date,omitempty"`
	EffectiveEndDate  *time.Time `json:"effective_end_date,omitempty"`
	SegmentStartDate  time.Time  `json:"segment_start_date"`
	SegmentEndDate    time.Time  `json:"segment_end_date"`
	YearDays          int32      `json:"year_days"`
	SegmentDays       int32      `json:"segment_days"`
	FullYearMinutes   int32      `json:"full_year_minutes"`
	ScheduleMinutes   int32      `json:"schedule_minutes"`
	OvertimeMinutes   int32      `json:"overtime_minutes"`
	GainedMinutes     int32      `json:"gained_minutes"`
	DeductedMinutes   int32      `json:"deducted_minutes"`
	RemainingMinutes  int32      `json:"remaining_minutes"`
}

type leaveBalanceDetailsResponse struct {
	managerLeaveBalanceResponse
	ContractAccruals []leaveContractAccrualResponse `json:"contract_accruals"`
}

func toCreateLeaveRequestParams(
	req createLeaveRequestRequest,
) (domain.CreateLeaveRequestParams, error) {
	startDate, err := time.Parse(leaveDateLayout, req.StartDate)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	endDate, err := time.Parse(leaveDateLayout, req.EndDate)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	startTime, err := parseLeaveTimePtr(req.StartTime)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	endTime, err := parseLeaveTimePtr(req.EndTime)
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	return domain.CreateLeaveRequestParams{
		LeaveType:    strings.TrimSpace(req.LeaveType),
		DurationType: strings.TrimSpace(req.DurationType),
		StartDate:    startDate.UTC(),
		EndDate:      endDate.UTC(),
		StartTime:    startTime,
		EndTime:      endTime,
		Reason:       req.Reason,
	}, nil
}

func toCreateLeaveRequestByAdminParams(
	req createLeaveRequestByAdminRequest,
) (domain.CreateLeaveRequestParams, error) {
	base, err := toCreateLeaveRequestParams(createLeaveRequestRequest{
		LeaveType:    req.LeaveType,
		DurationType: req.DurationType,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Reason:       req.Reason,
	})
	if err != nil {
		return domain.CreateLeaveRequestParams{}, err
	}
	base.EmployeeID = req.EmployeeID
	return base, nil
}

func toUpdateLeaveRequestParams(
	req updateLeaveRequestRequest,
) (domain.UpdateLeaveRequestParams, error) {
	startDate, err := parseLeaveDatePtr(req.StartDate)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}
	endDate, err := parseLeaveDatePtr(req.EndDate)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}

	startTime, err := parseLeaveTimePtr(req.StartTime)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}
	endTime, err := parseLeaveTimePtr(req.EndTime)
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, err
	}

	return domain.UpdateLeaveRequestParams{
		LeaveType:    req.LeaveType,
		DurationType: req.DurationType,
		StartDate:    startDate,
		EndDate:      endDate,
		StartTime:    startTime,
		EndTime:      endTime,
		Reason:       req.Reason,
	}, nil
}

func toUpdateLeaveRequestByAdminParams(
	req updateLeaveRequestByAdminRequest,
) (domain.UpdateLeaveRequestParams, string, error) {
	updateParams, err := toUpdateLeaveRequestParams(updateLeaveRequestRequest{
		LeaveType:    req.LeaveType,
		DurationType: req.DurationType,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Reason:       req.Reason,
	})
	if err != nil {
		return domain.UpdateLeaveRequestParams{}, "", err
	}
	return updateParams, req.AdminUpdateNote, nil
}

func toDecideLeaveRequestParams(
	req decideLeaveRequestByAdminRequest,
) domain.DecideLeaveRequestParams {
	return domain.DecideLeaveRequestParams{
		Decision:     req.Decision,
		DecisionNote: req.DecisionNote,
	}
}

func toListMyLeaveRequestsParams(
	employeeID uuid.UUID,
	req listMyLeaveRequestsRequest,
) domain.ListMyLeaveRequestsParams {
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

func bindListLeaveCalendarRequest(ctx *gin.Context) (listLeaveCalendarRequest, error) {
	var req listLeaveCalendarRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return listLeaveCalendarRequest{}, err
	}
	req.LeaveTypes = ctx.QueryArray("leave_types")
	if err := req.validate(); err != nil {
		return listLeaveCalendarRequest{}, err
	}
	return req, nil
}

func toListLeaveCalendarParams(req listLeaveCalendarRequest) domain.ListLeaveCalendarParams {
	month, err := time.Parse(leaveMonthLayout, req.Month)
	if err != nil {
		return domain.ListLeaveCalendarParams{}
	}

	return domain.ListLeaveCalendarParams{
		Month:          time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC),
		DepartmentID:   req.DepartmentID,
		LeaveTypes:     req.LeaveTypes,
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

func toGetLeaveBalanceDetailsParams(
	employeeID uuid.UUID,
	req getLeaveBalanceDetailsRequest,
) domain.GetLeaveBalanceDetailsParams {
	return domain.GetLeaveBalanceDetailsParams{
		EmployeeID: employeeID,
		Year:       req.Year,
	}
}

func toLeaveRequestResponse(item *domain.LeaveRequest) leaveRequestResponse {
	return leaveRequestResponse{
		ID:                  item.ID,
		EmployeeID:          item.EmployeeID,
		CreatedByEmployeeID: item.CreatedByEmployeeID,
		LeaveType:           item.LeaveType,
		Status:              item.Status,
		DurationType:        item.DurationType,
		RequestedMinutes:    item.RequestedMinutes,
		StartTime:           formatLeaveTimePtr(item.StartTime),
		EndTime:             formatLeaveTimePtr(item.EndTime),
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

func toLeaveCalendarEmployeeResponse(
	item domain.LeaveCalendarEmployee,
) leaveCalendarEmployeeResponse {
	records := make([]leaveCalendarRecordResponse, len(item.LeaveRecords))
	for i, record := range item.LeaveRecords {
		records[i] = leaveCalendarRecordResponse{
			LeaveRequestID:   record.LeaveRequestID,
			LeaveType:        record.LeaveType,
			Status:           record.Status,
			DurationType:     record.DurationType,
			RequestedMinutes: record.RequestedMinutes,
			StartDate:        record.StartDate,
			EndDate:          record.EndDate,
			Reason:           record.Reason,
		}
	}

	return leaveCalendarEmployeeResponse{
		EmployeeID:     item.EmployeeID,
		EmployeeName:   item.EmployeeName,
		DepartmentName: item.DepartmentName,
		LeaveRecords:   records,
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
	deductedLeaves := make([]deductedLeaveResponse, len(item.DeductedLeaves))
	for i, dl := range item.DeductedLeaves {
		deductedLeaves[i] = deductedLeaveResponse{
			ID:              dl.ID,
			LeaveType:       dl.LeaveType,
			StartDate:       dl.StartDate.Format("2006-01-02"),
			EndDate:         dl.EndDate.Format("2006-01-02"),
			DurationMinutes: dl.DurationMinutes,
		}
	}

	return leaveBalanceResponse{
		EmployeeID:            item.EmployeeID,
		EmployeeName:          item.EmployeeName,
		Year:                  item.Year,
		LegalTotalMinutes:     item.LegalTotalMinutes,
		LegalUsedMinutes:      item.LegalUsedMinutes,
		LegalRemainingMinutes: item.LegalRemainingMinutes,
		TotalRemainingMinutes: item.TotalRemainingMinutes,
		DeductedLeaves:        deductedLeaves,
	}
}

func toManagerLeaveBalanceResponse(item domain.LeaveBalance) managerLeaveBalanceResponse {
	return managerLeaveBalanceResponse{
		EmployeeID:   item.EmployeeID,
		EmployeeName: item.EmployeeName,
		Year:         item.Year,
		LeaveBudget: leaveBudgetByTypeResponse{
			Legal: leaveBudgetTypeResponse{
				TotalMinutes:     item.LegalTotalMinutes,
				UsedMinutes:      item.LegalUsedMinutes,
				RemainingMinutes: item.LegalRemainingMinutes,
			},
			Budget: leaveBudgetTypeResponse{},
		},
		Contract: leaveContractDetailsResponse{
			ContractHours:     item.ContractHours,
			ContractType:      item.ContractType,
			ContractStartDate: item.ContractStartDate,
			ContractEndDate:   item.ContractEndDate,
			EffectiveEndDate:  item.EffectiveEndDate,
		},
	}
}

func toLeaveBalanceDetailsResponse(item domain.LeaveBalanceDetails) leaveBalanceDetailsResponse {
	accruals := make([]leaveContractAccrualResponse, len(item.ContractAccruals))
	for i, accrual := range item.ContractAccruals {
		accruals[i] = leaveContractAccrualResponse{
			ContractID:        accrual.ContractID,
			ContractType:      accrual.ContractType,
			ContractHours:     accrual.ContractHours,
			ContractStartDate: accrual.ContractStartDate,
			ContractEndDate:   accrual.ContractEndDate,
			EffectiveEndDate:  accrual.EffectiveEndDate,
			SegmentStartDate:  accrual.SegmentStartDate,
			SegmentEndDate:    accrual.SegmentEndDate,
			YearDays:          accrual.YearDays,
			SegmentDays:       accrual.SegmentDays,
			FullYearMinutes:   accrual.FullYearMinutes,
			ScheduleMinutes:   accrual.ScheduleMinutes,
			OvertimeMinutes:   accrual.OvertimeMinutes,
			GainedMinutes:     accrual.GainedMinutes,
			DeductedMinutes:   accrual.DeductedMinutes,
			RemainingMinutes:  accrual.RemainingMinutes,
		}
	}

	return leaveBalanceDetailsResponse{
		managerLeaveBalanceResponse: toManagerLeaveBalanceResponse(item.Balance),
		ContractAccruals:            accruals,
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

const leaveTimeLayout = "15:04"

func parseLeaveTimePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(leaveTimeLayout, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func formatLeaveTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(leaveTimeLayout)
	return &formatted
}

func (r listLeaveCalendarRequest) validate() error {
	for _, leaveType := range r.LeaveTypes {
		switch strings.TrimSpace(leaveType) {
		case "vacation", "personal", "sick", "pregnancy", "unpaid", "other":
		default:
			return fmt.Errorf("Field validation for 'LeaveTypes' failed on the 'oneof' tag")
		}
	}
	return nil
}
