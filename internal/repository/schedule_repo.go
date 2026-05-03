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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ScheduleRepository struct {
	store *db.Store
}

func NewScheduleRepository(store *db.Store) domain.ScheduleRepository {
	return &ScheduleRepository{store: store}
}

func (r *ScheduleRepository) CreateSchedule(
	ctx context.Context,
	params domain.CreateScheduleParams,
) (*domain.CreateScheduleResponse, error) {
	row, err := r.store.CreateSchedule(ctx, db.CreateScheduleParams{
		EmployeeID:             params.EmployeeID,
		LocationID:             params.LocationID,
		LocationShiftID:        params.LocationShiftID,
		ShiftNameSnapshot:      params.ShiftNameSnapshot,
		ShiftStartTimeSnapshot: toPgTimePtr(params.ShiftStartTimeSnapshot),
		ShiftEndTimeSnapshot:   toPgTimePtr(params.ShiftEndTimeSnapshot),
		IsCustom:               params.IsCustom,
		CreatedByEmployeeID:    params.CreatedByEmployeeID,
		StartDatetime:          conv.PgTimestamptzFromTime(params.StartDatetime),
		EndDatetime:            conv.PgTimestamptzFromTime(params.EndDatetime),
	})
	if err != nil {
		return nil, err
	}

	return &domain.CreateScheduleResponse{
		ID:              row.ID,
		EmployeeID:      row.EmployeeID,
		LocationID:      row.LocationID,
		LocationName:    row.LocationName,
		StartDatetime:   conv.TimeFromPgTimestamptz(row.StartDatetime),
		EndDatetime:     conv.TimeFromPgTimestamptz(row.EndDatetime),
		CreatedAt:       conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       conv.TimeFromPgTimestamptz(row.UpdatedAt),
		LocationShiftID: row.LocationShiftID,
		ShiftName:       row.ShiftNameSnapshot,
	}, nil
}

func (r *ScheduleRepository) GetSchedulesByLocationInRange(
	ctx context.Context,
	locationID uuid.UUID,
	startDate, endDate time.Time,
) ([]domain.GetSchedulesByLocationInRangeResponse, error) {
	rows, err := r.store.GetSchedulesByLocationInRange(ctx, db.GetSchedulesByLocationInRangeParams{
		LocationID: locationID,
		StartDate:  conv.PgDateFromTime(startDate),
		EndDate:    conv.PgDateFromTime(endDate),
	})
	if err != nil {
		return nil, err
	}

	dayMap := make(map[string][]domain.Shift, len(rows))
	for _, row := range rows {
		day := conv.TimeFromPgDate(row.Day).Format("2006-01-02")
		dayMap[day] = append(dayMap[day], domain.Shift{
			ScheduleID:        row.ShiftID,
			EmployeeID:        row.EmployeeID,
			EmployeeFirstName: row.EmployeeFirstName,
			EmployeeLastName:  row.EmployeeLastName,
			StartTime:         conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndTime:           conv.TimeFromPgTimestamptz(row.EndDatetime),
			LocationID:        row.LocationID,
			ShiftName:         &row.ShiftName,
			LocationShiftID:   row.LocationShiftID,
			IsCustom:          row.IsCustom,
		})
	}

	response := make([]domain.GetSchedulesByLocationInRangeResponse, 0)
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		response = append(response, domain.GetSchedulesByLocationInRangeResponse{
			Date:   key,
			Shifts: append([]domain.Shift{}, dayMap[key]...),
		})
	}
	return response, nil
}

