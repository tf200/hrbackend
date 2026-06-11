package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOvertimeNotFound       = errors.New("overtime entry not found")
	ErrOvertimeForbidden      = errors.New("overtime entry is not accessible by the actor")
	ErrOvertimeInvalidRequest = errors.New("invalid overtime entry")
	ErrOvertimeStateInvalid   = errors.New(
		"overtime entry is not in a valid state for this operation",
	)
)

const (
	OvertimeStatusSubmitted = "submitted"
	OvertimeStatusApproved  = "approved"
	OvertimeStatusRejected  = "rejected"

	OvertimeReasonClientCrisis             = "client_crisis"
	OvertimeReasonUnderstaffing            = "understaffing"
	OvertimeReasonMeetingConsultation      = "meeting_consultation"
	OvertimeReasonTrainingEducation        = "training_education"
	OvertimeReasonCompletingAdministration = "completing_administration"
	OvertimeReasonHandover                 = "handover"
	OvertimeReasonEmergency                = "emergency"
	OvertimeReasonProjectWork              = "project_work"
	OvertimeReasonEventActivity            = "event_activity"
	OvertimeReasonOther                    = "other"
)

type OvertimeEntry struct {
	ID                   uuid.UUID
	EmployeeID           uuid.UUID
	EmployeeName         string
	ScheduleID           *uuid.UUID
	PaidPeriodID         *uuid.UUID
	EntryDate            time.Time
	Minutes              int32
	Reason               string
	Description          *string
	Status               string
	SubmittedAt          time.Time
	ApprovedAt           *time.Time
	ApprovedByEmployeeID *uuid.UUID
	ApprovedByName       *string
	RejectionReason      *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type OvertimeEntryPage struct {
	Items      []OvertimeEntry
	TotalCount int64
}

type OvertimeStats struct {
	TotalApprovedMinutes  int64
	TotalAwaitingApproval int64
	TotalApproved         int64
	TotalSubmitted        int64
}

type CreateOvertimeEntryParams struct {
	EmployeeID  uuid.UUID
	ScheduleID  *uuid.UUID
	EntryDate   time.Time
	Minutes     int32
	Reason      string
	Description *string
}

type ListOvertimeEntriesParams struct {
	Limit  int32
	Offset int32
	Status *string
}

type ListMyOvertimeEntriesParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Status     *string
}

type DecideOvertimeEntryParams struct {
	Decision        string
	RejectionReason *string
}

type UpdateOvertimeEntryParams struct {
	EmployeeID  uuid.UUID
	ScheduleID  *uuid.UUID
	EntryDate   *time.Time
	Minutes     *int32
	Reason      *string
	Description *string
}

type OvertimeTxRepository interface {
	GetOvertimeEntryForUpdate(
		ctx context.Context,
		overtimeEntryID uuid.UUID,
	) (*OvertimeEntry, error)
	ApproveOvertimeEntry(
		ctx context.Context,
		overtimeEntryID, approvedByEmployeeID uuid.UUID,
	) (*OvertimeEntry, error)
	RejectOvertimeEntry(
		ctx context.Context,
		overtimeEntryID uuid.UUID,
		rejectionReason *string,
	) (*OvertimeEntry, error)
	UpdateOvertimeEntryByAdmin(
		ctx context.Context,
		overtimeEntryID uuid.UUID,
		params UpdateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	DeleteOvertimeEntry(ctx context.Context, overtimeEntryID uuid.UUID) error
}

type OvertimeRepository interface {
	WithTx(ctx context.Context, fn func(tx OvertimeTxRepository) error) error
	CreateOvertimeEntry(
		ctx context.Context,
		params CreateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	GetOvertimeEntryByID(ctx context.Context, id uuid.UUID) (*OvertimeEntry, error)
	ListOvertimeEntries(
		ctx context.Context,
		params ListOvertimeEntriesParams,
	) (*OvertimeEntryPage, error)
	ListMyOvertimeEntries(
		ctx context.Context,
		params ListMyOvertimeEntriesParams,
	) (*OvertimeEntryPage, error)
	GetCurrentMonthOvertimeStats(ctx context.Context) (*OvertimeStats, error)
	GetMyCurrentMonthOvertimeStats(
		ctx context.Context,
		employeeID uuid.UUID,
	) (*OvertimeStats, error)
}

type OvertimeService interface {
	CreateOvertimeEntry(
		ctx context.Context,
		actorEmployeeID uuid.UUID,
		params CreateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	CreateOvertimeEntryByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params CreateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	DecideOvertimeEntryByAdmin(
		ctx context.Context,
		adminEmployeeID, overtimeEntryID uuid.UUID,
		params DecideOvertimeEntryParams,
	) (*OvertimeEntry, error)
	UpdateOvertimeEntryByAdmin(
		ctx context.Context,
		adminEmployeeID, overtimeEntryID uuid.UUID,
		params UpdateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	UpdateMyOvertimeEntry(
		ctx context.Context,
		actorEmployeeID, overtimeEntryID uuid.UUID,
		params UpdateOvertimeEntryParams,
	) (*OvertimeEntry, error)
	DeleteOvertimeEntryByAdmin(
		ctx context.Context,
		adminEmployeeID, overtimeEntryID uuid.UUID,
	) error
	DeleteMyOvertimeEntry(
		ctx context.Context,
		actorEmployeeID, overtimeEntryID uuid.UUID,
	) error
	GetOvertimeEntryByID(ctx context.Context, overtimeEntryID uuid.UUID) (*OvertimeEntry, error)
	GetMyOvertimeEntryByID(
		ctx context.Context,
		actorEmployeeID, overtimeEntryID uuid.UUID,
	) (*OvertimeEntry, error)
	ListOvertimeEntries(
		ctx context.Context,
		params ListOvertimeEntriesParams,
	) (*OvertimeEntryPage, error)
	ListMyOvertimeEntries(
		ctx context.Context,
		params ListMyOvertimeEntriesParams,
	) (*OvertimeEntryPage, error)
	GetCurrentMonthOvertimeStats(ctx context.Context) (*OvertimeStats, error)
	GetMyCurrentMonthOvertimeStats(
		ctx context.Context,
		employeeID uuid.UUID,
	) (*OvertimeStats, error)
}
