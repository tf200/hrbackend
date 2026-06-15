package handler

import "github.com/gin-gonic/gin"

func RegisterEmployeeSignatureProfileRoutes(
	rg *gin.RouterGroup,
	handler *EmployeeSignatureProfileHandler,
	auth gin.HandlerFunc,
) {
	signature := rg.Group("/signature-profile/me")
	signature.Use(auth)
	signature.GET("", handler.GetMySignatureProfile)
	signature.PUT("", handler.UpsertMySignatureProfile)
	signature.DELETE("", handler.DeleteMySignatureProfile)
	signature.POST("/upload-url", handler.RequestUploadURL)
	signature.GET("/image-url", handler.GetSignatureImageURL)
}
