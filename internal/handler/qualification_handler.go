package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QualificationHandler struct {
	service domain.QualificationTypeService
}

func NewQualificationHandler(service domain.QualificationTypeService) *QualificationHandler {
	return &QualificationHandler{service: service}
}

func (h *QualificationHandler) ListQualificationTypes(ctx *gin.Context) {
	types, err := h.service.ListQualificationTypes(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list qualification types", ""))
		return
	}

	response := make([]qualificationTypeResponse, len(types))
	for i := range types {
		response[i] = toQualificationTypeResponse(&types[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Qualification types retrieved successfully"))
}

func (h *QualificationHandler) CreateQualificationType(ctx *gin.Context) {
	var req createQualificationTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request: name is required", ""))
		return
	}

	qualification, err := h.service.CreateQualificationType(ctx.Request.Context(), domain.CreateQualificationTypeParams{
		Name: req.Name,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create qualification type", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toQualificationTypeResponse(qualification), "Qualification type created successfully"))
}

func (h *QualificationHandler) UpdateQualificationType(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid qualification type id", ""))
		return
	}

	var req createQualificationTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request: name is required", ""))
		return
	}

	qualification, err := h.service.UpdateQualificationType(ctx.Request.Context(), id, domain.CreateQualificationTypeParams{
		Name: req.Name,
	})
	if err != nil {
		if errors.Is(err, domain.ErrQualificationTypeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail("qualification type not found", ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update qualification type", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toQualificationTypeResponse(qualification), "Qualification type updated successfully"))
}
