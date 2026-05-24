package handler

import "github.com/gin-gonic/gin"

func RegisterAttachmentRoutes(
	rg *gin.RouterGroup,
	handler *AttachmentHandler,
	auth gin.HandlerFunc,
) {
	rg.POST(
		"/attachments/upload-url",
		auth,
		handler.RequestUploadURL,
	)
}
