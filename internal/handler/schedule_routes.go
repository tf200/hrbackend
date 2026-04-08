package handler

import "github.com/gin-gonic/gin"

func RegisterScheduleRoutes(
	rg *gin.RouterGroup,
	handler *ScheduleHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/schedules", auth, requirePermission("SCHEDULE.CREATE"), handler.CreateSchedule)
	rg.GET(
		"/locations/:id/schedules",
		auth,
		requirePermission("SCHEDULE.VIEW"),
		handler.GetSchedulesByLocationInRange,
	)
	rg.GET(
		"/schedules/by-employee-day",
		auth,
		requirePermission("SCHEDULE.VIEW"),
		handler.GetEmployeeSchedulesByDay,
	)
	rg.GET("/schedules/:id", auth, requirePermission("SCHEDULE.VIEW"), handler.GetScheduleByID)
	rg.PUT("/schedules/:id", auth, requirePermission("SCHEDULE.UPDATE"), handler.UpdateSchedule)
	rg.DELETE("/schedules/:id", auth, requirePermission("SCHEDULE.DELETE"), handler.DeleteSchedule)
	rg.POST(
		"/schedules/auto_generate",
		auth,
		requirePermission("SCHEDULE.CREATE"),
		handler.AutoGenerateSchedules,
	)
	rg.POST(
		"/schedules/save_generated",
		auth,
		requirePermission("SCHEDULE.CREATE"),
		handler.SaveGeneratedSchedules,
	)
}
