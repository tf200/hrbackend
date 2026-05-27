package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterSalaryRoutes(
	rg *gin.RouterGroup,
	handler *SalaryHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET(
		"/salary-scale-steps",
		auth,
		requirePermission(permission.Employee.Create),
		handler.ListSalaryScaleSteps,
	)
}
