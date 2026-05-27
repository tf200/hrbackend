package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterScheduleRoutes(
	rg *gin.RouterGroup,
	handler *ScheduleHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST("/schedules", auth, requirePermission(permission.Schedule.Create), handler.CreateSchedule)
	rg.GET(
		"/locations/:id/schedules",
		auth,
		requirePermission(permission.Schedule.View),
		handler.GetSchedulesByLocationInRange,
	)
	rg.GET(
		"/schedules/by-employee-day",
		auth,
		requirePermission(permission.Schedule.View),
		handler.GetEmployeeSchedulesByDay,
	)
	rg.GET(
		"/employees/:id/schedules",
		auth,
		requirePermission(permission.Schedule.View),
		handler.GetEmployeeSchedulesTimeline,
	)
	rg.GET(
		"/me/shifts/overview",
		auth,
		requirePermission(permission.Portal.EmployeeAccess),
		handler.GetMyShiftOverview,
	)
	rg.GET(
		"/me/shifts/upcoming",
		auth,
		requirePermission(permission.Portal.EmployeeAccess),
		handler.GetMyUpcomingShifts,
	)
	rg.GET(
		"/me/shifts/past",
		auth,
		requirePermission(permission.Portal.EmployeeAccess),
		handler.GetMyPastShifts,
	)
	rg.GET("/schedules/:id", auth, requirePermission(permission.Schedule.View), handler.GetScheduleByID)
	rg.PUT("/schedules/:id", auth, requirePermission(permission.Schedule.Update), handler.UpdateSchedule)
	rg.DELETE("/schedules/:id", auth, requirePermission(permission.Schedule.Delete), handler.DeleteSchedule)
	rg.POST(
		"/schedules/auto_generate",
		auth,
		requirePermission(permission.Schedule.Create),
		handler.AutoGenerateSchedules,
	)
	rg.POST(
		"/schedules/save_generated",
		auth,
		requirePermission(permission.Schedule.Create),
		handler.SaveGeneratedSchedules,
	)
}
