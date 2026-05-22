package handler

import "github.com/gin-gonic/gin"

func RegisterSalaryRoutes(
	rg *gin.RouterGroup,
	handler *SalaryHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/salary-scale-steps",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.ListSalaryScaleSteps,
	)
}
