package handler

import "github.com/gin-gonic/gin"

func RegisterAuthorizationRoutes(
	rg *gin.RouterGroup,
	handler *AuthorizationHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/authorizations",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListAuthorizations,
	)
	rg.POST(
		"/authorizations",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.CreateAuthorization,
	)
	rg.PUT(
		"/authorizations/:id",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.UpdateAuthorization,
	)
}
