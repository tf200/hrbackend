package service

import (
	"context"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
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

func (s *TrainingService) AssignTrainingToEmployee(
	ctx context.Context,
	params domain.AssignTrainingToEmployeeParams,
) (*domain.EmployeeTrainingAssignment, error) {
	if params.EmployeeID == uuid.Nil || params.TrainingID == uuid.Nil || params.DueAt.IsZero() {
		return nil, domain.ErrTrainingInvalidRequest
	}

	assignment, err := s.repository.AssignTrainingToEmployee(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"TrainingService.AssignTrainingToEmployee",
				"failed to assign training to employee",
				err,
				zap.String("training_id", params.TrainingID.String()),
				zap.String("employee_id", params.EmployeeID.String()),
			)
		}
		return nil, err
	}

	return assignment, nil
}

func (s *TrainingService) CancelTrainingAssignment(
	ctx context.Context,
	params domain.CancelTrainingAssignmentParams,
) (*domain.EmployeeTrainingAssignment, error) {
	if params.AssignmentID == uuid.Nil {
		return nil, domain.ErrTrainingInvalidRequest
	}

	params.CancellationReason = trimStringPtr(params.CancellationReason)

	assignment, err := s.repository.CancelTrainingAssignment(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"TrainingService.CancelTrainingAssignment",
				"failed to cancel training assignment",
				err,
				zap.String("assignment_id", params.AssignmentID.String()),
			)
		}
		return nil, err
	}

	return assignment, nil
}

func (s *TrainingService) ListTrainingAssignments(
	ctx context.Context,
	params domain.ListTrainingAssignmentsParams,
) (*domain.TrainingAssignmentPage, error) {
	params.EmployeeSearch = trimStringPtr(params.EmployeeSearch)

	if params.Status != nil {
		normalized := strings.TrimSpace(strings.ToLower(*params.Status))
		switch normalized {
		case "":
			params.Status = nil
		case "assigned", "in_progress", "completed", "cancelled":
			params.Status = &normalized
		default:
			return nil, domain.ErrTrainingInvalidRequest
		}
	}

	page, err := s.repository.ListTrainingAssignments(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"TrainingService.ListTrainingAssignments",
				"failed to list training assignments",
				err,
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
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
