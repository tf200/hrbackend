package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ListSalaryScaleStepsParams struct {
	ActiveOnly bool
}

type SalaryScaleStepOption struct {
	ID            uuid.UUID
	SalaryTableID uuid.UUID
	Step          string
	IPNumber      *int32
	MonthlySalary float64
	HourlyRate    float64
	Label         string
}

type SalaryScaleGroup struct {
	Scale int32
	Steps []SalaryScaleStepOption
}

type SalaryScaleStepsMeta struct {
	SalaryTableID   uuid.UUID
	CaoCode         string
	SalaryTableName string
	EffectiveFrom   time.Time
	EffectiveTo     *time.Time
	ScaleCount      int
}

type SalaryScaleStepsResult struct {
	Groups []SalaryScaleGroup
	Meta   *SalaryScaleStepsMeta
}

type SalaryRepository interface {
	ListSalaryScaleSteps(ctx context.Context, params ListSalaryScaleStepsParams) (*SalaryScaleStepsResult, error)
}

type SalaryService interface {
	ListSalaryScaleSteps(ctx context.Context, params ListSalaryScaleStepsParams) (*SalaryScaleStepsResult, error)
}
