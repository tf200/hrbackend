package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type SalaryRepository struct {
	store *db.Store
}

func NewSalaryRepository(store *db.Store) domain.SalaryRepository {
	return &SalaryRepository{store: store}
}

// WithTxSalary implements domain.SalaryRepository.
func (r *SalaryRepository) WithTxSalary(
	ctx context.Context,
	fn func(tx domain.SalaryTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&salaryTxRepo{queries: q})
	})
}

func (r *SalaryRepository) ListSalaryScaleSteps(
	ctx context.Context,
	params domain.ListSalaryScaleStepsParams,
) (*domain.SalaryScaleStepsResult, error) {
	rows, err := r.store.ListSalaryScaleSteps(ctx, &params.ActiveOnly)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &domain.SalaryScaleStepsResult{}, nil
	}

	var groups []domain.SalaryScaleGroup
	var currentGroup *domain.SalaryScaleGroup

	for _, row := range rows {
		if currentGroup == nil || currentGroup.Scale != row.Scale {
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}
			currentGroup = &domain.SalaryScaleGroup{
				Scale: row.Scale,
				Steps: nil,
			}
		}

		label := fmt.Sprintf("Scale %d / Step %s - €%.2f/mo", row.Scale, row.Step, row.MonthlySalary)

		step := domain.SalaryScaleStepOption{
			ID:            row.ID,
			SalaryTableID: row.SalaryTableID,
			Step:          row.Step,
			IPNumber:      row.IpNumber,
			MonthlySalary: row.MonthlySalary,
			HourlyRate:    row.HourlyRate,
			Label:         label,
		}

		currentGroup.Steps = append(currentGroup.Steps, step)
	}

	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}

	first := rows[0]
	meta := &domain.SalaryScaleStepsMeta{
		SalaryTableID:   first.SalaryTableID,
		CaoCode:         first.CaoCode,
		SalaryTableName: first.SalaryTableName,
		EffectiveFrom:   conv.TimeFromPgDate(first.EffectiveFrom),
		EffectiveTo:     conv.TimePtrFromPgDate(first.EffectiveTo),
		ScaleCount:      len(groups),
	}

	return &domain.SalaryScaleStepsResult{
		Groups: groups,
		Meta:   meta,
	}, nil
}

func (r *SalaryRepository) GetPayrollPreviewEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.EmployeeDetail, error) {
	row, err := r.store.GetEmployeeProfileByID(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row), nil
}

func (r *SalaryRepository) ListPayrollPreviewWorkItems(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) ([]domain.PayrollWorkItem, error) {
	rows, err := r.store.ListPayrollPreviewWorkItems(ctx, db.ListPayrollPreviewWorkItemsParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: conv.PgTimestamptzFromTime(params.PeriodStart),
		PeriodEnd:   conv.PgTimestamptzFromTime(params.PeriodEnd),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollWorkItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPayrollWorkItem(row))
	}
	return items, nil
}

func (r *SalaryRepository) ListNationalHolidays(
	ctx context.Context,
	countryCode string,
	startDate, endDate time.Time,
) ([]domain.NationalHoliday, error) {
	rows, err := r.store.ListNationalHolidaysInRange(ctx, db.ListNationalHolidaysInRangeParams{
		CountryCode: strings.TrimSpace(strings.ToUpper(countryCode)),
		StartDate:   conv.PgDateFromTime(startDate),
		EndDate:     conv.PgDateFromTime(endDate),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.NationalHoliday, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.NationalHoliday{
			Date: conv.TimeFromPgDate(row.HolidayDate),
			Name: row.Name,
		})
	}

	return items, nil
}

