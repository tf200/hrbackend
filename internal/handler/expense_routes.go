package handler

import "github.com/gin-gonic/gin"

func RegisterExpenseRoutes(
	rg *gin.RouterGroup,
	handler *ExpenseHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST(
		"/expense-requests/admin",
		auth,
		requirePermission("EXPENSE.REQUEST.CREATE"),
		handler.CreateExpenseRequestByAdmin,
	)
	rg.GET(
		"/expense-requests",
		auth,
		requirePermission("EXPENSE.REQUEST.VIEW_ALL"),
		handler.ListExpenseRequests,
	)
	rg.GET(
		"/expense-requests/:id",
		auth,
		requirePermission("EXPENSE.REQUEST.VIEW_ALL"),
		handler.GetExpenseRequestByID,
	)
	rg.PUT(
		"/expense-requests/:id/admin",
		auth,
		requirePermission("EXPENSE.REQUEST.UPDATE"),
		handler.UpdateExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/decision",
		auth,
		requirePermission("EXPENSE.REQUEST.DECIDE"),
		handler.DecideExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/cancel",
		auth,
		requirePermission("EXPENSE.REQUEST.UPDATE"),
		handler.CancelExpenseRequestByAdmin,
	)
	rg.POST(
		"/expense-requests/:id/mark-reimbursed",
		auth,
		requirePermission("EXPENSE.REQUEST.MARK_REIMBURSED"),
		handler.MarkExpenseRequestReimbursedByAdmin,
	)
}
