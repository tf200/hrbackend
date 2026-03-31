package pdf

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"
)

type ContractData struct {
	ID     int64  `json:"ContractID"`
	Status string `json:"Status"`

	StartDate      string `json:"StartDate"`
	EndDate        string `json:"EndDate"`
	ReminderPeriod int    `json:"ReminderPeriod"`

	SenderName        string `json:"SenderName"`
	SenderStreet      string `json:"SenderStreet"`
	SenderHouseNumber string `json:"SenderHouseNumber"`
	SenderPostalCode  string `json:"SenderPostalCode"`
	SenderCity        string `json:"SenderCity"`
	SenderContactInfo string `json:"SenderContactInfo"`

	ClientFirstName   string `json:"ClientFirstName"`
	ClientLastName    string `json:"ClientLastName"`
	ClientAddress     string `json:"ClientAddress"`
	ClientContactInfo string `json:"ClientContactInfo"`

	CareType        string `json:"CareType"`
	CareName        string `json:"CareName"`
	FinancingAct    string `json:"FinancingAct"`
	FinancingOption string `json:"FinancingOption"`

	Hours            float64 `json:"Hours"`
	HoursType        string  `json:"HoursType"`
	AmbulanteDisplay string  `json:"AmbulanteDisplay"`

	Price          float64 `json:"Price"`
	PriceTimeUnit  string  `json:"PriceTimeUnit"`
	Vat            float64 `json:"Vat"`
	TypeName       string  `json:"TypeName"`
	GenerationDate string  `json:"GenerationDate"`
}

func (s *pdfService) generateContractPDF(contractData ContractData) (multipart.File, error) {
	headerLines := []string{
		fmt.Sprintf("Contract ID: %d", contractData.ID),
		fmt.Sprintf("Status: %s", contractData.Status),
		fmt.Sprintf("Generation date: %s", contractData.GenerationDate),
		fmt.Sprintf("Period: %s to %s", contractData.StartDate, contractData.EndDate),
		fmt.Sprintf("Reminder period (days): %d", contractData.ReminderPeriod),
	}

	sections := []documentSection{
		{
			Title: "Sender",
			Lines: []string{
				fmt.Sprintf("Name: %s", contractData.SenderName),
				fmt.Sprintf("Address: %s %s, %s %s", contractData.SenderStreet, contractData.SenderHouseNumber, contractData.SenderPostalCode, contractData.SenderCity),
				fmt.Sprintf("Contact info: %s", contractData.SenderContactInfo),
			},
		},
		{
			Title: "Client",
			Lines: []string{
				fmt.Sprintf("Name: %s %s", contractData.ClientFirstName, contractData.ClientLastName),
				fmt.Sprintf("Address: %s", contractData.ClientAddress),
				fmt.Sprintf("Contact info: %s", contractData.ClientContactInfo),
			},
		},
		{
			Title: "Care specification",
			Lines: []string{
				fmt.Sprintf("Care type: %s", contractData.CareType),
				fmt.Sprintf("Care name: %s", contractData.CareName),
				fmt.Sprintf("Financing act: %s", contractData.FinancingAct),
				fmt.Sprintf("Financing option: %s", contractData.FinancingOption),
				fmt.Sprintf("Hours: %.2f (%s)", contractData.Hours, contractData.HoursType),
			},
		},
		{
			Title: "Financial terms",
			Lines: []string{
				fmt.Sprintf("Price: EUR %.2f per %s", contractData.Price, contractData.PriceTimeUnit),
				fmt.Sprintf("VAT: %.2f%%", contractData.Vat),
				fmt.Sprintf("Contract type: %s", contractData.TypeName),
			},
		},
	}

	pdfBytes, err := buildSectionsPDF("Care Contract", headerLines, sections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contract pdf: %w", err)
	}

	return toMultipartFile(pdfBytes), nil
}

func (s *pdfService) uploadContractPDF(ctx context.Context, pdfFile multipart.File, contractID int64) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("contract/%s/contract-%d.pdf", timestamp, contractID)

	key, _, err := s.bucketClient.Upload(ctx, pdfFile, filename, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to B2: %w", err)
	}
	return key, nil
}

func (s *pdfService) GenerateAndUploadContractPDF(ctx context.Context, contractData ContractData) (string, error) {
	pdfFile, err := s.generateContractPDF(contractData)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	fileURL, err := s.uploadContractPDF(ctx, pdfFile, contractData.ID)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	return fileURL, nil
}
