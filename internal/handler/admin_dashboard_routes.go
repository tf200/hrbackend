package handler

import "github.com/gin-gonic/gin"

func RegisterAdminDashboardRoutes(
	rg *gin.RouterGroup,
	handler *AdminDashboardHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/admin/dashboard/availability",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.GetAvailabilityStats,
	)
	rg.GET(
		"/admin/dashboard/open-shift-coverage",
		auth,
		requirePermission("SCHEDULE.VIEW"),
		handler.ListOpenShiftCoverage,
	)
	rg.GET(
		"/admin/dashboard/critical-actions",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.GetCriticalActionStats,
	)
	rg.GET(
		"/admin/dashboard/payroll-totals",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetPayrollTotalStats,
	)
	rg.GET(
		"/admin/dashboard/risk-radar",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.GetRiskRadarStats,
	)
	rg.GET(
		"/admin/dashboard/team-health",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListTeamHealthByDepartment,
	)
}
