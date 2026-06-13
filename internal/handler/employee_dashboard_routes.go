package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterEmployeeDashboardRoutes(
	rg *gin.RouterGroup,
	handler *EmployeeDashboardHandler,
	auth gin.HandlerFunc,
	_ func(permission.Permission) gin.HandlerFunc,
) {
	dashboard := rg.Group("/employee/dashboard", auth)
	{
		dashboard.GET("/kpis", handler.GetKPIs)
		dashboard.GET("/request-tracking", handler.ListPendingRequests)
	}
}
