package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
)

type AdminDashboardRepository struct {
	store *db.Store
}

func NewAdminDashboardRepository(store *db.Store) *AdminDashboardRepository {
	return &AdminDashboardRepository{store: store}
}

func (r *AdminDashboardRepository) GetKPIs(ctx context.Context) (*domain.AdminDashboardKPI, error) {
	row, err := r.store.GetAdminDashboardKPIs(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.AdminDashboardKPI{
		TotalEmployees:   row.TotalEmployees,
		EmployeesPresent: row.EmployeesPresent,
		TotalDocuments:   row.TotalDocuments,
		ProcessingDocs:   row.ProcessingDocs,
	}, nil
}

func (r *AdminDashboardRepository) ListRecentEmployees(
	ctx context.Context,
	params domain.ListRecentEmployeesParams,
) (*domain.RecentDashboardEmployeePage, error) {
	rows, err := r.store.ListRecentDashboardEmployees(ctx, db.ListRecentDashboardEmployeesParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.store.CountRecentDashboardEmployees(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.RecentDashboardEmployee, len(rows))
	for i, row := range rows {
		items[i] = domain.RecentDashboardEmployee{
			ID:                 row.ID,
			FirstName:          row.FirstName,
			LastName:           row.LastName,
			OrganizationalRole: row.OrganizationalRoleName,
			DepartmentName:     row.DepartmentName,
			LocationName:       row.LocationName,
			CreatedAt:          row.CreatedAt.Time,
		}
	}

	return &domain.RecentDashboardEmployeePage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (r *AdminDashboardRepository) GetFullTimeEmployeeBreakdowns(
	ctx context.Context,
) (*domain.FullTimeEmployeeBreakdowns, error) {
	deptRows, err := r.store.ListFullTimeEmployeesByDepartment(ctx)
	if err != nil {
		return nil, err
	}

	locRows, err := r.store.ListFullTimeEmployeesByLocation(ctx)
	if err != nil {
		return nil, err
	}

	deptItems := make([]domain.FullTimeEmployeeDeptBreakdownItem, len(deptRows))
	for i, row := range deptRows {
		deptItems[i] = domain.FullTimeEmployeeDeptBreakdownItem{
			DepartmentID:   row.DepartmentID,
			DepartmentName: row.DepartmentName,
			TotalEmployees: row.TotalEmployees,
		}
	}

	locItems := make([]domain.FullTimeEmployeeLocBreakdownItem, len(locRows))
	for i, row := range locRows {
		locItems[i] = domain.FullTimeEmployeeLocBreakdownItem{
			LocationID:     row.LocationID,
			LocationName:   row.LocationName,
			TotalEmployees: row.TotalEmployees,
		}
	}

	return &domain.FullTimeEmployeeBreakdowns{
		ByDepartment: deptItems,
		ByLocation:   locItems,
	}, nil
}

func (r *AdminDashboardRepository) ListLeaveAbsenceTrendPoints(
	ctx context.Context,
	params domain.ListLeaveAbsenceTrendPointsParams,
) ([]domain.LeaveAbsenceTrendPoint, error) {
	rows, err := r.store.ListLeaveAbsenceTrendPoints(ctx, db.ListLeaveAbsenceTrendPointsParams{
		FromDate: conv.PgDateFromTime(params.FromDate),
		ToDate:   conv.PgDateFromTime(params.ToDate),
	})
	if err != nil {
		return nil, err
	}

	points := make([]domain.LeaveAbsenceTrendPoint, len(rows))
	for i, row := range rows {
		points[i] = domain.LeaveAbsenceTrendPoint{
			Month:        conv.TimeFromPgDate(row.MonthStart),
			EmployeesOut: row.EmployeesOut,
		}
	}

	return points, nil
}

func (r *AdminDashboardRepository) ListEndingContractAlerts(
	ctx context.Context,
	params domain.ListUpcomingDashboardAlertsParams,
) ([]domain.EndingContractAlert, error) {
	rows, err := r.store.ListEndingContractAlerts(ctx, db.ListEndingContractAlertsParams{
		ToDate: conv.PgDateFromTime(params.ToDate),
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, err
	}

	alerts := make([]domain.EndingContractAlert, len(rows))
	for i, row := range rows {
		alerts[i] = domain.EndingContractAlert{
			EmployeeID:      row.EmployeeID,
			EmployeeName:    row.EmployeeName,
			ContractID:      row.ContractID,
			ContractType:    row.ContractType,
			ContractEndDate: conv.TimeFromPgDate(row.ContractEndDate),
			DaysRemaining:   row.DaysRemaining,
			Department:      row.Department,
			Location:        row.Location,
		}
	}

	return alerts, nil
}

func (r *AdminDashboardRepository) ListExpiringCredentialAlerts(
	ctx context.Context,
	params domain.ListUpcomingDashboardAlertsParams,
) ([]domain.ExpiringCredentialAlert, error) {
	rows, err := r.store.ListExpiringCredentialAlerts(ctx, db.ListExpiringCredentialAlertsParams{
		ToDate: conv.PgDateFromTime(params.ToDate),
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, err
	}

	alerts := make([]domain.ExpiringCredentialAlert, len(rows))
	for i, row := range rows {
		alerts[i] = domain.ExpiringCredentialAlert{
			EmployeeID:     row.EmployeeID,
			EmployeeName:   row.EmployeeName,
			CredentialID:   row.CredentialID,
			CredentialType: row.CredentialType,
			Name:           row.Name,
			ExpiryDate:     conv.TimeFromPgDate(row.ExpiryDate),
			DaysRemaining:  row.DaysRemaining,
		}
	}

	return alerts, nil
}

func (r *AdminDashboardRepository) ListReturningFromLeaveAlerts(
	ctx context.Context,
	params domain.ListUpcomingDashboardAlertsParams,
) ([]domain.ReturningFromLeaveAlert, error) {
	rows, err := r.store.ListReturningFromLeaveAlerts(ctx, db.ListReturningFromLeaveAlertsParams{
		ToDate: conv.PgDateFromTime(params.ToDate),
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, err
	}

	alerts := make([]domain.ReturningFromLeaveAlert, len(rows))
	for i, row := range rows {
		alerts[i] = domain.ReturningFromLeaveAlert{
			EmployeeID:      row.EmployeeID,
			EmployeeName:    row.EmployeeName,
			LeaveRequestID:  row.LeaveRequestID,
			LeaveType:       row.LeaveType,
			LeaveEndDate:    conv.TimeFromPgDate(row.LeaveEndDate),
			ReturnDate:      conv.TimeFromPgDate(row.ReturnDate),
			DaysUntilReturn: row.DaysUntilReturn,
		}
	}

	return alerts, nil
}
