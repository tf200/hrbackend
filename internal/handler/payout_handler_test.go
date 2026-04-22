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

func TestPayoutHandlerGetPayrollMonthORTOverviewSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeID := uuid.New()
	payPeriodID := uuid.New()
	service := &fakePayoutService{
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
	handler := NewPayoutHandler(service)
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

func TestPayoutHandlerGetORTRulesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	roster := domain.IrregularHoursProfileRoster
	service := &fakePayoutService{
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
	handler := NewPayoutHandler(service)
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

type fakePayoutService struct {
	ortOverviewPage   *domain.PayrollMonthORTOverviewPage
	ortOverviewParams domain.PayrollMonthORTOverviewParams
	ortOverviewErr    error
	ortRules          *domain.ORTRulesResponse
	ortRulesErr       error
}

func (f *fakePayoutService) CreatePayoutRequest(
	_ context.Context,
	_ uuid.UUID,
	_ domain.CreatePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) DecidePayoutRequestByAdmin(
	_ context.Context,
	_, _ uuid.UUID,
	_ domain.DecidePayoutRequestParams,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) MarkPayoutRequestPaidByAdmin(
	_ context.Context,
	_, _ uuid.UUID,
) (*domain.PayoutRequest, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) ListMyPayoutRequests(
	_ context.Context,
	_ domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) ListPayoutRequests(
	_ context.Context,
	_ domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) PreviewPayroll(
	_ context.Context,
	_ domain.PayrollPreviewParams,
) (*domain.PayrollPreview, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) PreviewMyPayroll(
	_ context.Context,
	_ uuid.UUID,
	_, _ time.Time,
) (*domain.PayrollPreview, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) ClosePayPeriod(
	_ context.Context,
	_ uuid.UUID,
	_ domain.ClosePayPeriodParams,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) GetPayPeriodByID(
	_ context.Context,
	_ uuid.UUID,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) ListPayPeriods(
	_ context.Context,
	_ domain.ListPayPeriodsParams,
) (*domain.PayPeriodPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) MarkPayPeriodPaidByAdmin(
	_ context.Context,
	_, _ uuid.UUID,
) (*domain.PayPeriod, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) GetPayrollMonthSummary(
	_ context.Context,
	_ domain.PayrollMonthSummaryParams,
) (*domain.PayrollMonthSummaryPage, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) GetPayrollMonthORTOverview(
	_ context.Context,
	params domain.PayrollMonthORTOverviewParams,
) (*domain.PayrollMonthORTOverviewPage, error) {
	f.ortOverviewParams = params
	if f.ortOverviewErr != nil {
		return nil, f.ortOverviewErr
	}
	return f.ortOverviewPage, nil
}

func (f *fakePayoutService) GetORTRules(_ context.Context) (*domain.ORTRulesResponse, error) {
	if f.ortRulesErr != nil {
		return nil, f.ortRulesErr
	}
	return f.ortRules, nil
}

func (f *fakePayoutService) GetPayrollMonthDetail(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ *string,
) (*domain.PayrollMonthDetail, error) {
	panic("unexpected call")
}

func (f *fakePayoutService) ExportPayrollMonthPDF(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ *string,
) ([]byte, string, error) {
	panic("unexpected call")
}

var _ domain.PayoutService = (*fakePayoutService)(nil)

func stringPtr(v string) *string {
	return &v
}
