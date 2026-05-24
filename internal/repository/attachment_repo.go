package repository

import (
	"context"

	"github.com/google/uuid"
	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type AttachmentRepository struct {
	queries db.Querier
}

func NewAttachmentRepository(queries db.Querier) domain.AttachmentRepository {
	return &AttachmentRepository{queries: queries}
}

func (r *AttachmentRepository) CreateAttachment(
	ctx context.Context,
	params domain.CreateAttachmentParams,
) (*domain.Attachment, error) {
	row, err := r.queries.CreateAttachment(ctx, db.CreateAttachmentParams{
		Name:   params.Name,
		File:   params.File,
		Size:   params.Size,
		IsUsed: params.IsUsed,
		Tag:    params.Tag,
	})
	if err != nil {
		return nil, err
	}

	return toDomainAttachment(row), nil
}

func (r *AttachmentRepository) GetAttachment(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Attachment, error) {
	row, err := r.queries.GetAttachment(ctx, id)
	if err != nil {
		return nil, err
	}

	return toDomainAttachment(row), nil
}

func (r *AttachmentRepository) UpdateAttachmentUsed(
	ctx context.Context,
	params domain.UpdateAttachmentUsedParams,
) (*domain.Attachment, error) {
	row, err := r.queries.UpdateAttachmentUsed(ctx, db.UpdateAttachmentUsedParams{
		Uuid:   params.UUID,
		IsUsed: params.IsUsed,
	})
	if err != nil {
		return nil, err
	}

	return toDomainAttachment(row), nil
}

func (r *AttachmentRepository) DeleteAttachment(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.queries.DeleteAttachment(ctx, id)
}

func toDomainAttachment(row db.AttachmentFile) *domain.Attachment {
	return &domain.Attachment{
		UUID:    row.Uuid,
		Name:    row.Name,
		File:    row.File,
		Size:    row.Size,
		IsUsed:  row.IsUsed,
		Tag:     row.Tag,
		Updated: conv.TimeFromPgTimestamptz(row.Updated),
		Created: conv.TimeFromPgTimestamptz(row.Created),
	}
}
