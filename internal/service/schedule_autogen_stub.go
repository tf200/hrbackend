//go:build !ortools

package service

import (
	"context"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func (s *ScheduleService) AutoGenerateSchedules(
	ctx context.Context,
	req *domain.AutoGenerateSchedulesRequest,
) (*domain.AutoGenerateSchedulesResponse, error) {
	if err := s.validateAutoGenerateRequest(req); err != nil {
		return nil, err
	}
	return nil, domain.ErrScheduleAutogenUnavailable
}

func (s *ScheduleService) SaveGeneratedSchedules(
	ctx context.Context,
	creatorID uuid.UUID,
	req *domain.SaveGeneratedSchedulesRequest,
) error {
	if len(req.Slots) == 0 {
		return nil
	}
	return domain.ErrScheduleAutogenUnavailable
}
