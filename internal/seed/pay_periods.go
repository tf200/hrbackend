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

type PayPeriodSeed struct {
	Alias                  string
	EmployeeAlias          string
	CreatedByEmployeeAlias string
	PaidByEmployeeAlias    *string
	Status                 string
	PeriodStart            time.Time
	PeriodEnd              time.Time
}

type PayPeriodsSeeder struct {
	Periods []PayPeriodSeed
}

func (s PayPeriodsSeeder) Name() string {
	return "pay_periods"
}

func (s PayPeriodsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Periods) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed pay_periods: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed pay_periods: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	payoutRepo := repository.NewPayoutRepository(store)
	payoutService := service.NewPayoutService(payoutRepo, nil)

	for _, item := range s.Periods {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed pay_periods: alias is required")
		}
		employeeAlias := strings.TrimSpace(item.EmployeeAlias)
		if employeeAlias == "" {
			return fmt.Errorf("seed pay_periods[%s]: employee alias is required", alias)
		}
		if strings.TrimSpace(item.CreatedByEmployeeAlias) == "" {
			return fmt.Errorf("seed pay_periods[%s]: created-by employee alias is required", alias)
		}
		if !isValidSeedPayPeriodStatus(item.Status) {
			return fmt.Errorf("seed pay_periods[%s]: unsupported status %q", alias, item.Status)
		}
		if item.PeriodStart.IsZero() || item.PeriodEnd.IsZero() {
			return fmt.Errorf("seed pay_periods[%s]: period start and end are required", alias)
		}
		if dateOnlyUTC(item.PeriodEnd).Before(dateOnlyUTC(item.PeriodStart)) {
			return fmt.Errorf("seed pay_periods[%s]: period end must be on or after period start", alias)
		}
		if strings.TrimSpace(item.Status) == domain.PayPeriodStatusPaid && normalizeOptionalAlias(item.PaidByEmployeeAlias) == "" {
			return fmt.Errorf("seed pay_periods[%s]: paid-by employee alias is required for paid periods", alias)
		}

		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed pay_periods[%s]: employee alias %q missing in seed state", alias, employeeAlias)
		}
		creatorID, ok := env.State.EmployeeID(strings.TrimSpace(item.CreatedByEmployeeAlias))
		if !ok {
			return fmt.Errorf(
				"seed pay_periods[%s]: created-by employee alias %q missing in seed state",
				alias,
				item.CreatedByEmployeeAlias,
			)
		}

		existing, err := findExistingPayPeriod(ctx, payoutService, employeeID, item.PeriodStart, item.PeriodEnd)
		if err != nil {
			return fmt.Errorf("seed pay_periods[%s]: inspect existing period: %w", alias, err)
		}
		if existing == nil {
			created, err := payoutService.ClosePayPeriod(ctx, creatorID, domain.ClosePayPeriodParams{
				EmployeeID:  employeeID,
				PeriodStart: dateOnlyUTC(item.PeriodStart),
				PeriodEnd:   dateOnlyUTC(item.PeriodEnd),
			})
			if err != nil {
				return fmt.Errorf("seed pay_periods[%s]: close pay period: %w", alias, err)
			}
			existing = created
		}

		if existing.Status == strings.TrimSpace(item.Status) {
			env.State.PutPayPeriod(alias, existing.ID)
			continue
		}

		if existing.Status == domain.PayPeriodStatusDraft && strings.TrimSpace(item.Status) == domain.PayPeriodStatusPaid {
			paidByEmployeeID, err := resolveRequiredEmployeeAliasForPayPeriod(env, item.PaidByEmployeeAlias, alias, "paid_by")
			if err != nil {
				return err
			}
			updated, err := payoutService.MarkPayPeriodPaidByAdmin(ctx, paidByEmployeeID, existing.ID)
			if err != nil {
				return fmt.Errorf("seed pay_periods[%s]: mark paid: %w", alias, err)
			}
			env.State.PutPayPeriod(alias, updated.ID)
			continue
		}

		return fmt.Errorf(
			"seed pay_periods[%s]: existing pay period has status %q; refusing to mutate to %q",
			alias,
			existing.Status,
			item.Status,
		)
	}

	return nil
}

func findExistingPayPeriod(
	ctx context.Context,
	payoutService domain.PayoutService,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) (*domain.PayPeriod, error) {
	page, err := payoutService.ListPayPeriods(ctx, domain.ListPayPeriodsParams{
		Limit:  500,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	start := dateOnlyUTC(periodStart)
	end := dateOnlyUTC(periodEnd)
	for _, item := range page.Items {
		if item.EmployeeID != employeeID {
			continue
		}
		if dateOnlyUTC(item.PeriodStart).Equal(start) && dateOnlyUTC(item.PeriodEnd).Equal(end) {
			copy := item
			return &copy, nil
		}
	}
	return nil, nil
}

func resolveRequiredEmployeeAliasForPayPeriod(
	env Env,
	alias *string,
	periodAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed pay_periods[%s]: %s employee alias is required", periodAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed pay_periods[%s]: %s employee alias %q missing in seed state",
			periodAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}

func isValidSeedPayPeriodStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case domain.PayPeriodStatusDraft, domain.PayPeriodStatusPaid:
		return true
	default:
		return false
	}
}
