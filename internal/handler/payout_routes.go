package handler

import (
	"hrbackend/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func RegisterPayoutRoutes(
	rg *gin.RouterGroup,
	handler *PayoutHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST(
		"/payout-requests/admin",
		auth,
		requirePermission(permission.Payout.Request.Decide),
		handler.CreatePayoutRequestByAdmin,
	)
	rg.POST(
		"/payout-requests",
		auth,
		requirePermission(permission.Payout.Request.Create),
		handler.CreatePayoutRequest,
	)
	rg.GET(
		"/payout-requests/my",
		auth,
		requirePermission(permission.Payout.Request.View),
		handler.ListMyPayoutRequests,
	)
	rg.GET(
		"/payout-requests",
		auth,
		requirePermission(permission.Payout.Request.ViewAll),
		handler.ListPayoutRequests,
	)
	rg.POST(
		"/payout-requests/:id/decision",
		auth,
		requirePermission(permission.Payout.Request.Decide),
		handler.DecidePayoutRequestByAdmin,
	)
	rg.POST(
		"/payout-requests/:id/mark-paid",
		auth,
		requirePermission(permission.Payout.Request.MarkPaid),
		handler.MarkPayoutRequestPaidByAdmin,
	)
}
