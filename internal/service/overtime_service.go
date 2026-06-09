package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OvertimeService struct {
	repository          domain.OvertimeRepository
	employeeRepo        domain.EmployeeRepository
	notificationService domain.NotificationService
	logger              domain.Logger
}

func NewOvertimeService(
	repository domain.OvertimeRepository,
	employeeRepo domain.EmployeeRepository,
	notificationService domain.NotificationService,
	logger domain.Logger,
) domain.OvertimeService {
	return &OvertimeService{
		repository:          repository,
		employeeRepo:        employeeRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (s *OvertimeService) CreateOvertimeEntry(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	params.EmployeeID = actorEmployeeID
	entry, err := s.createOvertimeEntry(ctx, params, "OvertimeService.CreateOvertimeEntry")
	if err != nil {
		return nil, err
	}

	// Trigger notification to administrators
	if s.notificationService != nil {
		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				Roles: []string{"admin"},
			},
			Message: fmt.Sprintf(
				"%s has requested %d minutes of overtime for the shift on %s.",
				entry.EmployeeName,
				entry.Minutes,
				entry.EntryDate.Format("2006-01-02"),
			),
			Data: domain.OvertimeRequestCreatedNotificationData{
				OvertimeEntryID: entry.ID,
				EmployeeID:      entry.EmployeeID,
				EmployeeName:    entry.EmployeeName,
				Minutes:         entry.Minutes,
				EntryDate:       entry.EntryDate,
				Reason:          entry.Reason,
			},
		})
	}

	return entry, nil
}

func (s *OvertimeService) CreateOvertimeEntryByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.CreateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	return s.createOvertimeEntry(ctx, params, "OvertimeService.CreateOvertimeEntryByAdmin")
}

