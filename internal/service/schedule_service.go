package service

import (
	"context"
	"fmt"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ScheduleService struct {
	repository  domain.ScheduleRepository
	asynqClient domain.TaskQueue
	logger      domain.Logger
}

func NewScheduleService(
	repository domain.ScheduleRepository,
	asynqClient domain.TaskQueue,
	logger domain.Logger,
) domain.ScheduleService {
	return &ScheduleService{
		repository:  repository,
		asynqClient: asynqClient,
		logger:      logger,
	}
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	creatorID uuid.UUID,
	req *domain.CreateScheduleRequest,
) ([]domain.CreateScheduleResponse, error) {
	if err := s.validateCreateScheduleRequest(req); err != nil {
		return nil, err
	}

	recurrence, err := s.resolveRecurrence(req)
	if err != nil {
		return nil, err
	}

	results := make([]domain.CreateScheduleResponse, 0)
	if req.IsCustom {
		dates := s.buildCustomScheduleDates(*req.StartDatetime, recurrence)
		duration := req.EndDatetime.Sub(*req.StartDatetime)
		for _, date := range dates {
			start := time.Date(
				date.Year(),
				date.Month(),
				date.Day(),
				req.StartDatetime.Hour(),
				req.StartDatetime.Minute(),
				req.StartDatetime.Second(),
				req.StartDatetime.Nanosecond(),
				req.StartDatetime.Location(),
			)
			end := start.Add(duration)
			for _, assigneeID := range req.EmployeeIDs {
				res, createErr := s.createCustomSchedule(
					ctx,
					creatorID,
					assigneeID,
					req.LocationID,
					start,
					end,
				)
				if createErr != nil {
					return nil, createErr
				}
				results = append(results, *res)
				s.sendNotificationForNewSchedule(
					ctx,
					res.ID,
					creatorID,
					assigneeID,
					res.StartDatetime,
					res.EndDatetime,
					res.LocationName,
				)
			}
		}
		return results, nil
	}

	locationShift, locationTZ, err := s.getPresetScheduleContext(ctx, req)
	if err != nil {
		return nil, err
	}

	baseDate, err := time.ParseInLocation("2006-01-02", *req.ShiftDate, locationTZ)
	if err != nil {
		s.logError(
			ctx,
			"CreateSchedule",
			"invalid shift_date format",
			err,
			zap.String("shift_date", *req.ShiftDate),
		)
		return nil, fmt.Errorf("invalid shift_date format: %w", err)
	}

	dates := s.buildCustomScheduleDates(baseDate, recurrence)
	for _, date := range dates {
		for _, assigneeID := range req.EmployeeIDs {
			res, createErr := s.createPresetScheduleForDate(
				ctx,
				creatorID,
				assigneeID,
				req.LocationID,
				req.LocationShiftID,
				locationShift,
				date,
				locationTZ,
			)
			if createErr != nil {
				return nil, createErr
			}
			results = append(results, *res)
			s.sendNotificationForNewSchedule(
				ctx,
				res.ID,
				creatorID,
				assigneeID,
				res.StartDatetime,
				res.EndDatetime,
				res.LocationName,
			)
		}
	}
	return results, nil
}

func (s *ScheduleService) GetSchedulesByLocationInRange(
	ctx context.Context,
	locationID uuid.UUID,
	req *domain.GetSchedulesByLocationInRangeRequest,
) ([]domain.GetSchedulesByLocationInRangeResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		s.logError(
			ctx,
			"GetSchedulesByLocationInRange",
			"invalid start_date format",
			err,
			zap.String("start_date", req.StartDate),
		)
		return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		s.logError(
			ctx,
			"GetSchedulesByLocationInRange",
			"invalid end_date format",
			err,
			zap.String("end_date", req.EndDate),
		)
		return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		s.logError(
			ctx,
			"GetSchedulesByLocationInRange",
			"end_date is before start_date",
			nil,
			zap.String("start_date", req.StartDate),
			zap.String("end_date", req.EndDate),
		)
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	rows, err := s.repository.GetSchedulesByLocationInRange(ctx, locationID, startDate, endDate)
	if err != nil {
		s.logError(ctx, "GetSchedulesByLocationInRange", "failed to list schedules by range", err)
		return nil, fmt.Errorf("failed to list schedules by range: %w", err)
	}
	return rows, nil
}

func (s *ScheduleService) GetScheduleByID(
	ctx context.Context,
	scheduleID uuid.UUID,
) (*domain.GetScheduleByIdResponse, error) {
	item, err := s.repository.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		s.logError(ctx, "GetScheduleByID", "failed to get schedule by id", err)
		return nil, fmt.Errorf("failed to get schedule by ID: %w", err)
	}
	return item, nil
}

