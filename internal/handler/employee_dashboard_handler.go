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

type EmployeeDashboardHandler struct {
	service domain.EmployeeDashboardService
}

func NewEmployeeDashboardHandler(
	service domain.EmployeeDashboardService,
) *EmployeeDashboardHandler {
	return &EmployeeDashboardHandler{service: service}
}

func (h *EmployeeDashboardHandler) GetKPIs(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	kpis, err := h.service.GetKPIs(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(mapEmployeeDashboardErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toEmployeeDashboardKPIsResponse(kpis),
			"Employee dashboard KPIs retrieved successfully",
		),
	)
}

func (h *EmployeeDashboardHandler) ListPendingRequests(ctx *gin.Context) {
	var req employeeDashboardPendingRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	items, err := h.service.ListPendingRequests(
		ctx.Request.Context(),
		domain.ListEmployeeDashboardPendingRequestsParams{
			EmployeeID: employeeID,
			RecentDays: req.Days,
			Limit:      req.Limit,
		},
	)
	if err != nil {
		ctx.JSON(mapEmployeeDashboardErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toEmployeeDashboardPendingRequestsResponse(items),
			"Employee dashboard pending requests retrieved successfully",
		),
	)
}

func mapEmployeeDashboardErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrEmployeeDashboardInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrEmployeeNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrSalaryInvalidRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
