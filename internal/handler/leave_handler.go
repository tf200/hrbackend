package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/domain/permission"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterLeaveRoutes(
	rg *gin.RouterGroup,
	handler *LeaveHandler,
	auth gin.HandlerFunc,
	requirePermission func(permission.Permission) gin.HandlerFunc,
) {
	rg.POST(
		"/leave-requests",
		auth,
		requirePermission(permission.Leave.Request.Create),
		handler.CreateLeaveRequest,
	)
	rg.POST(
		"/leave-requests/admin",
		auth,
		requirePermission(permission.Leave.Request.UpdateAll),
		handler.CreateLeaveRequestByAdmin,
	)
	rg.POST(
		"/leave-requests/:id/decision",
		auth,
		requirePermission(permission.Leave.Request.Decide),
		handler.DecideLeaveRequestByAdmin,
	)
	rg.PUT(
		"/leave-requests/:id",
		auth,
		requirePermission(permission.Leave.Request.Update),
		handler.UpdateLeaveRequest,
	)
	rg.PUT(
		"/leave-requests/:id/admin",
		auth,
		requirePermission(permission.Leave.Request.UpdateAll),
		handler.UpdateLeaveRequestByAdmin,
	)
	rg.GET(
		"/leave-requests",
		auth,
		requirePermission(permission.Leave.Request.ViewAll),
		handler.ListLeaveRequests,
	)
	rg.GET(
		"/leave-requests/calendar",
		auth,
		requirePermission(permission.Leave.Request.ViewAll),
		handler.ListLeaveCalendar,
	)
	rg.GET(
		"/leave-requests/my",
		auth,
		requirePermission(permission.Leave.Request.View),
		handler.ListMyLeaveRequests,
	)
	rg.GET(
		"/leave-requests/my/stats",
		auth,
		requirePermission(permission.Leave.Request.View),
		handler.GetMyLeaveRequestStats,
	)
	rg.GET(
		"/leave-requests/stats",
		auth,
		requirePermission(permission.Leave.Request.ViewAll),
		handler.GetLeaveRequestStats,
	)

	rg.GET(
		"/leave-balances",
		auth,
		requirePermission(permission.Leave.Balance.ViewAll),
		handler.ListLeaveBalances,
	)
	rg.GET(
		"/leave-balances/my",
		auth,
		requirePermission(permission.Leave.Balance.View),
		handler.GetMyLeaveBalance,
	)
	rg.GET(
		"/employees/:id/leave-balance",
		auth,
		requirePermission(permission.Leave.Balance.ViewAll),
		handler.GetEmployeeLeaveBalanceDetails,
	)
}

type LeaveHandler struct {
	service domain.LeaveService
}

func NewLeaveHandler(service domain.LeaveService) *LeaveHandler {
	return &LeaveHandler{service: service}
}

func (h *LeaveHandler) CreateLeaveRequest(ctx *gin.Context) {
	var req createLeaveRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateLeaveRequestParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateLeaveRequest(ctx.Request.Context(), employeeID, params)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toLeaveRequestResponse(item), "Leave request created successfully"),
	)
}

func (h *LeaveHandler) CreateLeaveRequestByAdmin(ctx *gin.Context) {
	var req createLeaveRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateLeaveRequestByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateLeaveRequestByAdmin(ctx.Request.Context(), adminEmployeeID, params)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toLeaveRequestResponse(item), "Leave request created successfully"),
	)
}

func (h *LeaveHandler) UpdateLeaveRequest(ctx *gin.Context) {
	leaveRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid leave request id", ""))
		return
	}

	var req updateLeaveRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toUpdateLeaveRequestParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateLeaveRequest(
		ctx.Request.Context(),
		employeeID,
		leaveRequestID,
		params,
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toLeaveRequestResponse(item), "Leave request updated successfully"),
	)
}

