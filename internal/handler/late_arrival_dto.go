package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

const lateArrivalDateLayout = "2006-01-02"

type createLateArrivalRequest struct {
	ArrivalDate string `json:"arrival_date" binding:"required,datetime=2006-01-02"`
	ArrivalTime string `json:"arrival_time" binding:"required"`
	Reason      string `json:"reason"       binding:"required"`
}

type createLateArrivalByAdminRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id"  binding:"required"`
	ArrivalDate string    `json:"arrival_date" binding:"required,datetime=2006-01-02"`
	ArrivalTime string    `json:"arrival_time" binding:"required"`
	Reason      string    `json:"reason"       binding:"required"`
}

type listMyLateArrivalsRequest struct {
	httpapi.PageRequest
	DateFrom *string `form:"date_from" binding:"omitempty,datetime=2006-01-02"`
	DateTo   *string `form:"date_to"   binding:"omitempty,datetime=2006-01-02"`
}

type listLateArrivalsRequest struct {
	httpapi.PageRequest
	EmployeeSearch *string `form:"employee_search" binding:"omitempty,max=120"`
	DateFrom       *string `form:"date_from"       binding:"omitempty,datetime=2006-01-02"`
	DateTo         *string `form:"date_to"         binding:"omitempty,datetime=2006-01-02"`
}

type createLateArrivalResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ScheduleID          uuid.UUID  `json:"schedule_id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	CreatedByEmployeeID *uuid.UUID `json:"created_by_employee_id,omitempty"`
	ArrivalDate         time.Time  `json:"arrival_date"`
	ArrivalTime         string     `json:"arrival_time"`
	Reason              string     `json:"reason"`
	ShiftStartDatetime  time.Time  `json:"shift_start_datetime"`
	ShiftEndDatetime    time.Time  `json:"shift_end_datetime"`
	ShiftName           string     `json:"shift_name"`
	LocationName        string     `json:"location_name"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type lateArrivalListItemResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ScheduleID          uuid.UUID  `json:"schedule_id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	EmployeeName        string     `json:"employee_name"`
	CreatedByEmployeeID *uuid.UUID `json:"created_by_employee_id,omitempty"`
	ArrivalDate         time.Time  `json:"arrival_date"`
	ArrivalTime         string     `json:"arrival_time"`
	Reason              string     `json:"reason"`
	ShiftStartDatetime  time.Time  `json:"shift_start_datetime"`
	ShiftEndDatetime    time.Time  `json:"shift_end_datetime"`
	ShiftName           string     `json:"shift_name"`
	LocationName        string     `json:"location_name"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func toCreateLateArrivalParams(
	employeeID, createdByEmployeeID uuid.UUID,
	req createLateArrivalRequest,
) (domain.LateArrivalCreateParams, error) {
	arrivalDate, err := time.Parse(lateArrivalDateLayout, req.ArrivalDate)
	if err != nil {
		return domain.LateArrivalCreateParams{}, err
	}

	return domain.LateArrivalCreateParams{
		EmployeeID:          employeeID,
		CreatedByEmployeeID: createdByEmployeeID,
		ArrivalDate:         arrivalDate.UTC(),
		ArrivalTime:         req.ArrivalTime,
		Reason:              req.Reason,
	}, nil
}

func toCreateLateArrivalByAdminParams(
	adminEmployeeID uuid.UUID,
	req createLateArrivalByAdminRequest,
) (domain.LateArrivalCreateParams, error) {
	return toCreateLateArrivalParams(req.EmployeeID, adminEmployeeID, createLateArrivalRequest{
		ArrivalDate: req.ArrivalDate,
		ArrivalTime: req.ArrivalTime,
		Reason:      req.Reason,
	})
}

func toListMyLateArrivalsParams(
	employeeID uuid.UUID,
	req listMyLateArrivalsRequest,
) (domain.ListMyLateArrivalsParams, error) {
	dateFrom, err := parseLateArrivalDatePtr(req.DateFrom)
	if err != nil {
		return domain.ListMyLateArrivalsParams{}, err
	}
	dateTo, err := parseLateArrivalDatePtr(req.DateTo)
	if err != nil {
		return domain.ListMyLateArrivalsParams{}, err
	}

	return domain.ListMyLateArrivalsParams{
		EmployeeID: employeeID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}, nil
}

func toListLateArrivalsParams(req listLateArrivalsRequest) (domain.ListLateArrivalsParams, error) {
	dateFrom, err := parseLateArrivalDatePtr(req.DateFrom)
	if err != nil {
		return domain.ListLateArrivalsParams{}, err
	}
	dateTo, err := parseLateArrivalDatePtr(req.DateTo)
	if err != nil {
		return domain.ListLateArrivalsParams{}, err
	}

	return domain.ListLateArrivalsParams{
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		EmployeeSearch: req.EmployeeSearch,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
	}, nil
}

func toCreateLateArrivalResponse(item *domain.CreateLateArrivalResult) createLateArrivalResponse {
	return createLateArrivalResponse{
		ID:                  item.ID,
		ScheduleID:          item.ScheduleID,
		EmployeeID:          item.EmployeeID,
		CreatedByEmployeeID: item.CreatedByEmployeeID,
		ArrivalDate:         item.ArrivalDate,
		ArrivalTime:         item.ArrivalTime,
		Reason:              item.Reason,
		ShiftStartDatetime:  item.ShiftStartDatetime,
		ShiftEndDatetime:    item.ShiftEndDatetime,
		ShiftName:           item.ShiftName,
		LocationName:        item.LocationName,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func toLateArrivalListItemResponse(item domain.LateArrivalListItem) lateArrivalListItemResponse {
	return lateArrivalListItemResponse{
		ID:                  item.ID,
		ScheduleID:          item.ScheduleID,
		EmployeeID:          item.EmployeeID,
		EmployeeName:        item.EmployeeName,
		CreatedByEmployeeID: item.CreatedByEmployeeID,
		ArrivalDate:         item.ArrivalDate,
		ArrivalTime:         item.ArrivalTime,
		Reason:              item.Reason,
		ShiftStartDatetime:  item.ShiftStartDatetime,
		ShiftEndDatetime:    item.ShiftEndDatetime,
		ShiftName:           item.ShiftName,
		LocationName:        item.LocationName,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func parseLateArrivalDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(lateArrivalDateLayout, trimmed)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
