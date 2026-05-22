package handler

import (
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type listSalaryScaleStepsRequest struct {
	ActiveOnly bool `form:"active_only"`
}

func toListSalaryScaleStepsParams(req listSalaryScaleStepsRequest) domain.ListSalaryScaleStepsParams {
	return domain.ListSalaryScaleStepsParams{
		ActiveOnly: req.ActiveOnly,
	}
}

type salaryScaleStepsResponse struct {
	Groups []salaryScaleGroupResponse `json:"groups"`
	Meta   *salaryScaleStepsMeta      `json:"meta"`
}

type salaryScaleGroupResponse struct {
	Scale int32                        `json:"scale"`
	Steps []salaryScaleStepOptionResponse `json:"steps"`
}

type salaryScaleStepOptionResponse struct {
	ID            uuid.UUID `json:"id"`
	SalaryTableID uuid.UUID `json:"salary_table_id"`
	Step          string    `json:"step"`
	IPNumber      *int32    `json:"ip_number"`
	MonthlySalary float64   `json:"monthly_salary"`
	HourlyRate    float64   `json:"hourly_rate"`
	Label         string    `json:"label"`
}

type salaryScaleStepsMeta struct {
	SalaryTableID   uuid.UUID `json:"salary_table_id"`
	CaoCode         string    `json:"cao_code"`
	SalaryTableName string    `json:"salary_table_name"`
	EffectiveFrom   time.Time `json:"effective_from"`
	EffectiveTo     *time.Time `json:"effective_to"`
	ScaleCount      int       `json:"scale_count"`
}

func toSalaryScaleStepsResponse(result *domain.SalaryScaleStepsResult) salaryScaleStepsResponse {
	groups := make([]salaryScaleGroupResponse, len(result.Groups))
	for i, g := range result.Groups {
		steps := make([]salaryScaleStepOptionResponse, len(g.Steps))
		for j, s := range g.Steps {
			steps[j] = salaryScaleStepOptionResponse{
				ID:            s.ID,
				SalaryTableID: s.SalaryTableID,
				Step:          s.Step,
				IPNumber:      s.IPNumber,
				MonthlySalary: s.MonthlySalary,
				HourlyRate:    s.HourlyRate,
				Label:         s.Label,
			}
		}
		groups[i] = salaryScaleGroupResponse{
			Scale: g.Scale,
			Steps: steps,
		}
	}

	var meta *salaryScaleStepsMeta
	if result.Meta != nil {
		meta = &salaryScaleStepsMeta{
			SalaryTableID:   result.Meta.SalaryTableID,
			CaoCode:         result.Meta.CaoCode,
			SalaryTableName: result.Meta.SalaryTableName,
			EffectiveFrom:   result.Meta.EffectiveFrom,
			EffectiveTo:     result.Meta.EffectiveTo,
			ScaleCount:      result.Meta.ScaleCount,
		}
	}

	return salaryScaleStepsResponse{
		Groups: groups,
		Meta:   meta,
	}
}
