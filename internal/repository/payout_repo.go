package repository

import (
	"context"
	"errors"
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

type PayoutRepository struct {
	store *db.Store
}

func NewPayoutRepository(store *db.Store) domain.PayoutRepository {
	return &PayoutRepository{store: store}
}

const payrollPreviewTimeEntriesSQL = `
	SELECT
		te.id,
		te.employee_id,
		ep.first_name AS employee_first_name,
		ep.last_name AS employee_last_name,
		COALESCE(
			NULLIF(btrim(te.activity_description), ''),
			NULLIF(btrim(te.activity_category), ''),
			NULLIF(btrim(te.project_name), ''),
			NULLIF(btrim(te.client_name), ''),
			''
		) AS label,
		te.entry_date,
		te.start_time,
		te.end_time,
		te.break_minutes,
		te.hour_type::text,
		COALESCE(cc.contract_type, ep.contract_type)::text AS contract_type,
		COALESCE(cc.contract_rate, ep.contract_rate) AS contract_rate,
		COALESCE(cc.irregular_hours_profile, ep.irregular_hours_profile)::text AS irregular_hours_profile
	FROM time_entries te
	JOIN employee_profile ep ON ep.id = te.employee_id
	LEFT JOIN LATERAL (
		SELECT c.contract_type, c.contract_rate, c.irregular_hours_profile
		FROM employee_contract_changes c
		WHERE c.employee_id = te.employee_id
		  AND c.effective_from <= te.entry_date
		ORDER BY c.effective_from DESC, c.created_at DESC
		LIMIT 1
	) cc ON TRUE
	WHERE te.employee_id = $1
	  AND te.status = 'approved'::time_entry_status_enum
	  AND te.hour_type IN ('normal'::time_entry_hour_type_enum, 'overtime'::time_entry_hour_type_enum, 'travel'::time_entry_hour_type_enum, 'training'::time_entry_hour_type_enum)
	  AND te.entry_date >= $2
	  AND te.entry_date <= $3
`

const payrollMonthApprovedTimeEntriesSQL = `
	SELECT
		te.id,
		te.employee_id,
		ep.first_name AS employee_first_name,
		ep.last_name AS employee_last_name,
		COALESCE(
			NULLIF(btrim(te.activity_description), ''),
			NULLIF(btrim(te.activity_category), ''),
			NULLIF(btrim(te.project_name), ''),
			NULLIF(btrim(te.client_name), ''),
			''
		) AS label,
		te.entry_date,
		te.start_time,
		te.end_time,
		te.break_minutes,
		te.hour_type::text,
		COALESCE(cc.contract_type, ep.contract_type)::text AS contract_type,
		COALESCE(cc.contract_rate, ep.contract_rate) AS contract_rate,
		COALESCE(cc.irregular_hours_profile, ep.irregular_hours_profile)::text AS irregular_hours_profile
	FROM time_entries te
	JOIN employee_profile ep ON ep.id = te.employee_id
	LEFT JOIN LATERAL (
		SELECT c.contract_type, c.contract_rate, c.irregular_hours_profile
		FROM employee_contract_changes c
		WHERE c.employee_id = te.employee_id
		  AND c.effective_from <= te.entry_date
		ORDER BY c.effective_from DESC, c.created_at DESC
		LIMIT 1
	) cc ON TRUE
	WHERE te.employee_id = ANY($1::uuid[])
	  AND te.status = 'approved'::time_entry_status_enum
	  AND te.hour_type IN ('normal'::time_entry_hour_type_enum, 'overtime'::time_entry_hour_type_enum, 'travel'::time_entry_hour_type_enum, 'training'::time_entry_hour_type_enum)
	  AND te.entry_date >= $2
	  AND te.entry_date <= $3
	ORDER BY te.employee_id ASC, te.entry_date ASC, te.start_time ASC, te.created_at ASC
`

func (r *PayoutRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.PayoutTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&payoutTxRepo{queries: q})
	})
}

