package service

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QualificationService struct {
	repo   domain.QualificationTypeRepository
	logger domain.Logger
}

func NewQualificationService(
	repo domain.QualificationTypeRepository,
	logger domain.Logger,
) domain.QualificationTypeService {
	return &QualificationService{repo: repo, logger: logger}
}

func (s *QualificationService) ListQualificationTypes(
	ctx context.Context,
) ([]domain.QualificationType, error) {
	items, err := s.repo.ListQualificationTypes(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"QualificationService.ListQualificationTypes",
				"failed to list qualification types",
				err,
				zap.Error(err),
			)
		}
		return nil, err
	}
	return items, nil
}

func (s *QualificationService) CreateQualificationType(
	ctx context.Context,
	params domain.CreateQualificationTypeParams,
) (*domain.QualificationType, error) {
	if strings.TrimSpace(params.Name) == "" {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"QualificationService.CreateQualificationType",
				"name is required",
				nil,
			)
		}
		return nil, fmt.Errorf("name is required")
	}

	params.Code = generateCodeFromName(params.Name)

	result, err := s.repo.CreateQualificationType(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"QualificationService.CreateQualificationType",
				"failed to create qualification type",
				err,
				zap.Error(err),
			)
		}
		return nil, err
	}
	return result, nil
}

func (s *QualificationService) UpdateQualificationType(
	ctx context.Context,
	id uuid.UUID,
	params domain.CreateQualificationTypeParams,
) (*domain.QualificationType, error) {
	if strings.TrimSpace(params.Name) == "" {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"QualificationService.UpdateQualificationType",
				"name is required",
				nil,
			)
		}
		return nil, fmt.Errorf("name is required")
	}

	params.Code = generateCodeFromName(params.Name)

	result, err := s.repo.UpdateQualificationType(ctx, id, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"QualificationService.UpdateQualificationType",
				"failed to update qualification type",
				err,
				zap.Error(err),
			)
		}
		return nil, err
	}
	return result, nil
}

func generateCodeFromName(name string) string {
	lower := strings.ToLower(name)
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		"'", "",
		".", "",
		",", "",
		"(", "",
		")", "",
		"/", "_",
	)
	return replacer.Replace(lower)
}
