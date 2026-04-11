package handler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type createScheduleRequest struct {
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
	LocationID  uuid.UUID   `json:"location_id"`
	IsCustom    bool        `json:"is_custom"`
	Recurrence  *string     `json:"recurrence,omitempty"`

	StartDatetime *time.Time `json:"start_datetime,omitempty"`
	EndDatetime   *time.Time `json:"end_datetime,omitempty"`

	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty"`
	ShiftDate       *string    `json:"shift_date,omitempty"`
}

type getSchedulesByLocationInRangeRequest struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date"   binding:"required"`
}

type getEmployeeSchedulesByDayRequest struct {
	EmployeeID string `form:"employee_id" binding:"required"`
	Date       string `form:"date"        binding:"required"`
}

type getEmployeeSchedulesTimelineRequest struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date"   binding:"required"`
}

var uuidExtractRegex = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

type updateScheduleRequest struct {
	EmployeeID *uuid.UUID `json:"employee_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	IsCustom   *bool      `json:"is_custom,omitempty"`

	StartDatetime *time.Time `json:"start_datetime,omitempty"`
	EndDatetime   *time.Time `json:"end_datetime,omitempty"`

	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty"`
	ShiftDate       *string    `json:"shift_date,omitempty"`
}

type autoGenerateSchedulesRequest struct {
	LocationID  uuid.UUID   `json:"location_id"`
	Week        int32       `json:"week"`
	Year        int32       `json:"year"`
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
}

type schedulePlanSlotRequest struct {
	Date        string      `json:"date"`
	ShiftID     uuid.UUID   `json:"shift_id"`
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
}

type saveGeneratedSchedulesRequest struct {
	PlanID     uuid.UUID                 `json:"plan_id"`
	LocationID uuid.UUID                 `json:"location_id"`
	Week       int32                     `json:"week"`
	Year       int32                     `json:"year"`
	Slots      []schedulePlanSlotRequest `json:"slots"`
}

func toCreateScheduleRequest(req createScheduleRequest) *domain.CreateScheduleRequest {
	return &domain.CreateScheduleRequest{
		EmployeeIDs:     req.EmployeeIDs,
		LocationID:      req.LocationID,
		IsCustom:        req.IsCustom,
		Recurrence:      trimStringPtr(req.Recurrence),
		StartDatetime:   req.StartDatetime,
		EndDatetime:     req.EndDatetime,
		LocationShiftID: req.LocationShiftID,
		ShiftDate:       trimStringPtr(req.ShiftDate),
	}
}

func toGetSchedulesByLocationInRangeRequest(
	req getSchedulesByLocationInRangeRequest,
) *domain.GetSchedulesByLocationInRangeRequest {
	return &domain.GetSchedulesByLocationInRangeRequest{
		StartDate: strings.TrimSpace(req.StartDate),
		EndDate:   strings.TrimSpace(req.EndDate),
	}
}

func toGetEmployeeSchedulesByDayRequest(
	req getEmployeeSchedulesByDayRequest,
) (*domain.GetEmployeeSchedulesByDayRequest, error) {
	normalizedEmployeeID := normalizeUUIDValue(req.EmployeeID)
	employeeID, err := uuid.Parse(normalizedEmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee_id must be a valid uuid")
	}

	return &domain.GetEmployeeSchedulesByDayRequest{
		EmployeeID: employeeID,
		Date:       strings.TrimSpace(req.Date),
	}, nil
}

func toGetEmployeeSchedulesTimelineRequest(
	req getEmployeeSchedulesTimelineRequest,
	employeeID uuid.UUID,
) *domain.GetEmployeeSchedulesTimelineRequest {
	return &domain.GetEmployeeSchedulesTimelineRequest{
		EmployeeID: employeeID,
		StartDate:  strings.TrimSpace(req.StartDate),
		EndDate:    strings.TrimSpace(req.EndDate),
	}
}

func normalizeUUIDValue(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return ""
	}
	if _, err := uuid.Parse(raw); err == nil {
		return raw
	}

	return uuidExtractRegex.FindString(raw)
}

func toUpdateScheduleRequest(req updateScheduleRequest) *domain.UpdateScheduleRequest {
	return &domain.UpdateScheduleRequest{
		EmployeeID:      req.EmployeeID,
		LocationID:      req.LocationID,
		IsCustom:        req.IsCustom,
		StartDatetime:   req.StartDatetime,
		EndDatetime:     req.EndDatetime,
		LocationShiftID: req.LocationShiftID,
		ShiftDate:       trimStringPtr(req.ShiftDate),
	}
}

func toAutoGenerateSchedulesRequest(
	req autoGenerateSchedulesRequest,
) *domain.AutoGenerateSchedulesRequest {
	return &domain.AutoGenerateSchedulesRequest{
		LocationID:  req.LocationID,
		Week:        req.Week,
		Year:        req.Year,
		EmployeeIDs: req.EmployeeIDs,
	}
}

func toSaveGeneratedSchedulesRequest(
	req saveGeneratedSchedulesRequest,
) *domain.SaveGeneratedSchedulesRequest {
	slots := make([]domain.SchedulePlanSlot, len(req.Slots))
	for i, slot := range req.Slots {
		slots[i] = domain.SchedulePlanSlot{
			Date:        strings.TrimSpace(slot.Date),
			ShiftID:     slot.ShiftID,
			EmployeeIDs: slot.EmployeeIDs,
		}
	}

	return &domain.SaveGeneratedSchedulesRequest{
		PlanID:     req.PlanID,
		LocationID: req.LocationID,
		Week:       req.Week,
		Year:       req.Year,
		Slots:      slots,
	}
}
