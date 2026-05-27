package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TypeNewAppointment          = "new_appointment"
	TypeNewScheduleNotification = "new_schedule_notification"
	TypeSystemReminder          = "system_reminder"
	TypeShiftSwapRequested      = "shift_swap_requested"
)

type Notification struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Type      string             `json:"type"`
	Message   string             `json:"message"`
	IsRead    bool               `json:"is_read"`
	Data      NotificationData   `json:"data"`
	ReadAt    *time.Time         `json:"read_at"`
	CreatedAt time.Time          `json:"created_at"`
}

type NotificationRecipients struct {
	UserIDs     []uuid.UUID `json:"user_ids,omitempty"`
	EmployeeIDs []uuid.UUID `json:"employee_ids,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
}

type NotificationRequest struct {
	Recipients NotificationRecipients `json:"recipients"`
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Data       NotificationData       `json:"data"`
	CreatedAt  *time.Time             `json:"created_at,omitempty"`
}

type NotificationData struct {
	NewScheduleNotification *NewScheduleNotificationData `json:"new_schedule_notification,omitempty"`
	ShiftSwapNotification   *ShiftSwapNotificationData   `json:"shift_swap_notification,omitempty"`
}

type ShiftSwapNotificationData struct {
	SwapID                uuid.UUID `json:"swap_id"`
	RequesterEmployeeID   uuid.UUID `json:"requester_employee_id"`
	RequesterEmployeeName string    `json:"requester_employee_name"`
	RecipientEmployeeID   uuid.UUID `json:"recipient_employee_id"`
	RecipientEmployeeName string    `json:"recipient_employee_name"`
	Status                string    `json:"status"`
}

type NewScheduleNotificationData struct {
	ScheduleID uuid.UUID `json:"schedule_id"`
	CreatedBy  uuid.UUID `json:"created_by"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Location   string    `json:"location"`
}

type CreateNotificationsParams struct {
	UserIDs   []uuid.UUID
	Type      string
	Message   string
	Data      []byte
	CreatedAt time.Time
}

type NotificationRepository interface {
	CreateNotifications(ctx context.Context, params CreateNotificationsParams) ([]Notification, error)
	ListUserIDsByEmployeeIDs(ctx context.Context, employeeIDs []uuid.UUID) ([]uuid.UUID, error)
	ListUserIDsByRoles(ctx context.Context, roleNames []string) ([]uuid.UUID, error)
	ListUserIDsByPermissions(ctx context.Context, permissionNames []string) ([]uuid.UUID, error)
}

type RealtimeSender interface {
	SendToUser(userID uuid.UUID, message []byte)
}

type NotificationService interface {
	Notify(ctx context.Context, req NotificationRequest)
}
