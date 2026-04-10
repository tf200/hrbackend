package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type createShiftSwapRequest struct {
	RecipientEmployeeID uuid.UUID  `json:"recipient_employee_id" binding:"required"`
	RequesterScheduleID uuid.UUID  `json:"requester_schedule_id" binding:"required"`
	RecipientScheduleID uuid.UUID  `json:"recipient_schedule_id" binding:"required"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
}

type createAdminShiftSwapRequest struct {
	RequesterEmployeeID uuid.UUID `json:"requester_employee_id" binding:"required"`
	RecipientEmployeeID uuid.UUID `json:"recipient_employee_id" binding:"required"`
	RequesterScheduleID uuid.UUID `json:"requester_schedule_id" binding:"required"`
	RecipientScheduleID uuid.UUID `json:"recipient_schedule_id" binding:"required"`
	Note                *string   `json:"note,omitempty"`
}

type respondShiftSwapRequest struct {
	Decision string  `json:"decision"       binding:"required,oneof=accept reject"`
	Note     *string `json:"note,omitempty"`
}

type adminDecisionShiftSwapRequest struct {
	Decision string  `json:"decision"       binding:"required,oneof=approve reject"`
	Note     *string `json:"note,omitempty"`
}

type listShiftSwapRequestsRequest struct {
	httpapi.PageRequest
	Status     *string    `form:"status"      binding:"omitempty,oneof=pending_recipient recipient_rejected pending_admin admin_rejected confirmed cancelled expired"`
	Filter     *string    `form:"filter"      binding:"omitempty,oneof=open to_approve history"`
	EmployeeID *uuid.UUID `form:"employee_id"`
}

type createShiftSwapResponse struct {
	ID                  uuid.UUID  `json:"id"`
	RequesterEmployeeID uuid.UUID  `json:"requester_employee_id"`
	RecipientEmployeeID uuid.UUID  `json:"recipient_employee_id"`
	RequesterScheduleID uuid.UUID  `json:"requester_schedule_id"`
	RecipientScheduleID uuid.UUID  `json:"recipient_schedule_id"`
	Status              string     `json:"status"`
	RequestedAt         time.Time  `json:"requested_at"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	Direction           string     `json:"direction,omitempty"`
}

