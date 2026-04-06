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

type ShiftSwapRequestSeed struct {
	Alias                  string
	RequesterEmployeeAlias string
	RecipientEmployeeAlias string
	RequesterScheduleAlias string
	RecipientScheduleAlias string
	Status                 string
	ExpiresAt              *time.Time
	RecipientResponseNote  *string
	AdminEmployeeAlias     *string
	AdminDecisionNote      *string
}

type ShiftSwapRequestsSeeder struct {
	Requests []ShiftSwapRequestSeed
}

func (s ShiftSwapRequestsSeeder) Name() string {
	return "shift_swap_requests"
}

func (s ShiftSwapRequestsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Requests) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed shift_swap_requests: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed shift_swap_requests: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	scheduleRepo := repository.NewScheduleRepository(store)
	scheduleService := service.NewScheduleService(scheduleRepo, nil, nil)

	requestsByRequester := make(map[string][]ShiftSwapRequestSeed)
	for _, item := range s.Requests {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed shift_swap_requests: alias is required")
		}
		requesterAlias := strings.TrimSpace(item.RequesterEmployeeAlias)
		if requesterAlias == "" {
			return fmt.Errorf("seed shift_swap_requests[%s]: requester employee alias is required", alias)
		}
		if strings.TrimSpace(item.RecipientEmployeeAlias) == "" {
			return fmt.Errorf("seed shift_swap_requests[%s]: recipient employee alias is required", alias)
		}
		if strings.TrimSpace(item.RequesterScheduleAlias) == "" {
			return fmt.Errorf("seed shift_swap_requests[%s]: requester schedule alias is required", alias)
		}
		if strings.TrimSpace(item.RecipientScheduleAlias) == "" {
			return fmt.Errorf("seed shift_swap_requests[%s]: recipient schedule alias is required", alias)
		}
		if !isValidSeedShiftSwapStatus(item.Status) {
			return fmt.Errorf("seed shift_swap_requests[%s]: unsupported status %q", alias, item.Status)
		}
		if wantsShiftSwapAdminDecision(item.Status) && normalizeOptionalAlias(item.AdminEmployeeAlias) == "" {
			return fmt.Errorf("seed shift_swap_requests[%s]: admin employee alias is required for status %q", alias, item.Status)
		}
		requestsByRequester[requesterAlias] = append(requestsByRequester[requesterAlias], item)
	}

	for requesterAlias, items := range requestsByRequester {
		requesterEmployeeID, ok := env.State.EmployeeID(requesterAlias)
		if !ok {
			return fmt.Errorf(
				"seed shift_swap_requests[%s]: requester employee alias missing in seed state",
				requesterAlias,
			)
		}

		existing, err := scheduleService.ListMyShiftSwapRequests(ctx, requesterEmployeeID)
		if err != nil {
			return fmt.Errorf("seed shift_swap_requests[%s]: list existing requests: %w", requesterAlias, err)
		}

		for _, item := range items {
			requesterScheduleID, ok := env.State.ScheduleID(strings.TrimSpace(item.RequesterScheduleAlias))
			if !ok {
				return fmt.Errorf(
					"seed shift_swap_requests[%s]: requester schedule alias %q missing in seed state",
					item.Alias,
					item.RequesterScheduleAlias,
				)
			}
			recipientScheduleID, ok := env.State.ScheduleID(strings.TrimSpace(item.RecipientScheduleAlias))
			if !ok {
				return fmt.Errorf(
					"seed shift_swap_requests[%s]: recipient schedule alias %q missing in seed state",
					item.Alias,
					item.RecipientScheduleAlias,
				)
			}
			recipientEmployeeID, ok := env.State.EmployeeID(strings.TrimSpace(item.RecipientEmployeeAlias))
			if !ok {
				return fmt.Errorf(
					"seed shift_swap_requests[%s]: recipient employee alias %q missing in seed state",
					item.Alias,
					item.RecipientEmployeeAlias,
				)
			}

			if exactShiftSwapRequestExists(existing, requesterEmployeeID, recipientEmployeeID, requesterScheduleID, recipientScheduleID, item) {
				continue
			}

			current := findComparableShiftSwapRequest(
				existing,
				requesterEmployeeID,
				recipientEmployeeID,
				requesterScheduleID,
				recipientScheduleID,
			)
			if current == nil {
				created, err := scheduleService.CreateShiftSwapRequest(ctx, requesterEmployeeID, &domain.CreateShiftSwapRequest{
					RecipientEmployeeID: recipientEmployeeID,
					RequesterScheduleID: requesterScheduleID,
					RecipientScheduleID: recipientScheduleID,
					ExpiresAt:           normalizeShiftSwapOptionalTime(item.ExpiresAt),
				})
				if err != nil {
					return fmt.Errorf("seed shift_swap_requests[%s]: create: %w", item.Alias, err)
				}

				details, err := scheduleRepo.GetShiftSwapRequestDetailsByID(ctx, created.ID)
				if err != nil {
					return fmt.Errorf("seed shift_swap_requests[%s]: load created request: %w", item.Alias, err)
				}
				existing = append(existing, *details)
				current = details
			}

			updated, err := advanceShiftSwapRequestToDesiredState(ctx, scheduleService, scheduleRepo, env, current, item)
			if err != nil {
				return fmt.Errorf("seed shift_swap_requests[%s]: %w", item.Alias, err)
			}
			replaceShiftSwapRequest(existing, *updated)
		}
	}

	return nil
}

