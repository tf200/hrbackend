package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSignDocumentNotFound              = errors.New("sign document not found")
	ErrSignDocumentInvalidRequest        = errors.New("invalid sign document request")
	ErrSignDocumentInvalidStatus         = errors.New("invalid sign document status")
	ErrSignDocumentNotAuthorized         = errors.New("not authorized for sign document")
	ErrSignDocumentRecipientNotFound     = errors.New("sign document recipient not found")
	ErrSignDocumentRequiredFieldsMissing = errors.New("required signing fields are missing")
	ErrSignDocumentConsentRequired       = errors.New("signing consent is required")
	ErrSignDocumentSigningOrderBlocked   = errors.New("previous signers must sign first")
)

type SignDocument struct {
	ID                  uuid.UUID
	Title               string
	SourceAttachmentID  uuid.UUID
	SourceFileKey       string
	SignedFileKey       *string
	Status              string
	CreatedByEmployeeID uuid.UUID
	RelatedEntityType   *string
	RelatedEntityID     *uuid.UUID
	ExpiresAt           *time.Time
	SentAt              *time.Time
	CompletedAt         *time.Time
	CancelledAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Recipients          []SignDocumentRecipient
	Fields              []SignDocumentField
	Events              []SignDocumentEvent
}

type SignDocumentRecipient struct {
	ID            uuid.UUID
	DocumentID    uuid.UUID
	EmployeeID    uuid.UUID
	Name          string
	Email         *string
	SigningOrder  int32
	Status        string
	ViewedAt      *time.Time
	SignedAt      *time.Time
	DeclinedAt    *time.Time
	DeclineReason *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SignDocumentField struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	RecipientID uuid.UUID
	Type        string
	PageNumber  int32
	X           float64
	Y           float64
	Width       float64
	Height      float64
	Required    bool
	Label       *string
	Value       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EmployeeSignatureProfile struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	Type         string
	TypedName    *string
	ImageFileKey *string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SignDocumentSignature struct {
	ID                    uuid.UUID
	DocumentID            uuid.UUID
	RecipientID           uuid.UUID
	EmployeeID            uuid.UUID
	SignatureProfileID    *uuid.UUID
	SignatureText         *string
	SignatureImageFileKey *string
	ConsentText           string
	IPAddress             *string
	UserAgent             *string
	SignatureHash         string
	SignedAt              time.Time
}

type SignDocumentEvent struct {
	ID              uuid.UUID
	DocumentID      uuid.UUID
	RecipientID     *uuid.UUID
	ActorEmployeeID *uuid.UUID
	Event           string
	IPAddress       *string
	UserAgent       *string
	Metadata        any
	CreatedAt       time.Time
}

type CreateSignDocumentRecipientParams struct {
	EmployeeID   uuid.UUID
	SigningOrder int32
}

type CreateSignDocumentParams struct {
	Title              string
	SourceAttachmentID uuid.UUID
	Recipients         []CreateSignDocumentRecipientParams
	RelatedEntityType  *string
	RelatedEntityID    *uuid.UUID
	ExpiresAt          *time.Time
}

type UpsertSignDocumentFieldParams struct {
	RecipientID uuid.UUID
	Type        string
	PageNumber  int32
	X           float64
	Y           float64
	Width       float64
	Height      float64
	Required    bool
	Label       *string
}

type SignDocumentFieldValueParams struct {
	FieldID uuid.UUID
	Value   string
}

type SignDocumentSignParams struct {
	DocumentID             uuid.UUID
	SignatureText          *string
	SignatureImageFileKey  *string
	SaveSignatureForFuture bool
	FieldValues            []SignDocumentFieldValueParams
	ConsentAccepted        bool
	ConsentText            string
	IPAddress              *string
	UserAgent              *string
}

type SignDocumentRepository interface {
	WithTx(ctx context.Context, fn func(tx SignDocumentRepository) error) error
	CreateDocument(ctx context.Context, actorEmployeeID uuid.UUID, attachment *Attachment, params CreateSignDocumentParams) (*SignDocument, error)
	CreateRecipient(ctx context.Context, documentID uuid.UUID, params CreateSignDocumentRecipientParams) (*SignDocumentRecipient, error)
	GetDocumentByID(ctx context.Context, documentID uuid.UUID) (*SignDocument, error)
	ListDocumentsByCreator(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]SignDocument, error)
	ListDocumentsForEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]SignDocument, error)
	ListRecipients(ctx context.Context, documentID uuid.UUID) ([]SignDocumentRecipient, error)
	GetRecipientForEmployee(ctx context.Context, documentID, employeeID uuid.UUID) (*SignDocumentRecipient, error)
	ReplaceFields(ctx context.Context, documentID uuid.UUID, fields []UpsertSignDocumentFieldParams) ([]SignDocumentField, error)
	ListFields(ctx context.Context, documentID uuid.UUID) ([]SignDocumentField, error)
	ListFieldsForRecipient(ctx context.Context, documentID, recipientID uuid.UUID) ([]SignDocumentField, error)
	SendDocument(ctx context.Context, documentID uuid.UUID) (*SignDocument, error)
	MarkRecipientViewed(ctx context.Context, recipientID uuid.UUID) (*SignDocumentRecipient, error)
	CountUnsignedPriorRecipients(ctx context.Context, recipientID uuid.UUID) (int32, error)
	CreateSignatureProfile(ctx context.Context, employeeID uuid.UUID, typ string, typedName, imageFileKey *string, isDefault bool) (*EmployeeSignatureProfile, error)
	CreateSignature(ctx context.Context, params SignDocumentSignParams, recipient SignDocumentRecipient, profileID *uuid.UUID, signatureHash string) (*SignDocumentSignature, error)
	UpdateFieldValue(ctx context.Context, fieldID, recipientID uuid.UUID, value string) (*SignDocumentField, error)
	MarkRecipientSigned(ctx context.Context, recipientID uuid.UUID) (*SignDocumentRecipient, error)
	CountUnsignedRecipients(ctx context.Context, documentID uuid.UUID) (int32, error)
	MarkDocumentPartiallySigned(ctx context.Context, documentID uuid.UUID) (*SignDocument, error)
	MarkDocumentCompleted(ctx context.Context, documentID uuid.UUID, signedFileKey string) (*SignDocument, error)
	CancelDocument(ctx context.Context, documentID uuid.UUID) (*SignDocument, error)
	CreateEvent(ctx context.Context, event SignDocumentEvent) error
	ListEvents(ctx context.Context, documentID uuid.UUID) ([]SignDocumentEvent, error)
	ListSignatures(ctx context.Context, documentID uuid.UUID) ([]SignDocumentSignature, error)
}

