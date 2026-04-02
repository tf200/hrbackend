package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTimeEntryNotFound       = errors.New("time entry not found")
	ErrTimeEntryForbidden      = errors.New("time entry is not accessible by the actor")
	ErrTimeEntryInvalidRequest = errors.New("invalid time entry")
	ErrTimeEntryStateInvalid   = errors.New("time entry is not in a valid state for this operation")
)

const (
	TimeEntryStatusDraft     = "draft"
	TimeEntryStatusSubmitted = "submitted"
	TimeEntryStatusApproved  = "approved"
	TimeEntryStatusRejected  = "rejected"

	TimeEntryHourTypeNormal   = "normal"
	TimeEntryHourTypeOvertime = "overtime"
	TimeEntryHourTypeTravel   = "travel"
	TimeEntryHourTypeLeave    = "leave"
	TimeEntryHourTypeSick     = "sick"
	TimeEntryHourTypeTraining = "training"
)

type TimeEntry struct {
	ID                   uuid.UUID
	EmployeeID           uuid.UUID
	EmployeeName         string
	ScheduleID           *uuid.UUID
	EntryDate            time.Time
	StartTime            string
	EndTime              string
	BreakMinutes         int32
	HourType             string
	ProjectName          *string
	ProjectNumber        *string
	ClientName           *string
	ActivityCategory     *string
	ActivityDescription  *string
	Status               string
	SubmittedAt          *time.Time
	ApprovedAt           *time.Time
	ApprovedByEmployeeID *uuid.UUID
	ApprovedByName       *string
	RejectionReason      *string
	Notes                *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type TimeEntryPage struct {
	Items      []TimeEntry
	TotalCount int64
}

type CreateTimeEntryParams struct {
	EmployeeID          uuid.UUID
	ScheduleID          *uuid.UUID
	EntryDate           time.Time
	StartTime           string
	EndTime             string
	BreakMinutes        int32
	HourType            string
	ProjectName         *string
	ProjectNumber       *string
	ClientName          *string
	ActivityCategory    *string
	ActivityDescription *string
	Notes               *string
}

type ListTimeEntriesParams struct {
	Limit          int32
	Offset         int32
	EmployeeID     *uuid.UUID
	EmployeeSearch *string
	Status         *string
}

type ListMyTimeEntriesParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Status     *string
}

type DecideTimeEntryParams struct {
	Decision        string
	RejectionReason *string
}

type TimeEntryTxRepository interface {
	GetTimeEntryForUpdate(ctx context.Context, timeEntryID uuid.UUID) (*TimeEntry, error)
	ApproveTimeEntry(ctx context.Context, timeEntryID, approvedByEmployeeID uuid.UUID) (*TimeEntry, error)
	RejectTimeEntry(ctx context.Context, timeEntryID uuid.UUID, rejectionReason *string) (*TimeEntry, error)
}

type TimeEntryRepository interface {
	WithTx(ctx context.Context, fn func(tx TimeEntryTxRepository) error) error
	CreateTimeEntry(ctx context.Context, params CreateTimeEntryParams) (*TimeEntry, error)
	GetTimeEntryByID(ctx context.Context, id uuid.UUID) (*TimeEntry, error)
	ListTimeEntries(ctx context.Context, params ListTimeEntriesParams) (*TimeEntryPage, error)
	ListMyTimeEntries(ctx context.Context, params ListMyTimeEntriesParams) (*TimeEntryPage, error)
}

type TimeEntryService interface {
	CreateTimeEntry(ctx context.Context, actorEmployeeID uuid.UUID, params CreateTimeEntryParams) (*TimeEntry, error)
	CreateTimeEntryByAdmin(ctx context.Context, adminEmployeeID uuid.UUID, params CreateTimeEntryParams) (*TimeEntry, error)
	DecideTimeEntryByAdmin(ctx context.Context, adminEmployeeID, timeEntryID uuid.UUID, params DecideTimeEntryParams) (*TimeEntry, error)
	GetTimeEntryByID(ctx context.Context, timeEntryID uuid.UUID) (*TimeEntry, error)
	GetMyTimeEntryByID(ctx context.Context, actorEmployeeID, timeEntryID uuid.UUID) (*TimeEntry, error)
	ListTimeEntries(ctx context.Context, params ListTimeEntriesParams) (*TimeEntryPage, error)
	ListMyTimeEntries(ctx context.Context, params ListMyTimeEntriesParams) (*TimeEntryPage, error)
}
