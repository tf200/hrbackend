package handler

import "github.com/gin-gonic/gin"

func RegisterPayoutRoutes(
	rg *gin.RouterGroup,
	handler *PayoutHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/payroll-preview/my",
		auth,
		requirePermission("PAYOUT.REQUEST.VIEW"),
		handler.PreviewMyPayroll,
	)
	rg.GET(
		"/payroll-preview",
		auth,
		requirePermission("PAYOUT.REQUEST.VIEW_ALL"),
		handler.PreviewPayroll,
	)
	rg.GET(
		"/payroll/ort-rules",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetORTRules,
	)
	rg.GET(
		"/payroll-month-summary",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetPayrollMonthSummary,
	)
	rg.GET(
		"/payroll-month-summary/ort-overview",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetPayrollMonthORTOverview,
	)
	rg.GET(
		"/payroll-month-summary/zzp",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetZZPPayrollMonthSummary,
	)
	rg.GET(
		"/payroll-month-summary/details",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.GetPayrollMonthDetail,
	)
	rg.GET(
		"/payroll-month-summary/export-pdf",
		auth,
		requirePermission("PAY_PERIOD.MONTH_SUMMARY_VIEW"),
		handler.ExportPayrollMonthSummaryPDF,
	)
	rg.POST(
		"/pay-periods/close",
		auth,
		requirePermission("PAY_PERIOD.CLOSE"),
		handler.ClosePayPeriod,
	)
	rg.GET("/pay-periods", auth, requirePermission("PAY_PERIOD.VIEW_ALL"), handler.ListPayPeriods)
	rg.GET(
		"/pay-periods/:id",
		auth,
		requirePermission("PAY_PERIOD.VIEW_ALL"),
		handler.GetPayPeriodByID,
	)
	rg.POST(
		"/pay-periods/:id/mark-paid",
		auth,
		requirePermission("PAY_PERIOD.MARK_PAID"),
		handler.MarkPayPeriodPaidByAdmin,
	)
	rg.POST(
		"/payout-requests",
		auth,
		requirePermission("PAYOUT.REQUEST.CREATE"),
		handler.CreatePayoutRequest,
	)
	rg.GET(
		"/payout-requests/my",
		auth,
		requirePermission("PAYOUT.REQUEST.VIEW"),
		handler.ListMyPayoutRequests,
	)
	rg.GET(
		"/payout-requests",
		auth,
		requirePermission("PAYOUT.REQUEST.VIEW_ALL"),
		handler.ListPayoutRequests,
	)
	rg.POST(
		"/payout-requests/:id/decision",
		auth,
		requirePermission("PAYOUT.REQUEST.DECIDE"),
		handler.DecidePayoutRequestByAdmin,
	)
	rg.POST(
		"/payout-requests/:id/mark-paid",
		auth,
		requirePermission("PAYOUT.REQUEST.MARK_PAID"),
		handler.MarkPayoutRequestPaidByAdmin,
	)
}
