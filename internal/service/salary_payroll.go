package service

import (
	"context"
	"errors"
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

// --- Public methods on *SalaryService ---

func (s *SalaryService) PreviewMyPayroll(
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

func (s *SalaryService) GetORTRules(_ context.Context) (*domain.ORTRulesResponse, error) {
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

func (s *SalaryService) PreviewPayroll(
	ctx context.Context,
	params domain.PayrollPreviewParams,
) (*domain.PayrollPreview, error) {
	normalized, err := normalizePayrollPreviewParams(params)
	if err != nil {
		return nil, err
	}

	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, normalized.EmployeeID)
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

	workItems, err := s.repo.ListPayrollPreviewWorkItems(ctx, normalized)
	if err != nil {
		s.logError(ctx, "PreviewPayroll", "failed to list payroll work items", err,
			zap.String("employee_id", normalized.EmployeeID.String()),
		)
		return nil, fmt.Errorf("failed to list payroll work items: %w", err)
	}

	return s.buildPayrollPreview(ctx, employee, normalized, workItems)
}

func (s *SalaryService) ClosePayPeriod(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.ClosePayPeriodParams,
) (*domain.PayPeriod, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrSalaryInvalidRequest
	}

	normalized, err := normalizeClosePayPeriodParams(params)
	if err != nil {
		return nil, err
	}

	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, normalized.EmployeeID)
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
	err = s.repo.WithTxSalary(ctx, func(tx domain.SalaryTxRepository) error {
		existing, err := tx.GetPayPeriodByEmployeePeriod(
			ctx,
			normalized.EmployeeID,
			normalized.PeriodStart,
			normalized.PeriodEnd,
			normalized.PayrollGroup,
		)
		if err != nil && err != domain.ErrPayPeriodNotFound {
			return err
		}
		if existing != nil {
			return domain.ErrPayPeriodAlreadyExists
		}

		cutoffParams := domain.PayrollPreviewParams{
			EmployeeID:  normalized.EmployeeID,
			PeriodStart: normalized.PeriodStart,
			PeriodEnd:   normalized.CutoffAt,
		}

		overtimeIDs, err := tx.LockPayrollOvertimeEntries(ctx, cutoffParams)
		if err != nil {
			return err
		}

		workItems, err := tx.LockPayrollPreviewWorkItems(ctx, cutoffParams)
		if err != nil {
			return err
		}
		workItems = filterPayrollWorkItemsByPayrollGroup(workItems, normalized.PayrollGroup)

		fixedBaseLineItems := []domain.PayrollPreviewLineItem{}
		if normalized.PayrollGroup == domain.PayrollGroupFixed {
			if isPayrollPeriodRange(normalized.PeriodStart, normalized.PeriodEnd) {
				fixedBaseLineItems, err = s.buildFixedPeriodBasePreviewLineItems(
					ctx,
					normalized.EmployeeID,
					normalized.PeriodStart,
					normalized.PeriodEnd,
				)
			} else {
				fixedBaseLineItems, err = s.buildFixedBasePreviewLineItems(ctx, normalized.EmployeeID, normalized.PeriodStart, normalized.PeriodEnd)
			}
			if err != nil {
				return err
			}
		}
		if len(workItems) == 0 && len(fixedBaseLineItems) == 0 {
			return domain.ErrPayPeriodNoEntries
		}

		preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
			EmployeeID:  normalized.EmployeeID,
			PeriodStart: normalized.PeriodStart,
			PeriodEnd:   normalized.PeriodEnd,
		}, workItems)
		if err != nil {
			return err
		}
		applyFixedBaseLineItems(preview, fixedBaseLineItems)

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
				buildPayPeriodLineItem(item, workItems),
			)
			if err != nil {
				return err
			}
			created.LineItems = append(created.LineItems, *createdLine)
		}

		overtimeEntryIDs := uniquePreviewOvertimeEntryIDs(preview.LineItems)
		if len(overtimeEntryIDs) == 0 && len(overtimeIDs) > 0 {
			return domain.ErrPayPeriodNoEntries
		}
		if len(overtimeEntryIDs) > 0 {
			if err := tx.AssignOvertimeEntriesToPayPeriod(ctx, created.ID, overtimeEntryIDs); err != nil {
				return err
			}
		}

		scheduleIDs := uniquePreviewScheduleIDs(preview.LineItems)
		if len(scheduleIDs) > 0 {
			if err := tx.AssignSchedulesToPayPeriod(ctx, created.ID, scheduleIDs); err != nil {
				return err
			}
		}

		leavePayoutRequestIDs := uniquePreviewLeavePayoutRequestIDs(preview.LineItems)
		if len(leavePayoutRequestIDs) > 0 {
			if err := tx.AssignLeavePayoutRequestsToPayPeriod(ctx, created.ID, leavePayoutRequestIDs); err != nil {
				return err
			}
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SalaryService) GetPayPeriodByID(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	if payPeriodID == uuid.Nil {
		return nil, domain.ErrSalaryInvalidRequest
	}

	period, err := s.repo.GetPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}

	lineItems, err := s.repo.ListPayPeriodLineItems(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}
	period.LineItems = lineItems
	return period, nil
}

func (s *SalaryService) PreviewPayrollMonthClose(
	ctx context.Context,
	params domain.ClosePayrollMonthParams,
) (*domain.PayrollMonthCloseResult, error) {
	normalized, monthStart, _, err := normalizeClosePayrollMonthParams(params)
	if err != nil {
		return nil, err
	}

	result := &domain.PayrollMonthCloseResult{
		Month:        monthStart,
		PayrollGroup: normalized.PayrollGroup,
		CutoffAt:     normalized.CutoffAt,
	}
	selected := uuidSet(normalized.EmployeeIDs)

	if normalized.PayrollGroup == domain.PayrollGroupOnCall {
		page, err := s.GetOnCallPayrollMonthSummary(
			ctx,
			domain.PayrollMonthSummaryParams{Month: monthStart, Limit: 100000},
		)
		if err != nil {
			return nil, err
		}
		for _, row := range page.Items {
			if len(selected) > 0 && !selected[row.EmployeeID] {
				continue
			}
			item, err := s.previewPayrollMonthCloseEmployee(
				ctx,
				row.EmployeeID,
				row.EmployeeName,
				normalized,
			)
			if err != nil {
				return nil, err
			}
			item.PendingEntries = row.PendingEntryCount
			appendMonthClosePreviewItem(result, item)
		}
		return result, nil
	}

	page, err := s.GetFixedPayrollMonthSummary(
		ctx,
		domain.PayrollMonthSummaryParams{Month: monthStart, Limit: 100000},
	)
	if err != nil {
		return nil, err
	}
	for _, row := range page.Items {
		if len(selected) > 0 && !selected[row.EmployeeID] {
			continue
		}
		item, err := s.previewPayrollMonthCloseEmployee(
			ctx,
			row.EmployeeID,
			row.EmployeeName,
			normalized,
		)
		if err != nil {
			return nil, err
		}
		item.PendingEntries = row.PendingEntryCount
		appendMonthClosePreviewItem(result, item)
	}
	return result, nil
}

func (s *SalaryService) previewPayrollMonthCloseEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	employeeName string,
	params domain.ClosePayrollMonthParams,
) (domain.PayrollMonthCloseEmployeeResult, error) {
	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, employeeID)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	workItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		[]uuid.UUID{employeeID},
		params.Month,
		params.CutoffAt,
	)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	workItems = filterPayrollWorkItemsByPayrollGroup(workItems, params.PayrollGroup)
	fixedBaseLineItems := []domain.PayrollPreviewLineItem{}
	if params.PayrollGroup == domain.PayrollGroupFixed {
		fixedBaseLineItems, err = s.buildFixedBasePreviewLineItems(
			ctx,
			employeeID,
			params.Month,
			params.Month.AddDate(0, 1, -1),
		)
		if err != nil {
			return domain.PayrollMonthCloseEmployeeResult{}, err
		}
	}
	if len(workItems) == 0 && len(fixedBaseLineItems) == 0 {
		return domain.PayrollMonthCloseEmployeeResult{
			EmployeeID:   employeeID,
			EmployeeName: employeeName,
			Status:       "skipped",
			Reason:       domain.ErrPayPeriodNoEntries.Error(),
		}, nil
	}
	preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
		EmployeeID:  employeeID,
		PeriodStart: params.Month,
		PeriodEnd:   params.CutoffAt,
	}, workItems)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	applyFixedBaseLineItems(preview, fixedBaseLineItems)
	return domain.PayrollMonthCloseEmployeeResult{
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
		Status:       "ready",
		GrossAmount:  preview.GrossAmount,
	}, nil
}

func (s *SalaryService) ClosePayrollMonthByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.ClosePayrollMonthParams,
) (*domain.PayrollMonthCloseResult, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrSalaryInvalidRequest
	}
	preview, err := s.PreviewPayrollMonthClose(ctx, params)
	if err != nil {
		return nil, err
	}

	monthEnd := preview.Month.AddDate(0, 1, -1)
	result := &domain.PayrollMonthCloseResult{
		Month:        preview.Month,
		PayrollGroup: preview.PayrollGroup,
		CutoffAt:     preview.CutoffAt,
		Items:        make([]domain.PayrollMonthCloseEmployeeResult, 0, len(preview.Items)),
	}

	for _, item := range preview.Items {
		if item.Status != "ready" {
			result.Items = append(result.Items, item)
			result.SkippedCount++
			continue
		}
		period, err := s.ClosePayPeriod(ctx, adminEmployeeID, domain.ClosePayPeriodParams{
			EmployeeID:   item.EmployeeID,
			PeriodStart:  preview.Month,
			PeriodEnd:    monthEnd,
			PayrollGroup: preview.PayrollGroup,
			CutoffAt:     preview.CutoffAt,
		})
		if err != nil {
			status := "failed"
			if errors.Is(err, domain.ErrPayPeriodAlreadyExists) ||
				errors.Is(err, domain.ErrPayPeriodNoEntries) {
				status = "skipped"
			}
			result.Items = append(result.Items, domain.PayrollMonthCloseEmployeeResult{
				EmployeeID:   item.EmployeeID,
				EmployeeName: item.EmployeeName,
				Status:       status,
				Reason:       err.Error(),
			})
			if status == "skipped" {
				result.SkippedCount++
			} else {
				result.FailedCount++
			}
			continue
		}
		result.Items = append(result.Items, domain.PayrollMonthCloseEmployeeResult{
			EmployeeID:   item.EmployeeID,
			EmployeeName: item.EmployeeName,
			Status:       "closed",
			PayPeriodID:  &period.ID,
			GrossAmount:  period.GrossAmount,
		})
		result.ClosedCount++
	}

	return result, nil
}

