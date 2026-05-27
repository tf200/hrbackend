package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterAdminDashboardRoutes(
	rg *gin.RouterGroup,
	handler *AdminDashboardHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET(
		"/admin/dashboard/availability",
		auth,
		requirePermission(permission.Employee.View),
		handler.GetAvailabilityStats,
	)
	rg.GET(
		"/admin/dashboard/open-shift-coverage",
		auth,
		requirePermission(permission.Schedule.View),
		handler.ListOpenShiftCoverage,
	)
	rg.GET(
		"/admin/dashboard/critical-actions",
		auth,
		requirePermission(permission.Employee.View),
		handler.GetCriticalActionStats,
	)
	rg.GET(
		"/admin/dashboard/payroll-totals",
		auth,
		requirePermission(permission.PayPeriod.MonthSummaryView),
		handler.GetPayrollTotalStats,
	)
	rg.GET(
		"/admin/dashboard/risk-radar",
		auth,
		requirePermission(permission.Employee.View),
		handler.GetRiskRadarStats,
	)
	rg.GET(
		"/admin/dashboard/team-health",
		auth,
		requirePermission(permission.Employee.View),
		handler.ListTeamHealthByDepartment,
	)
}
