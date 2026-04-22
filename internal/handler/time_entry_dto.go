package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
)

const timeEntryDateLayout = "2006-01-02"

type createTimeEntryRequest struct {
	ScheduleID          *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate           string     `json:"entry_date"            binding:"required,datetime=2006-01-02"`
	StartTime           string     `json:"start_time"            binding:"required"`
	EndTime             string     `json:"end_time"              binding:"required"`
	BreakMinutes        int32      `json:"break_minutes"`
	HourType            string     `json:"hour_type"             binding:"required,oneof=normal overtime travel leave sick training"`
	ProjectName         *string    `json:"project_name"`
	ProjectNumber       *string    `json:"project_number"`
	ClientName          *string    `json:"client_name"`
	ActivityCategory    *string    `json:"activity_category"`
	ActivityDescription *string    `json:"activity_description"`
	Notes               *string    `json:"notes"`
}

type createTimeEntryByAdminRequest struct {
	EmployeeID          uuid.UUID  `json:"employee_id"           binding:"required"`
	ScheduleID          *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate           string     `json:"entry_date"            binding:"required,datetime=2006-01-02"`
	StartTime           string     `json:"start_time"            binding:"required"`
	EndTime             string     `json:"end_time"              binding:"required"`
	BreakMinutes        int32      `json:"break_minutes"`
	HourType            string     `json:"hour_type"             binding:"required,oneof=normal overtime travel leave sick training"`
	ProjectName         *string    `json:"project_name"`
	ProjectNumber       *string    `json:"project_number"`
	ClientName          *string    `json:"client_name"`
	ActivityCategory    *string    `json:"activity_category"`
	ActivityDescription *string    `json:"activity_description"`
	Notes               *string    `json:"notes"`
}

type decideTimeEntryByAdminRequest struct {
	Decision        string  `json:"decision"         binding:"required,oneof=approve reject"`
	RejectionReason *string `json:"rejection_reason"`
}

type updateTimeEntryByAdminRequest struct {
	ScheduleID          *uuid.UUID `json:"schedule_id,omitempty"`
	EntryDate           *string    `json:"entry_date"            binding:"omitempty,datetime=2006-01-02"`
	StartTime           *string    `json:"start_time"`
	EndTime             *string    `json:"end_time"`
	BreakMinutes        *int32     `json:"break_minutes"`
	HourType            *string    `json:"hour_type"             binding:"omitempty,oneof=normal overtime travel leave sick training"`
	ProjectName         *string    `json:"project_name"`
	ProjectNumber       *string    `json:"project_number"`
	ClientName          *string    `json:"client_name"`
	ActivityCategory    *string    `json:"activity_category"`
	ActivityDescription *string    `json:"activity_description"`
	Notes               *string    `json:"notes"`
	Status              *string    `json:"status"                binding:"omitempty,oneof=submitted"`
	AdminUpdateNote     string     `json:"admin_update_note"     binding:"required"`
}

type listTimeEntriesRequest struct {
	httpapi.PageRequest
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
	Status         *string `form:"status"          binding:"omitempty,oneof=draft submitted approved rejected"`
}

type listMyTimeEntriesRequest struct {
	httpapi.PageRequest
	Status *string `form:"status" binding:"omitempty,oneof=draft submitted approved rejected"`
}

type timeEntryResponse struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	EmployeeName         string     `json:"employee_name"`
	ScheduleID           *uuid.UUID `json:"schedule_id,omitempty"`
	IsPaid               bool       `json:"is_paid"`
	EntryDate            time.Time  `json:"entry_date"`
	StartTime            string     `json:"start_time"`
	EndTime              string     `json:"end_time"`
	BreakMinutes         int32      `json:"break_minutes"`
	HourType             string     `json:"hour_type"`
	ProjectName          *string    `json:"project_name,omitempty"`
	ProjectNumber        *string    `json:"project_number,omitempty"`
	ClientName           *string    `json:"client_name,omitempty"`
	ActivityCategory     *string    `json:"activity_category,omitempty"`
	ActivityDescription  *string    `json:"activity_description,omitempty"`
	Status               string     `json:"status"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	ApprovedByEmployeeID *uuid.UUID `json:"approved_by_employee_id,omitempty"`
	ApprovedByName       *string    `json:"approved_by_name,omitempty"`
	RejectionReason      *string    `json:"rejection_reason,omitempty"`
	Notes                *string    `json:"notes,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type timeEntryStatsResponse struct {
	TotalHours            float64 `json:"total_hours"`
	TotalAwaitingApproval int64   `json:"total_awaiting_approval"`
	TotalApproved         int64   `json:"total_approved"`
	TotalConcepts         int64   `json:"total_concepts"`
}

func toCreateTimeEntryParams(req createTimeEntryRequest) (domain.CreateTimeEntryParams, error) {
	entryDate, err := time.Parse(timeEntryDateLayout, req.EntryDate)
	if err != nil {
		return domain.CreateTimeEntryParams{}, err
	}

	return domain.CreateTimeEntryParams{
		ScheduleID:          req.ScheduleID,
		EntryDate:           entryDate.UTC(),
		StartTime:           strings.TrimSpace(req.StartTime),
		EndTime:             strings.TrimSpace(req.EndTime),
		BreakMinutes:        req.BreakMinutes,
		HourType:            strings.TrimSpace(req.HourType),
		ProjectName:         req.ProjectName,
		ProjectNumber:       req.ProjectNumber,
		ClientName:          req.ClientName,
		ActivityCategory:    req.ActivityCategory,
		ActivityDescription: req.ActivityDescription,
		Notes:               req.Notes,
	}, nil
}