func (h *LeaveHandler) UpdateLeaveRequestByAdmin(ctx *gin.Context) {
	leaveRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid leave request id", ""))
		return
	}

	var req updateLeaveRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, note, err := toUpdateLeaveRequestByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateLeaveRequestByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		leaveRequestID,
		params,
		note,
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toLeaveRequestResponse(item), "Leave request updated successfully"),
	)
}

func (h *LeaveHandler) DecideLeaveRequestByAdmin(ctx *gin.Context) {
	leaveRequestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid leave request id", ""))
		return
	}

	var req decideLeaveRequestByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.DecideLeaveRequestByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		leaveRequestID,
		toDecideLeaveRequestParams(req),
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toLeaveRequestResponse(item), "Leave request decided successfully"),
	)
}

func (h *LeaveHandler) ListMyLeaveRequests(ctx *gin.Context) {
	var req listMyLeaveRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	page, err := h.service.ListMyLeaveRequests(
		ctx.Request.Context(),
		toListMyLeaveRequestsParams(employeeID, req),
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]leaveRequestListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toLeaveRequestListItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Leave requests retrieved successfully"))
}

func (h *LeaveHandler) ListLeaveRequests(ctx *gin.Context) {
	var req listLeaveRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListLeaveRequests(ctx.Request.Context(), toListLeaveRequestsParams(req))
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]leaveRequestListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toLeaveRequestListItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Leave requests retrieved successfully"))
}

func (h *LeaveHandler) ListLeaveCalendar(ctx *gin.Context) {
	req, err := bindListLeaveCalendarRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	items, err := h.service.ListLeaveCalendar(ctx.Request.Context(), toListLeaveCalendarParams(req))
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]leaveCalendarEmployeeResponse, len(items))
	for i, item := range items {
		results[i] = toLeaveCalendarEmployeeResponse(item)
	}

	ctx.JSON(http.StatusOK, httpapi.OK(results, "Leave calendar retrieved successfully"))
}

func (h *LeaveHandler) GetMyLeaveRequestStats(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	stats, err := h.service.GetMyLeaveRequestStats(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toLeaveRequestStatsResponse(stats),
			"Leave request stats retrieved successfully",
		),
	)
}

func (h *LeaveHandler) GetLeaveRequestStats(ctx *gin.Context) {
	stats, err := h.service.GetLeaveRequestStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toLeaveRequestStatsResponse(stats),
			"Leave request stats retrieved successfully",
		),
	)
}

func (h *LeaveHandler) ListLeaveBalances(ctx *gin.Context) {
	var req listLeaveBalancesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListLeaveBalances(ctx.Request.Context(), toListLeaveBalancesParams(req))
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]managerLeaveBalanceResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toManagerLeaveBalanceResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Leave balances retrieved successfully"))
}

func (h *LeaveHandler) GetMyLeaveBalance(ctx *gin.Context) {
	var req getLeaveBalanceDetailsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	details, err := h.service.GetLeaveBalanceDetails(
		ctx.Request.Context(),
		domain.GetLeaveBalanceDetailsParams{
			EmployeeID: employeeID,
			Year:       req.Year,
		},
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toLeaveBalanceResponse(details.Balance),
			"Leave balance retrieved successfully",
		),
	)
}

func (h *LeaveHandler) GetEmployeeLeaveBalanceDetails(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee id", ""))
		return
	}

	var req getLeaveBalanceDetailsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	details, err := h.service.GetLeaveBalanceDetails(
		ctx.Request.Context(),
		toGetLeaveBalanceDetailsParams(employeeID, req),
	)
	if err != nil {
		ctx.JSON(mapLeaveErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toLeaveBalanceDetailsResponse(*details),
			"Leave balance details retrieved successfully",
		),
	)
}

func mapLeaveErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrLeaveRequestInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrLeaveBalanceInvalidAdjust):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrLeaveDurationInvalid):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrLeaveRequestForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrLeaveRequestNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrLeaveRequestStateInvalid):
		return http.StatusConflict
	case errors.Is(err, domain.ErrLeaveBalanceInsufficient):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
