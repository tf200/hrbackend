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
	default:
		return http.StatusInternalServerError
	}
}