func exactShiftSwapRequestExists(
	existing []domain.ShiftSwapResponse,
	requesterEmployeeID, recipientEmployeeID, requesterScheduleID, recipientScheduleID uuid.UUID,
	item ShiftSwapRequestSeed,
) bool {
	for _, current := range existing {
		if sameShiftSwapRequest(current, requesterEmployeeID, recipientEmployeeID, requesterScheduleID, recipientScheduleID, item, true) {
			return true
		}
	}
	return false
}

func findComparableShiftSwapRequest(
	existing []domain.ShiftSwapResponse,
	requesterEmployeeID, recipientEmployeeID, requesterScheduleID, recipientScheduleID uuid.UUID,
) *domain.ShiftSwapResponse {
	for _, current := range existing {
		if current.RequesterEmployeeID != requesterEmployeeID || current.RecipientEmployeeID != recipientEmployeeID {
			continue
		}
		if current.RequesterSchedule.ID != requesterScheduleID || current.RecipientSchedule.ID != recipientScheduleID {
			continue
		}
		copy := current
		return &copy
	}
	return nil
}

func sameShiftSwapRequest(
	current domain.ShiftSwapResponse,
	requesterEmployeeID, recipientEmployeeID, requesterScheduleID, recipientScheduleID uuid.UUID,
	item ShiftSwapRequestSeed,
	includeStatus bool,
) bool {
	if current.RequesterEmployeeID != requesterEmployeeID || current.RecipientEmployeeID != recipientEmployeeID {
		return false
	}
	if current.RequesterSchedule.ID != requesterScheduleID || current.RecipientSchedule.ID != recipientScheduleID {
		return false
	}
	if !sameOptionalTime(current.ExpiresAt, item.ExpiresAt) {
		return false
	}
	if includeStatus {
		if current.Status != strings.TrimSpace(item.Status) {
			return false
		}
		if !sameOptionalText(current.RecipientResponseNote, item.RecipientResponseNote) {
			return false
		}
		if !sameOptionalText(current.AdminDecisionNote, item.AdminDecisionNote) {
			return false
		}
	}
	return true
}

