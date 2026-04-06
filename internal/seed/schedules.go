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

type ScheduleSeed struct {
	Alias                  string
	EmployeeAlias          string
	LocationAlias          string
	CreatedByEmployeeAlias string
	IsCustom               bool
	ShiftSlot              int16
	ShiftDate              time.Time
}

type SchedulesSeeder struct {
	Schedules []ScheduleSeed
}

func (s SchedulesSeeder) Name() string {
	return "schedules"
}

func (s SchedulesSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Schedules) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed schedules: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed schedules: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	scheduleRepo := repository.NewScheduleRepository(store)
	scheduleService := service.NewScheduleService(scheduleRepo, nil, nil)

	shiftsByLocationAlias := make(map[string]map[int16]domain.ScheduleLocationShift)

	for _, item := range s.Schedules {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed schedules: alias is required")
		}
		if strings.TrimSpace(item.EmployeeAlias) == "" {
			return fmt.Errorf("seed schedules[%s]: employee alias is required", item.Alias)
		}
		if strings.TrimSpace(item.LocationAlias) == "" {
			return fmt.Errorf("seed schedules[%s]: location alias is required", item.Alias)
		}
		if strings.TrimSpace(item.CreatedByEmployeeAlias) == "" {
			return fmt.Errorf("seed schedules[%s]: created-by employee alias is required", item.Alias)
		}
		if item.IsCustom {
			return fmt.Errorf("seed schedules[%s]: only preset schedules are supported in the baseline seed", item.Alias)
		}
		if item.ShiftSlot < 1 || item.ShiftSlot > 3 {
			return fmt.Errorf("seed schedules[%s]: shift slot must be between 1 and 3", item.Alias)
		}
		if item.ShiftDate.IsZero() {
			return fmt.Errorf("seed schedules[%s]: shift date is required", item.Alias)
		}

		employeeID, ok := env.State.EmployeeID(strings.TrimSpace(item.EmployeeAlias))
		if !ok {
			return fmt.Errorf("seed schedules[%s]: employee alias %q missing in seed state", item.Alias, item.EmployeeAlias)
		}
		locationID, ok := env.State.LocationID(strings.TrimSpace(item.LocationAlias))
		if !ok {
			return fmt.Errorf("seed schedules[%s]: location alias %q missing in seed state", item.Alias, item.LocationAlias)
		}
		creatorID, ok := env.State.EmployeeID(strings.TrimSpace(item.CreatedByEmployeeAlias))
		if !ok {
			return fmt.Errorf(
				"seed schedules[%s]: created-by employee alias %q missing in seed state",
				item.Alias,
				item.CreatedByEmployeeAlias,
			)
		}

		locationShift, err := resolveShiftBySlot(ctx, scheduleRepo, shiftsByLocationAlias, item.LocationAlias, locationID, item.ShiftSlot)
		if err != nil {
			return fmt.Errorf("seed schedules[%s]: resolve location shift: %w", item.Alias, err)
		}

		day := item.ShiftDate.UTC().Format("2006-01-02")
		existing, err := scheduleRepo.GetSchedulesByLocationInRange(ctx, locationID, dateOnlyUTC(item.ShiftDate), dateOnlyUTC(item.ShiftDate))
		if err != nil {
			return fmt.Errorf("seed schedules[%s]: list existing schedules: %w", item.Alias, err)
		}

		if found := findMatchingPresetSchedule(existing, employeeID, locationShift.ID); found != nil {
			env.State.PutSchedule(item.Alias, found.ScheduleID)
			continue
		}
		if conflicting := findConflictingSchedule(existing, employeeID, day); conflicting != nil {
			return fmt.Errorf(
				"seed schedules[%s]: employee already has a different schedule on %s at location %q",
				item.Alias,
				day,
				item.LocationAlias,
			)
		}

		locationShiftID := locationShift.ID
		shiftDate := day
		created, err := scheduleService.CreateSchedule(ctx, creatorID, &domain.CreateScheduleRequest{
			EmployeeIDs:     []uuid.UUID{employeeID},
			LocationID:      locationID,
			IsCustom:        false,
			LocationShiftID: &locationShiftID,
			ShiftDate:       &shiftDate,
		})
		if err != nil {
			return fmt.Errorf("seed schedules[%s]: create: %w", item.Alias, err)
		}
		if len(created) != 1 {
			return fmt.Errorf("seed schedules[%s]: expected 1 created schedule, got %d", item.Alias, len(created))
		}

		env.State.PutSchedule(item.Alias, created[0].ID)
	}

	return nil
}

func resolveShiftBySlot(
	ctx context.Context,
	repo domain.ScheduleRepository,
	cache map[string]map[int16]domain.ScheduleLocationShift,
	locationAlias string,
	locationID uuid.UUID,
	slot int16,
) (*domain.ScheduleLocationShift, error) {
	if cached, ok := cache[locationAlias]; ok {
		if shift, exists := cached[slot]; exists {
			copy := shift
			return &copy, nil
		}
	}

	items, err := repo.GetShiftsByLocationID(ctx, locationID)
	if err != nil {
		return nil, err
	}
	if _, ok := cache[locationAlias]; !ok {
		cache[locationAlias] = make(map[int16]domain.ScheduleLocationShift)
	}
	for _, item := range items {
		resolvedSlot := inferShiftSlot(item)
		cache[locationAlias][resolvedSlot] = item
	}

	shift, ok := cache[locationAlias][slot]
	if !ok {
		return nil, fmt.Errorf("no location shift found for slot %d", slot)
	}
	copy := shift
	return &copy, nil
}

func inferShiftSlot(item domain.ScheduleLocationShift) int16 {
	switch strings.TrimSpace(item.ShiftName) {
	case "Ochtenddienst":
		return 1
	case "Avonddienst":
		return 2
	case "Slaapdienst of Waakdienst":
		return 3
	default:
		startHour, _, _, _ := shiftMicrosecondsToTimeComponents(item.StartMicroseconds)
		switch {
		case startHour < 12:
			return 1
		case startHour < 20:
			return 2
		default:
			return 3
		}
	}
}

func findMatchingPresetSchedule(
	items []domain.GetSchedulesByLocationInRangeResponse,
	employeeID, locationShiftID uuid.UUID,
) *domain.Shift {
	for _, day := range items {
		for _, shift := range day.Shifts {
			if shift.EmployeeID == employeeID && shift.LocationShiftID != nil && *shift.LocationShiftID == locationShiftID {
				copy := shift
				return &copy
			}
		}
	}
	return nil
}

func shiftMicrosecondsToTimeComponents(microseconds int64) (hour, min, sec, nano int) {
	totalSeconds := microseconds / 1_000_000
	hour = int(totalSeconds / 3600)
	min = int((totalSeconds % 3600) / 60)
	sec = int(totalSeconds % 60)
	nano = int((microseconds % 1_000_000) * 1000)
	return
}

func findConflictingSchedule(
	items []domain.GetSchedulesByLocationInRangeResponse,
	employeeID uuid.UUID,
	day string,
) *domain.Shift {
	for _, item := range items {
		if item.Date != day {
			continue
		}
		for _, shift := range item.Shifts {
			if shift.EmployeeID == employeeID {
				copy := shift
				return &copy
			}
		}
	}
	return nil
}