func (s *ScheduleService) UpdateSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	updaterEmployeeID uuid.UUID,
	req *domain.UpdateScheduleRequest,
) (*domain.UpdateScheduleResponse, error) {
	existingSchedule, err := s.repository.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		s.logError(ctx, "UpdateSchedule", "failed to fetch existing schedule", err)
		return nil, fmt.Errorf("failed to fetch existing schedule: %w", err)
	}

	isCustom := s.determineScheduleType(req, existingSchedule)
	var res *domain.UpdateScheduleResponse
	if isCustom {
		res, err = s.updateCustomSchedule(ctx, scheduleID, existingSchedule, req)
		if err != nil {
			return nil, err
		}
	} else {
		res, err = s.updatePresetSchedule(ctx, scheduleID, existingSchedule, req)
		if err != nil {
			return nil, err
		}
	}

	s.sendNotificationForUpdatedSchedule(
		ctx,
		res.ID,
		updaterEmployeeID,
		res.EmployeeID,
		res.StartDatetime,
		res.EndDatetime,
		res.LocationName,
	)
	return res, nil
}

func (s *ScheduleService) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	if err := s.repository.DeleteSchedule(ctx, scheduleID); err != nil {
		s.logError(ctx, "DeleteSchedule", "failed to delete schedule", err)
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	return nil
}

func (s *ScheduleService) validateCreateScheduleRequest(req *domain.CreateScheduleRequest) error {
	if len(req.EmployeeIDs) == 0 {
		return fmt.Errorf("employee_ids is required")
	}

	seen := make(map[uuid.UUID]struct{}, len(req.EmployeeIDs))
	for _, employeeID := range req.EmployeeIDs {
		if employeeID == uuid.Nil {
			return fmt.Errorf("employee_ids contains invalid uuid")
		}
		if _, exists := seen[employeeID]; exists {
			return fmt.Errorf("employee_ids must not contain duplicates")
		}
		seen[employeeID] = struct{}{}
	}

	if req.LocationID == uuid.Nil {
		return fmt.Errorf("location_id is required")
	}
	if req.IsCustom {
		return s.validateCustomSchedule(req)
	}
	return s.validatePresetSchedule(req)
}

func (s *ScheduleService) resolveRecurrence(req *domain.CreateScheduleRequest) (string, error) {
	if req.Recurrence == nil || *req.Recurrence == "" {
		return domain.CreateScheduleRecurrenceNone, nil
	}

	switch *req.Recurrence {
	case domain.CreateScheduleRecurrenceNone,
		domain.CreateScheduleRecurrenceEndOfWeek,
		domain.CreateScheduleRecurrenceEndOfMonth:
		return *req.Recurrence, nil
	default:
		return "", fmt.Errorf("recurrence must be one of: none, end_of_week, end_of_month")
	}
}

