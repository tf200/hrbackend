package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrQualificationTypeNotFound = errors.New("qualification type not found")

type QualificationType struct {
	ID        uuid.UUID
	Code      string
	Name      string
	AppContext *string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateQualificationTypeParams struct {
	Code string
	Name string
}

type QualificationTypeRepository interface {
	ListQualificationTypes(ctx context.Context) ([]QualificationType, error)
	CreateQualificationType(ctx context.Context, params CreateQualificationTypeParams) (*QualificationType, error)
	UpdateQualificationType(ctx context.Context, id uuid.UUID, params CreateQualificationTypeParams) (*QualificationType, error)
}

type QualificationTypeService interface {
	ListQualificationTypes(ctx context.Context) ([]QualificationType, error)
	CreateQualificationType(ctx context.Context, params CreateQualificationTypeParams) (*QualificationType, error)
	UpdateQualificationType(ctx context.Context, id uuid.UUID, params CreateQualificationTypeParams) (*QualificationType, error)
}
