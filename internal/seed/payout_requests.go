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

type PayoutRequestSeed struct {
	Alias                          string
	EmployeeAlias                  string
	BalanceAdjustedByEmployeeAlias string
	RequestedHours                 int32
	BalanceYear                    int32
	Status                         string
	RequestNote                    *string
	DecisionByEmployeeAlias        *string
	PaidByEmployeeAlias            *string
	SalaryMonth                    *time.Time
	DecisionNote                   *string
}

type PayoutRequestsSeeder struct {
	Requests []PayoutRequestSeed
}

func (s PayoutRequestsSeeder) Name() string {
	return "payout_requests"
}

func (s PayoutRequestsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Requests) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed payout_requests: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed payout_requests: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	leaveRepo := repository.NewLeaveRepository(store)
	leaveService := service.NewLeaveService(leaveRepo, nil)
	payoutRepo := repository.NewPayoutRepository(store)
	payoutService := service.NewPayoutService(payoutRepo, nil)

	requestsByEmployee := make(map[string][]PayoutRequestSeed)
	for _, item := range s.Requests {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed payout_requests: alias is required")
		}
		employeeAlias := strings.TrimSpace(item.EmployeeAlias)
		if employeeAlias == "" {
			return fmt.Errorf("seed payout_requests[%s]: employee alias is required", alias)
		}
		if strings.TrimSpace(item.BalanceAdjustedByEmployeeAlias) == "" {
			return fmt.Errorf("seed payout_requests[%s]: balance adjustment actor alias is required", alias)
		}
		if item.RequestedHours <= 0 {
			return fmt.Errorf("seed payout_requests[%s]: requested hours must be positive", alias)
		}
		if item.BalanceYear < 2000 || item.BalanceYear > 2100 {
			return fmt.Errorf("seed payout_requests[%s]: balance year is invalid", alias)
		}
		if !isValidSeedPayoutStatus(item.Status) {
			return fmt.Errorf("seed payout_requests[%s]: unsupported status %q", alias, item.Status)
		}
		if wantsPayoutApproval(item.Status) && (item.SalaryMonth == nil || item.SalaryMonth.IsZero()) {
			return fmt.Errorf("seed payout_requests[%s]: salary month is required for approved/paid requests", alias)
		}

		requestsByEmployee[employeeAlias] = append(requestsByEmployee[employeeAlias], item)
	}

	for employeeAlias, items := range requestsByEmployee {
		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed payout_requests[%s]: employee alias missing in seed state", employeeAlias)
		}

		existing, err := listAllPayoutRequestsForEmployee(ctx, payoutService, employeeID)
		if err != nil {
			return fmt.Errorf("seed payout_requests[%s]: list existing requests: %w", employeeAlias, err)
		}

		for _, item := range items {
			if exactPayoutRequestExists(existing, item) {
				continue
			}

			current := findComparablePayoutRequest(existing, item)
			if current == nil {
				if err := ensureExtraLeaveHoursForPayout(ctx, leaveService, env, employeeID, item); err != nil {
					return fmt.Errorf("seed payout_requests[%s]: ensure extra leave hours: %w", item.Alias, err)
				}

				created, err := payoutService.CreatePayoutRequest(ctx, employeeID, domain.CreatePayoutRequestParams{
					RequestedHours: item.RequestedHours,
					BalanceYear:    item.BalanceYear,
					RequestNote:    normalizeOptionalText(item.RequestNote),
				})
				if err != nil {
					return fmt.Errorf("seed payout_requests[%s]: create: %w", item.Alias, err)
				}

				existing = append(existing, *created)
				current = created
			}

			if current.Status == item.Status {
				continue
			}

			updated, err := advancePayoutRequestToDesiredState(ctx, payoutService, env, current, item)
			if err != nil {
				return fmt.Errorf("seed payout_requests[%s]: %w", item.Alias, err)
			}
			replacePayoutRequest(existing, *updated)
		}
	}

	return nil
}

