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

type TimeEntrySeed struct {
	Alias                    string
	ScheduleAlias            string
	EmployeeAlias            string
	SubmittedByEmployeeAlias *string
	ApprovedByEmployeeAlias  *string
	Status                   string
	StartTime                string
	EndTime                  string
	BreakMinutes             int32
	HourType                 string
	Notes                    *string
	RejectionReason          *string
}

type TimeEntriesSeeder struct {
	Entries []TimeEntrySeed
}

func (s TimeEntriesSeeder) Name() string {
	return "time_entries"
}

func (s TimeEntriesSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Entries) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed time_entries: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed time_entries: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	timeEntryRepo := repository.NewTimeEntryRepository(store)
	timeEntryService := service.NewTimeEntryService(timeEntryRepo, nil)
	scheduleRepo := repository.NewScheduleRepository(store)

	entriesByEmployee := make(map[string][]TimeEntrySeed)
	for _, item := range s.Entries {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed time_entries: alias is required")
		}
		if strings.TrimSpace(item.ScheduleAlias) == "" {
			return fmt.Errorf("seed time_entries[%s]: schedule alias is required", alias)
		}
		if strings.TrimSpace(item.EmployeeAlias) == "" {
			return fmt.Errorf("seed time_entries[%s]: employee alias is required", alias)
		}
		if !isValidSeedTimeEntryStatus(item.Status) {
			return fmt.Errorf("seed time_entries[%s]: unsupported status %q", alias, item.Status)
		}
		if wantsTimeEntryDecision(item.Status) && item.ApprovedByEmployeeAlias == nil {
			return fmt.Errorf("seed time_entries[%s]: approver alias is required for status %q", alias, item.Status)
		}
		if strings.TrimSpace(item.Status) == domain.TimeEntryStatusRejected && normalizeOptionalText(item.RejectionReason) == nil {
			return fmt.Errorf("seed time_entries[%s]: rejection reason is required for rejected entries", alias)
		}
		entriesByEmployee[strings.TrimSpace(item.EmployeeAlias)] = append(entriesByEmployee[strings.TrimSpace(item.EmployeeAlias)], item)
	}

	for employeeAlias, items := range entriesByEmployee {
		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed time_entries[%s]: employee alias missing in seed state", employeeAlias)
		}

		existing, err := listAllTimeEntriesForEmployee(ctx, timeEntryService, employeeID)
		if err != nil {
			return fmt.Errorf("seed time_entries[%s]: list existing entries: %w", employeeAlias, err)
		}

		for _, item := range items {
			scheduleID, ok := env.State.ScheduleID(strings.TrimSpace(item.ScheduleAlias))
			if !ok {
				return fmt.Errorf("seed time_entries[%s]: schedule alias %q missing in seed state", item.Alias, item.ScheduleAlias)
			}
			scheduleDetail, err := scheduleRepo.GetScheduleByID(ctx, scheduleID)
			if err != nil {
				return fmt.Errorf("seed time_entries[%s]: load schedule: %w", item.Alias, err)
			}

			if exactTimeEntryExists(existing, scheduleID, item) {
				continue
			}

			current := findComparableTimeEntry(existing, scheduleID, item)
			if current == nil {
				created, err := timeEntryService.CreateTimeEntry(ctx, employeeID, domain.CreateTimeEntryParams{
					ScheduleID:   &scheduleID,
					EntryDate:    scheduleDetail.StartDatetime.UTC(),
					StartTime:    strings.TrimSpace(item.StartTime),
					EndTime:      strings.TrimSpace(item.EndTime),
					BreakMinutes: item.BreakMinutes,
					HourType:     strings.TrimSpace(item.HourType),
					Notes:        normalizeOptionalText(item.Notes),
				})
				if err != nil {
					return fmt.Errorf("seed time_entries[%s]: create: %w", item.Alias, err)
				}

				existing = append(existing, *created)
				current = created
			}

			updated, err := advanceTimeEntryToDesiredState(ctx, env, current, item)
			if err != nil {
				return fmt.Errorf("seed time_entries[%s]: %w", item.Alias, err)
			}
			replaceTimeEntry(existing, *updated)
		}
	}

	return nil
}

