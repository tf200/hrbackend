package pdf

import (
	"context"
	"fmt"
	"mime/multipart"
	"sort"
	"time"

	"github.com/google/uuid"
)

type InvoicePDFData struct {
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
	InvoiceDetails      []InvoiceDetail
	TotalAmount         float64
	ExtraItems          map[string]string
}

type InvoiceDetail struct {
	CareType      string
	Periods       []InvoicePeriod
	Price         float64
	PriceTimeUnit string
	PreVatTotal   float64
	Total         float64
}

type InvoicePeriod struct {
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	AcommodationTimeFrame string    `json:"accommodation_time_frame,omitempty"`
	AmbulanteTotalMinutes float64   `json:"ambulante_total_minutes,omitempty"`
}

func sumPreVat(details []InvoiceDetail) float64 {
	var total float64
	for _, d := range details {
		total += d.PreVatTotal
	}
	return total
}

func sumVat(details []InvoiceDetail) float64 {
	var total float64
	for _, d := range details {
		total += d.Total - d.PreVatTotal
	}
	return total
}

func (s *pdfService) generateInvoicePDF(invoiceData InvoicePDFData) (multipart.File, error) {
	totalAmount := invoiceData.TotalAmount
	if totalAmount == 0 {
		for _, detail := range invoiceData.InvoiceDetails {
			totalAmount += detail.Total
		}
	}

	headerLines := []string{
		fmt.Sprintf("Invoice number: %s", invoiceData.InvoiceNumber),
		fmt.Sprintf("Invoice date: %s", invoiceData.InvoiceDate.Format("2006-01-02")),
		fmt.Sprintf("Due date: %s", invoiceData.DueDate.Format("2006-01-02")),
		fmt.Sprintf("Sender: %s", invoiceData.SenderName),
		fmt.Sprintf("Sender contact: %s", invoiceData.SenderContactPerson),
		fmt.Sprintf("Sender address: %s %s, %s %s", invoiceData.SenderStreet, invoiceData.SenderHouseNumber, invoiceData.SenderPostalCode, invoiceData.SenderCity),
	}

	sections := make([]documentSection, 0, len(invoiceData.InvoiceDetails)+2)
	for idx, detail := range invoiceData.InvoiceDetails {
		lines := []string{
			fmt.Sprintf("Care type: %s", detail.CareType),
			fmt.Sprintf("Price: %s per %s", formatCurrency(detail.Price), detail.PriceTimeUnit),
		}

		for periodIdx, period := range detail.Periods {
			periodLine := fmt.Sprintf("Period %d: %s to %s", periodIdx+1, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
			lines = append(lines, periodLine)
			if period.AcommodationTimeFrame != "" {
				lines = append(lines, fmt.Sprintf("Accommodation timeframe: %s", period.AcommodationTimeFrame))
			}
			if period.AmbulanteTotalMinutes > 0 {
				lines = append(lines, fmt.Sprintf("Ambulante minutes: %.2f", period.AmbulanteTotalMinutes))
			}
		}

		lines = append(lines,
			fmt.Sprintf("Subtotal (excl. VAT): %s", formatCurrency(detail.PreVatTotal)),
			fmt.Sprintf("Total (incl. VAT): %s", formatCurrency(detail.Total)),
		)

		sections = append(sections, documentSection{
			Title: fmt.Sprintf("Invoice detail #%d", idx+1),
			Lines: lines,
		})
	}

	if len(invoiceData.ExtraItems) > 0 {
		keys := make([]string, 0, len(invoiceData.ExtraItems))
		for key := range invoiceData.ExtraItems {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		extraLines := make([]string, 0, len(keys))
		for _, key := range keys {
			extraLines = append(extraLines, fmt.Sprintf("%s: %s", key, invoiceData.ExtraItems[key]))
		}

		sections = append(sections, documentSection{
			Title: "Extra items",
			Lines: extraLines,
		})
	}

	sections = append(sections, documentSection{
		Title: "Totals",
		Lines: []string{
			fmt.Sprintf("Total excl. VAT: %s", formatCurrency(sumPreVat(invoiceData.InvoiceDetails))),
			fmt.Sprintf("VAT total: %s", formatCurrency(sumVat(invoiceData.InvoiceDetails))),
			fmt.Sprintf("Total incl. VAT: %s", formatCurrency(totalAmount)),
		},
	})

	pdfBytes, err := buildSectionsPDF(fmt.Sprintf("Invoice %s", invoiceData.InvoiceNumber), headerLines, sections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice pdf: %w", err)
	}

	return toMultipartFile(pdfBytes), nil
}

func formatCurrency(value float64) string {
	return fmt.Sprintf("EUR %.2f", value)
}

func (s *pdfService) uploadInvoicePDF(ctx context.Context, pdfFile multipart.File, invoiceID uuid.UUID) (string, int64, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("invoice_reports/%s/invoice_report_%s.pdf", timestamp, invoiceID.String())

	key, size, err := s.bucketClient.Upload(ctx, pdfFile, filename, "application/pdf")
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload PDF to B2: %w", err)
	}
	return key, size, nil
}

func (s *pdfService) GenerateAndUploadInvoicePDF(ctx context.Context, invoiceData InvoicePDFData) (string, int64, error) {
	pdfFile, err := s.generateInvoicePDF(invoiceData)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate PDF: %w", err)
	}

	fileURL, size, err := s.uploadInvoicePDF(ctx, pdfFile, invoiceData.ID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload PDF: %w", err)
	}

	return fileURL, size, nil
}