func ensureExtraLeaveHoursForPayout(
	ctx context.Context,
	leaveService domain.LeaveService,
	env Env,
	employeeID uuid.UUID,
	item PayoutRequestSeed,
) error {
	page, err := leaveService.ListMyLeaveBalances(ctx, domain.ListMyLeaveBalancesParams{
		EmployeeID: employeeID,
		Year:       int32Ptr(item.BalanceYear),
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		return err
	}

	var extraRemaining int32
	if len(page.Items) > 0 {
		extraRemaining = page.Items[0].ExtraRemaining
	}
	if extraRemaining >= item.RequestedHours {
		return nil
	}

	actorID, ok := env.State.EmployeeID(strings.TrimSpace(item.BalanceAdjustedByEmployeeAlias))
	if !ok {
		return fmt.Errorf("balance adjustment actor alias %q missing in seed state", item.BalanceAdjustedByEmployeeAlias)
	}

	_, err = leaveService.AdjustLeaveBalance(ctx, domain.AdjustLeaveBalanceParams{
		AdminEmployeeID: actorID,
		EmployeeID:      employeeID,
		Year:            item.BalanceYear,
		ExtraHoursDelta: item.RequestedHours - extraRemaining,
		Reason:          fmt.Sprintf("seed payout request %s extra-hour top-up", strings.TrimSpace(item.Alias)),
	})
	return err
}

func listAllPayoutRequestsForEmployee(
	ctx context.Context,
	payoutService domain.PayoutService,
	employeeID uuid.UUID,
) ([]domain.PayoutRequest, error) {
	page, err := payoutService.ListMyPayoutRequests(ctx, domain.ListMyPayoutRequestsParams{
		EmployeeID: employeeID,
		Limit:      500,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func exactPayoutRequestExists(existing []domain.PayoutRequest, item PayoutRequestSeed) bool {
	for _, current := range existing {
		if samePayoutRequest(current, item, true) {
			return true
		}
	}
	return false
}

func findComparablePayoutRequest(existing []domain.PayoutRequest, item PayoutRequestSeed) *domain.PayoutRequest {
	for _, current := range existing {
		if samePayoutRequest(current, item, false) {
			copy := current
			return &copy
		}
	}
	return nil
}

func samePayoutRequest(current domain.PayoutRequest, item PayoutRequestSeed, includeStatus bool) bool {
	if current.RequestedHours != item.RequestedHours || current.BalanceYear != item.BalanceYear {
		return false
	}
	if normalizeOptionalText(current.RequestNote) == nil && normalizeOptionalText(item.RequestNote) != nil {
		return false
	}
	if normalizeOptionalText(current.RequestNote) != nil && normalizeOptionalText(item.RequestNote) == nil {
		return false
	}
	if normalizeOptionalText(current.RequestNote) != nil &&
		*normalizeOptionalText(current.RequestNote) != *normalizeOptionalText(item.RequestNote) {
		return false
	}
	if includeStatus {
		if current.Status != strings.TrimSpace(item.Status) {
			return false
		}
		if !sameOptionalDate(current.SalaryMonth, item.SalaryMonth) {
			return false
		}
		if normalizeOptionalText(current.DecisionNote) == nil && normalizeOptionalText(item.DecisionNote) != nil {
			return false
		}
		if normalizeOptionalText(current.DecisionNote) != nil && normalizeOptionalText(item.DecisionNote) == nil {
			return false
		}
		if normalizeOptionalText(current.DecisionNote) != nil &&
			*normalizeOptionalText(current.DecisionNote) != *normalizeOptionalText(item.DecisionNote) {
			return false
		}
	}
	return true
}

func advancePayoutRequestToDesiredState(
	ctx context.Context,
	payoutService domain.PayoutService,
	env Env,
	current *domain.PayoutRequest,
	item PayoutRequestSeed,
) (*domain.PayoutRequest, error) {
	desiredStatus := strings.TrimSpace(item.Status)

	if current.Status == domain.PayoutRequestStatusPending {
		if desiredStatus == domain.PayoutRequestStatusPending {
			return current, nil
		}

		decisionByAlias := item.DecisionByEmployeeAlias
		decisionByEmployeeID, err := resolveRequiredEmployeeAliasForPayout(env, decisionByAlias, item.Alias, "decision")
		if err != nil {
			return nil, err
		}

		updated, err := payoutService.DecidePayoutRequestByAdmin(ctx, decisionByEmployeeID, current.ID, domain.DecidePayoutRequestParams{
			Decision:     payoutDecisionFromStatus(desiredStatus),
			DecisionNote: normalizeOptionalText(item.DecisionNote),
			SalaryMonth:  normalizeOptionalTime(item.SalaryMonth),
		})
		if err != nil {
			return nil, err
		}
		current = updated
	}

	if current.Status == domain.PayoutRequestStatusApproved && desiredStatus == domain.PayoutRequestStatusPaid {
		paidByEmployeeID, err := resolveRequiredEmployeeAliasForPayout(env, item.PaidByEmployeeAlias, item.Alias, "paid_by")
		if err != nil {
			return nil, err
		}
		return payoutService.MarkPayoutRequestPaidByAdmin(ctx, paidByEmployeeID, current.ID)
	}

	if current.Status != desiredStatus {
		return nil, fmt.Errorf(
			"existing request has status %q; refusing to mutate to %q",
			current.Status,
			desiredStatus,
		)
	}

	return current, nil
}

func replacePayoutRequest(existing []domain.PayoutRequest, updated domain.PayoutRequest) {
	for idx := range existing {
		if existing[idx].ID == updated.ID {
			existing[idx] = updated
			return
		}
	}
}

func resolveRequiredEmployeeAliasForPayout(
	env Env,
	alias *string,
	requestAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed payout_requests[%s]: %s employee alias is required", requestAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed payout_requests[%s]: %s employee alias %q missing in seed state",
			requestAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}

func isValidSeedPayoutStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case domain.PayoutRequestStatusPending,
		domain.PayoutRequestStatusApproved,
		domain.PayoutRequestStatusRejected,
		domain.PayoutRequestStatusPaid:
		return true
	default:
		return false
	}
}

func wantsPayoutApproval(status string) bool {
	switch strings.TrimSpace(status) {
	case domain.PayoutRequestStatusApproved, domain.PayoutRequestStatusPaid:
		return true
	default:
		return false
	}
}

func payoutDecisionFromStatus(status string) string {
	switch strings.TrimSpace(status) {
	case domain.PayoutRequestStatusApproved, domain.PayoutRequestStatusPaid:
		return "approve"
	case domain.PayoutRequestStatusRejected:
		return "reject"
	default:
		return strings.TrimSpace(status)
	}
}

func normalizeOptionalTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	normalized := time.Date(value.UTC().Year(), value.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	return &normalized
}

func sameOptionalDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return normalizeOptionalTime(a).Equal(*normalizeOptionalTime(b))
}

func int32Ptr(value int32) *int32 {
	return &value
}
