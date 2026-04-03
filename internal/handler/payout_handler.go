package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PayoutHandler struct {
	service domain.PayoutService
}

func NewPayoutHandler(service domain.PayoutService) *PayoutHandler {
	return &PayoutHandler{service: service}
}

func (h *PayoutHandler) CreatePayoutRequest(ctx *gin.Context) {
	var req createPayoutRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.CreatePayoutRequest(ctx.Request.Context(), employeeID, toCreatePayoutRequestParams(employeeID, req))
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toPayoutRequestResponse(*item), "Payout request created successfully"))
}

func (h *PayoutHandler) ListMyPayoutRequests(ctx *gin.Context) {
	var req listMyPayoutRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	page, err := h.service.ListMyPayoutRequests(ctx.Request.Context(), toListMyPayoutRequestsParams(employeeID, req))
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, toPayoutRequestResponses(page.Items), page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Payout requests retrieved successfully"))
}

func (h *PayoutHandler) ListPayoutRequests(ctx *gin.Context) {
	var req listPayoutRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListPayoutRequests(ctx.Request.Context(), toListPayoutRequestsParams(req))
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, toPayoutRequestResponses(page.Items), page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Payout requests retrieved successfully"))
}

func (h *PayoutHandler) PreviewPayroll(ctx *gin.Context) {
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
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayrollPreviewResponse(preview), "Payroll preview retrieved successfully"))
}

func (h *PayoutHandler) PreviewMyPayroll(ctx *gin.Context) {
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

	preview, err := h.service.PreviewMyPayroll(ctx.Request.Context(), employeeID, periodStart, periodEnd)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayrollPreviewResponse(preview), "Payroll preview retrieved successfully"))
}

func (h *PayoutHandler) GetPayrollMonthSummary(ctx *gin.Context) {
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

	page, err := h.service.GetPayrollMonthSummary(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, toPayrollMonthSummaryResponses(page.Items), page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Payroll month summary retrieved successfully"))
}

func (h *PayoutHandler) ClosePayPeriod(ctx *gin.Context) {
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
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toPayPeriodResponse(item), "Pay period created successfully"))
}

func (h *PayoutHandler) ListPayPeriods(ctx *gin.Context) {
	var req listPayPeriodsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListPayPeriods(ctx.Request.Context(), toListPayPeriodsParams(req))
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, toPayPeriodResponses(page.Items), page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Pay periods retrieved successfully"))
}

func (h *PayoutHandler) GetPayPeriodByID(ctx *gin.Context) {
	payPeriodID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid pay period id", ""))
		return
	}

	item, err := h.service.GetPayPeriodByID(ctx.Request.Context(), payPeriodID)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayPeriodResponse(item), "Pay period retrieved successfully"))
}

func (h *PayoutHandler) MarkPayPeriodPaidByAdmin(ctx *gin.Context) {
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

	item, err := h.service.MarkPayPeriodPaidByAdmin(ctx.Request.Context(), adminEmployeeID, payPeriodID)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayPeriodResponse(item), "Pay period marked as paid successfully"))
}

func (h *PayoutHandler) DecidePayoutRequestByAdmin(ctx *gin.Context) {
	payoutRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid payout request id", ""))
		return
	}

	var req decidePayoutRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toDecidePayoutRequestParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.DecidePayoutRequestByAdmin(ctx.Request.Context(), adminEmployeeID, payoutRequestID, params)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayoutRequestResponse(*item), "Payout request decided successfully"))
}

func (h *PayoutHandler) MarkPayoutRequestPaidByAdmin(ctx *gin.Context) {
	payoutRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid payout request id", ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.MarkPayoutRequestPaidByAdmin(ctx.Request.Context(), adminEmployeeID, payoutRequestID)
	if err != nil {
		ctx.JSON(mapPayoutErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPayoutRequestResponse(*item), "Payout request marked as paid successfully"))
}

func mapPayoutErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrPayoutRequestInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrEmployeeNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPayoutRequestForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrPayoutRequestNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPayoutRequestStateInvalid):
		return http.StatusConflict
	case errors.Is(err, domain.ErrPayoutRequestInsufficientHours):
		return http.StatusConflict
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
