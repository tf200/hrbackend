package handler

import "github.com/gin-gonic/gin"

func RegisterPayoutRoutes(
	rg *gin.RouterGroup,
	handler *PayoutHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET("/payroll-preview/my", auth, requirePermission("PAYOUT.REQUEST.VIEW"), handler.PreviewMyPayroll)
	rg.GET("/payroll-preview", auth, requirePermission("PAYOUT.REQUEST.VIEW_ALL"), handler.PreviewPayroll)
	rg.POST("/payout-requests", auth, requirePermission("PAYOUT.REQUEST.CREATE"), handler.CreatePayoutRequest)
	rg.GET("/payout-requests/my", auth, requirePermission("PAYOUT.REQUEST.VIEW"), handler.ListMyPayoutRequests)
	rg.GET("/payout-requests", auth, requirePermission("PAYOUT.REQUEST.VIEW_ALL"), handler.ListPayoutRequests)
	rg.POST("/payout-requests/:id/decision", auth, requirePermission("PAYOUT.REQUEST.DECIDE"), handler.DecidePayoutRequestByAdmin)
	rg.POST("/payout-requests/:id/mark-paid", auth, requirePermission("PAYOUT.REQUEST.MARK_PAID"), handler.MarkPayoutRequestPaidByAdmin)
}
