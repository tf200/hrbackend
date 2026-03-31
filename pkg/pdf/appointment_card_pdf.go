package pdf

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type AppointmentCard struct {
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

func (s *pdfService) GenerateAppointmentCardPDF(_ context.Context, cardData AppointmentCard) ([]byte, error) {
	return s.generateAppointmentCardPDFBytes(cardData)
}

func (s *pdfService) generateAppointmentCardPDF(appointmentCardData AppointmentCard) (multipart.File, error) {
	pdfBytes, err := s.generateAppointmentCardPDFBytes(appointmentCardData)
	if err != nil {
		return nil, err
	}
	return toMultipartFile(pdfBytes), nil
}

func (s *pdfService) generateAppointmentCardPDFBytes(appointmentCardData AppointmentCard) ([]byte, error) {
	headerLines := []string{
		fmt.Sprintf("Client: %s", appointmentCardData.ClientName),
		fmt.Sprintf("Date: %s", appointmentCardData.Date),
		fmt.Sprintf("Mentor: %s", fallbackString(appointmentCardData.Mentor, "N/A")),
	}

	sections := []documentSection{
		{Title: "Algemene Informatie", Lines: appointmentCardData.GeneralInformation},
		{Title: "Belangrijke Contacten", Lines: appointmentCardData.ImportantContacts},
		{Title: "Huishouden", Lines: appointmentCardData.HouseholdInfo},
		{Title: "Organisatie Afspraken", Lines: appointmentCardData.OrganizationAgreements},
		{Title: "Jeugdreclassering Afspraken", Lines: appointmentCardData.YouthOfficerAgreements},
		{Title: "Behandel Afspraken", Lines: appointmentCardData.TreatmentAgreements},
		{Title: "Rookregels", Lines: appointmentCardData.SmokingRules},
		{Title: "Werk", Lines: appointmentCardData.Work},
		{Title: "School/Stage", Lines: appointmentCardData.SchoolInternship},
		{Title: "Reizen", Lines: appointmentCardData.Travel},
		{Title: "Verlof", Lines: appointmentCardData.Leave},
	}

	pdfBytes, err := buildSectionsPDF("Afsprakenkaart", headerLines, sections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate appointment card pdf: %w", err)
	}

	return pdfBytes, nil
}

func (s *pdfService) uploadAppointmentCardPDF(ctx context.Context, pdfFile multipart.File, appointmentCardID uuid.UUID) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("appointment_cards/%s/appointment_card_%s.pdf", timestamp, appointmentCardID.String())

	key, _, err := s.bucketClient.Upload(ctx, pdfFile, filename, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to B2: %w", err)
	}

	return key, nil
}

func (s *pdfService) GenerateAndUploadAppointmentCardPDF(ctx context.Context, cardData AppointmentCard) (string, error) {
	pdfFile, err := s.generateAppointmentCardPDF(cardData)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	fileURL, err := s.uploadAppointmentCardPDF(ctx, pdfFile, cardData.ID)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	return fileURL, nil
}

func fallbackString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
