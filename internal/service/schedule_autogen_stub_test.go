package service

import (
	"context"
	"errors"
	"testing"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestAutoGenerateSchedulesUnavailableWithoutORTools(t *testing.T) {
	service := &ScheduleService{}

	_, err := service.AutoGenerateSchedules(context.Background(), &domain.AutoGenerateSchedulesRequest{
		LocationID:  uuid.New(),
		Week:        1,
		Year:        2026,
		EmployeeIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, domain.ErrScheduleAutogenUnavailable) {
		t.Fatalf("expected ErrScheduleAutogenUnavailable, got %v", err)
	}
}

func TestSaveGeneratedSchedulesUnavailableWithoutORTools(t *testing.T) {
	service := &ScheduleService{}

	err := service.SaveGeneratedSchedules(context.Background(), uuid.New(), &domain.SaveGeneratedSchedulesRequest{
		PlanID:     uuid.New(),
		LocationID: uuid.New(),
		Week:       1,
		Year:       2026,
		Slots: []domain.SchedulePlanSlot{
			{Date: "2026-01-01", ShiftID: uuid.New(), EmployeeIDs: []uuid.UUID{uuid.New()}},
		},
	})
	if !errors.Is(err, domain.ErrScheduleAutogenUnavailable) {
		t.Fatalf("expected ErrScheduleAutogenUnavailable, got %v", err)
	}
}
