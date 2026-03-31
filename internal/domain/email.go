package domain

import (
	"context"
	"time"
)

type SMTPClient interface {
	Send(subject, body string, to []string) error
}

type EmailCredentials struct {
	Name     string
	Email    string
	Password string
}

type IncidentEmail struct {
	IncidentID   string
	ReportedBy   string
	ClientName   string
	IncidentType string
	Severity     string
	Location     string
	DocumentLink string
}

type AcceptedRegistrationFormEmail struct {
	ReferrerName        string
	ChildName           string
	ChildBSN            string
	AppointmentDate     string
	AppointmentLocation string
}

type ProcessRegistrationFormEmail struct {
	RecipientName string
	ClientName    string
	Location      string
	Link          string
}

type ClientContractReminderEmail struct {
	ClientID           string
	ClientFirstName    string
	ClientLastName     string
	ContractID         string
	CareType           string
	ContractStartDate  time.Time
	ContractEndDate    time.Time
	ContractStatus     string
	ReminderType       string
	LastReminderSentAt *time.Time
	CurrentDate        time.Time
	CurrentYear        int
}

type EmailSender interface {
	SendCredentials(ctx context.Context, to []string, data EmailCredentials) error
	SendIncident(ctx context.Context, to []string, data IncidentEmail) error
	SendIncidentWithAttachment(ctx context.Context, to []string, data IncidentEmail, attachmentName string, attachmentBytes []byte) error
	SendAcceptedRegistrationForm(ctx context.Context, to []string, data AcceptedRegistrationFormEmail) error
	SendProcessRegistrationForm(ctx context.Context, to []string, data ProcessRegistrationFormEmail) error
	SendClientContractReminder(ctx context.Context, to []string, data ClientContractReminderEmail) error
}
