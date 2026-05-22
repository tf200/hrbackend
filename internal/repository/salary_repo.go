package repository

import (
	"context"
	"fmt"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type SalaryRepository struct {
	store *db.Store
}

func NewSalaryRepository(store *db.Store) domain.SalaryRepository {
	return &SalaryRepository{store: store}
}

func (r *SalaryRepository) ListSalaryScaleSteps(
	ctx context.Context,
	params domain.ListSalaryScaleStepsParams,
) (*domain.SalaryScaleStepsResult, error) {
	rows, err := r.store.ListSalaryScaleSteps(ctx, &params.ActiveOnly)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &domain.SalaryScaleStepsResult{}, nil
	}

	var groups []domain.SalaryScaleGroup
	var currentGroup *domain.SalaryScaleGroup

	for _, row := range rows {
		if currentGroup == nil || currentGroup.Scale != row.Scale {
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}
			currentGroup = &domain.SalaryScaleGroup{
				Scale: row.Scale,
				Steps: nil,
			}
		}

		label := fmt.Sprintf("Scale %d / Step %s - €%.2f/mo", row.Scale, row.Step, row.MonthlySalary)

		step := domain.SalaryScaleStepOption{
			ID:            row.ID,
			SalaryTableID: row.SalaryTableID,
			Step:          row.Step,
			IPNumber:      row.IpNumber,
			MonthlySalary: row.MonthlySalary,
			HourlyRate:    row.HourlyRate,
			Label:         label,
		}

		currentGroup.Steps = append(currentGroup.Steps, step)
	}

	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}

	first := rows[0]
	meta := &domain.SalaryScaleStepsMeta{
		SalaryTableID:   first.SalaryTableID,
		CaoCode:         first.CaoCode,
		SalaryTableName: first.SalaryTableName,
		EffectiveFrom:   conv.TimeFromPgDate(first.EffectiveFrom),
		EffectiveTo:     conv.TimePtrFromPgDate(first.EffectiveTo),
		ScaleCount:      len(groups),
	}

	return &domain.SalaryScaleStepsResult{
		Groups: groups,
		Meta:   meta,
	}, nil
}
