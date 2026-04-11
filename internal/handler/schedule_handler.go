package handler

import (
	"errors"
	"fmt"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	service domain.ScheduleService
}

func NewScheduleHandler(service domain.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) CreateSchedule(ctx *gin.Context) {
	var req createScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	schedules, err := h.service.CreateSchedule(
		ctx.Request.Context(),
		employeeID,
		toCreateScheduleRequest(req),
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to create schedule: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(schedules, "Schedules created successfully"))
}

func (h *ScheduleHandler) GetSchedulesByLocationInRange(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	var req getSchedulesByLocationInRangeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	response, err := h.service.GetSchedulesByLocationInRange(
		ctx.Request.Context(),
		locationID,
		toGetSchedulesByLocationInRangeRequest(req),
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to get schedules by range: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Schedules retrieved successfully"))
}

func (h *ScheduleHandler) GetEmployeeSchedulesByDay(ctx *gin.Context) {
	var req getEmployeeSchedulesByDayRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	domainReq, err := toGetEmployeeSchedulesByDayRequest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	response, err := h.service.GetEmployeeSchedulesByDay(
		ctx.Request.Context(),
		domainReq,
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to get employee schedules by day: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Schedules retrieved successfully"))
}

func (h *ScheduleHandler) GetEmployeeSchedulesTimeline(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req getEmployeeSchedulesTimelineRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	response, err := h.service.GetEmployeeSchedulesTimeline(
		ctx.Request.Context(),
		toGetEmployeeSchedulesTimelineRequest(req, employeeID),
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to get employee schedules timeline: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Schedules retrieved successfully"))
}

func (h *ScheduleHandler) GetScheduleByID(ctx *gin.Context) {
	scheduleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid schedule ID format", ""))
		return
	}

	item, err := h.service.GetScheduleByID(ctx.Request.Context(), scheduleID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to get schedule: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(item, "Schedule retrieved successfully"))
}

func (h *ScheduleHandler) UpdateSchedule(ctx *gin.Context) {
	scheduleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid schedule ID format", ""))
		return
	}

	var req updateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	item, err := h.service.UpdateSchedule(
		ctx.Request.Context(),
		scheduleID,
		employeeID,
		toUpdateScheduleRequest(req),
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to update schedule: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(item, "Schedule updated successfully"))
}

func (h *ScheduleHandler) DeleteSchedule(ctx *gin.Context) {
	scheduleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid schedule ID format", ""))
		return
	}

	if err := h.service.DeleteSchedule(ctx.Request.Context(), scheduleID); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to delete schedule: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK[any](nil, "Schedule deleted successfully"))
}

func (h *ScheduleHandler) AutoGenerateSchedules(ctx *gin.Context) {
	var req autoGenerateSchedulesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	generatedSchedule, err := h.service.AutoGenerateSchedules(
		ctx.Request.Context(),
		toAutoGenerateSchedulesRequest(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrWeekNotEmpty) {
			ctx.JSON(http.StatusConflict, httpapi.Fail(err.Error(), ""))
			return
		}
		if errors.Is(err, domain.ErrScheduleAutogenUnavailable) {
			ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to auto-generate schedules: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(generatedSchedule, "Schedules auto-generated successfully"))
}

func (h *ScheduleHandler) SaveGeneratedSchedules(ctx *gin.Context) {
	var req saveGeneratedSchedulesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	if err := h.service.SaveGeneratedSchedules(
		ctx.Request.Context(),
		employeeID,
		toSaveGeneratedSchedulesRequest(req),
	); err != nil {
		if errors.Is(err, domain.ErrScheduleAutogenUnavailable) {
			ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail(fmt.Sprintf("failed to save generated schedules: %v", err), ""),
		)
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK[any](nil, "Generated schedules saved successfully"))
}
