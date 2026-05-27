package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterQualificationRoutes(
	rg *gin.RouterGroup,
	handler *QualificationHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET(
		"/qualification-types",
		auth,
		requirePermission(permission.Employee.View),
		handler.ListQualificationTypes,
	)
	rg.POST(
		"/qualification-types",
		auth,
		requirePermission(permission.Employee.Create),
		handler.CreateQualificationType,
	)
	rg.PUT(
		"/qualification-types/:id",
		auth,
		requirePermission(permission.Employee.Create),
		handler.UpdateQualificationType,
	)
}
