package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type TrainingRepository struct {
	queries db.Querier
}

func NewTrainingRepository(queries db.Querier) domain.TrainingRepository {
	return &TrainingRepository{queries: queries}
}

func (r *TrainingRepository) CreateTrainingCatalogItem(
	ctx context.Context,
	params domain.CreateTrainingCatalogItemParams,
) (*domain.TrainingCatalogItem, error) {
	item, err := r.queries.CreateTrainingCatalogItem(ctx, db.CreateTrainingCatalogItemParams{
		Title:                    params.Title,
		Description:              params.Description,
		Category:                 params.Category,
		EstimatedDurationMinutes: params.EstimatedDurationMinutes,
		CreatedByEmployeeID:      params.CreatedByEmployeeID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainTrainingCatalogItem(item), nil
}

func toDomainTrainingCatalogItem(item db.TrainingCatalogItem) *domain.TrainingCatalogItem {
	return &domain.TrainingCatalogItem{
		ID:                       item.ID,
		Title:                    item.Title,
		Description:              item.Description,
		Category:                 item.Category,
		EstimatedDurationMinutes: item.EstimatedDurationMinutes,
		IsActive:                 item.IsActive,
		CreatedByEmployeeID:      item.CreatedByEmployeeID,
		CreatedAt:                conv.TimeFromPgTimestamptz(item.CreatedAt),
		UpdatedAt:                conv.TimeFromPgTimestamptz(item.UpdatedAt),
	}
}

var _ domain.TrainingRepository = (*TrainingRepository)(nil)
