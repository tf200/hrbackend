package seed

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type OrganizationSeed struct {
	Alias               string
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	PhoneNumber         *string
	Email               *string
	KvkNumber           *string
	BtwNumber           *string
}

type OrganisationsSeeder struct {
	Organizations []OrganizationSeed
}

func (s OrganisationsSeeder) Name() string {
	return "organisations"
}

func (s OrganisationsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Organizations) == 0 {
		return nil
	}

	for _, item := range s.Organizations {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed organisations: alias is required")
		}
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("seed organisations: name is required for alias %q", item.Alias)
		}

		var id uuid.UUID
		err := env.DB.QueryRow(ctx, `
			WITH existing AS (
				SELECT id
				FROM organisations
				WHERE name = $1
				LIMIT 1
			), updated AS (
				UPDATE organisations
				SET
					street = $2,
					house_number = $3,
					house_number_addition = $4,
					postal_code = $5,
					city = $6,
					phone_number = $7,
					email = $8,
					kvk_number = $9,
					btw_number = $10,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = (SELECT id FROM existing)
				RETURNING id
			), inserted AS (
				INSERT INTO organisations (
					name,
					street,
					house_number,
					house_number_addition,
					postal_code,
					city,
					phone_number,
					email,
					kvk_number,
					btw_number
				)
				SELECT
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				WHERE NOT EXISTS (SELECT 1 FROM existing)
				RETURNING id
			)
			SELECT id FROM updated
			UNION ALL
			SELECT id FROM inserted
			LIMIT 1
		`, item.Name, item.Street, item.HouseNumber, item.HouseNumberAddition, item.PostalCode,
			item.City, item.PhoneNumber, item.Email, item.KvkNumber, item.BtwNumber).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed organisations[%s]: %w", item.Alias, err)
		}

		if env.State != nil {
			env.State.PutOrganization(item.Alias, id)
		}
	}

	return nil
}
