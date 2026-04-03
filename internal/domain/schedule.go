package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWeekNotEmpty               = errors.New("week is not empty")
	ErrScheduleAutogenUnavailable = errors.New(
		"schedule auto-generation is only available in the Docker dev environment",
	)
	ErrScheduleNotFound                = errors.New("schedule not found")
	ErrInvalidSchedule                 = errors.New("invalid schedule")
	ErrInvalidLocationTZ               = errors.New("invalid location timezone")
	ErrShiftSwapNotFound               = errors.New("shift swap request not found")
	ErrShiftSwapInvalidRequest         = errors.New("invalid shift swap request")
	ErrShiftSwapStateInvalid           = errors.New("shift swap request is not in a valid state")
	ErrShiftSwapExpired                = errors.New("shift swap request has expired")
	ErrShiftSwapConflict               = errors.New("swap would create schedule overlap conflict")
	ErrShiftSwapScheduleOwnership      = errors.New("schedule ownership is invalid for this swap")
	ErrShiftSwapDuplicateActiveRequest = errors.New(
		"one of the schedules is already in an active swap request",
	)
)

type CreateScheduleRequest struct {
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
	LocationID  uuid.UUID   `json:"location_id"`
	IsCustom    bool        `json:"is_custom"            example:"true"`
	Recurrence  *string     `json:"recurrence,omitempty" example:"end_of_week"`

	StartDatetime *time.Time `json:"start_datetime,omitempty" example:"2023-10-01T09:00:00Z"`
	EndDatetime   *time.Time `json:"end_datetime,omitempty"   example:"2023-10-01T17:00:00Z"`

	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty" example:"1"`
	ShiftDate       *string    `json:"shift_date,omitempty"        example:"2023-10-01"`
}

type CreateScheduleResponse struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	LocationID      uuid.UUID  `json:"location_id"`
	LocationName    string     `json:"location_name"`
	StartDatetime   time.Time  `json:"start_datetime"`
	EndDatetime     time.Time  `json:"end_datetime"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty"`
	ShiftName       *string    `json:"shift_name,omitempty"`
}

const (
	CreateScheduleRecurrenceNone       = "none"
	CreateScheduleRecurrenceEndOfWeek  = "end_of_week"
	CreateScheduleRecurrenceEndOfMonth = "end_of_month"
)

type GetSchedulesByLocationInRangeRequest struct {
	StartDate string `form:"start_date" binding:"required" example:"2026-02-01"`
	EndDate   string `form:"end_date"   binding:"required" example:"2026-02-29"`
}

type Shift struct {
	ScheduleID        uuid.UUID  `json:"schedule_id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	EmployeeFirstName string     `json:"employee_first_name"`
	EmployeeLastName  string     `json:"employee_last_name"`
	StartTime         time.Time  `json:"start_time"`
	EndTime           time.Time  `json:"end_time"`
	LocationID        uuid.UUID  `json:"location_id"`
	ShiftName         *string    `json:"shift_name,omitempty"`
	LocationShiftID   *uuid.UUID `json:"location_shift_id,omitempty"`
	IsCustom          bool       `json:"is_custom"`
}

type GetSchedulesByLocationInRangeResponse struct {
	Date   string  `json:"date"`
	Shifts []Shift `json:"shifts"`
}

type GetScheduleByIdResponse struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	EmployeeFirstName string     `json:"employee_first_name"`
	EmployeeLastName  string     `json:"employee_last_name"`
	LocationID        uuid.UUID  `json:"location_id"`
	LocationName      string     `json:"location_name"`
	LocationShiftID   *uuid.UUID `json:"location_shift_id,omitempty"`
	LocationShiftName *string    `json:"shift_name,omitempty"`
	StartDatetime     time.Time  `json:"start_datetime"`
	EndDatetime       time.Time  `json:"end_datetime"`
	IsCustom          bool       `json:"is_custom"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type UpdateScheduleRequest struct {
	EmployeeID *uuid.UUID `json:"employee_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	IsCustom   *bool      `json:"is_custom,omitempty"   example:"true"`

	StartDatetime *time.Time `json:"start_datetime,omitempty" example:"2023-10-01T09:00:00Z"`
	EndDatetime   *time.Time `json:"end_datetime,omitempty"   example:"2023-10-01T17:00:00Z"`

	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty" example:"1"`
	ShiftDate       *string    `json:"shift_date,omitempty"        example:"2023-10-01"`
}

type UpdateScheduleResponse struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	LocationID      uuid.UUID  `json:"location_id"`
	StartDatetime   time.Time  `json:"start_datetime"`
	EndDatetime     time.Time  `json:"end_datetime"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LocationName    string     `json:"location_name"`
	LocationShiftID *uuid.UUID `json:"location_shift_id,omitempty"`
	ShiftName       *string    `json:"shift_name,omitempty"`
}

type AutoGenerateSchedulesRequest struct {
	LocationID  uuid.UUID   `json:"location_id"`
	Week        int32       `json:"week"`
	Year        int32       `json:"year"`
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
}

type SchedulePlanConstraints struct {
	MaxStaffPerShift int32 `json:"max_staff_per_shift"`
	AllowEmptyShift  bool  `json:"allow_empty_shift"`
}

type SchedulePlanEmployee struct {
	ID            uuid.UUID `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	TargetMinutes int64     `json:"target_minutes"`
}

