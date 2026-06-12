package handler

import "github.com/gin-gonic/gin"

func RegisterBugReportRoutes(
	rg *gin.RouterGroup,
	handler *BugReportHandler,
	auth gin.HandlerFunc,
) {
	rg.POST("/bug-reports", auth, handler.CreateBugReport)
}
