package service

import (
	"testing"

	"hrbackend/internal/domain"
)

func TestHasAnyScheduledShift(t *testing.T) {
	t.Run("returns false when all days are empty", func(t *testing.T) {
		rows := []domain.GetSchedulesByLocationInRangeResponse{
			{Date: "2026-01-05", Shifts: []domain.Shift{}},
			{Date: "2026-01-06", Shifts: nil},
		}

		if hasAnyScheduledShift(rows) {
			t.Fatalf("expected false for empty week")
		}
	})

	t.Run("returns true when at least one day has shifts", func(t *testing.T) {
		rows := []domain.GetSchedulesByLocationInRangeResponse{
			{Date: "2026-01-05", Shifts: []domain.Shift{}},
			{Date: "2026-01-06", Shifts: []domain.Shift{{}}},
		}

		if !hasAnyScheduledShift(rows) {
			t.Fatalf("expected true when a shift exists")
		}
	})
}
