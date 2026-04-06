package seed

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	dbrepo "hrbackend/internal/repository/db"
	"hrbackend/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LateArrivalSeed struct {
	Alias                  string
	EmployeeAlias          string
	CreatedByEmployeeAlias *string
	ArrivalDate            time.Time
	ArrivalTime            string
	Reason                 string
}

type LateArrivalsSeeder struct {
	Arrivals []LateArrivalSeed
}

func (s LateArrivalsSeeder) Name() string {
	return "late_arrivals"
}

func (s LateArrivalsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Arrivals) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed late_arrivals: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed late_arrivals: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	lateArrivalRepo := repository.NewLateArrivalRepository(store)
	lateArrivalService := service.NewLateArrivalService(lateArrivalRepo, nil)

	arrivalsByEmployee := make(map[string][]LateArrivalSeed)
	for _, item := range s.Arrivals {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed late_arrivals: alias is required")
		}
		employeeAlias := strings.TrimSpace(item.EmployeeAlias)
		if employeeAlias == "" {
			return fmt.Errorf("seed late_arrivals[%s]: employee alias is required", alias)
		}
		if item.ArrivalDate.IsZero() {
			return fmt.Errorf("seed late_arrivals[%s]: arrival date is required", alias)
		}
		if strings.TrimSpace(item.ArrivalTime) == "" {
			return fmt.Errorf("seed late_arrivals[%s]: arrival time is required", alias)
		}
		if strings.TrimSpace(item.Reason) == "" {
			return fmt.Errorf("seed late_arrivals[%s]: reason is required", alias)
		}
		arrivalsByEmployee[employeeAlias] = append(arrivalsByEmployee[employeeAlias], item)
	}

	for employeeAlias, items := range arrivalsByEmployee {
		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed late_arrivals[%s]: employee alias missing in seed state", employeeAlias)
		}

		existing, err := listAllLateArrivalsForEmployee(ctx, lateArrivalService, employeeID)
		if err != nil {
			return fmt.Errorf("seed late_arrivals[%s]: list existing arrivals: %w", employeeAlias, err)
		}

		for _, item := range items {
			if exactLateArrivalExists(existing, item) {
				continue
			}
			if current := findComparableLateArrival(existing, item); current != nil {
				return fmt.Errorf(
					"seed late_arrivals[%s]: existing late arrival already present with different details for %s",
					item.Alias,
					lateArrivalDateOnlyUTC(item.ArrivalDate).Format("2006-01-02"),
				)
			}

			createdByAlias := normalizeOptionalAlias(item.CreatedByEmployeeAlias)
			params := domain.LateArrivalCreateParams{
				EmployeeID:          employeeID,
				CreatedByEmployeeID: employeeID,
				ArrivalDate:         lateArrivalDateOnlyUTC(item.ArrivalDate),
				ArrivalTime:         strings.TrimSpace(item.ArrivalTime),
				Reason:              strings.TrimSpace(item.Reason),
			}
			if createdByAlias == "" || createdByAlias == employeeAlias {
				if _, err := lateArrivalService.CreateLateArrival(ctx, params); err != nil {
					return fmt.Errorf("seed late_arrivals[%s]: create: %w", item.Alias, err)
				}
				continue
			}

			createdByEmployeeID, err := resolveRequiredEmployeeAliasForLateArrival(env, item.CreatedByEmployeeAlias, item.Alias, "created_by")
			if err != nil {
				return err
			}
			params.CreatedByEmployeeID = createdByEmployeeID
			if _, err := lateArrivalService.CreateLateArrivalByAdmin(ctx, params); err != nil {
				return fmt.Errorf("seed late_arrivals[%s]: create by admin: %w", item.Alias, err)
			}
		}
	}

	return nil
}

func listAllLateArrivalsForEmployee(
	ctx context.Context,
	lateArrivalService domain.LateArrivalService,
	employeeID uuid.UUID,
) ([]domain.LateArrivalListItem, error) {
	page, err := lateArrivalService.ListMyLateArrivals(ctx, domain.ListMyLateArrivalsParams{
		EmployeeID: employeeID,
		Limit:      500,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func exactLateArrivalExists(existing []domain.LateArrivalListItem, item LateArrivalSeed) bool {
	for _, current := range existing {
		if sameLateArrival(current.LateArrival, item, true) {
			return true
		}
	}
	return false
}

func findComparableLateArrival(existing []domain.LateArrivalListItem, item LateArrivalSeed) *domain.LateArrival {
	for _, current := range existing {
		if sameLateArrival(current.LateArrival, item, false) {
			copy := current.LateArrival
			return &copy
		}
	}
	return nil
}

func sameLateArrival(current domain.LateArrival, item LateArrivalSeed, includeReason bool) bool {
	if !lateArrivalDateOnlyUTC(current.ArrivalDate).Equal(lateArrivalDateOnlyUTC(item.ArrivalDate)) {
		return false
	}
	if strings.TrimSpace(current.ArrivalTime) != normalizeLateArrivalTime(item.ArrivalTime) {
		return false
	}
	if includeReason && strings.TrimSpace(current.Reason) != strings.TrimSpace(item.Reason) {
		return false
	}
	return true
}

func normalizeLateArrivalTime(value string) string {
	trimmed := strings.TrimSpace(value)
	if parsed, err := time.Parse("15:04:05", trimmed); err == nil {
		return parsed.Format("15:04:05")
	}
	if parsed, err := time.Parse("15:04", trimmed); err == nil {
		return parsed.Format("15:04:05")
	}
	return trimmed
}

func lateArrivalDateOnlyUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func resolveRequiredEmployeeAliasForLateArrival(
	env Env,
	alias *string,
	arrivalAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed late_arrivals[%s]: %s employee alias is required", arrivalAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed late_arrivals[%s]: %s employee alias %q missing in seed state",
			arrivalAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}