type ScheduleShiftTemplate struct {
	ShiftID         uuid.UUID `json:"shift_id"`
	Name            string    `json:"name"`
	StartMinute     int32     `json:"start_minute"`
	EndMinute       int32     `json:"end_minute"`
	DurationMinutes int64     `json:"duration_minutes"`
	Overnight       bool      `json:"overnight"`
}

type SchedulePlanSlot struct {
	Date        string      `json:"date"`
	ShiftID     uuid.UUID   `json:"shift_id"`
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
}

type ScheduleEmployeeSummary struct {
	EmployeeID      uuid.UUID         `json:"employee_id"`
	TargetMinutes   int64             `json:"target_minutes"`
	AssignedMinutes int64             `json:"assigned_minutes"`
	OvertimeMinutes int64             `json:"overtime_minutes"`
	ShiftCounts     map[uuid.UUID]int `json:"shift_counts"`
}

type SchedulePlanWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AutoGenerateSchedulesResponse struct {
	Status         string                    `json:"status"`
	PlanID         uuid.UUID                 `json:"plan_id"`
	LocationID     uuid.UUID                 `json:"location_id"`
	Timezone       string                    `json:"timezone"`
	Week           int32                     `json:"week"`
	Year           int32                     `json:"year"`
	WeekStartDate  string                    `json:"week_start_date"`
	Constraints    SchedulePlanConstraints   `json:"constraints"`
	Employees      []SchedulePlanEmployee    `json:"employees"`
	ShiftTemplates []ScheduleShiftTemplate   `json:"shift_templates"`
	Slots          []SchedulePlanSlot        `json:"slots"`
	Summary        []ScheduleEmployeeSummary `json:"summary"`
	Warnings       []SchedulePlanWarning     `json:"warnings,omitempty"`
}

type SaveGeneratedSchedulesRequest struct {
	PlanID     uuid.UUID          `json:"plan_id"`
	LocationID uuid.UUID          `json:"location_id"`
	Week       int32              `json:"week"`
	Year       int32              `json:"year"`
	Slots      []SchedulePlanSlot `json:"slots"`
}

type ScheduleLocation struct {
	ID       uuid.UUID
	Timezone string
}

type ScheduleLocationShift struct {
	ID                uuid.UUID
	LocationID        uuid.UUID
	ShiftName         string
	StartMicroseconds int64
	EndMicroseconds   int64
}

type ScheduleEmployeeContractHours struct {
	ID            uuid.UUID
	FirstName     string
	LastName      string
	ContractHours *float64
}

type CreateScheduleParams struct {
	EmployeeID             uuid.UUID
	LocationID             uuid.UUID
	LocationShiftID        *uuid.UUID
	ShiftNameSnapshot      *string
	ShiftStartTimeSnapshot *int64
	ShiftEndTimeSnapshot   *int64
	IsCustom               bool
	CreatedByEmployeeID    uuid.UUID
	StartDatetime          time.Time
	EndDatetime            time.Time
}

