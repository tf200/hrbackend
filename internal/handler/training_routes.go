package handler

import "github.com/gin-gonic/gin"

func RegisterTrainingRoutes(
	rg *gin.RouterGroup,
	handler *TrainingHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	training := rg.Group("/training")
	training.Use(auth)

	training.GET(
		"/catalog",
		requirePermission("TRAINING.CATALOG.VIEW"),
		handler.ListTrainingCatalogItems,
	)
	training.GET(
		"/assignments",
		requirePermission("TRAINING.ASSIGNMENTS.VIEW"),
		handler.ListTrainingAssignments,
	)
	training.POST(
		"/assignments",
		requirePermission("TRAINING.ASSIGN"),
		handler.AssignTrainingToEmployee,
	)
	training.POST(
		"/assignments/:assignment_id/cancel",
		requirePermission("TRAINING.ASSIGN"),
		handler.CancelTrainingAssignment,
	)
	training.POST(
		"/catalog",
		requirePermission("TRAINING.CATALOG.CREATE"),
		handler.CreateTrainingCatalogItem,
	)
}
