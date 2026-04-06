package seed

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	dbrepo "hrbackend/internal/repository/db"
	"hrbackend/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EmployeeHandbookAssignmentSeed struct {
	EmployeeAlias      string
	TemplateAlias      string
	ActorEmployeeAlias *string
}

type EmployeeHandbookAssignmentsSeeder struct {
	Assignments []EmployeeHandbookAssignmentSeed
}

func (s EmployeeHandbookAssignmentsSeeder) Name() string {
	return "employee_handbook_assignments"
}

func (s EmployeeHandbookAssignmentsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Assignments) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed employee_handbook_assignments: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed employee_handbook_assignments: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	handbookRepo := repository.NewHandbookRepository(store)
	handbookService := service.NewHandbookService(handbookRepo, nil)

	for _, item := range s.Assignments {
		if strings.TrimSpace(item.EmployeeAlias) == "" {
			return fmt.Errorf("seed employee_handbook_assignments: employee alias is required")
		}
		if strings.TrimSpace(item.TemplateAlias) == "" {
			return fmt.Errorf("seed employee_handbook_assignments[%s]: template alias is required", item.EmployeeAlias)
		}

		employeeID, ok := env.State.EmployeeID(strings.TrimSpace(item.EmployeeAlias))
		if !ok {
			return fmt.Errorf(
				"seed employee_handbook_assignments[%s]: missing employee alias in seed state",
				item.EmployeeAlias,
			)
		}

		templateID, ok := env.State.HandbookID(strings.TrimSpace(item.TemplateAlias))
		if !ok {
			return fmt.Errorf(
				"seed employee_handbook_assignments[%s]: missing template alias %q in seed state",
				item.EmployeeAlias,
				item.TemplateAlias,
			)
		}

		actorEmployeeID, err := resolveOptionalAssignmentActor(env, item)
		if err != nil {
			return fmt.Errorf("seed employee_handbook_assignments[%s]: %w", item.EmployeeAlias, err)
		}

		active, err := handbookRepo.GetActiveEmployeeHandbookByEmployeeID(ctx, employeeID)
		if err == nil && active.TemplateID == templateID {
			continue
		}
		if err != nil && err != domain.ErrActiveHandbookNotFound {
			return fmt.Errorf("lookup active handbook: %w", err)
		}

		if _, err := handbookService.AssignTemplateToEmployee(ctx, actorEmployeeID, domain.AssignTemplateToEmployeeParams{
			EmployeeID: employeeID,
			TemplateID: templateID,
		}); err != nil {
			return fmt.Errorf("assign template to employee: %w", err)
		}
	}

	return nil
}

func resolveOptionalAssignmentActor(env Env, item EmployeeHandbookAssignmentSeed) (uuid.UUID, error) {
	if item.ActorEmployeeAlias == nil || strings.TrimSpace(*item.ActorEmployeeAlias) == "" {
		return uuid.Nil, nil
	}

	employeeID, ok := env.State.EmployeeID(strings.TrimSpace(*item.ActorEmployeeAlias))
	if !ok {
		return uuid.Nil, fmt.Errorf("missing actor employee alias %q in seed state", strings.TrimSpace(*item.ActorEmployeeAlias))
	}
	return employeeID, nil
}
