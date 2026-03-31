package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TaskQueueCritical = "critical"
	TaskQueueDefault  = "default"
	TaskQueueLow      = "low"
)

type TaskEnqueueOptions struct {
	Queue    string
	MaxRetry int
}

type IncidentTaskPayload struct {
	ID                      uuid.UUID
	EmployeeID              uuid.UUID
	EmployeeFirstName       string
	EmployeeLastName        string
	LocationID              uuid.UUID
	ReporterInvolvement     string
	InformedParties         []string
	OccurredAt              time.Time
	IncidentType            string
	SeverityOfIncident      string
	IncidentExplanation     *string
	RecurrenceRisk          string
	IncidentPreventSteps    *string
	IncidentTakenMeasures   *string
	CauseCategories         []string
	CauseExplanation        *string
	PhysicalInjury          string
	PhysicalInjuryDesc      *string
	PsychologicalDamage     string
	PsychologicalDamageDesc *string
	NeededConsultation      string
	FollowUpActions         []string
	FollowUpNotes           *string
	IsEmployeeAbsent        bool
	AdditionalDetails       *string
	ClientID                uuid.UUID
	LocationName            string
	Emails                  []string
}

type EmailDeliveryTaskPayload struct {
	To           string
	Name         string
	UserEmail    string
	UserPassword string
}

type AcceptedRegistrationFormTaskPayload struct {
	ReferrerName        string
	ChildName           string
	ChildBSN            string
	AppointmentDate     string
	AppointmentLocation string
	To                  string
}

type ProcessRegistrationFormEmailTaskPayload struct {
	ReferrerName string
	ClientName   string
	Location     string
	Link         string
	To           []string
}

type IncidentConfirmedEmailTaskPayload struct {
	IncidentID uuid.UUID
}

type NotificationTaskPayload struct {
	RecipientUserIDs []uuid.UUID
	Type             string
	Data             NotificationTaskData
	CreatedAt        time.Time
	Message          string
}

type NotificationTaskData struct {
	NewAppointment          *NewAppointmentTaskData
	NewClientAssignment     *NewClientAssignmentTaskData
	ClientContractReminder  *ClientContractReminderTaskData
	NewIncidentReport       *NewIncidentReportTaskData
	NewScheduleNotification *NewScheduleNotificationTaskData
}

type NewAppointmentTaskData struct {
	AppointmentID uuid.UUID
	CreatedBy     string
	StartTime     time.Time
	EndTime       time.Time
	Location      string
}

type NewClientAssignmentTaskData struct {
	ClientID        uuid.UUID
	ClientFirstName string
	ClientLastName  string
	ClientLocation  *string
}

type ClientContractReminderTaskData struct {
	ClientID           uuid.UUID
	ClientFirstName    string
	ClientLastName     string
	ContractID         uuid.UUID
	CareType           string
	ContractStart      time.Time
	ContractEnd        time.Time
	ReminderType       string
	LastReminderSentAt *time.Time
}

type NewIncidentReportTaskData struct {
	ID                 uuid.UUID
	EmployeeID         uuid.UUID
	EmployeeFirstName  string
	EmployeeLastName   string
	LocationID         uuid.UUID
	LocationName       string
	ClientID           uuid.UUID
	ClientFirstName    string
	ClientLastName     string
	SeverityOfIncident string
}

type NewScheduleNotificationTaskData struct {
	ScheduleID uuid.UUID
	CreatedBy  uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	Location   string
}

type TaskQueue interface {
	EnqueueEmailDelivery(ctx context.Context, payload EmailDeliveryTaskPayload, opts *TaskEnqueueOptions) error
	EnqueueIncident(ctx context.Context, payload IncidentTaskPayload, opts *TaskEnqueueOptions) error
	EnqueueIncidentConfirmedEmail(ctx context.Context, payload IncidentConfirmedEmailTaskPayload, opts *TaskEnqueueOptions) error
	EnqueueNotificationTask(ctx context.Context, payload NotificationTaskPayload, opts *TaskEnqueueOptions) error
	EnqueueAcceptedRegistration(ctx context.Context, payload AcceptedRegistrationFormTaskPayload, opts *TaskEnqueueOptions) error
	EnqueueProcessRegistrationFormEmail(ctx context.Context, payload ProcessRegistrationFormEmailTaskPayload, opts *TaskEnqueueOptions) error
	Close() error
}
