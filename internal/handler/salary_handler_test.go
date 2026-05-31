package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestSalaryHandlerGetPayrollMonthORTOverviewSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeID := uuid.New()
	payPeriodID := uuid.New()
	service := &fakeSalaryService{
		ortOverviewPage: &domain.PayrollMonthORTOverviewPage{
			Month: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			Distribution: []domain.PayrollMultiplierSummary{
				{
					RatePercent:   45,
					WorkedMinutes: 120,
					PaidMinutes:   120,
					BaseAmount:    20,
					PremiumAmount: 9,
				},
			},
			Items: []domain.PayrollMonthORTOverviewRow{
				{
					EmployeeID:        employeeID,
					EmployeeName:      "Annie Case",
					Month:             time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
					IsCurrentMonth:    true,
					IsLocked:          false,
					HasLockedSnapshot: true,
					DataSource:        "live",
					WorkedMinutes:     120,
					PaidMinutes:       120,
					BaseAmount:        20,
					PremiumAmount:     9,
					PayPeriodID:       &payPeriodID,
					Distribution: []domain.PayrollMultiplierSummary{
						{
							RatePercent:   45,
							WorkedMinutes: 120,
							PaidMinutes:   120,
							BaseAmount:    20,
							PremiumAmount: 9,
						},
					},
				},
			},
			TotalCount: 1,
		},
	}

	router := gin.New()
	handler := NewSalaryHandler(service)
	router.GET("/payroll-month-summary/ort-overview", handler.GetPayrollMonthORTOverview)

	req := httptest.NewRequest(
		http.MethodGet,
		"/payroll-month-summary/ort-overview?month=2026-04&page=1&page_size=5&employee_search=annie",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.ortOverviewParams.Month.Format("2006-01") != "2026-04" {
		t.Fatalf("unexpected parsed month: %s", service.ortOverviewParams.Month.Format("2006-01"))
	}
	if service.ortOverviewParams.Limit != 5 || service.ortOverviewParams.Offset != 0 {
		t.Fatalf("unexpected pagination params: %#v", service.ortOverviewParams)
	}
	if service.ortOverviewParams.EmployeeSearch == nil || *service.ortOverviewParams.EmployeeSearch != "annie" {
		t.Fatalf("unexpected employee_search: %#v", service.ortOverviewParams.EmployeeSearch)
	}

	var response struct {
		Success bool                            `json:"success"`
		Message string                          `json:"message"`
		Data    payrollMonthORTOverviewResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "Payroll month ORT overview retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if response.Data.Month != "2026-04" {
		t.Fatalf("unexpected response month: %s", response.Data.Month)
	}
	if response.Data.Count != 1 || response.Data.PageSize != 5 || len(response.Data.Results) != 1 {
		t.Fatalf("unexpected page response: %#v", response.Data)
	}
	if len(response.Data.Distribution) != 1 || response.Data.Distribution[0].RatePercent != 45 {
		t.Fatalf("unexpected distribution: %#v", response.Data.Distribution)
	}
	if response.Data.Results[0].EmployeeID != employeeID {
		t.Fatalf("unexpected employee id: %s", response.Data.Results[0].EmployeeID)
	}
}

func TestSalaryHandlerGetORTRulesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	roster := domain.IrregularHoursProfileRoster
	service := &fakeSalaryService{
		ortRules: &domain.ORTRulesResponse{
			Rules: []domain.ORTRule{
				{
					Order:                 1,
					RatePercent:           25,
					Label:                 "Roster evening",
					Description:           "Roster profile from 19:00 to before 22:00 applies 25% ORT.",
					ContractType:          "loondienst",
					IrregularHoursProfile: &roster,
					DayType:               "any",
					TimeFrom:              stringPtr("19:00"),
					TimeTo:                stringPtr("22:00"),
				},
			},
		},
	}

	router := gin.New()
	handler := NewSalaryHandler(service)
	router.GET("/payroll/ort-rules", handler.GetORTRules)

	req := httptest.NewRequest(http.MethodGet, "/payroll/ort-rules", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Success bool             `json:"success"`
		Message string           `json:"message"`
		Data    ortRulesResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "ORT rules retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if len(response.Data.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(response.Data.Rules))
	}
	if response.Data.Rules[0].RatePercent != 25 || response.Data.Rules[0].IrregularHoursProfile == nil {
		t.Fatalf("unexpected rule payload: %#v", response.Data.Rules[0])
	}
}