func toCreateTimeEntryByAdminParams(
	req createTimeEntryByAdminRequest,
) (domain.CreateTimeEntryParams, error) {
	base, err := toCreateTimeEntryParams(createTimeEntryRequest{
		ScheduleID:          req.ScheduleID,
		EntryDate:           req.EntryDate,
		StartTime:           req.StartTime,
		EndTime:             req.EndTime,
		BreakMinutes:        req.BreakMinutes,
		HourType:            req.HourType,
		ProjectName:         req.ProjectName,
		ProjectNumber:       req.ProjectNumber,
		ClientName:          req.ClientName,
		ActivityCategory:    req.ActivityCategory,
		ActivityDescription: req.ActivityDescription,
		Notes:               req.Notes,
	})
	if err != nil {
		return domain.CreateTimeEntryParams{}, err
	}
	base.EmployeeID = req.EmployeeID
	return base, nil
}

func toDecideTimeEntryParams(req decideTimeEntryByAdminRequest) domain.DecideTimeEntryParams {
	return domain.DecideTimeEntryParams{
		Decision:        strings.TrimSpace(req.Decision),
		RejectionReason: ptr.TrimString(req.RejectionReason),
	}
}

func toUpdateTimeEntryByAdminParams(
	req updateTimeEntryByAdminRequest,
) (domain.UpdateTimeEntryByAdminParams, string, error) {
	entryDate, err := parseTimeEntryDatePtr(req.EntryDate)
	if err != nil {
		return domain.UpdateTimeEntryByAdminParams{}, "", err
	}

	status := req.Status
	if status != nil {
		trimmed := strings.TrimSpace(*status)
		status = &trimmed
	}

	return domain.UpdateTimeEntryByAdminParams{
		ScheduleID:          req.ScheduleID,
		EntryDate:           entryDate,
		StartTime:           ptr.TrimString(req.StartTime),
		EndTime:             ptr.TrimString(req.EndTime),
		BreakMinutes:        req.BreakMinutes,
		HourType:            ptr.TrimString(req.HourType),
		ProjectName:         ptr.TrimString(req.ProjectName),
		ProjectNumber:       ptr.TrimString(req.ProjectNumber),
		ClientName:          ptr.TrimString(req.ClientName),
		ActivityCategory:    ptr.TrimString(req.ActivityCategory),
		ActivityDescription: ptr.TrimString(req.ActivityDescription),
		Notes:               ptr.TrimString(req.Notes),
		Status:              status,
	}, req.AdminUpdateNote, nil
}

func toListTimeEntriesParams(req listTimeEntriesRequest) domain.ListTimeEntriesParams {
	return domain.ListTimeEntriesParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		EmployeeSearch: req.EmployeeSearch,
		Status:         req.Status,
	}
}

func toListMyTimeEntriesParams(
	employeeID uuid.UUID,
	req listMyTimeEntriesRequest,
) domain.ListMyTimeEntriesParams {
	return domain.ListMyTimeEntriesParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Status:     req.Status,
	}
}

func toTimeEntryResponse(item *domain.TimeEntry) timeEntryResponse {
	return timeEntryResponse{
		ID:                   item.ID,
		EmployeeID:           item.EmployeeID,
		EmployeeName:         item.EmployeeName,
		ScheduleID:           item.ScheduleID,
		IsPaid:               item.PaidPeriodID != nil,
		EntryDate:            item.EntryDate,
		StartTime:            item.StartTime,
		EndTime:              item.EndTime,
		BreakMinutes:         item.BreakMinutes,
		HourType:             item.HourType,
		ProjectName:          item.ProjectName,
		ProjectNumber:        item.ProjectNumber,
		ClientName:           item.ClientName,
		ActivityCategory:     item.ActivityCategory,
		ActivityDescription:  item.ActivityDescription,
		Status:               item.Status,
		SubmittedAt:          item.SubmittedAt,
		ApprovedAt:           item.ApprovedAt,
		ApprovedByEmployeeID: item.ApprovedByEmployeeID,
		ApprovedByName:       item.ApprovedByName,
		RejectionReason:      item.RejectionReason,
		Notes:                item.Notes,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}

func toTimeEntryResponses(items []domain.TimeEntry) []timeEntryResponse {
	results := make([]timeEntryResponse, len(items))
	for i, item := range items {
		results[i] = toTimeEntryResponse(&item)
	}
	return results
}

func toTimeEntryStatsResponse(stats *domain.TimeEntryStats) timeEntryStatsResponse {
	return timeEntryStatsResponse{
		TotalHours:            stats.TotalHours,
		TotalAwaitingApproval: stats.TotalAwaitingApproval,
		TotalApproved:         stats.TotalApproved,
		TotalConcepts:         stats.TotalConcepts,
	}
}

func parseTimeEntryDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(timeEntryDateLayout, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