func (s *ScheduleService) buildCustomScheduleDates(
	baseDate time.Time,
	recurrence string,
) []time.Time {
	dayStart := time.Date(
		baseDate.Year(),
		baseDate.Month(),
		baseDate.Day(),
		0,
		0,
		0,
		0,
		baseDate.Location(),
	)
	endDate := dayStart

	switch recurrence {
	case domain.CreateScheduleRecurrenceEndOfWeek:
		daysUntilSunday := (7 - int(dayStart.Weekday())) % 7
		endDate = dayStart.AddDate(0, 0, daysUntilSunday)
	case domain.CreateScheduleRecurrenceEndOfMonth:
		endDate = time.Date(dayStart.Year(), dayStart.Month()+1, 0, 0, 0, 0, 0, dayStart.Location())
	}

	dates := make([]time.Time, 0)
	for date := dayStart; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date)
	}
	return dates
}

func (s *ScheduleService) validateCustomSchedule(req *domain.CreateScheduleRequest) error {
	if req.StartDatetime == nil || req.EndDatetime == nil {
		return fmt.Errorf("start_datetime and end_datetime are required for custom schedules")
	}
	if req.StartDatetime.After(*req.EndDatetime) {
		return fmt.Errorf("start_datetime must be before end_datetime")
	}
	if req.LocationShiftID != nil || req.ShiftDate != nil {
		return fmt.Errorf(
			"location_shift_id and shift_date should not be provided for custom schedules",
		)
	}
	return nil
}

func (s *ScheduleService) validatePresetSchedule(req *domain.CreateScheduleRequest) error {
	if req.LocationShiftID == nil || req.ShiftDate == nil {
		return fmt.Errorf(
			"location_shift_id and shift_date are required for preset shift schedules",
		)
	}
	if req.StartDatetime != nil || req.EndDatetime != nil {
		return fmt.Errorf(
			"start_datetime and end_datetime should not be provided for preset shift schedules",
		)
	}
	return nil
}

func (s *ScheduleService) createCustomSchedule(
	ctx context.Context,
	creatorID, assigneeID, locationID uuid.UUID,
	startDatetime, endDatetime time.Time,
) (*domain.CreateScheduleResponse, error) {
	schedule, err := s.repository.CreateSchedule(ctx, domain.CreateScheduleParams{
		EmployeeID:             assigneeID,
		LocationID:             locationID,
		IsCustom:               true,
		LocationShiftID:        nil,
		ShiftNameSnapshot:      nil,
		ShiftStartTimeSnapshot: nil,
		ShiftEndTimeSnapshot:   nil,
		CreatedByEmployeeID:    creatorID,
		StartDatetime:          startDatetime,
		EndDatetime:            endDatetime,
	})
	if err != nil {
		s.logError(ctx, "createCustomSchedule", "failed to create custom schedule", err)
		return nil, fmt.Errorf("failed to create custom schedule: %w", err)
	}
	return schedule, nil
}

func (s *ScheduleService) getPresetScheduleContext(
	ctx context.Context,
	req *domain.CreateScheduleRequest,
) (*domain.ScheduleLocationShift, *time.Location, error) {
	if err := s.validatePresetSchedule(req); err != nil {
		s.logError(ctx, "getPresetScheduleContext", "preset schedule validation failed", err)
		return nil, nil, err
	}

	locationShift, err := s.repository.GetShiftByID(ctx, *req.LocationShiftID)
	if err != nil {
		s.logError(ctx, "getPresetScheduleContext", "failed to fetch location shift", err)
		return nil, nil, fmt.Errorf("failed to fetch location shift: %w", err)
	}
	if locationShift.LocationID != req.LocationID {
		return nil, nil, fmt.Errorf("location shift does not belong to the specified location")
	}

	location, err := s.repository.GetLocationByID(ctx, req.LocationID)
	if err != nil {
		s.logError(ctx, "getPresetScheduleContext", "failed to fetch location", err)
		return nil, nil, fmt.Errorf("failed to fetch location: %w", err)
	}
	locationTZ, err := time.LoadLocation(location.Timezone)
	if err != nil {
		s.logError(
			ctx,
			"getPresetScheduleContext",
			"invalid location timezone",
			err,
			zap.String("location_timezone", location.Timezone),
		)
		return nil, nil, fmt.Errorf("invalid location timezone: %w", err)
	}

	return locationShift, locationTZ, nil
}

