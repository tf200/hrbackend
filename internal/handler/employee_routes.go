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
	rg.POST(
		"/employees/:id/education",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.AddEducation,
	)
	rg.GET(
		"/employees/:id/education",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListEducation,
	)
	rg.PUT(
		"/employees/:id/education/:education_id",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.UpdateEducation,
	)
	rg.DELETE(
		"/employees/:id/education/:education_id",
		auth,
		requirePermission("EMPLOYEE.DELETE"),
		handler.DeleteEducation,
	)
	rg.POST(
		"/employees/:id/experience",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.AddExperience,
	)
	rg.GET(
		"/employees/:id/experience",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListExperience,
	)
	rg.PUT(
		"/employees/:id/experience/:experience_id",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.UpdateExperience,
	)
	rg.DELETE(
		"/employees/:id/experience/:experience_id",
		auth,
		requirePermission("EMPLOYEE.DELETE"),
		handler.DeleteExperience,
	)
	rg.POST(
		"/employees/:id/qualifications",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.AddQualification,
	)
	rg.GET(
		"/employees/:id/qualifications",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListQualification,
	)
	rg.PUT(
		"/employees/:id/qualifications/:qualification_id",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.UpdateQualification,
	)
	rg.DELETE(
		"/employees/:id/qualifications/:qualification_id",
		auth,
		requirePermission("EMPLOYEE.DELETE"),
		handler.DeleteQualification,
	)
	rg.POST(
		"/employees/:id/authorizations",
		auth,
		requirePermission("EMPLOYEE.CREATE"),
		handler.AddEmployeeAuthorization,
	)
	rg.GET(
		"/employees/:id/authorizations",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListEmployeeAuthorizations,
	)
	rg.PUT(
		"/employees/:id/authorizations/:authorization_id",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.UpdateEmployeeAuthorization,
	)
	rg.DELETE(
		"/employees/:id/authorizations/:authorization_id",
		auth,
		requirePermission("EMPLOYEE.DELETE"),
		handler.DeleteEmployeeAuthorization,
	)
	rg.GET(
		"/employees/:id/contracts",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.ListEmployeeContracts,
	)
	rg.POST(
		"/employees/:id/contracts",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.CreateContract,
	)
	rg.PUT(
		"/employees/:id/contracts/:contract_id",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.UpdateContract,
	)
	rg.POST(
		"/employees/:id/contracts/:contract_id/amendments",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.CreateContractAmendment,
	)
	rg.POST(
		"/employees/:id/reset_password",
		auth,
		requirePermission("EMPLOYEE.UPDATE"),
		handler.ResetPassword,
	)
	rg.GET(
		"/employees/emails",
		auth,
		requirePermission("EMPLOYEE.VIEW"),
		handler.SearchEmployeesByNameOrEmail,
	)
}
