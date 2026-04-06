package seed

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type DepartmentSeed struct {
	Alias       string
	Name        string
	Description *string
}

type DepartmentsSeeder struct {
	Departments []DepartmentSeed
}

func (s DepartmentsSeeder) Name() string {
	return "departments"
}

func (s DepartmentsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Departments) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed departments: state is required")
	}

	for _, item := range s.Departments {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed departments: alias is required")
		}
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("seed departments: name is required for alias %q", item.Alias)
		}

		var id uuid.UUID
		err := env.DB.QueryRow(ctx, `
			INSERT INTO departments (
				name,
				description,
				department_head_employee_id
			) VALUES (
				$1, $2, $3
			)
			ON CONFLICT (name) DO UPDATE
			SET
				description = EXCLUDED.description,
				updated_at = CURRENT_TIMESTAMP
			RETURNING id
		`, item.Name, item.Description, nil).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed departments[%s]: %w", item.Alias, err)
		}

		env.State.PutDepartment(item.Alias, id)
	}

	return nil
}
