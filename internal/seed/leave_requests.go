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

type LeaveRequestSeed struct {
	Alias                   string
	EmployeeAlias           string
	CreatedByEmployeeAlias  *string
	DecisionByEmployeeAlias *string
	LeaveType               string
	Status                  string
	StartDate               time.Time
	EndDate                 time.Time
	Reason                  *string
	DecisionNote            *string
}

type LeaveRequestsSeeder struct {
	Requests []LeaveRequestSeed
}

func (s LeaveRequestsSeeder) Name() string {
	return "leave_requests"
}

func (s LeaveRequestsSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Requests) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed leave_requests: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed leave_requests: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	leaveRepo := repository.NewLeaveRepository(store)
	leaveService := service.NewLeaveService(leaveRepo, nil)

	requestsByEmployee := make(map[string][]LeaveRequestSeed)
	for _, item := range s.Requests {
		alias := strings.TrimSpace(item.Alias)
		if alias == "" {
			return fmt.Errorf("seed leave_requests: alias is required")
		}
		employeeAlias := strings.TrimSpace(item.EmployeeAlias)
		if employeeAlias == "" {
			return fmt.Errorf("seed leave_requests[%s]: employee alias is required", alias)
		}
		if strings.TrimSpace(item.LeaveType) == "" {
			return fmt.Errorf("seed leave_requests[%s]: leave type is required", alias)
		}
		if strings.TrimSpace(item.Status) == "" {
			return fmt.Errorf("seed leave_requests[%s]: status is required", alias)
		}
		if item.StartDate.IsZero() || item.EndDate.IsZero() {
			return fmt.Errorf("seed leave_requests[%s]: start and end dates are required", alias)
		}
		if leaveDateOnlyUTC(item.EndDate).Before(leaveDateOnlyUTC(item.StartDate)) {
			return fmt.Errorf("seed leave_requests[%s]: end date must be on or after start date", alias)
		}
		requestsByEmployee[employeeAlias] = append(requestsByEmployee[employeeAlias], item)
	}

	for employeeAlias, items := range requestsByEmployee {
		employeeID, ok := env.State.EmployeeID(employeeAlias)
		if !ok {
			return fmt.Errorf("seed leave_requests[%s]: employee alias missing in seed state", employeeAlias)
		}

		existing, err := listAllLeaveRequestsForEmployee(ctx, leaveService, employeeID)
		if err != nil {
			return fmt.Errorf("seed leave_requests[%s]: list existing requests: %w", employeeAlias, err)
		}

		for _, item := range items {
			if exactLeaveRequestExists(existing, item) {
				continue
			}

			current := findComparableLeaveRequest(existing, item)
			if current == nil {
				created, err := createSeededLeaveRequest(ctx, leaveService, env, employeeID, item)
				if err != nil {
					return fmt.Errorf("seed leave_requests[%s]: create: %w", item.Alias, err)
				}
				existing = append(existing, domain.LeaveRequestListItem{
					LeaveRequest: *created,
				})
				current = created
			}

			if current.Status == item.Status {
				continue
			}

			if current.Status != "pending" || (item.Status != "approved" && item.Status != "rejected") {
				return fmt.Errorf(
					"seed leave_requests[%s]: existing request has status %q; refusing to mutate to %q",
					item.Alias,
					current.Status,
					item.Status,
				)
			}

			decisionByEmployeeID, err := resolveRequiredEmployeeAlias(env, item.DecisionByEmployeeAlias, item.Alias, "decision")
			if err != nil {
				return err
			}
			updated, err := leaveService.DecideLeaveRequestByAdmin(ctx, decisionByEmployeeID, current.ID, domain.DecideLeaveRequestParams{
				Decision:     leaveDecisionFromStatus(item.Status),
				DecisionNote: normalizeOptionalText(item.DecisionNote),
			})
			if err != nil {
				return fmt.Errorf("seed leave_requests[%s]: decide: %w", item.Alias, err)
			}
			replaceLeaveRequest(existing, *updated)
		}
	}

	return nil
}

