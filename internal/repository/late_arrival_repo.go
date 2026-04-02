package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type LateArrivalRepository struct {
	queries db.Querier
}

func NewLateArrivalRepository(queries db.Querier) domain.LateArrivalRepository {
	return &LateArrivalRepository{queries: queries}
}

func (r *LateArrivalRepository) ListAssignedSchedulesForEmployeeOnDate(
	ctx context.Context,
	employeeID uuid.UUID,
	arrivalDate time.Time,
) ([]domain.AssignedScheduleForDate, error) {
	rows, err := r.queries.ListAssignedSchedulesForEmployeeOnDate(ctx, db.ListAssignedSchedulesForEmployeeOnDateParams{
		EmployeeID:  employeeID,
		ArrivalDate: conv.PgDateFromTime(arrivalDate),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.AssignedScheduleForDate, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.AssignedScheduleForDate{
			ScheduleID:       row.ScheduleID,
			EmployeeID:       row.EmployeeID,
			StartDatetime:    conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndDatetime:      conv.TimeFromPgTimestamptz(row.EndDatetime),
			LocationTimezone: row.LocationTimezone,
			LocationName:     row.LocationName,
			ShiftName:        row.ShiftName,
		})
	}
	return items, nil
}

func (r *LateArrivalRepository) CreateLateArrival(
	ctx context.Context,
	params domain.LateArrivalCreateParams,
	scheduleID uuid.UUID,
) (*domain.LateArrival, error) {
	arrivalTime, err := conv.PgTimeFromString(strings.TrimSpace(params.ArrivalTime))
	if err != nil {
		return nil, domain.ErrLateArrivalInvalidRequest
	}

	row, err := r.queries.CreateLateArrival(ctx, db.CreateLateArrivalParams{
		ScheduleID:          scheduleID,
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: &params.CreatedByEmployeeID,
		ArrivalDate:         conv.PgDateFromTime(params.ArrivalDate),
		ArrivalTime:         arrivalTime,
		Reason:              params.Reason,
	})
	if err != nil {
		if isLateArrivalUniqueViolation(err) {
			return nil, domain.ErrLateArrivalConflict
		}
		return nil, err
	}

	item := toDomainLateArrival(row)
	return &item, nil
}

func (r *LateArrivalRepository) ListMyLateArrivals(
	ctx context.Context,
	params domain.ListMyLateArrivalsParams,
) (*domain.LateArrivalPage, error) {
	rows, err := r.queries.ListMyLateArrivalsPaginated(ctx, db.ListMyLateArrivalsPaginatedParams{
		EmployeeID: params.EmployeeID,
		DateFrom:   pgDateFromPtr(params.DateFrom),
		DateTo:     pgDateFromPtr(params.DateTo),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.LateArrivalPage{
		Items: make([]domain.LateArrivalListItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainLateArrivalListItem(
			row.ID,
			row.ScheduleID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.ArrivalDate,
			row.ArrivalTime,
			row.Reason,
			row.ShiftStartDatetime,
			row.ShiftEndDatetime,
			row.ShiftName,
			row.LocationName,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *LateArrivalRepository) ListLateArrivals(
	ctx context.Context,
	params domain.ListLateArrivalsParams,
) (*domain.LateArrivalPage, error) {
	var employeeSearch *string
	if params.EmployeeSearch != nil {
		trimmed := strings.TrimSpace(*params.EmployeeSearch)
		if trimmed != "" {
			employeeSearch = &trimmed
		}
	}

	rows, err := r.queries.ListLateArrivalsPaginated(ctx, db.ListLateArrivalsPaginatedParams{
		EmployeeSearch: employeeSearch,
		DateFrom:       pgDateFromPtr(params.DateFrom),
		DateTo:         pgDateFromPtr(params.DateTo),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.LateArrivalPage{
		Items: make([]domain.LateArrivalListItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainLateArrivalListItem(
			row.ID,
			row.ScheduleID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.ArrivalDate,
			row.ArrivalTime,
			row.Reason,
			row.ShiftStartDatetime,
			row.ShiftEndDatetime,
			row.ShiftName,
			row.LocationName,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func toDomainLateArrival(row db.LateArrival) domain.LateArrival {
	return domain.LateArrival{
		ID:                  row.ID,
		ScheduleID:          row.ScheduleID,
		EmployeeID:          row.EmployeeID,
		CreatedByEmployeeID: row.CreatedByEmployeeID,
		ArrivalDate:         conv.TimeFromPgDate(row.ArrivalDate),
		ArrivalTime:         conv.StringFromPgTime(row.ArrivalTime),
		Reason:              row.Reason,
		CreatedAt:           conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:           conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainLateArrivalListItem(
	id uuid.UUID,
	scheduleID uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	createdByEmployeeID *uuid.UUID,
	arrivalDate pgtype.Date,
	arrivalTime pgtype.Time,
	reason string,
	shiftStartDatetime pgtype.Timestamptz,
	shiftEndDatetime pgtype.Timestamptz,
	shiftName string,
	locationName string,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.LateArrivalListItem {
	return domain.LateArrivalListItem{
		LateArrival: domain.LateArrival{
			ID:                  id,
			ScheduleID:          scheduleID,
			EmployeeID:          employeeID,
			CreatedByEmployeeID: createdByEmployeeID,
			ArrivalDate:         conv.TimeFromPgDate(arrivalDate),
			ArrivalTime:         conv.StringFromPgTime(arrivalTime),
			Reason:              reason,
			CreatedAt:           conv.TimeFromPgTimestamptz(createdAt),
			UpdatedAt:           conv.TimeFromPgTimestamptz(updatedAt),
		},
		EmployeeName:       employeeName,
		ShiftStartDatetime: conv.TimeFromPgTimestamptz(shiftStartDatetime),
		ShiftEndDatetime:   conv.TimeFromPgTimestamptz(shiftEndDatetime),
		ShiftName:          shiftName,
		LocationName:       locationName,
	}
}

func isLateArrivalUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		strings.Contains(pgErr.ConstraintName, "late_arrivals_unique_schedule")
}

var _ domain.LateArrivalRepository = (*LateArrivalRepository)(nil)
