package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"
)

type EmployeeSignatureProfileHandler struct {
	service domain.EmployeeSignatureProfileService
}

func NewEmployeeSignatureProfileHandler(
	service domain.EmployeeSignatureProfileService,
) *EmployeeSignatureProfileHandler {
	return &EmployeeSignatureProfileHandler{service: service}
}

func (h *EmployeeSignatureProfileHandler) GetMySignatureProfile(ctx *gin.Context) {
	employeeID, ok := employeeIDFromRequest(ctx)
	if !ok {
		return
	}
	profile, err := h.service.GetMySignatureProfile(ctx.Request.Context(), employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeSignatureProfileNotFound) {
			ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeSignatureProfileResponse(nil, nil), "Signature profile retrieved"))
			return
		}
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	var imageURL *string
	if profile.ImageFileKey != nil {
		url, err := h.service.GetSignatureImageURL(ctx.Request.Context(), employeeID)
		if err == nil {
			imageURL = &url
		}
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeSignatureProfileResponse(profile, imageURL), "Signature profile retrieved"))
}

func (h *EmployeeSignatureProfileHandler) UpsertMySignatureProfile(ctx *gin.Context) {
	employeeID, ok := employeeIDFromRequest(ctx)
	if !ok {
		return
	}
	var req upsertEmployeeSignatureProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", "INVALID_REQUEST"))
		return
	}
	profile, err := h.service.UpsertMySignatureProfile(
		ctx.Request.Context(),
		employeeID,
		toUpsertEmployeeSignatureProfileParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), employeeSignatureProfileErrorCode(err)))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeSignatureProfileResponse(profile, nil), "Signature profile saved"))
}

func (h *EmployeeSignatureProfileHandler) DeleteMySignatureProfile(ctx *gin.Context) {
	employeeID, ok := employeeIDFromRequest(ctx)
	if !ok {
		return
	}
	if err := h.service.DeleteMySignatureProfile(ctx.Request.Context(), employeeID); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), employeeSignatureProfileErrorCode(err)))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(gin.H{"deleted": true}, "Signature profile deleted"))
}

func (h *EmployeeSignatureProfileHandler) RequestUploadURL(ctx *gin.Context) {
	employeeID, ok := employeeIDFromRequest(ctx)
	if !ok {
		return
	}
	var req requestSignatureUploadURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", "INVALID_REQUEST"))
		return
	}
	resp, err := h.service.RequestUploadURL(
		ctx.Request.Context(),
		employeeID,
		toRequestSignatureUploadURLParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), employeeSignatureProfileErrorCode(err)))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toSignatureUploadURLResponse(resp), "Signature upload URL generated"))
}

func (h *EmployeeSignatureProfileHandler) GetSignatureImageURL(ctx *gin.Context) {
	employeeID, ok := employeeIDFromRequest(ctx)
	if !ok {
		return
	}
	url, err := h.service.GetSignatureImageURL(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), employeeSignatureProfileErrorCode(err)))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(signatureImageURLResponse{URL: url}, "Signature image URL generated"))
}

func employeeIDFromRequest(ctx *gin.Context) (uuid.UUID, bool) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return uuid.Nil, false
	}
	return employeeID, true
}

func employeeSignatureProfileErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrEmployeeSignatureProfileInvalid):
		return "INVALID_SIGNATURE_PROFILE"
	case errors.Is(err, domain.ErrEmployeeSignatureProfileNotFound):
		return "SIGNATURE_PROFILE_NOT_FOUND"
	case errors.Is(err, domain.ErrEmployeeSignatureProfileRequired):
		return "SIGNATURE_PROFILE_REQUIRED"
	default:
		return ""
	}
}
