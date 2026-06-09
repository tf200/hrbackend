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
	dashboard := rg.Group("/admin/dashboard", auth)
	{
		dashboard.GET("/kpis", requirePermission(permission.Employee.View), handler.GetKPIs)
		dashboard.GET(
			"/recent-employees",
			requirePermission(permission.Employee.View),
			handler.ListRecentEmployees,
		)
		dashboard.GET(
			"/full-time-employee-breakdowns",
			requirePermission(permission.Employee.View),
			handler.GetFullTimeEmployeeBreakdowns,
		)
		dashboard.GET(
			"/leave-absence-trends",
			requirePermission(permission.Employee.View),
			handler.GetLeaveAbsenceTrends,
		)
		dashboard.GET(
			"/upcoming-alerts",
			requirePermission(permission.Employee.View),
			handler.GetUpcomingDashboardAlerts,
		)
	}
}
