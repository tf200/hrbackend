package repository

import (
	"context"
	"errors"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/jackc/pgx/v5"
)

type EmployeeDashboardRepository struct {
	store *db.Store
}

func NewEmployeeDashboardRepository(store *db.Store) *EmployeeDashboardRepository {
	return &EmployeeDashboardRepository{store: store}
}

func (r *EmployeeDashboardRepository) GetKPIs(
	ctx context.Context,
	params domain.GetEmployeeDashboardKPIsParams,
) (*domain.EmployeeDashboardRepositoryKPI, error) {
	row, err := r.store.GetEmployeeDashboardKPIs(ctx, db.GetEmployeeDashboardKPIsParams{
		EmployeeID: params.EmployeeID,
		Year:       params.Year,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return &domain.EmployeeDashboardRepositoryKPI{
		LeaveBalance: domain.EmployeeDashboardLeaveBalanceKPI{
			Year:         row.Year,
			UsedMinutes:  row.LeaveUsedMinutes,
			TotalMinutes: row.LeaveTotalMinutes,
		},
		PendingLeaveRequests: row.PendingLeaveRequests,
		PendingSignatures:    row.PendingSignatures,
	}, nil
}

func (r *EmployeeDashboardRepository) ListPendingRequests(
	ctx context.Context,
	params domain.ListEmployeeDashboardPendingRequestsParams,
) ([]domain.EmployeeDashboardPendingRequest, error) {
	rows, err := r.store.ListEmployeeDashboardPendingRequests(
		ctx,
		db.ListEmployeeDashboardPendingRequestsParams{
			EmployeeID: params.EmployeeID,
			Since:      conv.PgTimestamptzFromTime(params.Since),
			Limit:      params.Limit,
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.EmployeeDashboardPendingRequest, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainEmployeeDashboardPendingRequest(row))
	}
	return items, nil
}

func toDomainEmployeeDashboardPendingRequest(
	row db.ListEmployeeDashboardPendingRequestsRow,
) domain.EmployeeDashboardPendingRequest {
	var durationMinutes *int32
	if row.RequestType != "expense" {
		durationMinutes = &row.DurationMinutes
	}

	return domain.EmployeeDashboardPendingRequest{
		ID:              row.ID,
		RequestType:     row.RequestType,
		Status:          row.Status,
		SubmittedAt:     conv.TimeFromPgTimestamptz(row.SubmittedAt),
		RequestDate:     conv.TimeFromPgDate(row.RequestDate),
		Title:           row.Title,
		Description:     row.Description,
		DurationMinutes: durationMinutes,
		Amount:          row.Amount,
		Currency:        row.Currency,
	}
}
