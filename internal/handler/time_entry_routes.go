package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterTimeEntryRoutes(
	rg *gin.RouterGroup,
	handler *TimeEntryHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST("/time-entries", auth, requirePermission(permission.TimeEntry.Create), handler.CreateTimeEntry)
	rg.POST(
		"/time-entries/admin",
		auth,
		requirePermission(permission.TimeEntry.CreateAll),
		handler.CreateTimeEntryByAdmin,
	)
	rg.POST(
		"/time-entries/:id/decision",
		auth,
		requirePermission(permission.TimeEntry.Decide),
		handler.DecideTimeEntryByAdmin,
	)
	rg.PUT(
		"/time-entries/:id/admin",
		auth,
		requirePermission(permission.TimeEntry.UpdateAll),
		handler.UpdateTimeEntryByAdmin,
	)
	rg.PUT(
		"/time-entries/my/:id",
		auth,
		requirePermission(permission.TimeEntry.Update),
		handler.UpdateMyTimeEntry,
	)
	rg.GET("/time-entries", auth, requirePermission(permission.TimeEntry.ViewAll), handler.ListTimeEntries)
	rg.GET(
		"/time-entries/stats",
		auth,
		requirePermission(permission.TimeEntry.ViewAll),
		handler.GetTimeEntryStats,
	)
	rg.GET(
		"/time-entries/my",
		auth,
		requirePermission(permission.TimeEntry.View),
		handler.ListMyTimeEntries,
	)
	rg.GET(
		"/time-entries/my/stats",
		auth,
		requirePermission(permission.TimeEntry.View),
		handler.GetMyTimeEntryStats,
	)
	rg.GET(
		"/time-entries/:id",
		auth,
		requirePermission(permission.TimeEntry.ViewAll),
		handler.GetTimeEntryByID,
	)
	rg.GET(
		"/time-entries/my/:id",
		auth,
		requirePermission(permission.TimeEntry.View),
		handler.GetMyTimeEntryByID,
	)
}
