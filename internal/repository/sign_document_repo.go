package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type SignDocumentRepository struct {
	store *db.Store
}

func NewSignDocumentRepository(store *db.Store) domain.SignDocumentRepository {
	return &SignDocumentRepository{store: store}
}

func (r *SignDocumentRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.SignDocumentRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&SignDocumentRepository{store: &db.Store{Queries: q, ConnPool: r.store.ConnPool}})
	})
}

func (r *SignDocumentRepository) CreateDocument(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	attachment *domain.Attachment,
	params domain.CreateSignDocumentParams,
) (*domain.SignDocument, error) {
	row, err := r.store.CreateSignDocument(ctx, db.CreateSignDocumentParams{
		Title:               params.Title,
		SourceAttachmentID:  params.SourceAttachmentID,
		SourceFileKey:       attachment.File,
		CreatedByEmployeeID: actorEmployeeID,
		RelatedEntityType:   params.RelatedEntityType,
		RelatedEntityID:     params.RelatedEntityID,
		ExpiresAt:           pgTimestamptzFromPtr(params.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) CreateRecipient(
	ctx context.Context,
	documentID uuid.UUID,
	params domain.CreateSignDocumentRecipientParams,
) (*domain.SignDocumentRecipient, error) {
	row, err := r.store.CreateSignDocumentRecipient(
		ctx,
		db.CreateSignDocumentRecipientParams{
			DocumentID:   documentID,
			EmployeeID:   params.EmployeeID,
			SigningOrder: params.SigningOrder,
		},
	)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocumentRecipient(row)
	return &model, nil
}

func (r *SignDocumentRepository) GetDocumentByID(
	ctx context.Context,
	documentID uuid.UUID,
) (*domain.SignDocument, error) {
	row, err := r.store.GetSignDocumentByID(ctx, documentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSignDocumentNotFound
		}
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) ListDocumentsByCreator(
	ctx context.Context,
	employeeID uuid.UUID,
	limit, offset int32,
) ([]domain.SignDocument, error) {
	rows, err := r.store.ListSignDocumentsByCreator(
		ctx,
		db.ListSignDocumentsByCreatorParams{
			CreatedByEmployeeID: employeeID,
			Limit:               limit,
			Offset:              offset,
		},
	)
	return toDomainSignDocuments(rows), err
}

func (r *SignDocumentRepository) ListDocumentsForEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	limit, offset int32,
) ([]domain.SignDocument, error) {
	rows, err := r.store.ListSignDocumentsForEmployee(
		ctx,
		db.ListSignDocumentsForEmployeeParams{EmployeeID: employeeID, Limit: limit, Offset: offset},
	)
	return toDomainSignDocuments(rows), err
}

func (r *SignDocumentRepository) ListRecipients(
	ctx context.Context,
	documentID uuid.UUID,
) ([]domain.SignDocumentRecipient, error) {
	rows, err := r.store.ListSignDocumentRecipients(ctx, documentID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.SignDocumentRecipient, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSignDocumentRecipient(row))
	}
	return items, nil
}

func (r *SignDocumentRepository) GetRecipientForEmployee(
	ctx context.Context,
	documentID, employeeID uuid.UUID,
) (*domain.SignDocumentRecipient, error) {
	row, err := r.store.GetSignDocumentRecipientForEmployee(
		ctx,
		db.GetSignDocumentRecipientForEmployeeParams{
			DocumentID: documentID,
			EmployeeID: employeeID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSignDocumentRecipientNotFound
		}
		return nil, err
	}
	model := toDomainSignDocumentRecipient(row)
	return &model, nil
}

func (r *SignDocumentRepository) ReplaceFields(
	ctx context.Context,
	documentID uuid.UUID,
	fields []domain.UpsertSignDocumentFieldParams,
) ([]domain.SignDocumentField, error) {
	if err := r.store.DeleteSignDocumentFields(ctx, documentID); err != nil {
		return nil, err
	}
	items := make([]domain.SignDocumentField, 0, len(fields))
	for _, f := range fields {
		row, err := r.store.CreateSignDocumentField(
			ctx,
			db.CreateSignDocumentFieldParams{
				DocumentID:  documentID,
				RecipientID: f.RecipientID,
				Type:        db.SignDocumentFieldTypeEnum(f.Type),
				PageNumber:  f.PageNumber,
				X:           f.X,
				Y:           f.Y,
				Width:       f.Width,
				Height:      f.Height,
				Required:    f.Required,
				Label:       f.Label,
			},
		)
		if err != nil {
			return nil, err
		}
		items = append(items, toDomainSignDocumentField(row))
	}
	return items, nil
}

func (r *SignDocumentRepository) ListFields(
	ctx context.Context,
	documentID uuid.UUID,
) ([]domain.SignDocumentField, error) {
	rows, err := r.store.ListSignDocumentFields(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return toDomainSignDocumentFields(rows), nil
}

func (r *SignDocumentRepository) ListFieldsForRecipient(
	ctx context.Context,
	documentID, recipientID uuid.UUID,
) ([]domain.SignDocumentField, error) {
	rows, err := r.store.ListSignDocumentFieldsForRecipient(
		ctx,
		db.ListSignDocumentFieldsForRecipientParams{
			DocumentID:  documentID,
			RecipientID: recipientID,
		},
	)
	if err != nil {
		return nil, err
	}
	return toDomainSignDocumentFields(rows), nil
}

func (r *SignDocumentRepository) SendDocument(
	ctx context.Context,
	documentID uuid.UUID,
) (*domain.SignDocument, error) {
	row, err := r.store.SendSignDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) MarkRecipientViewed(
	ctx context.Context,
	recipientID uuid.UUID,
) (*domain.SignDocumentRecipient, error) {
	row, err := r.store.MarkSignDocumentRecipientViewed(ctx, recipientID)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocumentRecipient(row)
	return &model, nil
}

func (r *SignDocumentRepository) CountUnsignedPriorRecipients(
	ctx context.Context,
	recipientID uuid.UUID,
) (int32, error) {
	return r.store.CountUnsignedPriorSignDocumentRecipients(ctx, recipientID)
}

func (r *SignDocumentRepository) CreateSignature(
	ctx context.Context,
	params domain.CreateSignDocumentSignatureParams,
) (*domain.SignDocumentSignature, error) {
	row, err := r.store.CreateSignDocumentSignature(
		ctx,
		db.CreateSignDocumentSignatureParams{
			DocumentID:            params.DocumentID,
			RecipientID:           params.RecipientID,
			EmployeeID:            params.EmployeeID,
			SignatureProfileID:    params.SignatureProfileID,
			SignatureText:         params.SignatureText,
			SignatureImageFileKey: params.SignatureImageFileKey,
			ConsentText:           params.ConsentText,
			IpAddress:             params.IPAddress,
			UserAgent:             params.UserAgent,
			SignatureHash:         params.SignatureHash,
		},
	)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocumentSignature(row)
	return &model, nil
}

func (r *SignDocumentRepository) UpdateFieldValue(
	ctx context.Context,
	fieldID, recipientID uuid.UUID,
	value string,
) (*domain.SignDocumentField, error) {
	row, err := r.store.UpdateSignDocumentFieldValue(
		ctx,
		db.UpdateSignDocumentFieldValueParams{ID: fieldID, RecipientID: recipientID, Value: &value},
	)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocumentField(row)
	return &model, nil
}

func (r *SignDocumentRepository) MarkRecipientSigned(
	ctx context.Context,
	recipientID uuid.UUID,
) (*domain.SignDocumentRecipient, error) {
	row, err := r.store.MarkSignDocumentRecipientSigned(ctx, recipientID)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocumentRecipient(row)
	return &model, nil
}

func (r *SignDocumentRepository) CountUnsignedRecipients(
	ctx context.Context,
	documentID uuid.UUID,
) (int32, error) {
	return r.store.CountUnsignedSignDocumentRecipients(ctx, documentID)
}

func (r *SignDocumentRepository) MarkDocumentPartiallySigned(
	ctx context.Context,
	documentID uuid.UUID,
) (*domain.SignDocument, error) {
	row, err := r.store.MarkSignDocumentPartiallySigned(ctx, documentID)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) MarkDocumentCompleted(
	ctx context.Context,
	documentID uuid.UUID,
	signedFileKey string,
) (*domain.SignDocument, error) {
	row, err := r.store.MarkSignDocumentCompleted(
		ctx,
		db.MarkSignDocumentCompletedParams{ID: documentID, SignedFileKey: &signedFileKey},
	)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) CancelDocument(
	ctx context.Context,
	documentID uuid.UUID,
) (*domain.SignDocument, error) {
	row, err := r.store.CancelSignDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	model := toDomainSignDocument(row)
	return &model, nil
}

func (r *SignDocumentRepository) CreateEvent(
	ctx context.Context,
	event domain.SignDocumentEvent,
) error {
	_, err := r.store.CreateSignDocumentEvent(
		ctx,
		db.CreateSignDocumentEventParams{
			DocumentID:      event.DocumentID,
			RecipientID:     event.RecipientID,
			ActorEmployeeID: event.ActorEmployeeID,
			Event:           db.SignDocumentEventEnum(event.Event),
			IpAddress:       event.IPAddress,
			UserAgent:       event.UserAgent,
			Column7:         event.Metadata,
		},
	)
	return err
}

func (r *SignDocumentRepository) ListEvents(
	ctx context.Context,
	documentID uuid.UUID,
) ([]domain.SignDocumentEvent, error) {
	rows, err := r.store.ListSignDocumentEvents(ctx, documentID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.SignDocumentEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSignDocumentEvent(row))
	}
	return items, nil
}

func (r *SignDocumentRepository) ListSignatures(
	ctx context.Context,
	documentID uuid.UUID,
) ([]domain.SignDocumentSignature, error) {
	rows, err := r.store.ListSignDocumentSignatures(ctx, documentID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.SignDocumentSignature, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSignDocumentSignature(row))
	}
	return items, nil
}

