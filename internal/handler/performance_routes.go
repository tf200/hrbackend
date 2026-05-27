package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterPerformanceRoutes(
	rg *gin.RouterGroup,
	handler *PerformanceHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET(
		"/performance-mine",
		auth,
		handler.GetMine,
	)

	rg.GET(
		"/performance-assessment-catalog",
		auth,
		requirePermission(permission.Performance.Assessment.ViewAll),
		handler.ListAssessmentCatalog,
	)
	rg.POST(
		"/performance-assessments",
		auth,
		requirePermission(permission.Performance.Assessment.Create),
		handler.CreateAssessment,
	)
	rg.GET(
		"/performance-assessments",
		auth,
		requirePermission(permission.Performance.Assessment.ViewAll),
		handler.ListAssessments,
	)
	rg.GET(
		"/performance-assessments/:id",
		auth,
		requirePermission(permission.Performance.Assessment.ViewAll),
		handler.GetAssessmentByID,
	)
	rg.DELETE(
		"/performance-assessments/:id",
		auth,
		requirePermission(permission.Performance.Assessment.Delete),
		handler.DeleteAssessment,
	)
	rg.GET(
		"/performance-assessments/:id/scores",
		auth,
		requirePermission(permission.Performance.Assessment.ViewAll),
		handler.ListAssessmentScores,
	)

	rg.GET(
		"/performance-work-assignments",
		auth,
		requirePermission(permission.Performance.WorkAssignment.ViewAll),
		handler.ListWorkAssignments,
	)
	rg.GET(
		"/performance-work-assignments/:id",
		auth,
		requirePermission(permission.Performance.WorkAssignment.ViewAll),
		handler.GetWorkAssignmentByID,
	)
	rg.POST(
		"/performance-work-assignments/:id/decision",
		auth,
		requirePermission(permission.Performance.WorkAssignment.Decide),
		handler.DecideWorkAssignment,
	)

	rg.GET(
		"/performance-upcoming",
		auth,
		requirePermission(permission.Performance.Assessment.ViewAll),
		handler.ListUpcoming,
	)
	rg.POST(
		"/performance-upcoming/invitations",
		auth,
		requirePermission(permission.Performance.Upcoming.Invite),
		handler.SendUpcomingInvitations,
	)

	rg.GET(
		"/performance-stats",
		auth,
		requirePermission(permission.Performance.Stats),
		handler.GetStats,
	)
}