func (s *SalaryService) PreviewPayrollPeriodClose(
	ctx context.Context,
	params domain.ClosePayrollPeriodParams,
) (*domain.PayrollPeriodCloseResult, error) {
	normalized, periodStart, periodEnd, err := normalizeClosePayrollPeriodParams(params)
	if err != nil {
		return nil, err
	}

	result := &domain.PayrollPeriodCloseResult{
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		PayrollGroup: normalized.PayrollGroup,
		CutoffAt:     normalized.CutoffAt,
	}
	selected := uuidSet(normalized.EmployeeIDs)

	if normalized.PayrollGroup == domain.PayrollGroupOnCall {
		page, err := s.GetOnCallPayrollPeriodSummary(
			ctx,
			domain.PayrollPeriodSummaryParams{
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Limit:       100000,
			},
		)
		if err != nil {
			return nil, err
		}
		for _, row := range page.Items {
			if len(selected) > 0 && !selected[row.EmployeeID] {
				continue
			}
			item, err := s.previewPayrollPeriodCloseEmployee(
				ctx,
				row.EmployeeID,
				row.EmployeeName,
				normalized,
			)
			if err != nil {
				return nil, err
			}
			item.PendingEntries = row.PendingEntryCount
			appendPeriodClosePreviewItem(result, item)
		}
		return result, nil
	}

	page, err := s.GetFixedPayrollPeriodSummary(
		ctx,
		domain.PayrollPeriodSummaryParams{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Limit:       100000,
		},
	)
	if err != nil {
		return nil, err
	}
	for _, row := range page.Items {
		if len(selected) > 0 && !selected[row.EmployeeID] {
			continue
		}
		item, err := s.previewPayrollPeriodCloseEmployee(
			ctx,
			row.EmployeeID,
			row.EmployeeName,
			normalized,
		)
		if err != nil {
			return nil, err
		}
		item.PendingEntries = row.PendingEntryCount
		appendPeriodClosePreviewItem(result, item)
	}
	return result, nil
}

func (s *SalaryService) previewPayrollPeriodCloseEmployee(
	ctx context.Context,
	employeeID uuid.UUID,
	employeeName string,
	params domain.ClosePayrollPeriodParams,
) (domain.PayrollMonthCloseEmployeeResult, error) {
	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, employeeID)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	workItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		[]uuid.UUID{employeeID},
		params.PeriodStart,
		params.CutoffAt,
	)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	workItems = filterPayrollWorkItemsByPayrollGroup(workItems, params.PayrollGroup)
	fixedBaseLineItems := []domain.PayrollPreviewLineItem{}
	if params.PayrollGroup == domain.PayrollGroupFixed {
		fixedBaseLineItems, err = s.buildFixedPeriodBasePreviewLineItems(
			ctx,
			employeeID,
			params.PeriodStart,
			params.PeriodEnd,
		)
		if err != nil {
			return domain.PayrollMonthCloseEmployeeResult{}, err
		}
	}
	if len(workItems) == 0 && len(fixedBaseLineItems) == 0 {
		return domain.PayrollMonthCloseEmployeeResult{
			EmployeeID:   employeeID,
			EmployeeName: employeeName,
			Status:       "skipped",
			Reason:       domain.ErrPayPeriodNoEntries.Error(),
		}, nil
	}
	preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
		EmployeeID:  employeeID,
		PeriodStart: params.PeriodStart,
		PeriodEnd:   params.CutoffAt,
	}, workItems)
	if err != nil {
		return domain.PayrollMonthCloseEmployeeResult{}, err
	}
	applyFixedBaseLineItems(preview, fixedBaseLineItems)
	return domain.PayrollMonthCloseEmployeeResult{
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
		Status:       "ready",
		GrossAmount:  preview.GrossAmount,
	}, nil
}

func (s *SalaryService) ClosePayrollPeriodByAdmin(
	ctx context.Context,
	adminEmployeeID uuid.UUID,
	params domain.ClosePayrollPeriodParams,
) (*domain.PayrollPeriodCloseResult, error) {
	if adminEmployeeID == uuid.Nil {
		return nil, domain.ErrSalaryInvalidRequest
	}
	preview, err := s.PreviewPayrollPeriodClose(ctx, params)
	if err != nil {
		return nil, err
	}

	result := &domain.PayrollPeriodCloseResult{
		PeriodStart:  preview.PeriodStart,
		PeriodEnd:    preview.PeriodEnd,
		PayrollGroup: preview.PayrollGroup,
		CutoffAt:     preview.CutoffAt,
		Items:        make([]domain.PayrollMonthCloseEmployeeResult, 0, len(preview.Items)),
	}

	for _, item := range preview.Items {
		if item.Status != "ready" {
			result.Items = append(result.Items, item)
			result.SkippedCount++
			continue
		}
		period, err := s.ClosePayPeriod(ctx, adminEmployeeID, domain.ClosePayPeriodParams{
			EmployeeID:   item.EmployeeID,
			PeriodStart:  preview.PeriodStart,
			PeriodEnd:    preview.PeriodEnd,
			PayrollGroup: preview.PayrollGroup,
			CutoffAt:     preview.CutoffAt,
		})
		if err != nil {
			status := "failed"
			if errors.Is(err, domain.ErrPayPeriodAlreadyExists) ||
				errors.Is(err, domain.ErrPayPeriodNoEntries) {
				status = "skipped"
			}
			result.Items = append(result.Items, domain.PayrollMonthCloseEmployeeResult{
				EmployeeID:   item.EmployeeID,
				EmployeeName: item.EmployeeName,
				Status:       status,
				Reason:       err.Error(),
			})
			if status == "skipped" {
				result.SkippedCount++
			} else {
				result.FailedCount++
			}
			continue
		}
		result.Items = append(result.Items, domain.PayrollMonthCloseEmployeeResult{
			EmployeeID:   item.EmployeeID,
			EmployeeName: item.EmployeeName,
			Status:       "closed",
			PayPeriodID:  &period.ID,
			GrossAmount:  period.GrossAmount,
		})
		result.ClosedCount++
	}

	return result, nil
}

func (s *SalaryService) loadPayPeriodWithLineItems(
	ctx context.Context,
	payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	period, err := s.repo.GetPayPeriodByID(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}

	lineItems, err := s.repo.ListPayPeriodLineItems(ctx, payPeriodID)
	if err != nil {
		return nil, err
	}
	period.LineItems = lineItems
	return period, nil
}

func (s *SalaryService) ListPayPeriods(
	ctx context.Context,
	params domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	if params.Status != nil && !isValidPayPeriodStatus(*params.Status) {
		return nil, domain.ErrSalaryInvalidRequest
	}
	return s.repo.ListPayPeriods(ctx, params)
}