func (r *SalaryRepository) GetPayPeriodByID(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	row, err := r.store.GetPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayPeriodNotFound
		}
		return nil, err
	}
	model := toDomainPayPeriod(
		row.ID,
		row.EmployeeID,
		fullName(row.EmployeeFirstName, row.EmployeeLastName),
		row.PeriodStart,
		row.PeriodEnd,
		row.PayrollGroup,
		row.CutoffAt,
		string(row.Status),
		row.BaseGrossAmount,
		row.IrregularGrossAmount,
		row.GrossAmount,
		row.PaidAt,
		row.CreatedByEmployeeID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *SalaryRepository) ListPayPeriods(
	ctx context.Context,
	params domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	rows, err := r.store.ListPayPeriodsPaginated(ctx, db.ListPayPeriodsPaginatedParams{
		Status:         toDBPayPeriodStatusPtr(params.Status),
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Offset:         params.Offset,
		Limit:          params.Limit,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.PayPeriodPage{
		Items: make([]domain.PayPeriod, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainPayPeriod(
			row.ID,
			row.EmployeeID,
			fullName(row.EmployeeFirstName, row.EmployeeLastName),
			row.PeriodStart,
			row.PeriodEnd,
			row.PayrollGroup,
			row.CutoffAt,
			string(row.Status),
			row.BaseGrossAmount,
			row.IrregularGrossAmount,
			row.GrossAmount,
			row.PaidAt,
			row.CreatedByEmployeeID,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *SalaryRepository) ListPayPeriodLineItems(
	ctx context.Context,
	payPeriodID uuid.UUID,
) ([]domain.PayPeriodLineItem, error) {
	rows, err := r.store.ListPayPeriodLineItemsByPayPeriodID(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayPeriodLineItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPayPeriodLineItem(row))
	}
	return items, nil
}

func (r *SalaryRepository) ListPayrollMonthEmployees(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthEmployee, int64, error) {
	rows, err := r.store.ListPayrollMonthEmployeesPaginated(
		ctx,
		db.ListPayrollMonthEmployeesPaginatedParams{
			EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
			Offset:         params.Offset,
			Limit:          params.Limit,
			MonthStart:     conv.PgDateFromTime(monthStart),
			MonthEnd:       conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, 0, err
	}

	items := make([]domain.PayrollMonthEmployee, 0, len(rows))
	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}
	for _, row := range rows {
		items = append(items, domain.PayrollMonthEmployee{
			EmployeeID:   row.EmployeeID,
			EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		})
	}
	return items, totalCount, nil
}

func (r *SalaryRepository) ListPayrollMonthEmployeesAll(
	ctx context.Context,
	params domain.PayrollMonthORTOverviewParams,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthEmployee, error) {
	rows, err := r.store.ListPayrollMonthEmployeesAll(
		ctx,
		db.ListPayrollMonthEmployeesAllParams{
			EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
			MonthStart:     conv.PgDateFromTime(monthStart),
			MonthEnd:       conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollMonthEmployee, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PayrollMonthEmployee{
			EmployeeID:   row.EmployeeID,
			EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		})
	}
	return items, nil
}

func (r *SalaryRepository) ListFixedPayrollMonthEmployees(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthEmployee, int64, error) {
	rows, err := r.store.ListFixedPayrollMonthEmployeesPaginated(
		ctx,
		db.ListFixedPayrollMonthEmployeesPaginatedParams{
			EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
			Offset:         params.Offset,
			Limit:          params.Limit,
			MonthStart:     conv.PgDateFromTime(monthStart),
			MonthEnd:       conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, 0, err
	}

	items := make([]domain.PayrollMonthEmployee, 0, len(rows))
	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}
	for _, row := range rows {
		items = append(items, domain.PayrollMonthEmployee{
			EmployeeID:   row.EmployeeID,
			EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		})
	}
	return items, totalCount, nil
}

func (r *SalaryRepository) ListOnCallPayrollMonthEmployees(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthEmployee, int64, error) {
	rows, err := r.store.ListOnCallPayrollMonthEmployeesPaginated(
		ctx,
		db.ListOnCallPayrollMonthEmployeesPaginatedParams{
			EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
			Offset:         params.Offset,
			Limit:          params.Limit,
			MonthStart:     conv.PgDateFromTime(monthStart),
			MonthEnd:       conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, 0, err
	}

	items := make([]domain.PayrollMonthEmployee, 0, len(rows))
	var totalCount int64
	if len(rows) > 0 {
		totalCount = rows[0].TotalCount
	}
	for _, row := range rows {
		items = append(items, domain.PayrollMonthEmployee{
			EmployeeID:   row.EmployeeID,
			EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		})
	}
	return items, totalCount, nil
}

func (r *SalaryRepository) ListFixedPayrollContractSegments(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.FixedPayrollContractSegmentSource, error) {
	if len(employeeIDs) == 0 {
		return []domain.FixedPayrollContractSegmentSource{}, nil
	}

	rows, err := r.store.ListFixedPayrollContractSegments(
		ctx,
		db.ListFixedPayrollContractSegmentsParams{
			EmployeeIds: employeeIDs,
			MonthStart:  conv.PgDateFromTime(monthStart),
			MonthEnd:    conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FixedPayrollContractSegmentSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.FixedPayrollContractSegmentSource{
			EmployeeID:           row.EmployeeID,
			ContractID:           row.ContractID,
			ContractType:         string(row.ContractType),
			ActiveFrom:           conv.TimeFromPgDate(row.ActiveFrom),
			ActiveUntil:          conv.TimeFromPgDate(row.ActiveUntil),
			HoursPerWeek:         row.HoursPerWeek,
			FullTimeHoursPerWeek: row.FullTimeHoursPerWeek,
			MonthlySalary:        row.MonthlySalary,
			HourlyRate:           row.HourlyRate,
		})
	}
	return items, nil
}

func (r *SalaryRepository) ListPayPeriodsByEmployeesAndRange(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayPeriod, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayPeriod{}, nil
	}

	rows, err := r.store.ListPayPeriodsByEmployeeIDsAndRange(
		ctx,
		db.ListPayPeriodsByEmployeeIDsAndRangeParams{
			EmployeeIds: employeeIDs,
			MonthStart:  conv.PgDateFromTime(monthStart),
			MonthEnd:    conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayPeriod, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPayPeriod(
			row.ID,
			row.EmployeeID,
			fullName(row.EmployeeFirstName, row.EmployeeLastName),
			row.PeriodStart,
			row.PeriodEnd,
			row.PayrollGroup,
			row.CutoffAt,
			string(row.Status),
			row.BaseGrossAmount,
			row.IrregularGrossAmount,
			row.GrossAmount,
			row.PaidAt,
			row.CreatedByEmployeeID,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return items, nil
}

func (r *SalaryRepository) ListPayrollMonthLockedMultiplierSummaries(
	ctx context.Context,
	payPeriodIDs []uuid.UUID,
) ([]domain.PayrollLockedMultiplierSummary, error) {
	if len(payPeriodIDs) == 0 {
		return []domain.PayrollLockedMultiplierSummary{}, nil
	}

	rows, err := r.store.ListLockedPayPeriodMultiplierSummaries(ctx, payPeriodIDs)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollLockedMultiplierSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PayrollLockedMultiplierSummary{
			PayPeriodID:   row.PayPeriodID,
			RatePercent:   row.AppliedRatePercent,
			WorkedMinutes: row.WorkedMinutes,
			PaidMinutes:   row.PaidMinutes,
			BaseAmount:    row.BaseAmount,
			PremiumAmount: row.PremiumAmount,
		})
	}
	return items, nil
}

func (r *SalaryRepository) ListPayrollMonthApprovedWorkItems(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollWorkItem, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollWorkItem{}, nil
	}

	rows, err := r.store.ListPayrollMonthApprovedWorkItems(ctx, db.ListPayrollMonthApprovedWorkItemsParams{
		EmployeeIds: employeeIDs,
		MonthStart:  conv.PgTimestamptzFromTime(monthStart),
		MonthEnd:    conv.PgTimestamptzFromTime(monthEnd),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollWorkItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPayrollWorkItemFromApproved(row))
	}
	return items, nil
}

func (r *SalaryRepository) ListPayrollMonthPendingSummaries(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthPendingSummary, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollMonthPendingSummary{}, nil
	}

	rows, err := r.store.ListPayrollMonthPendingOvertimeSummaries(
		ctx,
		db.ListPayrollMonthPendingOvertimeSummariesParams{
			EmployeeIds: employeeIDs,
			MonthStart:  conv.PgDateFromTime(monthStart),
			MonthEnd:    conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollMonthPendingSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PayrollMonthPendingSummary{
			EmployeeID:           row.EmployeeID,
			PendingEntryCount:    row.PendingEntryCount,
			PendingWorkedMinutes: row.PendingWorkedMinutes,
		})
	}
	return items, nil
}

func (r *SalaryRepository) ListPayrollMonthPendingEntries(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthPendingEntry, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollMonthPendingEntry{}, nil
	}

	rows, err := r.store.ListPayrollMonthPendingOvertimeEntries(
		ctx,
		db.ListPayrollMonthPendingOvertimeEntriesParams{
			EmployeeIds: employeeIDs,
			MonthStart:  conv.PgDateFromTime(monthStart),
			MonthEnd:    conv.PgDateFromTime(monthEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollMonthPendingEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PayrollMonthPendingEntry{
			EmployeeID:    row.EmployeeID,
			WorkedMinutes: row.WorkedMinutes,
			ContractType:  string(row.ContractType),
		})
	}
	return items, nil
}

func (r *SalaryRepository) ListPendingOvertimeEntriesDetail(
	ctx context.Context,
	employeeID uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollPendingEntryDetail, error) {
	sql := `
		SELECT
			id,
			entry_date,
			minutes AS worked_minutes,
			status::text
		FROM overtime_entries
		WHERE employee_id = $1
		  AND entry_date >= $2
		  AND entry_date <= $3
		  AND status = 'submitted'::overtime_status_enum
		ORDER BY entry_date ASC, created_at ASC
	`
	rows, err := r.store.ConnPool.Query(ctx, sql, employeeID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.PayrollPendingEntryDetail
	for rows.Next() {
		var item domain.PayrollPendingEntryDetail
		var entryDate pgtype.Date
		var status string
		err := rows.Scan(
			&item.ID,
			&entryDate,
			&item.WorkedMinutes,
			&status,
		)
		if err != nil {
			return nil, err
		}
		item.WorkDate = conv.TimeFromPgDate(entryDate)
		item.Status = status
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SalaryRepository) ListPayoutRequestsByEmployeeAndMonth(
	ctx context.Context,
	employeeID uuid.UUID,
	salaryMonth time.Time,
) ([]domain.PayoutRequest, error) {
	sql := `
		SELECT
			id, employee_id, created_by_employee_id, requested_hours, balance_year,
			hourly_rate, gross_amount, pay_period_start, status,
			request_note, decision_note, decided_by_employee_id, paid_by_employee_id,
			requested_at, decided_at, paid_at, created_at, updated_at
		FROM leave_payout_requests
		WHERE employee_id = $1
		  AND pay_period_start >= $2
		  AND pay_period_start < $2::date + interval '1 month'
		ORDER BY requested_at DESC
	`
	rows, err := r.store.ConnPool.Query(ctx, sql, employeeID, salaryMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.PayoutRequest
	for rows.Next() {
		var (
			id                  uuid.UUID
			empID               uuid.UUID
			createdByEmpID      uuid.UUID
			requestedHours      int32
			balanceYear         int32
			hourlyRate          float64
			grossAmount         float64
			payPeriodStart      pgtype.Date
			status              string
			requestNote         *string
			decisionNote        *string
			decidedByEmployeeID *uuid.UUID
			paidByEmployeeID    *uuid.UUID
			requestedAt         pgtype.Timestamptz
			decidedAt           pgtype.Timestamptz
			paidAt              pgtype.Timestamptz
			createdAt           pgtype.Timestamptz
			updatedAt           pgtype.Timestamptz
		)
		err := rows.Scan(
			&id, &empID, &createdByEmpID, &requestedHours, &balanceYear,
			&hourlyRate, &grossAmount, &payPeriodStart, &status,
			&requestNote, &decisionNote, &decidedByEmployeeID, &paidByEmployeeID,
			&requestedAt, &decidedAt, &paidAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, toDomainPayoutRequest(
			id, empID, "", createdByEmpID,
			requestedHours, balanceYear, hourlyRate, grossAmount, payPeriodStart, status,
			requestNote, decisionNote, decidedByEmployeeID, paidByEmployeeID,
			requestedAt, decidedAt, paidAt, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type salaryTxRepo struct {
	queries *db.Queries
}

func (r *salaryTxRepo) GetPayPeriodByEmployeePeriod(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
	payrollGroup string,
) (*domain.PayPeriod, error) {
	row, err := r.queries.GetPayPeriodByEmployeePeriod(ctx, db.GetPayPeriodByEmployeePeriodParams{
		EmployeeID:   employeeID,
		PeriodStart:  conv.PgDateFromTime(periodStart),
		PeriodEnd:    conv.PgDateFromTime(periodEnd),
		PayrollGroup: payrollGroup,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayPeriodNotFound
		}
		return nil, err
	}

	model := toDomainPayPeriod(
		row.ID,
		row.EmployeeID,
		fullName(row.EmployeeFirstName, row.EmployeeLastName),
		row.PeriodStart,
		row.PeriodEnd,
		row.PayrollGroup,
		row.CutoffAt,
		string(row.Status),
		row.BaseGrossAmount,
		row.IrregularGrossAmount,
		row.GrossAmount,
		row.PaidAt,
		row.CreatedByEmployeeID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *salaryTxRepo) LockPayrollOvertimeEntries(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) ([]uuid.UUID, error) {
	ids, err := r.queries.LockPayrollOvertimeEntries(ctx, db.LockPayrollOvertimeEntriesParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: conv.PgDateFromTime(params.PeriodStart),
		CutoffDate:  conv.PgDateFromTime(params.PeriodEnd),
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *salaryTxRepo) LockPayrollPreviewWorkItems(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) ([]domain.PayrollWorkItem, error) {
	rows, err := r.queries.LockPayrollPreviewWorkItems(ctx, db.LockPayrollPreviewWorkItemsParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: conv.PgTimestamptzFromTime(params.PeriodStart),
		CutoffAt:    conv.PgTimestamptzFromTime(params.PeriodEnd),
		CutoffDate:  conv.PgDateFromTime(params.PeriodEnd),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollWorkItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPayrollWorkItemFromLock(row))
	}
	return items, nil
}

func (r *salaryTxRepo) CreatePayPeriod(
	ctx context.Context,
	params domain.ClosePayPeriodParams,
	createdByEmployeeID uuid.UUID,
	preview domain.PayrollPreview,
) (*domain.PayPeriod, error) {
	row, err := r.queries.CreatePayPeriod(ctx, db.CreatePayPeriodParams{
		EmployeeID:           params.EmployeeID,
		PeriodStart:          conv.PgDateFromTime(params.PeriodStart),
		PeriodEnd:            conv.PgDateFromTime(params.PeriodEnd),
		PayrollGroup:         params.PayrollGroup,
		CutoffAt:             conv.PgTimestamptzFromTime(params.CutoffAt),
		BaseGrossAmount:      preview.BaseGrossAmount,
		IrregularGrossAmount: preview.IrregularGrossAmount,
		GrossAmount:          preview.GrossAmount,
		CreatedByEmployeeID:  &createdByEmployeeID,
	})
	if err != nil {
		if isPayPeriodUniqueViolation(err) {
			return nil, domain.ErrPayPeriodAlreadyExists
		}
		return nil, err
	}

	model := toDomainPayPeriod(
		row.ID,
		row.EmployeeID,
		preview.EmployeeName,
		row.PeriodStart,
		row.PeriodEnd,
		row.PayrollGroup,
		row.CutoffAt,
		string(row.Status),
		row.BaseGrossAmount,
		row.IrregularGrossAmount,
		row.GrossAmount,
		row.PaidAt,
		row.CreatedByEmployeeID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *salaryTxRepo) CreatePayPeriodLineItem(
	ctx context.Context,
	payPeriodID uuid.UUID,
	item domain.PayPeriodLineItem,
) (*domain.PayPeriodLineItem, error) {
	row, err := r.queries.CreatePayPeriodLineItem(ctx, db.CreatePayPeriodLineItemParams{
		PayPeriodID:           payPeriodID,
		ScheduleID:            item.ScheduleID,
		OvertimeEntryID:       item.OvertimeEntryID,
		LeavePayoutRequestID:  item.LeavePayoutRequestID,
		ContractType:          db.EmployeeContractTypeEnum(item.ContractType),
		WorkDate:              conv.PgDateFromTime(item.WorkDate),
		LineType:              item.LineType,
		IrregularHoursProfile: db.IrregularHoursProfileEnum(item.IrregularHoursProfile),
		AppliedRatePercent:    item.AppliedRatePercent,
		MinutesWorked:         item.MinutesWorked,
		BaseAmount:            item.BaseAmount,
		PremiumAmount:         item.PremiumAmount,
		Metadata:              item.Metadata,
	})
	if err != nil {
		return nil, err
	}
	model := toDomainPayPeriodLineItem(row)
	return &model, nil
}

func (r *salaryTxRepo) AssignOvertimeEntriesToPayPeriod(
	ctx context.Context,
	payPeriodID uuid.UUID,
	overtimeEntryIDs []uuid.UUID,
) error {
	return r.queries.AssignOvertimeEntriesToPayPeriod(ctx, db.AssignOvertimeEntriesToPayPeriodParams{
		PayPeriodID:      &payPeriodID,
		OvertimeEntryIds: overtimeEntryIDs,
	})
}

func (r *salaryTxRepo) AssignSchedulesToPayPeriod(
	ctx context.Context,
	payPeriodID uuid.UUID,
	scheduleIDs []uuid.UUID,
) error {
	return r.queries.AssignSchedulesToPayPeriod(ctx, db.AssignSchedulesToPayPeriodParams{
		PayPeriodID: &payPeriodID,
		ScheduleIds: scheduleIDs,
	})
}

func (r *salaryTxRepo) AssignLeavePayoutRequestsToPayPeriod(
	ctx context.Context,
	payPeriodID uuid.UUID,
	leavePayoutRequestIDs []uuid.UUID,
) error {
	return r.queries.AssignLeavePayoutRequestsToPayPeriod(
		ctx,
		db.AssignLeavePayoutRequestsToPayPeriodParams{
			PayPeriodID:           &payPeriodID,
			LeavePayoutRequestIds: leavePayoutRequestIDs,
		},
	)
}

func (r *salaryTxRepo) GetPayPeriodForUpdate(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	row, err := r.queries.LockPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayPeriodNotFound
		}
		return nil, err
	}
	model := toDomainPayPeriod(
		row.ID,
		row.EmployeeID,
		"",
		row.PeriodStart,
		row.PeriodEnd,
		row.PayrollGroup,
		row.CutoffAt,
		string(row.Status),
		row.BaseGrossAmount,
		row.IrregularGrossAmount,
		row.GrossAmount,
		row.PaidAt,
		row.CreatedByEmployeeID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *salaryTxRepo) MarkPayPeriodPaid(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	row, err := r.queries.MarkPayPeriodPaid(ctx, payPeriodID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayPeriodNotFound
		}
		return nil, err
	}
	model := toDomainPayPeriod(
		row.ID,
		row.EmployeeID,
		"",
		row.PeriodStart,
		row.PeriodEnd,
		row.PayrollGroup,
		row.CutoffAt,
		string(row.Status),
		row.BaseGrossAmount,
		row.IrregularGrossAmount,
		row.GrossAmount,
		row.PaidAt,
		row.CreatedByEmployeeID,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *salaryTxRepo) MarkLeavePayoutRequestsPaidByPayPeriod(
	ctx context.Context,
	payPeriodID, paidByEmployeeID uuid.UUID,
) error {
	return r.queries.MarkLeavePayoutRequestsPaidByPayPeriod(
		ctx,
		db.MarkLeavePayoutRequestsPaidByPayPeriodParams{
			PaidByEmployeeID: &paidByEmployeeID,
			PayPeriodID:      &payPeriodID,
		},
	)
}

// --- Helper functions ---

func toDomainPayPeriod(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	periodStart pgtype.Date,
	periodEnd pgtype.Date,
	payrollGroup string,
	cutoffAt pgtype.Timestamptz,
	status string,
	baseGrossAmount float64,
	irregularGrossAmount float64,
	grossAmount float64,
	paidAt pgtype.Timestamptz,
	createdByEmployeeID *uuid.UUID,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.PayPeriod {
	return domain.PayPeriod{
		ID:                   id,
		EmployeeID:           employeeID,
		EmployeeName:         employeeName,
		PeriodStart:          conv.TimeFromPgDate(periodStart),
		PeriodEnd:            conv.TimeFromPgDate(periodEnd),
		PayrollGroup:         payrollGroup,
		CutoffAt:             timePtrFromPgTimestamptz(cutoffAt),
		Status:               status,
		BaseGrossAmount:      baseGrossAmount,
		IrregularGrossAmount: irregularGrossAmount,
		GrossAmount:          grossAmount,
		PaidAt:               timePtrFromPgTimestamptz(paidAt),
		CreatedByEmployeeID:  createdByEmployeeID,
		CreatedAt:            conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:            conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func toDomainPayPeriodLineItem(row db.PayPeriodLineItem) domain.PayPeriodLineItem {
	return domain.PayPeriodLineItem{
		ID:                    row.ID,
		PayPeriodID:           row.PayPeriodID,
		ScheduleID:            row.ScheduleID,
		OvertimeEntryID:       row.OvertimeEntryID,
		LeavePayoutRequestID:  row.LeavePayoutRequestID,
		ContractType:          string(row.ContractType),
		WorkDate:              conv.TimeFromPgDate(row.WorkDate),
		LineType:              row.LineType,
		IrregularHoursProfile: string(row.IrregularHoursProfile),
		AppliedRatePercent:    row.AppliedRatePercent,
		MinutesWorked:         row.MinutesWorked,
		BaseAmount:            row.BaseAmount,
		PremiumAmount:         row.PremiumAmount,
		Metadata:              row.Metadata,
		CreatedAt:             conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:             conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainPayrollWorkItem(row db.ListPayrollPreviewWorkItemsRow) domain.PayrollWorkItem {
	return domain.PayrollWorkItem{
		ID:           row.SourceID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		Label:        fmt.Sprintf("%v", row.Label),
		WorkDate:     conv.TimeFromPgDate(row.WorkDate),
		StartTime:    conv.StringFromPgTime(row.StartTimeVal),
		EndTime:      conv.StringFromPgTime(row.EndTimeVal),
		BreakMinutes: row.BreakMinutes,
		MinutesWorked: func() float64 {
			return float64(row.MinutesWorked)
		}(),
		SourceType:            row.SourceType,
		ScheduleID:            scheduleIDPtr(row.SourceType, row.ScheduleID),
		OvertimeEntryID:       row.OvertimeEntryID,
		LeavePayoutRequestID:  row.LeavePayoutRequestID,
		ContractType:          string(row.ContractType),
		ContractRate:          &row.ContractRate,
		GrossAmountOverride:   row.GrossAmountOverride,
		IrregularHoursProfile: row.IrregularHoursProfile,
	}
}

func toDomainPayrollWorkItemFromApproved(row db.ListPayrollMonthApprovedWorkItemsRow) domain.PayrollWorkItem {
	return domain.PayrollWorkItem{
		ID:           row.SourceID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		Label:        fmt.Sprintf("%v", row.Label),
		WorkDate:     conv.TimeFromPgDate(row.WorkDate),
		StartTime:    conv.StringFromPgTime(row.StartTimeVal),
		EndTime:      conv.StringFromPgTime(row.EndTimeVal),
		BreakMinutes: row.BreakMinutes,
		MinutesWorked: func() float64 {
			return float64(row.MinutesWorked)
		}(),
		SourceType:            row.SourceType,
		ScheduleID:            scheduleIDPtr(row.SourceType, row.ScheduleID),
		OvertimeEntryID:       row.OvertimeEntryID,
		LeavePayoutRequestID:  row.LeavePayoutRequestID,
		ContractType:          string(row.ContractType),
		ContractRate:          &row.ContractRate,
		GrossAmountOverride:   row.GrossAmountOverride,
		IrregularHoursProfile: row.IrregularHoursProfile,
	}
}

func toDomainPayrollWorkItemFromLock(row db.LockPayrollPreviewWorkItemsRow) domain.PayrollWorkItem {
	return domain.PayrollWorkItem{
		ID:           row.SourceID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: fullName(row.EmployeeFirstName, row.EmployeeLastName),
		Label:        fmt.Sprintf("%v", row.Label),
		WorkDate:     conv.TimeFromPgDate(row.WorkDate),
		StartTime:    conv.StringFromPgTime(row.StartTimeVal),
		EndTime:      conv.StringFromPgTime(row.EndTimeVal),
		BreakMinutes: row.BreakMinutes,
		MinutesWorked: func() float64 {
			return float64(row.MinutesWorked)
		}(),
		SourceType:            row.SourceType,
		ScheduleID:            scheduleIDPtr(row.SourceType, row.ScheduleID),
		OvertimeEntryID:       row.OvertimeEntryID,
		LeavePayoutRequestID:  row.LeavePayoutRequestID,
		ContractType:          string(row.ContractType),
		ContractRate:          &row.ContractRate,
		GrossAmountOverride:   row.GrossAmountOverride,
		IrregularHoursProfile: row.IrregularHoursProfile,
	}
}

func toDBPayPeriodStatus(value string) (db.PayPeriodStatusEnum, bool) {
	switch db.PayPeriodStatusEnum(strings.TrimSpace(value)) {
	case db.PayPeriodStatusEnumDraft, db.PayPeriodStatusEnumPaid:
		return db.PayPeriodStatusEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBPayPeriodStatusPtr(value *string) *db.PayPeriodStatusEnum {
	if value == nil {
		return nil
	}
	parsed, ok := toDBPayPeriodStatus(*value)
	if !ok {
		return nil
	}
	return enumPtr(parsed)
}

func isPayPeriodUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" &&
		strings.Contains(pgErr.ConstraintName, "pay_periods_unique_employee_period")
}

var _ domain.SalaryRepository = (*SalaryRepository)(nil)
