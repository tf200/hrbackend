package pdf

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/phpdave11/gofpdf"

	"hrbackend/internal/domain"
)

func TestStampSignDocumentPDF(t *testing.T) {
	source := testPDF(t)
	documentID := uuid.New()
	recipientID := uuid.New()
	employeeID := uuid.New()
	fieldID := uuid.New()
	name := "Jane Employee"
	value := "Jane Employee"
	signedAt := time.Date(2026, 6, 1, 14, 30, 0, 0, time.UTC)

	service := NewSignDocumentStamper()
	out, err := service.StampSignDocumentPDF(context.Background(), domain.SignDocumentPDFStampInput{
		SourcePDF: source,
		Recipients: []domain.SignDocumentRecipient{{
			ID:         recipientID,
			DocumentID: documentID,
			EmployeeID: employeeID,
			Name:       name,
		}},
		Fields: []domain.SignDocumentField{{
			ID:          fieldID,
			DocumentID:  documentID,
			RecipientID: recipientID,
			Type:        "signature",
			PageNumber:  1,
			X:           0.1,
			Y:           0.7,
			Width:       0.5,
			Height:      0.1,
			Required:    true,
			Value:       &value,
		}},
		Signatures: []domain.SignDocumentSignature{{
			DocumentID:    documentID,
			RecipientID:   recipientID,
			EmployeeID:    employeeID,
			SignatureText: &name,
			ConsentText:   "I agree",
			SignatureHash: "hash",
			SignedAt:      signedAt,
		}},
	})
	if err != nil {
		t.Fatalf("StampSignDocumentPDF returned error: %v", err)
	}
	if bytes.Equal(source, out) {
		t.Fatal("expected stamped PDF to differ from source PDF")
	}
	if _, err := api.PageCount(bytes.NewReader(out), model.NewDefaultConfiguration()); err != nil {
		t.Fatalf("stamped PDF is not valid: %v", err)
	}
}

func TestStampSignDocumentPDFRejectsInvalidPage(t *testing.T) {
	service := NewSignDocumentStamper()
	_, err := service.StampSignDocumentPDF(context.Background(), domain.SignDocumentPDFStampInput{
		SourcePDF: testPDF(t),
		Fields: []domain.SignDocumentField{{
			ID:         uuid.New(),
			Type:       "text",
			PageNumber: 2,
			X:          0.1,
			Y:          0.1,
			Width:      0.2,
			Height:     0.1,
		}},
	})
	if err == nil {
		t.Fatal("expected invalid page error")
	}
}

func testPDF(t *testing.T) []byte {
	t.Helper()
	doc := gofpdf.New("P", "pt", "A4", "")
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	doc.Text(72, 72, "Document body")
	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("generate test PDF: %v", err)
	}
	return buf.Bytes()
}
