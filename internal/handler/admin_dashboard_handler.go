package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
)

type AdminDashboardHandler struct {
	service domain.AdminDashboardService
}

func NewAdminDashboardHandler(service domain.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: service}
}

func (h *AdminDashboardHandler) GetKPIs(ctx *gin.Context) {
	kpis, err := h.service.GetKPIs(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get dashboard KPIs", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toAdminDashboardKPIsResponse(kpis), "Dashboard KPIs retrieved successfully"),
	)
}

func (h *AdminDashboardHandler) GetFullTimeEmployeeBreakdowns(ctx *gin.Context) {
	breakdowns, err := h.service.GetFullTimeEmployeeBreakdowns(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get full-time employee breakdowns", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toFullTimeEmployeeBreakdownsResponse(breakdowns),
			"Full-time employee breakdowns retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) GetLeaveAbsenceTrends(ctx *gin.Context) {
	var req getLeaveAbsenceTrendsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	trends, err := h.service.GetLeaveAbsenceTrends(
		ctx.Request.Context(),
		domain.GetLeaveAbsenceTrendsParams{
			View: req.View,
			Year: req.Year,
		},
	)
	if err != nil {
		if errors.Is(err, domain.ErrAdminDashboardInvalidRequest) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get leave absence trends", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toLeaveAbsenceTrendsResponse(trends),
			"Leave and absence trends retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) GetUpcomingDashboardAlerts(ctx *gin.Context) {
	var req getUpcomingDashboardAlertsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	alerts, err := h.service.GetUpcomingDashboardAlerts(
		ctx.Request.Context(),
		domain.GetUpcomingDashboardAlertsParams{
			Days:  req.Days,
			Limit: req.Limit,
		},
	)
	if err != nil {
		if errors.Is(err, domain.ErrAdminDashboardInvalidRequest) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get upcoming dashboard alerts", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toUpcomingDashboardAlertsResponse(alerts),
			"Upcoming dashboard alerts retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) ListRecentEmployees(ctx *gin.Context) {
	var req listRecentEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	limit := int32(req.PageSize)
	if limit <= 0 || limit > 6 {
		limit = 6
	}
	offset := int32((req.Page - 1) * req.PageSize)

	page, err := h.service.ListRecentEmployees(
		ctx.Request.Context(),
		domain.ListRecentEmployeesParams{
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to list recent employees", ""),
		)
		return
	}

	results := make([]recentEmployeeItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toRecentEmployeeItemResponse(item)
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount),
			"Recent employees retrieved successfully",
		),
	)
}
