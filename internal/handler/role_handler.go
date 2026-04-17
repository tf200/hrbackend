package handler

import (
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
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
