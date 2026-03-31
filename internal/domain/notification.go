package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Notification Types
	TypeNewAppointment          = "new_appointment"
	TypeNewScheduleNotification = "new_schedule_notification"
	TypeSystemReminder          = "system_reminder"
)

type NotificationPayload struct {
	RecipientUserIDs []uuid.UUID
	Type             string
	Data             NotificationData
	CreatedAt        time.Time
	Message          string
}

type NotificationData struct {
	NewScheduleNotification *NewScheduleNotificationData
}

type NewScheduleNotificationData struct {
	ScheduleID uuid.UUID
	CreatedBy  uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	Location   string
}