type UpdateScheduleParams struct {
	EmployeeID             uuid.UUID
	LocationID             uuid.UUID
	LocationShiftID        *uuid.UUID
	ShiftNameSnapshot      *string
	ShiftStartTimeSnapshot *int64
	ShiftEndTimeSnapshot   *int64
	IsCustom               bool
	StartDatetime          time.Time
	EndDatetime            time.Time
}

type CreateShiftSwapRequest struct {
	RecipientEmployeeID uuid.UUID
	RequesterScheduleID uuid.UUID
	RecipientScheduleID uuid.UUID
	ExpiresAt           *time.Time
}

type CreateShiftSwapResponse struct {
	ID                  uuid.UUID
	RequesterEmployeeID uuid.UUID
	RecipientEmployeeID uuid.UUID
	RequesterScheduleID uuid.UUID
	RecipientScheduleID uuid.UUID
	Status              string
	RequestedAt         time.Time
	ExpiresAt           *time.Time
	Direction           string
}

type RespondShiftSwapRequest struct {
	Decision string
	Note     *string
}

type AdminDecisionShiftSwapRequest struct {
	Decision string
	Note     *string
}

type ListShiftSwapRequestsParams struct {
	Limit      int32
	Offset     int32
	Status     *string
	Filter     *string
	EmployeeID *uuid.UUID
}

type ShiftSwapScheduleSnapshot struct {
	ID            uuid.UUID
	EmployeeID    uuid.UUID
	EmployeeName  string
	ShiftName     string
	StartDatetime time.Time
	EndDatetime   time.Time
}

type ShiftSwapResponse struct {
	ID                    uuid.UUID
	RequesterEmployeeID   uuid.UUID
	RequesterEmployeeName string
	RecipientEmployeeID   uuid.UUID
	RecipientEmployeeName string
	RequesterSchedule     ShiftSwapScheduleSnapshot
	RecipientSchedule     ShiftSwapScheduleSnapshot
	Status                string
	RequestedAt           time.Time
	RecipientRespondedAt  *time.Time
	AdminDecidedAt        *time.Time
	RecipientResponseNote *string
	AdminDecisionNote     *string
	AdminEmployeeID       *uuid.UUID
	AdminEmployeeName     *string
	ExpiresAt             *time.Time
	Direction             string
}

type ShiftSwapPage struct {
	Items      []ShiftSwapResponse
	TotalCount int64
}

