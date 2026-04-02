package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PayoutService struct {
	repository domain.PayoutRepository
	logger     domain.Logger
}

func NewPayoutService(repository domain.PayoutRepository, logger domain.Logger) domain.PayoutService {
	return &PayoutService{
		repository: repository,
		logger:     logger,
	}
}

func (s *PayoutService) CreatePayoutRequest(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CreatePayoutRequestParams) (*domain.PayoutRequest, error) {
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

		if err := tx.EnsureLeaveBalanceForYear(ctx, params.EmployeeID, params.BalanceYear); err != nil {
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

func (s *PayoutService) DecidePayoutRequestByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID, params domain.DecidePayoutRequestParams) (*domain.PayoutRequest, error) {
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
			if err := tx.EnsureLeaveBalanceForYear(ctx, current.EmployeeID, current.BalanceYear); err != nil {
				return err
			}
			balance, err := tx.GetPayoutBalanceForUpdate(ctx, current.EmployeeID, current.BalanceYear)
			if err != nil {
				return err
			}
			if balance.ExtraRemaining < current.RequestedHours {
				return domain.ErrPayoutRequestInsufficientHours
			}
			if _, err := tx.ApplyLeaveBalanceDeduction(ctx, balance.LeaveBalanceID, current.RequestedHours, 0); err != nil {
				return err
			}

			updated, err = tx.ApprovePayoutRequest(ctx, payoutRequestID, adminEmployeeID, params.SalaryMonth.UTC(), params.DecisionNote)
			return err
		}

		updated, err = tx.RejectPayoutRequest(ctx, payoutRequestID, adminEmployeeID, params.DecisionNote)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PayoutService) MarkPayoutRequestPaidByAdmin(ctx context.Context, adminEmployeeID, payoutRequestID uuid.UUID) (*domain.PayoutRequest, error) {
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

func (s *PayoutService) ListMyPayoutRequests(ctx context.Context, params domain.ListMyPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListMyPayoutRequests(ctx, params)
}

func (s *PayoutService) ListPayoutRequests(ctx context.Context, params domain.ListPayoutRequestsParams) (*domain.PayoutRequestPage, error) {
	if params.Status != nil && !isValidPayoutStatus(*params.Status) {
		return nil, domain.ErrPayoutRequestInvalidRequest
	}
	return s.repository.ListPayoutRequests(ctx, params)
}

func (s *PayoutService) PreviewMyPayroll(ctx context.Context, actorEmployeeID uuid.UUID, periodStart, periodEnd time.Time) (*domain.PayrollPreview, error) {
	return s.PreviewPayroll(ctx, domain.PayrollPreviewParams{
		EmployeeID:  actorEmployeeID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
}

func (s *PayoutService) PreviewPayroll(ctx context.Context, params domain.PayrollPreviewParams) (*domain.PayrollPreview, error) {
	normalized, err := normalizePayrollPreviewParams(params)
	if err != nil {
		return nil, err
	}

	employee, err := s.repository.GetPayrollPreviewEmployee(ctx, normalized.EmployeeID)
	if err != nil {
		if err == domain.ErrEmployeeNotFound {
			return nil, err
		}
		s.logError(ctx, "PreviewPayroll", "failed to get employee", err, zap.String("employee_id", normalized.EmployeeID.String()))
		return nil, fmt.Errorf("failed to get employee for payroll preview: %w", err)
	}

	entries, err := s.repository.ListPayrollPreviewTimeEntries(ctx, normalized)
	if err != nil {
		s.logError(ctx, "PreviewPayroll", "failed to list payroll time entries", err,
			zap.String("employee_id", normalized.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to list payroll time entries: %w", err)
	}

	holidays, err := s.repository.ListNationalHolidays(ctx, "NL", normalized.PeriodStart, normalized.PeriodEnd.AddDate(0, 0, 1))
	if err != nil {
		s.logError(ctx, "PreviewPayroll", "failed to list national holidays", err)
		return nil, fmt.Errorf("failed to list national holidays: %w", err)
	}

	holidaySet := make(map[string]struct{}, len(holidays))
	for _, holiday := range holidays {
		holidaySet[holiday.Date.UTC().Format(time.DateOnly)] = struct{}{}
	}

	preview := &domain.PayrollPreview{
		EmployeeID:   employee.ID,
		EmployeeName: strings.TrimSpace(employee.FirstName + " " + employee.LastName),
		PeriodStart:  normalized.PeriodStart,
		PeriodEnd:    normalized.PeriodEnd,
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

		lineItems, workedMinutes, baseAmount, premiumAmount, err := buildPayrollPreviewLineItems(entry, *entry.ContractRate, holidaySet)
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

func normalizePayrollPreviewParams(params domain.PayrollPreviewParams) (domain.PayrollPreviewParams, error) {
	if params.EmployeeID == uuid.Nil || params.PeriodStart.IsZero() || params.PeriodEnd.IsZero() {
		return domain.PayrollPreviewParams{}, domain.ErrPayoutRequestInvalidRequest
	}

	start := time.Date(params.PeriodStart.UTC().Year(), params.PeriodStart.UTC().Month(), params.PeriodStart.UTC().Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(params.PeriodEnd.UTC().Year(), params.PeriodEnd.UTC().Month(), params.PeriodEnd.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if end.Before(start) {
		return domain.PayrollPreviewParams{}, domain.ErrPayoutRequestInvalidRequest
	}

	params.PeriodStart = start
	params.PeriodEnd = end
	return params, nil
}

func buildPayrollPreviewLineItems(entry domain.PayrollPreviewTimeEntry, hourlyRate float64, holidaySet map[string]struct{}) ([]domain.PayrollPreviewLineItem, int32, float64, float64, error) {
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
	segmentWorkDate := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)

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
			BaseAmount:            baseAmount,
			PremiumAmount:         premiumAmount,
		})
	}

	return items, totalMinutes - entry.BreakMinutes, baseTotal, premiumTotal, nil
}

func parseTimeEntryBounds(entryDate time.Time, startTime, endTime string) (time.Time, time.Time, error) {
	baseDate := time.Date(entryDate.UTC().Year(), entryDate.UTC().Month(), entryDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
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

	start := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), startParsed.Hour(), startParsed.Minute(), startParsed.Second(), 0, time.UTC)
	end := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), endParsed.Hour(), endParsed.Minute(), endParsed.Second(), 0, time.UTC)
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

func (s *PayoutService) logError(ctx context.Context, operation, message string, err error, fields ...zap.Field) {
	if s.logger == nil {
		return
	}
	s.logger.LogError(ctx, "PayoutService."+operation, message, err, fields...)
}
