package pdf

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"hrbackend/internal/domain"
)

type SignDocumentStamper struct{}

func NewSignDocumentStamper() *SignDocumentStamper {
	return &SignDocumentStamper{}
}

func (s *SignDocumentStamper) StampSignDocumentPDF(ctx context.Context, input domain.SignDocumentPDFStampInput) ([]byte, error) {
	if len(input.SourcePDF) == 0 {
		return nil, fmt.Errorf("source PDF is empty")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	reader := bytes.NewReader(input.SourcePDF)
	pageDims, err := api.PageDims(reader, model.NewDefaultConfiguration())
	if err != nil {
		return nil, fmt.Errorf("read PDF page dimensions: %w", err)
	}
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, err
	}

	stamps, err := signDocumentWatermarks(input, pageDims)
	if err != nil {
		return nil, err
	}
	if len(stamps) == 0 {
		return input.SourcePDF, nil
	}

	var out bytes.Buffer
	if err := api.AddWatermarksSliceMap(reader, &out, stamps, model.NewDefaultConfiguration()); err != nil {
		return nil, fmt.Errorf("stamp sign document PDF: %w", err)
	}
	return out.Bytes(), nil
}

func signDocumentWatermarks(input domain.SignDocumentPDFStampInput, pageDims []types.Dim) (map[int][]*model.Watermark, error) {
	recipients := make(map[string]domain.SignDocumentRecipient, len(input.Recipients))
	for _, recipient := range input.Recipients {
		recipients[recipient.ID.String()] = recipient
	}
	signatures := make(map[string]domain.SignDocumentSignature, len(input.Signatures))
	for _, signature := range input.Signatures {
		signatures[signature.RecipientID.String()] = signature
	}

	stamps := map[int][]*model.Watermark{}
	for _, field := range input.Fields {
		if field.PageNumber <= 0 || int(field.PageNumber) > len(pageDims) {
			return nil, fmt.Errorf("field %s references invalid page %d", field.ID, field.PageNumber)
		}
		if field.X < 0 || field.Y < 0 || field.Width <= 0 || field.Height <= 0 || field.X+field.Width > 1 || field.Y+field.Height > 1 {
			return nil, fmt.Errorf("field %s has invalid coordinates", field.ID)
		}

		text, ok := signDocumentFieldStampText(field, recipients[field.RecipientID.String()], signatures[field.RecipientID.String()])
		if !ok {
			continue
		}
		wm, err := signDocumentTextWatermark(text, field, pageDims[field.PageNumber-1])
		if err != nil {
			return nil, err
		}
		stamps[int(field.PageNumber)] = append(stamps[int(field.PageNumber)], wm)
	}
	return stamps, nil
}

func signDocumentFieldStampText(field domain.SignDocumentField, recipient domain.SignDocumentRecipient, signature domain.SignDocumentSignature) (string, bool) {
	value := ""
	if field.Value != nil {
		value = strings.TrimSpace(*field.Value)
	}

	switch field.Type {
	case "signature":
		name := value
		if name == "" && signature.SignatureText != nil {
			name = strings.TrimSpace(*signature.SignatureText)
		}
		if name == "" {
			name = recipient.Name
		}
		if name == "" {
			return "", false
		}
		return fmt.Sprintf("%s\nSigned electronically on %s", name, signature.SignedAt.Format("2006-01-02 15:04")), true
	case "initials", "text":
		return value, value != ""
	case "date":
		if value != "" {
			return value, true
		}
		if !signature.SignedAt.IsZero() {
			return signature.SignedAt.Format(time.DateOnly), true
		}
	case "checkbox":
		if signDocumentTruthy(value) {
			return "X", true
		}
	}
	return "", false
}

func signDocumentTextWatermark(text string, field domain.SignDocumentField, dim types.Dim) (*model.Watermark, error) {
	x := field.X * dim.Width
	y := field.Y * dim.Height
	fontSize := int(math.Max(8, math.Min(16, field.Height*dim.Height*0.45)))
	if field.Type == "signature" {
		fontSize = int(math.Max(9, math.Min(18, field.Height*dim.Height*0.32)))
	}

	desc := fmt.Sprintf("font:%s, points:%d, scale:1 abs, pos:tl, off:%s -%s, fillc:#111111, rot:0", signDocumentFont(field.Type), fontSize, signDocumentFloat(x), signDocumentFloat(y))
	return api.TextWatermark(text, desc, true, false, types.POINTS)
}

func signDocumentFont(fieldType string) string {
	if fieldType == "signature" {
		return "Times-Italic"
	}
	return "Helvetica"
}

func signDocumentFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func signDocumentTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "checked", "x", "on":
		return true
	default:
		return false
	}
}