func listAllTimeEntriesForEmployee(
	ctx context.Context,
	timeEntryService domain.TimeEntryService,
	employeeID uuid.UUID,
) ([]domain.TimeEntry, error) {
	page, err := timeEntryService.ListMyTimeEntries(ctx, domain.ListMyTimeEntriesParams{
		EmployeeID: employeeID,
		Limit:      500,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func exactTimeEntryExists(existing []domain.TimeEntry, scheduleID uuid.UUID, item TimeEntrySeed) bool {
	for _, current := range existing {
		if sameTimeEntry(current, scheduleID, item, true) {
			return true
		}
	}
	return false
}

func findComparableTimeEntry(existing []domain.TimeEntry, scheduleID uuid.UUID, item TimeEntrySeed) *domain.TimeEntry {
	for _, current := range existing {
		if sameTimeEntry(current, scheduleID, item, false) {
			copy := current
			return &copy
		}
	}
	return nil
}

func sameTimeEntry(current domain.TimeEntry, scheduleID uuid.UUID, item TimeEntrySeed, includeStatus bool) bool {
	if current.ScheduleID == nil || *current.ScheduleID != scheduleID {
		return false
	}
	if current.StartTime != strings.TrimSpace(item.StartTime) || current.EndTime != strings.TrimSpace(item.EndTime) {
		return false
	}
	if current.BreakMinutes != item.BreakMinutes || current.HourType != strings.TrimSpace(item.HourType) {
		return false
	}
	if normalizeOptionalText(current.Notes) == nil && normalizeOptionalText(item.Notes) != nil {
		return false
	}
	if normalizeOptionalText(current.Notes) != nil && normalizeOptionalText(item.Notes) == nil {
		return false
	}
	if normalizeOptionalText(current.Notes) != nil &&
		*normalizeOptionalText(current.Notes) != *normalizeOptionalText(item.Notes) {
		return false
	}
	if includeStatus {
		if current.Status != strings.TrimSpace(item.Status) {
			return false
		}
		if normalizeOptionalText(current.RejectionReason) == nil && normalizeOptionalText(item.RejectionReason) != nil {
			return false
		}
		if normalizeOptionalText(current.RejectionReason) != nil && normalizeOptionalText(item.RejectionReason) == nil {
			return false
		}
		if normalizeOptionalText(current.RejectionReason) != nil &&
			*normalizeOptionalText(current.RejectionReason) != *normalizeOptionalText(item.RejectionReason) {
			return false
		}
	}
	return true
}

func advanceTimeEntryToDesiredState(
	ctx context.Context,
	env Env,
	current *domain.TimeEntry,
	item TimeEntrySeed,
) (*domain.TimeEntry, error) {
	desiredStatus := strings.TrimSpace(item.Status)

	if current.Status == domain.TimeEntryStatusDraft && desiredStatus != domain.TimeEntryStatusDraft {
		if err := markTimeEntrySubmitted(ctx, env, current.ID); err != nil {
			return nil, fmt.Errorf("submit time entry: %w", err)
		}
		current.Status = domain.TimeEntryStatusSubmitted
	}

	if current.Status == domain.TimeEntryStatusSubmitted && wantsTimeEntryDecision(desiredStatus) {
		approverID, err := resolveRequiredEmployeeAliasForTimeEntry(env, item.ApprovedByEmployeeAlias, item.Alias, "approved_by")
		if err != nil {
			return nil, err
		}
		updated, err := service.NewTimeEntryService(repository.NewTimeEntryRepository(dbrepo.NewStoreWithTx(env.DB.(pgx.Tx))), nil).
			DecideTimeEntryByAdmin(ctx, approverID, current.ID, domain.DecideTimeEntryParams{
				Decision:        timeEntryDecisionFromStatus(desiredStatus),
				RejectionReason: normalizeOptionalText(item.RejectionReason),
			})
		if err != nil {
			return nil, err
		}
		current = updated
	}

	if current.Status != desiredStatus {
		return nil, fmt.Errorf("existing entry has status %q; refusing to mutate to %q", current.Status, desiredStatus)
	}

	return current, nil
}

func markTimeEntrySubmitted(ctx context.Context, env Env, timeEntryID uuid.UUID) error {
	tag, err := env.DB.Exec(
		ctx,
		`UPDATE time_entries
		 SET status = 'submitted'::time_entry_status_enum,
		     submitted_at = COALESCE(submitted_at, NOW()),
		     updated_at = NOW()
		 WHERE id = $1
		   AND status = 'draft'::time_entry_status_enum`,
		timeEntryID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 draft time entry to submit, affected %d", tag.RowsAffected())
	}
	return nil
}

func replaceTimeEntry(existing []domain.TimeEntry, updated domain.TimeEntry) {
	for idx := range existing {
		if existing[idx].ID == updated.ID {
			existing[idx] = updated
			return
		}
	}
}

func resolveRequiredEmployeeAliasForTimeEntry(
	env Env,
	alias *string,
	entryAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed time_entries[%s]: %s employee alias is required", entryAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed time_entries[%s]: %s employee alias %q missing in seed state",
			entryAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}

func isValidSeedTimeEntryStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case domain.TimeEntryStatusDraft,
		domain.TimeEntryStatusSubmitted,
		domain.TimeEntryStatusApproved,
		domain.TimeEntryStatusRejected:
		return true
	default:
		return false
	}
}

func wantsTimeEntryDecision(status string) bool {
	switch strings.TrimSpace(status) {
	case domain.TimeEntryStatusApproved, domain.TimeEntryStatusRejected:
		return true
	default:
		return false
	}
}

func timeEntryDecisionFromStatus(status string) string {
	switch strings.TrimSpace(status) {
	case domain.TimeEntryStatusApproved:
		return "approve"
	case domain.TimeEntryStatusRejected:
		return "reject"
	default:
		return strings.TrimSpace(status)
	}
}
