package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	UUID    uuid.UUID
	Name    string
	File    string
	Size    int32
	IsUsed  bool
	Tag     *string
	Updated time.Time
	Created time.Time
}

type CreateAttachmentParams struct {
	Name   string
	File   string
	Size   int32
	IsUsed bool
	Tag    *string
}

type UpdateAttachmentUsedParams struct {
	UUID   uuid.UUID
	IsUsed bool
}

type AttachmentRepository interface {
	CreateAttachment(ctx context.Context, params CreateAttachmentParams) (*Attachment, error)
	GetAttachment(ctx context.Context, id uuid.UUID) (*Attachment, error)
	UpdateAttachmentUsed(
		ctx context.Context,
		params UpdateAttachmentUsedParams,
	) (*Attachment, error)
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
}

type AttachmentService interface {
	RequestUploadURL(
		ctx context.Context,
		filename string,
		size int64,
		tag *string,
	) (*UploadURLResponse, error)
	GetAttachment(ctx context.Context, id uuid.UUID) (*Attachment, error)
}

type UploadURLResponse struct {
	AttachmentID uuid.UUID
	UploadURL    string
	FileKey      string
}
