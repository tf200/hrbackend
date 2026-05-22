package seed

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SalaryTableSeed struct {
	CAOCode              string
	Name                 string
	EffectiveFrom        time.Time
	EffectiveTo          *time.Time
	FullTimeHoursPerWeek float64
	FullTimeHoursPerYear float64
	SourceURL            *string
	Steps                []SalaryScaleStepSeed
}

type SalaryScaleStepSeed struct {
	Scale         int
	Step          string
	IPNumber      *int
	MonthlySalary float64
}

type SalaryTablesSeeder struct {
	Tables []SalaryTableSeed
}

func (s SalaryTablesSeeder) Name() string {
	return "salary_tables"
}

func (s SalaryTablesSeeder) Seed(ctx context.Context, env Env) error {
	for _, table := range s.Tables {
		if strings.TrimSpace(table.CAOCode) == "" {
			return fmt.Errorf("seed salary_tables: cao code is required")
		}
		if strings.TrimSpace(table.Name) == "" {
			return fmt.Errorf("seed salary_tables[%s]: name is required", table.CAOCode)
		}
		if table.FullTimeHoursPerWeek <= 0 || table.FullTimeHoursPerYear <= 0 {
			return fmt.Errorf("seed salary_tables[%s]: full-time hours must be positive", table.CAOCode)
		}

		var tableID uuid.UUID
		err := env.DB.QueryRow(ctx, `
			INSERT INTO cao_salary_tables (
				cao_code,
				name,
				effective_from,
				effective_to,
				full_time_hours_per_week,
				full_time_hours_per_year,
				source_url
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (cao_code, effective_from) DO UPDATE
			SET
				name = EXCLUDED.name,
				effective_to = EXCLUDED.effective_to,
				full_time_hours_per_week = EXCLUDED.full_time_hours_per_week,
				full_time_hours_per_year = EXCLUDED.full_time_hours_per_year,
				source_url = EXCLUDED.source_url,
				updated_at = CURRENT_TIMESTAMP
			RETURNING id
		`, table.CAOCode, table.Name, table.EffectiveFrom, table.EffectiveTo,
			table.FullTimeHoursPerWeek, table.FullTimeHoursPerYear, table.SourceURL).Scan(&tableID)
		if err != nil {
			return fmt.Errorf("seed salary_tables[%s %s]: %w", table.CAOCode, table.EffectiveFrom.Format("2006-01-02"), err)
		}

		for _, step := range table.Steps {
			if step.Scale <= 0 {
				return fmt.Errorf("seed salary_tables[%s]: scale must be positive", table.CAOCode)
			}
			if strings.TrimSpace(step.Step) == "" {
				return fmt.Errorf("seed salary_tables[%s]: step is required for scale %d", table.CAOCode, step.Scale)
			}
			if step.MonthlySalary <= 0 {
				return fmt.Errorf("seed salary_tables[%s]: monthly salary must be positive for scale %d step %s", table.CAOCode, step.Scale, step.Step)
			}

			_, err := env.DB.Exec(ctx, `
				INSERT INTO cao_salary_scale_steps (
					salary_table_id,
					scale,
					step,
					ip_number,
					monthly_salary,
					hourly_rate
				)
				VALUES ($1, $2, $3, $4, $5, ROUND(($5::numeric * 12 / $6::numeric), 4))
				ON CONFLICT (salary_table_id, scale, step) DO UPDATE
				SET
					ip_number = EXCLUDED.ip_number,
					monthly_salary = EXCLUDED.monthly_salary,
					hourly_rate = EXCLUDED.hourly_rate,
					updated_at = CURRENT_TIMESTAMP
			`, tableID, step.Scale, step.Step, step.IPNumber, step.MonthlySalary, table.FullTimeHoursPerYear)
			if err != nil {
				return fmt.Errorf("seed salary_tables[%s scale %d step %s]: %w", table.CAOCode, step.Scale, step.Step, err)
			}
		}
	}

	return nil
}
