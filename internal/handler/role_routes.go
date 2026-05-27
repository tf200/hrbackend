package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(
	rg *gin.RouterGroup,
	handler *RoleHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET("/permissions", auth, requirePermission(permission.Role.View), handler.ListAllPermissions)
	rg.GET("/roles", auth, requirePermission(permission.Role.View), handler.ListRoles)
	rg.GET("/roles/:id/permissions", auth, requirePermission(permission.Role.View), handler.ListRolePermissions)
}
