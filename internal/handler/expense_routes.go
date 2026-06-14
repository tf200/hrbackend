package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterExpenseRoutes(
	rg *gin.RouterGroup,
	handler *ExpenseHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST(
		"/expense-requests",
		auth,
		requirePermission(permission.Expense.Request.Create),
		handler.CreateExpenseRequest,
	)
	rg.POST(
		"/expense-requests/admin",
		auth,
		requirePermission(permission.Expense.Request.Create),
		handler.CreateExpenseRequestByAdmin,
	)
	rg.GET(
		"/expense-requests/my",
		auth,
		requirePermission(permission.Expense.Request.View),
		handler.ListMyExpenseRequests,
	)
	rg.GET(
		"/expense-requests/my/stats",
		auth,
		requirePermission(permission.Expense.Request.View),
		handler.GetMyExpenseStats,
	)
	rg.GET(
		"/expense-requests",
		auth,
		requirePermission(permission.Expense.Request.ViewAll),
		handler.ListExpenseRequests,
	)
	rg.GET(
		"/expense-requests/:id",
		auth,
		requirePermission(permission.Expense.Request.ViewAll),
		handler.GetExpenseRequestByID,
	)
	rg.PUT(
		"/expense-requests/:id",
		auth,
		requirePermission(permission.Expense.Request.Update),
		handler.UpdateExpenseRequest,
	)
	rg.DELETE(
		"/expense-requests/:id",
		auth,
		requirePermission(permission.Expense.Request.Update),
		handler.DeleteExpenseRequest,
	)
	rg.PUT(
		"/expense-requests/:id/admin",
		auth,
		requirePermission(permission.Expense.Request.Update),
		handler.UpdateExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/decision",
		auth,
		requirePermission(permission.Expense.Request.Decide),
		handler.DecideExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/cancel",
		auth,
		requirePermission(permission.Expense.Request.Update),
		handler.CancelExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/mark-reimbursed",
		auth,
		requirePermission(permission.Expense.Request.MarkReimbursed),
		handler.MarkExpenseRequestReimbursedByAdmin,
	)
}
