package asynq

import (
	"time"

	"github.com/google/uuid"
)

type IncidentPayload struct {
	ID                      uuid.UUID `json:"id"`
	EmployeeID              uuid.UUID `json:"employee_id"`
	EmployeeFirstName       string    `json:"employee_first_name"`
	EmployeeLastName        string    `json:"employee_last_name"`
	LocationID              uuid.UUID `json:"location_id"`
	ReporterInvolvement     string    `json:"reporter_involvement"`
	InformedParties         []string  `json:"informed_parties"`
	OccurredAt              time.Time `json:"occurred_at"`
	IncidentType            string    `json:"incident_type"`
	SeverityOfIncident      string    `json:"severity_of_incident"`
	IncidentExplanation     *string   `json:"incident_explanation"`
	RecurrenceRisk          string    `json:"recurrence_risk"`
	IncidentPreventSteps    *string   `json:"incident_prevent_steps"`
	IncidentTakenMeasures   *string   `json:"incident_taken_measures"`
	CauseCategories         []string  `json:"cause_categories"`
	CauseExplanation        *string   `json:"cause_explanation"`
	PhysicalInjury          string    `json:"physical_injury"`
	PhysicalInjuryDesc      *string   `json:"physical_injury_desc"`
	PsychologicalDamage     string    `json:"psychological_damage"`
	PsychologicalDamageDesc *string   `json:"psychological_damage_desc"`
	NeededConsultation      string    `json:"needed_consultation"`
	FollowUpActions         []string  `json:"follow_up_actions"`
	FollowUpNotes           *string   `json:"follow_up_notes"`
	IsEmployeeAbsent        bool      `json:"is_employee_absent"`
	AdditionalDetails       *string   `json:"additional_details"`
	ClientID                uuid.UUID `json:"client_id"`
	LocationName            string    `json:"location_name"`
	Emails                  []string  `json:"emails"`
}

type EmailDeliveryPayload struct {
	To           string `json:"to"`
	Name         string `json:"name"`
	UserEmail    string `json:"user_email"`
	UserPassword string `json:"user_password"`
}

type AcceptedRegistrationFormPayload struct {
	ReferrerName        string `json:"referrer_name"`
	ChildName           string `json:"child_name"`
	ChildBSN            string `json:"child_bsn"`
	AppointmentDate     string `json:"appointment_date"`
	AppointmentLocation string `json:"appointment_location"`
	To                  string `json:"to"`
}

type ProcessRegistrationFormEmailPayload struct {
	ReferrerName string   `json:"referrer_name"`
	ClientName   string   `json:"client_name"`
	Location     string   `json:"location"`
	Link         string   `json:"link"`
	To           []string `json:"to"`
}

type IncidentConfirmedEmailPayload struct {
	IncidentID uuid.UUID `json:"incident_id"`
}

type NotificationPayload struct {
	RecipientUserIDs []uuid.UUID      `json:"recipient_user_ids"`
	Type             string           `json:"type"`
	Data             NotificationData `json:"data"`
	CreatedAt        time.Time        `json:"created_at"`
	Message          string           `json:"message"`
}

type NotificationData struct {
	NewAppointment          *NewAppointmentData          `json:"new_appointment,omitempty"`
	NewClientAssignment     *NewClientAssignmentData     `json:"new_client_assignment,omitempty"`
	ClientContractReminder  *ClientContractReminderData  `json:"client_contract_reminder,omitempty"`
	NewIncidentReport       *NewIncidentReportData       `json:"new_incident_report,omitempty"`
	NewScheduleNotification *NewScheduleNotificationData `json:"new_schedule_notification,omitempty"`
}

type NewAppointmentData struct {
	AppointmentID uuid.UUID `json:"appointment_id"`
	CreatedBy     string    `json:"created_by"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Location      string    `json:"location"`
}

type NewClientAssignmentData struct {
	ClientID        uuid.UUID `json:"client_id"`
	ClientFirstName string    `json:"client_first_name"`
	ClientLastName  string    `json:"client_last_name"`
	ClientLocation  *string   `json:"client_location"`
}

type ClientContractReminderData struct {
	ClientID           uuid.UUID  `json:"client_id"`
	ClientFirstName    string     `json:"client_first_name"`
	ClientLastName     string     `json:"client_last_name"`
	ContractID         uuid.UUID  `json:"contract_id"`
	CareType           string     `json:"care_type"`
	ContractStart      time.Time  `json:"contract_start"`
	ContractEnd        time.Time  `json:"contract_end"`
	ReminderType       string     `json:"reminder_type"`
	LastReminderSentAt *time.Time `json:"last_reminder_sent_at,omitempty"`
}

type NewIncidentReportData struct {
	ID                 uuid.UUID `json:"id"`
	EmployeeID         uuid.UUID `json:"employee_id"`
	EmployeeFirstName  string    `json:"employee_first_name"`
	EmployeeLastName   string    `json:"employee_last_name"`
	LocationID         uuid.UUID `json:"location_id"`
	LocationName       string    `json:"location_name"`
	ClientID           uuid.UUID `json:"client_id"`
	ClientFirstName    string    `json:"client_first_name"`
	ClientLastName     string    `json:"client_last_name"`
	SeverityOfIncident string    `json:"severity_of_incident"`
}

type NewScheduleNotificationData struct {
	ScheduleID uuid.UUID `json:"schedule_id"`
	CreatedBy  uuid.UUID `json:"created_by"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Location   string    `json:"location"`
}
