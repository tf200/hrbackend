package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleHandler struct {
	service domain.RoleService
}

func NewRoleHandler(service domain.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) ListRoles(ctx *gin.Context) {
	items, err := h.service.ListRoles(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list roles", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toRoleResponses(items), "Roles retrieved successfully"),
	)
}

func (h *RoleHandler) ListAllPermissions(ctx *gin.Context) {
	items, err := h.service.ListAllPermissions(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list permissions", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toPermissionCatalogResponses(items),
			"Permissions retrieved successfully",
		),
	)
}

func (h *RoleHandler) ListRolePermissions(ctx *gin.Context) {
	roleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid role ID", ""))
		return
	}

	items, err := h.service.ListRolePermissions(ctx.Request.Context(), roleID)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail("role not found", ""))
			return
		}

		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list role permissions", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toRolePermissionResponses(items),
			"Role permissions retrieved successfully",
		),
	)
}