type ShiftSwapRequestRecord struct {
	ID                    uuid.UUID
	RequesterEmployeeID   uuid.UUID
	RecipientEmployeeID   uuid.UUID
	RequesterScheduleID   uuid.UUID
	RecipientScheduleID   uuid.UUID
	Status                string
	RequestedAt           time.Time
	RecipientRespondedAt  *time.Time
	AdminDecidedAt        *time.Time
	RecipientResponseNote *string
	AdminDecisionNote     *string
	AdminEmployeeID       *uuid.UUID
	ExpiresAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ScheduleSwapValidation struct {
	ID            uuid.UUID
	EmployeeID    uuid.UUID
	LocationID    uuid.UUID
	StartDatetime time.Time
	EndDatetime   time.Time
}

type ScheduleRepository interface {
	CreateSchedule(
		ctx context.Context,
		params CreateScheduleParams,
	) (*CreateScheduleResponse, error)
	GetSchedulesByLocationInRange(
		ctx context.Context,
		locationID uuid.UUID,
		startDate, endDate time.Time,
	) ([]GetSchedulesByLocationInRangeResponse, error)
	GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*GetScheduleByIdResponse, error)
	UpdateSchedule(
		ctx context.Context,
		scheduleID uuid.UUID,
		params UpdateScheduleParams,
	) (*UpdateScheduleResponse, error)
	DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error
	GetLocationByID(ctx context.Context, locationID uuid.UUID) (*ScheduleLocation, error)
	GetShiftByID(ctx context.Context, shiftID uuid.UUID) (*ScheduleLocationShift, error)
	GetShiftsByLocationID(
		ctx context.Context,
		locationID uuid.UUID,
	) ([]ScheduleLocationShift, error)
	ListEmployeesWithContractHours(
		ctx context.Context,
		employeeIDs []uuid.UUID,
	) ([]ScheduleEmployeeContractHours, error)
	WithTx(ctx context.Context, fn func(tx ScheduleRepository) error) error
	ExpirePendingShiftSwapRequests(ctx context.Context) error
	GetScheduleForSwapValidation(
		ctx context.Context,
		scheduleID uuid.UUID,
	) (*ScheduleSwapValidation, error)
	CreateShiftSwapRequest(
		ctx context.Context,
		params CreateShiftSwapRequest,
		requesterEmployeeID uuid.UUID,
	) (*ShiftSwapRequestRecord, error)
	UpdateShiftSwapStatusAfterRecipientResponse(
		ctx context.Context,
		swapID, recipientEmployeeID uuid.UUID,
		status string,
		note *string,
	) (*ShiftSwapRequestRecord, error)
	UpdateShiftSwapAdminDecision(
		ctx context.Context,
		swapID uuid.UUID,
		status string,
		note *string,
		adminEmployeeID uuid.UUID,
	) (*ShiftSwapRequestRecord, error)
	MarkShiftSwapConfirmed(
		ctx context.Context,
		swapID uuid.UUID,
		note *string,
		adminEmployeeID uuid.UUID,
	) (*ShiftSwapRequestRecord, error)
	GetShiftSwapRequestByID(ctx context.Context, swapID uuid.UUID) (*ShiftSwapRequestRecord, error)
	GetShiftSwapRequestDetailsByID(
		ctx context.Context,
		swapID uuid.UUID,
	) (*ShiftSwapResponse, error)
	ListMyShiftSwapRequests(ctx context.Context, employeeID uuid.UUID) ([]ShiftSwapResponse, error)
	ListShiftSwapRequests(
		ctx context.Context,
		params ListShiftSwapRequestsParams,
	) (*ShiftSwapPage, error)
	LockSchedulesByIDsForSwap(
		ctx context.Context,
		ids []uuid.UUID,
	) ([]ScheduleSwapValidation, error)
	LockShiftSwapRequestForAdminDecision(
		ctx context.Context,
		swapID uuid.UUID,
	) (*ShiftSwapRequestRecord, error)
	CountScheduleOverlapsForEmployee(
		ctx context.Context,
		employeeID uuid.UUID,
		excludedScheduleIDs []uuid.UUID,
		conflictStart, conflictEnd time.Time,
	) (int64, error)
	UpdateScheduleEmployeeAssignment(ctx context.Context, scheduleID, employeeID uuid.UUID) error
}

type ScheduleService interface {
	CreateSchedule(
		ctx context.Context,
		creatorID uuid.UUID,
		req *CreateScheduleRequest,
	) ([]CreateScheduleResponse, error)
	GetSchedulesByLocationInRange(
		ctx context.Context,
		locationID uuid.UUID,
		req *GetSchedulesByLocationInRangeRequest,
	) ([]GetSchedulesByLocationInRangeResponse, error)
	GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*GetScheduleByIdResponse, error)
	UpdateSchedule(
		ctx context.Context,
		scheduleID uuid.UUID,
		updaterEmployeeID uuid.UUID,
		req *UpdateScheduleRequest,
	) (*UpdateScheduleResponse, error)
	DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error
	AutoGenerateSchedules(
		ctx context.Context,
		req *AutoGenerateSchedulesRequest,
	) (*AutoGenerateSchedulesResponse, error)
	SaveGeneratedSchedules(
		ctx context.Context,
		creatorID uuid.UUID,
		req *SaveGeneratedSchedulesRequest,
	) error
	CreateShiftSwapRequest(
		ctx context.Context,
		requesterEmployeeID uuid.UUID,
		req *CreateShiftSwapRequest,
	) (*CreateShiftSwapResponse, error)
	RespondToShiftSwapRequest(
		ctx context.Context,
		recipientEmployeeID, swapID uuid.UUID,
		req *RespondShiftSwapRequest,
	) (*ShiftSwapResponse, error)
	AdminDecisionShiftSwapRequest(
		ctx context.Context,
		adminEmployeeID, swapID uuid.UUID,
		req *AdminDecisionShiftSwapRequest,
	) (*ShiftSwapResponse, error)
	ListMyShiftSwapRequests(ctx context.Context, employeeID uuid.UUID) ([]ShiftSwapResponse, error)
	ListShiftSwapRequests(
		ctx context.Context,
		params ListShiftSwapRequestsParams,
	) (*ShiftSwapPage, error)
}