func createSeededLeaveRequest(
	ctx context.Context,
	leaveService domain.LeaveService,
	env Env,
	employeeID uuid.UUID,
	item LeaveRequestSeed,
) (*domain.LeaveRequest, error) {
	params := domain.CreateLeaveRequestParams{
		EmployeeID: employeeID,
		LeaveType:  strings.TrimSpace(item.LeaveType),
		StartDate:  leaveDateOnlyUTC(item.StartDate),
		EndDate:    leaveDateOnlyUTC(item.EndDate),
		Reason:     normalizeOptionalText(item.Reason),
	}

	createdByAlias := normalizeOptionalAlias(item.CreatedByEmployeeAlias)
	if createdByAlias == "" || createdByAlias == strings.TrimSpace(item.EmployeeAlias) {
		return leaveService.CreateLeaveRequest(ctx, employeeID, params)
	}

	createdByEmployeeID, err := resolveRequiredEmployeeAlias(env, item.CreatedByEmployeeAlias, item.Alias, "created_by")
	if err != nil {
		return nil, err
	}
	return leaveService.CreateLeaveRequestByAdmin(ctx, createdByEmployeeID, params)
}

func listAllLeaveRequestsForEmployee(
	ctx context.Context,
	leaveService domain.LeaveService,
	employeeID uuid.UUID,
) ([]domain.LeaveRequestListItem, error) {
	page, err := leaveService.ListMyLeaveRequests(ctx, domain.ListMyLeaveRequestsParams{
		EmployeeID: employeeID,
		Limit:      500,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func exactLeaveRequestExists(existing []domain.LeaveRequestListItem, item LeaveRequestSeed) bool {
	for _, current := range existing {
		if sameLeaveRequest(current.LeaveRequest, item, true) {
			return true
		}
	}
	return false
}

func findComparableLeaveRequest(existing []domain.LeaveRequestListItem, item LeaveRequestSeed) *domain.LeaveRequest {
	for _, current := range existing {
		if sameLeaveRequest(current.LeaveRequest, item, false) {
			copy := current.LeaveRequest
			return &copy
		}
	}
	return nil
}

func sameLeaveRequest(current domain.LeaveRequest, item LeaveRequestSeed, includeStatus bool) bool {
	if current.LeaveType != strings.TrimSpace(item.LeaveType) {
		return false
	}
	if !leaveDateOnlyUTC(current.StartDate).Equal(leaveDateOnlyUTC(item.StartDate)) {
		return false
	}
	if !leaveDateOnlyUTC(current.EndDate).Equal(leaveDateOnlyUTC(item.EndDate)) {
		return false
	}
	if normalizeOptionalText(current.Reason) == nil && normalizeOptionalText(item.Reason) != nil {
		return false
	}
	if normalizeOptionalText(current.Reason) != nil && normalizeOptionalText(item.Reason) == nil {
		return false
	}
	if normalizeOptionalText(current.Reason) != nil &&
		*normalizeOptionalText(current.Reason) != *normalizeOptionalText(item.Reason) {
		return false
	}
	if includeStatus {
		if current.Status != strings.TrimSpace(item.Status) {
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

func replaceLeaveRequest(existing []domain.LeaveRequestListItem, updated domain.LeaveRequest) {
	for idx := range existing {
		if existing[idx].ID == updated.ID {
			existing[idx].LeaveRequest = updated
			return
		}
	}
}

func resolveRequiredEmployeeAlias(
	env Env,
	alias *string,
	requestAlias string,
	fieldName string,
) (uuid.UUID, error) {
	resolved := normalizeOptionalAlias(alias)
	if resolved == "" {
		return uuid.Nil, fmt.Errorf("seed leave_requests[%s]: %s employee alias is required", requestAlias, fieldName)
	}
	employeeID, ok := env.State.EmployeeID(resolved)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"seed leave_requests[%s]: %s employee alias %q missing in seed state",
			requestAlias,
			fieldName,
			resolved,
		)
	}
	return employeeID, nil
}

func normalizeOptionalAlias(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func leaveDateOnlyUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func leaveDecisionFromStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "approved":
		return "approve"
	case "rejected":
		return "reject"
	default:
		return strings.TrimSpace(status)
	}
}