func (s *SalaryService) GetFixedPayrollMonthSummary(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.FixedPayrollMonthSummaryPage, error) {
	normalized, monthStart, monthEnd, isCurrentMonth, err := normalizePayrollMonthSummaryParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	employees, totalCount, err := s.repo.ListFixedPayrollMonthEmployees(
		ctx,
		normalized,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollMonthSummary",
			"failed to list fixed payroll employees",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll employees: %w", err)
	}
	if len(employees) == 0 {
		return &domain.FixedPayrollMonthSummaryPage{
			Items:      []domain.FixedPayrollMonthSummaryRow{},
			TotalCount: totalCount,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	contractSources, err := s.repo.ListFixedPayrollContractSegments(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollMonthSummary",
			"failed to list fixed payroll contract segments",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll contract segments: %w", err)
	}
	contractSegmentsByEmployee := buildFixedPayrollContractSegmentsByEmployee(
		contractSources,
		monthStart,
		monthEnd,
	)

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollMonthSummary",
			"failed to list fixed payroll work items",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll work items: %w", err)
	}
	loondienst := "loondienst"
	fixedWorkItems := filterPayrollWorkItemsByContractType(approvedWorkItems, &loondienst)

	pendingEntries, err := s.repo.ListPayrollMonthPendingEntries(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollMonthSummary",
			"failed to list fixed payroll pending entries",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll pending entries: %w", err)
	}
	pendingByEmployee := buildPendingSummaryMap(pendingEntries, &loondienst)

	holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	adjustmentsByEmployee, err := buildFixedPayrollAdjustments(
		fixedWorkItems,
		holidaySet,
		fixedPayrollAsOf(monthStart, monthEnd, isCurrentMonth),
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FixedPayrollMonthSummaryRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.FixedPayrollMonthSummaryRow{
			EmployeeID:       employee.EmployeeID,
			EmployeeName:     employee.EmployeeName,
			Month:            normalized.Month,
			IsCurrentMonth:   isCurrentMonth,
			CalculationMode:  "fixed_contract",
			DataSource:       "live",
			ContractSegments: contractSegmentsByEmployee[employee.EmployeeID],
		}

		for _, segment := range row.ContractSegments {
			row.ContractBaseAmount = roundCurrency(row.ContractBaseAmount + segment.BaseAmount)
			row.ContractPaidMinutes = roundCurrency(
				row.ContractPaidMinutes + contractSegmentPaidMinutes(segment, monthEnd),
			)
		}

		if adjustment, ok := adjustmentsByEmployee[employee.EmployeeID]; ok {
			applyFixedPayrollAdjustmentSummary(&row, adjustment)
		}
		if pending, ok := pendingByEmployee[employee.EmployeeID]; ok {
			row.PendingEntryCount = pending.PendingEntryCount
			row.PendingWorkedMinutes = pending.PendingWorkedMinutes
		}

		row.PayableGrossAmount = roundCurrency(
			row.ContractBaseAmount +
				row.ActualORTAmount +
				row.ApprovedOvertimeAmount +
				row.LeavePayoutAmount,
		)
		row.ProjectedGrossAmount = roundCurrency(row.PayableGrossAmount + row.ForecastORTAmount)
		items = append(items, row)
	}

	return &domain.FixedPayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *SalaryService) GetFixedPayrollPeriodSummary(
	ctx context.Context,
	params domain.PayrollPeriodSummaryParams,
) (*domain.FixedPayrollMonthSummaryPage, error) {
	normalized, periodStart, periodEnd, isCurrentPeriod, err := normalizePayrollPeriodSummaryParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	listParams := domain.PayrollMonthSummaryParams{
		Limit:          normalized.Limit,
		Offset:         normalized.Offset,
		EmployeeSearch: normalized.EmployeeSearch,
		ContractType:   normalized.ContractType,
	}
	employees, totalCount, err := s.repo.ListFixedPayrollMonthEmployees(
		ctx,
		listParams,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollPeriodSummary",
			"failed to list fixed payroll employees",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll employees: %w", err)
	}
	if len(employees) == 0 {
		return &domain.FixedPayrollMonthSummaryPage{
			Items:      []domain.FixedPayrollMonthSummaryRow{},
			TotalCount: totalCount,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	contractSources, err := s.repo.ListFixedPayrollContractSegments(
		ctx,
		employeeIDs,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollPeriodSummary",
			"failed to list fixed payroll contract segments",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll contract segments: %w", err)
	}
	contractSegmentsByEmployee := buildFixedPayrollPeriodContractSegmentsByEmployee(
		contractSources,
		periodStart,
		periodEnd,
	)

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		employeeIDs,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollPeriodSummary",
			"failed to list fixed payroll work items",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll work items: %w", err)
	}
	loondienst := "loondienst"
	fixedWorkItems := filterPayrollWorkItemsByContractType(approvedWorkItems, &loondienst)

	pendingEntries, err := s.repo.ListPayrollMonthPendingEntries(
		ctx,
		employeeIDs,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetFixedPayrollPeriodSummary",
			"failed to list fixed payroll pending entries",
			err,
		)
		return nil, fmt.Errorf("failed to list fixed payroll pending entries: %w", err)
	}
	pendingByEmployee := buildPendingSummaryMap(pendingEntries, &loondienst)

	holidaySet, err := s.loadHolidaySet(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	adjustmentsByEmployee, err := buildFixedPayrollAdjustments(
		fixedWorkItems,
		holidaySet,
		fixedPayrollAsOf(periodStart, periodEnd, isCurrentPeriod),
	)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FixedPayrollMonthSummaryRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.FixedPayrollMonthSummaryRow{
			EmployeeID:       employee.EmployeeID,
			EmployeeName:     employee.EmployeeName,
			Month:            periodStart,
			IsCurrentMonth:   isCurrentPeriod,
			CalculationMode:  "fixed_contract_four_week",
			DataSource:       "live",
			ContractSegments: contractSegmentsByEmployee[employee.EmployeeID],
		}

		for _, segment := range row.ContractSegments {
			row.ContractBaseAmount = roundCurrency(row.ContractBaseAmount + segment.BaseAmount)
			row.ContractPaidMinutes = roundCurrency(
				row.ContractPaidMinutes + contractSegmentPeriodPaidMinutes(
					segment,
					periodStart,
					periodEnd,
				),
			)
		}

		if adjustment, ok := adjustmentsByEmployee[employee.EmployeeID]; ok {
			applyFixedPayrollAdjustmentSummary(&row, adjustment)
		}
		if pending, ok := pendingByEmployee[employee.EmployeeID]; ok {
			row.PendingEntryCount = pending.PendingEntryCount
			row.PendingWorkedMinutes = pending.PendingWorkedMinutes
		}

		row.PayableGrossAmount = roundCurrency(
			row.ContractBaseAmount +
				row.ActualORTAmount +
				row.ApprovedOvertimeAmount +
				row.LeavePayoutAmount,
		)
		row.ProjectedGrossAmount = roundCurrency(row.PayableGrossAmount + row.ForecastORTAmount)
		items = append(items, row)
	}

	return &domain.FixedPayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *SalaryService) GetOnCallPayrollMonthSummary(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.OnCallPayrollMonthSummaryPage, error) {
	normalized, monthStart, monthEnd, isCurrentMonth, err := normalizePayrollMonthSummaryParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	employees, totalCount, err := s.repo.ListOnCallPayrollMonthEmployees(
		ctx,
		normalized,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollMonthSummary",
			"failed to list on-call payroll employees",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll employees: %w", err)
	}
	if len(employees) == 0 {
		return &domain.OnCallPayrollMonthSummaryPage{
			Items:      []domain.OnCallPayrollMonthSummaryRow{},
			TotalCount: totalCount,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollMonthSummary",
			"failed to list on-call payroll work items",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll work items: %w", err)
	}
	onCall := "on_call"
	onCallWorkItems := filterPayrollWorkItemsByContractType(approvedWorkItems, &onCall)

	pendingEntries, err := s.repo.ListPayrollMonthPendingEntries(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollMonthSummary",
			"failed to list on-call payroll pending entries",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll pending entries: %w", err)
	}
	pendingByEmployee := buildPendingSummaryMap(pendingEntries, &onCall)

	holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	summariesByEmployee, err := buildOnCallPayrollSummaries(onCallWorkItems, holidaySet)
	if err != nil {
		return nil, err
	}

	items := make([]domain.OnCallPayrollMonthSummaryRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.OnCallPayrollMonthSummaryRow{
			EmployeeID:      employee.EmployeeID,
			EmployeeName:    employee.EmployeeName,
			Month:           normalized.Month,
			IsCurrentMonth:  isCurrentMonth,
			CalculationMode: "on_call_hours",
			DataSource:      "live",
		}

		if summary, ok := summariesByEmployee[employee.EmployeeID]; ok {
			applyOnCallPayrollSummary(&row, summary)
		}
		if pending, ok := pendingByEmployee[employee.EmployeeID]; ok {
			row.PendingEntryCount = pending.PendingEntryCount
			row.PendingWorkedMinutes = pending.PendingWorkedMinutes
		}

		row.PayableGrossAmount = roundCurrency(
			row.WorkedHoursAmount + row.ApprovedOvertimeAmount + row.LeavePayoutAmount,
		)
		if !shouldIncludeOnCallPayrollRow(row) {
			continue
		}
		items = append(items, row)
	}

	return &domain.OnCallPayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *SalaryService) GetOnCallPayrollPeriodSummary(
	ctx context.Context,
	params domain.PayrollPeriodSummaryParams,
) (*domain.OnCallPayrollMonthSummaryPage, error) {
	normalized, periodStart, periodEnd, isCurrentPeriod, err := normalizePayrollPeriodSummaryParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	listParams := domain.PayrollMonthSummaryParams{
		Limit:          normalized.Limit,
		Offset:         normalized.Offset,
		EmployeeSearch: normalized.EmployeeSearch,
		ContractType:   normalized.ContractType,
	}
	employees, totalCount, err := s.repo.ListOnCallPayrollMonthEmployees(
		ctx,
		listParams,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollPeriodSummary",
			"failed to list on-call payroll employees",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll employees: %w", err)
	}
	if len(employees) == 0 {
		return &domain.OnCallPayrollMonthSummaryPage{
			Items:      []domain.OnCallPayrollMonthSummaryRow{},
			TotalCount: totalCount,
		}, nil
	}

	employeeIDs := make([]uuid.UUID, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.EmployeeID)
	}

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		employeeIDs,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollPeriodSummary",
			"failed to list on-call payroll work items",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll work items: %w", err)
	}
	onCall := "on_call"
	onCallWorkItems := filterPayrollWorkItemsByContractType(approvedWorkItems, &onCall)

	pendingEntries, err := s.repo.ListPayrollMonthPendingEntries(
		ctx,
		employeeIDs,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetOnCallPayrollPeriodSummary",
			"failed to list on-call payroll pending entries",
			err,
		)
		return nil, fmt.Errorf("failed to list on-call payroll pending entries: %w", err)
	}
	pendingByEmployee := buildPendingSummaryMap(pendingEntries, &onCall)

	holidaySet, err := s.loadHolidaySet(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	summariesByEmployee, err := buildOnCallPayrollSummaries(onCallWorkItems, holidaySet)
	if err != nil {
		return nil, err
	}

	items := make([]domain.OnCallPayrollMonthSummaryRow, 0, len(employees))
	for _, employee := range employees {
		row := domain.OnCallPayrollMonthSummaryRow{
			EmployeeID:      employee.EmployeeID,
			EmployeeName:    employee.EmployeeName,
			Month:           periodStart,
			IsCurrentMonth:  isCurrentPeriod,
			CalculationMode: "on_call_four_week",
			DataSource:      "live",
		}

		if summary, ok := summariesByEmployee[employee.EmployeeID]; ok {
			applyOnCallPayrollSummary(&row, summary)
		}
		if pending, ok := pendingByEmployee[employee.EmployeeID]; ok {
			row.PendingEntryCount = pending.PendingEntryCount
			row.PendingWorkedMinutes = pending.PendingWorkedMinutes
		}
		row.PayableGrossAmount = roundCurrency(
			row.WorkedHoursAmount +
				row.ApprovedOvertimeAmount +
				row.LeavePayoutAmount,
		)
		if !shouldIncludeOnCallPayrollRow(row) {
			continue
		}
		items = append(items, row)
	}

	return &domain.OnCallPayrollMonthSummaryPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (s *SalaryService) GetFixedPayrollMonthStats(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthStats, error) {
	params.Limit = 100000
	params.Offset = 0

	page, err := s.GetFixedPayrollMonthSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	stats := &domain.PayrollMonthStats{Month: params.Month}
	for _, item := range page.Items {
		stats.TotalBaseContractPay = roundCurrency(
			stats.TotalBaseContractPay + item.ContractBaseAmount,
		)
		stats.TotalORTPay = roundCurrency(stats.TotalORTPay + item.ActualORTAmount)
		stats.TotalOvertimePay = roundCurrency(stats.TotalOvertimePay + item.ApprovedOvertimeAmount)
		stats.TotalRequestedLeaveHoursPay = roundCurrency(
			stats.TotalRequestedLeaveHoursPay + item.LeavePayoutAmount,
		)
		stats.TotalRequestedLeaveHours = roundCurrency(
			stats.TotalRequestedLeaveHours + float64(item.LeavePayoutMinutes)/60,
		)
		stats.TotalGrossPayable = roundCurrency(stats.TotalGrossPayable + item.PayableGrossAmount)
	}

	return stats, nil
}

func (s *SalaryService) GetFixedPayrollPeriodStats(
	ctx context.Context,
	params domain.PayrollPeriodSummaryParams,
) (*domain.PayrollMonthStats, error) {
	params.Limit = 100000
	params.Offset = 0

	page, err := s.GetFixedPayrollPeriodSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	stats := &domain.PayrollMonthStats{Month: params.PeriodStart}
	for _, item := range page.Items {
		stats.TotalBaseContractPay = roundCurrency(
			stats.TotalBaseContractPay + item.ContractBaseAmount,
		)
		stats.TotalORTPay = roundCurrency(stats.TotalORTPay + item.ActualORTAmount)
		stats.TotalOvertimePay = roundCurrency(stats.TotalOvertimePay + item.ApprovedOvertimeAmount)
		stats.TotalRequestedLeaveHoursPay = roundCurrency(
			stats.TotalRequestedLeaveHoursPay + item.LeavePayoutAmount,
		)
		stats.TotalRequestedLeaveHours = roundCurrency(
			stats.TotalRequestedLeaveHours + float64(item.LeavePayoutMinutes)/60,
		)
		stats.TotalGrossPayable = roundCurrency(stats.TotalGrossPayable + item.PayableGrossAmount)
	}

	return stats, nil
}

func (s *SalaryService) GetOnCallPayrollMonthStats(
	ctx context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthStats, error) {
	params.Limit = 100000
	params.Offset = 0

	page, err := s.GetOnCallPayrollMonthSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	stats := &domain.PayrollMonthStats{Month: params.Month}
	for _, item := range page.Items {
		stats.TotalBaseContractPay = roundCurrency(
			stats.TotalBaseContractPay + item.WorkedHoursAmount,
		)
		stats.TotalOvertimePay = roundCurrency(stats.TotalOvertimePay + item.ApprovedOvertimeAmount)
		stats.TotalRequestedLeaveHoursPay = roundCurrency(
			stats.TotalRequestedLeaveHoursPay + item.LeavePayoutAmount,
		)
		stats.TotalRequestedLeaveHours = roundCurrency(
			stats.TotalRequestedLeaveHours + float64(item.LeavePayoutMinutes)/60,
		)
		stats.TotalGrossPayable = roundCurrency(stats.TotalGrossPayable + item.PayableGrossAmount)
	}

	return stats, nil
}

func (s *SalaryService) GetOnCallPayrollPeriodStats(
	ctx context.Context,
	params domain.PayrollPeriodSummaryParams,
) (*domain.PayrollMonthStats, error) {
	params.Limit = 100000
	params.Offset = 0

	page, err := s.GetOnCallPayrollPeriodSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	stats := &domain.PayrollMonthStats{Month: params.PeriodStart}
	for _, item := range page.Items {
		stats.TotalBaseContractPay = roundCurrency(
			stats.TotalBaseContractPay + item.WorkedHoursAmount,
		)
		stats.TotalOvertimePay = roundCurrency(stats.TotalOvertimePay + item.ApprovedOvertimeAmount)
		stats.TotalRequestedLeaveHoursPay = roundCurrency(
			stats.TotalRequestedLeaveHoursPay + item.LeavePayoutAmount,
		)
		stats.TotalRequestedLeaveHours = roundCurrency(
			stats.TotalRequestedLeaveHours + float64(item.LeavePayoutMinutes)/60,
		)
		stats.TotalGrossPayable = roundCurrency(stats.TotalGrossPayable + item.PayableGrossAmount)
	}

	return stats, nil
}

func (s *SalaryService) GetPayrollPeriodOptions(
	_ context.Context,
	year *int,
) ([]domain.PayrollPeriodOption, error) {
	selectedYear := domain.CurrentPayrollISOYear(time.Now().UTC())
	if year != nil {
		if *year <= 0 {
			return nil, domain.ErrSalaryInvalidRequest
		}
		selectedYear = *year
	}
	return domain.PayrollPeriodOptionsForYear(selectedYear, time.Now().UTC()), nil
}

func (s *SalaryService) GetPayrollMonthORTOverview(
	ctx context.Context,
	params domain.PayrollMonthORTOverviewParams,
) (*domain.PayrollMonthORTOverviewPage, error) {
	normalized, monthStart, monthEnd, isCurrentMonth, err := normalizePayrollMonthORTOverviewParams(
		params,
	)
	if err != nil {
		return nil, err
	}

	employees, err := s.repo.ListPayrollMonthEmployeesAll(ctx, normalized, monthStart, monthEnd)
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

	lockedPayPeriods, err := s.repo.ListPayPeriodsByEmployeesAndRange(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
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

	lockedDistributionByPeriod := make(
		map[uuid.UUID][]domain.PayrollMultiplierSummary,
		len(payPeriodIDs),
	)
	if len(payPeriodIDs) > 0 {
		lockedSummaries, err := s.repo.ListPayrollMonthLockedMultiplierSummaries(ctx, payPeriodIDs)
		if err != nil {
			s.logError(
				ctx,
				"GetPayrollMonthORTOverview",
				"failed to list locked pay period multiplier summaries",
				err,
			)
			return nil, fmt.Errorf(
				"failed to list locked payroll month multiplier summaries: %w",
				err,
			)
		}
		lockedDistributionByPeriod = buildLockedORTDistributionMap(lockedSummaries)
	}

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		employeeIDs,
		monthStart,
		monthEnd,
	)
	if err != nil {
		s.logError(
			ctx,
			"GetPayrollMonthORTOverview",
			"failed to list approved payroll work items",
			err,
		)
		return nil, fmt.Errorf("failed to list approved payroll month work items: %w", err)
	}

	liveSummaries := make(map[uuid.UUID]payrollMonthLiveSummary)
	if len(approvedWorkItems) > 0 {
		holidaySet, err := s.loadHolidaySet(ctx, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		liveSummaries, err = buildPayrollMonthLiveSummaries(approvedWorkItems, holidaySet)
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

func (s *SalaryService) GetPayrollMonthDetail(
	ctx context.Context,
	employeeID uuid.UUID,
	month time.Time,
	contractType *string,
) (*domain.PayrollMonthDetail, error) {
	if employeeID == uuid.Nil || month.IsZero() {
		return nil, domain.ErrSalaryInvalidRequest
	}
	normalizedContractType, err := normalizePayrollContractType(contractType)
	if err != nil {
		return nil, err
	}

	monthStart := time.Date(month.UTC().Year(), month.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)

	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	payPeriods, err := s.repo.ListPayPeriodsByEmployeesAndRange(
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
		filteredPayPeriod := filterPayPeriodByContractType(
			selectedPayPeriod,
			normalizedContractType,
		)
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

	approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
		ctx,
		[]uuid.UUID{employeeID},
		monthStart,
		monthEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list approved payroll work items for detail: %w", err)
	}
	approvedWorkItems = filterPayrollWorkItemsByContractType(
		approvedWorkItems,
		normalizedContractType,
	)

	preview, err := s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
		EmployeeID:  employeeID,
		PeriodStart: monthStart,
		PeriodEnd:   monthEnd,
	}, approvedWorkItems)
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

func (s *SalaryService) ExportPayrollMonthPDF(
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

func (s *SalaryService) GetMySalaryPage(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) (*domain.SalaryPageData, error) {
	if employeeID == uuid.Nil || periodStart.IsZero() || periodEnd.IsZero() {
		return nil, domain.ErrSalaryInvalidRequest
	}

	monthStart := time.Date(
		periodStart.UTC().Year(),
		periodStart.UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	employee, err := s.repo.GetPayrollPreviewEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	periodContractSegments, err := s.repo.ListFixedPayrollContractSegments(
		ctx,
		[]uuid.UUID{employeeID},
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(ctx, "GetMySalaryPage", "failed to list fixed payroll contract segments", err)
		return nil, fmt.Errorf("failed to list fixed payroll contract segments: %w", err)
	}
	applySalaryPagePeriodContract(employee, periodContractSegments)

	payPeriods, err := s.repo.ListPayPeriodsByEmployeesAndRange(
		ctx, []uuid.UUID{employeeID}, periodStart, periodEnd,
	)
	if err != nil {
		s.logError(ctx, "GetMySalaryPage", "failed to list pay periods", err)
		return nil, fmt.Errorf("failed to list pay periods: %w", err)
	}

	var selectedPayPeriod *domain.PayPeriod
	for _, period := range payPeriods {
		if period.EmployeeID != employeeID {
			continue
		}
		if period.PeriodStart.Equal(periodStart) && period.PeriodEnd.Equal(periodEnd) {
			item, loadErr := s.loadPayPeriodWithLineItems(ctx, period.ID)
			if loadErr != nil {
				return nil, loadErr
			}
			selectedPayPeriod = item
			break
		}
	}

	var dataSource string
	var payPeriod *domain.PayPeriod
	var preview *domain.PayrollPreview

	if selectedPayPeriod != nil {
		dataSource = "locked"
		payPeriod = selectedPayPeriod
		if isLoondienstContractType(employee.ContractType) {
			var scheduleBaseSum float64
			for i, item := range payPeriod.LineItems {
				if item.LineType == string(domain.PayrollSourceSchedule) {
					scheduleBaseSum = roundCurrency(scheduleBaseSum + item.BaseAmount)
					payPeriod.LineItems[i].BaseAmount = 0
				}
			}
			payPeriod.BaseGrossAmount = roundCurrency(payPeriod.BaseGrossAmount - scheduleBaseSum)
			payPeriod.GrossAmount = roundCurrency(
				payPeriod.BaseGrossAmount + payPeriod.IrregularGrossAmount,
			)
		}
	} else {
		dataSource = "live"
		approvedWorkItems, err := s.repo.ListPayrollMonthApprovedWorkItems(
			ctx, []uuid.UUID{employeeID}, periodStart, periodEnd,
		)
		if err != nil {
			s.logError(ctx, "GetMySalaryPage", "failed to list approved work items", err)
			return nil, fmt.Errorf("failed to list approved work items: %w", err)
		}

		fixedBaseLineItems := []domain.PayrollPreviewLineItem{}
		if isLoondienstContractType(employee.ContractType) {
			fixedBaseLineItems, err = s.buildFixedPeriodBasePreviewLineItems(ctx, employeeID, periodStart, periodEnd)
			if err != nil {
				return nil, err
			}
		}

		if len(approvedWorkItems) > 0 || len(fixedBaseLineItems) > 0 {
			preview, err = s.buildPayrollPreview(ctx, employee, domain.PayrollPreviewParams{
				EmployeeID:  employeeID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}, approvedWorkItems)
			if err != nil {
				return nil, err
			}

			if isLoondienstContractType(employee.ContractType) {
				for i, item := range preview.LineItems {
					if item.SourceType == domain.PayrollSourceSchedule {
						preview.LineItems[i].BaseAmount = 0
					}
				}
				var newBaseGross float64
				for _, item := range preview.LineItems {
					if item.SourceType != domain.PayrollSourceSchedule {
						newBaseGross = roundCurrency(newBaseGross + item.BaseAmount)
					}
				}
				preview.BaseGrossAmount = newBaseGross
				applyFixedBaseLineItems(preview, fixedBaseLineItems)
			}
		}
	}

	pendingEntries, err := s.repo.ListPendingOvertimeEntriesDetail(
		ctx,
		employeeID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		s.logError(ctx, "GetMySalaryPage", "failed to list pending entries", err)
		return nil, fmt.Errorf("failed to list pending entries: %w", err)
	}
	if pendingEntries == nil {
		pendingEntries = []domain.PayrollPendingEntryDetail{}
	}

	leavePayouts, err := s.repo.ListPayoutRequestsByEmployeeAndMonth(ctx, employeeID, monthStart)
	if err != nil {
		s.logError(ctx, "GetMySalaryPage", "failed to list leave payouts", err)
		return nil, fmt.Errorf("failed to list leave payouts: %w", err)
	}
	if leavePayouts == nil {
		leavePayouts = []domain.PayoutRequest{}
	}

	extraRemaining := int32(0)

	return &domain.SalaryPageData{
		EmployeeID:            employeeID,
		EmployeeName:          strings.TrimSpace(employee.FirstName + " " + employee.LastName),
		Month:                 monthStart,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
		ContractType:          employee.ContractType,
		ContractRate:          employee.ContractRate,
		ContractHours:         employee.ContractHours,
		IrregularHoursProfile: "",
		ContractStartDate:     employee.ContractStartDate,
		ContractEndDate:       employee.ContractEndDate,
		DataSource:            dataSource,
		PayPeriod:             payPeriod,
		Preview:               preview,
		PendingEntries:        pendingEntries,
		LeavePayoutRequests:   leavePayouts,
		ExtraLeaveRemaining:   extraRemaining,
	}, nil
}

func applySalaryPagePeriodContract(
	employee *domain.EmployeeDetail,
	segments []domain.FixedPayrollContractSegmentSource,
) {
	if employee == nil || len(segments) == 0 {
		return
	}

	segment := segments[0]
	for _, candidate := range segments[1:] {
		if candidate.ActiveFrom.Before(segment.ActiveFrom) {
			segment = candidate
		}
	}

	employee.ContractType = segment.ContractType
	employee.ContractRate = &segment.HourlyRate
	employee.ContractHours = &segment.HoursPerWeek
	employee.ContractStartDate = &segment.ActiveFrom
	employee.ContractEndDate = &segment.ActiveUntil
}

func (s *SalaryService) MarkPayPeriodPaidByAdmin(
	ctx context.Context,
	adminEmployeeID, payPeriodID uuid.UUID,
) (*domain.PayPeriod, error) {
	if adminEmployeeID == uuid.Nil || payPeriodID == uuid.Nil {
		return nil, domain.ErrSalaryInvalidRequest
	}

	var updated *domain.PayPeriod
	err := s.repo.WithTxSalary(ctx, func(tx domain.SalaryTxRepository) error {
		current, err := tx.GetPayPeriodForUpdate(ctx, payPeriodID)
		if err != nil {
			return err
		}
		if current.Status != domain.PayPeriodStatusDraft {
			return domain.ErrPayPeriodStateInvalid
		}

		updated, err = tx.MarkPayPeriodPaid(ctx, payPeriodID)
		if err != nil {
			return err
		}
		return tx.MarkLeavePayoutRequestsPaidByPayPeriod(ctx, payPeriodID, adminEmployeeID)
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// --- Private methods on *SalaryService ---

func (s *SalaryService) buildPayrollPreview(
	ctx context.Context,
	employee *domain.EmployeeDetail,
	params domain.PayrollPreviewParams,
	workItems []domain.PayrollWorkItem,
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

	for _, item := range workItems {
		if !isPayrollEligibleContractType(item.ContractType) {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if item.ContractRate == nil || *item.ContractRate <= 0 {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(item.IrregularHoursProfile) {
			return nil, domain.ErrSalaryInvalidRequest
		}

		var lineItems []domain.PayrollPreviewLineItem
		var workedMinutes int32
		var baseAmount float64
		var premiumAmount float64

		if item.SourceType == domain.PayrollSourceLeavePayout {
			lineItems, workedMinutes, baseAmount, premiumAmount, err = buildLeavePayoutLineItems(
				item,
			)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
		} else if item.SourceType == domain.PayrollSourceOvertime && item.StartTime == "" {
			lineItems, workedMinutes, baseAmount, premiumAmount = buildSimpleOvertimeLineItems(item, *item.ContractRate)
		} else {
			lineItems, workedMinutes, baseAmount, premiumAmount, err = buildPayrollPreviewLineItems(item, *item.ContractRate, holidaySet)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
		}

		preview.TotalWorkedMinutes += workedMinutes
		preview.BaseGrossAmount = roundCurrency(preview.BaseGrossAmount + baseAmount)
		preview.IrregularGrossAmount = roundCurrency(preview.IrregularGrossAmount + premiumAmount)
		preview.LineItems = append(preview.LineItems, lineItems...)
	}

	preview.GrossAmount = roundCurrency(preview.BaseGrossAmount + preview.IrregularGrossAmount)
	return preview, nil
}

func (s *SalaryService) buildFixedBasePreviewLineItems(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) ([]domain.PayrollPreviewLineItem, error) {
	segments, err := s.repo.ListFixedPayrollContractSegments(
		ctx,
		[]uuid.UUID{employeeID},
		periodStart,
		periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list fixed payroll contract segments: %w", err)
	}
	byEmployee := buildFixedPayrollContractSegmentsByEmployee(segments, periodStart, periodEnd)
	contractSegments := byEmployee[employeeID]
	lineItems := make([]domain.PayrollPreviewLineItem, 0, len(contractSegments))
	for _, segment := range contractSegments {
		lineItems = append(lineItems, domain.PayrollPreviewLineItem{
			SourceType:            "fixed_base",
			Label:                 "Fixed monthly base salary",
			ContractType:          segment.ContractType,
			WorkDate:              segment.ActiveFrom,
			IrregularHoursProfile: "none",
			BaseAmount:            segment.BaseAmount,
			PaidMinutes:           contractSegmentPaidMinutes(segment, periodEnd),
		})
	}
	return lineItems, nil
}

func (s *SalaryService) buildFixedPeriodBasePreviewLineItems(
	ctx context.Context,
	employeeID uuid.UUID,
	periodStart, periodEnd time.Time,
) ([]domain.PayrollPreviewLineItem, error) {
	segments, err := s.repo.ListFixedPayrollContractSegments(
		ctx,
		[]uuid.UUID{employeeID},
		periodStart,
		periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list fixed payroll contract segments: %w", err)
	}
	byEmployee := buildFixedPayrollPeriodContractSegmentsByEmployee(
		segments,
		periodStart,
		periodEnd,
	)
	contractSegments := byEmployee[employeeID]
	lineItems := make([]domain.PayrollPreviewLineItem, 0, len(contractSegments))
	for _, segment := range contractSegments {
		lineItems = append(lineItems, domain.PayrollPreviewLineItem{
			SourceType:            "fixed_base",
			Label:                 "Fixed 4-week base salary",
			ContractType:          segment.ContractType,
			WorkDate:              segment.ActiveFrom,
			IrregularHoursProfile: "none",
			BaseAmount:            segment.BaseAmount,
			PaidMinutes: contractSegmentPeriodPaidMinutes(
				segment,
				periodStart,
				periodEnd,
			),
		})
	}
	return lineItems, nil
}

func applyFixedBaseLineItems(
	preview *domain.PayrollPreview,
	lineItems []domain.PayrollPreviewLineItem,
) {
	for _, item := range lineItems {
		preview.BaseGrossAmount = roundCurrency(preview.BaseGrossAmount + item.BaseAmount)
		preview.LineItems = append(preview.LineItems, item)
	}
	preview.GrossAmount = roundCurrency(preview.BaseGrossAmount + preview.IrregularGrossAmount)
}

func (s *SalaryService) buildLockedSnapshotMap(
	ctx context.Context,
	payPeriods []domain.PayPeriod,
	contractType *string,
) (map[uuid.UUID]lockedPayrollSnapshot, error) {
	snapshots := make(map[uuid.UUID]lockedPayrollSnapshot, len(payPeriods))
	for _, period := range payPeriods {
		lineItems, err := s.repo.ListPayPeriodLineItems(ctx, period.ID)
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

func (s *SalaryService) loadHolidaySet(
	ctx context.Context,
	startDate, endDate time.Time,
) (map[string]struct{}, error) {
	holidays, err := s.repo.ListNationalHolidays(
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

// --- Standalone helper functions ---

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
	case "loondienst", "permanent", "temporary":
		normalized := "LOONDIENST"
		return &normalized, nil
	case "zzp", "on_call":
		normalized := "ZZP"
		return &normalized, nil
	default:
		return nil, domain.ErrSalaryInvalidRequest
	}
}

func normalizePayrollPreviewParams(
	params domain.PayrollPreviewParams,
) (domain.PayrollPreviewParams, error) {
	if params.EmployeeID == uuid.Nil || params.PeriodStart.IsZero() || params.PeriodEnd.IsZero() {
		return domain.PayrollPreviewParams{}, domain.ErrSalaryInvalidRequest
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
		return domain.PayrollPreviewParams{}, domain.ErrSalaryInvalidRequest
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
		EmployeeID:   normalized.EmployeeID,
		PeriodStart:  normalized.PeriodStart,
		PeriodEnd:    normalized.PeriodEnd,
		PayrollGroup: normalizePayrollGroupOrDefault(params.PayrollGroup),
		CutoffAt:     normalizePayPeriodCutoff(params.CutoffAt, normalized.PeriodEnd),
	}, nil
}

func normalizePayrollGroupOrDefault(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case domain.PayrollGroupOnCall:
		return domain.PayrollGroupOnCall
	default:
		return domain.PayrollGroupFixed
	}
}

func normalizePayPeriodCutoff(cutoffAt, periodEnd time.Time) time.Time {
	if cutoffAt.IsZero() {
		return time.Date(
			periodEnd.UTC().Year(),
			periodEnd.UTC().Month(),
			periodEnd.UTC().Day(),
			23,
			59,
			59,
			0,
			time.UTC,
		)
	}
	return cutoffAt.UTC()
}

func normalizePayrollMonthSummaryParams(
	params domain.PayrollMonthSummaryParams,
) (domain.PayrollMonthSummaryParams, time.Time, time.Time, bool, error) {
	if params.Month.IsZero() {
		return domain.PayrollMonthSummaryParams{}, time.Time{}, time.Time{}, false, domain.ErrSalaryInvalidRequest
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

func normalizePayrollPeriodSummaryParams(
	params domain.PayrollPeriodSummaryParams,
) (domain.PayrollPeriodSummaryParams, time.Time, time.Time, bool, error) {
	if params.PeriodStart.IsZero() || params.PeriodEnd.IsZero() {
		return domain.PayrollPeriodSummaryParams{}, time.Time{}, time.Time{}, false, domain.ErrSalaryInvalidRequest
	}

	normalizedContractType, err := normalizePayrollContractType(params.ContractType)
	if err != nil {
		return domain.PayrollPeriodSummaryParams{}, time.Time{}, time.Time{}, false, err
	}
	params.ContractType = normalizedContractType

	periodStart := time.Date(
		params.PeriodStart.UTC().Year(),
		params.PeriodStart.UTC().Month(),
		params.PeriodStart.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	periodEnd := time.Date(
		params.PeriodEnd.UTC().Year(),
		params.PeriodEnd.UTC().Month(),
		params.PeriodEnd.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if !isPayrollPeriodRange(periodStart, periodEnd) {
		return domain.PayrollPeriodSummaryParams{}, time.Time{}, time.Time{}, false, domain.ErrSalaryInvalidRequest
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	currentPeriod := !today.Before(periodStart) && !today.After(periodEnd)

	params.PeriodStart = periodStart
	params.PeriodEnd = periodEnd
	return params, periodStart, periodEnd, currentPeriod, nil
}

func normalizeClosePayrollMonthParams(
	params domain.ClosePayrollMonthParams,
) (domain.ClosePayrollMonthParams, time.Time, time.Time, error) {
	if params.Month.IsZero() {
		return domain.ClosePayrollMonthParams{}, time.Time{}, time.Time{}, domain.ErrSalaryInvalidRequest
	}
	monthStart := time.Date(
		params.Month.UTC().Year(),
		params.Month.UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	monthEnd := monthStart.AddDate(0, 1, -1)
	payrollGroup := normalizePayrollGroupOrDefault(params.PayrollGroup)
	cutoffAt := normalizePayPeriodCutoff(params.CutoffAt, monthEnd)
	if cutoffAt.Before(monthStart) || cutoffAt.After(monthEnd.AddDate(0, 0, 1)) {
		return domain.ClosePayrollMonthParams{}, time.Time{}, time.Time{}, domain.ErrSalaryInvalidRequest
	}
	return domain.ClosePayrollMonthParams{
		PayrollGroup: payrollGroup,
		Month:        monthStart,
		EmployeeIDs:  params.EmployeeIDs,
		CutoffAt:     cutoffAt,
	}, monthStart, monthEnd, nil
}

func normalizeClosePayrollPeriodParams(
	params domain.ClosePayrollPeriodParams,
) (domain.ClosePayrollPeriodParams, time.Time, time.Time, error) {
	if params.PeriodStart.IsZero() || params.PeriodEnd.IsZero() {
		return domain.ClosePayrollPeriodParams{}, time.Time{}, time.Time{}, domain.ErrSalaryInvalidRequest
	}
	periodStart := time.Date(
		params.PeriodStart.UTC().Year(),
		params.PeriodStart.UTC().Month(),
		params.PeriodStart.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	periodEnd := time.Date(
		params.PeriodEnd.UTC().Year(),
		params.PeriodEnd.UTC().Month(),
		params.PeriodEnd.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if !isPayrollPeriodRange(periodStart, periodEnd) {
		return domain.ClosePayrollPeriodParams{}, time.Time{}, time.Time{}, domain.ErrSalaryInvalidRequest
	}
	payrollGroup := normalizePayrollGroupOrDefault(params.PayrollGroup)
	cutoffAt := normalizePayPeriodCutoff(params.CutoffAt, periodEnd)
	if cutoffAt.Before(periodStart) || cutoffAt.After(periodEnd.AddDate(0, 0, 1)) {
		return domain.ClosePayrollPeriodParams{}, time.Time{}, time.Time{}, domain.ErrSalaryInvalidRequest
	}
	return domain.ClosePayrollPeriodParams{
		PayrollGroup: payrollGroup,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		EmployeeIDs:  params.EmployeeIDs,
		CutoffAt:     cutoffAt,
	}, periodStart, periodEnd, nil
}

func isPayrollPeriodRange(periodStart, periodEnd time.Time) bool {
	if periodEnd.Before(periodStart) || !domain.IsPayrollPeriodStart(periodStart) {
		return false
	}
	expectedStart, expectedEnd := domain.ResolvePayrollPeriod(periodStart)
	return periodStart.Equal(expectedStart) && periodEnd.Equal(expectedEnd)
}

func appendMonthClosePreviewItem(
	result *domain.PayrollMonthCloseResult,
	item domain.PayrollMonthCloseEmployeeResult,
) {
	if item.Status == "" {
		item.Status = "ready"
	}
	if item.GrossAmount <= 0 {
		item.Status = "skipped"
		item.Reason = domain.ErrPayPeriodNoEntries.Error()
		result.SkippedCount++
	} else {
		result.ClosedCount++
	}
	result.Items = append(result.Items, item)
}

func appendPeriodClosePreviewItem(
	result *domain.PayrollPeriodCloseResult,
	item domain.PayrollMonthCloseEmployeeResult,
) {
	if item.Status == "" {
		item.Status = "ready"
	}
	if item.GrossAmount <= 0 {
		item.Status = "skipped"
		item.Reason = domain.ErrPayPeriodNoEntries.Error()
		result.SkippedCount++
	} else {
		result.ClosedCount++
	}
	result.Items = append(result.Items, item)
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]bool {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if id != uuid.Nil {
			set[id] = true
		}
	}
	return set
}

func normalizePayrollMonthORTOverviewParams(
	params domain.PayrollMonthORTOverviewParams,
) (domain.PayrollMonthORTOverviewParams, time.Time, time.Time, bool, error) {
	if params.Month.IsZero() {
		return domain.PayrollMonthORTOverviewParams{}, time.Time{}, time.Time{}, false, domain.ErrSalaryInvalidRequest
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

func buildSimpleOvertimeLineItems(
	item domain.PayrollWorkItem,
	hourlyRate float64,
) ([]domain.PayrollPreviewLineItem, int32, float64, float64) {
	totalMinutes := int32(math.Round(item.MinutesWorked))
	paidMinutes := float64(totalMinutes)
	baseAmount := roundCurrency(hourlyRate * paidMinutes / 60)

	lineItem := domain.PayrollPreviewLineItem{
		ScheduleID:            item.ScheduleID,
		OvertimeEntryID:       item.OvertimeEntryID,
		SourceType:            item.SourceType,
		Label:                 item.Label,
		ContractType:          item.ContractType,
		WorkDate:              item.WorkDate,
		StartTime:             "",
		EndTime:               "",
		BreakMinutes:          0,
		IrregularHoursProfile: item.IrregularHoursProfile,
		AppliedRatePercent:    0,
		MinutesWorked:         totalMinutes,
		PaidMinutes:           paidMinutes,
		BaseAmount:            baseAmount,
		PremiumAmount:         0,
	}

	return []domain.PayrollPreviewLineItem{lineItem}, totalMinutes, baseAmount, 0
}

func buildLeavePayoutLineItems(
	item domain.PayrollWorkItem,
) ([]domain.PayrollPreviewLineItem, int32, float64, float64, error) {
	if item.LeavePayoutRequestID == nil || item.GrossAmountOverride == nil ||
		*item.GrossAmountOverride <= 0 {
		return nil, 0, 0, 0, domain.ErrSalaryInvalidRequest
	}

	totalMinutes := int32(math.Round(item.MinutesWorked))
	if totalMinutes <= 0 {
		return nil, 0, 0, 0, domain.ErrSalaryInvalidRequest
	}
	baseAmount := roundCurrency(*item.GrossAmountOverride)
	paidMinutes := float64(totalMinutes)

	lineItem := domain.PayrollPreviewLineItem{
		LeavePayoutRequestID:  item.LeavePayoutRequestID,
		SourceType:            item.SourceType,
		Label:                 item.Label,
		ContractType:          item.ContractType,
		WorkDate:              item.WorkDate,
		StartTime:             "",
		EndTime:               "",
		BreakMinutes:          0,
		IrregularHoursProfile: item.IrregularHoursProfile,
		AppliedRatePercent:    0,
		MinutesWorked:         0,
		PaidMinutes:           paidMinutes,
		BaseAmount:            baseAmount,
		PremiumAmount:         0,
	}

	return []domain.PayrollPreviewLineItem{lineItem}, 0, baseAmount, 0, nil
}

func buildPayrollPreviewLineItems(
	item domain.PayrollWorkItem,
	hourlyRate float64,
	holidaySet map[string]struct{},
) ([]domain.PayrollPreviewLineItem, int32, float64, float64, error) {
	start, end, err := parseWorkItemBounds(item.WorkDate, item.StartTime, item.EndTime)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	totalMinutes := int32(end.Sub(start).Minutes())
	if totalMinutes <= 0 {
		return nil, 0, 0, 0, domain.ErrSalaryInvalidRequest
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
	segmentRate := appliedPayrollRateForMinute(item, current, holidaySet)
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
			nextRate = appliedPayrollRateForMinute(item, next, holidaySet)
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
		paidMinutes := float64(segment.minutes)
		baseAmount := roundCurrency(hourlyRate * paidMinutes / 60)
		premiumAmount := roundCurrency(baseAmount * segment.rate / 100)

		baseTotal = roundCurrency(baseTotal + baseAmount)
		premiumTotal = roundCurrency(premiumTotal + premiumAmount)

		items = append(items, domain.PayrollPreviewLineItem{
			ScheduleID:            item.ScheduleID,
			OvertimeEntryID:       item.OvertimeEntryID,
			SourceType:            item.SourceType,
			Label:                 item.Label,
			ContractType:          item.ContractType,
			WorkDate:              segment.workDate,
			StartTime:             segment.start.Format("15:04"),
			EndTime:               segment.end.Format("15:04"),
			BreakMinutes:          0,
			IrregularHoursProfile: item.IrregularHoursProfile,
			AppliedRatePercent:    segment.rate,
			MinutesWorked:         segment.minutes,
			PaidMinutes:           roundCurrency(paidMinutes),
			BaseAmount:            baseAmount,
			PremiumAmount:         premiumAmount,
		})
	}

	return items, totalMinutes, baseTotal, premiumTotal, nil
}

func parseWorkItemBounds(
	workDate time.Time,
	startTime, endTime string,
) (time.Time, time.Time, error) {
	baseDate := time.Date(
		workDate.UTC().Year(),
		workDate.UTC().Month(),
		workDate.UTC().Day(),
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

func roundRatio(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func appliedPayrollRateForMinute(
	item domain.PayrollWorkItem,
	minute time.Time,
	holidaySet map[string]struct{},
) float64 {
	if !isPayrollORTEligibleContractType(item.ContractType) {
		return 0
	}
	return ortRateForMinute(item.IrregularHoursProfile, minute, holidaySet)
}

func isPayrollEligibleContractType(contractType string) bool {
	switch strings.ToLower(strings.TrimSpace(contractType)) {
	case "permanent", "temporary", "on_call":
		return true
	default:
		return false
	}
}

func isPayrollORTEligibleContractType(contractType string) bool {
	return isLoondienstContractType(contractType)
}

func isLoondienstContractType(contractType string) bool {
	switch strings.ToLower(strings.TrimSpace(contractType)) {
	case "permanent", "temporary":
		return true
	default:
		return false
	}
}

func payoutHoursToMinutes(hours int32) int32 {
	return hours * 60
}

type payrollMonthLiveSummary struct {
	WorkedMinutes        int32
	PaidMinutes          float64
	BaseGrossAmount      float64
	IrregularGrossAmount float64
	GrossAmount          float64
	MultiplierSummaries  []domain.PayrollMultiplierSummary
}

type fixedPayrollAdjustmentSummary struct {
	ActualORTAmount         float64
	ForecastORTAmount       float64
	ApprovedOvertimeAmount  float64
	LeavePayoutAmount       float64
	ScheduledActualMinutes  int32
	ScheduledFutureMinutes  int32
	ApprovedOvertimeMinutes int32
	LeavePayoutMinutes      int32
	ORTBreakdown            []domain.FixedPayrollORTBreakdown
	OvertimeBreakdown       []domain.FixedPayrollOvertimeBreakdown
	LeavePayoutBreakdown    []domain.FixedPayrollLeavePayoutBreakdown
}

type onCallPayrollSummary struct {
	WorkedMinutes           int32
	WorkedHoursAmount       float64
	ApprovedOvertimeMinutes int32
	ApprovedOvertimeAmount  float64
	LeavePayoutMinutes      int32
	LeavePayoutAmount       float64
	WorkedHoursBreakdown    []domain.OnCallPayrollWorkedHoursBreakdown
	OvertimeBreakdown       []domain.OnCallPayrollOvertimeBreakdown
	LeavePayoutBreakdown    []domain.OnCallPayrollLeavePayoutBreakdown
}

func buildFixedPayrollContractSegmentsByEmployee(
	sources []domain.FixedPayrollContractSegmentSource,
	monthStart, monthEnd time.Time,
) map[uuid.UUID][]domain.FixedPayrollContractSegment {
	segmentsByEmployee := make(map[uuid.UUID][]domain.FixedPayrollContractSegment)
	daysInMonth := float64(monthEnd.Day())

	for _, source := range sources {
		activeDays := inclusiveDateDays(source.ActiveFrom, source.ActiveUntil)
		if activeDays <= 0 || daysInMonth <= 0 || source.FullTimeHoursPerWeek <= 0 {
			continue
		}

		prorationRatio := float64(activeDays) / daysInMonth
		baseAmount := roundCurrency(
			source.MonthlySalary *
				(source.HoursPerWeek / source.FullTimeHoursPerWeek) *
				prorationRatio,
		)

		segment := domain.FixedPayrollContractSegment{
			ContractID:           source.ContractID,
			ContractType:         source.ContractType,
			ActiveFrom:           source.ActiveFrom,
			ActiveUntil:          source.ActiveUntil,
			HoursPerWeek:         source.HoursPerWeek,
			FullTimeHoursPerWeek: source.FullTimeHoursPerWeek,
			MonthlySalary:        source.MonthlySalary,
			HourlyRate:           source.HourlyRate,
			ProrationRatio:       roundRatio(prorationRatio),
			BaseAmount:           baseAmount,
		}
		segmentsByEmployee[source.EmployeeID] = append(
			segmentsByEmployee[source.EmployeeID],
			segment,
		)
	}

	return segmentsByEmployee
}

func buildFixedPayrollPeriodContractSegmentsByEmployee(
	sources []domain.FixedPayrollContractSegmentSource,
	periodStart, periodEnd time.Time,
) map[uuid.UUID][]domain.FixedPayrollContractSegment {
	segmentsByEmployee := make(map[uuid.UUID][]domain.FixedPayrollContractSegment)
	periodDays := float64(inclusiveDateDays(periodStart, periodEnd))

	for _, source := range sources {
		activeDays := inclusiveDateDays(source.ActiveFrom, source.ActiveUntil)
		if activeDays <= 0 || periodDays <= 0 || source.HourlyRate <= 0 ||
			source.HoursPerWeek <= 0 {
			continue
		}

		prorationRatio := float64(activeDays) / periodDays
		periodWeeks := periodDays / 7
		baseAmount := roundCurrency(
			source.HourlyRate * source.HoursPerWeek * periodWeeks * prorationRatio,
		)

		segment := domain.FixedPayrollContractSegment{
			ContractID:           source.ContractID,
			ContractType:         source.ContractType,
			ActiveFrom:           source.ActiveFrom,
			ActiveUntil:          source.ActiveUntil,
			HoursPerWeek:         source.HoursPerWeek,
			FullTimeHoursPerWeek: source.FullTimeHoursPerWeek,
			MonthlySalary:        source.MonthlySalary,
			HourlyRate:           source.HourlyRate,
			ProrationRatio:       roundRatio(prorationRatio),
			BaseAmount:           baseAmount,
		}
		segmentsByEmployee[source.EmployeeID] = append(
			segmentsByEmployee[source.EmployeeID],
			segment,
		)
	}

	return segmentsByEmployee
}

func buildFixedPayrollAdjustments(
	workItems []domain.PayrollWorkItem,
	holidaySet map[string]struct{},
	asOf time.Time,
) (map[uuid.UUID]fixedPayrollAdjustmentSummary, error) {
	summaries := make(map[uuid.UUID]fixedPayrollAdjustmentSummary)

	for _, item := range workItems {
		if !isPayrollEligibleContractType(item.ContractType) ||
			!isLoondienstContractType(item.ContractType) {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if item.ContractRate == nil || *item.ContractRate <= 0 {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(item.IrregularHoursProfile) {
			return nil, domain.ErrSalaryInvalidRequest
		}

		summary := summaries[item.EmployeeID]
		switch item.SourceType {
		case domain.PayrollSourceSchedule:
			lineItems, _, _, _, err := buildPayrollPreviewLineItems(
				item,
				*item.ContractRate,
				holidaySet,
			)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
			for _, line := range lineItems {
				breakdown := domain.FixedPayrollORTBreakdown{
					ScheduleID:    line.ScheduleID,
					WorkDate:      line.WorkDate,
					StartTime:     line.StartTime,
					EndTime:       line.EndTime,
					RatePercent:   line.AppliedRatePercent,
					Minutes:       line.MinutesWorked,
					BasisAmount:   line.BaseAmount,
					PremiumAmount: line.PremiumAmount,
				}
				if isFixedPayrollLineActual(line, asOf) {
					breakdown.Status = "actual"
					summary.ActualORTAmount = roundCurrency(
						summary.ActualORTAmount + line.PremiumAmount,
					)
					summary.ScheduledActualMinutes += line.MinutesWorked
				} else {
					breakdown.Status = "forecast"
					summary.ForecastORTAmount = roundCurrency(summary.ForecastORTAmount + line.PremiumAmount)
					summary.ScheduledFutureMinutes += line.MinutesWorked
				}
				summary.ORTBreakdown = append(summary.ORTBreakdown, breakdown)
			}

		case domain.PayrollSourceOvertime:
			lineItems, err := buildFixedPayrollOvertimeLineItems(
				item,
				*item.ContractRate,
				holidaySet,
			)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
			for _, line := range lineItems {
				amount := roundCurrency(line.BaseAmount + line.PremiumAmount)
				summary.ApprovedOvertimeAmount = roundCurrency(
					summary.ApprovedOvertimeAmount + amount,
				)
				summary.ApprovedOvertimeMinutes += line.MinutesWorked
				summary.OvertimeBreakdown = append(
					summary.OvertimeBreakdown,
					domain.FixedPayrollOvertimeBreakdown{
						OvertimeEntryID: line.OvertimeEntryID,
						WorkDate:        line.WorkDate,
						Minutes:         line.MinutesWorked,
						Amount:          amount,
						Status:          "approved",
					},
				)
			}

		case domain.PayrollSourceLeavePayout:
			lineItems, _, baseAmount, _, err := buildLeavePayoutLineItems(item)
			if err != nil || len(lineItems) == 0 {
				return nil, domain.ErrSalaryInvalidRequest
			}
			minutes := int32(math.Round(item.MinutesWorked))
			summary.LeavePayoutAmount = roundCurrency(summary.LeavePayoutAmount + baseAmount)
			summary.LeavePayoutMinutes += minutes
			summary.LeavePayoutBreakdown = append(
				summary.LeavePayoutBreakdown,
				domain.FixedPayrollLeavePayoutBreakdown{
					LeavePayoutRequestID: item.LeavePayoutRequestID,
					SalaryMonth:          item.WorkDate,
					RequestedHours:       minutes / 60,
					Minutes:              minutes,
					Amount:               baseAmount,
					Status:               "approved",
				},
			)
		}

		summaries[item.EmployeeID] = summary
	}

	return summaries, nil
}

func buildFixedPayrollOvertimeLineItems(
	item domain.PayrollWorkItem,
	hourlyRate float64,
	holidaySet map[string]struct{},
) ([]domain.PayrollPreviewLineItem, error) {
	if item.StartTime == "" {
		lineItems, _, _, _ := buildSimpleOvertimeLineItems(item, hourlyRate)
		return lineItems, nil
	}
	lineItems, _, _, _, err := buildPayrollPreviewLineItems(item, hourlyRate, holidaySet)
	return lineItems, err
}

func applyFixedPayrollAdjustmentSummary(
	row *domain.FixedPayrollMonthSummaryRow,
	summary fixedPayrollAdjustmentSummary,
) {
	row.ActualORTAmount = summary.ActualORTAmount
	row.ForecastORTAmount = summary.ForecastORTAmount
	row.ApprovedOvertimeAmount = summary.ApprovedOvertimeAmount
	row.LeavePayoutAmount = summary.LeavePayoutAmount
	row.ScheduledActualMinutes = summary.ScheduledActualMinutes
	row.ScheduledFutureMinutes = summary.ScheduledFutureMinutes
	row.ApprovedOvertimeMinutes = summary.ApprovedOvertimeMinutes
	row.LeavePayoutMinutes = summary.LeavePayoutMinutes
	row.ORTBreakdown = summary.ORTBreakdown
	row.OvertimeBreakdown = summary.OvertimeBreakdown
	row.LeavePayoutBreakdown = summary.LeavePayoutBreakdown
}

func buildOnCallPayrollSummaries(
	workItems []domain.PayrollWorkItem,
	holidaySet map[string]struct{},
) (map[uuid.UUID]onCallPayrollSummary, error) {
	summaries := make(map[uuid.UUID]onCallPayrollSummary)

	for _, item := range workItems {
		if !isPayrollEligibleContractType(item.ContractType) || item.ContractType != "on_call" {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if item.ContractRate == nil || *item.ContractRate <= 0 {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(item.IrregularHoursProfile) {
			return nil, domain.ErrSalaryInvalidRequest
		}

		summary := summaries[item.EmployeeID]
		switch item.SourceType {
		case domain.PayrollSourceSchedule:
			lineItems, _, _, _, err := buildPayrollPreviewLineItems(
				item,
				*item.ContractRate,
				holidaySet,
			)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
			for _, line := range lineItems {
				amount := roundCurrency(line.BaseAmount + line.PremiumAmount)
				summary.WorkedMinutes += line.MinutesWorked
				summary.WorkedHoursAmount = roundCurrency(summary.WorkedHoursAmount + amount)
				summary.WorkedHoursBreakdown = append(
					summary.WorkedHoursBreakdown,
					domain.OnCallPayrollWorkedHoursBreakdown{
						ScheduleID:  line.ScheduleID,
						WorkDate:    line.WorkDate,
						StartTime:   line.StartTime,
						EndTime:     line.EndTime,
						Minutes:     line.MinutesWorked,
						HourlyRate:  *item.ContractRate,
						BaseAmount:  line.BaseAmount,
						TotalAmount: amount,
					},
				)
			}

		case domain.PayrollSourceOvertime:
			lineItems, _, _, _ := buildSimpleOvertimeLineItems(item, *item.ContractRate)
			for _, line := range lineItems {
				amount := roundCurrency(line.BaseAmount + line.PremiumAmount)
				summary.ApprovedOvertimeMinutes += line.MinutesWorked
				summary.ApprovedOvertimeAmount = roundCurrency(
					summary.ApprovedOvertimeAmount + amount,
				)
				summary.OvertimeBreakdown = append(
					summary.OvertimeBreakdown,
					domain.OnCallPayrollOvertimeBreakdown{
						OvertimeEntryID: line.OvertimeEntryID,
						WorkDate:        line.WorkDate,
						Minutes:         line.MinutesWorked,
						HourlyRate:      *item.ContractRate,
						Amount:          amount,
						Status:          "approved",
					},
				)
			}

		case domain.PayrollSourceLeavePayout:
			_, _, baseAmount, _, err := buildLeavePayoutLineItems(item)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
			minutes := int32(math.Round(item.MinutesWorked))
			summary.LeavePayoutAmount = roundCurrency(summary.LeavePayoutAmount + baseAmount)
			summary.LeavePayoutMinutes += minutes
			summary.LeavePayoutBreakdown = append(
				summary.LeavePayoutBreakdown,
				domain.OnCallPayrollLeavePayoutBreakdown{
					LeavePayoutRequestID: item.LeavePayoutRequestID,
					SalaryMonth:          item.WorkDate,
					RequestedHours:       minutes / 60,
					Minutes:              minutes,
					Amount:               baseAmount,
					Status:               "approved",
				},
			)
		}

		summaries[item.EmployeeID] = summary
	}

	return summaries, nil
}

func applyOnCallPayrollSummary(
	row *domain.OnCallPayrollMonthSummaryRow,
	summary onCallPayrollSummary,
) {
	row.WorkedMinutes = summary.WorkedMinutes
	row.WorkedHoursAmount = summary.WorkedHoursAmount
	row.ApprovedOvertimeMinutes = summary.ApprovedOvertimeMinutes
	row.ApprovedOvertimeAmount = summary.ApprovedOvertimeAmount
	row.LeavePayoutMinutes = summary.LeavePayoutMinutes
	row.LeavePayoutAmount = summary.LeavePayoutAmount
	row.WorkedHoursBreakdown = summary.WorkedHoursBreakdown
	row.OvertimeBreakdown = summary.OvertimeBreakdown
	row.LeavePayoutBreakdown = summary.LeavePayoutBreakdown
}

func shouldIncludeOnCallPayrollRow(row domain.OnCallPayrollMonthSummaryRow) bool {
	return row.WorkedMinutes > 0 ||
		row.WorkedHoursAmount > 0 ||
		row.ApprovedOvertimeMinutes > 0 ||
		row.ApprovedOvertimeAmount > 0 ||
		row.LeavePayoutMinutes > 0 ||
		row.LeavePayoutAmount > 0 ||
		row.PayableGrossAmount > 0 ||
		row.PendingEntryCount > 0 ||
		row.PendingWorkedMinutes > 0
}

func contractSegmentPaidMinutes(
	segment domain.FixedPayrollContractSegment,
	monthEnd time.Time,
) float64 {
	activeDays := inclusiveDateDays(segment.ActiveFrom, segment.ActiveUntil)
	monthDays := monthEnd.Day()
	if activeDays <= 0 || monthDays <= 0 {
		return 0
	}
	return roundCurrency(
		segment.HoursPerWeek * 52 / 12 * float64(activeDays) / float64(monthDays) * 60,
	)
}

func contractSegmentPeriodPaidMinutes(
	segment domain.FixedPayrollContractSegment,
	periodStart, periodEnd time.Time,
) float64 {
	activeDays := inclusiveDateDays(segment.ActiveFrom, segment.ActiveUntil)
	periodDays := inclusiveDateDays(periodStart, periodEnd)
	if activeDays <= 0 || periodDays <= 0 {
		return 0
	}
	periodWeeks := float64(periodDays) / 7
	return roundCurrency(
		segment.HoursPerWeek * periodWeeks * float64(activeDays) / float64(periodDays) * 60,
	)
}

func inclusiveDateDays(start, end time.Time) int {
	startDate := time.Date(
		start.UTC().Year(),
		start.UTC().Month(),
		start.UTC().Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endDate := time.Date(end.UTC().Year(), end.UTC().Month(), end.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if endDate.Before(startDate) {
		return 0
	}
	return int(endDate.Sub(startDate).Hours()/24) + 1
}

func fixedPayrollAsOf(monthStart, monthEnd time.Time, isCurrentMonth bool) time.Time {
	if isCurrentMonth {
		return time.Now().UTC()
	}
	currentMonth := time.Date(
		time.Now().UTC().Year(),
		time.Now().UTC().Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if monthStart.Before(currentMonth) {
		return monthEnd.AddDate(0, 0, 1)
	}
	return monthStart
}

func isFixedPayrollLineActual(line domain.PayrollPreviewLineItem, asOf time.Time) bool {
	_, end, err := parseWorkItemBounds(line.WorkDate, line.StartTime, line.EndTime)
	if err != nil {
		return !line.WorkDate.After(asOf)
	}
	return !end.After(asOf)
}

func buildPayrollMonthLiveSummaries(
	workItems []domain.PayrollWorkItem,
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
	for _, item := range workItems {
		if !isPayrollEligibleContractType(item.ContractType) {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if item.ContractRate == nil || *item.ContractRate <= 0 {
			return nil, domain.ErrSalaryInvalidRequest
		}
		if !isValidPayrollIrregularHoursProfile(item.IrregularHoursProfile) {
			return nil, domain.ErrSalaryInvalidRequest
		}

		var lineItems []domain.PayrollPreviewLineItem
		var workedMinutes int32
		var baseAmount float64
		var premiumAmount float64
		var err error

		if item.SourceType == domain.PayrollSourceLeavePayout {
			lineItems, workedMinutes, baseAmount, premiumAmount, err = buildLeavePayoutLineItems(
				item,
			)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
		} else if item.SourceType == domain.PayrollSourceOvertime && item.StartTime == "" {
			lineItems, workedMinutes, baseAmount, premiumAmount = buildSimpleOvertimeLineItems(item, *item.ContractRate)
		} else {
			lineItems, workedMinutes, baseAmount, premiumAmount, err = buildPayrollPreviewLineItems(item, *item.ContractRate, holidaySet)
			if err != nil {
				return nil, domain.ErrSalaryInvalidRequest
			}
		}

		acc := accumulators[item.EmployeeID]
		if acc == nil {
			acc = &liveAccumulator{
				MultiplierByRate: make(map[float64]*domain.PayrollMultiplierSummary),
			}
			accumulators[item.EmployeeID] = acc
		}

		acc.WorkedMinutes += workedMinutes
		acc.BaseGrossAmount = roundCurrency(acc.BaseGrossAmount + baseAmount)
		acc.IrregularGrossAmount = roundCurrency(acc.IrregularGrossAmount + premiumAmount)
		for _, line := range lineItems {
			acc.PaidMinutes = roundCurrency(acc.PaidMinutes + line.PaidMinutes)
			bucket := acc.MultiplierByRate[line.AppliedRatePercent]
			if bucket == nil {
				bucket = &domain.PayrollMultiplierSummary{RatePercent: line.AppliedRatePercent}
				acc.MultiplierByRate[line.AppliedRatePercent] = bucket
			}
			bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + line.PaidMinutes)
			bucket.PaidMinutes = roundCurrency(bucket.PaidMinutes + line.PaidMinutes)
			bucket.BaseAmount = roundCurrency(bucket.BaseAmount + line.BaseAmount)
			bucket.PremiumAmount = roundCurrency(bucket.PremiumAmount + line.PremiumAmount)
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

func buildLiveShiftCountMap(workItems []domain.PayrollWorkItem) map[uuid.UUID]int32 {
	counts := make(map[uuid.UUID]int32)
	for _, item := range workItems {
		if item.SourceType == domain.PayrollSourceSchedule {
			counts[item.EmployeeID]++
		}
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

func buildLockedPayrollSnapshot(lineItems []domain.PayPeriodLineItem) lockedPayrollSnapshot {
	multiplierBuckets := make(map[float64]*domain.PayrollMultiplierSummary)
	var workedMinutes float64
	var paidMinutes float64
	var baseGrossAmount float64
	var irregularGrossAmount float64

	for _, item := range lineItems {
		itemWorkedMinutes := item.MinutesWorked
		if item.LineType == domain.PayrollSourceLeavePayout {
			itemWorkedMinutes = 0
		}
		workedMinutes = roundCurrency(workedMinutes + itemWorkedMinutes)
		paidMinutes = roundCurrency(paidMinutes + item.MinutesWorked)
		baseGrossAmount = roundCurrency(baseGrossAmount + item.BaseAmount)
		irregularGrossAmount = roundCurrency(irregularGrossAmount + item.PremiumAmount)

		bucket := multiplierBuckets[item.AppliedRatePercent]
		if bucket == nil {
			bucket = &domain.PayrollMultiplierSummary{RatePercent: item.AppliedRatePercent}
			multiplierBuckets[item.AppliedRatePercent] = bucket
		}
		bucket.WorkedMinutes = roundCurrency(bucket.WorkedMinutes + itemWorkedMinutes)
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
		ShiftCount:           int32(len(lineItems)),
		MultiplierSummaries:  sortedMultiplierSummaries(multiplierBuckets),
	}
}

func filterPayrollWorkItemsByContractType(
	workItems []domain.PayrollWorkItem,
	contractType *string,
) []domain.PayrollWorkItem {
	if contractType == nil {
		return workItems
	}

	filtered := make([]domain.PayrollWorkItem, 0, len(workItems))
	for _, item := range workItems {
		if matchesPayrollContractType(item.ContractType, *contractType) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterPayrollWorkItemsByPayrollGroup(
	workItems []domain.PayrollWorkItem,
	payrollGroup string,
) []domain.PayrollWorkItem {
	contractType := payrollGroupContractType(payrollGroup)
	return filterPayrollWorkItemsByContractType(workItems, &contractType)
}

func payrollGroupContractType(payrollGroup string) string {
	if payrollGroup == domain.PayrollGroupOnCall {
		return domain.PayrollGroupOnCall
	}
	return "loondienst"
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
	a := strings.ToLower(strings.TrimSpace(actual))
	e := strings.ToLower(strings.TrimSpace(expected))
	switch e {
	case "loondienst":
		return a == "permanent" || a == "temporary"
	case "zzp":
		return a == "on_call"
	default:
		return a == e
	}
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

func buildPayPeriodLineItem(
	item domain.PayrollPreviewLineItem,
	workItems []domain.PayrollWorkItem,
) domain.PayPeriodLineItem {
	metadata := map[string]any{
		"source_type":  item.SourceType,
		"start_time":   item.StartTime,
		"end_time":     item.EndTime,
		"paid_minutes": roundCurrency(item.PaidMinutes),
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		payload = []byte(`{}`)
	}

	return domain.PayPeriodLineItem{
		ScheduleID:            item.ScheduleID,
		OvertimeEntryID:       item.OvertimeEntryID,
		LeavePayoutRequestID:  item.LeavePayoutRequestID,
		ContractType:          item.ContractType,
		WorkDate:              item.WorkDate,
		LineType:              item.SourceType,
		IrregularHoursProfile: item.IrregularHoursProfile,
		AppliedRatePercent:    item.AppliedRatePercent,
		MinutesWorked:         roundCurrency(item.PaidMinutes),
		BaseAmount:            item.BaseAmount,
		PremiumAmount:         item.PremiumAmount,
		Metadata:              payload,
	}
}

func uniquePreviewOvertimeEntryIDs(items []domain.PayrollPreviewLineItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(items))
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.OvertimeEntryID == nil {
			continue
		}
		if _, ok := seen[*item.OvertimeEntryID]; ok {
			continue
		}
		seen[*item.OvertimeEntryID] = struct{}{}
		result = append(result, *item.OvertimeEntryID)
	}
	return result
}

func uniquePreviewScheduleIDs(items []domain.PayrollPreviewLineItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0)
	for _, item := range items {
		if item.ScheduleID == nil || *item.ScheduleID == uuid.Nil {
			continue
		}
		if _, ok := seen[*item.ScheduleID]; ok {
			continue
		}
		seen[*item.ScheduleID] = struct{}{}
		ids = append(ids, *item.ScheduleID)
	}
	return ids
}

func uniquePreviewLeavePayoutRequestIDs(items []domain.PayrollPreviewLineItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(items))
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.LeavePayoutRequestID == nil {
			continue
		}
		if _, ok := seen[*item.LeavePayoutRequestID]; ok {
			continue
		}
		seen[*item.LeavePayoutRequestID] = struct{}{}
		result = append(result, *item.LeavePayoutRequestID)
	}
	return result
}
