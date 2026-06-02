package handler

import (
	"github.com/gin-gonic/gin"

	"hrbackend/internal/domain/permission"
)

func RegisterSignDocumentRoutes(rg *gin.RouterGroup, handler *SignDocumentHandler, auth gin.HandlerFunc, requirePermission func(permission.Permission) gin.HandlerFunc) {
	sign := rg.Group("/sign-documents")
	sign.Use(auth)
	sign.GET("", requirePermission(permission.SignDocument.View), handler.ListCreatedDocuments)
	sign.POST("", requirePermission(permission.SignDocument.Create), handler.CreateDocument)
	sign.GET("/me", requirePermission(permission.SignDocument.Self.View), handler.ListMySigningDocuments)
	sign.GET("/me/:document_id", requirePermission(permission.SignDocument.Self.View), handler.GetMySigningDocument)
	sign.POST("/me/:document_id/view", requirePermission(permission.SignDocument.Self.View), handler.MarkViewed)
	sign.POST("/me/:document_id/sign", requirePermission(permission.SignDocument.Self.Sign), handler.Sign)
	sign.GET("/:document_id", requirePermission(permission.SignDocument.View), handler.GetDocument)
	sign.PATCH("/:document_id/fields", requirePermission(permission.SignDocument.Update), handler.SetFields)
	sign.POST("/:document_id/send", requirePermission(permission.SignDocument.Send), handler.SendDocument)
	sign.POST("/:document_id/cancel", requirePermission(permission.SignDocument.Cancel), handler.CancelDocument)
	sign.GET("/:document_id/source-url", requirePermission(permission.SignDocument.View), handler.GetSourceURL)
	sign.GET("/:document_id/signed-url", requirePermission(permission.SignDocument.View), handler.GetSignedURL)
}
