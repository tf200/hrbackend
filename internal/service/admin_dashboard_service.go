package service

import (
	"context"
	"fmt"
	"time"

	"hrbackend/internal/domain"
)

const (
	adminDashboardViewLast6Months = "last_6_months"
	adminDashboardViewYearly      = "yearly"
	defaultDashboardAlertDays     = 30
	defaultDashboardAlertLimit    = 6
	maxDashboardAlertLimit        = 20
	maxRecentEmployees            = 6
)

type AdminDashboardService struct {
	repo   domain.AdminDashboardRepository
	logger domain.Logger
}

func NewAdminDashboardService(
	repo domain.AdminDashboardRepository,
	logger domain.Logger,
) *AdminDashboardService {
	return &AdminDashboardService{repo: repo, logger: logger}
}

func (s *AdminDashboardService) GetKPIs(ctx context.Context) (*domain.AdminDashboardKPI, error) {
	kpis, err := s.repo.GetKPIs(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to get dashboard KPIs", err)
		return nil, err
	}
	return kpis, nil
}

func (s *AdminDashboardService) ListRecentEmployees(
	ctx context.Context,
	params domain.ListRecentEmployeesParams,
) (*domain.RecentDashboardEmployeePage, error) {
	if params.Limit <= 0 || params.Limit > maxRecentEmployees {
		params.Limit = maxRecentEmployees
	}

	page, err := s.repo.ListRecentEmployees(ctx, params)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to list recent employees", err)
		return nil, err
	}
	return page, nil
}

func (s *AdminDashboardService) GetFullTimeEmployeeBreakdowns(ctx context.Context) (*domain.FullTimeEmployeeBreakdowns, error) {
	breakdowns, err := s.repo.GetFullTimeEmployeeBreakdowns(ctx)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to get full-time employee breakdowns", err)
		return nil, err
	}
	return breakdowns, nil
}

func (s *AdminDashboardService) GetLeaveAbsenceTrends(
	ctx context.Context,
	params domain.GetLeaveAbsenceTrendsParams,
) (*domain.LeaveAbsenceTrends, error) {
	view := params.View
	if view == "" {
		view = adminDashboardViewYearly
	}

	now := time.Now()
	var year *int32
	var fromDate, toDate time.Time
	switch view {
	case adminDashboardViewYearly:
		y := int32(now.Year())
		if params.Year != nil {
			y = *params.Year
		}
		if y < 1 {
			return nil, fmt.Errorf("%w: year must be greater than 0", domain.ErrAdminDashboardInvalidRequest)
		}
		year = &y
		fromDate = time.Date(int(y), time.January, 1, 0, 0, 0, 0, time.UTC)
		toDate = time.Date(int(y), time.December, 1, 0, 0, 0, 0, time.UTC)
	case adminDashboardViewLast6Months:
		toDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		fromDate = toDate.AddDate(0, -5, 0)
	default:
		return nil, fmt.Errorf("%w: unsupported view", domain.ErrAdminDashboardInvalidRequest)
	}

	points, err := s.repo.ListLeaveAbsenceTrendPoints(ctx, domain.ListLeaveAbsenceTrendPointsParams{
		FromDate: fromDate,
		ToDate:   toDate,
	})
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to get leave absence trends", err)
		return nil, err
	}

	return &domain.LeaveAbsenceTrends{
		View:   view,
		Year:   year,
		Points: points,
	}, nil
}

func (s *AdminDashboardService) GetUpcomingDashboardAlerts(
	ctx context.Context,
	params domain.GetUpcomingDashboardAlertsParams,
) (*domain.UpcomingDashboardAlerts, error) {
	if params.Days == 0 {
		params.Days = defaultDashboardAlertDays
	}
	if params.Days < 0 {
		return nil, fmt.Errorf("%w: days must be greater than 0", domain.ErrAdminDashboardInvalidRequest)
	}

	if params.Limit <= 0 {
		params.Limit = defaultDashboardAlertLimit
	}
	if params.Limit > maxDashboardAlertLimit {
		params.Limit = maxDashboardAlertLimit
	}

	now := time.Now()
	toDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(params.Days))
	listParams := domain.ListUpcomingDashboardAlertsParams{
		ToDate: toDate,
		Limit:  params.Limit,
	}

	endingContracts, err := s.repo.ListEndingContractAlerts(ctx, listParams)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to list ending contract alerts", err)
		return nil, err
	}

	expiringCredentials, err := s.repo.ListExpiringCredentialAlerts(ctx, listParams)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to list expiring credential alerts", err)
		return nil, err
	}

	returningFromLeave, err := s.repo.ListReturningFromLeaveAlerts(ctx, listParams)
	if err != nil {
		s.logger.LogError(ctx, "AdminDashboardService", "failed to list returning from leave alerts", err)
		return nil, err
	}

	return &domain.UpcomingDashboardAlerts{
		EndingContracts:     endingContracts,
		ExpiringCredentials: expiringCredentials,
		ReturningFromLeave:  returningFromLeave,
	}, nil
}
