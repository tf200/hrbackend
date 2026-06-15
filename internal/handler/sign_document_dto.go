package handler

import (
	"time"

	"github.com/google/uuid"

	"hrbackend/internal/domain"
)

type createSignDocumentRequest struct {
	Title              string                        `json:"title"                binding:"required"`
	SourceAttachmentID uuid.UUID                     `json:"source_attachment_id" binding:"required"`
	Recipients         []createSignDocumentRecipient `json:"recipients"           binding:"required"`
	RelatedEntityType  *string                       `json:"related_entity_type"`
	RelatedEntityID    *uuid.UUID                    `json:"related_entity_id"`
	ExpiresAt          *time.Time                    `json:"expires_at"`
}

type createSignDocumentRecipient struct {
	EmployeeID   uuid.UUID `json:"employee_id"   binding:"required"`
	SigningOrder int32     `json:"signing_order"`
}

type setSignDocumentFieldsRequest struct {
	Fields []signDocumentFieldInput `json:"fields" binding:"required"`
}

type signDocumentFieldInput struct {
	RecipientID uuid.UUID `json:"recipient_id" binding:"required"`
	Type        string    `json:"type"         binding:"required"`
	PageNumber  int32     `json:"page_number"  binding:"required"`
	X           float64   `json:"x"            binding:"required"`
	Y           float64   `json:"y"            binding:"required"`
	Width       float64   `json:"width"        binding:"required"`
	Height      float64   `json:"height"       binding:"required"`
	Required    *bool     `json:"required"`
	Label       *string   `json:"label"`
}

type signSignDocumentRequest struct {
	FieldValues     []signDocumentFieldValueIn `json:"field_values"     binding:"required"`
	ConsentAccepted bool                       `json:"consent_accepted" binding:"required"`
	ConsentText     string                     `json:"consent_text"`
}

type signDocumentFieldValueIn struct {
	FieldID uuid.UUID `json:"field_id" binding:"required"`
	Value   string    `json:"value"    binding:"required"`
}

type signDocumentURLResponse struct {
	URL string `json:"url"`
}

func toCreateSignDocumentParams(req createSignDocumentRequest) domain.CreateSignDocumentParams {
	recipients := make([]domain.CreateSignDocumentRecipientParams, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		recipients = append(
			recipients,
			domain.CreateSignDocumentRecipientParams{
				EmployeeID:   r.EmployeeID,
				SigningOrder: r.SigningOrder,
			},
		)
	}
	return domain.CreateSignDocumentParams{
		Title:              req.Title,
		SourceAttachmentID: req.SourceAttachmentID,
		Recipients:         recipients,
		RelatedEntityType:  req.RelatedEntityType,
		RelatedEntityID:    req.RelatedEntityID,
		ExpiresAt:          req.ExpiresAt,
	}
}

func toSignDocumentFieldParams(
	inputs []signDocumentFieldInput,
) []domain.UpsertSignDocumentFieldParams {
	fields := make([]domain.UpsertSignDocumentFieldParams, 0, len(inputs))
	for _, input := range inputs {
		required := true
		if input.Required != nil {
			required = *input.Required
		}
		fields = append(
			fields,
			domain.UpsertSignDocumentFieldParams{
				RecipientID: input.RecipientID,
				Type:        input.Type,
				PageNumber:  input.PageNumber,
				X:           input.X,
				Y:           input.Y,
				Width:       input.Width,
				Height:      input.Height,
				Required:    required,
				Label:       input.Label,
			},
		)
	}
	return fields
}

func toSignDocumentSignParams(
	documentID uuid.UUID,
	req signSignDocumentRequest,
	ipAddress, userAgent *string,
) domain.SignDocumentSignParams {
	values := make([]domain.SignDocumentFieldValueParams, 0, len(req.FieldValues))
	for _, value := range req.FieldValues {
		values = append(
			values,
			domain.SignDocumentFieldValueParams{FieldID: value.FieldID, Value: value.Value},
		)
	}
	return domain.SignDocumentSignParams{
		DocumentID:      documentID,
		FieldValues:     values,
		ConsentAccepted: req.ConsentAccepted,
		ConsentText:     req.ConsentText,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	}
}
