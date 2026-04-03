package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type LateArrivalService struct {
	repository domain.LateArrivalRepository
	logger     domain.Logger
}

func NewLateArrivalService(
	repository domain.LateArrivalRepository,
	logger domain.Logger,
) domain.LateArrivalService {
	return &LateArrivalService{
		repository: repository,
		logger:     logger,
	}
}

func (s *LateArrivalService) CreateLateArrival(
	ctx context.Context,
	params domain.LateArrivalCreateParams,
) (*domain.CreateLateArrivalResult, error) {
	if params.EmployeeID == uuid.Nil || params.CreatedByEmployeeID == uuid.Nil {
		return nil, domain.ErrLateArrivalInvalidRequest
	}
	params.CreatedByEmployeeID = params.EmployeeID
	return s.createLateArrival(ctx, params)
}

func (s *LateArrivalService) CreateLateArrivalByAdmin(
	ctx context.Context,
	params domain.LateArrivalCreateParams,
) (*domain.CreateLateArrivalResult, error) {
	if params.EmployeeID == uuid.Nil || params.CreatedByEmployeeID == uuid.Nil {
		return nil, domain.ErrLateArrivalInvalidRequest
	}
	return s.createLateArrival(ctx, params)
}

func (s *LateArrivalService) createLateArrival(
	ctx context.Context,
	params domain.LateArrivalCreateParams,
) (*domain.CreateLateArrivalResult, error) {
	if params.ArrivalDate.IsZero() {
		return nil, domain.ErrLateArrivalInvalidRequest
	}

	reason := strings.TrimSpace(params.Reason)
	if reason == "" {
		return nil, fmt.Errorf("%w: reason is required", domain.ErrLateArrivalInvalidRequest)
	}
	params.Reason = reason

	hour, minute, second, normalizedArrivalTime, err := parseLateArrivalTime(params.ArrivalTime)
	if err != nil {
		return nil, err
	}
	params.ArrivalTime = normalizedArrivalTime

	schedules, err := s.repository.ListAssignedSchedulesForEmployeeOnDate(
		ctx,
		params.EmployeeID,
		params.ArrivalDate,
	)
	if err != nil {
		return nil, err
	}
	if len(schedules) == 0 {
		return nil, fmt.Errorf(
			"%w: no assigned shift found for selected date",
			domain.ErrLateArrivalInvalidRequest,
		)
	}
	if len(schedules) > 1 {
		return nil, fmt.Errorf(
			"%w: multiple assigned shifts found for selected date",
			domain.ErrLateArrivalConflict,
		)
	}

	resolved := schedules[0]
	locationTZ, err := time.LoadLocation(resolved.LocationTimezone)
	if err != nil {
		return nil, err
	}

	arrivalAtLocal := time.Date(
		params.ArrivalDate.Year(),
		params.ArrivalDate.Month(),
		params.ArrivalDate.Day(),
		hour,
		minute,
		second,
		0,
		locationTZ,
	)
	shiftStartLocal := resolved.StartDatetime.In(locationTZ)
	if !arrivalAtLocal.After(shiftStartLocal) {
		return nil, fmt.Errorf(
			"%w: arrival_time must be after shift start",
			domain.ErrLateArrivalInvalidRequest,
		)
	}

	created, err := s.repository.CreateLateArrival(ctx, params, resolved.ScheduleID)
	if err != nil {
		return nil, err
	}

	return &domain.CreateLateArrivalResult{
		ID:                  created.ID,
		ScheduleID:          created.ScheduleID,
		EmployeeID:          created.EmployeeID,
		CreatedByEmployeeID: created.CreatedByEmployeeID,
		ArrivalDate:         created.ArrivalDate,
		ArrivalTime:         created.ArrivalTime,
		Reason:              created.Reason,
		ShiftStartDatetime:  resolved.StartDatetime,
		ShiftEndDatetime:    resolved.EndDatetime,
		ShiftName:           resolved.ShiftName,
		LocationName:        resolved.LocationName,
		CreatedAt:           created.CreatedAt,
		UpdatedAt:           created.UpdatedAt,
	}, nil
}

func (s *LateArrivalService) ListMyLateArrivals(
	ctx context.Context,
	params domain.ListMyLateArrivalsParams,
) (*domain.LateArrivalPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrLateArrivalInvalidRequest
	}
	if err := validateLateArrivalDateRange(params.DateFrom, params.DateTo); err != nil {
		return nil, err
	}
	return s.repository.ListMyLateArrivals(ctx, params)
}

func (s *LateArrivalService) ListLateArrivals(
	ctx context.Context,
	params domain.ListLateArrivalsParams,
) (*domain.LateArrivalPage, error) {
	if err := validateLateArrivalDateRange(params.DateFrom, params.DateTo); err != nil {
		return nil, err
	}
	return s.repository.ListLateArrivals(ctx, params)
}

func parseLateArrivalTime(value string) (hour, minute, second int, normalized string, err error) {
	input := strings.TrimSpace(value)
	if input == "" {
		return 0, 0, 0, "", domain.ErrLateArrivalInvalidRequest
	}

	var parsed time.Time
	parsed, err = time.Parse("15:04:05", input)
	if err != nil {
		parsed, err = time.Parse("15:04", input)
		if err != nil {
			return 0, 0, 0, "", fmt.Errorf(
				"%w: invalid arrival_time format",
				domain.ErrLateArrivalInvalidRequest,
			)
		}
	}

	return parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Format("15:04:05"), nil
}

func validateLateArrivalDateRange(dateFrom, dateTo *time.Time) error {
	if dateFrom == nil || dateTo == nil {
		return nil
	}
	if dateTo.Before(*dateFrom) {
		return fmt.Errorf(
			"%w: date_to must be on or after date_from",
			domain.ErrLateArrivalInvalidRequest,
		)
	}
	return nil
}

var _ domain.LateArrivalService = (*LateArrivalService)(nil)
