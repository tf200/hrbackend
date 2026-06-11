package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterOvertimeRoutes(
	rg *gin.RouterGroup,
	handler *OvertimeHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST(
		"/overtime-entries",
		auth,
		requirePermission(permission.Overtime.Create),
		handler.CreateOvertimeEntry,
	)
	rg.POST(
		"/overtime-entries/admin",
		auth,
		requirePermission(permission.Overtime.CreateAll),
		handler.CreateOvertimeEntryByAdmin,
	)
	rg.POST(
		"/overtime-entries/:id/decision",
		auth,
		requirePermission(permission.Overtime.Decide),
		handler.DecideOvertimeEntryByAdmin,
	)
	rg.PUT(
		"/overtime-entries/:id/admin",
		auth,
		requirePermission(permission.Overtime.UpdateAll),
		handler.UpdateOvertimeEntryByAdmin,
	)
	rg.PUT(
		"/overtime-entries/my/:id",
		auth,
		requirePermission(permission.Overtime.Update),
		handler.UpdateMyOvertimeEntry,
	)
	rg.DELETE(
		"/overtime-entries/:id/admin",
		auth,
		requirePermission(permission.Overtime.DeleteAll),
		handler.DeleteOvertimeEntryByAdmin,
	)
	rg.DELETE(
		"/overtime-entries/my/:id",
		auth,
		requirePermission(permission.Overtime.Delete),
		handler.DeleteMyOvertimeEntry,
	)
	rg.GET(
		"/overtime-entries",
		auth,
		requirePermission(permission.Overtime.ViewAll),
		handler.ListOvertimeEntries,
	)
	rg.GET(
		"/overtime-entries/stats",
		auth,
		requirePermission(permission.Overtime.ViewAll),
		handler.GetOvertimeStats,
	)
	rg.GET(
		"/overtime-entries/my",
		auth,
		requirePermission(permission.Overtime.View),
		handler.ListMyOvertimeEntries,
	)
	rg.GET(
		"/overtime-entries/my/stats",
		auth,
		requirePermission(permission.Overtime.View),
		handler.GetMyOvertimeStats,
	)
	rg.GET(
		"/overtime-entries/:id",
		auth,
		requirePermission(permission.Overtime.ViewAll),
		handler.GetOvertimeEntryByID,
	)
	rg.GET(
		"/overtime-entries/my/:id",
		auth,
		requirePermission(permission.Overtime.View),
		handler.GetMyOvertimeEntryByID,
	)
}