func (r *ScheduleRepository) GetEmployeeSchedulesByDay(
	ctx context.Context,
	employeeID uuid.UUID,
	date time.Time,
) ([]domain.EmployeeScheduleByDay, error) {
	rows, err := r.store.GetEmployeeSchedulesByDay(ctx, db.GetEmployeeSchedulesByDayParams{
		EmployeeID: employeeID,
		Date:       conv.PgDateFromTime(date),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeScheduleByDay, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeScheduleByDay{
			ScheduleID:      row.ScheduleID,
			EmployeeID:      row.EmployeeID,
			LocationID:      row.LocationID,
			LocationName:    row.LocationName,
			StartTime:       conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndTime:         conv.TimeFromPgTimestamptz(row.EndDatetime),
			ShiftName:       row.ShiftName,
			LocationShiftID: row.LocationShiftID,
			IsCustom:        row.IsCustom,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) GetEmployeeSchedulesTimelineEntries(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) ([]domain.EmployeeScheduleTimelineEntry, error) {
	rows, err := r.store.GetEmployeeSchedules(ctx, db.GetEmployeeSchedulesParams{
		EmployeeID:  employeeID,
		PeriodStart: conv.PgTimestamptzFromTime(periodStart),
		PeriodEnd:   conv.PgTimestamptzFromTime(periodEnd),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeScheduleTimelineEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeScheduleTimelineEntry{
			ScheduleID:   row.ID,
			StartTime:    conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndTime:      conv.TimeFromPgTimestamptz(row.EndDatetime),
			LocationID:   row.LocationID,
			LocationName: row.LocationName,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) GetEmployeeNextShift(
	ctx context.Context,
	employeeID uuid.UUID,
	now time.Time,
) (*domain.EmployeeShiftOverviewNextShift, error) {
	row, err := r.store.GetEmployeeNextShift(ctx, db.GetEmployeeNextShiftParams{
		EmployeeID: employeeID,
		Now:        conv.PgTimestamptzFromTime(now),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.EmployeeShiftOverviewNextShift{
		ScheduleID:   row.ScheduleID,
		ShiftName:    row.ShiftName,
		LocationID:   row.LocationID,
		LocationName: row.LocationName,
		Address:      formatLocationAddress(row.Street, row.HouseNumber, row.HouseNumberAddition, row.PostalCode, row.City),
		StartTime:    conv.TimeFromPgTimestamptz(row.StartDatetime),
		EndTime:      conv.TimeFromPgTimestamptz(row.EndDatetime),
		Date:         conv.TimeFromPgDate(row.ShiftDate).Format("2006-01-02"),
		IsCustom:     row.IsCustom,
		Colleagues:   []domain.EmployeeShiftOverviewColleague{},
	}, nil
}

func (r *ScheduleRepository) ListEmployeeShiftColleagues(
	ctx context.Context,
	scheduleID, employeeID uuid.UUID,
) ([]domain.EmployeeShiftOverviewColleague, error) {
	rows, err := r.store.ListEmployeeShiftColleagues(ctx, db.ListEmployeeShiftColleaguesParams{
		ScheduleID: scheduleID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeShiftOverviewColleague, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeShiftOverviewColleague{
			EmployeeID: row.EmployeeID,
			FirstName:  row.FirstName,
			LastName:   row.LastName,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) GetEmployeeShiftOverviewStats(
	ctx context.Context,
	employeeID uuid.UUID,
	now, weekStart, weekEnd, monthStart, monthEnd time.Time,
) (*domain.EmployeeShiftOverviewStats, error) {
	row, err := r.store.GetEmployeeShiftOverviewStats(ctx, db.GetEmployeeShiftOverviewStatsParams{
		MonthStart: conv.PgTimestamptzFromTime(monthStart),
		MonthEnd:   conv.PgTimestamptzFromTime(monthEnd),
		Now:        conv.PgTimestamptzFromTime(now),
		WeekStart:  conv.PgTimestamptzFromTime(weekStart),
		WeekEnd:    conv.PgTimestamptzFromTime(weekEnd),
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.EmployeeShiftOverviewStats{
		UpcomingCount:  row.UpcomingCount,
		CompletedCount: row.CompletedCount,
		PlannedHours:   row.PlannedHours,
	}, nil
}

func (r *ScheduleRepository) ListEmployeeWeekShiftCounts(
	ctx context.Context,
	employeeID uuid.UUID,
	weekStart, weekEnd time.Time,
) ([]domain.EmployeeShiftOverviewDayCount, error) {
	rows, err := r.store.ListEmployeeWeekShiftCounts(ctx, db.ListEmployeeWeekShiftCountsParams{
		EmployeeID: employeeID,
		WeekStart:  conv.PgTimestamptzFromTime(weekStart),
		WeekEnd:    conv.PgTimestamptzFromTime(weekEnd),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeShiftOverviewDayCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeShiftOverviewDayCount{
			Day:        conv.TimeFromPgDate(row.Day),
			ShiftCount: row.ShiftCount,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) ListEmployeeMonthShiftCounts(
	ctx context.Context,
	employeeID uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.EmployeeShiftOverviewMonthDayCount, error) {
	rows, err := r.store.ListEmployeeMonthShiftCounts(ctx, db.ListEmployeeMonthShiftCountsParams{
		EmployeeID: employeeID,
		MonthStart: conv.PgTimestamptzFromTime(monthStart),
		MonthEnd:   conv.PgTimestamptzFromTime(monthEnd),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeShiftOverviewMonthDayCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeShiftOverviewMonthDayCount{
			Day:        int(row.Day),
			ShiftCount: row.ShiftCount,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) GetEmployeeScheduleManager(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.EmployeeShiftOverviewManager, error) {
	row, err := r.store.GetEmployeeScheduleManager(ctx, employeeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.EmployeeShiftOverviewManager{
		FirstName: row.FirstName,
		LastName:  row.LastName,
	}, nil
}

func (r *ScheduleRepository) GetScheduleByID(
	ctx context.Context,
	scheduleID uuid.UUID,
) (*domain.GetScheduleByIdResponse, error) {
	row, err := r.store.GetScheduleById(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	return &domain.GetScheduleByIdResponse{
		ID:                row.ID,
		EmployeeID:        row.EmployeeID,
		EmployeeFirstName: row.EmployeeFirstName,
		EmployeeLastName:  row.EmployeeLastName,
		LocationID:        row.LocationID,
		LocationName:      row.LocationName,
		LocationShiftID:   row.LocationShiftID,
		LocationShiftName: &row.LocationShiftName,
		StartDatetime:     conv.TimeFromPgTimestamptz(row.StartDatetime),
		EndDatetime:       conv.TimeFromPgTimestamptz(row.EndDatetime),
		IsCustom:          row.IsCustom,
		CreatedAt:         conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *ScheduleRepository) UpdateSchedule(
	ctx context.Context,
	scheduleID uuid.UUID,
	params domain.UpdateScheduleParams,
) (*domain.UpdateScheduleResponse, error) {
	row, err := r.store.UpdateSchedule(ctx, db.UpdateScheduleParams{
		ID:                     scheduleID,
		EmployeeID:             params.EmployeeID,
		LocationID:             params.LocationID,
		LocationShiftID:        params.LocationShiftID,
		StartDatetime:          conv.PgTimestamptzFromTime(params.StartDatetime),
		EndDatetime:            conv.PgTimestamptzFromTime(params.EndDatetime),
		ShiftNameSnapshot:      params.ShiftNameSnapshot,
		ShiftStartTimeSnapshot: toPgTimePtr(params.ShiftStartTimeSnapshot),
		ShiftEndTimeSnapshot:   toPgTimePtr(params.ShiftEndTimeSnapshot),
		IsCustom:               params.IsCustom,
	})
	if err != nil {
		return nil, err
	}

	return &domain.UpdateScheduleResponse{
		ID:              row.ID,
		EmployeeID:      row.EmployeeID,
		LocationID:      row.LocationID,
		StartDatetime:   conv.TimeFromPgTimestamptz(row.StartDatetime),
		EndDatetime:     conv.TimeFromPgTimestamptz(row.EndDatetime),
		CreatedAt:       conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       conv.TimeFromPgTimestamptz(row.UpdatedAt),
		LocationName:    row.LocationName,
		LocationShiftID: row.LocationShiftID,
		ShiftName:       row.ShiftNameSnapshot,
	}, nil
}

func (r *ScheduleRepository) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	return r.store.DeleteSchedule(ctx, scheduleID)
}

func (r *ScheduleRepository) GetLocationByID(
	ctx context.Context,
	locationID uuid.UUID,
) (*domain.ScheduleLocation, error) {
	location, err := r.store.GetLocation(ctx, locationID)
	if err != nil {
		return nil, err
	}

	return &domain.ScheduleLocation{
		ID:       location.ID,
		Timezone: location.Timezone,
	}, nil
}

func (r *ScheduleRepository) GetShiftByID(
	ctx context.Context,
	shiftID uuid.UUID,
) (*domain.ScheduleLocationShift, error) {
	shift, err := r.store.GetShiftByID(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	return toDomainScheduleLocationShift(shift), nil
}

func (r *ScheduleRepository) GetShiftsByLocationID(
	ctx context.Context,
	locationID uuid.UUID,
) ([]domain.ScheduleLocationShift, error) {
	shifts, err := r.store.GetShiftsByLocationID(ctx, locationID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ScheduleLocationShift, len(shifts))
	for i, shift := range shifts {
		result[i] = *toDomainScheduleLocationShift(shift)
	}
	return result, nil
}

func (r *ScheduleRepository) ListEmployeesWithContractHours(
	ctx context.Context,
	employeeIDs []uuid.UUID,
) ([]domain.ScheduleEmployeeContractHours, error) {
	rows, err := r.store.ListEmployeesWithContractHours(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ScheduleEmployeeContractHours, len(rows))
	for i, row := range rows {
		result[i] = domain.ScheduleEmployeeContractHours{
			ID:            row.ID,
			FirstName:     row.FirstName,
			LastName:      row.LastName,
			ContractHours: row.ContractHours,
		}
	}
	return result, nil
}

func toDomainScheduleLocationShift(shift db.LocationShift) *domain.ScheduleLocationShift {
	return &domain.ScheduleLocationShift{
		ID:                shift.ID,
		LocationID:        shift.LocationID,
		ShiftName:         shift.ShiftName,
		StartMicroseconds: shift.StartTime.Microseconds,
		EndMicroseconds:   shift.EndTime.Microseconds,
	}
}

func toPgTimePtr(value *int64) pgtype.Time {
	if value == nil {
		return pgtype.Time{Valid: false}
	}
	return pgtype.Time{
		Microseconds: *value,
		Valid:        true,
	}
}

func (r *ScheduleRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.ScheduleRepository) error,
) error {
	if r.store == nil {
		return errors.New("schedule repository transaction store is not configured")
	}

	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		txRepo := &ScheduleRepository{
			store: &db.Store{
				Queries:  q,
				ConnPool: r.store.ConnPool,
			},
		}
		return fn(txRepo)
	})
}

func (r *ScheduleRepository) ExpirePendingShiftSwapRequests(ctx context.Context) error {
	return r.store.ExpirePendingShiftSwapRequests(ctx)
}

func (r *ScheduleRepository) GetScheduleForSwapValidation(
	ctx context.Context,
	scheduleID uuid.UUID,
) (*domain.ScheduleSwapValidation, error) {
	row, err := r.store.GetScheduleForSwapValidation(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	return &domain.ScheduleSwapValidation{
		ID:            row.ID,
		EmployeeID:    row.EmployeeID,
		LocationID:    row.LocationID,
		StartDatetime: conv.TimeFromPgTimestamptz(row.StartDatetime),
		EndDatetime:   conv.TimeFromPgTimestamptz(row.EndDatetime),
	}, nil
}

func (r *ScheduleRepository) CreateShiftSwapRequest(
	ctx context.Context,
	params domain.CreateShiftSwapRequest,
	requesterEmployeeID uuid.UUID,
) (*domain.ShiftSwapRequestRecord, error) {
	createParams := db.CreateShiftSwapRequestParams{
		RequesterEmployeeID: requesterEmployeeID,
		RecipientEmployeeID: params.RecipientEmployeeID,
		RequesterScheduleID: params.RequesterScheduleID,
		RecipientScheduleID: params.RecipientScheduleID,
		Status:              db.ShiftSwapStatusEnumPendingRecipient,
		ExpiresAt:           pgtype.Timestamptz{Valid: false},
	}
	if params.ExpiresAt != nil {
		createParams.ExpiresAt = conv.PgTimestamptzFromTime(params.ExpiresAt.UTC())
	}

	row, err := r.store.CreateShiftSwapRequest(ctx, createParams)
	if err != nil {
		if isShiftSwapUniqueViolation(err) {
			return nil, domain.ErrShiftSwapDuplicateActiveRequest
		}
		return nil, err
	}

	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) UpdateShiftSwapStatusAfterRecipientResponse(
	ctx context.Context,
	swapID, recipientEmployeeID uuid.UUID,
	status string,
	note *string,
) (*domain.ShiftSwapRequestRecord, error) {
	dbStatus, ok := parseDBShiftSwapStatus(status)
	if !ok {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	row, err := r.store.UpdateShiftSwapStatusAfterRecipientResponse(
		ctx,
		db.UpdateShiftSwapStatusAfterRecipientResponseParams{
			Status:                dbStatus,
			RecipientResponseNote: note,
			ID:                    swapID,
			RecipientEmployeeID:   recipientEmployeeID,
		},
	)
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) UpdateShiftSwapAdminDecision(
	ctx context.Context,
	swapID uuid.UUID,
	status string,
	note *string,
	adminEmployeeID uuid.UUID,
) (*domain.ShiftSwapRequestRecord, error) {
	dbStatus, ok := parseDBShiftSwapStatus(status)
	if !ok {
		return nil, domain.ErrShiftSwapInvalidRequest
	}
	row, err := r.store.UpdateShiftSwapAdminDecision(ctx, db.UpdateShiftSwapAdminDecisionParams{
		Status:            dbStatus,
		AdminDecisionNote: note,
		AdminEmployeeID:   &adminEmployeeID,
		ID:                swapID,
	})
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) MarkShiftSwapConfirmed(
	ctx context.Context,
	swapID uuid.UUID,
	note *string,
	adminEmployeeID uuid.UUID,
) (*domain.ShiftSwapRequestRecord, error) {
	row, err := r.store.MarkShiftSwapConfirmed(ctx, db.MarkShiftSwapConfirmedParams{
		ID:                swapID,
		AdminDecisionNote: note,
		AdminEmployeeID:   &adminEmployeeID,
	})
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) GetShiftSwapRequestByID(
	ctx context.Context,
	swapID uuid.UUID,
) (*domain.ShiftSwapRequestRecord, error) {
	row, err := r.store.GetShiftSwapRequestByID(ctx, swapID)
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) GetShiftSwapRequestDetailsByID(
	ctx context.Context,
	swapID uuid.UUID,
) (*domain.ShiftSwapResponse, error) {
	row, err := r.store.GetShiftSwapRequestDetailsByID(ctx, swapID)
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapDetails(row, uuid.Nil)
	return &item, nil
}

func (r *ScheduleRepository) ListMyShiftSwapRequests(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.ShiftSwapResponse, error) {
	rows, err := r.store.ListMyShiftSwapRequests(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.ShiftSwapResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainShiftSwapListRow(row, employeeID))
	}
	return result, nil
}

func (r *ScheduleRepository) ListShiftSwapRequests(
	ctx context.Context,
	params domain.ListShiftSwapRequestsParams,
) (*domain.ShiftSwapPage, error) {
	queryArg := db.ListShiftSwapRequestsPaginatedParams{
		Status:     nil,
		EmployeeID: params.EmployeeID,
		Filter:     params.Filter,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}
	if params.Status != nil {
		if parsed, ok := parseDBShiftSwapStatus(*params.Status); ok {
			queryArg.Status = enumPtr(parsed)
		}
	}

	rows, err := r.store.ListShiftSwapRequestsPaginated(ctx, queryArg)
	if err != nil {
		return nil, err
	}

	page := &domain.ShiftSwapPage{
		Items: make([]domain.ShiftSwapResponse, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainShiftSwapPaginatedRow(row, uuid.Nil))
	}
	return page, nil
}

func (r *ScheduleRepository) LockSchedulesByIDsForSwap(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.ScheduleSwapValidation, error) {
	rows, err := r.store.LockSchedulesByIDsForSwap(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ScheduleSwapValidation, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ScheduleSwapValidation{
			ID:            row.ID,
			EmployeeID:    row.EmployeeID,
			StartDatetime: conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.EndDatetime),
		})
	}
	return result, nil
}

func (r *ScheduleRepository) LockShiftSwapRequestForAdminDecision(
	ctx context.Context,
	swapID uuid.UUID,
) (*domain.ShiftSwapRequestRecord, error) {
	row, err := r.store.LockShiftSwapRequestForAdminDecision(ctx, swapID)
	if err != nil {
		return nil, err
	}
	item := toDomainShiftSwapRequestRecord(row)
	return &item, nil
}

func (r *ScheduleRepository) CountScheduleOverlapsForEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	excludedScheduleIDs []uuid.UUID,
	conflictStart, conflictEnd time.Time,
) (int64, error) {
	return r.store.CountScheduleOverlapsForEmployee(ctx, db.CountScheduleOverlapsForEmployeeParams{
		EmployeeID:          employeeID,
		ExcludedScheduleIds: excludedScheduleIDs,
		ConflictStart:       conv.PgTimestamptzFromTime(conflictStart),
		ConflictEnd:         conv.PgTimestamptzFromTime(conflictEnd),
	})
}

func (r *ScheduleRepository) UpdateScheduleEmployeeAssignment(
	ctx context.Context,
	scheduleID, employeeID uuid.UUID,
) error {
	return r.store.UpdateScheduleEmployeeAssignment(ctx, db.UpdateScheduleEmployeeAssignmentParams{
		ID:         scheduleID,
		EmployeeID: employeeID,
	})
}

func toDomainShiftSwapRequestRecord(row db.ShiftSwapRequest) domain.ShiftSwapRequestRecord {
	return domain.ShiftSwapRequestRecord{
		ID:                    row.ID,
		RequesterEmployeeID:   row.RequesterEmployeeID,
		RecipientEmployeeID:   row.RecipientEmployeeID,
		RequesterScheduleID:   row.RequesterScheduleID,
		RecipientScheduleID:   row.RecipientScheduleID,
		Status:                string(row.Status),
		RequestedAt:           conv.TimeFromPgTimestamptz(row.RequestedAt),
		RecipientRespondedAt:  toTimePtr(row.RecipientRespondedAt),
		AdminDecidedAt:        toTimePtr(row.AdminDecidedAt),
		RecipientResponseNote: row.RecipientResponseNote,
		AdminDecisionNote:     row.AdminDecisionNote,
		AdminEmployeeID:       row.AdminEmployeeID,
		ExpiresAt:             toTimePtr(row.ExpiresAt),
		CreatedAt:             conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:             conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func formatLocationAddress(
	street string,
	houseNumber string,
	houseNumberAddition *string,
	postalCode string,
	city string,
) string {
	number := strings.TrimSpace(houseNumber)
	if houseNumberAddition != nil && strings.TrimSpace(*houseNumberAddition) != "" {
		number = strings.TrimSpace(number + " " + strings.TrimSpace(*houseNumberAddition))
	}

	return strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(street + " " + number),
		strings.TrimSpace(postalCode + " " + city),
	}, ", "))
}

func toDomainShiftSwapDetails(
	row db.GetShiftSwapRequestDetailsByIDRow,
	viewerEmployeeID uuid.UUID,
) domain.ShiftSwapResponse {
	requesterName := strings.TrimSpace(row.RequesterFirstName + " " + row.RequesterLastName)
	recipientName := strings.TrimSpace(row.RecipientFirstName + " " + row.RecipientLastName)

	resp := domain.ShiftSwapResponse{
		ID:                    row.ID,
		RequesterEmployeeID:   row.RequesterEmployeeID,
		RequesterEmployeeName: requesterName,
		RecipientEmployeeID:   row.RecipientEmployeeID,
		RecipientEmployeeName: recipientName,
		RequesterSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RequesterScheduleID,
			EmployeeID:    row.RequesterEmployeeID,
			EmployeeName:  requesterName,
			ShiftName:     row.RequesterScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RequesterScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RequesterScheduleEndDatetime),
		},
		RecipientSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RecipientScheduleID,
			EmployeeID:    row.RecipientEmployeeID,
			EmployeeName:  recipientName,
			ShiftName:     row.RecipientScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RecipientScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RecipientScheduleEndDatetime),
		},
		Status:                string(row.Status),
		RequestedAt:           conv.TimeFromPgTimestamptz(row.RequestedAt),
		RecipientRespondedAt:  toTimePtr(row.RecipientRespondedAt),
		AdminDecidedAt:        toTimePtr(row.AdminDecidedAt),
		RecipientResponseNote: row.RecipientResponseNote,
		AdminDecisionNote:     row.AdminDecisionNote,
		AdminEmployeeID:       row.AdminEmployeeID,
		ExpiresAt:             toTimePtr(row.ExpiresAt),
	}
	if row.AdminFirstName != nil && row.AdminLastName != nil {
		name := strings.TrimSpace(*row.AdminFirstName + " " + *row.AdminLastName)
		resp.AdminEmployeeName = &name
	}
	if viewerEmployeeID == row.RequesterEmployeeID {
		resp.Direction = "sent"
	} else if viewerEmployeeID == row.RecipientEmployeeID {
		resp.Direction = "received"
	}
	return resp
}

func toDomainShiftSwapListRow(
	row db.ListMyShiftSwapRequestsRow,
	viewerEmployeeID uuid.UUID,
) domain.ShiftSwapResponse {
	requesterName := strings.TrimSpace(row.RequesterFirstName + " " + row.RequesterLastName)
	recipientName := strings.TrimSpace(row.RecipientFirstName + " " + row.RecipientLastName)

	resp := domain.ShiftSwapResponse{
		ID:                    row.ID,
		RequesterEmployeeID:   row.RequesterEmployeeID,
		RequesterEmployeeName: requesterName,
		RecipientEmployeeID:   row.RecipientEmployeeID,
		RecipientEmployeeName: recipientName,
		RequesterSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RequesterScheduleID,
			EmployeeID:    row.RequesterEmployeeID,
			EmployeeName:  requesterName,
			ShiftName:     row.RequesterScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RequesterScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RequesterScheduleEndDatetime),
		},
		RecipientSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RecipientScheduleID,
			EmployeeID:    row.RecipientEmployeeID,
			EmployeeName:  recipientName,
			ShiftName:     row.RecipientScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RecipientScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RecipientScheduleEndDatetime),
		},
		Status:                string(row.Status),
		RequestedAt:           conv.TimeFromPgTimestamptz(row.RequestedAt),
		RecipientRespondedAt:  toTimePtr(row.RecipientRespondedAt),
		AdminDecidedAt:        toTimePtr(row.AdminDecidedAt),
		RecipientResponseNote: row.RecipientResponseNote,
		AdminDecisionNote:     row.AdminDecisionNote,
		AdminEmployeeID:       row.AdminEmployeeID,
		ExpiresAt:             toTimePtr(row.ExpiresAt),
	}
	if row.AdminFirstName != nil && row.AdminLastName != nil {
		name := strings.TrimSpace(*row.AdminFirstName + " " + *row.AdminLastName)
		resp.AdminEmployeeName = &name
	}
	if viewerEmployeeID == row.RequesterEmployeeID {
		resp.Direction = "sent"
	} else if viewerEmployeeID == row.RecipientEmployeeID {
		resp.Direction = "received"
	}
	return resp
}

func toDomainShiftSwapPaginatedRow(
	row db.ListShiftSwapRequestsPaginatedRow,
	viewerEmployeeID uuid.UUID,
) domain.ShiftSwapResponse {
	requesterName := strings.TrimSpace(row.RequesterFirstName + " " + row.RequesterLastName)
	recipientName := strings.TrimSpace(row.RecipientFirstName + " " + row.RecipientLastName)

	resp := domain.ShiftSwapResponse{
		ID:                    row.ID,
		RequesterEmployeeID:   row.RequesterEmployeeID,
		RequesterEmployeeName: requesterName,
		RecipientEmployeeID:   row.RecipientEmployeeID,
		RecipientEmployeeName: recipientName,
		RequesterSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RequesterScheduleID,
			EmployeeID:    row.RequesterEmployeeID,
			EmployeeName:  requesterName,
			ShiftName:     row.RequesterScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RequesterScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RequesterScheduleEndDatetime),
		},
		RecipientSchedule: domain.ShiftSwapScheduleSnapshot{
			ID:            row.RecipientScheduleID,
			EmployeeID:    row.RecipientEmployeeID,
			EmployeeName:  recipientName,
			ShiftName:     row.RecipientScheduleShiftName,
			StartDatetime: conv.TimeFromPgTimestamptz(row.RecipientScheduleStartDatetime),
			EndDatetime:   conv.TimeFromPgTimestamptz(row.RecipientScheduleEndDatetime),
		},
		Status:                string(row.Status),
		RequestedAt:           conv.TimeFromPgTimestamptz(row.RequestedAt),
		RecipientRespondedAt:  toTimePtr(row.RecipientRespondedAt),
		AdminDecidedAt:        toTimePtr(row.AdminDecidedAt),
		RecipientResponseNote: row.RecipientResponseNote,
		AdminDecisionNote:     row.AdminDecisionNote,
		AdminEmployeeID:       row.AdminEmployeeID,
		ExpiresAt:             toTimePtr(row.ExpiresAt),
	}
	if row.AdminFirstName != nil && row.AdminLastName != nil {
		name := strings.TrimSpace(*row.AdminFirstName + " " + *row.AdminLastName)
		resp.AdminEmployeeName = &name
	}
	if viewerEmployeeID == row.RequesterEmployeeID {
		resp.Direction = "sent"
	} else if viewerEmployeeID == row.RecipientEmployeeID {
		resp.Direction = "received"
	}
	return resp
}

func parseDBShiftSwapStatus(value string) (db.ShiftSwapStatusEnum, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(db.ShiftSwapStatusEnumPendingRecipient):
		return db.ShiftSwapStatusEnumPendingRecipient, true
	case string(db.ShiftSwapStatusEnumRecipientRejected):
		return db.ShiftSwapStatusEnumRecipientRejected, true
	case string(db.ShiftSwapStatusEnumPendingAdmin):
		return db.ShiftSwapStatusEnumPendingAdmin, true
	case string(db.ShiftSwapStatusEnumAdminRejected):
		return db.ShiftSwapStatusEnumAdminRejected, true
	case string(db.ShiftSwapStatusEnumConfirmed):
		return db.ShiftSwapStatusEnumConfirmed, true
	case string(db.ShiftSwapStatusEnumCancelled):
		return db.ShiftSwapStatusEnumCancelled, true
	case string(db.ShiftSwapStatusEnumExpired):
		return db.ShiftSwapStatusEnumExpired, true
	default:
		return "", false
	}
}

func (r *ScheduleRepository) ListEmployeeUpcomingShifts(
	ctx context.Context,
	employeeID uuid.UUID,
	now time.Time,
) ([]domain.EmployeeUpcomingShift, error) {
	rows, err := r.store.ListEmployeeUpcomingShifts(ctx, db.ListEmployeeUpcomingShiftsParams{
		EmployeeID: employeeID,
		Now:        conv.PgTimestamptzFromTime(now),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeUpcomingShift, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeUpcomingShift{
			ScheduleID:   row.ScheduleID,
			ShiftName:    row.ShiftName,
			IsCustom:     row.IsCustom,
			LocationID:   row.LocationID,
			LocationName: row.LocationName,
			Address:      formatLocationAddress(row.Street, row.HouseNumber, row.HouseNumberAddition, row.PostalCode, row.City),
			StartTime:    conv.TimeFromPgTimestamptz(row.StartDatetime),
			EndTime:      conv.TimeFromPgTimestamptz(row.EndDatetime),
			Date:         conv.TimeFromPgDate(row.ShiftDate).Format("2006-01-02"),
			Colleagues:   nil,
		})
	}
	return result, nil
}

func (r *ScheduleRepository) ListShiftColleaguesByScheduleIDs(
	ctx context.Context,
	scheduleIDs []uuid.UUID,
	employeeID uuid.UUID,
) ([]domain.EmployeeUpcomingShiftColleagueRow, error) {
	rows, err := r.store.ListShiftColleaguesByScheduleIDs(ctx, db.ListShiftColleaguesByScheduleIDsParams{
		ScheduleIds: scheduleIDs,
		EmployeeID:  employeeID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeUpcomingShiftColleagueRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.EmployeeUpcomingShiftColleagueRow{
			ScheduleID: row.ScheduleID,
			EmployeeID: row.EmployeeID,
			FirstName:  row.FirstName,
			LastName:   row.LastName,
		})
	}
	return result, nil
}

func toTimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func isShiftSwapUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" &&
			(strings.Contains(pgErr.ConstraintName, "uq_shift_swap_active_requester_schedule") ||
				strings.Contains(pgErr.ConstraintName, "uq_shift_swap_active_recipient_schedule") ||
				strings.Contains(pgErr.ConstraintName, "uq_shift_swap_active_schedule_any"))
	}
	return false
}
