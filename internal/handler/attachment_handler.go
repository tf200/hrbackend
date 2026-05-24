package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
)

type AttachmentHandler struct {
	service domain.AttachmentService
}

func NewAttachmentHandler(service domain.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{service: service}
}

func (h *AttachmentHandler) RequestUploadURL(ctx *gin.Context) {
	var req requestUploadURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	resp, err := h.service.RequestUploadURL(
		ctx.Request.Context(),
		req.Filename,
		req.Size,
		req.Tag,
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(err.Error(), ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toUploadURLResponse(resp),
			"Upload URL generated successfully",
		),
	)
}
