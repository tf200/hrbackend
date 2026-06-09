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

type OvertimeHandler struct {
	service domain.OvertimeService
}

func NewOvertimeHandler(service domain.OvertimeService) *OvertimeHandler {
	return &OvertimeHandler{service: service}
}

func (h *OvertimeHandler) CreateOvertimeEntry(ctx *gin.Context) {
	var req createOvertimeEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateOvertimeEntryParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateOvertimeEntry(ctx.Request.Context(), employeeID, params)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry created successfully"),
	)
}

func (h *OvertimeHandler) CreateOvertimeEntryByAdmin(ctx *gin.Context) {
	var req createOvertimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateOvertimeEntryByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateOvertimeEntryByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		params,
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry created successfully"),
	)
}

func (h *OvertimeHandler) DecideOvertimeEntryByAdmin(ctx *gin.Context) {
	overtimeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid overtime entry id", ""))
		return
	}

	var req decideOvertimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.DecideOvertimeEntryByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		overtimeEntryID,
		toDecideOvertimeEntryParams(req),
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry decided successfully"),
	)
}

func (h *OvertimeHandler) UpdateOvertimeEntryByAdmin(ctx *gin.Context) {
	overtimeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid overtime entry id", ""))
		return
	}

	var req updateOvertimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toUpdateOvertimeEntryByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateOvertimeEntryByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		overtimeEntryID,
		params,
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry updated successfully"),
	)
}

func (h *OvertimeHandler) UpdateMyOvertimeEntry(ctx *gin.Context) {
	overtimeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid overtime entry id", ""))
		return
	}

	var req updateMyOvertimeEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toUpdateMyOvertimeEntryParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateMyOvertimeEntry(
		ctx.Request.Context(),
		employeeID,
		overtimeEntryID,
		params,
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry updated successfully"),
	)
}

func (h *OvertimeHandler) GetOvertimeEntryByID(ctx *gin.Context) {
	overtimeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid overtime entry id", ""))
		return
	}

	item, err := h.service.GetOvertimeEntryByID(ctx.Request.Context(), overtimeEntryID)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry retrieved successfully"),
	)
}

func (h *OvertimeHandler) GetMyOvertimeEntryByID(ctx *gin.Context) {
	overtimeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid overtime entry id", ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.GetMyOvertimeEntryByID(
		ctx.Request.Context(),
		employeeID,
		overtimeEntryID,
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toOvertimeEntryResponse(item), "Overtime entry retrieved successfully"),
	)
}

func (h *OvertimeHandler) ListOvertimeEntries(ctx *gin.Context) {
	var req listOvertimeEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListOvertimeEntries(
		ctx.Request.Context(),
		toListOvertimeEntriesParams(req),
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toOvertimeEntryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Overtime entries retrieved successfully"))
}

func (h *OvertimeHandler) ListMyOvertimeEntries(ctx *gin.Context) {
	var req listMyOvertimeEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	page, err := h.service.ListMyOvertimeEntries(
		ctx.Request.Context(),
		toListMyOvertimeEntriesParams(employeeID, req),
	)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toOvertimeEntryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "My overtime entries retrieved successfully"))
}

func (h *OvertimeHandler) GetOvertimeStats(ctx *gin.Context) {
	stats, err := h.service.GetCurrentMonthOvertimeStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toOvertimeStatsResponse(stats),
			"Overtime stats retrieved successfully",
		),
	)
}

func (h *OvertimeHandler) GetMyOvertimeStats(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	stats, err := h.service.GetMyCurrentMonthOvertimeStats(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(mapOvertimeErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toOvertimeStatsResponse(stats),
			"Overtime stats retrieved successfully",
		),
	)
}

func mapOvertimeErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrOvertimeInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrOvertimeNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrOvertimeForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrOvertimeStateInvalid):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
