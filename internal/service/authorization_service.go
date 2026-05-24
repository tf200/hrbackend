package service

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthorizationService struct {
	repo   domain.AuthorizationRepository
	logger domain.Logger
}

func NewAuthorizationService(
	repo domain.AuthorizationRepository,
	logger domain.Logger,
) domain.AuthorizationService {
	return &AuthorizationService{repo: repo, logger: logger}
}

func (s *AuthorizationService) ListAuthorizations(
	ctx context.Context,
) ([]domain.Authorization, error) {
	items, err := s.repo.ListAuthorizations(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.ListAuthorizations", "failed to list authorizations", err, zap.Error(err))
		}
		return nil, err
	}
	return items, nil
}

func (s *AuthorizationService) CreateAuthorization(
	ctx context.Context,
	params domain.CreateAuthorizationParams,
) (*domain.Authorization, error) {
	if strings.TrimSpace(params.Name) == "" {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.CreateAuthorization", "name is required", nil)
		}
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(params.Category) == "" {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.CreateAuthorization", "category is required", nil)
		}
		return nil, fmt.Errorf("category is required")
	}

	result, err := s.repo.CreateAuthorization(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.CreateAuthorization", "failed to create authorization", err, zap.Error(err))
		}
		return nil, err
	}
	return result, nil
}

func (s *AuthorizationService) UpdateAuthorization(
	ctx context.Context,
	id uuid.UUID,
	params domain.CreateAuthorizationParams,
) (*domain.Authorization, error) {
	if strings.TrimSpace(params.Name) == "" {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.UpdateAuthorization", "name is required", nil)
		}
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(params.Category) == "" {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.UpdateAuthorization", "category is required", nil)
		}
		return nil, fmt.Errorf("category is required")
	}

	result, err := s.repo.UpdateAuthorization(ctx, id, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "AuthorizationService.UpdateAuthorization", "failed to update authorization", err, zap.Error(err))
		}
		return nil, err
	}
	return result, nil
}
