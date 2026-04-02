package handler

import "github.com/gin-gonic/gin"

func RegisterDepartmentRoutes(
	rg *gin.RouterGroup,
	handler *DepartmentHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/departments", auth, requirePermission("EMPLOYEE.CREATE"), handler.CreateDepartment)
	rg.GET("/departments", auth, requirePermission("EMPLOYEE.VIEW"), handler.ListDepartments)
	rg.GET("/departments/:id", auth, requirePermission("EMPLOYEE.VIEW"), handler.GetDepartmentByID)
	rg.PUT("/departments/:id", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateDepartment)
	rg.DELETE("/departments/:id", auth, requirePermission("EMPLOYEE.DELETE"), handler.DeleteDepartment)
}
