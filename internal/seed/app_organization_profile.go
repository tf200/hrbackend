package seed

import (
	"context"
	"fmt"
)

type AppOrganizationProfileDefaults struct {
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
}

type AppOrganizationProfileSeeder struct {
	Defaults AppOrganizationProfileDefaults
}

func (s AppOrganizationProfileSeeder) Name() string {
	return "app_organization_profile"
}

func (s AppOrganizationProfileSeeder) Seed(ctx context.Context, env Env) error {
	_, err := env.DB.Exec(ctx, `
		INSERT INTO app_organization_profile (
			singleton,
			name,
			default_timezone,
			email,
			phone_number,
			website,
			hq_street,
			hq_house_number,
			hq_house_number_addition,
			hq_postal_code,
			hq_city
		)
		VALUES (
			TRUE,
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (singleton) DO UPDATE
		SET
			name = EXCLUDED.name,
			default_timezone = EXCLUDED.default_timezone,
			email = EXCLUDED.email,
			phone_number = EXCLUDED.phone_number,
			website = EXCLUDED.website,
			hq_street = EXCLUDED.hq_street,
			hq_house_number = EXCLUDED.hq_house_number,
			hq_house_number_addition = EXCLUDED.hq_house_number_addition,
			hq_postal_code = EXCLUDED.hq_postal_code,
			hq_city = EXCLUDED.hq_city,
			updated_at = CURRENT_TIMESTAMP
	`, s.Defaults.Name, s.Defaults.DefaultTimezone, s.Defaults.Email, s.Defaults.PhoneNumber,
		s.Defaults.Website, s.Defaults.HQStreet, s.Defaults.HQHouseNumber,
		s.Defaults.HQHouseNumberAddition, s.Defaults.HQPostalCode, s.Defaults.HQCity)
	if err != nil {
		return fmt.Errorf("seed app_organization_profile: %w", err)
	}
	return nil
}
