package service

import (
	"context"
	"fmt"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultEmployeeDashboardPendingRequestDays  = 90
	maxEmployeeDashboardPendingRequestDays      = 180
	defaultEmployeeDashboardPendingRequestLimit = 10
	maxEmployeeDashboardPendingRequestLimit     = 20
)

type EmployeeDashboardService struct {
	repo          domain.EmployeeDashboardRepository
	salaryService employeeDashboardSalaryReader
	logger        domain.Logger
}

type employeeDashboardSalaryReader interface {
	GetMySalaryPage(
		ctx context.Context,
		employeeID uuid.UUID,
		periodStart, periodEnd time.Time,
	) (*domain.SalaryPageData, error)
}

func NewEmployeeDashboardService(
	repo domain.EmployeeDashboardRepository,
	salaryService employeeDashboardSalaryReader,
	logger domain.Logger,
) domain.EmployeeDashboardService {
	return &EmployeeDashboardService{
		repo:          repo,
		salaryService: salaryService,
		logger:        logger,
	}
}

func (s *EmployeeDashboardService) GetKPIs(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.EmployeeDashboardKPI, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrEmployeeDashboardInvalidRequest
	}

	now := time.Now().UTC()
	base, err := s.repo.GetKPIs(ctx, domain.GetEmployeeDashboardKPIsParams{
		EmployeeID: employeeID,
		Year:       int32(now.Year()),
	})
	if err != nil {
		s.logError(ctx, "GetKPIs", "failed to get dashboard KPI counts", err,
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}

	periodStart, periodEnd := domain.ResolvePayrollPeriod(now)
	if periodStart.IsZero() || periodEnd.IsZero() {
		return nil, domain.ErrEmployeeDashboardInvalidRequest
	}

	salaryPage, err := s.salaryService.GetMySalaryPage(ctx, employeeID, periodStart, periodEnd)
	if err != nil {
		s.logError(ctx, "GetKPIs", "failed to get current salary page", err,
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}

	return &domain.EmployeeDashboardKPI{
		LeaveBalance:         base.LeaveBalance,
		PendingLeaveRequests: base.PendingLeaveRequests,
		EstimatedCurrentPay: domain.EmployeeDashboardPayKPI{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			GrossAmount: employeeDashboardGrossAmount(salaryPage),
			DataSource:  salaryPage.DataSource,
		},
		PendingSignatures: base.PendingSignatures,
	}, nil
}

func (s *EmployeeDashboardService) ListPendingRequests(
	ctx context.Context,
	params domain.ListEmployeeDashboardPendingRequestsParams,
) ([]domain.EmployeeDashboardPendingRequest, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrEmployeeDashboardInvalidRequest
	}

	if params.RecentDays == 0 {
		params.RecentDays = defaultEmployeeDashboardPendingRequestDays
	}
	if params.RecentDays < 0 {
		return nil, fmt.Errorf(
			"%w: days must be greater than 0",
			domain.ErrEmployeeDashboardInvalidRequest,
		)
	}
	if params.RecentDays > maxEmployeeDashboardPendingRequestDays {
		params.RecentDays = maxEmployeeDashboardPendingRequestDays
	}

	if params.Limit <= 0 {
		params.Limit = defaultEmployeeDashboardPendingRequestLimit
	}
	if params.Limit > maxEmployeeDashboardPendingRequestLimit {
		params.Limit = maxEmployeeDashboardPendingRequestLimit
	}
	params.Since = time.Now().UTC().AddDate(0, 0, -int(params.RecentDays))

	items, err := s.repo.ListPendingRequests(ctx, params)
	if err != nil {
		s.logError(ctx, "ListPendingRequests", "failed to list pending requests", err,
			zap.String("employee_id", params.EmployeeID.String()),
		)
		return nil, err
	}
	return items, nil
}

func employeeDashboardGrossAmount(data *domain.SalaryPageData) float64 {
	if data == nil {
		return 0
	}
	if data.PayPeriod != nil {
		return data.PayPeriod.GrossAmount
	}
	if data.Preview != nil {
		return data.Preview.GrossAmount
	}
	return 0
}

func (s *EmployeeDashboardService) logError(
	ctx context.Context,
	method, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger != nil {
		s.logger.LogError(ctx, "EmployeeDashboardService."+method, message, err, fields...)
	}
}
