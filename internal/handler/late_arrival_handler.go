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

func RegisterLateArrivalRoutes(
	rg *gin.RouterGroup,
	handler *LateArrivalHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST("/late-arrivals", auth, requirePermission("LATE_ARRIVAL.CREATE"), handler.CreateLateArrival)
	rg.POST("/late-arrivals/admin", auth, requirePermission("LATE_ARRIVAL.CREATE_ALL"), handler.CreateLateArrivalByAdmin)
	rg.GET("/late-arrivals/my", auth, requirePermission("LATE_ARRIVAL.VIEW"), handler.ListMyLateArrivals)
	rg.GET("/late-arrivals", auth, requirePermission("LATE_ARRIVAL.VIEW_ALL"), handler.ListLateArrivals)
}

type LateArrivalHandler struct {
	service domain.LateArrivalService
}

func NewLateArrivalHandler(service domain.LateArrivalService) *LateArrivalHandler {
	return &LateArrivalHandler{service: service}
}

func (h *LateArrivalHandler) CreateLateArrival(ctx *gin.Context) {
	var req createLateArrivalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateLateArrivalParams(employeeID, employeeID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateLateArrival(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapLateArrivalErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toCreateLateArrivalResponse(item), "Late arrival created successfully"))
}

func (h *LateArrivalHandler) CreateLateArrivalByAdmin(ctx *gin.Context) {
	var req createLateArrivalByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateLateArrivalByAdminParams(adminEmployeeID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateLateArrivalByAdmin(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapLateArrivalErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toCreateLateArrivalResponse(item), "Late arrival created successfully"))
}

func (h *LateArrivalHandler) ListMyLateArrivals(ctx *gin.Context) {
	var req listMyLateArrivalsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toListMyLateArrivalsParams(employeeID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListMyLateArrivals(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapLateArrivalErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]lateArrivalListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toLateArrivalListItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Late arrivals retrieved successfully"))
}

func (h *LateArrivalHandler) ListLateArrivals(ctx *gin.Context) {
	var req listLateArrivalsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	params, err := toListLateArrivalsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListLateArrivals(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(mapLateArrivalErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	results := make([]lateArrivalListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toLateArrivalListItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Late arrivals retrieved successfully"))
}

func mapLateArrivalErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrLateArrivalInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrLateArrivalConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