func (s *ScheduleService) createPresetScheduleForDate(
	ctx context.Context,
	creatorID, assigneeID, locationID uuid.UUID,
	locationShiftID *uuid.UUID,
	locationShift *domain.ScheduleLocationShift,
	shiftDate time.Time,
	locationTZ *time.Location,
) (*domain.CreateScheduleResponse, error) {
	shiftDate = time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		0,
		0,
		0,
		0,
		locationTZ,
	)
	startHour, startMin, startSec, startNano := microsecondsToTimeComponents(
		locationShift.StartMicroseconds,
	)
	endHour, endMin, endSec, endNano := microsecondsToTimeComponents(locationShift.EndMicroseconds)

	startDatetime := time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		startHour,
		startMin,
		startSec,
		startNano,
		locationTZ,
	)
	endDatetime := time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		endHour,
		endMin,
		endSec,
		endNano,
		locationTZ,
	)
	if locationShift.EndMicroseconds < locationShift.StartMicroseconds {
		endDatetime = endDatetime.AddDate(0, 0, 1)
	}

	shiftStart := locationShift.StartMicroseconds
	shiftEnd := locationShift.EndMicroseconds
	schedule, err := s.repository.CreateSchedule(ctx, domain.CreateScheduleParams{
		EmployeeID:             assigneeID,
		LocationID:             locationID,
		LocationShiftID:        locationShiftID,
		ShiftNameSnapshot:      &locationShift.ShiftName,
		ShiftStartTimeSnapshot: &shiftStart,
		ShiftEndTimeSnapshot:   &shiftEnd,
		IsCustom:               false,
		CreatedByEmployeeID:    creatorID,
		StartDatetime:          startDatetime,
		EndDatetime:            endDatetime,
	})
	if err != nil {
		s.logError(ctx, "createPresetSchedule", "failed to create preset schedule", err)
		return nil, fmt.Errorf("failed to create preset schedule: %w", err)
	}
	return schedule, nil
}

func (s *ScheduleService) sendNotificationForNewSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	creatorID, recipientID uuid.UUID,
	startTime, endTime time.Time,
	locationName string,
) {
	if s.asynqClient == nil {
		return
	}
	notifData := &domain.NewScheduleNotificationTaskData{
		ScheduleID: scheduleID,
		CreatedBy:  creatorID,
		StartTime:  startTime,
		EndTime:    endTime,
		Location:   locationName,
	}
	err := s.asynqClient.EnqueueNotificationTask(ctx, domain.NotificationTaskPayload{
		RecipientUserIDs: []uuid.UUID{recipientID},
		Type:             domain.TypeNewScheduleNotification,
		Data:             domain.NotificationTaskData{NewScheduleNotification: notifData},
		CreatedAt:        time.Now(),
		Message:          "notifData.NewScheduleMessage()",
	}, &domain.TaskEnqueueOptions{MaxRetry: 3})
	if err != nil {
		s.logError(
			ctx,
			"sendNotificationForNewSchedule",
			"failed to enqueue new schedule notification",
			err,
			zap.String("schedule_id", scheduleID.String()),
		)
	}
}

func (s *ScheduleService) determineScheduleType(
	req *domain.UpdateScheduleRequest,
	existingSchedule *domain.GetScheduleByIdResponse,
) bool {
	if req.IsCustom != nil {
		isCustom := *req.IsCustom
		if !isCustom {
			req.StartDatetime = nil
			req.EndDatetime = nil
		}
		return isCustom
	}
	if req.LocationShiftID != nil || req.ShiftDate != nil {
		req.StartDatetime = nil
		req.EndDatetime = nil
		return false
	}
	return existingSchedule.LocationShiftID == nil
}

