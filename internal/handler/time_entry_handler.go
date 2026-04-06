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

type TimeEntryHandler struct {
	service domain.TimeEntryService
}

func NewTimeEntryHandler(service domain.TimeEntryService) *TimeEntryHandler {
	return &TimeEntryHandler{service: service}
}

func (h *TimeEntryHandler) CreateTimeEntry(ctx *gin.Context) {
	var req createTimeEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateTimeEntryParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateTimeEntry(ctx.Request.Context(), employeeID, params)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toTimeEntryResponse(item), "Time entry created successfully"),
	)
}

func (h *TimeEntryHandler) CreateTimeEntryByAdmin(ctx *gin.Context) {
	var req createTimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, err := toCreateTimeEntryByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.CreateTimeEntryByAdmin(ctx.Request.Context(), adminEmployeeID, params)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toTimeEntryResponse(item), "Time entry created successfully"),
	)
}

func (h *TimeEntryHandler) DecideTimeEntryByAdmin(ctx *gin.Context) {
	timeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid time entry id", ""))
		return
	}

	var req decideTimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.DecideTimeEntryByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		timeEntryID,
		toDecideTimeEntryParams(req),
	)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toTimeEntryResponse(item), "Time entry decided successfully"),
	)
}

func (h *TimeEntryHandler) UpdateTimeEntryByAdmin(ctx *gin.Context) {
	timeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid time entry id", ""))
		return
	}

	var req updateTimeEntryByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	adminEmployeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if adminEmployeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	params, adminUpdateNote, err := toUpdateTimeEntryByAdminParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	item, err := h.service.UpdateTimeEntryByAdmin(
		ctx.Request.Context(),
		adminEmployeeID,
		timeEntryID,
		params,
		adminUpdateNote,
	)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toTimeEntryResponse(item), "Time entry updated successfully"),
	)
}

func (h *TimeEntryHandler) GetTimeEntryByID(ctx *gin.Context) {
	timeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid time entry id", ""))
		return
	}

	item, err := h.service.GetTimeEntryByID(ctx.Request.Context(), timeEntryID)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toTimeEntryResponse(item), "Time entry retrieved successfully"),
	)
}

func (h *TimeEntryHandler) GetMyTimeEntryByID(ctx *gin.Context) {
	timeEntryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid time entry id", ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.GetMyTimeEntryByID(ctx.Request.Context(), employeeID, timeEntryID)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toTimeEntryResponse(item), "Time entry retrieved successfully"),
	)
}

func (h *TimeEntryHandler) ListTimeEntries(ctx *gin.Context) {
	var req listTimeEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListTimeEntries(ctx.Request.Context(), toListTimeEntriesParams(req))
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toTimeEntryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Time entries retrieved successfully"))
}

func (h *TimeEntryHandler) ListMyTimeEntries(ctx *gin.Context) {
	var req listMyTimeEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	page, err := h.service.ListMyTimeEntries(
		ctx.Request.Context(),
		toListMyTimeEntriesParams(employeeID, req),
	)
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	response := httpapi.NewPageResponse(
		ctx,
		req.PageRequest,
		toTimeEntryResponses(page.Items),
		page.TotalCount,
	)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Time entries retrieved successfully"))
}

func (h *TimeEntryHandler) GetTimeEntryStats(ctx *gin.Context) {
	stats, err := h.service.GetCurrentMonthTimeEntryStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapTimeEntryErrorStatus(err), httpapi.Fail(err.Error(), ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toTimeEntryStatsResponse(stats),
			"Time entry stats retrieved successfully",
		),
	)
}

func mapTimeEntryErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrTimeEntryInvalidRequest):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTimeEntryNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrTimeEntryForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrTimeEntryStateInvalid):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
