package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLateArrivalInvalidRequest = errors.New("invalid late arrival request")
	ErrLateArrivalConflict       = errors.New("late arrival conflict")
)

type LateArrival struct {
	ID                  uuid.UUID
	ScheduleID          uuid.UUID
	EmployeeID          uuid.UUID
	CreatedByEmployeeID *uuid.UUID
	ArrivalDate         time.Time
	ArrivalTime         string
	Reason              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type LateArrivalListItem struct {
	LateArrival
	EmployeeName       string
	ShiftStartDatetime time.Time
	ShiftEndDatetime   time.Time
	ShiftName          string
	LocationName       string
}

type LateArrivalPage struct {
	Items      []LateArrivalListItem
	TotalCount int64
}

type AssignedScheduleForDate struct {
	ScheduleID       uuid.UUID
	EmployeeID       uuid.UUID
	StartDatetime    time.Time
	EndDatetime      time.Time
	LocationTimezone string
	LocationName     string
	ShiftName        string
}

type LateArrivalCreateParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	ArrivalDate         time.Time
	ArrivalTime         string
	Reason              string
}

type CreateLateArrivalResult struct {
	ID                  uuid.UUID
	ScheduleID          uuid.UUID
	EmployeeID          uuid.UUID
	CreatedByEmployeeID *uuid.UUID
	ArrivalDate         time.Time
	ArrivalTime         string
	Reason              string
	ShiftStartDatetime  time.Time
	ShiftEndDatetime    time.Time
	ShiftName           string
	LocationName        string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ListMyLateArrivalsParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	DateFrom   *time.Time
	DateTo     *time.Time
}

type ListLateArrivalsParams struct {
	Limit          int32
	Offset         int32
	EmployeeSearch *string
	DateFrom       *time.Time
	DateTo         *time.Time
}

type LateArrivalRepository interface {
	ListAssignedSchedulesForEmployeeOnDate(ctx context.Context, employeeID uuid.UUID, arrivalDate time.Time) ([]AssignedScheduleForDate, error)
	CreateLateArrival(ctx context.Context, params LateArrivalCreateParams, scheduleID uuid.UUID) (*LateArrival, error)
	ListMyLateArrivals(ctx context.Context, params ListMyLateArrivalsParams) (*LateArrivalPage, error)
	ListLateArrivals(ctx context.Context, params ListLateArrivalsParams) (*LateArrivalPage, error)
}

type LateArrivalService interface {
	CreateLateArrival(ctx context.Context, params LateArrivalCreateParams) (*CreateLateArrivalResult, error)
	CreateLateArrivalByAdmin(ctx context.Context, params LateArrivalCreateParams) (*CreateLateArrivalResult, error)
	ListMyLateArrivals(ctx context.Context, params ListMyLateArrivalsParams) (*LateArrivalPage, error)
	ListLateArrivals(ctx context.Context, params ListLateArrivalsParams) (*LateArrivalPage, error)
}
