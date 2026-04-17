package handler

import "github.com/gin-gonic/gin"

func RegisterRoleRoutes(
	rg *gin.RouterGroup,
	handler *RoleHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET("/permissions", auth, requirePermission("ROLE.VIEW"), handler.ListAllPermissions)
	rg.GET("/roles", auth, requirePermission("ROLE.VIEW"), handler.ListRoles)
}
