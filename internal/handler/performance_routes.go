package handler

import "github.com/gin-gonic/gin"

func RegisterPerformanceRoutes(
	rg *gin.RouterGroup,
	handler *PerformanceHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/performance-mine",
		auth,
		handler.GetMine,
	)

	rg.GET(
		"/performance-assessment-catalog",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.VIEW_ALL"),
		handler.ListAssessmentCatalog,
	)
	rg.POST(
		"/performance-assessments",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.CREATE"),
		handler.CreateAssessment,
	)
	rg.GET(
		"/performance-assessments",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.VIEW_ALL"),
		handler.ListAssessments,
	)
	rg.GET(
		"/performance-assessments/:id",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.VIEW_ALL"),
		handler.GetAssessmentByID,
	)
	rg.DELETE(
		"/performance-assessments/:id",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.DELETE"),
		handler.DeleteAssessment,
	)
	rg.GET(
		"/performance-assessments/:id/scores",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.VIEW_ALL"),
		handler.ListAssessmentScores,
	)

	rg.GET(
		"/performance-work-assignments",
		auth,
		requirePermission("PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL"),
		handler.ListWorkAssignments,
	)
	rg.GET(
		"/performance-work-assignments/:id",
		auth,
		requirePermission("PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL"),
		handler.GetWorkAssignmentByID,
	)
	rg.POST(
		"/performance-work-assignments/:id/decision",
		auth,
		requirePermission("PERFORMANCE.WORK_ASSIGNMENT.DECIDE"),
		handler.DecideWorkAssignment,
	)

	rg.GET(
		"/performance-upcoming",
		auth,
		requirePermission("PERFORMANCE.ASSESSMENT.VIEW_ALL"),
		handler.ListUpcoming,
	)
	rg.POST(
		"/performance-upcoming/invitations",
		auth,
		requirePermission("PERFORMANCE.UPCOMING.INVITE"),
		handler.SendUpcomingInvitations,
	)

	rg.GET(
		"/performance-stats",
		auth,
		requirePermission("PERFORMANCE.STATS.VIEW"),
		handler.GetStats,
	)
}
