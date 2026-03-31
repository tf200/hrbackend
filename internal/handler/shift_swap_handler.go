package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterShiftSwapRoutes(
	rg *gin.RouterGroup,
	handler *ShiftSwapHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/shift-swaps", auth, requirePermission("SCHEDULE_SWAP.REQUEST"), handler.CreateShiftSwapRequest)
	rg.POST("/shift-swaps/:id/respond", auth, requirePermission("SCHEDULE_SWAP.RESPOND"), handler.RespondShiftSwapRequest)
	rg.POST("/shift-swaps/:id/admin-decision", auth, requirePermission("SCHEDULE_SWAP.APPROVE"), handler.AdminDecisionShiftSwapRequest)
	rg.GET("/shift-swaps", auth, requirePermission("SCHEDULE_SWAP.APPROVE"), handler.ListShiftSwapRequests)
	rg.GET("/shift-swaps/my", auth, requirePermission("SCHEDULE_SWAP.VIEW"), handler.ListMyShiftSwapRequests)
}

type ShiftSwapHandler struct {
	service domain.ScheduleService
}

func NewShiftSwapHandler(service domain.ScheduleService) *ShiftSwapHandler {
	return &ShiftSwapHandler{service: service}
}

func (h *ShiftSwapHandler) CreateShiftSwapRequest(ctx *gin.Context) {
	var req createShiftSwapRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.CreateShiftSwapRequest(ctx.Request.Context(), employeeID, &domain.CreateShiftSwapRequest{
		RecipientEmployeeID: req.RecipientEmployeeID,
		RequesterScheduleID: req.RequesterScheduleID,
		RecipientScheduleID: req.RecipientScheduleID,
		ExpiresAt:           req.ExpiresAt,
	})
	if err != nil {
		ctx.JSON(mapShiftSwapErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toCreateShiftSwapResponse(item), "Shift swap request created successfully"))
}

func (h *ShiftSwapHandler) RespondShiftSwapRequest(ctx *gin.Context) {
	swapID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	var req respondShiftSwapRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.RespondToShiftSwapRequest(ctx.Request.Context(), employeeID, swapID, &domain.RespondShiftSwapRequest{
		Decision: req.Decision,
		Note:     req.Note,
	})
	if err != nil {
		ctx.JSON(mapShiftSwapErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toShiftSwapResponse(*item), "Shift swap request response recorded successfully"))
}

func (h *ShiftSwapHandler) AdminDecisionShiftSwapRequest(ctx *gin.Context) {
	swapID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	var req adminDecisionShiftSwapRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.AdminDecisionShiftSwapRequest(ctx.Request.Context(), employeeID, swapID, &domain.AdminDecisionShiftSwapRequest{
		Decision: req.Decision,
		Note:     req.Note,
	})
	if err != nil {
		ctx.JSON(mapShiftSwapErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toShiftSwapResponse(*item), "Shift swap request admin decision recorded successfully"))
}

func (h *ShiftSwapHandler) ListMyShiftSwapRequests(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	items, err := h.service.ListMyShiftSwapRequests(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(mapShiftSwapErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := make([]shiftSwapResponse, len(items))
	for i, item := range items {
		response[i] = toShiftSwapResponse(item)
	}
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Shift swap requests retrieved successfully"))
}

func (h *ShiftSwapHandler) ListShiftSwapRequests(ctx *gin.Context) {
	var req listShiftSwapRequestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListShiftSwapRequests(ctx.Request.Context(), toListShiftSwapParams(req))
	if err != nil {
		ctx.JSON(mapShiftSwapErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]shiftSwapResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toShiftSwapResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Shift swap requests retrieved successfully"))
}

func mapShiftSwapErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrShiftSwapInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrScheduleNotFound),
		errors.Is(err, domain.ErrShiftSwapNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrShiftSwapStateInvalid),
		errors.Is(err, domain.ErrShiftSwapDuplicateActiveRequest),
		errors.Is(err, domain.ErrShiftSwapExpired),
		errors.Is(err, domain.ErrShiftSwapScheduleOwnership),
		errors.Is(err, domain.ErrShiftSwapConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