func (s *OvertimeService) DecideOvertimeEntryByAdmin(
	ctx context.Context,
	adminEmployeeID, overtimeEntryID uuid.UUID,
	params domain.DecideOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if adminEmployeeID == uuid.Nil || overtimeEntryID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	rejectionReason := trimOvertimeStringPtr(params.RejectionReason)
	if decision == "reject" && rejectionReason == nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	var updated *domain.OvertimeEntry
	err := s.repository.WithTx(ctx, func(tx domain.OvertimeTxRepository) error {
		current, err := tx.GetOvertimeEntryForUpdate(ctx, overtimeEntryID)
		if err != nil {
			return err
		}
		if current.Status != domain.OvertimeStatusSubmitted {
			return domain.ErrOvertimeStateInvalid
		}

		if decision == "approve" {
			updated, err = tx.ApproveOvertimeEntry(ctx, overtimeEntryID, adminEmployeeID)
			return err
		}

		updated, err = tx.RejectOvertimeEntry(ctx, overtimeEntryID, rejectionReason)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Trigger notification to the employee who made the request
	if s.notificationService != nil && updated.EmployeeID != uuid.Nil {
		decidedByName := "An administrator"
		if s.employeeRepo != nil {
			emp, err := s.employeeRepo.GetEmployeeByID(ctx, adminEmployeeID)
			if err == nil && emp != nil {
				decidedByName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
			}
		}

		rejectionReasonStr := ""
		if updated.RejectionReason != nil {
			rejectionReasonStr = *updated.RejectionReason
		}

		var message string
		if updated.Status == "approved" {
			message = fmt.Sprintf(
				"Your overtime request for %d minutes on %s has been approved by %s.",
				updated.Minutes,
				updated.EntryDate.Format("2006-01-02"),
				decidedByName,
			)
		} else if updated.Status == "rejected" {
			if rejectionReasonStr != "" {
				message = fmt.Sprintf("Your overtime request for %d minutes on %s has been rejected by %s. Reason: %s", updated.Minutes, updated.EntryDate.Format("2006-01-02"), decidedByName, rejectionReasonStr)
			} else {
				message = fmt.Sprintf("Your overtime request for %d minutes on %s has been rejected by %s.", updated.Minutes, updated.EntryDate.Format("2006-01-02"), decidedByName)
			}
		}

		s.notificationService.Notify(ctx, domain.NotificationRequest{
			Recipients: domain.NotificationRecipients{
				EmployeeIDs: []uuid.UUID{updated.EmployeeID},
			},
			Message: message,
			Data: domain.OvertimeRequestDecidedNotificationData{
				OvertimeEntryID:     updated.ID,
				EmployeeID:          updated.EmployeeID,
				Status:              updated.Status,
				Minutes:             updated.Minutes,
				EntryDate:           updated.EntryDate,
				DecidedByEmployeeID: adminEmployeeID,
				DecidedByName:       decidedByName,
				RejectionReason:     rejectionReasonStr,
			},
		})
	}

	return updated, nil
}

func (s *OvertimeService) UpdateOvertimeEntryByAdmin(
	ctx context.Context,
	adminEmployeeID, overtimeEntryID uuid.UUID,
	params domain.UpdateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if adminEmployeeID == uuid.Nil || overtimeEntryID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	var updated *domain.OvertimeEntry
	err := s.repository.WithTx(ctx, func(tx domain.OvertimeTxRepository) error {
		current, err := tx.GetOvertimeEntryForUpdate(ctx, overtimeEntryID)
		if err != nil {
			return err
		}
		if current.PaidPeriodID != nil {
			return domain.ErrOvertimeStateInvalid
		}

		params.EmployeeID = current.EmployeeID
		normalized, err := normalizeUpdateOvertimeEntryParams(params)
		if err != nil {
			return err
		}
		updated, err = tx.UpdateOvertimeEntryByAdmin(ctx, overtimeEntryID, normalized)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *OvertimeService) UpdateMyOvertimeEntry(
	ctx context.Context,
	actorEmployeeID, overtimeEntryID uuid.UUID,
	params domain.UpdateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if actorEmployeeID == uuid.Nil || overtimeEntryID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	var updated *domain.OvertimeEntry
	err := s.repository.WithTx(ctx, func(tx domain.OvertimeTxRepository) error {
		current, err := tx.GetOvertimeEntryForUpdate(ctx, overtimeEntryID)
		if err != nil {
			return err
		}
		if current.EmployeeID != actorEmployeeID {
			return domain.ErrOvertimeForbidden
		}
		if current.Status != domain.OvertimeStatusSubmitted || current.PaidPeriodID != nil {
			return domain.ErrOvertimeStateInvalid
		}

		params.EmployeeID = current.EmployeeID
		normalized, err := normalizeUpdateOvertimeEntryParams(params)
		if err != nil {
			return err
		}
		updated, err = tx.UpdateOvertimeEntryByAdmin(ctx, overtimeEntryID, normalized)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *OvertimeService) createOvertimeEntry(
	ctx context.Context,
	params domain.CreateOvertimeEntryParams,
	operation string,
) (*domain.OvertimeEntry, error) {
	normalizedParams, err := normalizeCreateOvertimeEntryParams(params)
	if err != nil {
		return nil, err
	}

	item, err := s.repository.CreateOvertimeEntry(ctx, normalizedParams)
	if err != nil {
		s.logError(ctx, operation, "failed to create overtime entry", err,
			zap.String("employee_id", normalizedParams.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to create overtime entry: %w", err)
	}

	return item, nil
}

func (s *OvertimeService) GetOvertimeEntryByID(
	ctx context.Context,
	overtimeEntryID uuid.UUID,
) (*domain.OvertimeEntry, error) {
	if overtimeEntryID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	item, err := s.repository.GetOvertimeEntryByID(ctx, overtimeEntryID)
	if err != nil {
		if err == domain.ErrOvertimeNotFound {
			return nil, err
		}
		s.logError(ctx, "OvertimeService.GetOvertimeEntryByID", "failed to get overtime entry", err,
			zap.String("overtime_entry_id", overtimeEntryID.String()),
		)
		return nil, fmt.Errorf("failed to get overtime entry: %w", err)
	}

	return item, nil
}

func (s *OvertimeService) GetMyOvertimeEntryByID(
	ctx context.Context,
	actorEmployeeID, overtimeEntryID uuid.UUID,
) (*domain.OvertimeEntry, error) {
	if actorEmployeeID == uuid.Nil || overtimeEntryID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	item, err := s.GetOvertimeEntryByID(ctx, overtimeEntryID)
	if err != nil {
		return nil, err
	}
	if item.EmployeeID != actorEmployeeID {
		return nil, domain.ErrOvertimeForbidden
	}

	return item, nil
}

func (s *OvertimeService) ListOvertimeEntries(
	ctx context.Context,
	params domain.ListOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	normalizedParams, err := normalizeListOvertimeEntriesParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListOvertimeEntries(ctx, normalizedParams)
	if err != nil {
		s.logError(
			ctx,
			"OvertimeService.ListOvertimeEntries",
			"failed to list overtime entries",
			err,
		)
		return nil, fmt.Errorf("failed to list overtime entries: %w", err)
	}

	return page, nil
}

func (s *OvertimeService) ListMyOvertimeEntries(
	ctx context.Context,
	params domain.ListMyOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	normalizedParams, err := normalizeListMyOvertimeEntriesParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListMyOvertimeEntries(ctx, normalizedParams)
	if err != nil {
		s.logError(
			ctx,
			"OvertimeService.ListMyOvertimeEntries",
			"failed to list my overtime entries",
			err,
			zap.String("employee_id", normalizedParams.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to list my overtime entries: %w", err)
	}

	return page, nil
}

func (s *OvertimeService) GetCurrentMonthOvertimeStats(
	ctx context.Context,
) (*domain.OvertimeStats, error) {
	stats, err := s.repository.GetCurrentMonthOvertimeStats(ctx)
	if err != nil {
		s.logError(
			ctx,
			"OvertimeService.GetCurrentMonthOvertimeStats",
			"failed to get current month overtime stats",
			err,
		)
		return nil, fmt.Errorf("failed to get current month overtime stats: %w", err)
	}

	return stats, nil
}

func (s *OvertimeService) GetMyCurrentMonthOvertimeStats(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.OvertimeStats, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrOvertimeInvalidRequest
	}

	stats, err := s.repository.GetMyCurrentMonthOvertimeStats(ctx, employeeID)
	if err != nil {
		s.logError(
			ctx,
			"OvertimeService.GetMyCurrentMonthOvertimeStats",
			"failed to get my current month overtime stats",
			err,
			zap.String("employee_id", employeeID.String()),
		)
		return nil, fmt.Errorf("failed to get my current month overtime stats: %w", err)
	}

	return stats, nil
}

func normalizeCreateOvertimeEntryParams(
	params domain.CreateOvertimeEntryParams,
) (domain.CreateOvertimeEntryParams, error) {
	if params.EmployeeID == uuid.Nil {
		return domain.CreateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}
	if params.EntryDate.IsZero() {
		return domain.CreateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}
	if params.Minutes <= 0 {
		return domain.CreateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}

	normalizedReason := strings.ToLower(strings.TrimSpace(params.Reason))
	if !isValidOvertimeReason(normalizedReason) {
		return domain.CreateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}
	params.Reason = normalizedReason

	return params, nil
}

func normalizeUpdateOvertimeEntryParams(
	params domain.UpdateOvertimeEntryParams,
) (domain.UpdateOvertimeEntryParams, error) {
	normalized := domain.UpdateOvertimeEntryParams{
		EmployeeID: params.EmployeeID,
	}
	var hasUpdates bool

	if params.ScheduleID != nil {
		normalized.ScheduleID = params.ScheduleID
		hasUpdates = true
	}
	if params.EntryDate != nil {
		dateOnly := params.EntryDate.UTC()
		dateOnly = time.Date(
			dateOnly.Year(),
			dateOnly.Month(),
			dateOnly.Day(),
			0,
			0,
			0,
			0,
			time.UTC,
		)
		normalized.EntryDate = &dateOnly
		hasUpdates = true
	}
	if params.Minutes != nil {
		normalized.Minutes = params.Minutes
		hasUpdates = true
	}
	if params.Reason != nil {
		normalizedReason := strings.ToLower(strings.TrimSpace(*params.Reason))
		normalized.Reason = &normalizedReason
		hasUpdates = true
	}
	if params.Description != nil {
		normalized.Description = trimOvertimeStringPtr(params.Description)
		hasUpdates = true
	}

	if !hasUpdates {
		return domain.UpdateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}

	if normalized.Minutes != nil && *normalized.Minutes <= 0 {
		return domain.UpdateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}
	if normalized.Reason != nil && !isValidOvertimeReason(*normalized.Reason) {
		return domain.UpdateOvertimeEntryParams{}, domain.ErrOvertimeInvalidRequest
	}

	return normalized, nil
}

func normalizeListOvertimeEntriesParams(
	params domain.ListOvertimeEntriesParams,
) (domain.ListOvertimeEntriesParams, error) {
	if params.Status == nil {
		return params, nil
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(*params.Status))
	if !isValidOvertimeStatus(normalizedStatus) {
		return domain.ListOvertimeEntriesParams{}, domain.ErrOvertimeInvalidRequest
	}

	params.Status = &normalizedStatus
	return params, nil
}

func normalizeListMyOvertimeEntriesParams(
	params domain.ListMyOvertimeEntriesParams,
) (domain.ListMyOvertimeEntriesParams, error) {
	if params.EmployeeID == uuid.Nil {
		return domain.ListMyOvertimeEntriesParams{}, domain.ErrOvertimeInvalidRequest
	}
	if params.Status == nil {
		return params, nil
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(*params.Status))
	if !isValidOvertimeStatus(normalizedStatus) {
		return domain.ListMyOvertimeEntriesParams{}, domain.ErrOvertimeInvalidRequest
	}

	params.Status = &normalizedStatus
	return params, nil
}

func isValidOvertimeReason(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.OvertimeReasonClientCrisis,
		domain.OvertimeReasonUnderstaffing,
		domain.OvertimeReasonMeetingConsultation,
		domain.OvertimeReasonTrainingEducation,
		domain.OvertimeReasonCompletingAdministration,
		domain.OvertimeReasonHandover,
		domain.OvertimeReasonEmergency,
		domain.OvertimeReasonProjectWork,
		domain.OvertimeReasonEventActivity,
		domain.OvertimeReasonOther:
		return true
	default:
		return false
	}
}

func isValidOvertimeStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.OvertimeStatusSubmitted,
		domain.OvertimeStatusApproved,
		domain.OvertimeStatusRejected:
		return true
	default:
		return false
	}
}

func (s *OvertimeService) logError(
	ctx context.Context,
	operation, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger == nil || err == nil {
		return
	}
	s.logger.LogError(ctx, operation, message, err, fields...)
}

func trimOvertimeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
