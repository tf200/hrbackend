package service

import (
	"context"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
)

type SalaryService struct {
	repo   domain.SalaryRepository
	logger domain.Logger
}

func NewSalaryService(repo domain.SalaryRepository, logger domain.Logger) domain.SalaryService {
	return &SalaryService{repo: repo, logger: logger}
}

func (s *SalaryService) ListSalaryScaleSteps(
	ctx context.Context,
	params domain.ListSalaryScaleStepsParams,
) (*domain.SalaryScaleStepsResult, error) {
	result, err := s.repo.ListSalaryScaleSteps(ctx, params)
	if err != nil {
		s.logError(ctx, "ListSalaryScaleSteps", "failed to list salary scale steps", err)
		return nil, err
	}
	return result, nil
}

func (s *SalaryService) logError(
	ctx context.Context,
	method, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger != nil {
		s.logger.LogError(ctx, "SalaryService."+method, message, err, fields...)
	}
}
