package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"hrbackend/internal/domain"
)

const signDocumentConsentText = "I agree to electronically sign this document."

type SignDocumentService struct {
	repo                domain.SignDocumentRepository
	attachmentRepo      domain.AttachmentRepository
	employeeRepo        domain.EmployeeRepository
	notificationService domain.NotificationService
	storage             domain.Storage
	pdfStamper          domain.SignDocumentPDFStamper
	logger              domain.Logger
}

func NewSignDocumentService(
	repo domain.SignDocumentRepository,
	attachmentRepo domain.AttachmentRepository,
	employeeRepo domain.EmployeeRepository,
	notificationService domain.NotificationService,
	storage domain.Storage,
	pdfStamper domain.SignDocumentPDFStamper,
	logger domain.Logger,
) domain.SignDocumentService {
	return &SignDocumentService{
		repo:                repo,
		attachmentRepo:      attachmentRepo,
		employeeRepo:        employeeRepo,
		notificationService: notificationService,
		storage:             storage,
		pdfStamper:          pdfStamper,
		logger:              logger,
	}
}

func (s *SignDocumentService) CreateDocument(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreateSignDocumentParams,
) (*domain.SignDocument, error) {
	if actorEmployeeID == uuid.Nil || params.SourceAttachmentID == uuid.Nil ||
		strings.TrimSpace(params.Title) == "" ||
		len(params.Recipients) == 0 {
		return nil, domain.ErrSignDocumentInvalidRequest
	}
	params.Title = strings.TrimSpace(params.Title)
	attachment, err := s.attachmentRepo.GetAttachment(ctx, params.SourceAttachmentID)
	if err != nil {
		return nil, err
	}

	var doc *domain.SignDocument
	err = s.repo.WithTx(ctx, func(tx domain.SignDocumentRepository) error {
		created, err := tx.CreateDocument(ctx, actorEmployeeID, attachment, params)
		if err != nil {
			return err
		}
		doc = created
		for _, recipient := range params.Recipients {
			if recipient.EmployeeID == uuid.Nil {
				return domain.ErrSignDocumentInvalidRequest
			}
			if recipient.SigningOrder <= 0 {
				recipient.SigningOrder = 1
			}
			createdRecipient, err := tx.CreateRecipient(ctx, created.ID, recipient)
			if err != nil {
				return err
			}
			doc.Recipients = append(doc.Recipients, *createdRecipient)
		}
		return tx.CreateEvent(
			ctx,
			domain.SignDocumentEvent{
				DocumentID:      created.ID,
				ActorEmployeeID: &actorEmployeeID,
				Event:           "created",
				Metadata: map[string]any{
					"source_attachment_id": params.SourceAttachmentID.String(),
				},
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *SignDocumentService) SetFields(
	ctx context.Context,
	actorEmployeeID, documentID uuid.UUID,
	fields []domain.UpsertSignDocumentFieldParams,
) ([]domain.SignDocumentField, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CreatedByEmployeeID != actorEmployeeID {
		return nil, domain.ErrSignDocumentNotAuthorized
	}
	if doc.Status != "draft" {
		return nil, domain.ErrSignDocumentInvalidStatus
	}
	if len(fields) == 0 {
		return nil, domain.ErrSignDocumentRequiredFieldsMissing
	}
	recipients, err := s.repo.ListRecipients(ctx, documentID)
	if err != nil {
		return nil, err
	}
	recipientSet := map[uuid.UUID]bool{}
	for _, recipient := range recipients {
		recipientSet[recipient.ID] = true
	}
	for _, field := range fields {
		if !recipientSet[field.RecipientID] || field.PageNumber <= 0 || field.Width <= 0 ||
			field.Height <= 0 ||
			field.X < 0 ||
			field.Y < 0 ||
			field.X > 1 ||
			field.Y > 1 ||
			field.Width > 1 ||
			field.Height > 1 {
			return nil, domain.ErrSignDocumentInvalidRequest
		}
		if field.Type == "" {
			return nil, domain.ErrSignDocumentInvalidRequest
		}
	}
	return s.repo.ReplaceFields(ctx, documentID, fields)
}

func (s *SignDocumentService) SendDocument(
	ctx context.Context,
	actorEmployeeID, documentID uuid.UUID,
) (*domain.SignDocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CreatedByEmployeeID != actorEmployeeID {
		return nil, domain.ErrSignDocumentNotAuthorized
	}
	if doc.Status != "draft" {
		return nil, domain.ErrSignDocumentInvalidStatus
	}
	fields, err := s.repo.ListFields(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, domain.ErrSignDocumentRequiredFieldsMissing
	}

	var sent *domain.SignDocument
	err = s.repo.WithTx(ctx, func(tx domain.SignDocumentRepository) error {
		var err error
		sent, err = tx.SendDocument(ctx, documentID)
		if err != nil {
			return err
		}
		return tx.CreateEvent(
			ctx,
			domain.SignDocumentEvent{
				DocumentID:      documentID,
				ActorEmployeeID: &actorEmployeeID,
				Event:           "sent",
				Metadata:        map[string]any{},
			},
		)
	})
	if err != nil {
		return nil, err
	}

	hydrated, err := s.hydrate(ctx, sent)
	if err != nil {
		return nil, err
	}

	// Trigger notifications to recipients
	var recipientEmployeeIDs []uuid.UUID
	for _, r := range hydrated.Recipients {
		if r.EmployeeID != uuid.Nil && r.EmployeeID != actorEmployeeID {
			recipientEmployeeIDs = append(recipientEmployeeIDs, r.EmployeeID)
		}
	}

	if s.notificationService != nil && len(recipientEmployeeIDs) > 0 {
		requesterName := "Someone"
		if s.employeeRepo != nil {
			emp, err := s.employeeRepo.GetEmployeeByID(ctx, actorEmployeeID)
			if err == nil && emp != nil {
				requesterName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				EmployeeIDs: recipientEmployeeIDs,
			},
			Message: fmt.Sprintf(
				"%s requested your signature on a document: %s",
				requesterName,
				hydrated.Title,
			),
			Data: domain.SignDocumentRequestedNotificationData{
				DocumentID:          hydrated.ID,
				DocumentTitle:       hydrated.Title,
				RequesterEmployeeID: actorEmployeeID,
				RequesterName:       requesterName,
			},
		})
	}

	return hydrated, nil
}

func (s *SignDocumentService) GetDocument(
	ctx context.Context,
	actorEmployeeID, documentID uuid.UUID,
) (*domain.SignDocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CreatedByEmployeeID != actorEmployeeID {
		return nil, domain.ErrSignDocumentNotAuthorized
	}
	return s.hydrate(ctx, doc)
}

func (s *SignDocumentService) ListMyCreatedDocuments(
	ctx context.Context,
	employeeID uuid.UUID,
	limit, offset int32,
) ([]domain.SignDocument, error) {
	return s.repo.ListDocumentsByCreator(ctx, employeeID, normalizeLimit(limit), offset)
}

func (s *SignDocumentService) ListMySigningDocuments(
	ctx context.Context,
	employeeID uuid.UUID,
	limit, offset int32,
) ([]domain.SignDocument, error) {
	return s.repo.ListDocumentsForEmployee(ctx, employeeID, normalizeLimit(limit), offset)
}

func (s *SignDocumentService) GetMySigningDocument(
	ctx context.Context,
	employeeID, documentID uuid.UUID,
) (*domain.SignDocument, error) {
	if _, err := s.repo.GetRecipientForEmployee(ctx, documentID, employeeID); err != nil {
		return nil, err
	}
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return s.hydrate(ctx, doc)
}

func (s *SignDocumentService) MarkViewed(
	ctx context.Context,
	employeeID, documentID uuid.UUID,
	ipAddress, userAgent *string,
) (*domain.SignDocumentRecipient, error) {
	recipient, err := s.repo.GetRecipientForEmployee(ctx, documentID, employeeID)
	if err != nil {
		return nil, err
	}
	viewed, err := s.repo.MarkRecipientViewed(ctx, recipient.ID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.CreateEvent(
		ctx,
		domain.SignDocumentEvent{
			DocumentID:      documentID,
			RecipientID:     &recipient.ID,
			ActorEmployeeID: &employeeID,
			Event:           "viewed",
			IPAddress:       ipAddress,
			UserAgent:       userAgent,
			Metadata:        map[string]any{},
		},
	)
	return viewed, nil
}

func (s *SignDocumentService) Sign(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.SignDocumentSignParams,
) (*domain.SignDocument, error) {
	if employeeID == uuid.Nil || params.DocumentID == uuid.Nil || !params.ConsentAccepted {
		return nil, domain.ErrSignDocumentConsentRequired
	}
	if params.ConsentText == "" {
		params.ConsentText = signDocumentConsentText
	}
	doc, err := s.repo.GetDocumentByID(ctx, params.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Status != "sent" && doc.Status != "partially_signed" {
		return nil, domain.ErrSignDocumentInvalidStatus
	}
	recipient, err := s.repo.GetRecipientForEmployee(ctx, params.DocumentID, employeeID)
	if err != nil {
		return nil, err
	}
	if recipient.Status == "signed" {
		return nil, domain.ErrSignDocumentInvalidStatus
	}
	prior, err := s.repo.CountUnsignedPriorRecipients(ctx, recipient.ID)
	if err != nil {
		return nil, err
	}
	if prior > 0 {
		return nil, domain.ErrSignDocumentSigningOrderBlocked
	}
	fields, err := s.repo.ListFieldsForRecipient(ctx, params.DocumentID, recipient.ID)
	if err != nil {
		return nil, err
	}
	if !hasRequiredFieldValues(fields, params.FieldValues) {
		return nil, domain.ErrSignDocumentRequiredFieldsMissing
	}

	var signedDoc *domain.SignDocument
	err = s.repo.WithTx(ctx, func(tx domain.SignDocumentRepository) error {
		var profileID *uuid.UUID
		if params.SaveSignatureForFuture {
			typ := "typed"
			if params.SignatureImageFileKey != nil {
				typ = "drawn"
			}
			profile, err := tx.CreateSignatureProfile(
				ctx,
				employeeID,
				typ,
				params.SignatureText,
				params.SignatureImageFileKey,
				true,
			)
			if err != nil {
				return err
			}
			profileID = &profile.ID
		}
		hash := signatureHash(params, employeeID, recipient.ID)
		if _, err := tx.CreateSignature(ctx, params, *recipient, profileID, hash); err != nil {
			return err
		}
		for _, fieldValue := range params.FieldValues {
			if _, err := tx.UpdateFieldValue(ctx, fieldValue.FieldID, recipient.ID, fieldValue.Value); err != nil {
				return err
			}
		}
		if _, err := tx.MarkRecipientSigned(ctx, recipient.ID); err != nil {
			return err
		}
		remaining, err := tx.CountUnsignedRecipients(ctx, params.DocumentID)
		if err != nil {
			return err
		}
		if remaining > 0 {
			signedDoc, err = tx.MarkDocumentPartiallySigned(ctx, params.DocumentID)
			if err != nil {
				return err
			}
		} else {
			signedKey, err := s.createSignedPDF(ctx, *doc)
			if err != nil {
				return err
			}
			signedDoc, err = tx.MarkDocumentCompleted(ctx, params.DocumentID, signedKey)
			if err != nil {
				return err
			}
			if err := tx.CreateEvent(ctx, domain.SignDocumentEvent{DocumentID: params.DocumentID, ActorEmployeeID: &employeeID, Event: "completed", IPAddress: params.IPAddress, UserAgent: params.UserAgent, Metadata: map[string]any{}}); err != nil {
				return err
			}
		}
		return tx.CreateEvent(
			ctx,
			domain.SignDocumentEvent{
				DocumentID:      params.DocumentID,
				RecipientID:     &recipient.ID,
				ActorEmployeeID: &employeeID,
				Event:           "signed",
				IPAddress:       params.IPAddress,
				UserAgent:       params.UserAgent,
				Metadata:        map[string]any{"signature_hash": hash},
			},
		)
	})
	if err != nil {
		return nil, err
	}

	hydrated, err := s.hydrate(ctx, signedDoc)
	if err != nil {
		return nil, err
	}

	// Trigger notification to document creator
	if s.notificationService != nil && doc.CreatedByEmployeeID != uuid.Nil &&
		doc.CreatedByEmployeeID != employeeID {
		signerName := "Someone"
		if s.employeeRepo != nil {
			emp, err := s.employeeRepo.GetEmployeeByID(ctx, employeeID)
			if err == nil && emp != nil {
				signerName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		isCompleted := (hydrated.Status == "completed")
		var message string
		if isCompleted {
			message = fmt.Sprintf(
				"%s has signed the document: %s. The document is now fully signed and completed.",
				signerName,
				hydrated.Title,
			)
		} else {
			message = fmt.Sprintf("%s has signed the document: %s.", signerName, hydrated.Title)
		}

		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				EmployeeIDs: []uuid.UUID{doc.CreatedByEmployeeID},
			},
			Message: message,
			Data: domain.SignDocumentSignedNotificationData{
				DocumentID:       hydrated.ID,
				DocumentTitle:    hydrated.Title,
				SignerEmployeeID: employeeID,
				SignerName:       signerName,
				IsCompleted:      isCompleted,
			},
		})
	}

	return hydrated, nil
}

func (s *SignDocumentService) CancelDocument(
	ctx context.Context,
	actorEmployeeID, documentID uuid.UUID,
) (*domain.SignDocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CreatedByEmployeeID != actorEmployeeID {
		return nil, domain.ErrSignDocumentNotAuthorized
	}
	cancelled, err := s.repo.CancelDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	_ = s.repo.CreateEvent(
		ctx,
		domain.SignDocumentEvent{
			DocumentID:      documentID,
			ActorEmployeeID: &actorEmployeeID,
			Event:           "cancelled",
			Metadata:        map[string]any{},
		},
	)
	return cancelled, nil
}

func (s *SignDocumentService) GetSourceURL(
	ctx context.Context,
	employeeID, documentID uuid.UUID,
) (string, error) {
	doc, err := s.authorizedDocument(ctx, employeeID, documentID)
	if err != nil {
		return "", err
	}
	return s.storage.GeneratePresignedURL(ctx, doc.SourceFileKey, 15*time.Minute)
}

func (s *SignDocumentService) GetSignedURL(
	ctx context.Context,
	employeeID, documentID uuid.UUID,
) (string, error) {
	doc, err := s.authorizedDocument(ctx, employeeID, documentID)
	if err != nil {
		return "", err
	}
	if doc.SignedFileKey == nil {
		return "", domain.ErrSignDocumentInvalidStatus
	}
	return s.storage.GeneratePresignedURL(ctx, *doc.SignedFileKey, 15*time.Minute)
}

func (s *SignDocumentService) authorizedDocument(
	ctx context.Context,
	employeeID, documentID uuid.UUID,
) (*domain.SignDocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CreatedByEmployeeID == employeeID {
		return doc, nil
	}
	if _, err := s.repo.GetRecipientForEmployee(ctx, documentID, employeeID); err == nil {
		return doc, nil
	}
	return nil, domain.ErrSignDocumentNotAuthorized
}

func (s *SignDocumentService) hydrate(
	ctx context.Context,
	doc *domain.SignDocument,
) (*domain.SignDocument, error) {
	if doc == nil {
		return nil, domain.ErrSignDocumentNotFound
	}
	recipients, err := s.repo.ListRecipients(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	fields, err := s.repo.ListFields(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	events, err := s.repo.ListEvents(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	doc.Recipients = recipients
	doc.Fields = fields
	doc.Events = events
	return doc, nil
}

func (s *SignDocumentService) createSignedPDF(
	ctx context.Context,
	doc domain.SignDocument,
) (string, error) {
	data, err := s.storage.Download(ctx, doc.SourceFileKey)
	if err != nil {
		return "", err
	}
	recipients, err := s.repo.ListRecipients(ctx, doc.ID)
	if err != nil {
		return "", err
	}
	fields, err := s.repo.ListFields(ctx, doc.ID)
	if err != nil {
		return "", err
	}
	signatures, err := s.repo.ListSignatures(ctx, doc.ID)
	if err != nil {
		return "", err
	}
	stamped, err := s.pdfStamper.StampSignDocumentPDF(
		ctx,
		domain.SignDocumentPDFStampInput{
			SourcePDF:  data,
			Recipients: recipients,
			Fields:     fields,
			Signatures: signatures,
		},
	)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf(
		"sign-documents/%s/signed-%s%s",
		doc.ID.String(),
		time.Now().Format("20060102150405"),
		filepath.Ext(doc.SourceFileKey),
	)
	if filepath.Ext(key) == "" {
		key += ".pdf"
	}
	uploadedKey, _, err := s.storage.Upload(
		ctx,
		nopMultipartFile{Reader: bytes.NewReader(stamped)},
		key,
		"application/pdf",
	)
	return uploadedKey, err
}

func hasRequiredFieldValues(
	fields []domain.SignDocumentField,
	values []domain.SignDocumentFieldValueParams,
) bool {
	provided := map[uuid.UUID]bool{}
	for _, value := range values {
		if strings.TrimSpace(value.Value) != "" {
			provided[value.FieldID] = true
		}
	}
	for _, field := range fields {
		if field.Required && !provided[field.ID] {
			return false
		}
	}
	return true
}
func signatureHash(params domain.SignDocumentSignParams, employeeID, recipientID uuid.UUID) string {
	h := sha256.New()
	_, _ = h.Write(
		[]byte(
			employeeID.String() + recipientID.String() + params.DocumentID.String() + params.ConsentText,
		),
	)
	if params.SignatureText != nil {
		_, _ = h.Write([]byte(*params.SignatureText))
	}
	if params.SignatureImageFileKey != nil {
		_, _ = h.Write([]byte(*params.SignatureImageFileKey))
	}
	return hex.EncodeToString(h.Sum(nil))
}
func normalizeLimit(limit int32) int32 {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

type nopMultipartFile struct{ *bytes.Reader }

func (f nopMultipartFile) Close() error { return nil }

var _ multipart.File = nopMultipartFile{}
var _ io.Reader = nopMultipartFile{}
