package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/pkg/ptr"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PayoutService struct {
	repository domain.PayoutRepository
	logger     domain.Logger
}

func NewPayoutService(
	repository domain.PayoutRepository,
	logger domain.Logger,
) domain.PayoutService {
	return &PayoutService{
		repository: repository,
		logger:     logger,
	}
}

func (s *PayoutService) CreatePayoutRequest(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreatePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.RequestedHours <= 0 || params.BalanceYear < 2000 || params.BalanceYear > 2100 {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	params.EmployeeID = actorEmployeeID
	params.CreatedByEmployeeID = actorEmployeeID

	var created *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		contract, err := tx.GetEmployeePayoutContract(ctx, params.EmployeeID)
		if err != nil {
			return err
		}
		if contract.ContractType != "loondienst" {
			return domain.ErrPayoutRequestInvalidRequest
		}
		if contract.ContractRate == nil || *contract.ContractRate <= 0 {
			return domain.ErrPayoutRequestInvalidRequest
		}

		if err := tx.EnsureLeaveBalanceForYear(
			ctx,
			params.EmployeeID,
			params.BalanceYear,
		); err != nil {
			return err
		}
		balance, err := tx.GetPayoutBalanceForUpdate(ctx, params.EmployeeID, params.BalanceYear)
		if err != nil {
			return err
		}
		if balance.ExtraRemaining < params.RequestedHours {
			return domain.ErrPayoutRequestInsufficientHours
		}

		hourlyRate := roundCurrency(*contract.ContractRate)
		grossAmount := roundCurrency(float64(params.RequestedHours) * hourlyRate)

		created, err = tx.CreatePayoutRequest(ctx, domain.CreatePayoutRequestTxParams{
			EmployeeID:          params.EmployeeID,
			CreatedByEmployeeID: params.CreatedByEmployeeID,
			RequestedHours:      params.RequestedHours,
			BalanceYear:         params.BalanceYear,
			HourlyRate:          hourlyRate,
			GrossAmount:         grossAmount,
			RequestNote:         params.RequestNote,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *PayoutService) DecidePayoutRequestByAdmin(
	ctx context.Context,
	adminEmployeeID, payoutRequestID uuid.UUID,
	params domain.DecidePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	decision := strings.ToLower(strings.TrimSpace(params.Decision))
	if decision != "approve" && decision != "reject" {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if decision == "approve" && (params.SalaryMonth == nil || params.SalaryMonth.IsZero()) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayoutRequestStatusPending {
			return domain.ErrPayoutRequestStateInvalid
		}

		if decision == "approve" {
			if err := tx.EnsureLeaveBalanceForYear(
				ctx,
				current.EmployeeID,
				current.BalanceYear,
			); err != nil {
				return err
			}
			balance, err := tx.GetPayoutBalanceForUpdate(
				ctx,
				current.EmployeeID,
				current.BalanceYear,
			)
			if err != nil {
				return err
			}
			if balance.ExtraRemaining < current.RequestedHours {
				return domain.ErrPayoutRequestInsufficientHours
			}
			if _, err := tx.ApplyLeaveBalanceDeduction(
				ctx,
				balance.LeaveBalanceID,
				current.RequestedHours,
				0,
			); err != nil {
				return err
			}

			updated, err = tx.ApprovePayoutRequest(
				ctx,
				payoutRequestID,
				adminEmployeeID,
				params.SalaryMonth.UTC(),
				params.DecisionNote,
			)
			return err
		}

		updated, err = tx.RejectPayoutRequest(
			ctx,
			payoutRequestID,
			adminEmployeeID,
			params.DecisionNote,
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) MarkPayoutRequestPaidByAdmin(
	ctx context.Context,
	adminEmployeeID, payoutRequestID uuid.UUID,
) (*domain.PayoutRequest, error) {
	if adminEmployeeID == uuid.Nil || payoutRequestID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayoutRequest
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayoutRequestForUpdate(ctx, payoutRequestID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayoutRequestStatusApproved {
			return domain.ErrPayoutRequestStateInvalid
		}

		updated, err = tx.MarkPayoutRequestPaid(ctx, payoutRequestID, adminEmployeeID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) ListMyPayoutRequests(
	ctx context.Context,
	params domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListMyPayoutRequests(ctx, params)
}

func (s *PayoutService) ListPayoutRequests(
	ctx context.Context,
	params domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListPayoutRequests(ctx, params)
}

func (s *PayoutService) PreviewMyPayroll(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) (*domain.PayrollPreview, error) {
	return s.PreviewPayroll(ctx, domain.PayrollPreviewParams{
		EmployeeID:  actorEmployeeID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
}

func (s *PayoutService) GetORTRules(_ context.Context) (*domain.ORTRulesResponse, error) {
	loondienst := "loondienst"
	nonLoondienst := "non_loondienst"
	roster := domain.IrregularHoursProfileRoster
	nonRoster := domain.IrregularHoursProfileNonRoster

	return &domain.ORTRulesResponse{
		Rules: []domain.ORTRule{
			{
				Order:        1,
				RatePercent:  45,
				Label:        "Public holiday",
				Description:  "Public holidays apply 45% ORT for all hours.",
				ContractType: loondienst,
				DayType:      "public_holiday",
			},
			{
				Order:        2,
				RatePercent:  45,
				Label:        "Sunday",
				Description:  "Sundays apply 45% ORT for all hours.",
				ContractType: loondienst,
				DayType:      "sunday",
			},
			{
				Order:        3,
				RatePercent:  45,
				Label:        "Night hours",
				Description:  "Any day from 22:00 to before 06:00 applies 45% ORT.",
				ContractType: loondienst,
				DayType:      "any",
				TimeFrom:     ptr.String("22:00"),
				TimeTo:       ptr.String("06:00"),
			},
			{
				Order:        4,
				RatePercent:  30,
				Label:        "Saturday daytime",
				Description:  "Saturdays from 06:00 to before 22:00 apply 30% ORT.",
				ContractType: loondienst,
				DayType:      "saturday",
				TimeFrom:     ptr.String("06:00"),
				TimeTo:       ptr.String("22:00"),
			},
			{
				Order:                 5,
				RatePercent:           25,
				Label:                 "Roster early morning",
				Description:           "Roster profile from 06:00 to before 07:00 applies 25% ORT.",
				ContractType:          loondienst,
				IrregularHoursProfile: &roster,
				DayType:               "any",
				TimeFrom:              ptr.String("06:00"),
				TimeTo:                ptr.String("07:00"),
			},
			{
				Order:                 6,
				RatePercent:           25,
				Label:                 "Roster evening",
				Description:           "Roster profile from 19:00 to before 22:00 applies 25% ORT.",
				ContractType:          loondienst,
				IrregularHoursProfile: &roster,
				DayType:               "any",
				TimeFrom:              ptr.String("19:00"),
				TimeTo:                ptr.String("22:00"),
			},
			{
				Order:                 7,
				RatePercent:           25,
				Label:                 "Non-roster evening",
				Description:           "Non-roster profile from 20:00 to before 22:00 applies 25% ORT.",
				ContractType:          loondienst,
				IrregularHoursProfile: &nonRoster,
				DayType:               "any",
				TimeFrom:              ptr.String("20:00"),
				TimeTo:                ptr.String("22:00"),
			},
			{
				Order:        8,
				RatePercent:  0,
				Label:        "Default loondienst fallback",
				Description:  "Hours not covered by any ORT window apply 0% ORT for loondienst.",
				ContractType: loondienst,
				DayType:      "any",
			},
			{
				Order:        9,
				RatePercent:  0,
				Label:        "Non-loondienst fallback",
				Description:  "Non-loondienst contract types, including ZZP, apply 0% ORT.",
				ContractType: nonLoondienst,
				DayType:      "any",
			},
		},
	}, nil
}

func (s *PayoutService) PreviewPayroll(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) (*domain.PayrollPreview, error) {
	normalized, err := normalizePayrollPreviewParams(params)
	if err != nil {
		return nil, err
	}

	employee, err := s.repository.GetPayrollPreviewEmployee(ctx, normalized.EmployeeID)
	if err != nil {
		if err == domain.ErrEmployeeNotFound {
			return nil, err
		}
		s.logError(
			ctx,
			"PreviewPayroll",
			"failed to get employee",
			err,
			zap.String("employee_id", normalized.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to get employee for payroll preview: %w", err)
	}

	entries, err := s.repository.ListPayrollPreviewTimeEntries(ctx, normalized)
	if err != nil {
		s.logError(ctx, "PreviewPayroll", "failed to list payroll time entries", err,
			zap.String("employee_id", normalized.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to list payroll time entries: %w", err)
	}

	return s.buildPayrollPreview(ctx, employee, normalized, entries)
}

func (s *PayoutService) ClosePayPeriod(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.ClosePayPeriodParams,
) (*domain.PayPeriod, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	normalized, err := normalizeClosePayPeriodParams(params)
	if err != nil {
		return nil, err
	}

	employee, err := s.repository.GetPayrollPreviewEmployee(ctx, normalized.EmployeeID)
	if err != nil {
		if err == domain.ErrEmployeeNotFound {
			return nil, err
		}
		s.logError(
			ctx,
			"ClosePayPeriod",
			"failed to get employee",
			err,
			zap.String("employee_id", normalized.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to get employee for pay period close: %w", err)
	}

	var result *domain.PayPeriod
	err = s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		existing, err := tx.GetPayPeriodByEmployeePeriod(
			ctx,
			normalized.EmployeeID,
			normalized.PeriodStart,
			normalized.PeriodEnd,
		)
		if err != nil && err != domain.ErrPayPeriodNotFound {
			return err
		}
		if existing != nil {
			return domain.ErrPayPeriodAlreadyExists
		}

		entries, err := tx.LockPayrollPreviewTimeEntries(ctx, domain.PayrollPreviewParams{
			EmployeeID:  normalized.EmployeeID,
			PeriodStart: normalized.PeriodStart,
			PeriodEnd:   normalized.PeriodEnd,
		})
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			return domain.ErrPayPeriodNoEntries
		}

		preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
			EmployeeID:  normalized.EmployeeID,
			PeriodStart: normalized.PeriodStart,
			PeriodEnd:   normalized.PeriodEnd,
		}, entries)
		if err != nil {
			return err
		}

		created, err := tx.CreatePayPeriod(ctx, normalized, adminEmployeeID, *preview)
		if err != nil {
			return err
		}

		created.EmployeeName = strings.TrimSpace(employee.FirstName + " " + employee.LastName)
		created.LineItems = make([]domain.PayPeriodLineItem, 0, len(preview.LineItems))
		for _, item := range preview.LineItems {
			createdLine, err := tx.CreatePayPeriodLineItem(
				ctx,
				created.ID,
				buildPayPeriodLineItem(item, entries),
			)
			if err != nil {
				return err
			}
			created.LineItems = append(created.LineItems, *createdLine)
		}

		timeEntryIDs := uniquePreviewTimeEntryIDs(preview.LineItems)
		if len(timeEntryIDs) == 0 {
			return domain.ErrPayPeriodNoEntries
		}
		if err := tx.AssignTimeEntriesToPayPeriod(ctx, created.ID, timeEntryIDs); err != nil {
			return err
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PayoutService) GetPayPeriodByID(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	if payPeriodID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	period, err := s.repository.GetPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}

	lineItems, err := s.repository.ListPayPeriodLineItems(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}
	period.LineItems = lineItems
	return period, nil
}

func (s *PayoutService) loadPayPeriodWithLineItems(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	period, err := s.repository.GetPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}

	lineItems, err := s.repository.ListPayPeriodLineItems(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}
	period.LineItems = lineItems
	return period, nil
}

func (s *PayoutService) ListPayPeriods(
	ctx context.Context,
	params domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	if params.Status != nil && !isValidPayPeriodStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListPayPeriods(ctx, params)
}

func (s *PayoutService) GetPayrollMonthSummary(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthSummaryPage, error) {
	normalized, monthStart, monthEnd, isCurrentMonth, err := normalizePayrollMonthSummaryParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	employees, totalCount, err := s.repository.ListPayrollMonthEmployees(
		ctx,
		normalized,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthSummary", "failed to list payroll month employees", err)
		return nil, fmt.Errorf("failed to list payroll month employees: %w", err)
	}

	if len(employees) == 0 {
		return &domain.PayrollMonthSummaryPage{
			Items:      []domain.PayrollMonthSummaryRow{},
			TotalCount: totalCount,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	lockedPayPeriods, err := s.repository.ListPayPeriodsByEmployeesAndRange(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthSummary", "failed to list locked pay periods", err)
		return nil, fmt.Errorf("failed to list pay periods for payroll month summary: %w", err)
	}
	lockedByEmployee := make(map[uuid.UUID]domain.PayPeriod, len(lockedPayPeriods))
	for _, payPeriod := range lockedPayPeriods {
		lockedByEmployee[payPeriod.EmployeeID] = payPeriod
	}
	lockedSnapshotByPeriod, err := s.buildLockedSnapshotMap(
		ctx,
		lockedPayPeriods,
		normalized.ContractType,
	)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthSummary", "failed to build locked payroll summaries", err)
		return nil, fmt.Errorf("failed to build locked payroll summaries: %w", err)
	}

	approvedEntries, err := s.repository.ListPayrollMonthApprovedTimeEntries(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetPayrollMonthSummary",
			"failed to list approved payroll time entries",
			err,
		)
		return nil, fmt.Errorf("failed to list approved payroll month time entries: %w", err)
	}
	filteredApprovedEntries := filterPayrollPreviewEntriesByContractType(
		approvedEntries,
		normalized.ContractType,
	)
	liveShiftCountByEmployee := buildLiveShiftCountMap(filteredApprovedEntries)

	pendingEntries, err := s.repository.ListPayrollMonthPendingEntries(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetPayrollMonthSummary",
			"failed to list pending payroll time entries",
			err,
		)
		return nil, fmt.Errorf("failed to list pending payroll month entries: %w", err)
	}
	pendingByEmployee := buildPendingSummaryMap(pendingEntries, normalized.ContractType)

	holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	liveSummaries, err := buildPayrollMonthLiveSummaries(filteredApprovedEntries, holidaySet)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PayrollMonthSummaryRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.PayrollMonthSummaryRow{
			EmployeeID:     employee.EmployeeID,
			EmployeeName:   employee.EmployeeName,
			Month:          normalized.Month,
			IsCurrentMonth: isCurrentMonth,
			DataSource:     "live",
		}

		if pending, ok := pendingByEmployee[employee.EmployeeID]; ok {
			row.PendingEntryCount = pending.PendingEntryCount
			row.PendingWorkedMinutes = pending.PendingWorkedMinutes
		}

		lockedPayPeriod, hasLockedSnapshot := lockedByEmployee[employee.EmployeeID]
		lockedSnapshot, hasMatchingLockedSnapshot := lockedSnapshotByPeriod[lockedPayPeriod.ID]
		if hasLockedSnapshot {
			row.HasLockedSnapshot = true
			row.PayPeriodID = &lockedPayPeriod.ID
			status := lockedPayPeriod.Status
			row.PayPeriodStatus = &status
			row.PaidAt = lockedPayPeriod.PaidAt
		}

		if isCurrentMonth {
			if live, ok := liveSummaries[employee.EmployeeID]; ok {
				applyLivePayrollMonthSummary(&row, live)
			}
			row.ShiftCount = liveShiftCountByEmployee[employee.EmployeeID]
		} else if hasLockedSnapshot {
			if hasMatchingLockedSnapshot {
				row.IsLocked = true
				row.DataSource = "locked"
				applyLockedPayrollMonthSummary(&row, lockedSnapshot)
			}
		} else if live, ok := liveSummaries[employee.EmployeeID]; ok {
			applyLivePayrollMonthSummary(&row, live)
			row.ShiftCount = liveShiftCountByEmployee[employee.EmployeeID]
		}

		if normalized.ContractType != nil &&
			!shouldIncludeContractFilteredPayrollRow(
				row,
				hasMatchingLockedSnapshot,
				liveSummaries[employee.EmployeeID],
				pendingByEmployee[employee.EmployeeID],
			) {
			continue
		}

		items = append(items, row)
	}

	if normalized.ContractType != nil {
		totalCount = int64(len(items))
	}

	return &domain.PayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *PayoutService) GetPayrollMonthORTOverview(
	ctx context.Context,
	params domain.PayrollMonthORTOverviewParams,
) (*domain.PayrollMonthORTOverviewPage, error) {
	normalized, monthStart, monthEnd, isCurrentMonth, err := normalizePayrollMonthORTOverviewParams(params)
	if err != nil {
		return nil, err
	}

	employees, err := s.repository.ListPayrollMonthEmployeesAll(ctx, normalized, monthStart, monthEnd)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthORTOverview", "failed to list payroll month employees", err)
		return nil, fmt.Errorf("failed to list payroll month employees: %w", err)
	}
	if len(employees) == 0 {
		return &domain.PayrollMonthORTOverviewPage{
			Month:        normalized.Month,
			Distribution: []domain.PayrollMultiplierSummary{},
			Items:        []domain.PayrollMonthORTOverviewRow{},
			TotalCount:   0,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	lockedPayPeriods, err := s.repository.ListPayPeriodsByEmployeesAndRange(ctx, employeeIDs, monthStart, monthEnd)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthORTOverview", "failed to list locked pay periods", err)
		return nil, fmt.Errorf("failed to list pay periods for payroll month ORT overview: %w", err)
	}

	lockedByEmployee := make(map[uuid.UUID]domain.PayPeriod, len(lockedPayPeriods))
	payPeriodIDs := make([]uuid.UUID, 0, len(lockedPayPeriods))
	for _, payPeriod := range lockedPayPeriods {
		lockedByEmployee[payPeriod.EmployeeID] = payPeriod
		payPeriodIDs = append(payPeriodIDs, payPeriod.ID)
	}

	lockedDistributionByPeriod := make(map[uuid.UUID][]domain.PayrollMultiplierSummary, len(payPeriodIDs))
	if len(payPeriodIDs) > 0 {
		lockedSummaries, err := s.repository.ListPayrollMonthLockedMultiplierSummaries(ctx, payPeriodIDs)
		if err != nil {
			s.logError(
				ctx,
				"GetPayrollMonthORTOverview",
				"failed to list locked pay period multiplier summaries",
				err,
			)
			return nil, fmt.Errorf("failed to list locked payroll month multiplier summaries: %w", err)
		}
		lockedDistributionByPeriod = buildLockedORTDistributionMap(lockedSummaries)
	}

	approvedEntries, err := s.repository.ListPayrollMonthApprovedTimeEntries(ctx, employeeIDs, monthStart, monthEnd)
	if err != nil {
		s.logError(
			ctx,
			"GetPayrollMonthORTOverview",
			"failed to list approved payroll time entries",
			err,
		)
		return nil, fmt.Errorf("failed to list approved payroll month time entries: %w", err)
	}

	liveSummaries := make(map[uuid.UUID]payrollMonthLiveSummary)
	if len(approvedEntries) > 0 {
		holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		liveSummaries, err = buildPayrollMonthLiveSummaries(approvedEntries, holidaySet)
		if err != nil {
			return nil, err
		}
	}

	totalBuckets := make(map[float64]*domain.PayrollMultiplierSummary)
	items := make([]domain.PayrollMonthORTOverviewRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.PayrollMonthORTOverviewRow{
			EmployeeID:     employee.EmployeeID,
			EmployeeName:   employee.EmployeeName,
			Month:          normalized.Month,
			IsCurrentMonth: isCurrentMonth,
		}

		lockedPayPeriod, hasLockedSnapshot := lockedByEmployee[employee.EmployeeID]
		if hasLockedSnapshot {
			row.HasLockedSnapshot = true
			row.PayPeriodID = &lockedPayPeriod.ID
			status := lockedPayPeriod.Status
			row.PayPeriodStatus = &status
			row.PaidAt = lockedPayPeriod.PaidAt
		}

		switch {
		case isCurrentMonth:
			live, ok := liveSummaries[employee.EmployeeID]
			if !ok {
				continue
			}
			row.DataSource = "live"
			row.Distribution = positiveMultiplierSummaries(live.MultiplierSummaries)
		case hasLockedSnapshot:
			row.IsLocked = true
			row.DataSource = "locked"
			row.Distribution = lockedDistributionByPeriod[lockedPayPeriod.ID]
		default:
			live, ok := liveSummaries[employee.EmployeeID]
			if !ok {
				continue
			}
			row.DataSource = "live"
			row.Distribution = positiveMultiplierSummaries(live.MultiplierSummaries)
		}

		if len(row.Distribution) == 0 {
			continue
		}

		applyORTOverviewTotals(&row)
		addMultiplierSummaries(totalBuckets, row.Distribution)
		items = append(items, row)
	}

	totalCount := int64(len(items))
	start := int(normalized.Offset)
	if start > len(items) {
		start = len(items)
	}
	end := start + int(normalized.Limit)
	if end > len(items) {
		end = len(items)
	}

	return &domain.PayrollMonthORTOverviewPage{
		Month:        normalized.Month,
		Distribution: sortedMultiplierSummaries(totalBuckets),
		Items:        items[start:end],
		TotalCount:   totalCount,
	}, nil
}

func (s *PayoutService) GetPayrollMonthDetail(
	ctx context.Context,
	employeeID uuid.UUID,
	month time.Time,
	contractType *string,
) (*domain.PayrollMonthDetail, error) {
	if employeeID == uuid.Nil || month.IsZero() {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	normalizedContractType, err := normalizePayrollContractType(contractType)
	if err != nil {
		return nil, err
	}

	monthStart := time.Date(month.UTC().Year(), month.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)

	employee, err := s.repository.GetPayrollPreviewEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	payPeriods, err := s.repository.ListPayPeriodsByEmployeesAndRange(
		ctx,
		[]uuid.UUID{employeeID},
		monthStart,
		monthEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list pay periods for detail: %w", err)
	}

	var selectedPayPeriod *domain.PayPeriod
	for _, period := range payPeriods {
		if period.EmployeeID != employeeID {
			continue
		}
		if period.PeriodStart.Equal(monthStart) && period.PeriodEnd.Equal(monthEnd) {
			item, getErr := s.loadPayPeriodWithLineItems(ctx, period.ID)
			if getErr != nil {
				return nil, getErr
			}
			selectedPayPeriod = item
			break
		}
	}

	if selectedPayPeriod != nil {
		filteredPayPeriod := filterPayPeriodByContractType(selectedPayPeriod, normalizedContractType)
		if normalizedContractType != nil && len(filteredPayPeriod.LineItems) == 0 {
			return nil, domain.ErrPayPeriodNotFound
		}
		return &domain.PayrollMonthDetail{
			EmployeeID:   employeeID,
			EmployeeName: strings.TrimSpace(employee.FirstName + " " + employee.LastName),
			Month:        monthStart,
			DataSource:   "locked",
			PayPeriod:    filteredPayPeriod,
		}, nil
	}

	// For live preview, use approved entries (not just unpaid entries)
	approvedEntries, err := s.repository.ListPayrollMonthApprovedTimeEntries(
		ctx,
		[]uuid.UUID{employeeID},
		monthStart,
		monthEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list approved payroll entries for detail: %w", err)
	}
	approvedEntries = filterPayrollPreviewEntriesByContractType(approvedEntries, normalizedContractType)

	preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
		EmployeeID:  employeeID,
		PeriodStart: monthStart,
		PeriodEnd:   monthEnd,
	}, approvedEntries)
	if err != nil {
		return nil, err
	}

	return &domain.PayrollMonthDetail{
		EmployeeID:   employeeID,
		EmployeeName: strings.TrimSpace(employee.FirstName + " " + employee.LastName),
		Month:        monthStart,
		DataSource:   "live",
		Preview:      preview,
	}, nil
}

func (s *PayoutService) ExportPayrollMonthPDF(
	ctx context.Context,
	employeeID uuid.UUID,
	month time.Time,
	contractType *string,
) ([]byte, string, error) {
	detail, err := s.GetPayrollMonthDetail(ctx, employeeID, month, contractType)
	if err != nil {
		return nil, "", err
	}

	pdfBytes, err := buildPayrollMonthDetailPDF(detail)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf(
		"salary_%s_%s.pdf",
		strings.ReplaceAll(strings.ToLower(detail.EmployeeName), " ", "_"),
		detail.Month.Format("2006-01"),
	)

	return pdfBytes, filename, nil
}

func (s *PayoutService) MarkPayPeriodPaidByAdmin(
	ctx context.Context,
	adminEmployeeID, payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	if adminEmployeeID == uuid.Nil || payPeriodID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}

	var updated *domain.PayPeriod
	err := s.repository.WithTx(ctx, func(tx domain.PayoutTxRepository) error {
		current, err := tx.GetPayPeriodForUpdate(ctx, payPeriodID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayPeriodStatusDraft {
			return domain.ErrPayPeriodStateInvalid
		}

		updated, err = tx.MarkPayPeriodPaid(ctx, payPeriodID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func isValidPayoutStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.PayoutRequestStatusPending,
		domain.PayoutRequestStatusApproved,
		domain.PayoutRequestStatusRejected,
		domain.PayoutRequestStatusPaid:
		return true
	default:
		return false
	}
}

func isValidPayPeriodStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.PayPeriodStatusDraft, domain.PayPeriodStatusPaid:
		return true
	default:
		return false
	}
}

func isValidPayrollIrregularHoursProfile(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.IrregularHoursProfileNone,
		domain.IrregularHoursProfileRoster,
		domain.IrregularHoursProfileNonRoster:
		return true
	default:
		return false
	}
}

func normalizePayrollContractType(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	switch strings.ToLower(strings.TrimSpace(*value)) {
	case "loondienst":
		normalized := "LOONDIENST"
		return &normalized, nil
	case "zzp":
		normalized := "ZZP"
		return &normalized, nil
	default:
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
}

func normalizePayrollPreviewParams(
	params domain.PayrollPreviewParams,
) (domain.PayrollPreviewParams, error) {
	if params.EmployeeID == uuid.Nil || params.PeriodStart.IsZero() || params.PeriodEnd.IsZero() {
		return domain.PayrollPreviewParams{}, domain.ErrPayoutRequestInvalidRequest
	}

	start := time.Date(
		params.PeriodStart.UTC().Year(),
		params.PeriodStart.UTC().Month(),
		params.PeriodStart.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	end := time.Date(
		params.PeriodEnd.UTC().Year(),
		params.PeriodEnd.UTC().Month(),
		params.PeriodEnd.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if end.Before(start) {
		return domain.PayrollPreviewParams{}, domain.ErrPayoutRequestInvalidRequest
	}

	params.PeriodStart = start
	params.PeriodEnd = end
	return params, nil
}

func normalizeClosePayPeriodParams(
	params domain.ClosePayPeriodParams,
) (domain.ClosePayPeriodParams, error) {
	normalized, err := normalizePayrollPreviewParams(domain.PayrollPreviewParams{
		EmployeeID:  params.EmployeeID,
		PeriodStart: params.PeriodStart,
		PeriodEnd:   params.PeriodEnd,
	})
	if err != nil {
		return domain.ClosePayPeriodParams{}, err
	}

	return domain.ClosePayPeriodParams{
		EmployeeID:  normalized.EmployeeID,
		PeriodStart: normalized.PeriodStart,
		PeriodEnd:   normalized.PeriodEnd,
	}, nil
}

func normalizePayrollMonthSummaryParams(
	params domain.PayrollMonthSummaryParams,
) (domain.PayrollMonthSummaryParams, time.Time, time.Time, bool, error) {
	if params.Month.IsZero() {
		return domain.PayrollMonthSummaryParams{}, time.Time{}, time.Time{}, false, domain.ErrPayoutRequestInvalidRequest
	}

	normalizedContractType, err := normalizePayrollContractType(params.ContractType)
	if err != nil {
		return domain.PayrollMonthSummaryParams{}, time.Time{}, time.Time{}, false, err
	}
	params.ContractType = normalizedContractType

	month := time.Date(
		params.Month.UTC().Year(),
		params.Month.UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	monthEnd := month.AddDate(0, 1, -1)
	now := time.Now().UTC()
	currentMonth := now.Year() == month.Year() && now.Month() == month.Month()

	params.Month = month
	return params, month, monthEnd, currentMonth, nil
}

func normalizePayrollMonthORTOverviewParams(
	params domain.PayrollMonthORTOverviewParams,
) (domain.PayrollMonthORTOverviewParams, time.Time, time.Time, bool, error) {
	if params.Month.IsZero() {
		return domain.PayrollMonthORTOverviewParams{}, time.Time{}, time.Time{}, false, domain.ErrPayoutRequestInvalidRequest
	}

	month := time.Date(
		params.Month.UTC().Year(),
		params.Month.UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	monthEnd := month.AddDate(0, 1, -1)
	now := time.Now().UTC()
	currentMonth := now.Year() == month.Year() && now.Month() == month.Month()

	params.Month = month
	return params, month, monthEnd, currentMonth, nil
}

func (s *PayoutService) buildPayrollPreview(
	ctx context.Context,
	employee *domain.EmployeeDetail,
	params domain.PayrollPreviewParams,
	entries []domain.PayrollPreviewTimeEntry,
) (*domain.PayrollPreview, error) {
	holidaySet, err := s.loadHolidaySet(ctx, params.PeriodStart, params.PeriodEnd)
	if err != nil {
		return nil, err
	}

	preview := &domain.PayrollPreview{
		EmployeeID:   employee.ID,
		EmployeeName: strings.TrimSpace(employee.FirstName + " " + employee.LastName),
		PeriodStart:  params.PeriodStart,
		PeriodEnd:    params.PeriodEnd,
		LineItems:    make([]domain.PayrollPreviewLineItem, 0),
	}

	for _, entry := range entries {
		if !isPayrollEligibleContractType(entry.ContractType) {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}
		if entry.ContractRate == nil || *entry.ContractRate <= 0 {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(entry.IrregularHoursProfile) {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}

		lineItems, workedMinutes, baseAmount, premiumAmount, err := buildPayrollPreviewLineItems(
			entry,
			*entry.ContractRate,
			holidaySet,
		)
		if err != nil {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}

		preview.TotalWorkedMinutes += workedMinutes
		preview.BaseGrossAmount = roundCurrency(preview.BaseGrossAmount + baseAmount)
		preview.IrregularGrossAmount = roundCurrency(preview.IrregularGrossAmount + premiumAmount)
		preview.LineItems = append(preview.LineItems, lineItems...)
	}

	preview.GrossAmount = roundCurrency(preview.BaseGrossAmount + preview.IrregularGrossAmount)
	return preview, nil
}

func buildPayrollPreviewLineItems(
	entry domain.PayrollPreviewTimeEntry,
	hourlyRate float64,
	holidaySet map[string]struct{},
) ([]domain.PayrollPreviewLineItem, int32, float64, float64, error) {
	start, end, err := parseTimeEntryBounds(entry.EntryDate, entry.StartTime, entry.EndTime)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	totalMinutes := int32(end.Sub(start).Minutes())
	if totalMinutes <= 0 || entry.BreakMinutes < 0 || entry.BreakMinutes >= totalMinutes {
		return nil, 0, 0, 0, domain.ErrPayoutRequestInvalidRequest
	}

	paidFactor := float64(totalMinutes-entry.BreakMinutes) / float64(totalMinutes)
	if paidFactor <= 0 {
		return nil, 0, 0, 0, domain.ErrPayoutRequestInvalidRequest
	}

	type segment struct {
		workDate time.Time
		start    time.Time
		end      time.Time
		rate     float64
		minutes  int32
	}

	segments := make([]segment, 0, 8)
	current := start
	segmentStart := start
	segmentRate := appliedPayrollRateForMinute(entry, current, holidaySet)
	segmentWorkDate := time.Date(
		current.Year(),
		current.Month(),
		current.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	for current.Before(end) {
		next := current.Add(time.Minute)
		nextRate := segmentRate
		nextWorkDate := segmentWorkDate
		if next.Before(end) {
			nextRate = appliedPayrollRateForMinute(entry, next, holidaySet)
			nextWorkDate = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
		}
		if next.Equal(end) || nextRate != segmentRate || !nextWorkDate.Equal(segmentWorkDate) {
			segments = append(segments, segment{
				workDate: segmentWorkDate,
				start:    segmentStart,
				end:      next,
				rate:     segmentRate,
				minutes:  int32(next.Sub(segmentStart).Minutes()),
			})
			segmentStart = next
			segmentRate = nextRate
			segmentWorkDate = nextWorkDate
		}
		current = next
	}

	items := make([]domain.PayrollPreviewLineItem, 0, len(segments))
	var baseTotal float64
	var premiumTotal float64
	for _, segment := range segments {
		paidMinutes := float64(segment.minutes) * paidFactor
		baseAmount := roundCurrency(hourlyRate * paidMinutes / 60)
		premiumAmount := roundCurrency(baseAmount * segment.rate / 100)

		baseTotal = roundCurrency(baseTotal + baseAmount)
		premiumTotal = roundCurrency(premiumTotal + premiumAmount)

		items = append(items, domain.PayrollPreviewLineItem{
			TimeEntryID:           entry.ID,
			ContractType:          entry.ContractType,
			WorkDate:              segment.workDate,
			HourType:              entry.HourType,
			StartTime:             segment.start.Format("15:04"),
			EndTime:               segment.end.Format("15:04"),
			IrregularHoursProfile: entry.IrregularHoursProfile,
			AppliedRatePercent:    segment.rate,
			MinutesWorked:         segment.minutes,
			PaidMinutes:           roundCurrency(paidMinutes),
			BaseAmount:            baseAmount,
			PremiumAmount:         premiumAmount,
		})
	}

	return items, totalMinutes - entry.BreakMinutes, baseTotal, premiumTotal, nil
}

func parseTimeEntryBounds(
	entryDate time.Time,
	startTime, endTime string,
) (time.Time, time.Time, error) {
	baseDate := time.Date(
		entryDate.UTC().Year(),
		entryDate.UTC().Month(),
		entryDate.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	startParsed, err := time.Parse("15:04:05", startTime)
	if err != nil {
		startParsed, err = time.Parse("15:04", startTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	endParsed, err := time.Parse("15:04:05", endTime)
	if err != nil {
		endParsed, err = time.Parse("15:04", endTime)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	start := time.Date(
		baseDate.Year(),
		baseDate.Month(),
		baseDate.Day(),
		startParsed.Hour(),
		startParsed.Minute(),
		startParsed.Second(),
		0,
		time.UTC,
	)
	end := time.Date(
		baseDate.Year(),
		baseDate.Month(),
		baseDate.Day(),
		endParsed.Hour(),
		endParsed.Minute(),
		endParsed.Second(),
		0,
		time.UTC,
	)
	if !end.After(start) {
		end = end.Add(24 * time.Hour)
	}
	return start, end, nil
}

func ortRateForMinute(profile string, minute time.Time, holidaySet map[string]struct{}) float64 {
	workDate := time.Date(minute.Year(), minute.Month(), minute.Day(), 0, 0, 0, 0, time.UTC)
	if _, ok := holidaySet[workDate.Format(time.DateOnly)]; ok {
		return 45
	}

	switch workDate.Weekday() {
	case time.Sunday:
		return 45
	case time.Saturday:
		minutesOfDay := minute.Hour()*60 + minute.Minute()
		if minutesOfDay >= 6*60 && minutesOfDay < 22*60 {
			return 30
		}
		if minutesOfDay >= 22*60 || minutesOfDay < 6*60 {
			return 45
		}
	}

	minutesOfDay := minute.Hour()*60 + minute.Minute()
	if minutesOfDay >= 22*60 || minutesOfDay < 6*60 {
		return 45
	}

	switch strings.ToLower(strings.TrimSpace(profile)) {
	case domain.IrregularHoursProfileRoster:
		if minutesOfDay >= 6*60 && minutesOfDay < 7*60 {
			return 25
		}
		if minutesOfDay >= 19*60 && minutesOfDay < 22*60 {
			return 25
		}
	case domain.IrregularHoursProfileNonRoster:
		if minutesOfDay >= 20*60 && minutesOfDay < 22*60 {
			return 25
		}
	}

	return 0
}

func roundCurrency(v float64) float64 {
	return math.Round(v*100) / 100
}

func appliedPayrollRateForMinute(
	entry domain.PayrollPreviewTimeEntry,
	minute time.Time,
	holidaySet map[string]struct{},
) float64 {
	if !isPayrollORTEligibleContractType(entry.ContractType) {
		return 0
	}
	return ortRateForMinute(entry.IrregularHoursProfile, minute, holidaySet)
}

func isPayrollEligibleContractType(contractType string) bool {
	switch strings.ToLower(strings.TrimSpace(contractType)) {
	case "loondienst", "zzp":
		return true
	default:
		return false
	}
}

func isPayrollORTEligibleContractType(contractType string) bool {
	return strings.EqualFold(strings.TrimSpace(contractType), "loondienst")
}

type payrollMonthLiveSummary struct {
	WorkedMinutes        int32
	PaidMinutes          float64
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	MultiplierSummaries  []domain.PayrollMultiplierSummary
}

func buildPayrollMonthLiveSummaries(
	entries []domain.PayrollPreviewTimeEntry,
	holidaySet map[string]struct{},
) (map[uuid.UUID]payrollMonthLiveSummary, error) {
	type liveAccumulator struct {
		WorkedMinutes        int32
		PaidMinutes          float64
		BaseGrossAmount      float64
		IrregularGrossAmount float64
		MultiplierByRate     map[float64]*domain.PayrollMultiplierSummary
	}

	accumulators := make(map[uuid.UUID]*liveAccumulator)
	for _, entry := range entries {
		if !isPayrollEligibleContractType(entry.ContractType) {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}
		if entry.ContractRate == nil || *entry.ContractRate <= 0 {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(entry.IrregularHoursProfile) {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}

		lineItems, workedMinutes, baseAmount, premiumAmount, err := buildPayrollPreviewLineItems(
			entry,
			*entry.ContractRate,
			holidaySet,
		)
		if err != nil {
			return nil, domain.ErrPayoutRequestInvalidRequest
		}

		acc := accumulators[entry.EmployeeID]
		if acc == nil {
			acc = &liveAccumulator{
				MultiplierByRate: make(map[float64]*domain.PayrollMultiplierSummary),
			}
			accumulators[entry.EmployeeID] = acc
		}

		acc.WorkedMinutes += workedMinutes
		acc.BaseGrossAmount = roundCurrency(acc.BaseGrossAmount + baseAmount)
		acc.IrregularGrossAmount = roundCurrency(acc.IrregularGrossAmount + premiumAmount)
		for _, item := range lineItems {
			acc.PaidMinutes = roundCurrency(acc.PaidMinutes + item.PaidMinutes)
			bucket := acc.MultiplierByRate[item.AppliedRatePercent]
			if bucket == nil {
				bucket = &domain.PayrollMultiplierSummary{RatePercent: item.AppliedRatePercent}
				acc.MultiplierByRate[item.AppliedRatePercent] = bucket
			}
			bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + item.PaidMinutes)
			bucket.PaidMinutes = roundCurrency(bucket.PaidMinutes + item.PaidMinutes)
			bucket.BaseAmount = roundCurrency(bucket.BaseAmount + item.BaseAmount)
			bucket.PremiumAmount = roundCurrency(bucket.PremiumAmount + item.PremiumAmount)
		}
	}

	results := make(map[uuid.UUID]payrollMonthLiveSummary, len(accumulators))
	for employeeID, acc := range accumulators {
		results[employeeID] = payrollMonthLiveSummary{
			WorkedMinutes:        acc.WorkedMinutes,
			PaidMinutes:          acc.PaidMinutes,
			BaseGrossAmount:      acc.BaseGrossAmount,
			IrregularGrossAmount: acc.IrregularGrossAmount,
			GrossAmount:          roundCurrency(acc.BaseGrossAmount + acc.IrregularGrossAmount),
			MultiplierSummaries:  sortedMultiplierSummaries(acc.MultiplierByRate),
		}
	}

	return results, nil
}

func buildLiveShiftCountMap(entries []domain.PayrollPreviewTimeEntry) map[uuid.UUID]int32 {
	counts := make(map[uuid.UUID]int32)
	for _, entry := range entries {
		counts[entry.EmployeeID]++
	}
	return counts
}

type lockedPayrollSnapshot struct {
	WorkedMinutes        int32
	PaidMinutes          float64
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	ShiftCount           int32
	MultiplierSummaries  []domain.PayrollMultiplierSummary
}

func (s *PayoutService) buildLockedSnapshotMap(
	ctx context.Context,
	payPeriods []domain.PayPeriod,
	contractType *string,
) (map[uuid.UUID]lockedPayrollSnapshot, error) {
	snapshots := make(map[uuid.UUID]lockedPayrollSnapshot, len(payPeriods))
	for _, period := range payPeriods {
		lineItems, err := s.repository.ListPayPeriodLineItems(ctx, period.ID)
		if err != nil {
			return nil, err
		}

		filteredLineItems := filterPayPeriodLineItemsByContractType(lineItems, contractType)
		if len(filteredLineItems) == 0 {
			continue
		}

		snapshots[period.ID] = buildLockedPayrollSnapshot(filteredLineItems)
	}
	return snapshots, nil
}

func buildLockedPayrollSnapshot(lineItems []domain.PayPeriodLineItem) lockedPayrollSnapshot {
	multiplierBuckets := make(map[float64]*domain.PayrollMultiplierSummary)
	uniqueTimeEntryIDs := make(map[uuid.UUID]struct{})
	var workedMinutes float64
	var paidMinutes float64
	var baseGrossAmount float64
	var irregularGrossAmount float64

	for _, item := range lineItems {
		workedMinutes = roundCurrency(workedMinutes + item.MinutesWorked)
		paidMinutes = roundCurrency(paidMinutes + item.MinutesWorked)
		baseGrossAmount = roundCurrency(baseGrossAmount + item.BaseAmount)
		irregularGrossAmount = roundCurrency(irregularGrossAmount + item.PremiumAmount)
		if item.TimeEntryID != nil {
			uniqueTimeEntryIDs[*item.TimeEntryID] = struct{}{}
		}

		bucket := multiplierBuckets[item.AppliedRatePercent]
		if bucket == nil {
			bucket = &domain.PayrollMultiplierSummary{RatePercent: item.AppliedRatePercent}
			multiplierBuckets[item.AppliedRatePercent] = bucket
		}
		bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + item.MinutesWorked)
		bucket.PaidMinutes = roundCurrency(bucket.PaidMinutes + item.MinutesWorked)
		bucket.BaseAmount = roundCurrency(bucket.BaseAmount + item.BaseAmount)
		bucket.PremiumAmount = roundCurrency(bucket.PremiumAmount + item.PremiumAmount)
	}

	return lockedPayrollSnapshot{
		WorkedMinutes:        int32(math.Round(workedMinutes)),
		PaidMinutes:          paidMinutes,
		BaseGrossAmount:      baseGrossAmount,
		IrregularGrossAmount: irregularGrossAmount,
		GrossAmount:          roundCurrency(baseGrossAmount + irregularGrossAmount),
		ShiftCount:           int32(len(uniqueTimeEntryIDs)),
		MultiplierSummaries:  sortedMultiplierSummaries(multiplierBuckets),
	}
}

func filterPayrollPreviewEntriesByContractType(
	entries []domain.PayrollPreviewTimeEntry,
	contractType *string,
) []domain.PayrollPreviewTimeEntry {
	if contractType == nil {
		return entries
	}

	filtered := make([]domain.PayrollPreviewTimeEntry, 0, len(entries))
	for _, entry := range entries {
		if matchesPayrollContractType(entry.ContractType, *contractType) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func filterPayPeriodLineItemsByContractType(
	items []domain.PayPeriodLineItem,
	contractType *string,
) []domain.PayPeriodLineItem {
	if contractType == nil {
		return items
	}

	filtered := make([]domain.PayPeriodLineItem, 0, len(items))
	for _, item := range items {
		if matchesPayrollContractType(item.ContractType, *contractType) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func buildPendingSummaryMap(
	entries []domain.PayrollMonthPendingEntry,
	contractType *string,
) map[uuid.UUID]domain.PayrollMonthPendingSummary {
	summaries := make(map[uuid.UUID]domain.PayrollMonthPendingSummary)
	for _, entry := range entries {
		if contractType != nil && !matchesPayrollContractType(entry.ContractType, *contractType) {
			continue
		}
		summary := summaries[entry.EmployeeID]
		summary.EmployeeID = entry.EmployeeID
		summary.PendingEntryCount++
		summary.PendingWorkedMinutes += entry.WorkedMinutes
		summaries[entry.EmployeeID] = summary
	}
	return summaries
}

func shouldIncludeContractFilteredPayrollRow(
	row domain.PayrollMonthSummaryRow,
	hasLockedSnapshot bool,
	live payrollMonthLiveSummary,
	pending domain.PayrollMonthPendingSummary,
) bool {
	if hasLockedSnapshot {
		return true
	}
	if row.WorkedMinutes > 0 || row.PaidMinutes > 0 || row.GrossAmount > 0 || row.ShiftCount > 0 {
		return true
	}
	return pending.PendingEntryCount > 0 || pending.PendingWorkedMinutes > 0 ||
		live.WorkedMinutes > 0 || live.PaidMinutes > 0 || live.GrossAmount > 0
}

func matchesPayrollContractType(actual, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
}

func filterPayPeriodByContractType(
	payPeriod *domain.PayPeriod,
	contractType *string,
) *domain.PayPeriod {
	if payPeriod == nil {
		return nil
	}

	filtered := *payPeriod
	filtered.LineItems = filterPayPeriodLineItemsByContractType(payPeriod.LineItems, contractType)
	if contractType == nil {
		return &filtered
	}

	snapshot := buildLockedPayrollSnapshot(filtered.LineItems)
	filtered.BaseGrossAmount = snapshot.BaseGrossAmount
	filtered.IrregularGrossAmount = snapshot.IrregularGrossAmount
	filtered.GrossAmount = snapshot.GrossAmount
	return &filtered
}

func sortedMultiplierSummaries(
	buckets map[float64]*domain.PayrollMultiplierSummary,
) []domain.PayrollMultiplierSummary {
	keys := make([]float64, 0, len(buckets))
	for rate := range buckets {
		keys = append(keys, rate)
	}
	sort.Float64s(keys)

	items := make([]domain.PayrollMultiplierSummary, 0, len(keys))
	for _, rate := range keys {
		items = append(items, *buckets[rate])
	}
	return items
}

func positiveMultiplierSummaries(
	items []domain.PayrollMultiplierSummary,
) []domain.PayrollMultiplierSummary {
	filtered := make([]domain.PayrollMultiplierSummary, 0, len(items))
	for _, item := range items {
		if item.RatePercent <= 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func addMultiplierSummaries(
	buckets map[float64]*domain.PayrollMultiplierSummary,
	items []domain.PayrollMultiplierSummary,
) {
	for _, item := range items {
		bucket := buckets[item.RatePercent]
		if bucket == nil {
			bucket = &domain.PayrollMultiplierSummary{RatePercent: item.RatePercent}
			buckets[item.RatePercent] = bucket
		}
		bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + item.WorkedMinutes)
		bucket.PaidMinutes = roundCurrency(bucket.PaidMinutes + item.PaidMinutes)
		bucket.BaseAmount = roundCurrency(bucket.BaseAmount + item.BaseAmount)
		bucket.PremiumAmount = roundCurrency(bucket.PremiumAmount + item.PremiumAmount)
	}
}

func buildLockedORTDistributionMap(
	summaries []domain.PayrollLockedMultiplierSummary,
) map[uuid.UUID][]domain.PayrollMultiplierSummary {
	bucketsByPeriod := make(map[uuid.UUID]map[float64]*domain.PayrollMultiplierSummary)
	for _, item := range summaries {
		if item.RatePercent <= 0 {
			continue
		}
		buckets := bucketsByPeriod[item.PayPeriodID]
		if buckets == nil {
			buckets = make(map[float64]*domain.PayrollMultiplierSummary)
			bucketsByPeriod[item.PayPeriodID] = buckets
		}
		bucket := buckets[item.RatePercent]
		if bucket == nil {
			bucket = &domain.PayrollMultiplierSummary{RatePercent: item.RatePercent}
			buckets[item.RatePercent] = bucket
		}
		bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + item.WorkedMinutes)
		bucket.PaidMinutes = roundCurrency(bucket.PaidMinutes + item.PaidMinutes)
		bucket.BaseAmount = roundCurrency(bucket.BaseAmount + item.BaseAmount)
		bucket.PremiumAmount = roundCurrency(bucket.PremiumAmount + item.PremiumAmount)
	}

	result := make(map[uuid.UUID][]domain.PayrollMultiplierSummary, len(bucketsByPeriod))
	for payPeriodID, buckets := range bucketsByPeriod {
		result[payPeriodID] = sortedMultiplierSummaries(buckets)
	}
	return result
}

func applyORTOverviewTotals(row *domain.PayrollMonthORTOverviewRow) {
	for _, item := range row.Distribution {
		row.WorkedMinutes = roundCurrency(row.WorkedMinutes + item.WorkedMinutes)
		row.PaidMinutes = roundCurrency(row.PaidMinutes + item.PaidMinutes)
		row.BaseAmount = roundCurrency(row.BaseAmount + item.BaseAmount)
		row.PremiumAmount = roundCurrency(row.PremiumAmount + item.PremiumAmount)
	}
}

func applyLivePayrollMonthSummary(
	row *domain.PayrollMonthSummaryRow,
	live payrollMonthLiveSummary,
) {
	row.DataSource = "live"
	row.WorkedMinutes = live.WorkedMinutes
	row.PaidMinutes = live.PaidMinutes
	row.BaseGrossAmount = live.BaseGrossAmount
	row.IrregularGrossAmount = live.IrregularGrossAmount
	row.GrossAmount = live.GrossAmount
	row.MultiplierSummaries = live.MultiplierSummaries
}

func applyLockedPayrollMonthSummary(
	row *domain.PayrollMonthSummaryRow,
	snapshot lockedPayrollSnapshot,
) {
	row.WorkedMinutes = snapshot.WorkedMinutes
	row.PaidMinutes = snapshot.PaidMinutes
	row.BaseGrossAmount = snapshot.BaseGrossAmount
	row.IrregularGrossAmount = snapshot.IrregularGrossAmount
	row.GrossAmount = snapshot.GrossAmount
	row.ShiftCount = snapshot.ShiftCount
	row.MultiplierSummaries = snapshot.MultiplierSummaries
}

func (s *PayoutService) loadHolidaySet(
	ctx context.Context,
	startDate, endDate time.Time,
) (map[string]struct{}, error) {
	holidays, err := s.repository.ListNationalHolidays(
		ctx,
		"NL",
		startDate,
		endDate.AddDate(0, 0, 1),
	)
	if err != nil {
		s.logError(ctx, "loadHolidaySet", "failed to list national holidays", err)
		return nil, fmt.Errorf("failed to list national holidays: %w", err)
	}

	holidaySet := make(map[string]struct{}, len(holidays))
	for _, holiday := range holidays {
		holidaySet[holiday.Date.UTC().Format(time.DateOnly)] = struct{}{}
	}
	return holidaySet, nil
}

func buildPayPeriodLineItem(
	item domain.PayrollPreviewLineItem,
	entries []domain.PayrollPreviewTimeEntry,
) domain.PayPeriodLineItem {
	metadata := map[string]any{
		"start_time":   item.StartTime,
		"end_time":     item.EndTime,
		"paid_minutes": roundCurrency(item.PaidMinutes),
	}

	if entry, ok := findPreviewTimeEntry(entries, item.TimeEntryID); ok {
		metadata["break_minutes"] = entry.BreakMinutes
		metadata["contract_type"] = entry.ContractType
		if entry.ContractRate != nil {
			metadata["contract_rate"] = roundCurrency(*entry.ContractRate)
		}
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		payload = []byte(`{}`)
	}

	timeEntryID := item.TimeEntryID
	return domain.PayPeriodLineItem{
		TimeEntryID:           &timeEntryID,
		ContractType:          item.ContractType,
		WorkDate:              item.WorkDate,
		LineType:              item.HourType,
		IrregularHoursProfile: item.IrregularHoursProfile,
		AppliedRatePercent:    item.AppliedRatePercent,
		MinutesWorked:         roundCurrency(item.PaidMinutes),
		BaseAmount:            item.BaseAmount,
		PremiumAmount:         item.PremiumAmount,
		Metadata:              payload,
	}
}

func findPreviewTimeEntry(
	entries []domain.PayrollPreviewTimeEntry,
	timeEntryID uuid.UUID,
) (domain.PayrollPreviewTimeEntry, bool) {
	for _, entry := range entries {
		if entry.ID == timeEntryID {
			return entry, true
		}
	}
	return domain.PayrollPreviewTimeEntry{}, false
}

func uniquePreviewTimeEntryIDs(items []domain.PayrollPreviewLineItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(items))
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.TimeEntryID]; ok {
			continue
		}
		seen[item.TimeEntryID] = struct{}{}
		result = append(result, item.TimeEntryID)
	}
	return result
}

func (s *PayoutService) logError(
	ctx context.Context,
	operation, message string,
	err error,
	fields ...zap.Field,
) {
	if s.logger == nil {
		return
	}
	s.logger.LogError(ctx, "PayoutService."+operation, message, err, fields...)
}
