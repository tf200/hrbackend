package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TrainingHandler struct {
	service domain.TrainingService
}

func NewTrainingHandler(service domain.TrainingService) *TrainingHandler {
	return &TrainingHandler{service: service}
}

func (h *TrainingHandler) CreateTrainingCatalogItem(ctx *gin.Context) {
	var req createTrainingCatalogItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.CreateTrainingCatalogItem(
		ctx.Request.Context(),
		toCreateTrainingCatalogItemParams(req, employeeID),
	)
	if err != nil {
		ctx.JSON(mapTrainingErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toTrainingCatalogItemResponse(item), "Training catalog item created successfully"),
	)
}

func (h *TrainingHandler) AssignTrainingToEmployee(ctx *gin.Context) {
	var req assignTrainingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	actorID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if actorID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.AssignTrainingToEmployee(
		ctx.Request.Context(),
		toAssignTrainingToEmployeeParams(req, actorID),
	)
	if err != nil {
		ctx.JSON(mapTrainingErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toTrainingAssignmentResponse(item), "Training assigned successfully"),
	)
}

func (h *TrainingHandler) CancelTrainingAssignment(ctx *gin.Context) {
	actorID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if actorID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	assignmentID, err := uuid.Parse(ctx.Param("assignment_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid assignment_id", ""))
		return
	}

	var req cancelTrainingAssignmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CancelTrainingAssignment(
		ctx.Request.Context(),
		toCancelTrainingAssignmentParams(assignmentID, req),
	)
	if err != nil {
		ctx.JSON(mapTrainingErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toTrainingAssignmentResponse(item), "Training assignment cancelled successfully"),
	)
}

func (h *TrainingHandler) ListTrainingAssignments(ctx *gin.Context) {
	var req listTrainingAssignmentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListTrainingAssignments(
		ctx.Request.Context(),
		toListTrainingAssignmentsParams(req),
	)
	if err != nil {
		ctx.JSON(mapTrainingErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]trainingAssignmentListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toTrainingAssignmentListItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Training assignments retrieved successfully"))
}

func (h *TrainingHandler) ListTrainingCatalogItems(ctx *gin.Context) {
	var req listTrainingCatalogItemsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListTrainingCatalogItems(
		ctx.Request.Context(),
		toListTrainingCatalogItemsParams(req),
	)
	if err != nil {
		ctx.JSON(mapTrainingErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toTrainingCatalogItemResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Training catalog items retrieved successfully"))
}

func mapTrainingErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrTrainingInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTrainingAssignmentNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrTrainingAssignmentNotCancellable):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTrainingAssignmentConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
