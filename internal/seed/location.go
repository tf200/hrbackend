package seed

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type LocationSeed struct {
	Alias               string
	OrganizationAlias   string
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	Timezone            string
	LocationType        string
}

type LocationSeeder struct {
	Locations []LocationSeed
}

func (s LocationSeeder) Name() string {
	return "location"
}

func (s LocationSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Locations) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed location: state is required")
	}

	for _, item := range s.Locations {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed location: alias is required")
		}
		if strings.TrimSpace(item.OrganizationAlias) == "" {
			return fmt.Errorf("seed location: organization alias is required for location %q", item.Alias)
		}
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("seed location: name is required for location %q", item.Alias)
		}

		organizationID, ok := env.State.OrganizationID(item.OrganizationAlias)
		if !ok {
			return fmt.Errorf(
				"seed location[%s]: missing organization alias %q in seed state",
				item.Alias,
				item.OrganizationAlias,
			)
		}

		var id uuid.UUID
		err := env.DB.QueryRow(ctx, `
			WITH existing AS (
				SELECT id
				FROM location
				WHERE organisation_id = $1
				  AND name = $2
				LIMIT 1
			), updated AS (
				UPDATE location
				SET
					street = $3,
					house_number = $4,
					house_number_addition = $5,
					postal_code = $6,
					city = $7,
					timezone = $8,
					location_type = $9::location_type_enum,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = (SELECT id FROM existing)
				RETURNING id
			), inserted AS (
				INSERT INTO location (
					organisation_id,
					name,
					street,
					house_number,
					house_number_addition,
					postal_code,
					city,
					timezone,
					location_type
				)
				SELECT
					$1, $2, $3, $4, $5, $6, $7, $8, $9::location_type_enum
				WHERE NOT EXISTS (SELECT 1 FROM existing)
				RETURNING id
			)
			SELECT id FROM updated
			UNION ALL
			SELECT id FROM inserted
			LIMIT 1
		`, organizationID, item.Name, item.Street, item.HouseNumber, item.HouseNumberAddition,
			item.PostalCode, item.City, item.Timezone, item.LocationType).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed location[%s]: %w", item.Alias, err)
		}

		env.State.PutLocation(item.Alias, id)
	}

	return nil
}
