package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"hrbackend/internal/domain"

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
	payPeriodIDs := make([]uuid.UUID, 0, len(lockedPayPeriods))
	for _, payPeriod := range lockedPayPeriods {
		lockedByEmployee[payPeriod.EmployeeID] = payPeriod
		payPeriodIDs = append(payPeriodIDs, payPeriod.ID)
	}

	lockedMultiplierRows, err := s.repository.ListPayrollMonthLockedMultiplierSummaries(
		ctx,
		payPeriodIDs,
	)
	if err != nil {
		s.logError(ctx, "GetPayrollMonthSummary", "failed to list locked multiplier summaries", err)
		return nil, fmt.Errorf("failed to list locked multiplier summaries: %w", err)
	}
	lockedMultiplierByPeriod := buildLockedMultiplierSummaryMap(lockedMultiplierRows)

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

	pendingRows, err := s.repository.ListPayrollMonthPendingSummaries(
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
		return nil, fmt.Errorf("failed to list pending payroll month summaries: %w", err)
	}
	pendingByEmployee := make(map[uuid.UUID]domain.PayrollMonthPendingSummary, len(pendingRows))
	for _, item := range pendingRows {
		pendingByEmployee[item.EmployeeID] = item
	}

	holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	liveSummaries, err := buildPayrollMonthLiveSummaries(approvedEntries, holidaySet)
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
		} else if hasLockedSnapshot {
			row.IsLocked = true
			row.DataSource = "locked"
			applyLockedPayrollMonthSummary(
				&row,
				lockedPayPeriod,
				lockedMultiplierByPeriod[lockedPayPeriod.ID],
			)
		} else if live, ok := liveSummaries[employee.EmployeeID]; ok {
			applyLivePayrollMonthSummary(&row, live)
		}

		items = append(items, row)
	}

	return &domain.PayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *PayoutService) GetPayrollMonthDetail(
	ctx context.Context,
	employeeID uuid.UUID,
	month time.Time,
) (*domain.PayrollMonthDetail, error) {
	if employeeID == uuid.Nil || month.IsZero() {
		return nil, domain.ErrPayoutRequestInvalidRequest
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
			item, getErr := s.repository.GetPayPeriodByID(ctx, period.ID)
			if getErr != nil {
				return nil, getErr
			}
			selectedPayPeriod = item
			break
		}
	}

	if selectedPayPeriod != nil {
		return &domain.PayrollMonthDetail{
			EmployeeID:   employeeID,
			EmployeeName: strings.TrimSpace(employee.FirstName + " " + employee.LastName),
			Month:        monthStart,
			DataSource:   "locked",
			PayPeriod:    selectedPayPeriod,
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
) ([]byte, string, error) {
	detail, err := s.GetPayrollMonthDetail(ctx, employeeID, month)
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
		if entry.ContractType != "loondienst" {
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
	segmentRate := ortRateForMinute(entry.IrregularHoursProfile, current, holidaySet)
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
			nextRate = ortRateForMinute(entry.IrregularHoursProfile, next, holidaySet)
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
		if entry.ContractType != "loondienst" {
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

func buildLockedMultiplierSummaryMap(
	rows []domain.PayrollLockedMultiplierSummary,
) map[uuid.UUID][]domain.PayrollMultiplierSummary {
	grouped := make(map[uuid.UUID][]domain.PayrollMultiplierSummary)
	for _, row := range rows {
		grouped[row.PayPeriodID] = append(grouped[row.PayPeriodID], domain.PayrollMultiplierSummary{
			RatePercent:   row.RatePercent,
			WorkedMinutes: row.WorkedMinutes,
			PaidMinutes:   row.PaidMinutes,
			BaseAmount:    row.BaseAmount,
			PremiumAmount: row.PremiumAmount,
		})
	}
	return grouped
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
	payPeriod domain.PayPeriod,
	multiplierSummaries []domain.PayrollMultiplierSummary,
) {
	var paidMinutes float64
	var workedMinutes float64
	for _, item := range multiplierSummaries {
		paidMinutes = roundCurrency(paidMinutes + item.PaidMinutes)
		workedMinutes = roundCurrency(workedMinutes + item.WorkedMinutes)
	}

	row.WorkedMinutes = int32(math.Round(workedMinutes))
	row.PaidMinutes = paidMinutes
	row.BaseGrossAmount = payPeriod.BaseGrossAmount
	row.IrregularGrossAmount = payPeriod.IrregularGrossAmount
	row.GrossAmount = payPeriod.GrossAmount
	row.MultiplierSummaries = multiplierSummaries
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
