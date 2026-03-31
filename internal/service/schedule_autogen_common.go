package service

import (
	"context"
	"fmt"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func isoWeekStartDate(year int, week int, loc *time.Location) (time.Time, error) {
	if week < 1 || week > 53 {
		return time.Time{}, fmt.Errorf("invalid ISO week: %d", week)
	}
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, loc)
	isoDow := int(jan4.Weekday())
	if isoDow == 0 {
		isoDow = 7
	}
	week1Start := jan4.AddDate(0, 0, -(isoDow - 1))
	return week1Start.AddDate(0, 0, (week-1)*7), nil
}

func (s *ScheduleService) ensureWeekEmpty(ctx context.Context, locationID uuid.UUID, week int32, year int32, locationTZ *time.Location) error {
	weekStart, err := isoWeekStartDate(int(year), int(week), locationTZ)
	if err != nil {
		return err
	}
	weekEnd := weekStart.AddDate(0, 0, 6)
	rows, err := s.repository.GetSchedulesByLocationInRange(ctx, locationID, weekStart, weekEnd)
	if err != nil {
		return fmt.Errorf("failed to check existing schedules: %w", err)
	}
	if len(rows) > 0 {
		return domain.ErrWeekNotEmpty
	}
	return nil
}

func (s *ScheduleService) validateAutoGenerateRequest(req *domain.AutoGenerateSchedulesRequest) error {
	if req.LocationID == uuid.Nil {
		return fmt.Errorf("location_id is required")
	}
	if req.Week < 1 || req.Week > 53 {
		return fmt.Errorf("invalid week")
	}
	if req.Year <= 0 {
		return fmt.Errorf("invalid year")
	}
	if len(req.EmployeeIDs) == 0 {
		return fmt.Errorf("employee_ids is required")
	}
	return nil
}

func (s *ScheduleService) loadAutoGenerateInputs(ctx context.Context, req *domain.AutoGenerateSchedulesRequest) ([]domain.ScheduleEmployeeContractHours, []domain.ScheduleLocationShift, string, *time.Location, error) {
	employees, err := s.repository.ListEmployeesWithContractHours(ctx, req.EmployeeIDs)
	if err != nil {
		s.logError(ctx, "AutoGenerateSchedules", "failed to fetch employee contract hours", err)
		return nil, nil, "", nil, err
	}
	if len(employees) == 0 {
		return nil, nil, "", nil, fmt.Errorf("no employees found with contract hours")
	}

	locationShifts, err := s.repository.GetShiftsByLocationID(ctx, req.LocationID)
	if err != nil {
		s.logError(ctx, "AutoGenerateSchedules", "failed to fetch location shifts", err)
		return nil, nil, "", nil, err
	}
	if len(locationShifts) == 0 {
		return nil, nil, "", nil, fmt.Errorf("no shifts configured for location")
	}

	location, err := s.repository.GetLocationByID(ctx, req.LocationID)
	if err != nil {
		s.logError(ctx, "AutoGenerateSchedules", "failed to fetch location", err)
		return nil, nil, "", nil, err
	}
	locationTZ, err := time.LoadLocation(location.Timezone)
	if err != nil {
		s.logError(ctx, "AutoGenerateSchedules", "invalid location timezone", err, zap.String("timezone", location.Timezone))
		return nil, nil, "", nil, fmt.Errorf("invalid location timezone: %w", err)
	}

	if err := s.ensureWeekEmpty(ctx, req.LocationID, req.Week, req.Year, locationTZ); err != nil {
		return nil, nil, "", nil, err
	}

	return employees, locationShifts, location.Timezone, locationTZ, nil
}
