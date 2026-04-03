package pdf

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IncidentReportData struct {
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
	ClientFirstName         string    `json:"client_firstname"`
	ClientLastName          string    `json:"client_lastname"`
	LocationName            string    `json:"location_name"`
}

func (s *pdfService) GenerateIncidentPDF(
	ctx context.Context,
	incidentData IncidentReportData,
) ([]byte, error) {
	_ = ctx

	headerLines := []string{
		fmt.Sprintf("Incident ID: %s", incidentData.ID),
		fmt.Sprintf("Occurred at: %s", formatTimeOrNA(incidentData.OccurredAt)),
		fmt.Sprintf("Location: %s (%s)", incidentData.LocationName, incidentData.LocationID),
		fmt.Sprintf(
			"Reporter: %s %s (%s)",
			incidentData.EmployeeFirstName,
			incidentData.EmployeeLastName,
			incidentData.EmployeeID,
		),
		fmt.Sprintf(
			"Client: %s %s (%s)",
			incidentData.ClientFirstName,
			incidentData.ClientLastName,
			incidentData.ClientID,
		),
	}

	sections := []documentSection{
		{
			Title: "Incident details",
			Lines: []string{
				fmt.Sprintf("Reporter involvement: %s", incidentData.ReporterInvolvement),
				fmt.Sprintf("Incident type: %s", incidentData.IncidentType),
				fmt.Sprintf("Severity: %s", incidentData.SeverityOfIncident),
				fmt.Sprintf("Informed parties: %s", joinOrNA(incidentData.InformedParties)),
			},
		},
		{
			Title: "Impact and risk",
			Lines: []string{
				fmt.Sprintf(
					"Incident explanation: %s",
					stringOrNA(incidentData.IncidentExplanation),
				),
				fmt.Sprintf("Recurrence risk: %s", incidentData.RecurrenceRisk),
				fmt.Sprintf("Preventive steps: %s", stringOrNA(incidentData.IncidentPreventSteps)),
				fmt.Sprintf("Taken measures: %s", stringOrNA(incidentData.IncidentTakenMeasures)),
			},
		},
		{
			Title: "Cause analysis",
			Lines: []string{
				fmt.Sprintf("Cause categories: %s", joinOrNA(incidentData.CauseCategories)),
				fmt.Sprintf("Cause explanation: %s", stringOrNA(incidentData.CauseExplanation)),
			},
		},
		{
			Title: "Damage and follow-up",
			Lines: []string{
				fmt.Sprintf("Physical injury: %s", incidentData.PhysicalInjury),
				fmt.Sprintf(
					"Physical injury description: %s",
					stringOrNA(incidentData.PhysicalInjuryDesc),
				),
				fmt.Sprintf("Psychological damage: %s", incidentData.PsychologicalDamage),
				fmt.Sprintf(
					"Psychological damage description: %s",
					stringOrNA(incidentData.PsychologicalDamageDesc),
				),
				fmt.Sprintf("Needed consultation: %s", incidentData.NeededConsultation),
				fmt.Sprintf("Follow-up actions: %s", joinOrNA(incidentData.FollowUpActions)),
				fmt.Sprintf("Follow-up notes: %s", stringOrNA(incidentData.FollowUpNotes)),
				fmt.Sprintf("Employee absent: %s", yesNo(incidentData.IsEmployeeAbsent)),
				fmt.Sprintf("Additional details: %s", stringOrNA(incidentData.AdditionalDetails)),
			},
		},
	}

	pdfBytes, err := buildSectionsPDF("Incident report", headerLines, sections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate incident pdf: %w", err)
	}

	return pdfBytes, nil
}

func (s *pdfService) generateIncidentPDF(incidentData IncidentReportData) (multipart.File, error) {
	pdfBytes, err := s.GenerateIncidentPDF(context.Background(), incidentData)
	if err != nil {
		return nil, err
	}
	return toMultipartFile(pdfBytes), nil
}

func (s *pdfService) uploadIncidentPDF(
	ctx context.Context,
	pdfFile multipart.File,
	incidentID uuid.UUID,
) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf(
		"incident_reports/%s/incident_report_%s.pdf",
		timestamp,
		incidentID.String(),
	)

	key, _, err := s.bucketClient.Upload(ctx, pdfFile, filename, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to B2: %w", err)
	}
	return key, nil
}

func (s *pdfService) GenerateAndUploadIncidentPDF(
	ctx context.Context,
	incidentData IncidentReportData,
) (string, error) {
	pdfFile, err := s.generateIncidentPDF(incidentData)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	filename, err := s.uploadIncidentPDF(ctx, pdfFile, incidentData.ID)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	return filename, nil
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func stringOrNA(value *string) string {
	if value == nil || *value == "" {
		return "N/A"
	}
	return *value
}

func joinOrNA(values []string) string {
	if len(values) == 0 {
		return "N/A"
	}
	return strings.Join(values, "; ")
}

func formatTimeOrNA(value time.Time) string {
	if value.IsZero() {
		return "N/A"
	}
	return value.Format(time.RFC3339)
}
