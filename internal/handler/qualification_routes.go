package handler

import "github.com/gin-gonic/gin"

func RegisterQualificationRoutes(
	rg *gin.RouterGroup,
	handler *QualificationHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/qualification-types",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListQualificationTypes,
	)
	rg.POST(
		"/qualification-types",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.CreateQualificationType,
	)
	rg.PUT(
		"/qualification-types/:id",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.UpdateQualificationType,
	)
}
