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

type ExpenseHandler struct {
	service domain.ExpenseService
}

func NewExpenseHandler(service domain.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) CreateExpenseRequestByAdmin(ctx *gin.Context) {
	var req createExpenseRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateExpenseRequestByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateExpenseRequestByAdmin(ctx.Request.Context(), adminEmployeeID, params)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request created successfully"),
	)
}

func (h *ExpenseHandler) ListExpenseRequests(ctx *gin.Context) {
	var req listExpenseRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListExpenseRequests(ctx.Request.Context(), toListExpenseRequestsParams(req))
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toExpenseRequestResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Expense requests retrieved successfully"))
}

func (h *ExpenseHandler) GetExpenseRequestByID(ctx *gin.Context) {
	expenseRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid expense request id", ""))
		return
	}

	item, err := h.service.GetExpenseRequestByID(ctx.Request.Context(), expenseRequestID)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request retrieved successfully"),
	)
}

func (h *ExpenseHandler) UpdateExpenseRequestByAdmin(ctx *gin.Context) {
	expenseRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid expense request id", ""))
		return
	}

	var req updateExpenseRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toUpdateExpenseRequestByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateExpenseRequestByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		expenseRequestID,
		params,
	)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request updated successfully"),
	)
}

func (h *ExpenseHandler) DecideExpenseRequestByAdmin(ctx *gin.Context) {
	expenseRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid expense request id", ""))
		return
	}

	var req decideExpenseRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.DecideExpenseRequestByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		expenseRequestID,
		toDecideExpenseRequestParams(req),
	)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request decided successfully"),
	)
}

func (h *ExpenseHandler) CancelExpenseRequestByAdmin(ctx *gin.Context) {
	expenseRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid expense request id", ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.CancelExpenseRequestByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		expenseRequestID,
	)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request cancelled successfully"),
	)
}

func (h *ExpenseHandler) MarkExpenseRequestReimbursedByAdmin(ctx *gin.Context) {
	expenseRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid expense request id", ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.MarkExpenseRequestReimbursedByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		expenseRequestID,
	)
	if err != nil {
		ctx.JSON(mapExpenseErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExpenseRequestResponse(*item), "Expense request marked reimbursed successfully"),
	)
}

func mapExpenseErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrExpenseRequestInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrExpenseRequestForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrExpenseRequestNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrExpenseRequestStateInvalid):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
