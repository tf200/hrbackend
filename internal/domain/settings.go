package domain

import (
	"context"
	"time"
)

type OrganizationProfile struct {
	Name                  string
	DefaultTimezone       string
	Email                 *string
	PhoneNumber           *string
	Website               *string
	HQStreet              *string
	HQHouseNumber         *string
	HQHouseNumberAddition *string
	HQPostalCode          *string
	HQCity                *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type UpdateOrganizationProfileParams struct {
	Name                  *string
	DefaultTimezone       *string
	Email                 *string
	PhoneNumber           *string
	Website               *string
	HQStreet              *string
	HQHouseNumber         *string
	HQHouseNumberAddition *string
	HQPostalCode          *string
	HQCity                *string
}

type SettingsRepository interface {
	GetOrganizationProfile(ctx context.Context) (*OrganizationProfile, error)
	UpdateOrganizationProfile(
		ctx context.Context,
		params UpdateOrganizationProfileParams,
	) (*OrganizationProfile, error)
}

type SettingsService interface {
	GetOrganizationProfile(ctx context.Context) (*OrganizationProfile, error)
	UpdateOrganizationProfile(
		ctx context.Context,
		params UpdateOrganizationProfileParams,
	) (*OrganizationProfile, error)
}