type shiftSwapScheduleSnapshot struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	ShiftName     string    `json:"shift_name"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
}

type shiftSwapResponse struct {
	ID                    uuid.UUID                 `json:"id"`
	RequesterEmployeeID   uuid.UUID                 `json:"requester_employee_id"`
	RequesterEmployeeName string                    `json:"requester_employee_name"`
	RecipientEmployeeID   uuid.UUID                 `json:"recipient_employee_id"`
	RecipientEmployeeName string                    `json:"recipient_employee_name"`
	RequesterSchedule     shiftSwapScheduleSnapshot `json:"requester_schedule"`
	RecipientSchedule     shiftSwapScheduleSnapshot `json:"recipient_schedule"`
	Status                string                    `json:"status"`
	RequestedAt           time.Time                 `json:"requested_at"`
	RecipientRespondedAt  *time.Time                `json:"recipient_responded_at,omitempty"`
	AdminDecidedAt        *time.Time                `json:"admin_decided_at,omitempty"`
	RecipientResponseNote *string                   `json:"recipient_response_note,omitempty"`
	AdminDecisionNote     *string                   `json:"admin_decision_note,omitempty"`
	AdminEmployeeID       *uuid.UUID                `json:"admin_employee_id,omitempty"`
	AdminEmployeeName     *string                   `json:"admin_employee_name,omitempty"`
	ExpiresAt             *time.Time                `json:"expires_at,omitempty"`
	Direction             string                    `json:"direction,omitempty"`
}

func toCreateShiftSwapParams(req createShiftSwapRequest) domain.CreateShiftSwapRequest {
	return domain.CreateShiftSwapRequest{
		RecipientEmployeeID: req.RecipientEmployeeID,
		RequesterScheduleID: req.RequesterScheduleID,
		RecipientScheduleID: req.RecipientScheduleID,
		ExpiresAt:           req.ExpiresAt,
	}
}

func toRespondShiftSwapParams(req respondShiftSwapRequest) domain.RespondShiftSwapRequest {
	return domain.RespondShiftSwapRequest{
		Decision: req.Decision,
		Note:     req.Note,
	}
}

func toAdminDecisionShiftSwapParams(
	req adminDecisionShiftSwapRequest,
) domain.AdminDecisionShiftSwapRequest {
	return domain.AdminDecisionShiftSwapRequest{
		Decision: req.Decision,
		Note:     req.Note,
	}
}

func toListShiftSwapParams(req listShiftSwapRequestsRequest) domain.ListShiftSwapRequestsParams {
	return domain.ListShiftSwapRequestsParams{
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		Status:     req.Status,
		Filter:     req.Filter,
		EmployeeID: req.EmployeeID,
	}
}

func toCreateShiftSwapResponse(item *domain.CreateShiftSwapResponse) createShiftSwapResponse {
	return createShiftSwapResponse{
		ID:                  item.ID,
		RequesterEmployeeID: item.RequesterEmployeeID,
		RecipientEmployeeID: item.RecipientEmployeeID,
		RequesterScheduleID: item.RequesterScheduleID,
		RecipientScheduleID: item.RecipientScheduleID,
		Status:              item.Status,
		RequestedAt:         item.RequestedAt,
		ExpiresAt:           item.ExpiresAt,
		Direction:           item.Direction,
	}
}

func toShiftSwapResponse(item domain.ShiftSwapResponse) shiftSwapResponse {
	requesterEmployeeName := item.RequesterEmployeeName
	if requesterEmployeeName == "" {
		requesterEmployeeName = item.RequesterSchedule.EmployeeName
	}

	recipientEmployeeName := item.RecipientEmployeeName
	if recipientEmployeeName == "" {
		recipientEmployeeName = item.RecipientSchedule.EmployeeName
	}

	return shiftSwapResponse{
		ID:                    item.ID,
		RequesterEmployeeID:   item.RequesterEmployeeID,
		RequesterEmployeeName: requesterEmployeeName,
		RecipientEmployeeID:   item.RecipientEmployeeID,
		RecipientEmployeeName: recipientEmployeeName,
		RequesterSchedule: shiftSwapScheduleSnapshot{
			ID:            item.RequesterSchedule.ID,
			EmployeeID:    item.RequesterSchedule.EmployeeID,
			EmployeeName:  item.RequesterSchedule.EmployeeName,
			ShiftName:     item.RequesterSchedule.ShiftName,
			StartDatetime: item.RequesterSchedule.StartDatetime,
			EndDatetime:   item.RequesterSchedule.EndDatetime,
		},
		RecipientSchedule: shiftSwapScheduleSnapshot{
			ID:            item.RecipientSchedule.ID,
			EmployeeID:    item.RecipientSchedule.EmployeeID,
			EmployeeName:  item.RecipientSchedule.EmployeeName,
			ShiftName:     item.RecipientSchedule.ShiftName,
			StartDatetime: item.RecipientSchedule.StartDatetime,
			EndDatetime:   item.RecipientSchedule.EndDatetime,
		},
		Status:                item.Status,
		RequestedAt:           item.RequestedAt,
		RecipientRespondedAt:  item.RecipientRespondedAt,
		AdminDecidedAt:        item.AdminDecidedAt,
		RecipientResponseNote: item.RecipientResponseNote,
		AdminDecisionNote:     item.AdminDecisionNote,
		AdminEmployeeID:       item.AdminEmployeeID,
		AdminEmployeeName:     item.AdminEmployeeName,
		ExpiresAt:             item.ExpiresAt,
		Direction:             item.Direction,
	}
}
