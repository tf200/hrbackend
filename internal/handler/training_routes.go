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

	training.POST(
		"/catalog",
		requirePermission("TRAINING.CATALOG.CREATE"),
		handler.CreateTrainingCatalogItem,
	)
}