func TestSalaryHandlerGetFixedPayrollMonthStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeSalaryService{
		fixedStats: &domain.PayrollMonthStats{
			Month:                       time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			TotalBaseContractPay:        4200,
			TotalORTPay:                 200,
			TotalOvertimePay:            125.5,
			TotalRequestedLeaveHoursPay: 50,
			TotalRequestedLeaveHours:    2,
			TotalGrossPayable:           4575.5,
		},
	}

	router := gin.New()
	handler := NewSalaryHandler(service)
	router.GET("/payroll-month-summary/fixed/stats", handler.GetFixedPayrollMonthStats)

	req := httptest.NewRequest(
		http.MethodGet,
		"/payroll-month-summary/fixed/stats?month=2026-04&employee_search=ann",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.fixedStatsParams.Month.Format("2006-01") != "2026-04" {
		t.Fatalf("unexpected parsed month: %s", service.fixedStatsParams.Month.Format("2006-01"))
	}
	if service.fixedStatsParams.EmployeeSearch == nil || *service.fixedStatsParams.EmployeeSearch != "ann" {
		t.Fatalf("unexpected employee_search: %#v", service.fixedStatsParams.EmployeeSearch)
	}

	var response struct {
		Success bool                      `json:"success"`
		Message string                    `json:"message"`
		Data    payrollMonthStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "Fixed payroll month stats retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if response.Data.Month != "2026-04" || response.Data.TotalORTPay != 200 || response.Data.TotalGrossPayable != 4575.5 {
		t.Fatalf("unexpected stats response: %#v", response.Data)
	}
}

func TestSalaryHandlerGetOnCallPayrollMonthStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeSalaryService{
		onCallStats: &domain.PayrollMonthStats{
			Month:                       time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			TotalBaseContractPay:        850,
			TotalOvertimePay:            100,
			TotalRequestedLeaveHoursPay: 25,
			TotalRequestedLeaveHours:    1,
			TotalGrossPayable:           975,
		},
	}

	router := gin.New()
	handler := NewSalaryHandler(service)
	router.GET("/payroll-month-summary/on-call/stats", handler.GetOnCallPayrollMonthStats)

	req := httptest.NewRequest(http.MethodGet, "/payroll-month-summary/on-call/stats?month=2026-04", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.onCallStatsParams.Month.Format("2006-01") != "2026-04" {
		t.Fatalf("unexpected parsed month: %s", service.onCallStatsParams.Month.Format("2006-01"))
	}

	var response struct {
		Success bool                      `json:"success"`
		Message string                    `json:"message"`
		Data    payrollMonthStatsResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "On-call payroll month stats retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if response.Data.Month != "2026-04" || response.Data.TotalBaseContractPay != 850 {
		t.Fatalf("unexpected stats response: %#v", response.Data)
	}
}

func TestSalaryPageResponseIncludesLiveLineItemLabelAndBreakMinutes(t *testing.T) {
	employeeID := uuid.New()
	month := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	rate := 24.5
	contractHours := 32.0

	response := toSalaryPageResponse(&domain.SalaryPageData{
		EmployeeID:            employeeID,
		EmployeeName:          "Jane Doe",
		Month:                 month,
		ContractType:          "loondienst",
		ContractRate:          &rate,
		ContractHours:         &contractHours,
		IrregularHoursProfile: "roster",
		DataSource:            "live",
		Preview: &domain.PayrollPreview{
			EmployeeID:           employeeID,
			EmployeeName:         "Jane Doe",
			PeriodStart:          month,
			PeriodEnd:            month.AddDate(0, 1, -1),
			TotalWorkedMinutes:   480,
			BaseGrossAmount:      183.75,
			IrregularGrossAmount: 45.94,
			GrossAmount:          229.69,
			LineItems: []domain.PayrollPreviewLineItem{
				{
					SourceType:         "schedule",
					Label:              "Evening care route",
					WorkDate:           time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
					StartTime:          "15:00",
					EndTime:            "23:00",
					BreakMinutes:       0,
					AppliedRatePercent: 25,
					MinutesWorked:      480,
					PaidMinutes:        450,
					BaseAmount:         183.75,
					PremiumAmount:      45.94,
				},
			},
		},
	})

	if response == nil {
		t.Fatalf("expected response")
	}
	if len(response.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(response.LineItems))
	}
	line := response.LineItems[0]
	if line.Label != "Evening care route" {
		t.Fatalf("expected label to be preserved, got %q", line.Label)
	}
	if line.BreakMinutes != 0 {
		t.Fatalf("expected break minutes 0, got %d", line.BreakMinutes)
	}
	if line.GrossAmount != 229.69 {
		t.Fatalf("expected gross amount 229.69, got %.2f", line.GrossAmount)
	}
}

