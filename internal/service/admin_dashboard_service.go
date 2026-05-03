package service

import (
	"context"

	"hrbackend/internal/domain"
)

type AdminDashboardService struct {
	repo   domain.AdminDashboardRepository
	logger domain.Logger
}

func NewAdminDashboardService(
	repo domain.AdminDashboardRepository,
	logger domain.Logger,
) domain.AdminDashboardService {
	return &AdminDashboardService{repo: repo, logger: logger}
}

func (s *AdminDashboardService) GetAvailabilityStats(
	ctx context.Context,
) (*domain.AvailabilityStats, error) {
	stats, err := s.repo.GetAvailabilityStats(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.GetAvailabilityStats", "failed", err)
		return nil, err
	}
	return stats, nil
}

func (s *AdminDashboardService) GetCriticalActionStats(
	ctx context.Context,
) (*domain.CriticalActionStats, error) {
	stats, err := s.repo.GetCriticalActionStats(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.GetCriticalActionStats", "failed", err)
		return nil, err
	}
	return stats, nil
}

func (s *AdminDashboardService) GetPayrollTotalStats(
	ctx context.Context,
) (*domain.PayrollTotalStats, error) {
	stats, err := s.repo.GetPayrollTotalStats(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.GetPayrollTotalStats", "failed", err)
		return nil, err
	}
	return stats, nil
}

func (s *AdminDashboardService) GetRiskRadarStats(
	ctx context.Context,
) (*domain.RiskRadarStats, error) {
	stats, err := s.repo.GetRiskRadarStats(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.GetRiskRadarStats", "failed", err)
		return nil, err
	}
	return stats, nil
}

func (s *AdminDashboardService) ListTeamHealthByDepartment(
	ctx context.Context,
) ([]domain.TeamHealthByDepartmentItem, error) {
	items, err := s.repo.ListTeamHealthByDepartment(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.ListTeamHealthByDepartment", "failed", err)
		return nil, err
	}
	return items, nil
}

func (s *AdminDashboardService) ListOpenShiftCoverage(
	ctx context.Context,
	days int32,
) ([]domain.OpenShiftCoverageItem, error) {
	items, err := s.repo.ListOpenShiftCoverage(ctx, days)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService.ListOpenShiftCoverage", "failed", err)
		return nil, err
	}
	return items, nil
}

// Ensure compile-time compliance.
var _ domain.AdminDashboardService = (*AdminDashboardService)(nil)
