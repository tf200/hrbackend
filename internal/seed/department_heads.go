package seed

import (
	"context"
	"fmt"
	"strings"
)

type DepartmentHeadSeed struct {
	DepartmentAlias string
	EmployeeAlias   string
}

type DepartmentHeadsSeeder struct {
	Assignments []DepartmentHeadSeed
}

func (s DepartmentHeadsSeeder) Name() string {
	return "department_heads"
}

func (s DepartmentHeadsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Assignments) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed department_heads: state is required")
	}

	for _, item := range s.Assignments {
		departmentAlias := strings.TrimSpace(item.DepartmentAlias)
		employeeAlias := strings.TrimSpace(item.EmployeeAlias)
		if departmentAlias == "" || employeeAlias == "" {
			return fmt.Errorf("seed department_heads: department alias and employee alias are required")
		}

		departmentID, ok := env.State.DepartmentID(departmentAlias)
		if !ok {
			return fmt.Errorf(
				"seed department_heads: missing department alias %q in seed state",
				departmentAlias,
			)
		}

		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf(
				"seed department_heads: missing employee alias %q in seed state",
				employeeAlias,
			)
		}

		if _, err := env.DB.Exec(ctx, `
			UPDATE departments
			SET
				department_head_employee_id = $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, employeeID, departmentID); err != nil {
			return fmt.Errorf(
				"seed department_heads[%s->%s]: %w",
				departmentAlias,
				employeeAlias,
				err,
			)
		}
	}

	return nil
}