type fakeSalaryService struct {
	ortOverviewPage   *domain.PayrollMonthORTOverviewPage
	ortOverviewParams domain.PayrollMonthORTOverviewParams
	ortOverviewErr    error
	ortRules          *domain.ORTRulesResponse
	ortRulesErr       error
	fixedStats        *domain.PayrollMonthStats
	fixedStatsParams  domain.PayrollMonthSummaryParams
	fixedStatsErr     error
	onCallStats       *domain.PayrollMonthStats
	onCallStatsParams domain.PayrollMonthSummaryParams
	onCallStatsErr    error
	salaryPageData    *domain.SalaryPageData
	salaryPageErr     error
}

func (f *fakeSalaryService) ListSalaryScaleSteps(
	_ context.Context,
	_ domain.ListSalaryScaleStepsParams,
) (*domain.SalaryScaleStepsResult, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) PreviewPayroll(
	_ context.Context,
	_ domain.PayrollPreviewParams,
) (*domain.PayrollPreview, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) PreviewMyPayroll(
	_ context.Context,
	_ uuid.UUID,
	_, _ time.Time,
) (*domain.PayrollPreview, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) ClosePayPeriod(
	_ context.Context,
	_ uuid.UUID,
	_ domain.ClosePayPeriodParams,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) PreviewPayrollMonthClose(
	_ context.Context,
	_ domain.ClosePayrollMonthParams,
) (*domain.PayrollMonthCloseResult, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) ClosePayrollMonthByAdmin(
	_ context.Context,
	_ uuid.UUID,
	_ domain.ClosePayrollMonthParams,
) (*domain.PayrollMonthCloseResult, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) GetPayPeriodByID(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) ListPayPeriods(
	_ context.Context,
	_ domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) MarkPayPeriodPaidByAdmin(
	_ context.Context,
	_, _ uuid.UUID,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) GetFixedPayrollMonthSummary(
	_ context.Context,
	_ domain.PayrollMonthSummaryParams,
) (*domain.FixedPayrollMonthSummaryPage, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) GetOnCallPayrollMonthSummary(
	_ context.Context,
	_ domain.PayrollMonthSummaryParams,
) (*domain.OnCallPayrollMonthSummaryPage, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) GetFixedPayrollMonthStats(
	_ context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthStats, error) {
	f.fixedStatsParams = params
	if f.fixedStatsErr != nil {
		return nil, f.fixedStatsErr
	}
	return f.fixedStats, nil
}
func (f *fakeSalaryService) GetOnCallPayrollMonthStats(
	_ context.Context,
	params domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthStats, error) {
	f.onCallStatsParams = params
	if f.onCallStatsErr != nil {
		return nil, f.onCallStatsErr
	}
	return f.onCallStats, nil
}
func (f *fakeSalaryService) GetPayrollMonthORTOverview(
	_ context.Context,
	params domain.PayrollMonthORTOverviewParams,
) (*domain.PayrollMonthORTOverviewPage, error) {
	f.ortOverviewParams = params
	if f.ortOverviewErr != nil {
		return nil, f.ortOverviewErr
	}
	return f.ortOverviewPage, nil
}
func (f *fakeSalaryService) GetORTRules(_ context.Context) (*domain.ORTRulesResponse, error) {
	if f.ortRulesErr != nil {
		return nil, f.ortRulesErr
	}
	return f.ortRules, nil
}
func (f *fakeSalaryService) GetPayrollMonthDetail(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ *string,
) (*domain.PayrollMonthDetail, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) ExportPayrollMonthPDF(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ *string,
) ([]byte, string, error) {
	panic("unexpected call")
}
func (f *fakeSalaryService) GetMySalaryPage(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
) (*domain.SalaryPageData, error) {
	if f.salaryPageErr != nil {
		return nil, f.salaryPageErr
	}
	return f.salaryPageData, nil
}

var _ domain.SalaryService = (*fakeSalaryService)(nil)

func stringPtr(v string) *string {
	return &v
}
