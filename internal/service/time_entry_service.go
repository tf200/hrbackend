package service

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TimeEntryService struct {
	repository domain.TimeEntryRepository
	logger     domain.Logger
}

func NewTimeEntryService(repository domain.TimeEntryRepository, logger domain.Logger) domain.TimeEntryService {
	return &TimeEntryService{
		repository: repository,
		logger:     logger,
	}
}

func (s *TimeEntryService) CreateTimeEntry(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CreateTimeEntryParams) (*domain.TimeEntry, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	params.EmployeeID = actorEmployeeID
	return s.createTimeEntry(ctx, params, "TimeEntryService.CreateTimeEntry")
}

func (s *TimeEntryService) CreateTimeEntryByAdmin(ctx context.Context, adminEmployeeID uuid.UUID, params domain.CreateTimeEntryParams) (*domain.TimeEntry, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	return s.createTimeEntry(ctx, params, "TimeEntryService.CreateTimeEntryByAdmin")
}

func (s *TimeEntryService) DecideTimeEntryByAdmin(ctx context.Context, adminEmployeeID, timeEntryID uuid.UUID, params domain.DecideTimeEntryParams) (*domain.TimeEntry, error) {
	if adminEmployeeID == uuid.Nil || timeEntryID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	rejectionReason := trimTimeEntryStringPtr(params.RejectionReason)
	if decision == "reject" && rejectionReason == nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	var updated *domain.TimeEntry
	err := s.repository.WithTx(ctx, func(tx domain.TimeEntryTxRepository) error {
		current, err := tx.GetTimeEntryForUpdate(ctx, timeEntryID)
		if err != nil {
			return err
		}
		if current.Status != domain.TimeEntryStatusSubmitted {
			return domain.ErrTimeEntryStateInvalid
		}

		if decision == "approve" {
			updated, err = tx.ApproveTimeEntry(ctx, timeEntryID, adminEmployeeID)
			return err
		}

		updated, err = tx.RejectTimeEntry(ctx, timeEntryID, rejectionReason)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TimeEntryService) createTimeEntry(ctx context.Context, params domain.CreateTimeEntryParams, operation string) (*domain.TimeEntry, error) {
	normalizedParams, err := normalizeCreateTimeEntryParams(params)
	if err != nil {
		return nil, err
	}

	item, err := s.repository.CreateTimeEntry(ctx, normalizedParams)
	if err != nil {
		s.logError(ctx, operation, "failed to create time entry", err,
			zap.String("employee_id", normalizedParams.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}

	return item, nil
}

func (s *TimeEntryService) GetTimeEntryByID(ctx context.Context, timeEntryID uuid.UUID) (*domain.TimeEntry, error) {
	if timeEntryID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	item, err := s.repository.GetTimeEntryByID(ctx, timeEntryID)
	if err != nil {
		if err == domain.ErrTimeEntryNotFound {
			return nil, err
		}
		s.logError(ctx, "TimeEntryService.GetTimeEntryByID", "failed to get time entry", err,
			zap.String("time_entry_id", timeEntryID.String()),
		)
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}

	return item, nil
}

func (s *TimeEntryService) GetMyTimeEntryByID(ctx context.Context, actorEmployeeID, timeEntryID uuid.UUID) (*domain.TimeEntry, error) {
	if actorEmployeeID == uuid.Nil || timeEntryID == uuid.Nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	item, err := s.GetTimeEntryByID(ctx, timeEntryID)
	if err != nil {
		return nil, err
	}
	if item.EmployeeID != actorEmployeeID {
		return nil, domain.ErrTimeEntryForbidden
	}

	return item, nil
}

func (s *TimeEntryService) ListTimeEntries(ctx context.Context, params domain.ListTimeEntriesParams) (*domain.TimeEntryPage, error) {
	normalizedParams, err := normalizeListTimeEntriesParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListTimeEntries(ctx, normalizedParams)
	if err != nil {
		s.logError(ctx, "TimeEntryService.ListTimeEntries", "failed to list time entries", err)
		return nil, fmt.Errorf("failed to list time entries: %w", err)
	}

	return page, nil
}

func (s *TimeEntryService) ListMyTimeEntries(ctx context.Context, params domain.ListMyTimeEntriesParams) (*domain.TimeEntryPage, error) {
	normalizedParams, err := normalizeListMyTimeEntriesParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListMyTimeEntries(ctx, normalizedParams)
	if err != nil {
		s.logError(ctx, "TimeEntryService.ListMyTimeEntries", "failed to list my time entries", err,
			zap.String("employee_id", normalizedParams.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to list my time entries: %w", err)
	}

	return page, nil
}

func normalizeCreateTimeEntryParams(params domain.CreateTimeEntryParams) (domain.CreateTimeEntryParams, error) {
	if params.EmployeeID == uuid.Nil {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	if params.EntryDate.IsZero() {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	if params.Hours <= 0 || params.Hours > 24 {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	if params.BreakMinutes < 0 {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	if float64(params.BreakMinutes) >= params.Hours*60 {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	normalizedHourType := strings.ToLower(strings.TrimSpace(params.HourType))
	if !isValidTimeEntryHourType(normalizedHourType) {
		return domain.CreateTimeEntryParams{}, domain.ErrTimeEntryInvalidRequest
	}
	params.HourType = normalizedHourType
	return params, nil
}

func normalizeListTimeEntriesParams(params domain.ListTimeEntriesParams) (domain.ListTimeEntriesParams, error) {
	if params.Status == nil {
		return params, nil
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(*params.Status))
	if !isValidTimeEntryStatus(normalizedStatus) {
		return domain.ListTimeEntriesParams{}, domain.ErrTimeEntryInvalidRequest
	}

	params.Status = &normalizedStatus
	return params, nil
}

func normalizeListMyTimeEntriesParams(params domain.ListMyTimeEntriesParams) (domain.ListMyTimeEntriesParams, error) {
	if params.EmployeeID == uuid.Nil {
		return domain.ListMyTimeEntriesParams{}, domain.ErrTimeEntryInvalidRequest
	}
	if params.Status == nil {
		return params, nil
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(*params.Status))
	if !isValidTimeEntryStatus(normalizedStatus) {
		return domain.ListMyTimeEntriesParams{}, domain.ErrTimeEntryInvalidRequest
	}

	params.Status = &normalizedStatus
	return params, nil
}

func isValidTimeEntryHourType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.TimeEntryHourTypeNormal,
		domain.TimeEntryHourTypeOvertime,
		domain.TimeEntryHourTypeTravel,
		domain.TimeEntryHourTypeLeave,
		domain.TimeEntryHourTypeSick,
		domain.TimeEntryHourTypeTraining:
		return true
	default:
		return false
	}
}

func isValidTimeEntryStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.TimeEntryStatusDraft,
		domain.TimeEntryStatusSubmitted,
		domain.TimeEntryStatusApproved,
		domain.TimeEntryStatusRejected:
		return true
	default:
		return false
	}
}

func (s *TimeEntryService) logError(ctx context.Context, operation, message string, err error, fields ...zap.Field) {
	if s.logger == nil || err == nil {
		return
	}
	s.logger.LogError(ctx, operation, message, err, fields...)
}

func trimTimeEntryStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