func (s *ScheduleService) updateCustomSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	existingSchedule *domain.GetScheduleByIdResponse,
	req *domain.UpdateScheduleRequest,
) (*domain.UpdateScheduleResponse, error) {
	if err := s.validateCustomScheduleUpdate(req); err != nil {
		s.logError(ctx, "updateCustomSchedule", "custom schedule validation failed", err)
		return nil, err
	}

	employeeID := existingSchedule.EmployeeID
	locationID := existingSchedule.LocationID
	startDatetime := existingSchedule.StartDatetime
	endDatetime := existingSchedule.EndDatetime

	if req.EmployeeID != nil {
		employeeID = *req.EmployeeID
	}
	if req.LocationID != nil {
		locationID = *req.LocationID
	}
	if req.StartDatetime != nil {
		startDatetime = *req.StartDatetime
	}
	if req.EndDatetime != nil {
		endDatetime = *req.EndDatetime
	}
	if startDatetime.After(endDatetime) {
		return nil, fmt.Errorf("start_datetime must be before end_datetime")
	}

	schedule, err := s.repository.UpdateSchedule(ctx, scheduleID, domain.UpdateScheduleParams{
		EmployeeID:             employeeID,
		LocationID:             locationID,
		LocationShiftID:        nil,
		ShiftNameSnapshot:      nil,
		ShiftStartTimeSnapshot: nil,
		ShiftEndTimeSnapshot:   nil,
		IsCustom:               true,
		StartDatetime:          startDatetime,
		EndDatetime:            endDatetime,
	})
	if err != nil {
		s.logError(ctx, "updateCustomSchedule", "failed to update custom schedule", err)
		return nil, fmt.Errorf("failed to update custom schedule: %w", err)
	}
	return schedule, nil
}

func (s *ScheduleService) updatePresetSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	existingSchedule *domain.GetScheduleByIdResponse,
	req *domain.UpdateScheduleRequest,
) (*domain.UpdateScheduleResponse, error) {
	employeeID := existingSchedule.EmployeeID
	locationID := existingSchedule.LocationID
	if req.EmployeeID != nil {
		employeeID = *req.EmployeeID
	}
	if req.LocationID != nil {
		locationID = *req.LocationID
	}

	var shiftIDToUse uuid.UUID
	if req.LocationShiftID != nil {
		shiftIDToUse = *req.LocationShiftID
	} else if existingSchedule.LocationShiftID != nil {
		shiftIDToUse = *existingSchedule.LocationShiftID
	} else {
		return nil, fmt.Errorf("location_shift_id is required for preset shift schedules")
	}

	shiftDateToUse := existingSchedule.StartDatetime.Format("2006-01-02")
	if req.ShiftDate != nil {
		shiftDateToUse = *req.ShiftDate
	}

	locationShift, err := s.repository.GetShiftByID(ctx, shiftIDToUse)
	if err != nil {
		s.logError(ctx, "updatePresetSchedule", "failed to fetch location shift", err)
		return nil, fmt.Errorf("invalid location_shift_id: %w", err)
	}
	if locationShift.LocationID != locationID {
		return nil, fmt.Errorf("location_shift_id does not belong to the specified location")
	}

	shiftDate, err := time.Parse("2006-01-02", shiftDateToUse)
	if err != nil {
		return nil, fmt.Errorf("invalid shift_date format, expected YYYY-MM-DD: %w", err)
	}

	location, err := s.repository.GetLocationByID(ctx, locationID)
	if err != nil {
		s.logError(ctx, "updatePresetSchedule", "failed to fetch location", err)
		return nil, fmt.Errorf("failed to fetch location: %w", err)
	}
	locationTZ, err := time.LoadLocation(location.Timezone)
	if err != nil {
		s.logError(
			ctx,
			"updatePresetSchedule",
			"invalid location timezone",
			err,
			zap.String("location_timezone", location.Timezone),
		)
		return nil, fmt.Errorf("invalid location timezone: %w", err)
	}

	shiftDate = time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		0,
		0,
		0,
		0,
		locationTZ,
	)
	startHour, startMin, startSec, startNano := microsecondsToTimeComponents(
		locationShift.StartMicroseconds,
	)
	endHour, endMin, endSec, endNano := microsecondsToTimeComponents(locationShift.EndMicroseconds)
	startDatetime := time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		startHour,
		startMin,
		startSec,
		startNano,
		locationTZ,
	)
	endDatetime := time.Date(
		shiftDate.Year(),
		shiftDate.Month(),
		shiftDate.Day(),
		endHour,
		endMin,
		endSec,
		endNano,
		locationTZ,
	)
	if locationShift.EndMicroseconds < locationShift.StartMicroseconds {
		endDatetime = endDatetime.AddDate(0, 0, 1)
	}

	shiftStart := locationShift.StartMicroseconds
	shiftEnd := locationShift.EndMicroseconds
	schedule, err := s.repository.UpdateSchedule(ctx, scheduleID, domain.UpdateScheduleParams{
		EmployeeID:             employeeID,
		LocationID:             locationID,
		LocationShiftID:        &shiftIDToUse,
		ShiftNameSnapshot:      &locationShift.ShiftName,
		ShiftStartTimeSnapshot: &shiftStart,
		ShiftEndTimeSnapshot:   &shiftEnd,
		IsCustom:               false,
		StartDatetime:          startDatetime,
		EndDatetime:            endDatetime,
	})
	if err != nil {
		s.logError(ctx, "updatePresetSchedule", "failed to update preset schedule", err)
		return nil, fmt.Errorf("failed to update preset schedule: %w", err)
	}
	return schedule, nil
}

