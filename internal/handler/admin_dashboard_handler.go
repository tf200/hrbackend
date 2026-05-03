package handler

import (
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

const defaultOpenShiftCoverageDays int32 = 7

func (h *AdminDashboardHandler) GetAvailabilityStats(ctx *gin.Context) {
	stats, err := h.service.GetAvailabilityStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get availability stats", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminAvailabilityStatsResponse(stats),
			"Availability stats retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) GetCriticalActionStats(ctx *gin.Context) {
	stats, err := h.service.GetCriticalActionStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get critical action stats", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminCriticalActionStatsResponse(stats),
			"Critical action stats retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) GetPayrollTotalStats(ctx *gin.Context) {
	stats, err := h.service.GetPayrollTotalStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get payroll total stats", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminPayrollTotalStatsResponse(stats),
			"Payroll total stats retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) GetRiskRadarStats(ctx *gin.Context) {
	stats, err := h.service.GetRiskRadarStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get risk radar stats", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminRiskRadarStatsResponse(stats),
			"Risk radar stats retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) ListTeamHealthByDepartment(ctx *gin.Context) {
	items, err := h.service.ListTeamHealthByDepartment(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get team health by department", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminTeamHealthByDepartmentResponse(items),
			"Team health by department retrieved successfully",
		),
	)
}

func (h *AdminDashboardHandler) ListOpenShiftCoverage(ctx *gin.Context) {
	var req listOpenShiftCoverageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid query parameters", err.Error()))
		return
	}

	days := defaultOpenShiftCoverageDays
	if req.Days != nil {
		days = *req.Days
	}

	items, err := h.service.ListOpenShiftCoverage(ctx.Request.Context(), days)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get open shift coverage", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toAdminOpenShiftCoverageResponse(days, items),
			"Open shift coverage retrieved successfully",
		),
	)
}
