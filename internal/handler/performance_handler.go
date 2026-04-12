package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PerformanceHandler struct {
	service domain.PerformanceService
}

func NewPerformanceHandler(service domain.PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{service: service}
}

func (h *PerformanceHandler) CreateAssessment(ctx *gin.Context) {
	var req createPerformanceAssessmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toCreatePerformanceAssessmentParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	created, err := h.service.CreateAssessment(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toPerformanceAssessmentResponse(created), "Assessment created successfully"),
	)
}

func (h *PerformanceHandler) ListAssessments(ctx *gin.Context) {
	var req listPerformanceAssessmentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toListPerformanceAssessmentsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListAssessments(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toPerformanceAssessmentResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Assessments retrieved successfully"))
}

func (h *PerformanceHandler) GetAssessmentByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid assessment id", ""))
		return
	}

	item, err := h.service.GetAssessmentByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceAssessmentResponse(item), "Assessment retrieved successfully"))
}

func (h *PerformanceHandler) DeleteAssessment(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid assessment id", ""))
		return
	}

	deleted, err := h.service.DeleteAssessment(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(gin.H{"deleted": deleted}, "Assessment deleted successfully"))
}

func (h *PerformanceHandler) ListAssessmentScores(ctx *gin.Context) {
	assessmentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid assessment id", ""))
		return
	}

	items, err := h.service.ListAssessmentScores(ctx.Request.Context(), assessmentID)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceAssessmentScoreResponses(items), "Assessment scores retrieved successfully"))
}

func (h *PerformanceHandler) ListWorkAssignments(ctx *gin.Context) {
	var req listPerformanceWorkAssignmentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toListPerformanceWorkAssignmentsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListWorkAssignments(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toPerformanceWorkAssignmentResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Work assignments retrieved successfully"))
}

func (h *PerformanceHandler) GetWorkAssignmentByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid work assignment id", ""))
		return
	}

	item, err := h.service.GetWorkAssignmentByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceWorkAssignmentResponse(item), "Work assignment retrieved successfully"))
}

func (h *PerformanceHandler) DecideWorkAssignment(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid work assignment id", ""))
		return
	}

	var req decidePerformanceWorkAssignmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.DecideWorkAssignment(
		ctx.Request.Context(),
		id,
		toDecidePerformanceWorkAssignmentParams(req),
	)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceWorkAssignmentResponse(item), "Work assignment decided successfully"))
}

func (h *PerformanceHandler) ListUpcoming(ctx *gin.Context) {
	var req listPerformanceUpcomingRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	windowDays := 14
	if req.WindowDays != nil {
		windowDays = *req.WindowDays
	}

	items, err := h.service.ListUpcoming(ctx.Request.Context(), windowDays)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceUpcomingResponses(items), "Upcoming assessments retrieved successfully"))
}

func (h *PerformanceHandler) SendUpcomingInvitations(ctx *gin.Context) {
	var req sendPerformanceUpcomingInvitationsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	sentCount, err := h.service.SendUpcomingInvitations(
		ctx.Request.Context(),
		req.EmployeeIDs,
		trimStringPtr(req.Message),
	)
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(sendPerformanceUpcomingInvitationsResponse{SentCount: sentCount}, "Invitations queued successfully"))
}

func (h *PerformanceHandler) GetStats(ctx *gin.Context) {
	stats, err := h.service.GetStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapPerformanceErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toPerformanceStatsResponse(stats), "Performance stats retrieved successfully"))
}

func mapPerformanceErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrPerformanceInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrPerformanceNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPerformanceStateInvalid):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
