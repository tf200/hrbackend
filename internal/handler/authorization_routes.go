package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterAuthorizationRoutes(
	rg *gin.RouterGroup,
	handler *AuthorizationHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.GET(
		"/authorizations",
		auth,
		requirePermission(permission.Employee.View),
		handler.ListAuthorizations,
	)
	rg.POST(
		"/authorizations",
		auth,
		requirePermission(permission.Employee.Create),
		handler.CreateAuthorization,
	)
	rg.PUT(
		"/authorizations/:id",
		auth,
		requirePermission(permission.Employee.Create),
		handler.UpdateAuthorization,
	)
}
