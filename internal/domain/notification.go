package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TypeNewAppointment          = "new_appointment"
	TypeNewScheduleNotification = "new_schedule_notification"
	TypeSystemReminder          = "system_reminder"
	TypeShiftSwapRequested      = "shift_swap_requested"
	TypeSignDocumentRequested   = "sign_document_requested"
	TypeSignDocumentSigned      = "sign_document_signed"
	TypeLeaveRequestCreated     = "leave_request_created"
	TypeLeaveRequestDecided     = "leave_request_decided"
	TypeOvertimeRequestCreated  = "overtime_request_created"
	TypeOvertimeRequestDecided  = "overtime_request_decided"
)

type SignDocumentRequestedNotificationData struct {
	DocumentID          uuid.UUID `json:"document_id"`
	DocumentTitle       string    `json:"document_title"`
	RequesterEmployeeID uuid.UUID `json:"requester_employee_id"`
	RequesterName       string    `json:"requester_name"`
}

func (SignDocumentRequestedNotificationData) NotificationType() string {
	return TypeSignDocumentRequested
}

type SignDocumentSignedNotificationData struct {
	DocumentID       uuid.UUID `json:"document_id"`
	DocumentTitle    string    `json:"document_title"`
	SignerEmployeeID uuid.UUID `json:"signer_employee_id"`
	SignerName       string    `json:"signer_name"`
	IsCompleted      bool      `json:"is_completed"`
}

func (SignDocumentSignedNotificationData) NotificationType() string {
	return TypeSignDocumentSigned
}

type LeaveRequestCreatedNotificationData struct {
	LeaveRequestID   uuid.UUID `json:"leave_request_id"`
	EmployeeID       uuid.UUID `json:"employee_id"`
	EmployeeName     string    `json:"employee_name"`
	LeaveType        string    `json:"leave_type"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	RequestedMinutes int32     `json:"requested_minutes"`
	Reason           string    `json:"reason"`
}

func (LeaveRequestCreatedNotificationData) NotificationType() string {
	return TypeLeaveRequestCreated
}

type LeaveRequestDecidedNotificationData struct {
	LeaveRequestID      uuid.UUID `json:"leave_request_id"`
	EmployeeID          uuid.UUID `json:"employee_id"`
	Status              string    `json:"status"`
	LeaveType           string    `json:"leave_type"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	DecidedByEmployeeID uuid.UUID `json:"decided_by_employee_id"`
	DecidedByName       string    `json:"decided_by_name"`
	DecisionNote        string    `json:"decision_note"`
}

func (LeaveRequestDecidedNotificationData) NotificationType() string {
	return TypeLeaveRequestDecided
}

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Message   string          `json:"message"`
	IsRead    bool            `json:"is_read"`
	Data      json.RawMessage `json:"data"`
	ReadAt    *time.Time      `json:"read_at"`
	CreatedAt time.Time       `json:"created_at"`
}

type NotificationRecipients struct {
	UserIDs     []uuid.UUID `json:"user_ids,omitempty"`
	EmployeeIDs []uuid.UUID `json:"employee_ids,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
}

type NotificationRequest struct {
	Recipients NotificationRecipients
	Message    string
	Data       NotificationData
	CreatedAt *time.Time
}

type NotificationData interface {
	NotificationType() string
}

type ShiftSwapNotificationData struct {
	SwapID                uuid.UUID `json:"swap_id"`
	RequesterEmployeeID   uuid.UUID `json:"requester_employee_id"`
	RequesterEmployeeName string    `json:"requester_employee_name"`
	RecipientEmployeeID   uuid.UUID `json:"recipient_employee_id"`
	RecipientEmployeeName string    `json:"recipient_employee_name"`
	Status                string    `json:"status"`
}

func (ShiftSwapNotificationData) NotificationType() string { return TypeShiftSwapRequested }

type NewScheduleNotificationData struct {
	ScheduleID uuid.UUID `json:"schedule_id"`
	CreatedBy  uuid.UUID `json:"created_by"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Location   string    `json:"location"`
}

func (NewScheduleNotificationData) NotificationType() string { return TypeNewScheduleNotification }

type OvertimeRequestCreatedNotificationData struct {
	OvertimeEntryID uuid.UUID `json:"overtime_entry_id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	Minutes         int32     `json:"minutes"`
	EntryDate       time.Time `json:"entry_date"`
	Reason          string    `json:"reason"`
}

func (OvertimeRequestCreatedNotificationData) NotificationType() string {
	return TypeOvertimeRequestCreated
}

type OvertimeRequestDecidedNotificationData struct {
	OvertimeEntryID     uuid.UUID `json:"overtime_entry_id"`
	EmployeeID          uuid.UUID `json:"employee_id"`
	Status              string    `json:"status"`
	Minutes             int32     `json:"minutes"`
	EntryDate           time.Time `json:"entry_date"`
	DecidedByEmployeeID uuid.UUID `json:"decided_by_employee_id"`
	DecidedByName       string    `json:"decided_by_name"`
	RejectionReason     string    `json:"rejection_reason,omitempty"`
}

func (OvertimeRequestDecidedNotificationData) NotificationType() string {
	return TypeOvertimeRequestDecided
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
	ListNotifications(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]Notification, error)
	CountNotifications(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkNotificationRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error
}

type RealtimeSender interface {
	SendToUser(userID uuid.UUID, message []byte)
}

type NotificationService interface {
	Notify(ctx context.Context, req NotificationRequest)
	ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int32) ([]Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}
