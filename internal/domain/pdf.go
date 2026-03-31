package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AppointmentCardPDF struct {
	ID                     uuid.UUID
	ClientName             string
	Date                   string
	Mentor                 string
	GeneralInformation     []string
	ImportantContacts      []string
	HouseholdInfo          []string
	OrganizationAgreements []string
	YouthOfficerAgreements []string
	TreatmentAgreements    []string
	SmokingRules           []string
	Work                   []string
	SchoolInternship       []string
	Travel                 []string
	Leave                  []string
}

type ContractPDF struct {
	ID     int64
	Status string

	StartDate      string
	EndDate        string
	ReminderPeriod int

	SenderName        string
	SenderStreet      string
	SenderHouseNumber string
	SenderPostalCode  string
	SenderCity        string
	SenderContactInfo string

	ClientFirstName   string
	ClientLastName    string
	ClientAddress     string
	ClientContactInfo string

	CareType        string
	CareName        string
	FinancingAct    string
	FinancingOption string

	Hours            float64
	HoursType        string
	AmbulanteDisplay string

	Price          float64
	PriceTimeUnit  string
	Vat            float64
	TypeName       string
	GenerationDate string
}

type IncidentReportPDF struct {
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
	ClientFirstName         string
	ClientLastName          string
	LocationName            string
}

type InvoicePeriodPDF struct {
	StartDate             time.Time
	EndDate               time.Time
	AcommodationTimeFrame string
	AmbulanteTotalMinutes float64
}

type InvoiceDetailPDF struct {
	CareType      string
	Periods       []InvoicePeriodPDF
	Price         float64
	PriceTimeUnit string
	PreVatTotal   float64
	Total         float64
}

type InvoicePDF struct {
	ID                  uuid.UUID
	SenderName          string
	SenderContactPerson string
	SenderStreet        string
	SenderHouseNumber   string
	SenderPostalCode    string
	SenderCity          string
	InvoiceNumber       string
	InvoiceDate         time.Time
	DueDate             time.Time
	InvoiceDetails      []InvoiceDetailPDF
	TotalAmount         float64
	ExtraItems          map[string]string
}

type PDFService interface {
	GenerateAppointmentCardPDF(ctx context.Context, cardData AppointmentCardPDF) ([]byte, error)
	GenerateAndUploadAppointmentCardPDF(ctx context.Context, cardData AppointmentCardPDF) (string, error)
	GenerateAndUploadInvoicePDF(ctx context.Context, invoiceData InvoicePDF) (string, int64, error)
	GenerateAndUploadContractPDF(ctx context.Context, contractData ContractPDF) (string, error)
	GenerateIncidentPDF(ctx context.Context, incidentData IncidentReportPDF) ([]byte, error)
	GenerateAndUploadIncidentPDF(ctx context.Context, incidentData IncidentReportPDF) (string, error)
}
