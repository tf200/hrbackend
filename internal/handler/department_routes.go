package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterDepartmentRoutes(
	rg *gin.RouterGroup,
	handler *DepartmentHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST(
		"/departments",
		auth,
		requirePermission(permission.Employee.Create),
		handler.CreateDepartment,
	)
	rg.GET(
		"/departments",
		auth,
		requirePermission(permission.Employee.View),
		handler.ListDepartments,
	)
	rg.GET(
		"/departments/:id",
		auth,
		requirePermission(permission.Employee.View),
		handler.GetDepartmentByID,
	)
	rg.PUT(
		"/departments/:id",
		auth,
		requirePermission(permission.Employee.Update),
		handler.UpdateDepartment,
	)
	rg.DELETE(
		"/departments/:id",
		auth,
		requirePermission(permission.Employee.Delete),
		handler.DeleteDepartment,
	)
}
