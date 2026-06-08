package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterTrainingRoutes(
	rg *gin.RouterGroup,
	handler *TrainingHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	training := rg.Group("/training")
	training.Use(auth)

	training.GET(
		"/catalog",
		requirePermission(permission.Training.Catalog.View),
		handler.ListTrainingCatalogItems,
	)
	training.GET(
		"/assignments",
		requirePermission(permission.Training.AssignmentsView),
		handler.ListTrainingAssignments,
	)
	training.GET(
		"/assignments/my",
		requirePermission(permission.Portal.EmployeeAccess),
		handler.ListMyTrainingAssignments,
	)
	training.GET(
		"/assignments/my/counts",
		requirePermission(permission.Portal.EmployeeAccess),
		handler.GetMyTrainingAssignmentsCounts,
	)
	training.POST(
		"/assignments",
		requirePermission(permission.Training.Assign),
		handler.AssignTrainingToEmployee,
	)
	training.POST(
		"/assignments/:assignment_id/cancel",
		requirePermission(permission.Training.Assign),
		handler.CancelTrainingAssignment,
	)
	training.POST(
		"/catalog",
		requirePermission(permission.Training.Catalog.Create),
		handler.CreateTrainingCatalogItem,
	)
}