func toDomainSignDocuments(rows []db.SignDocument) []domain.SignDocument {
	items := make([]domain.SignDocument, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSignDocument(row))
	}
	return items
}
func toDomainSignDocument(row db.SignDocument) domain.SignDocument {
	return domain.SignDocument{
		ID:                  row.ID,
		Title:               row.Title,
		SourceAttachmentID:  row.SourceAttachmentID,
		SourceFileKey:       row.SourceFileKey,
		SignedFileKey:       row.SignedFileKey,
		Status:              string(row.Status),
		CreatedByEmployeeID: row.CreatedByEmployeeID,
		RelatedEntityType:   row.RelatedEntityType,
		RelatedEntityID:     row.RelatedEntityID,
		ExpiresAt:           signDocumentTimePtrFromPgTimestamptz(row.ExpiresAt),
		SentAt:              signDocumentTimePtrFromPgTimestamptz(row.SentAt),
		CompletedAt:         signDocumentTimePtrFromPgTimestamptz(row.CompletedAt),
		CancelledAt:         signDocumentTimePtrFromPgTimestamptz(row.CancelledAt),
		CreatedAt:           conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:           conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
func toDomainSignDocumentRecipient(row db.SignDocumentRecipient) domain.SignDocumentRecipient {
	return domain.SignDocumentRecipient{
		ID:            row.ID,
		DocumentID:    row.DocumentID,
		EmployeeID:    row.EmployeeID,
		Name:          row.Name,
		Email:         row.Email,
		SigningOrder:  row.SigningOrder,
		Status:        string(row.Status),
		ViewedAt:      signDocumentTimePtrFromPgTimestamptz(row.ViewedAt),
		SignedAt:      signDocumentTimePtrFromPgTimestamptz(row.SignedAt),
		DeclinedAt:    signDocumentTimePtrFromPgTimestamptz(row.DeclinedAt),
		DeclineReason: row.DeclineReason,
		CreatedAt:     conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:     conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
func toDomainSignDocumentFields(rows []db.SignDocumentField) []domain.SignDocumentField {
	items := make([]domain.SignDocumentField, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSignDocumentField(row))
	}
	return items
}
func toDomainSignDocumentField(row db.SignDocumentField) domain.SignDocumentField {
	return domain.SignDocumentField{
		ID:          row.ID,
		DocumentID:  row.DocumentID,
		RecipientID: row.RecipientID,
		Type:        string(row.Type),
		PageNumber:  row.PageNumber,
		X:           row.X,
		Y:           row.Y,
		Width:       row.Width,
		Height:      row.Height,
		Required:    row.Required,
		Label:       row.Label,
		Value:       row.Value,
		CreatedAt:   conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:   conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainEmployeeSignatureProfile(
	row db.EmployeeSignatureProfile,
) domain.EmployeeSignatureProfile {
	return domain.EmployeeSignatureProfile{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		Type:         string(row.Type),
		TypedName:    row.TypedName,
		ImageFileKey: row.ImageFileKey,
		IsDefault:    true,
		CreatedAt:    conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:    conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
func toDomainSignDocumentSignature(row db.SignDocumentSignature) domain.SignDocumentSignature {
	return domain.SignDocumentSignature{
		ID:                    row.ID,
		DocumentID:            row.DocumentID,
		RecipientID:           row.RecipientID,
		EmployeeID:            row.EmployeeID,
		SignatureProfileID:    row.SignatureProfileID,
		SignatureText:         row.SignatureText,
		SignatureImageFileKey: row.SignatureImageFileKey,
		ConsentText:           row.ConsentText,
		IPAddress:             row.IpAddress,
		UserAgent:             row.UserAgent,
		SignatureHash:         row.SignatureHash,
		SignedAt:              conv.TimeFromPgTimestamptz(row.SignedAt),
	}
}
func toDomainSignDocumentEvent(row db.SignDocumentEvent) domain.SignDocumentEvent {
	return domain.SignDocumentEvent{
		ID:              row.ID,
		DocumentID:      row.DocumentID,
		RecipientID:     row.RecipientID,
		ActorEmployeeID: row.ActorEmployeeID,
		Event:           string(row.Event),
		IPAddress:       row.IpAddress,
		UserAgent:       row.UserAgent,
		Metadata:        row.Metadata,
		CreatedAt:       conv.TimeFromPgTimestamptz(row.CreatedAt),
	}
}

func signDocumentTimePtrFromPgTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
func pgTimestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}
