package service

import (
	"context"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

const (
	shiftSwapDecisionAccept  = "accept"
	shiftSwapDecisionReject  = "reject"
	shiftSwapDecisionApprove = "approve"
)

func (s *ScheduleService) CreateShiftSwapRequest(
	ctx context.Context,
	requesterEmployeeID uuid.UUID,
	req *domain.CreateShiftSwapRequest,
) (*domain.CreateShiftSwapResponse, error) {
	if req == nil || requesterEmployeeID == uuid.Nil {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	if req.RequesterScheduleID == uuid.Nil || req.RecipientScheduleID == uuid.Nil ||
		req.RecipientEmployeeID == uuid.Nil {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	if requesterEmployeeID == req.RecipientEmployeeID ||
		req.RequesterScheduleID == req.RecipientScheduleID {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now().UTC()) {
		return nil, domain.ErrShiftSwapInvalidRequest
	}

	requesterSchedule, err := s.repository.GetScheduleForSwapValidation(
		ctx,
		req.RequesterScheduleID,
	)
	if err != nil {
		return nil, domain.ErrScheduleNotFound
	}
	recipientSchedule, err := s.repository.GetScheduleForSwapValidation(
		ctx,
		req.RecipientScheduleID,
	)
	if err != nil {
		return nil, domain.ErrScheduleNotFound
	}

	now := time.Now().UTC()
	if requesterSchedule.StartDatetime.Before(now) || recipientSchedule.StartDatetime.Before(now) {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	if requesterSchedule.EmployeeID != requesterEmployeeID ||
		recipientSchedule.EmployeeID != req.RecipientEmployeeID {
		return nil, domain.ErrShiftSwapScheduleOwnership
	}

	_ = s.repository.ExpirePendingShiftSwapRequests(ctx)

	created, err := s.repository.CreateShiftSwapRequest(ctx, *req, requesterEmployeeID)
	if err != nil {
		return nil, err
	}

	resp := &domain.CreateShiftSwapResponse{
		ID:                  created.ID,
		RequesterEmployeeID: created.RequesterEmployeeID,
		RecipientEmployeeID: created.RecipientEmployeeID,
		RequesterScheduleID: created.RequesterScheduleID,
		RecipientScheduleID: created.RecipientScheduleID,
		Status:              created.Status,
		RequestedAt:         created.RequestedAt,
		ExpiresAt:           created.ExpiresAt,
	}
	if requesterEmployeeID == created.RequesterEmployeeID {
		resp.Direction = "sent"
	} else if requesterEmployeeID == created.RecipientEmployeeID {
		resp.Direction = "received"
	}

	return resp, nil
}

func (s *ScheduleService) RespondToShiftSwapRequest(
	ctx context.Context,
	recipientEmployeeID, swapID uuid.UUID,
	req *domain.RespondShiftSwapRequest,
) (*domain.ShiftSwapResponse, error) {
	if req == nil || recipientEmployeeID == uuid.Nil || swapID == uuid.Nil {
		return nil, domain.ErrShiftSwapInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(req.Decision))
	nextStatus := "recipient_rejected"
	if decision == shiftSwapDecisionAccept {
		nextStatus = "pending_admin"
	} else if decision != shiftSwapDecisionReject {
		return nil, domain.ErrShiftSwapInvalidRequest
	}

	updated, err := s.repository.UpdateShiftSwapStatusAfterRecipientResponse(
		ctx,
		swapID,
		recipientEmployeeID,
		nextStatus,
		req.Note,
	)
	if err != nil {
		existing, getErr := s.repository.GetShiftSwapRequestByID(ctx, swapID)
		if getErr != nil || existing == nil {
			return nil, domain.ErrShiftSwapNotFound
		}
		if existing.Status == "expired" ||
			(existing.ExpiresAt != nil && !existing.ExpiresAt.After(time.Now().UTC())) {
			return nil, domain.ErrShiftSwapExpired
		}
		return nil, domain.ErrShiftSwapStateInvalid
	}

	details, err := s.repository.GetShiftSwapRequestDetailsByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	if recipientEmployeeID == details.RequesterEmployeeID {
		details.Direction = "sent"
	} else if recipientEmployeeID == details.RecipientEmployeeID {
		details.Direction = "received"
	}
	return details, nil
}

func (s *ScheduleService) AdminDecisionShiftSwapRequest(
	ctx context.Context,
	adminEmployeeID, swapID uuid.UUID,
	req *domain.AdminDecisionShiftSwapRequest,
) (*domain.ShiftSwapResponse, error) {
	if req == nil || adminEmployeeID == uuid.Nil || swapID == uuid.Nil {
		return nil, domain.ErrShiftSwapInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(req.Decision))
	switch decision {
	case shiftSwapDecisionApprove:
		var confirmed *domain.ShiftSwapRequestRecord
		err := s.repository.WithTx(ctx, func(tx domain.ScheduleRepository) error {
			swapRow, err := tx.LockShiftSwapRequestForAdminDecision(ctx, swapID)
			if err != nil {
				return domain.ErrShiftSwapNotFound
			}
			if swapRow.Status == "expired" {
				return domain.ErrShiftSwapExpired
			}
			if swapRow.Status != "pending_admin" {
				return domain.ErrShiftSwapStateInvalid
			}
			if swapRow.ExpiresAt != nil && !swapRow.ExpiresAt.After(time.Now().UTC()) {
				return domain.ErrShiftSwapExpired
			}

			schedules, err := tx.LockSchedulesByIDsForSwap(
				ctx,
				[]uuid.UUID{swapRow.RequesterScheduleID, swapRow.RecipientScheduleID},
			)
			if err != nil {
				return err
			}
			if len(schedules) != 2 {
				return domain.ErrScheduleNotFound
			}

			scheduleByID := map[uuid.UUID]domain.ScheduleSwapValidation{}
			for _, sched := range schedules {
				scheduleByID[sched.ID] = sched
			}
			requesterSchedule, okReq := scheduleByID[swapRow.RequesterScheduleID]
			recipientSchedule, okRec := scheduleByID[swapRow.RecipientScheduleID]
			if !okReq || !okRec {
				return domain.ErrScheduleNotFound
			}
			if requesterSchedule.EmployeeID != swapRow.RequesterEmployeeID ||
				recipientSchedule.EmployeeID != swapRow.RecipientEmployeeID {
				return domain.ErrShiftSwapScheduleOwnership
			}
			now := time.Now().UTC()
			if requesterSchedule.StartDatetime.Before(now) ||
				recipientSchedule.StartDatetime.Before(now) {
				return domain.ErrShiftSwapInvalidRequest
			}

			excludeIDs := []uuid.UUID{requesterSchedule.ID, recipientSchedule.ID}
			requesterOverlap, err := tx.CountScheduleOverlapsForEmployee(
				ctx,
				swapRow.RequesterEmployeeID,
				excludeIDs,
				recipientSchedule.StartDatetime,
				recipientSchedule.EndDatetime,
			)
			if err != nil {
				return err
			}
			if requesterOverlap > 0 {
				return domain.ErrShiftSwapConflict
			}

			recipientOverlap, err := tx.CountScheduleOverlapsForEmployee(
				ctx,
				swapRow.RecipientEmployeeID,
				excludeIDs,
				requesterSchedule.StartDatetime,
				requesterSchedule.EndDatetime,
			)
			if err != nil {
				return err
			}
			if recipientOverlap > 0 {
				return domain.ErrShiftSwapConflict
			}

			if err := tx.UpdateScheduleEmployeeAssignment(
				ctx,
				requesterSchedule.ID,
				swapRow.RecipientEmployeeID,
			); err != nil {
				return err
			}
			if err := tx.UpdateScheduleEmployeeAssignment(
				ctx,
				recipientSchedule.ID,
				swapRow.RequesterEmployeeID,
			); err != nil {
				return err
			}

			var note *string
			if req.Note != nil && strings.TrimSpace(*req.Note) != "" {
				trimmed := strings.TrimSpace(*req.Note)
				note = &trimmed
			}
			confirmed, err = tx.MarkShiftSwapConfirmed(ctx, swapID, note, adminEmployeeID)
			if err != nil {
				return domain.ErrShiftSwapStateInvalid
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		details, err := s.repository.GetShiftSwapRequestDetailsByID(ctx, confirmed.ID)
		if err != nil {
			return nil, err
		}
		return details, nil

	case shiftSwapDecisionReject:
		updated, err := s.repository.UpdateShiftSwapAdminDecision(
			ctx,
			swapID,
			"admin_rejected",
			req.Note,
			adminEmployeeID,
		)
		if err != nil {
			existing, getErr := s.repository.GetShiftSwapRequestByID(ctx, swapID)
			if getErr != nil || existing == nil {
				return nil, domain.ErrShiftSwapNotFound
			}
			if existing.Status == "expired" ||
				(existing.ExpiresAt != nil && !existing.ExpiresAt.After(time.Now().UTC())) {
				return nil, domain.ErrShiftSwapExpired
			}
			return nil, domain.ErrShiftSwapStateInvalid
		}

		details, err := s.repository.GetShiftSwapRequestDetailsByID(ctx, updated.ID)
		if err != nil {
			return nil, err
		}
		return details, nil
	default:
		return nil, domain.ErrShiftSwapInvalidRequest
	}
}

func (s *ScheduleService) ListMyShiftSwapRequests(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.ShiftSwapResponse, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	return s.repository.ListMyShiftSwapRequests(ctx, employeeID)
}

func (s *ScheduleService) ListShiftSwapRequests(
	ctx context.Context,
	params domain.ListShiftSwapRequestsParams,
) (*domain.ShiftSwapPage, error) {
	if params.Status != nil {
		if !isValidShiftSwapStatus(*params.Status) {
			return nil, domain.ErrShiftSwapInvalidRequest
		}
	}
	if params.Filter != nil {
		if !isValidShiftSwapFilter(*params.Filter) {
			return nil, domain.ErrShiftSwapInvalidRequest
		}
	}
	return s.repository.ListShiftSwapRequests(ctx, params)
}

func isValidShiftSwapStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending_recipient",
		"recipient_rejected",
		"pending_admin",
		"admin_rejected",
		"confirmed",
		"cancelled",
		"expired":
		return true
	default:
		return false
	}
}

func isValidShiftSwapFilter(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open", "to_approve", "history":
		return true
	default:
		return false
	}
}
