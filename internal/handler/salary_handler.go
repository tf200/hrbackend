package handler

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SalaryHandler struct {
	service domain.SalaryService
}

func NewSalaryHandler(service domain.SalaryService) *SalaryHandler {
	return &SalaryHandler{service: service}
}

func (h *SalaryHandler) ListSalaryScaleSteps(ctx *gin.Context) {
	var req listSalaryScaleStepsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.ListSalaryScaleSteps(ctx.Request.Context(), toListSalaryScaleStepsParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list salary scale steps", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toSalaryScaleStepsResponse(result), "Salary scale steps retrieved successfully"),
	)
}

func (h *SalaryHandler) PreviewPayroll(ctx *gin.Context) {
	var req previewPayrollRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPreviewPayrollParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	preview, err := h.service.PreviewPayroll(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toPayrollPreviewResponse(preview), "Payroll preview retrieved successfully"),
	)
}

func (h *SalaryHandler) PreviewMyPayroll(ctx *gin.Context) {
	var req previewMyPayrollRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	periodStart, periodEnd, err := toPreviewMyPayrollDates(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	preview, err := h.service.PreviewMyPayroll(
		ctx.Request.Context(),
		employeeID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toPayrollPreviewResponse(preview), "Payroll preview retrieved successfully"),
	)
}

func (h *SalaryHandler) GetMySalaryPage(ctx *gin.Context) {
	var req salaryPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid month format, expected YYYY-MM", ""))
		return
	}

	data, err := h.service.GetMySalaryPage(ctx.Request.Context(), employeeID, month)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toSalaryPageResponse(data), "Salary page data retrieved successfully"),
	)
}

func (h *SalaryHandler) ExportMyPayrollMonthPDF(ctx *gin.Context) {
	var req salaryPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	month, err := time.Parse(payoutMonthLayout, req.Month)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid month format, expected YYYY-MM", ""))
		return
	}

	pdfBytes, filename, err := h.service.ExportPayrollMonthPDF(
		ctx.Request.Context(),
		employeeID,
		month,
		nil,
	)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	ctx.DataFromReader(http.StatusOK, int64(len(pdfBytes)), "application/pdf", bytes.NewReader(pdfBytes), nil)
}

func (h *SalaryHandler) GetORTRules(ctx *gin.Context) {
	rules, err := h.service.GetORTRules(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toORTRulesResponse(rules), "ORT rules retrieved successfully"))
}

func (h *SalaryHandler) GetFixedPayrollMonthSummary(ctx *gin.Context) {
	var req payrollMonthSummaryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPayrollMonthSummaryParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.GetFixedPayrollMonthSummary(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toFixedPayrollMonthSummaryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Fixed payroll month summary retrieved successfully"))
}

func (h *SalaryHandler) GetOnCallPayrollMonthSummary(ctx *gin.Context) {
	var req payrollMonthSummaryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPayrollMonthSummaryParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.GetOnCallPayrollMonthSummary(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toOnCallPayrollMonthSummaryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "On-call payroll month summary retrieved successfully"))
}

func (h *SalaryHandler) GetFixedPayrollMonthStats(ctx *gin.Context) {
	var req payrollMonthStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPayrollMonthStatsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	stats, err := h.service.GetFixedPayrollMonthStats(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayrollMonthStatsResponse(stats), "Fixed payroll month stats retrieved successfully"))
}

func (h *SalaryHandler) GetOnCallPayrollMonthStats(ctx *gin.Context) {
	var req payrollMonthStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPayrollMonthStatsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	stats, err := h.service.GetOnCallPayrollMonthStats(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayrollMonthStatsResponse(stats), "On-call payroll month stats retrieved successfully"))
}

func (h *SalaryHandler) GetPayrollMonthORTOverview(ctx *gin.Context) {
	var req payrollMonthSummaryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toPayrollMonthORTOverviewParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.GetPayrollMonthORTOverview(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	paged := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toPayrollMonthORTOverviewEmployeeResponses(page.Items),
		page.TotalCount,
	)
	response := toPayrollMonthORTOverviewResponse(page, paged)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Payroll month ORT overview retrieved successfully"))
}

