package service

import (
	"context"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
)

type SettingsService struct {
	repository domain.SettingsRepository
	logger     domain.Logger
}

func NewSettingsService(
	repository domain.SettingsRepository,
	logger domain.Logger,
) domain.SettingsService {
	return &SettingsService{
		repository: repository,
		logger:     logger,
	}
}

func (s *SettingsService) GetOrganizationProfile(
	ctx context.Context,
) (*domain.OrganizationProfile, error) {
	profile, err := s.repository.GetOrganizationProfile(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"SettingsService.GetOrganizationProfile",
				"failed to get organization profile",
				err,
				zap.String("resource", "app_organization_profile"),
			)
		}
		return nil, err
	}

	return profile, nil
}

func (s *SettingsService) UpdateOrganizationProfile(
	ctx context.Context,
	params domain.UpdateOrganizationProfileParams,
) (*domain.OrganizationProfile, error) {
	profile, err := s.repository.UpdateOrganizationProfile(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"SettingsService.UpdateOrganizationProfile",
				"failed to update organization profile",
				err,
				zap.String("resource", "app_organization_profile"),
			)
		}
		return nil, err
	}

	return profile, nil
}
