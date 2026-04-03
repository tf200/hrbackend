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

type PayoutRepository struct {
	store *db.Store
}

func NewPayoutRepository(store *db.Store) domain.PayoutRepository {
	return &PayoutRepository{store: store}
}

func (r *PayoutRepository) WithTx(ctx context.Context, fn func(tx domain.PayoutTxRepository) error) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&payoutTxRepo{queries: q})
	})
}

func (r *PayoutRepository) ListMyPayoutRequests(ctx context.Context, params domain.ListMyPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListMyPayoutRequestsPaginated(ctx, db.ListMyPayoutRequestsPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBNullPayoutStatus(params.Status),
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

func (r *PayoutRepository) ListPayoutRequests(ctx context.Context, params domain.ListPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListPayoutRequestsPaginated(ctx, db.ListPayoutRequestsPaginatedParams{
		Status:         toDBNullPayoutStatus(params.Status),
		EmployeeSearch: trimStringPtr(params.EmployeeSearch),
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

func (r *PayoutRepository) GetPayrollPreviewEmployee(ctx context.Context, employeeID uuid.UUID) (*domain.EmployeeDetail, error) {
	row, err := r.store.GetEmployeeProfileByID(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row), nil
}

func (r *PayoutRepository) ListPayrollPreviewTimeEntries(ctx context.Context, params domain.PayrollPreviewParams) ([]domain.PayrollPreviewTimeEntry, error) {
	rows, err := r.store.ListPayrollPreviewTimeEntries(ctx, db.ListPayrollPreviewTimeEntriesParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: conv.PgDateFromTime(params.PeriodStart),
		PeriodEnd:   conv.PgDateFromTime(params.PeriodEnd),
	})
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

func (r *PayoutRepository) ListNationalHolidays(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]domain.NationalHoliday, error) {
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

func (r *PayoutRepository) GetPayPeriodByID(ctx context.Context, payPeriodID uuid.UUID) (*domain.PayPeriod, error) {
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

func (r *PayoutRepository) ListPayPeriods(ctx context.Context, params domain.ListPayPeriodsParams) (*domain.PayPeriodPage, error) {
	rows, err := r.store.ListPayPeriodsPaginated(ctx, db.ListPayPeriodsPaginatedParams{
		Status:         toDBNullPayPeriodStatus(params.Status),
		EmployeeSearch: trimStringPtr(params.EmployeeSearch),
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

func (r *PayoutRepository) ListPayPeriodLineItems(ctx context.Context, payPeriodID uuid.UUID) ([]domain.PayPeriodLineItem, error) {
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

func (r *PayoutRepository) ListPayrollMonthEmployees(ctx context.Context, params domain.PayrollMonthSummaryParams, monthStart, monthEnd time.Time) ([]domain.PayrollMonthEmployee, int64, error) {
	rows, err := r.store.ListPayrollMonthEmployeesPaginated(ctx, db.ListPayrollMonthEmployeesPaginatedParams{
		EmployeeSearch: trimStringPtr(params.EmployeeSearch),
		Offset:         params.Offset,
		Limit:          params.Limit,
		MonthStart:     conv.PgDateFromTime(monthStart),
		MonthEnd:       conv.PgDateFromTime(monthEnd),
	})
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

func (r *PayoutRepository) ListPayPeriodsByEmployeesAndRange(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayPeriod, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayPeriod{}, nil
	}

	rows, err := r.store.ListPayPeriodsByEmployeeIDsAndRange(ctx, db.ListPayPeriodsByEmployeeIDsAndRangeParams{
		EmployeeIds: employeeIDs,
		MonthStart:  conv.PgDateFromTime(monthStart),
		MonthEnd:    conv.PgDateFromTime(monthEnd),
	})
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

func (r *PayoutRepository) ListPayrollMonthLockedMultiplierSummaries(ctx context.Context, payPeriodIDs []uuid.UUID) ([]domain.PayrollLockedMultiplierSummary, error) {
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

func (r *PayoutRepository) ListPayrollMonthApprovedTimeEntries(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollPreviewTimeEntry, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollPreviewTimeEntry{}, nil
	}

	rows, err := r.store.ListPayrollMonthApprovedTimeEntriesByEmployeeIDs(ctx, db.ListPayrollMonthApprovedTimeEntriesByEmployeeIDsParams{
		EmployeeIds: employeeIDs,
		MonthStart:  conv.PgDateFromTime(monthStart),
		MonthEnd:    conv.PgDateFromTime(monthEnd),
	})
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

func (r *PayoutRepository) ListPayrollMonthPendingSummaries(ctx context.Context, employeeIDs []uuid.UUID, monthStart, monthEnd time.Time) ([]domain.PayrollMonthPendingSummary, error) {
	if len(employeeIDs) == 0 {
		return []domain.PayrollMonthPendingSummary{}, nil
	}

	rows, err := r.store.ListPayrollMonthPendingSummariesByEmployeeIDs(ctx, db.ListPayrollMonthPendingSummariesByEmployeeIDsParams{
		EmployeeIds: employeeIDs,
		MonthStart:  conv.PgDateFromTime(monthStart),
		MonthEnd:    conv.PgDateFromTime(monthEnd),
	})
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

type payoutTxRepo struct {
	queries *db.Queries
}

func (r *payoutTxRepo) GetEmployeePayoutContract(ctx context.Context, employeeID uuid.UUID) (*domain.PayoutContract, error) {
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

func (r *payoutTxRepo) EnsureLeaveBalanceForYear(ctx context.Context, employeeID uuid.UUID, year int32) error {
	return r.queries.EnsureLeaveBalanceForYear(ctx, db.EnsureLeaveBalanceForYearParams{
		EmployeeID: employeeID,
		Year:       year,
	})
}

func (r *payoutTxRepo) GetPayoutBalanceForUpdate(ctx context.Context, employeeID uuid.UUID, year int32) (*domain.PayoutBalanceSnapshot, error) {
	row, err := r.queries.LockLeaveBalanceByEmployeeYear(ctx, db.LockLeaveBalanceByEmployeeYearParams{
		EmployeeID: employeeID,
		Year:       year,
	})
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

func (r *payoutTxRepo) CreatePayoutRequest(ctx context.Context, params domain.CreatePayoutRequestTxParams) (*domain.PayoutRequest, error) {
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

func (r *payoutTxRepo) GetPayoutRequestForUpdate(ctx context.Context, payoutRequestID uuid.UUID) (*domain.PayoutRequest, error) {
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

func (r *payoutTxRepo) ApprovePayoutRequest(ctx context.Context, payoutRequestID, decidedByEmployeeID uuid.UUID, salaryMonth time.Time, decisionNote *string) (*domain.PayoutRequest, error) {
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

func (r *payoutTxRepo) RejectPayoutRequest(ctx context.Context, payoutRequestID, decidedByEmployeeID uuid.UUID, decisionNote *string) (*domain.PayoutRequest, error) {
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

func (r *payoutTxRepo) MarkPayoutRequestPaid(ctx context.Context, payoutRequestID, paidByEmployeeID uuid.UUID) (*domain.PayoutRequest, error) {
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

func (r *payoutTxRepo) ApplyLeaveBalanceDeduction(ctx context.Context, balanceID uuid.UUID, extraHours, legalHours int32) (*domain.LeaveBalance, error) {
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

func (r *payoutTxRepo) GetPayPeriodByEmployeePeriod(ctx context.Context, employeeID uuid.UUID, periodStart, periodEnd time.Time) (*domain.PayPeriod, error) {
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

func (r *payoutTxRepo) LockPayrollPreviewTimeEntries(ctx context.Context, params domain.PayrollPreviewParams) ([]domain.PayrollPreviewTimeEntry, error) {
	rows, err := r.queries.LockPayrollPreviewTimeEntries(ctx, db.LockPayrollPreviewTimeEntriesParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: conv.PgDateFromTime(params.PeriodStart),
		PeriodEnd:   conv.PgDateFromTime(params.PeriodEnd),
	})
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

func (r *payoutTxRepo) CreatePayPeriod(ctx context.Context, params domain.ClosePayPeriodParams, createdByEmployeeID uuid.UUID, preview domain.PayrollPreview) (*domain.PayPeriod, error) {
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

func (r *payoutTxRepo) CreatePayPeriodLineItem(ctx context.Context, payPeriodID uuid.UUID, item domain.PayPeriodLineItem) (*domain.PayPeriodLineItem, error) {
	row, err := r.queries.CreatePayPeriodLineItem(ctx, db.CreatePayPeriodLineItemParams{
		PayPeriodID:           payPeriodID,
		TimeEntryID:           item.TimeEntryID,
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

func (r *payoutTxRepo) AssignTimeEntriesToPayPeriod(ctx context.Context, payPeriodID uuid.UUID, timeEntryIDs []uuid.UUID) error {
	return r.queries.AssignTimeEntriesToPayPeriod(ctx, db.AssignTimeEntriesToPayPeriodParams{
		PayPeriodID:  &payPeriodID,
		TimeEntryIds: timeEntryIDs,
	})
}

func (r *payoutTxRepo) GetPayPeriodForUpdate(ctx context.Context, payPeriodID uuid.UUID) (*domain.PayPeriod, error) {
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

func (r *payoutTxRepo) MarkPayPeriodPaid(ctx context.Context, payPeriodID uuid.UUID) (*domain.PayPeriod, error) {
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

func toDBNullPayoutStatus(value *string) db.NullPayoutRequestStatusEnum {
	if value == nil {
		return db.NullPayoutRequestStatusEnum{}
	}
	parsed, ok := toDBPayoutStatus(*value)
	if !ok {
		return db.NullPayoutRequestStatusEnum{}
	}
	return db.NullPayoutRequestStatusEnum{
		PayoutRequestStatusEnum: parsed,
		Valid:                   true,
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

func toDBNullPayPeriodStatus(value *string) db.NullPayPeriodStatusEnum {
	if value == nil {
		return db.NullPayPeriodStatusEnum{}
	}
	parsed, ok := toDBPayPeriodStatus(*value)
	if !ok {
		return db.NullPayPeriodStatusEnum{}
	}
	return db.NullPayPeriodStatusEnum{
		PayPeriodStatusEnum: parsed,
		Valid:               true,
	}
}

func isPayPeriodUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "pay_periods_unique_employee_period")
}

var _ domain.PayoutRepository = (*PayoutRepository)(nil)
