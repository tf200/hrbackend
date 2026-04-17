package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type SettingsRepository struct {
	queries db.Querier
}

func NewSettingsRepository(queries db.Querier) domain.SettingsRepository {
	return &SettingsRepository{queries: queries}
}

func (r *SettingsRepository) GetOrganizationProfile(
	ctx context.Context,
) (*domain.OrganizationProfile, error) {
	profile, err := r.queries.GetAppOrganizationProfile(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.OrganizationProfile{
		Name:                  profile.Name,
		DefaultTimezone:       profile.DefaultTimezone,
		Email:                 profile.Email,
		PhoneNumber:           profile.PhoneNumber,
		Website:               profile.Website,
		HQStreet:              profile.HqStreet,
		HQHouseNumber:         profile.HqHouseNumber,
		HQHouseNumberAddition: profile.HqHouseNumberAddition,
		HQPostalCode:          profile.HqPostalCode,
		HQCity:                profile.HqCity,
		CreatedAt:             conv.TimeFromPgTimestamptz(profile.CreatedAt),
		UpdatedAt:             conv.TimeFromPgTimestamptz(profile.UpdatedAt),
	}, nil
}

func (r *SettingsRepository) UpdateOrganizationProfile(
	ctx context.Context,
	params domain.UpdateOrganizationProfileParams,
) (*domain.OrganizationProfile, error) {
	profile, err := r.queries.UpdateAppOrganizationProfile(
		ctx,
		db.UpdateAppOrganizationProfileParams{
			Name:                  params.Name,
			DefaultTimezone:       params.DefaultTimezone,
			Email:                 params.Email,
			PhoneNumber:           params.PhoneNumber,
			Website:               params.Website,
			HqStreet:              params.HQStreet,
			HqHouseNumber:         params.HQHouseNumber,
			HqHouseNumberAddition: params.HQHouseNumberAddition,
			HqPostalCode:          params.HQPostalCode,
			HqCity:                params.HQCity,
		},
	)
	if err != nil {
		return nil, err
	}

	return &domain.OrganizationProfile{
		Name:                  profile.Name,
		DefaultTimezone:       profile.DefaultTimezone,
		Email:                 profile.Email,
		PhoneNumber:           profile.PhoneNumber,
		Website:               profile.Website,
		HQStreet:              profile.HqStreet,
		HQHouseNumber:         profile.HqHouseNumber,
		HQHouseNumberAddition: profile.HqHouseNumberAddition,
		HQPostalCode:          profile.HqPostalCode,
		HQCity:                profile.HqCity,
		CreatedAt:             conv.TimeFromPgTimestamptz(profile.CreatedAt),
		UpdatedAt:             conv.TimeFromPgTimestamptz(profile.UpdatedAt),
	}, nil
}
