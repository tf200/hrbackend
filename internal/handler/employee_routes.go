package handler

import "github.com/gin-gonic/gin"

func RegisterEmployeeRoutes(
	rg *gin.RouterGroup,
	handler *EmployeeHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/employees", auth, requirePermission("EMPLOYEE.CREATE"), handler.CreateEmployee)
	rg.GET("/employees", auth, requirePermission("EMPLOYEE.VIEW"), handler.ListEmployee)
	rg.GET("/employees/counts", auth, requirePermission("EMPLOYEE.VIEW"), handler.GetEmployeeCounts)
	rg.GET("/employees/:id", auth, requirePermission("EMPLOYEE.VIEW"), handler.GetEmployeeByID)
	rg.PUT("/employees/:id", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateEmployee)
	rg.GET("/employees/profile", auth, handler.GetEmployeeProfile)
	rg.PUT("/employees/:id/is_subcontractor", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateIsSubcontractor)
	rg.PUT("/employees/:id/contract_details", auth, requirePermission("EMPLOYEE.UPDATE"), handler.AddContractDetails)
	rg.GET("/employees/:id/contract_details", auth, requirePermission("EMPLOYEE.VIEW"), handler.GetContractDetails)
	rg.POST("/employees/:id/education", auth, requirePermission("EMPLOYEE.CREATE"), handler.AddEducation)
	rg.GET("/employees/:id/education", auth, requirePermission("EMPLOYEE.VIEW"), handler.ListEducation)
	rg.PUT("/employees/:id/education/:education_id", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateEducation)
	rg.DELETE("/employees/:id/education/:education_id", auth, requirePermission("EMPLOYEE.DELETE"), handler.DeleteEducation)
	rg.POST("/employees/:id/experience", auth, requirePermission("EMPLOYEE.CREATE"), handler.AddExperience)
	rg.GET("/employees/:id/experience", auth, requirePermission("EMPLOYEE.VIEW"), handler.ListExperience)
	rg.PUT("/employees/:id/experience/:experience_id", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateExperience)
	rg.DELETE("/employees/:id/experience/:experience_id", auth, requirePermission("EMPLOYEE.DELETE"), handler.DeleteExperience)
	rg.POST("/employees/:id/certification", auth, requirePermission("EMPLOYEE.CREATE"), handler.AddCertification)
	rg.GET("/employees/:id/certification", auth, requirePermission("EMPLOYEE.VIEW"), handler.ListCertification)
	rg.PUT("/employees/:id/certification/:certification_id", auth, requirePermission("EMPLOYEE.UPDATE"), handler.UpdateCertification)
	rg.DELETE("/employees/:id/certification/:certification_id", auth, requirePermission("EMPLOYEE.DELETE"), handler.DeleteCertification)
	rg.GET("/employees/emails", auth, requirePermission("EMPLOYEE.VIEW"), handler.SearchEmployeesByNameOrEmail)
}
