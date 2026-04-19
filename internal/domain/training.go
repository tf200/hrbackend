package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTrainingInvalidRequest = errors.New("invalid training request")
	ErrTrainingNotFound       = errors.New("training not found")
)

type TrainingCatalogItem struct {
	ID                       uuid.UUID
	Title                    string
	Description              *string
	Category                 *string
	EstimatedDurationMinutes *int32
	IsActive                 bool
	CreatedByEmployeeID      *uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CreateTrainingCatalogItemParams struct {
	Title                    string
	Description              *string
	Category                 *string
	EstimatedDurationMinutes *int32
	CreatedByEmployeeID      *uuid.UUID
}

type TrainingRepository interface {
	CreateTrainingCatalogItem(
		ctx context.Context,
		params CreateTrainingCatalogItemParams,
	) (*TrainingCatalogItem, error)
}

type TrainingService interface {
	CreateTrainingCatalogItem(
		ctx context.Context,
		params CreateTrainingCatalogItemParams,
	) (*TrainingCatalogItem, error)
}
