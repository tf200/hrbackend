package handler

import (
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterNotificationRoutes(
	rg *gin.RouterGroup,
	handler *NotificationHandler,
	auth gin.HandlerFunc,
) {
	rg.GET("/notifications", auth, handler.ListNotifications)
	rg.GET("/notifications/unread-count", auth, handler.GetUnreadCount)
	rg.PATCH("/notifications/:id/read", auth, handler.MarkAsRead)
	rg.PATCH("/notifications/read-all", auth, handler.MarkAllAsRead)
}

type NotificationHandler struct {
	service domain.NotificationService
}

func NewNotificationHandler(service domain.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) ListNotifications(ctx *gin.Context) {
	userID, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || userID == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	var req struct {
		httpapi.PageRequest
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	items, totalCount, err := h.service.ListNotifications(
		ctx.Request.Context(),
		userID.UserID,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list notifications", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			httpapi.NewPageResponse(ctx, req.PageRequest, items, totalCount),
			"Notifications retrieved successfully",
		),
	)
}

func (h *NotificationHandler) GetUnreadCount(ctx *gin.Context) {
	userID, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || userID == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	count, err := h.service.GetUnreadCount(ctx.Request.Context(), userID.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get unread count", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(map[string]int64{"count": count}, "Unread count retrieved successfully"),
	)
}

func (h *NotificationHandler) MarkAsRead(ctx *gin.Context) {
	userID, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || userID == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid notification id", ""))
		return
	}

	if err := h.service.MarkAsRead(ctx.Request.Context(), id, userID.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to mark notification as read", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(struct{}{}, "Notification marked as read"))
}

func (h *NotificationHandler) MarkAllAsRead(ctx *gin.Context) {
	userID, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || userID == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	if err := h.service.MarkAllAsRead(ctx.Request.Context(), userID.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to mark all notifications as read", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(struct{}{}, "All notifications marked as read"))
}
