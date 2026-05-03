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

func NewAdminDashboardRepository(store *db.Store) domain.AdminDashboardRepository {
	return &AdminDashboardRepository{store: store}
}

func (r *AdminDashboardRepository) GetAvailabilityStats(
	ctx context.Context,
) (*domain.AvailabilityStats, error) {
	row, err := r.store.GetEmployeeAvailabilityStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.AvailabilityStats{
		Active:       row.Active,
		OnLeave:      row.OnLeave,
		Sick:         row.Sick,
		OutOfService: row.OutOfService,
		Unscheduled:  row.Unscheduled,
	}, nil
}

func (r *AdminDashboardRepository) GetCriticalActionStats(
	ctx context.Context,
) (*domain.CriticalActionStats, error) {
	row, err := r.store.GetCriticalActionStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.CriticalActionStats{
		PendingLeaveRequests: row.PendingLeaveRequests,
		PendingShiftSwaps:    row.PendingShiftSwaps,
		PendingExpenseClaims: row.PendingExpenseClaims,
		PendingTimeEntries:   row.PendingTimeEntries,
	}, nil
}

func (r *AdminDashboardRepository) GetPayrollTotalStats(
	ctx context.Context,
) (*domain.PayrollTotalStats, error) {
	row, err := r.store.GetPayrollTotalStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.PayrollTotalStats{
		MonthStart:  conv.TimeFromPgDate(row.MonthStart),
		SalaryTotal: row.SalaryTotal,
		ZZPTotal:    row.ZzpTotal,
		ORTTotal:    row.OrtTotal,
	}, nil
}

func (r *AdminDashboardRepository) GetRiskRadarStats(
	ctx context.Context,
) (*domain.RiskRadarStats, error) {
	row, err := r.store.GetRiskRadarStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.RiskRadarStats{
		MonthStart:               conv.TimeFromPgDate(row.MonthStart),
		ContractsEndingThisMonth: row.ContractsEndingThisMonth,
		OverdueTraining:          row.OverdueTraining,
		LateArrivalsThisMonth:    row.LateArrivalsThisMonth,
	}, nil
}

func (r *AdminDashboardRepository) ListTeamHealthByDepartment(
	ctx context.Context,
) ([]domain.TeamHealthByDepartmentItem, error) {
	rows, err := r.store.ListTeamHealthByDepartment(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.TeamHealthByDepartmentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.TeamHealthByDepartmentItem{
			DepartmentID:    row.DepartmentID,
			DepartmentName:  row.DepartmentName,
			StaffingPercent: row.StaffingPercent,
			AbsencePercent:  row.AbsencePercent,
			TrainingPercent: row.TrainingPercent,
			Score:           row.Score,
			Risk:            row.Risk,
		})
	}

	return items, nil
}

func (r *AdminDashboardRepository) ListOpenShiftCoverage(
	ctx context.Context,
	days int32,
) ([]domain.OpenShiftCoverageItem, error) {
	rows, err := r.store.ListOpenShiftCoverage(ctx, days)
	if err != nil {
		return nil, err
	}

	items := make([]domain.OpenShiftCoverageItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.OpenShiftCoverageItem{
			LocationID:          row.LocationID,
			LocationName:        row.LocationName,
			Street:              row.Street,
			HouseNumber:         row.HouseNumber,
			HouseNumberAddition: row.HouseNumberAddition,
			PostalCode:          row.PostalCode,
			City:                row.City,
			ShiftID:             row.ShiftID,
			ShiftName:           row.ShiftName,
			ShiftDate:           conv.TimeFromPgDate(row.ShiftDate),
			OpenSlots:           row.OpenSlots,
		})
	}

	return items, nil
}