type SignDocumentService interface {
	CreateDocument(ctx context.Context, actorEmployeeID uuid.UUID, params CreateSignDocumentParams) (*SignDocument, error)
	SetFields(ctx context.Context, actorEmployeeID, documentID uuid.UUID, fields []UpsertSignDocumentFieldParams) ([]SignDocumentField, error)
	SendDocument(ctx context.Context, actorEmployeeID, documentID uuid.UUID) (*SignDocument, error)
	GetDocument(ctx context.Context, actorEmployeeID, documentID uuid.UUID) (*SignDocument, error)
	ListMyCreatedDocuments(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]SignDocument, error)
	ListMySigningDocuments(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]SignDocument, error)
	GetMySigningDocument(ctx context.Context, employeeID, documentID uuid.UUID) (*SignDocument, error)
	MarkViewed(ctx context.Context, employeeID, documentID uuid.UUID, ipAddress, userAgent *string) (*SignDocumentRecipient, error)
	Sign(ctx context.Context, employeeID uuid.UUID, params SignDocumentSignParams) (*SignDocument, error)
	CancelDocument(ctx context.Context, actorEmployeeID, documentID uuid.UUID) (*SignDocument, error)
	GetSourceURL(ctx context.Context, employeeID, documentID uuid.UUID) (string, error)
	GetSignedURL(ctx context.Context, employeeID, documentID uuid.UUID) (string, error)
}
