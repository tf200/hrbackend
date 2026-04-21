package service

import (
	"context"
	"strings"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
)

type TrainingService struct {
	repository domain.TrainingRepository
	logger     domain.Logger
}

func NewTrainingService(
	repository domain.TrainingRepository,
	logger domain.Logger,
) domain.TrainingService {
	return &TrainingService{
		repository: repository,
		logger:     logger,
	}
}

func (s *TrainingService) CreateTrainingCatalogItem(
	ctx context.Context,
	params domain.CreateTrainingCatalogItemParams,
) (*domain.TrainingCatalogItem, error) {
	params.Title = strings.TrimSpace(params.Title)
	params.Description = trimStringPtr(params.Description)
	params.Category = trimStringPtr(params.Category)

	if params.Title == "" {
		return nil, domain.ErrTrainingInvalidRequest
	}
	if params.EstimatedDurationMinutes != nil && *params.EstimatedDurationMinutes <= 0 {
		return nil, domain.ErrTrainingInvalidRequest
	}

	item, err := s.repository.CreateTrainingCatalogItem(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"TrainingService.CreateTrainingCatalogItem",
				"failed to create training catalog item",
				err,
				zap.String("title", params.Title),
			)
		}
		return nil, err
	}

	return item, nil
}

func (s *TrainingService) ListTrainingCatalogItems(
	ctx context.Context,
	params domain.ListTrainingCatalogItemsParams,
) (*domain.TrainingCatalogItemPage, error) {
	params.Search = trimStringPtr(params.Search)

	page, err := s.repository.ListTrainingCatalogItems(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"TrainingService.ListTrainingCatalogItems",
				"failed to list training catalog items",
				err,
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
}

var _ domain.TrainingService = (*TrainingService)(nil)
