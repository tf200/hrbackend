package handler

import (
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
)

type SalaryHandler struct {
	service domain.SalaryService
}

func NewSalaryHandler(service domain.SalaryService) *SalaryHandler {
	return &SalaryHandler{service: service}
}

func (h *SalaryHandler) ListSalaryScaleSteps(ctx *gin.Context) {
	var req listSalaryScaleStepsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.ListSalaryScaleSteps(ctx.Request.Context(), toListSalaryScaleStepsParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list salary scale steps", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toSalaryScaleStepsResponse(result), "Salary scale steps retrieved successfully"),
	)
}
