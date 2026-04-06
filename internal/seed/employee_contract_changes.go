package seed

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	dbrepo "hrbackend/internal/repository/db"
	"hrbackend/internal/service"

	"github.com/jackc/pgx/v5"
)

type EmployeeContractChangeSeed struct {
	EmployeeAlias         string
	ActorEmployeeAlias    string
	EffectiveFrom         time.Time
	ContractHours         float64
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
	ContractEndDate       *time.Time
}

type EmployeeContractChangesSeeder struct {
	Employees []EmployeeSeed
	Changes   []EmployeeContractChangeSeed
}

func (s EmployeeContractChangesSeeder) Name() string {
	return "employee_contract_changes"
}

func (s EmployeeContractChangesSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Changes) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed employee_contract_changes: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed employee_contract_changes: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	employeeRepo := repository.NewEmployeeRepository(store)
	employeeService := service.NewEmployeeService(employeeRepo, nil)

	employeeSeedByAlias := make(map[string]EmployeeSeed, len(s.Employees))
	for _, employee := range s.Employees {
		employeeSeedByAlias[employee.Alias] = employee
	}

	grouped := make(map[string][]EmployeeContractChangeSeed)
	for _, item := range s.Changes {
		if strings.TrimSpace(item.EmployeeAlias) == "" {
			return fmt.Errorf("seed employee_contract_changes: employee alias is required")
		}
		if strings.TrimSpace(item.ActorEmployeeAlias) == "" {
			return fmt.Errorf("seed employee_contract_changes[%s]: actor employee alias is required", item.EmployeeAlias)
		}
		if item.EffectiveFrom.IsZero() {
			return fmt.Errorf("seed employee_contract_changes[%s]: effective_from is required", item.EmployeeAlias)
		}
		grouped[item.EmployeeAlias] = append(grouped[item.EmployeeAlias], item)
	}

	for employeeAlias, desiredChanges := range grouped {
		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed employee_contract_changes[%s]: employee alias missing in seed state", employeeAlias)
		}

		employeeSeed, ok := employeeSeedByAlias[employeeAlias]
		if !ok {
			return fmt.Errorf("seed employee_contract_changes[%s]: baseline employee seed not found", employeeAlias)
		}

		slices.SortFunc(desiredChanges, func(a, b EmployeeContractChangeSeed) int {
			return contractDateOnlyUTC(a.EffectiveFrom).Compare(contractDateOnlyUTC(b.EffectiveFrom))
		})

		existingChanges, err := employeeService.ListContractChanges(ctx, employeeID)
		if err != nil {
			return fmt.Errorf("seed employee_contract_changes[%s]: list existing changes: %w", employeeAlias, err)
		}

		for _, existing := range existingChanges {
			if matchesDesiredChange(existing, desiredChanges) {
				continue
			}
			if matchesBootstrappedBaseline(existing, employeeSeed) {
				continue
			}
			return fmt.Errorf(
				"seed employee_contract_changes[%s]: existing contract history differs from seeded history; refusing to mutate",
				employeeAlias,
			)
		}

		for _, desired := range desiredChanges {
			if desiredChangeAlreadyExists(existingChanges, desired) {
				continue
			}

			actorID, ok := env.State.EmployeeID(strings.TrimSpace(desired.ActorEmployeeAlias))
			if !ok {
				return fmt.Errorf(
					"seed employee_contract_changes[%s]: actor employee alias %q missing in seed state",
					employeeAlias,
					desired.ActorEmployeeAlias,
				)
			}

			if _, err := employeeService.CreateContractChange(ctx, actorID, employeeID, domain.CreateEmployeeContractChangeParams{
				EffectiveFrom:         contractDateOnlyUTC(desired.EffectiveFrom),
				ContractHours:         desired.ContractHours,
				ContractType:          desired.ContractType,
				ContractRate:          desired.ContractRate,
				IrregularHoursProfile: desired.IrregularHoursProfile,
				ContractEndDate:       desired.ContractEndDate,
			}); err != nil {
				return fmt.Errorf("seed employee_contract_changes[%s]: create change: %w", employeeAlias, err)
			}
		}
	}

	return nil
}

func matchesDesiredChange(existing domain.EmployeeContractChange, desired []EmployeeContractChangeSeed) bool {
	for _, item := range desired {
		if matchesDesiredChangeExact(existing, item) {
			return true
		}
	}
	return false
}

func desiredChangeAlreadyExists(existing []domain.EmployeeContractChange, desired EmployeeContractChangeSeed) bool {
	for _, item := range existing {
		if matchesDesiredChangeExact(item, desired) {
			return true
		}
	}
	return false
}

func matchesDesiredChangeExact(existing domain.EmployeeContractChange, desired EmployeeContractChangeSeed) bool {
	return sameDate(existing.EffectiveFrom, desired.EffectiveFrom) &&
		floatEquals(existing.ContractHours, desired.ContractHours) &&
		existing.ContractType == desired.ContractType &&
		floatPtrEquals(existing.ContractRate, desired.ContractRate) &&
		existing.IrregularHoursProfile == desired.IrregularHoursProfile &&
		timePtrEquals(existing.ContractEndDate, desired.ContractEndDate)
}

func matchesBootstrappedBaseline(existing domain.EmployeeContractChange, employee EmployeeSeed) bool {
	if employee.ContractStartDate == nil || employee.ContractHours == nil {
		return false
	}

	return sameDate(existing.EffectiveFrom, *employee.ContractStartDate) &&
		floatEquals(existing.ContractHours, *employee.ContractHours) &&
		existing.ContractType == employee.ContractType &&
		floatPtrEquals(existing.ContractRate, employee.ContractRate) &&
		existing.IrregularHoursProfile == employee.IrregularHoursProfile &&
		timePtrEquals(existing.ContractEndDate, employee.ContractEndDate)
}

func sameDate(a, b time.Time) bool {
	return contractDateOnlyUTC(a).Equal(contractDateOnlyUTC(b))
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func floatPtrEquals(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return floatEquals(*a, *b)
}

func timePtrEquals(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return sameDate(*a, *b)
}

func contractDateOnlyUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
