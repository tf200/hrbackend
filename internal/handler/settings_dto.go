package handler

import (
	"time"

	"hrbackend/internal/domain"
)

type getOrganizationProfileResponse struct {
	Name                  string    `json:"name"`
	DefaultTimezone       string    `json:"default_timezone"`
	Email                 *string   `json:"email"`
	PhoneNumber           *string   `json:"phone_number"`
	Website               *string   `json:"website"`
	HQStreet              *string   `json:"hq_street"`
	HQHouseNumber         *string   `json:"hq_house_number"`
	HQHouseNumberAddition *string   `json:"hq_house_number_addition"`
	HQPostalCode          *string   `json:"hq_postal_code"`
	HQCity                *string   `json:"hq_city"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type updateOrganizationProfileRequest struct {
	Name                  *string `json:"name"`
	DefaultTimezone       *string `json:"default_timezone"`
	Email                 *string `json:"email"`
	PhoneNumber           *string `json:"phone_number"`
	Website               *string `json:"website"`
	HQStreet              *string `json:"hq_street"`
	HQHouseNumber         *string `json:"hq_house_number"`
	HQHouseNumberAddition *string `json:"hq_house_number_addition"`
	HQPostalCode          *string `json:"hq_postal_code"`
	HQCity                *string `json:"hq_city"`
}

type updateOrganizationProfileResponse = getOrganizationProfileResponse

func toGetOrganizationProfileResponse(
	profile *domain.OrganizationProfile,
) getOrganizationProfileResponse {
	return getOrganizationProfileResponse{
		Name:                  profile.Name,
		DefaultTimezone:       profile.DefaultTimezone,
		Email:                 profile.Email,
		PhoneNumber:           profile.PhoneNumber,
		Website:               profile.Website,
		HQStreet:              profile.HQStreet,
		HQHouseNumber:         profile.HQHouseNumber,
		HQHouseNumberAddition: profile.HQHouseNumberAddition,
		HQPostalCode:          profile.HQPostalCode,
		HQCity:                profile.HQCity,
		CreatedAt:             profile.CreatedAt,
		UpdatedAt:             profile.UpdatedAt,
	}
}

func toUpdateOrganizationProfileParams(
	req updateOrganizationProfileRequest,
) domain.UpdateOrganizationProfileParams {
	return domain.UpdateOrganizationProfileParams{
		Name:                  req.Name,
		DefaultTimezone:       req.DefaultTimezone,
		Email:                 req.Email,
		PhoneNumber:           req.PhoneNumber,
		Website:               req.Website,
		HQStreet:              req.HQStreet,
		HQHouseNumber:         req.HQHouseNumber,
		HQHouseNumberAddition: req.HQHouseNumberAddition,
		HQPostalCode:          req.HQPostalCode,
		HQCity:                req.HQCity,
	}
}
