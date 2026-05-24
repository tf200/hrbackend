package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthorizationHandler struct {
	service domain.AuthorizationService
}

func NewAuthorizationHandler(service domain.AuthorizationService) *AuthorizationHandler {
	return &AuthorizationHandler{service: service}
}

func (h *AuthorizationHandler) ListAuthorizations(ctx *gin.Context) {
	items, err := h.service.ListAuthorizations(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list authorizations", ""))
		return
	}

	response := make([]authorizationResponse, len(items))
	for i := range items {
		response[i] = toAuthorizationResponse(&items[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Authorizations retrieved successfully"))
}

func (h *AuthorizationHandler) CreateAuthorization(ctx *gin.Context) {
	var req createAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request: name and category are required", ""))
		return
	}

	item, err := h.service.CreateAuthorization(ctx.Request.Context(), domain.CreateAuthorizationParams{
		Name:           req.Name,
		Category:       req.Category,
		Description:    req.Description,
		RequiresExpiry: req.RequiresExpiry,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create authorization", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toAuthorizationResponse(item), "Authorization created successfully"))
}

func (h *AuthorizationHandler) UpdateAuthorization(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid authorization id", ""))
		return
	}

	var req createAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request: name and category are required", ""))
		return
	}

	item, err := h.service.UpdateAuthorization(ctx.Request.Context(), id, domain.CreateAuthorizationParams{
		Name:           req.Name,
		Category:       req.Category,
		Description:    req.Description,
		RequiresExpiry: req.RequiresExpiry,
	})
	if err != nil {
		if errors.Is(err, domain.ErrAuthorizationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail("authorization not found", ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update authorization", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toAuthorizationResponse(item), "Authorization updated successfully"))
}