func (r *PayoutRepository) ListMyPayoutRequests(
	ctx context.Context,
	params domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListMyPayoutRequestsPaginated(ctx, db.ListMyPayoutRequestsPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBPayoutStatusPtr(params.Status),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.PayoutRequestPage{
		Items: make([]domain.PayoutRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainPayoutRequest(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.RequestedHours,
			row.BalanceYear,
			row.HourlyRate,
			row.GrossAmount,
			row.SalaryMonth,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.PaidByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.PaidAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *PayoutRepository) ListPayoutRequests(
	ctx context.Context,
	params domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListPayoutRequestsPaginated(ctx, db.ListPayoutRequestsPaginatedParams{
		Status:         toDBPayoutStatusPtr(params.Status),
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.PayoutRequestPage{
		Items: make([]domain.PayoutRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainPayoutRequest(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.RequestedHours,
			row.BalanceYear,
			row.HourlyRate,
			row.GrossAmount,
			row.SalaryMonth,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.PaidByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.PaidAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *PayoutRepository) GetPayrollPreviewEmployee(
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

func (r *PayoutRepository) ListPayrollPreviewTimeEntries(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) ([]domain.PayrollPreviewTimeEntry, error) {
	rows, err := r.store.ConnPool.Query(ctx, payrollPreviewTimeEntriesSQL+`
		AND te.paid_period_id IS NULL
		ORDER BY te.entry_date ASC, te.start_time ASC, te.created_at ASC
	`, params.EmployeeID, params.PeriodStart, params.PeriodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PayrollPreviewTimeEntry, 0)
	for rows.Next() {
		item, scanErr := scanPayrollPreviewTimeEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PayoutRepository) ListNationalHolidays(
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

func (r *PayoutRepository) GetPayPeriodByID(
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

func (r *PayoutRepository) ListPayPeriods(
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

func (r *PayoutRepository) ListPayPeriodLineItems(
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

func (r *PayoutRepository) ListPayrollMonthEmployees(
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

func (r *PayoutRepository) ListPayrollMonthEmployeesAll(
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

func (r *PayoutRepository) ListPayPeriodsByEmployeesAndRange(
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

func (r *PayoutRepository) ListPayrollMonthLockedMultiplierSummaries(
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

func (r *PayoutRepository) ListPayrollMonthApprovedTimeEntries(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollPreviewTimeEntry, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollPreviewTimeEntry{}, nil
	}

	rows, err := r.store.ConnPool.Query(ctx, payrollMonthApprovedTimeEntriesSQL, employeeIDs, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PayrollPreviewTimeEntry, 0)
	for rows.Next() {
		item, scanErr := scanPayrollPreviewTimeEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PayoutRepository) ListPayrollMonthPendingSummaries(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthPendingSummary, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollMonthPendingSummary{}, nil
	}

	rows, err := r.store.ListPayrollMonthPendingSummariesByEmployeeIDs(
		ctx,
		db.ListPayrollMonthPendingSummariesByEmployeeIDsParams{
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

func (r *PayoutRepository) ListPayrollMonthPendingEntries(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollMonthPendingEntry, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollMonthPendingEntry{}, nil
	}

	rows, err := r.store.ListPayrollMonthPendingEntriesByEmployeeIDs(
		ctx,
		db.ListPayrollMonthPendingEntriesByEmployeeIDsParams{
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

func (r *PayoutRepository) ListPendingTimeEntriesDetail(
	ctx context.Context,
	employeeID uuid.UUID,
	monthStart, monthEnd time.Time,
) ([]domain.PayrollPendingEntryDetail, error) {
	sql := `
		SELECT
			id,
			entry_date,
			start_time,
			end_time,
			break_minutes,
			status,
			GREATEST(0, (
				CASE
					WHEN end_time > start_time THEN
						EXTRACT(EPOCH FROM end_time) - EXTRACT(EPOCH FROM start_time)
					ELSE
						EXTRACT(EPOCH FROM end_time) + 86400 - EXTRACT(EPOCH FROM start_time)
				END
			) / 60 - break_minutes)::INT AS worked_minutes
		FROM time_entries
		WHERE employee_id = $1
		  AND entry_date >= $2
		  AND entry_date <= $3
		  AND status IN ('draft'::time_entry_status_enum, 'submitted'::time_entry_status_enum)
		  AND hour_type IN ('normal'::time_entry_hour_type_enum, 'overtime'::time_entry_hour_type_enum, 'travel'::time_entry_hour_type_enum, 'training'::time_entry_hour_type_enum)
		ORDER BY entry_date ASC, start_time ASC
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
		var startTime pgtype.Time
		var endTime pgtype.Time
		var status string
		err := rows.Scan(
			&item.ID,
			&entryDate,
			&startTime,
			&endTime,
			&item.BreakMinutes,
			&status,
			&item.WorkedMinutes,
		)
		if err != nil {
			return nil, err
		}
		item.WorkDate = conv.TimeFromPgDate(entryDate)
		item.StartTime = conv.StringFromPgTime(startTime)
		item.EndTime = conv.StringFromPgTime(endTime)
		item.Status = status
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PayoutRepository) ListPayoutRequestsByEmployeeAndMonth(
	ctx context.Context,
	employeeID uuid.UUID,
	salaryMonth time.Time,
) ([]domain.PayoutRequest, error) {
	sql := `
		SELECT
			id, employee_id, created_by_employee_id, requested_hours, balance_year,
			hourly_rate, gross_amount, salary_month, status,
			request_note, decision_note, decided_by_employee_id, paid_by_employee_id,
			requested_at, decided_at, paid_at, created_at, updated_at
		FROM leave_payout_requests
		WHERE employee_id = $1
		  AND salary_month = $2
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
			salMonth            pgtype.Date
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
			&hourlyRate, &grossAmount, &salMonth, &status,
			&requestNote, &decisionNote, &decidedByEmployeeID, &paidByEmployeeID,
			&requestedAt, &decidedAt, &paidAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, toDomainPayoutRequest(
			id, empID, "", createdByEmpID,
			requestedHours, balanceYear, hourlyRate, grossAmount, salMonth, status,
			requestNote, decisionNote, decidedByEmployeeID, paidByEmployeeID,
			requestedAt, decidedAt, paidAt, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PayoutRepository) GetLeaveBalanceExtraRemaining(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) (int32, error) {
	sql := `
		SELECT COALESCE(
			(SELECT extra_total_hours - extra_used_hours
			 FROM leave_balances
			 WHERE employee_id = $1 AND year = $2),
			0
		)::INT AS extra_remaining
	`
	var extraRemaining int32
	err := r.store.ConnPool.QueryRow(ctx, sql, employeeID, year).Scan(&extraRemaining)
	if err != nil {
		return 0, err
	}
	return extraRemaining, nil
}

type payoutTxRepo struct {
	queries *db.Queries
}

func (r *payoutTxRepo) GetEmployeePayoutContract(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.PayoutContract, error) {
	row, err := r.queries.GetEmployeePayoutContract(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}

	return &domain.PayoutContract{
		ContractType: string(row.ContractType),
		ContractRate: row.ContractRate,
	}, nil
}

func (r *payoutTxRepo) EnsureLeaveBalanceForYear(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) error {
	return r.queries.EnsureLeaveBalanceForYear(ctx, db.EnsureLeaveBalanceForYearParams{
		EmployeeID: employeeID,
		Year:       year,
	})
}

func (r *payoutTxRepo) GetPayoutBalanceForUpdate(
	ctx context.Context,
	employeeID uuid.UUID,
	year int32,
) (*domain.PayoutBalanceSnapshot, error) {
	row, err := r.queries.LockLeaveBalanceByEmployeeYear(
		ctx,
		db.LockLeaveBalanceByEmployeeYearParams{
			EmployeeID: employeeID,
			Year:       year,
		},
	)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	return &domain.PayoutBalanceSnapshot{
		LeaveBalanceID: row.ID,
		ExtraRemaining: row.ExtraTotalHours - row.ExtraUsedHours,
	}, nil
}

func (r *payoutTxRepo) CreatePayoutRequest(
	ctx context.Context,
	params domain.CreatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.CreatePayoutRequest(ctx, db.CreatePayoutRequestParams{
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: params.CreatedByEmployeeID,
		RequestedHours:      params.RequestedHours,
		BalanceYear:         params.BalanceYear,
		HourlyRate:          params.HourlyRate,
		GrossAmount:         params.GrossAmount,
		RequestNote:         params.RequestNote,
	})
	if err != nil {
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) GetPayoutRequestForUpdate(
	ctx context.Context,
	payoutRequestID uuid.UUID,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.LockPayoutRequestByID(ctx, payoutRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) ApprovePayoutRequest(
	ctx context.Context,
	payoutRequestID, decidedByEmployeeID uuid.UUID,
	salaryMonth time.Time,
	decisionNote *string,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.ApprovePayoutRequest(ctx, db.ApprovePayoutRequestParams{
		ID:                  payoutRequestID,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
		SalaryMonth:         conv.PgDateFromTime(salaryMonth),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) RejectPayoutRequest(
	ctx context.Context,
	payoutRequestID, decidedByEmployeeID uuid.UUID,
	decisionNote *string,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.RejectPayoutRequest(ctx, db.RejectPayoutRequestParams{
		ID:                  payoutRequestID,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) MarkPayoutRequestPaid(
	ctx context.Context,
	payoutRequestID, paidByEmployeeID uuid.UUID,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.MarkPayoutRequestPaid(ctx, db.MarkPayoutRequestPaidParams{
		ID:               payoutRequestID,
		PaidByEmployeeID: &paidByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) ApplyLeaveBalanceDeduction(
	ctx context.Context,
	balanceID uuid.UUID,
	extraHours, legalHours int32,
) (*domain.LeaveBalance, error) {
	row, err := r.queries.ApplyLeaveBalanceDeduction(ctx, db.ApplyLeaveBalanceDeductionParams{
		ID:         balanceID,
		ExtraHours: extraHours,
		LegalHours: legalHours,
	})
	if err != nil {
		return nil, err
	}
	model := toDomainLeaveBalance(
		row.ID,
		row.EmployeeID,
		"",
		row.Year,
		row.LegalTotalHours,
		row.ExtraTotalHours,
		row.LegalUsedHours,
		row.ExtraUsedHours,
		nil,
		nil,
		nil,
		nil,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *payoutTxRepo) GetPayPeriodByEmployeePeriod(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) (*domain.PayPeriod, error) {
	row, err := r.queries.GetPayPeriodByEmployeePeriod(ctx, db.GetPayPeriodByEmployeePeriodParams{
		EmployeeID:  employeeID,
		PeriodStart: conv.PgDateFromTime(periodStart),
		PeriodEnd:   conv.PgDateFromTime(periodEnd),
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

func (r *payoutTxRepo) LockPayrollPreviewTimeEntries(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) ([]domain.PayrollPreviewTimeEntry, error) {
	rows, err := r.queries.LockPayrollPreviewTimeEntries(
		ctx,
		db.LockPayrollPreviewTimeEntriesParams{
			EmployeeID:  params.EmployeeID,
			PeriodStart: conv.PgDateFromTime(params.PeriodStart),
			PeriodEnd:   conv.PgDateFromTime(params.PeriodEnd),
		},
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollPreviewTimeEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PayrollPreviewTimeEntry{
			ID:                    row.ID,
			EmployeeID:            row.EmployeeID,
			EmployeeName:          fullName(row.EmployeeFirstName, row.EmployeeLastName),
			EntryDate:             conv.TimeFromPgDate(row.EntryDate),
			StartTime:             conv.StringFromPgTime(row.StartTime),
			EndTime:               conv.StringFromPgTime(row.EndTime),
			BreakMinutes:          row.BreakMinutes,
			HourType:              string(row.HourType),
			ContractType:          string(row.ContractType),
			ContractRate:          row.ContractRate,
			IrregularHoursProfile: string(row.IrregularHoursProfile),
		})
	}

	return items, nil
}

func (r *payoutTxRepo) CreatePayPeriod(
	ctx context.Context,
	params domain.ClosePayPeriodParams,
	createdByEmployeeID uuid.UUID,
	preview domain.PayrollPreview,
) (*domain.PayPeriod, error) {
	row, err := r.queries.CreatePayPeriod(ctx, db.CreatePayPeriodParams{
		EmployeeID:           params.EmployeeID,
		PeriodStart:          conv.PgDateFromTime(params.PeriodStart),
		PeriodEnd:            conv.PgDateFromTime(params.PeriodEnd),
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

func (r *payoutTxRepo) CreatePayPeriodLineItem(
	ctx context.Context,
	payPeriodID uuid.UUID,
	item domain.PayPeriodLineItem,
) (*domain.PayPeriodLineItem, error) {
	row, err := r.queries.CreatePayPeriodLineItem(ctx, db.CreatePayPeriodLineItemParams{
		PayPeriodID:           payPeriodID,
		TimeEntryID:           item.TimeEntryID,
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

func (r *payoutTxRepo) AssignTimeEntriesToPayPeriod(
	ctx context.Context,
	payPeriodID uuid.UUID,
	timeEntryIDs []uuid.UUID,
) error {
	return r.queries.AssignTimeEntriesToPayPeriod(ctx, db.AssignTimeEntriesToPayPeriodParams{
		PayPeriodID:  &payPeriodID,
		TimeEntryIds: timeEntryIDs,
	})
}

func (r *payoutTxRepo) GetPayPeriodForUpdate(
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

func (r *payoutTxRepo) MarkPayPeriodPaid(
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

func toDomainPayoutRequestFromRow(row db.LeavePayoutRequest) domain.PayoutRequest {
	return toDomainPayoutRequest(
		row.ID,
		row.EmployeeID,
		"",
		row.CreatedByEmployeeID,
		row.RequestedHours,
		row.BalanceYear,
		row.HourlyRate,
		row.GrossAmount,
		row.SalaryMonth,
		string(row.Status),
		row.RequestNote,
		row.DecisionNote,
		row.DecidedByEmployeeID,
		row.PaidByEmployeeID,
		row.RequestedAt,
		row.DecidedAt,
		row.PaidAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func stringFromDBValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		return ""
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayrollPreviewTimeEntry(row rowScanner) (domain.PayrollPreviewTimeEntry, error) {
	var (
		item                  domain.PayrollPreviewTimeEntry
		firstName             string
		lastName              string
		entryDate             pgtype.Date
		startTime             pgtype.Time
		endTime               pgtype.Time
		hourType              string
		contractType          string
		irregularHoursProfile string
	)
	err := row.Scan(
		&item.ID,
		&item.EmployeeID,
		&firstName,
		&lastName,
		&item.Label,
		&entryDate,
		&startTime,
		&endTime,
		&item.BreakMinutes,
		&hourType,
		&contractType,
		&item.ContractRate,
		&irregularHoursProfile,
	)
	if err != nil {
		return domain.PayrollPreviewTimeEntry{}, err
	}
	item.EmployeeName = fullName(firstName, lastName)
	item.EntryDate = conv.TimeFromPgDate(entryDate)
	item.StartTime = conv.StringFromPgTime(startTime)
	item.EndTime = conv.StringFromPgTime(endTime)
	item.HourType = hourType
	item.ContractType = contractType
	item.IrregularHoursProfile = irregularHoursProfile
	return item, nil
}

func toDomainPayoutRequest(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	createdByEmployeeID uuid.UUID,
	requestedHours int32,
	balanceYear int32,
	hourlyRate float64,
	grossAmount float64,
	salaryMonth pgtype.Date,
	status string,
	requestNote *string,
	decisionNote *string,
	decidedByEmployeeID *uuid.UUID,
	paidByEmployeeID *uuid.UUID,
	requestedAt pgtype.Timestamptz,
	decidedAt pgtype.Timestamptz,
	paidAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.PayoutRequest {
	return domain.PayoutRequest{
		ID:                  id,
		EmployeeID:          employeeID,
		EmployeeName:        employeeName,
		CreatedByEmployeeID: createdByEmployeeID,
		RequestedHours:      requestedHours,
		BalanceYear:         balanceYear,
		HourlyRate:          hourlyRate,
		GrossAmount:         grossAmount,
		SalaryMonth:         conv.TimePtrFromPgDate(salaryMonth),
		Status:              status,
		RequestNote:         requestNote,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: decidedByEmployeeID,
		PaidByEmployeeID:    paidByEmployeeID,
		RequestedAt:         conv.TimeFromPgTimestamptz(requestedAt),
		DecidedAt:           timePtrFromPgTimestamptz(decidedAt),
		PaidAt:              timePtrFromPgTimestamptz(paidAt),
		CreatedAt:           conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:           conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func toDomainPayPeriod(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	periodStart pgtype.Date,
	periodEnd pgtype.Date,
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
		TimeEntryID:           row.TimeEntryID,
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

func toDBPayoutStatus(value string) (db.PayoutRequestStatusEnum, bool) {
	switch db.PayoutRequestStatusEnum(strings.TrimSpace(value)) {
	case db.PayoutRequestStatusEnumPending,
		db.PayoutRequestStatusEnumApproved,
		db.PayoutRequestStatusEnumRejected,
		db.PayoutRequestStatusEnumPaid:
		return db.PayoutRequestStatusEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBPayoutStatusPtr(value *string) *db.PayoutRequestStatusEnum {
	if value == nil {
		return nil
	}
	parsed, ok := toDBPayoutStatus(*value)
	if !ok {
		return nil
	}
	return enumPtr(parsed)
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

var _ domain.PayoutRepository = (*PayoutRepository)(nil)