func advanceShiftSwapRequestToDesiredState(
	ctx context.Context,
	scheduleService domain.ScheduleService,
	scheduleRepo domain.ScheduleRepository,
	env Env,
	current *domain.ShiftSwapResponse,
	item ShiftSwapRequestSeed,
) (*domain.ShiftSwapResponse, error) {
	desiredStatus := strings.TrimSpace(item.Status)

	if current.Status == "pending_recipient" && desiredStatus != "pending_recipient" {
		updated, err := scheduleService.RespondToShiftSwapRequest(
			ctx,
			current.RecipientEmployeeID,
			current.ID,
			&domain.RespondShiftSwapRequest{
				Decision: shiftSwapRecipientDecisionFromStatus(desiredStatus),
				Note:     normalizeOptionalText(item.RecipientResponseNote),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("record recipient response: %w", err)
		}
		current = updated
	}

	if current.Status == "pending_admin" && wantsShiftSwapAdminDecision(desiredStatus) {
		adminEmployeeID, err := resolveRequiredEmployeeAliasForShiftSwap(env, item.AdminEmployeeAlias, item.Alias, "admin")
		if err != nil {
			return nil, err
		}
		updated, err := scheduleService.AdminDecisionShiftSwapRequest(
			ctx,
			adminEmployeeID,
			current.ID,
			&domain.AdminDecisionShiftSwapRequest{
				Decision: shiftSwapAdminDecisionFromStatus(desiredStatus),
				Note:     normalizeOptionalText(item.AdminDecisionNote),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("record admin decision: %w", err)
		}
		current = updated
	}

	// Refresh final state from the repository so computed statuses like expiry are reflected.
	refreshed, err := scheduleRepo.GetShiftSwapRequestDetailsByID(ctx, current.ID)
	if err != nil {
		return nil, fmt.Errorf("reload shift swap request: %w", err)
	}
	current = refreshed

	if !sameShiftSwapRequest(
		*current,
		current.RequesterEmployeeID,
		current.RecipientEmployeeID,
		current.RequesterSchedule.ID,
		current.RecipientSchedule.ID,
		item,
		true,
	) {
		return nil, fmt.Errorf("existing request has status %q; refusing to mutate to %q", current.Status, desiredStatus)
	}

	return current, nil
}

func replaceShiftSwapRequest(existing []domain.ShiftSwapResponse, updated domain.ShiftSwapResponse) {
	for idx := range existing {
		if existing[idx].ID == updated.ID {
			existing[idx] = updated
			return
		}
	}
}

func isValidSeedShiftSwapStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "pending_recipient", "recipient_rejected", "pending_admin", "admin_rejected", "confirmed":
		return true
	default:
		return false
	}
}

func wantsShiftSwapAdminDecision(status string) bool {
	switch strings.TrimSpace(status) {
	case "admin_rejected", "confirmed":
		return true
	default:
		return false
	}
}

func shiftSwapRecipientDecisionFromStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "recipient_rejected":
		return "reject"
	default:
		return "accept"
	}
}

func shiftSwapAdminDecisionFromStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "admin_rejected":
		return "reject"
	default:
		return "approve"
	}
}

func resolveRequiredEmployeeAliasForShiftSwap(
	env Env,
	alias *string,
	requestAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed shift_swap_requests[%s]: %s employee alias is required", requestAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed shift_swap_requests[%s]: %s employee alias %q missing in seed state",
			requestAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}

func sameOptionalText(left, right *string) bool {
	leftNormalized := normalizeOptionalText(left)
	rightNormalized := normalizeOptionalText(right)
	if leftNormalized == nil && rightNormalized == nil {
		return true
	}
	if leftNormalized == nil || rightNormalized == nil {
		return false
	}
	return *leftNormalized == *rightNormalized
}

func sameOptionalTime(left, right *time.Time) bool {
	leftNormalized := normalizeShiftSwapOptionalTime(left)
	rightNormalized := normalizeShiftSwapOptionalTime(right)
	if leftNormalized == nil && rightNormalized == nil {
		return true
	}
	if leftNormalized == nil || rightNormalized == nil {
		return false
	}
	return leftNormalized.Equal(*rightNormalized)
}

func normalizeShiftSwapOptionalTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	utc := value.UTC()
	normalized := time.Date(
		utc.Year(),
		utc.Month(),
		utc.Day(),
		utc.Hour(),
		utc.Minute(),
		utc.Second(),
		utc.Nanosecond(),
		time.UTC,
	)
	return &normalized
}
