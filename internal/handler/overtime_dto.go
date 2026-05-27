package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
)

const overtimeDateLayout = "2006-01-02"

type createOvertimeEntryRequest struct {
	ScheduleID  *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate   string     `json:"entry_date"  binding:"required,datetime=2006-01-02"`
	Minutes     int32      `json:"minutes"     binding:"required,min=1"`
	Reason      string     `json:"reason"      binding:"required"`
	Description *string    `json:"description"`
}

type createOvertimeEntryByAdminRequest struct {
	EmployeeID  uuid.UUID  `json:"employee_id"  binding:"required"`
	ScheduleID  *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate   string     `json:"entry_date"   binding:"required,datetime=2006-01-02"`
	Minutes     int32      `json:"minutes"      binding:"required,min=1"`
	Reason      string     `json:"reason"       binding:"required"`
	Description *string    `json:"description"`
}

type decideOvertimeEntryByAdminRequest struct {
	Decision        string  `json:"decision"         binding:"required,oneof=approve reject"`
	RejectionReason *string `json:"rejection_reason"`
}

type updateOvertimeEntryByAdminRequest struct {
	ScheduleID  *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate   *string    `json:"entry_date"           binding:"omitempty,datetime=2006-01-02"`
	Minutes     *int32     `json:"minutes"              binding:"omitempty,min=1"`
	Reason      *string    `json:"reason"`
	Description *string    `json:"description"`
}

type updateMyOvertimeEntryRequest struct {
	ScheduleID  *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate   *string    `json:"entry_date"   binding:"omitempty,datetime=2006-01-02"`
	Minutes     *int32     `json:"minutes"      binding:"omitempty,min=1"`
	Reason      *string    `json:"reason"`
	Description *string    `json:"description"`
}

type listOvertimeEntriesRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=submitted approved rejected"`
}

type listMyOvertimeEntriesRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=submitted approved rejected"`
}

type overtimeEntryResponse struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	EmployeeName         string     `json:"employee_name"`
	ScheduleID           *uuid.UUID `json:"schedule_id,omitempty"`
	IsPaid               bool       `json:"is_paid"`
	EntryDate            time.Time  `json:"entry_date"`
	Minutes              int32      `json:"minutes"`
	Reason               string     `json:"reason"`
	Description          *string    `json:"description,omitempty"`
	Status               string     `json:"status"`
	SubmittedAt          time.Time  `json:"submitted_at"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	ApprovedByEmployeeID *uuid.UUID `json:"approved_by_employee_id,omitempty"`
	ApprovedByName       *string    `json:"approved_by_name,omitempty"`
	RejectionReason      *string    `json:"rejection_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type overtimeStatsResponse struct {
	TotalApprovedMinutes  int64 `json:"total_approved_minutes"`
	TotalAwaitingApproval int64 `json:"total_awaiting_approval"`
	TotalApproved         int64 `json:"total_approved"`
	TotalSubmitted        int64 `json:"total_submitted"`
}

func toCreateOvertimeEntryParams(req createOvertimeEntryRequest) (domain.CreateOvertimeEntryParams, error) {
	entryDate, err := time.Parse(overtimeDateLayout, req.EntryDate)
	if err != nil {
		return domain.CreateOvertimeEntryParams{}, err
	}

	return domain.CreateOvertimeEntryParams{
		ScheduleID:  req.ScheduleID,
		EntryDate:   entryDate.UTC(),
		Minutes:     req.Minutes,
		Reason:      strings.TrimSpace(req.Reason),
		Description: req.Description,
	}, nil
}

func toCreateOvertimeEntryByAdminParams(
	req createOvertimeEntryByAdminRequest,
) (domain.CreateOvertimeEntryParams, error) {
	base, err := toCreateOvertimeEntryParams(createOvertimeEntryRequest{
		ScheduleID:  req.ScheduleID,
		EntryDate:   req.EntryDate,
		Minutes:     req.Minutes,
		Reason:      req.Reason,
		Description: req.Description,
	})
	if err != nil {
		return domain.CreateOvertimeEntryParams{}, err
	}
	base.EmployeeID = req.EmployeeID
	return base, nil
}

func toDecideOvertimeEntryParams(req decideOvertimeEntryByAdminRequest) domain.DecideOvertimeEntryParams {
	return domain.DecideOvertimeEntryParams{
		Decision:        strings.TrimSpace(req.Decision),
		RejectionReason: ptr.TrimString(req.RejectionReason),
	}
}

func toUpdateOvertimeEntryByAdminParams(
	req updateOvertimeEntryByAdminRequest,
) (domain.UpdateOvertimeEntryParams, error) {
	return toUpdateMyOvertimeEntryParams(updateMyOvertimeEntryRequest{
		ScheduleID:  req.ScheduleID,
		EntryDate:   req.EntryDate,
		Minutes:     req.Minutes,
		Reason:      req.Reason,
		Description: req.Description,
	})
}

func toUpdateMyOvertimeEntryParams(
	req updateMyOvertimeEntryRequest,
) (domain.UpdateOvertimeEntryParams, error) {
	entryDate, err := parseOvertimeDatePtr(req.EntryDate)
	if err != nil {
		return domain.UpdateOvertimeEntryParams{}, err
	}

	return domain.UpdateOvertimeEntryParams{
		ScheduleID:  req.ScheduleID,
		EntryDate:   entryDate,
		Minutes:     req.Minutes,
		Reason:      ptr.TrimString(req.Reason),
		Description: ptr.TrimString(req.Description),
	}, nil
}

func toListOvertimeEntriesParams(req listOvertimeEntriesRequest) domain.ListOvertimeEntriesParams {
	return domain.ListOvertimeEntriesParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
		Status: req.Status,
	}
}

func toListMyOvertimeEntriesParams(
	employeeID uuid.UUID,
	req listMyOvertimeEntriesRequest,
) domain.ListMyOvertimeEntriesParams {
	return domain.ListMyOvertimeEntriesParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Status:     req.Status,
	}
}

func toOvertimeEntryResponse(item *domain.OvertimeEntry) overtimeEntryResponse {
	return overtimeEntryResponse{
		ID:                   item.ID,
		EmployeeID:           item.EmployeeID,
		EmployeeName:         item.EmployeeName,
		ScheduleID:           item.ScheduleID,
		IsPaid:               item.PaidPeriodID != nil,
		EntryDate:            item.EntryDate,
		Minutes:              item.Minutes,
		Reason:               item.Reason,
		Description:          item.Description,
		Status:               item.Status,
		SubmittedAt:          item.SubmittedAt,
		ApprovedAt:           item.ApprovedAt,
		ApprovedByEmployeeID: item.ApprovedByEmployeeID,
		ApprovedByName:       item.ApprovedByName,
		RejectionReason:      item.RejectionReason,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}

func toOvertimeEntryResponses(items []domain.OvertimeEntry) []overtimeEntryResponse {
	results := make([]overtimeEntryResponse, len(items))
	for i, item := range items {
		results[i] = toOvertimeEntryResponse(&item)
	}
	return results
}

func toOvertimeStatsResponse(stats *domain.OvertimeStats) overtimeStatsResponse {
	return overtimeStatsResponse{
		TotalApprovedMinutes:  stats.TotalApprovedMinutes,
		TotalAwaitingApproval: stats.TotalAwaitingApproval,
		TotalApproved:         stats.TotalApproved,
		TotalSubmitted:        stats.TotalSubmitted,
	}
}

func parseOvertimeDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(overtimeDateLayout, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
