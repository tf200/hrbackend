package repository

import (
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func scheduleIDPtr(sourceType string, scheduleID uuid.UUID) *uuid.UUID {
	if sourceType == domain.PayrollSourceSchedule {
		return &scheduleID
	}
	return nil
}

func fullName(firstName, lastName string) string {
	return strings.TrimSpace(firstName + " " + lastName)
}

func nullableFullName(firstName, lastName *string) *string {
	if firstName == nil && lastName == nil {
		return nil
	}
	fn := ""
	ln := ""
	if firstName != nil {
		fn = *firstName
	}
	if lastName != nil {
		ln = *lastName
	}
	result := strings.TrimSpace(fn + " " + ln)
	if result == "" {
		return nil
	}
	return &result
}

func toNullablePgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}