func (h *SalaryHandler) GetPayrollMonthDetail(ctx *gin.Context) {
	var req payrollMonthDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID, month, contractType, err := toPayrollMonthDetailRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	detail, err := h.service.GetPayrollMonthDetail(
		ctx.Request.Context(),
		employeeID,
		month,
		contractType,
	)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toPayrollMonthDetailResponse(detail), "Payroll month detail retrieved successfully"),
	)
}

func (h *SalaryHandler) ExportPayrollMonthSummaryPDF(ctx *gin.Context) {
	var req payrollMonthDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID, month, contractType, err := toPayrollMonthDetailRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	pdfBytes, filename, err := h.service.ExportPayrollMonthPDF(
		ctx.Request.Context(),
		employeeID,
		month,
		contractType,
	)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	ctx.DataFromReader(
		http.StatusOK,
		int64(len(pdfBytes)),
		"application/pdf",
		bytes.NewReader(pdfBytes),
		nil,
	)
}

func (h *SalaryHandler) ClosePayPeriod(ctx *gin.Context) {
	var req closePayPeriodRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toClosePayPeriodParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.ClosePayPeriod(ctx.Request.Context(), adminEmployeeID, params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toPayPeriodResponse(item), "Pay period created successfully"),
	)
}

func (h *SalaryHandler) PreviewFixedPayrollMonthClose(ctx *gin.Context) {
	h.previewPayrollMonthClose(ctx, domain.PayrollGroupFixed)
}

func (h *SalaryHandler) CloseFixedPayrollMonth(ctx *gin.Context) {
	h.closePayrollMonth(ctx, domain.PayrollGroupFixed)
}

func (h *SalaryHandler) PreviewOnCallPayrollMonthClose(ctx *gin.Context) {
	h.previewPayrollMonthClose(ctx, domain.PayrollGroupOnCall)
}

func (h *SalaryHandler) CloseOnCallPayrollMonth(ctx *gin.Context) {
	h.closePayrollMonth(ctx, domain.PayrollGroupOnCall)
}

func (h *SalaryHandler) previewPayrollMonthClose(ctx *gin.Context, payrollGroup string) {
	var req closePayrollMonthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	params, err := toClosePayrollMonthParams(req, payrollGroup)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	result, err := h.service.PreviewPayrollMonthClose(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toPayrollMonthCloseResultResponse(result), "Payroll month close preview retrieved successfully"))
}

func (h *SalaryHandler) closePayrollMonth(ctx *gin.Context, payrollGroup string) {
	var req closePayrollMonthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	params, err := toClosePayrollMonthParams(req, payrollGroup)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	result, err := h.service.ClosePayrollMonthByAdmin(ctx.Request.Context(), adminEmployeeID, params)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusCreated, httpapi.OK(toPayrollMonthCloseResultResponse(result), "Payroll month closed successfully"))
}

func (h *SalaryHandler) ListPayPeriods(ctx *gin.Context) {
	var req listPayPeriodsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListPayPeriods(ctx.Request.Context(), toListPayPeriodsParams(req))
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toPayPeriodResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Pay periods retrieved successfully"))
}

func (h *SalaryHandler) GetPayPeriodByID(ctx *gin.Context) {
	payPeriodID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid pay period id", ""))
		return
	}

	item, err := h.service.GetPayPeriodByID(ctx.Request.Context(), payPeriodID)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toPayPeriodResponse(item), "Pay period retrieved successfully"),
	)
}

func (h *SalaryHandler) MarkPayPeriodPaidByAdmin(ctx *gin.Context) {
	payPeriodID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid pay period id", ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.MarkPayPeriodPaidByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		payPeriodID,
	)
	if err != nil {
		ctx.JSON(mapSalaryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toPayPeriodResponse(item), "Pay period marked as paid successfully"),
	)
}

func mapSalaryErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrSalaryInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrEmployeeNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPayPeriodNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPayPeriodStateInvalid):
		return http.StatusConflict
	case errors.Is(err, domain.ErrPayPeriodAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrPayPeriodNoEntries):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