func (s *ScheduleService) validateCustomScheduleUpdate(req *domain.UpdateScheduleRequest) error {
	if req.LocationShiftID != nil || req.ShiftDate != nil {
		return fmt.Errorf(
			"location_shift_id and shift_date should not be provided for custom schedules",
		)
	}
	return nil
}

func (s *ScheduleService) sendNotificationForUpdatedSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	updaterEmployeeID, recipientEmployeeID uuid.UUID,
	startTime, endTime time.Time,
	locationName string,
) {
	if s.asynqClient == nil {
		return
	}
	notifData := &domain.NewScheduleNotificationTaskData{
		ScheduleID: scheduleID,
		CreatedBy:  updaterEmployeeID,
		StartTime:  startTime,
		EndTime:    endTime,
		Location:   locationName,
	}
	err := s.asynqClient.EnqueueNotificationTask(ctx, domain.NotificationTaskPayload{
		RecipientUserIDs: []uuid.UUID{recipientEmployeeID},
		Type:             domain.TypeNewScheduleNotification,
		Data:             domain.NotificationTaskData{NewScheduleNotification: notifData},
		CreatedAt:        time.Now(),
		Message:          "notifData.UpdatedScheduleMessage()",
	}, &domain.TaskEnqueueOptions{MaxRetry: 3})
	if err != nil {
		s.logError(
			ctx,
			"sendNotificationForUpdatedSchedule",
			"failed to enqueue notification task",
			err,
		)
	}
}

func (s *ScheduleService) logError(
	ctx context.Context,
	operation, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger == nil {
		return
	}
	s.logger.LogError(ctx, "ScheduleService."+operation, message, err, fields...)
}

func microsecondsToTimeComponents(microseconds int64) (hour, min, sec, nano int) {
	totalSeconds := microseconds / 1_000_000
	hour = int(totalSeconds / 3600)
	min = int((totalSeconds % 3600) / 60)
	sec = int(totalSeconds % 60)
	nano = int((microseconds % 1_000_000) * 1000)
	return
}
