package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type BugReportHandler struct {
	service domain.BugReportService
}

func NewBugReportHandler(service domain.BugReportService) *BugReportHandler {
	return &BugReportHandler{service: service}
}

func (h *BugReportHandler) CreateBugReport(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	var req createBugReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateBugReport(
		ctx.Request.Context(),
		toCreateBugReportParams(payload.UserID, req),
	)
	if err != nil {
		ctx.JSON(mapBugReportErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toBugReportResponse(item), "Bug report submitted successfully."),
	)
}

func mapBugReportErrorStatus(err error) int {
	if errors.Is(err, domain.ErrBugReportInvalidRequest) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
