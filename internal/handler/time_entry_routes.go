package handler

import "github.com/gin-gonic/gin"

func RegisterTimeEntryRoutes(
	rg *gin.RouterGroup,
	handler *TimeEntryHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/time-entries", auth, requirePermission("TIME_ENTRY.CREATE"), handler.CreateTimeEntry)
	rg.POST(
		"/time-entries/admin",
		auth,
		requirePermission("TIME_ENTRY.CREATE_ALL"),
		handler.CreateTimeEntryByAdmin,
	)
	rg.POST(
		"/time-entries/:id/decision",
		auth,
		requirePermission("TIME_ENTRY.DECIDE"),
		handler.DecideTimeEntryByAdmin,
	)
	rg.GET("/time-entries", auth, requirePermission("TIME_ENTRY.VIEW_ALL"), handler.ListTimeEntries)
	rg.GET(
		"/time-entries/my",
		auth,
		requirePermission("TIME_ENTRY.VIEW"),
		handler.ListMyTimeEntries,
	)
	rg.GET(
		"/time-entries/:id",
		auth,
		requirePermission("TIME_ENTRY.VIEW_ALL"),
		handler.GetTimeEntryByID,
	)
	rg.GET(
		"/time-entries/my/:id",
		auth,
		requirePermission("TIME_ENTRY.VIEW"),
		handler.GetMyTimeEntryByID,
	)
}
