package repository

import (
	"context"
	"math"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type LeaveRepository struct {
	store *db.Store
}

func NewLeaveRepository(store *db.Store) domain.LeaveRepository {
	return &LeaveRepository{
		store: store,
	}
}

func (r *LeaveRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.LeaveTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&leaveTxRepo{queries: q})
	})
}

func (r *LeaveRepository) CreateLeaveRequest(
	ctx context.Context,
	params domain.CreateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	leaveType, ok := toDBLeaveType(params.LeaveType)
	if !ok {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	durationType, ok := toDBLeaveDurationType(params.DurationType)
	if !ok {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	row, err := r.store.CreateLeaveRequest(ctx, db.CreateLeaveRequestParams{
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: &params.CreatedByEmployeeID,
		LeaveType:           leaveType,
		DurationType:        durationType,
		RequestedMinutes:    params.RequestedMinutes,
		StartDate:           conv.PgDateFromTime(params.StartDate),
		EndDate:             conv.PgDateFromTime(params.EndDate),
		StartTime:           pgTimeFromTimePtr(params.StartTime),
		EndTime:             pgTimeFromTimePtr(params.EndTime),
		Reason:              params.Reason,
	})
	if err != nil {
		return nil, err
	}

	model := toDomainLeaveRequest(row)
	return &model, nil
}

func (r *LeaveRepository) GetActiveLeavePolicyByType(
	ctx context.Context,
	leaveType string,
) (*domain.LeavePolicy, error) {
	dbType, ok := toDBLeaveType(leaveType)
	if !ok {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	row, err := r.store.GetActiveLeavePolicyByType(ctx, dbType)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestInvalidRequest
		}
		return nil, err
	}

	return &domain.LeavePolicy{
		LeaveType:      string(row.LeaveType),
		DeductsBalance: row.DeductsBalance,
	}, nil
}

func (r *LeaveRepository) GetEmployeeContractAtDate(
	ctx context.Context,
	employeeID uuid.UUID,
	date time.Time,
) (*domain.LeaveContractAtDate, error) {
	row, err := r.store.GetEmployeeContractAtDate(ctx, db.GetEmployeeContractAtDateParams{
		EmployeeID: employeeID,
		TargetDate: conv.PgDateFromTime(date),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	return &domain.LeaveContractAtDate{
		EmployeeID:    row.EmployeeID,
		RosterFreeDay: string(row.RosterFreeDay),
	}, nil
}

func (r *LeaveRepository) ListMyLeaveRequests(
	ctx context.Context,
	params domain.ListMyLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	rows, err := r.store.ListMyLeaveRequestsPaginated(ctx, db.ListMyLeaveRequestsPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBLeaveStatusPtr(params.Status),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.LeaveRequestPage{
		Items: make([]domain.LeaveRequestListItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainLeaveRequestListItem(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.LeaveType,
			row.Status,
			row.DurationType,
			row.RequestedMinutes,
			row.StartTime,
			row.EndTime,
			row.StartDate,
			row.EndDate,
			row.Reason,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.CancelledAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *LeaveRepository) ListLeaveRequests(
	ctx context.Context,
	params domain.ListLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	rows, err := r.store.ListLeaveRequestsPaginated(ctx, db.ListLeaveRequestsPaginatedParams{
		Status:         toDBLeaveStatusPtr(params.Status),
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.LeaveRequestPage{
		Items: make([]domain.LeaveRequestListItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainLeaveRequestListItem(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.LeaveType,
			row.Status,
			row.DurationType,
			row.RequestedMinutes,
			row.StartTime,
			row.EndTime,
			row.StartDate,
			row.EndDate,
			row.Reason,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.CancelledAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *LeaveRepository) ListLeaveCalendar(
	ctx context.Context,
	params domain.ListLeaveCalendarParams,
) ([]domain.LeaveCalendarEmployee, error) {
	rows, err := r.store.ListLeaveCalendarRows(ctx, db.ListLeaveCalendarRowsParams{
		MonthStart:        conv.PgDateFromTime(params.Month),
		MonthEndExclusive: conv.PgDateFromTime(params.Month.AddDate(0, 1, 0)),
		DepartmentID:      params.DepartmentID,
		EmployeeSearch:    ptr.TrimString(params.EmployeeSearch),
		LeaveTypes:        toDBLeaveTypes(params.LeaveTypes),
	})
	if err != nil {
		return nil, err
	}

	return toDomainLeaveCalendar(rows), nil
}

func (r *LeaveRepository) GetMyLeaveRequestStats(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.LeaveRequestStats, error) {
	row, err := r.store.GetMyLeaveRequestStats(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return &domain.LeaveRequestStats{
		OpenRequests:     row.OpenRequests,
		ApprovedRequests: row.ApprovedRequests,
		RejectedRequests: row.RejectedRequests,
		SicknessAbsence:  row.SicknessAbsence,
	}, nil
}

func (r *LeaveRepository) GetLeaveRequestStats(
	ctx context.Context,
) (*domain.LeaveRequestStats, error) {
	row, err := r.store.GetLeaveRequestStats(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.LeaveRequestStats{
		OpenRequests:     row.OpenRequests,
		ApprovedRequests: row.ApprovedRequests,
		RejectedRequests: row.RejectedRequests,
		SicknessAbsence:  row.SicknessAbsence,
	}, nil
}

func (r *LeaveRepository) ListLeaveBalances(
	ctx context.Context,
	params domain.ListLeaveBalancesParams,
) (*domain.LeaveBalancePage, error) {
	rows, err := r.store.ListLeaveBalancesPaginated(ctx, db.ListLeaveBalancesPaginatedParams{
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Year:           params.Year,
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.LeaveBalancePage{
		Items: make([]domain.LeaveBalance, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainLeaveBalance(
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.Year,
			row.LegalTotalMinutes,
			row.LegalUsedMinutes,
			row.ContractHours,
			stringPtr(string(row.ContractType)),
			conv.TimePtrFromPgDate(row.ContractStartDate),
			conv.TimePtrFromPgDate(row.ContractEndDate),
			conv.TimePtrFromPgDate(row.EffectiveEndDate),
		))
	}

	return page, nil
}

func (r *LeaveRepository) getDeductedLeaves(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) ([]domain.DeductedLeaveSummary, error) {
	deductedRows, err := r.store.GetDeductedLeavesForEmployeeAndYear(
		ctx,
		db.GetDeductedLeavesForEmployeeAndYearParams{
			EmployeeID: employeeID,
			Year:       year,
		},
	)
	if err != nil {
		return nil, err
	}

	deductedLeaves := make([]domain.DeductedLeaveSummary, len(deductedRows))
	for i, dr := range deductedRows {
		deductedLeaves[i] = domain.DeductedLeaveSummary{
			ID:              dr.ID,
			LeaveType:       string(dr.LeaveType),
			StartDate:       conv.TimeFromPgDate(dr.StartDate),
			EndDate:         conv.TimeFromPgDate(dr.EndDate),
			DurationMinutes: dr.RequestedMinutes,
		}
	}
	return deductedLeaves, nil
}

func (r *LeaveRepository) getDeductedPayouts(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) ([]domain.DeductedPayoutSummary, error) {
	deductedRows, err := r.store.GetDeductedPayoutsForEmployeeAndYear(
		ctx,
		db.GetDeductedPayoutsForEmployeeAndYearParams{
			EmployeeID: employeeID,
			Year:       year,
		},
	)
	if err != nil {
		return nil, err
	}

	deductedPayouts := make([]domain.DeductedPayoutSummary, len(deductedRows))
	for i, dr := range deductedRows {
		deductedPayouts[i] = domain.DeductedPayoutSummary{
			ID:              dr.ID,
			RequestedHours:  dr.RequestedHours,
			DurationMinutes: dr.RequestedMinutes,
			BalanceYear:     dr.BalanceYear,
			PayPeriodStart:  conv.TimePtrFromPgDate(dr.PayPeriodStart),
			Status:          string(dr.Status),
			RequestedAt:     conv.TimeFromPgTimestamptz(dr.RequestedAt),
			DecidedAt:       timePtrFromPgTimestamptz(dr.DecidedAt),
			PaidAt:          timePtrFromPgTimestamptz(dr.PaidAt),
		}
	}
	return deductedPayouts, nil
}

func (r *LeaveRepository) GetLeaveBalanceDetails(
	ctx context.Context,
	params domain.GetLeaveBalanceDetailsParams,
) (*domain.LeaveBalanceDetails, error) {
	row, err := r.store.GetLeaveBalanceDetails(ctx, db.GetLeaveBalanceDetailsParams{
		EmployeeID: params.EmployeeID,
		Year:       params.Year,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}

	accrualRows, err := r.store.ListLeaveContractAccrualsForYear(
		ctx,
		db.ListLeaveContractAccrualsForYearParams{
			EmployeeID: params.EmployeeID,
			Year:       params.Year,
		},
	)
	if err != nil {
		return nil, err
	}

	deducted, err := r.getDeductedLeaves(ctx, row.EmployeeID, row.Year)
	if err != nil {
		return nil, err
	}
	deductedPayouts, err := r.getDeductedPayouts(ctx, row.EmployeeID, row.Year)
	if err != nil {
		return nil, err
	}

	balance := toDomainLeaveBalance(
		row.EmployeeID,
		strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
		row.Year,
		row.LegalTotalMinutes,
		row.LegalUsedMinutes,
		row.ContractHours,
		stringPtr(string(row.ContractType)),
		conv.TimePtrFromPgDate(row.ContractStartDate),
		conv.TimePtrFromPgDate(row.ContractEndDate),
		conv.TimePtrFromPgDate(row.EffectiveEndDate),
	)
	balance.DeductedLeaves = deducted
	balance.DeductedPayouts = deductedPayouts

	return &domain.LeaveBalanceDetails{
		Balance:          balance,
		ContractAccruals: toDomainLeaveContractAccruals(accrualRows),
	}, nil
}

type leaveTxRepo struct {
	queries *db.Queries
}

func (r *leaveTxRepo) GetLeaveRequestForUpdate(
	ctx context.Context,
	leaveRequestID uuid.UUID,
) (*domain.LeaveRequest, error) {
	row, err := r.queries.LockLeaveRequestByID(ctx, leaveRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	model := toDomainLeaveRequest(row)
	return &model, nil
}

func (r *leaveTxRepo) UpdateLeaveRequestEditableFields(
	ctx context.Context,
	leaveRequestID uuid.UUID,
	params domain.UpdateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	row, err := r.queries.UpdateLeaveRequestEditableFields(
		ctx,
		db.UpdateLeaveRequestEditableFieldsParams{
			ID: leaveRequestID,
			LeaveType: func() db.LeaveRequestTypeEnum {
				if params.LeaveType == nil {
					return ""
				}
				leaveType, ok := toDBLeaveType(*params.LeaveType)
				if !ok {
					return ""
				}
				return leaveType
			}(),
			DurationType: func() db.LeaveDurationTypeEnum {
				if params.DurationType == nil {
					return ""
				}
				durationType, ok := toDBLeaveDurationType(*params.DurationType)
				if !ok {
					return ""
				}
				return durationType
			}(),
			StartDate: func() pgtype.Date {
				if params.StartDate == nil {
					return pgtype.Date{}
				}
				return conv.PgDateFromTime(*params.StartDate)
			}(),
			EndDate: func() pgtype.Date {
				if params.EndDate == nil {
					return pgtype.Date{}
				}
				return conv.PgDateFromTime(*params.EndDate)
			}(),
			RequestedMinutes: func() int32 {
				if params.RequestedMinutes == nil {
					return 0
				}
				return *params.RequestedMinutes
			}(),
			StartTime: pgTimeFromTimePtr(params.StartTime),
			EndTime:   pgTimeFromTimePtr(params.EndTime),
			Reason:    params.Reason,
		},
	)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	model := toDomainLeaveRequest(row)
	return &model, nil
}

func (r *leaveTxRepo) UpdateLeaveRequestDecision(
	ctx context.Context,
	leaveRequestID uuid.UUID,
	status string,
	decisionNote *string,
	decidedByEmployeeID uuid.UUID,
) (*domain.LeaveRequest, error) {
	dbStatus, ok := toDBLeaveStatus(status)
	if !ok {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	row, err := r.queries.UpdateLeaveRequestDecision(ctx, db.UpdateLeaveRequestDecisionParams{
		ID:                  leaveRequestID,
		Status:              dbStatus,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	model := toDomainLeaveRequest(row)
	return &model, nil
}

func (r *leaveTxRepo) GetActiveLeavePolicyByType(
	ctx context.Context,
	leaveType string,
) (*domain.LeavePolicy, error) {
	dbType, ok := toDBLeaveType(leaveType)
	if !ok {
		return nil, domain.ErrLeaveRequestInvalidRequest
	}

	row, err := r.queries.GetActiveLeavePolicyByType(ctx, dbType)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestInvalidRequest
		}
		return nil, err
	}
	return &domain.LeavePolicy{
		LeaveType:      string(row.LeaveType),
		DeductsBalance: row.DeductsBalance,
	}, nil
}

func (r *leaveTxRepo) LockEmployeeForLeaveBalance(
	ctx context.Context,
	employeeID uuid.UUID,
) error {
	_, err := r.queries.LockEmployeeProfileForLeaveBalance(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return domain.ErrLeaveRequestNotFound
		}
	}
	return err
}

func (r *leaveTxRepo) GetLeaveHoursPerDay(
	ctx context.Context,
	employeeID uuid.UUID,
) (int32, error) {
	row, err := r.queries.GetEmployeeContractForLeave(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return 0, domain.ErrLeaveRequestNotFound
		}
		return 0, err
	}

	if row.ContractType != db.EmployeeContractTypeEnumPermanent {
		return 8, nil
	}
	if row.ContractHours != nil && *row.ContractHours > 0 {
		computed := int32(math.Round(*row.ContractHours / 5.0))
		if computed <= 0 {
			return 8, nil
		}
		return computed, nil
	}
	return 8, nil
}

func (r *leaveTxRepo) GetEmployeeContractAtDate(
	ctx context.Context,
	employeeID uuid.UUID,
	date time.Time,
) (*domain.LeaveContractAtDate, error) {
	row, err := r.queries.GetEmployeeContractAtDate(ctx, db.GetEmployeeContractAtDateParams{
		EmployeeID: employeeID,
		TargetDate: conv.PgDateFromTime(date),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	return &domain.LeaveContractAtDate{
		EmployeeID:    row.EmployeeID,
		RosterFreeDay: string(row.RosterFreeDay),
	}, nil
}

func (r *leaveTxRepo) ComputeLegalLeaveTotalForYear(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
	asOf time.Time,
) (int32, error) {
	return r.queries.ComputeLegalLeaveTotalForYear(ctx, db.ComputeLegalLeaveTotalForYearParams{
		EmployeeID: employeeID,
		Year:       year,
		AsOf:       conv.PgTimestamptzFromTime(asOf),
	})
}

func (r *leaveTxRepo) ComputeLegalLeaveUsedForYear(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) (int32, error) {
	return r.queries.ComputeLegalLeaveUsedForYear(ctx, db.ComputeLegalLeaveUsedForYearParams{
		EmployeeID: employeeID,
		Year:       year,
	})
}

func toDomainLeaveRequest(row db.LeaveRequest) domain.LeaveRequest {
	return domain.LeaveRequest{
		ID:                  row.ID,
		EmployeeID:          row.EmployeeID,
		CreatedByEmployeeID: row.CreatedByEmployeeID,
		LeaveType:           string(row.LeaveType),
		Status:              string(row.Status),
		DurationType:        string(row.DurationType),
		RequestedMinutes:    row.RequestedMinutes,
		StartTime:           timePtrFromPgTime(row.StartTime),
		EndTime:             timePtrFromPgTime(row.EndTime),
		StartDate:           conv.TimeFromPgDate(row.StartDate),
		EndDate:             conv.TimeFromPgDate(row.EndDate),
		Reason:              row.Reason,
		DecisionNote:        row.DecisionNote,
		DecidedByEmployeeID: row.DecidedByEmployeeID,
		RequestedAt:         conv.TimeFromPgTimestamptz(row.RequestedAt),
		DecidedAt:           timePtrFromPgTimestamptz(row.DecidedAt),
		CancelledAt:         timePtrFromPgTimestamptz(row.CancelledAt),
		CreatedAt:           conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:           conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainLeaveRequestListItem(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	createdByEmployeeID *uuid.UUID,
	leaveType db.LeaveRequestTypeEnum,
	status db.LeaveRequestStatusEnum,
	durationType db.LeaveDurationTypeEnum,
	requestedMinutes int32,
	startTime pgtype.Time,
	endTime pgtype.Time,
	startDate pgtype.Date,
	endDate pgtype.Date,
	reason *string,
	decisionNote *string,
	decidedByEmployeeID *uuid.UUID,
	requestedAt pgtype.Timestamptz,
	decidedAt pgtype.Timestamptz,
	cancelledAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.LeaveRequestListItem {
	return domain.LeaveRequestListItem{
		LeaveRequest: domain.LeaveRequest{
			ID:                  id,
			EmployeeID:          employeeID,
			CreatedByEmployeeID: createdByEmployeeID,
			LeaveType:           string(leaveType),
			Status:              string(status),
			DurationType:        string(durationType),
			RequestedMinutes:    requestedMinutes,
			StartTime:           timePtrFromPgTime(startTime),
			EndTime:             timePtrFromPgTime(endTime),
			StartDate:           conv.TimeFromPgDate(startDate),
			EndDate:             conv.TimeFromPgDate(endDate),
			Reason:              reason,
			DecisionNote:        decisionNote,
			DecidedByEmployeeID: decidedByEmployeeID,
			RequestedAt:         conv.TimeFromPgTimestamptz(requestedAt),
			DecidedAt:           timePtrFromPgTimestamptz(decidedAt),
			CancelledAt:         timePtrFromPgTimestamptz(cancelledAt),
			CreatedAt:           conv.TimeFromPgTimestamptz(createdAt),
			UpdatedAt:           conv.TimeFromPgTimestamptz(updatedAt),
		},
		EmployeeName: employeeName,
	}
}

func toDomainLeaveBalance(
	employeeID uuid.UUID,
	employeeName string,
	year int32,
	legalTotalMinutes int32,
	legalUsedMinutes int32,
	contractHours *float64,
	contractType *string,
	contractStartDate *time.Time,
	contractEndDate *time.Time,
	effectiveEndDate *time.Time,
) domain.LeaveBalance {
	legalRemaining := legalTotalMinutes - legalUsedMinutes
	return domain.LeaveBalance{
		EmployeeID:            employeeID,
		EmployeeName:          employeeName,
		Year:                  year,
		LegalTotalMinutes:     legalTotalMinutes,
		LegalUsedMinutes:      legalUsedMinutes,
		LegalRemainingMinutes: legalRemaining,
		TotalRemainingMinutes: legalRemaining,
		ContractHours:         contractHours,
		ContractType:          contractType,
		ContractStartDate:     contractStartDate,
		ContractEndDate:       contractEndDate,
		EffectiveEndDate:      effectiveEndDate,
	}
}

func toDomainLeaveContractAccruals(
	rows []db.ListLeaveContractAccrualsForYearRow,
) []domain.LeaveContractAccrual {
	items := make([]domain.LeaveContractAccrual, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.LeaveContractAccrual{
			ContractID:        row.ContractID,
			ContractType:      string(row.ContractType),
			ContractHours:     row.ContractHours,
			ContractStartDate: conv.TimeFromPgDate(row.ContractStartDate),
			ContractEndDate:   conv.TimePtrFromPgDate(row.ContractEndDate),
			EffectiveEndDate:  conv.TimePtrFromPgDate(row.EffectiveEndDate),
			SegmentStartDate:  conv.TimeFromPgDate(row.SegmentStartDate),
			SegmentEndDate:    conv.TimeFromPgDate(row.SegmentEndDate),
			YearDays:          row.YearDays,
			SegmentDays:       row.SegmentDays,
			FullYearMinutes:   row.FullYearMinutes,
			ScheduleMinutes:   row.ScheduleMinutes,
			OvertimeMinutes:   row.OvertimeMinutes,
			GainedMinutes:     row.GainedMinutes,
		})
	}
	return items
}

func toDomainLeaveCalendar(rows []db.ListLeaveCalendarRowsRow) []domain.LeaveCalendarEmployee {
	items := make([]domain.LeaveCalendarEmployee, 0, len(rows))
	byEmployee := make(map[uuid.UUID]int, len(rows))

	for _, row := range rows {
		idx, ok := byEmployee[row.EmployeeID]
		if !ok {
			items = append(items, domain.LeaveCalendarEmployee{
				EmployeeID: row.EmployeeID,
				EmployeeName: strings.TrimSpace(
					row.EmployeeFirstName + " " + row.EmployeeLastName,
				),
				DepartmentName: row.DepartmentName,
				LeaveRecords:   make([]domain.LeaveCalendarRecord, 0, 1),
			})
			idx = len(items) - 1
			byEmployee[row.EmployeeID] = idx
		}

		items[idx].LeaveRecords = append(items[idx].LeaveRecords, domain.LeaveCalendarRecord{
			LeaveRequestID:   row.LeaveRequestID,
			LeaveType:        string(row.LeaveType),
			Status:           string(row.Status),
			DurationType:     string(row.DurationType),
			RequestedMinutes: row.RequestedMinutes,
			StartDate:        conv.TimeFromPgDate(row.StartDate),
			EndDate:          conv.TimeFromPgDate(row.EndDate),
			Reason:           row.Reason,
		})
	}

	return items
}

func stringPtr(v string) *string {
	return &v
}

func toDBLeaveTypes(values []string) []db.LeaveRequestTypeEnum {
	if len(values) == 0 {
		return nil
	}

	items := make([]db.LeaveRequestTypeEnum, 0, len(values))
	for _, value := range values {
		dbValue, ok := toDBLeaveType(value)
		if !ok {
			continue
		}
		items = append(items, dbValue)
	}

	if len(items) == 0 {
		return nil
	}

	return items
}

func toDBLeaveType(value string) (db.LeaveRequestTypeEnum, bool) {
	switch db.LeaveRequestTypeEnum(strings.TrimSpace(value)) {
	case db.LeaveRequestTypeEnumVacation,
		db.LeaveRequestTypeEnumPersonal,
		db.LeaveRequestTypeEnumSick,
		db.LeaveRequestTypeEnumPregnancy,
		db.LeaveRequestTypeEnumUnpaid,
		db.LeaveRequestTypeEnumOther:
		return db.LeaveRequestTypeEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBLeaveStatus(value string) (db.LeaveRequestStatusEnum, bool) {
	switch db.LeaveRequestStatusEnum(strings.TrimSpace(value)) {
	case db.LeaveRequestStatusEnumPending,
		db.LeaveRequestStatusEnumApproved,
		db.LeaveRequestStatusEnumRejected,
		db.LeaveRequestStatusEnumCancelled,
		db.LeaveRequestStatusEnumExpired:
		return db.LeaveRequestStatusEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBLeaveStatusPtr(value *string) *db.LeaveRequestStatusEnum {
	if value == nil {
		return nil
	}
	parsed, ok := toDBLeaveStatus(*value)
	if !ok {
		return nil
	}
	return enumPtr(parsed)
}

func timePtrFromPgTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func timePtrFromPgTime(value pgtype.Time) *time.Time {
	if !value.Valid {
		return nil
	}
	totalSeconds := value.Microseconds / 1000000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	result := time.Date(0, 1, 1, int(hours), int(minutes), 0, 0, time.UTC)
	return &result
}

func pgTimeFromTimePtr(value *time.Time) pgtype.Time {
	if value == nil {
		return pgtype.Time{}
	}
	return pgtype.Time{
		Microseconds: int64(value.Hour())*3600000000 +
			int64(value.Minute())*60000000 +
			int64(value.Second())*1000000,
		Valid: true,
	}
}

func toDBLeaveDurationType(value string) (db.LeaveDurationTypeEnum, bool) {
	switch db.LeaveDurationTypeEnum(strings.TrimSpace(value)) {
	case db.LeaveDurationTypeEnumFullDay,
		db.LeaveDurationTypeEnumHours:
		return db.LeaveDurationTypeEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

var _ domain.LeaveRepository = (*LeaveRepository)(nil)
