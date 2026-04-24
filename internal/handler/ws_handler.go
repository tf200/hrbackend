package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"
	"hrbackend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const wsTicketQueryKey = "ticket"

type issueWebSocketTicketResponse struct {
	Ticket    string    `json:"ticket"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WebSocketHandler struct {
	service  domain.WebSocketAuthService
	hub      *ws.Hub
	logger   domain.Logger
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(
	service domain.WebSocketAuthService,
	hub *ws.Hub,
	logger domain.Logger,
	allowedOrigins string,
) *WebSocketHandler {
	return &WebSocketHandler{
		service: service,
		hub:     hub,
		logger:  logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     originChecker(allowedOrigins),
		},
	}
}

func RegisterWebSocketRoutes(rg *gin.RouterGroup, handler *WebSocketHandler, auth gin.HandlerFunc) {
	rg.POST("/ws/tickets", auth, handler.IssueTicket)
	rg.GET("/ws", handler.Connect)
}

func (h *WebSocketHandler) IssueTicket(ctx *gin.Context) {
	if h.service == nil {
		ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail("websocket ticket manager unavailable", ""))
		return
	}

	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("authorization payload not found", ""))
		return
	}

	result, err := h.service.IssueTicket(ctx.Request.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrExpiredWebSocketTicket):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrWebSocketTicketUnavailable):
			ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to issue websocket ticket", ""))
		}
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			issueWebSocketTicketResponse{Ticket: result.Ticket, ExpiresAt: result.ExpiresAt},
			"websocket ticket issued successfully",
		),
	)
}

func (h *WebSocketHandler) Connect(ctx *gin.Context) {
	if h.service == nil {
		ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail("websocket ticket manager unavailable", ""))
		return
	}
	if h.hub == nil {
		ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail("websocket hub unavailable", ""))
		return
	}
	if !h.upgrader.CheckOrigin(ctx.Request) {
		ctx.JSON(http.StatusForbidden, httpapi.Fail("websocket origin not allowed", ""))
		return
	}

	ticket := ctx.Query(wsTicketQueryKey)
	payload, err := h.service.ConsumeTicket(ctx.Request.Context(), ticket)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidWebSocketTicket),
			errors.Is(err, domain.ErrExpiredWebSocketTicket):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrWebSocketTicketUnavailable):
			ctx.JSON(http.StatusServiceUnavailable, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to authenticate websocket", ""))
		}
		return
	}

	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		if h.logger != nil {
			h.logger.LogWarn(
				ctx.Request.Context(),
				"WebSocketHandler.Connect",
				"failed to upgrade websocket connection",
				zap.Error(err),
				zap.String("user_id", payload.UserID.String()),
			)
		}
		return
	}

	client := ws.NewClient(h.hub, payload.UserID, conn)
	h.hub.Register(client)
	client.Start()
}

func originChecker(allowedOrigins string) func(r *http.Request) bool {
	allowed := make(map[string]struct{})
	allowAll := false

	for _, origin := range strings.Split(allowedOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAll = true
			continue
		}
		allowed[origin] = struct{}{}
	}

	return func(r *http.Request) bool {
		if allowAll || len(allowed) == 0 {
			return true
		}

		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true
		}

		_, ok := allowed[origin]
		return ok
	}
}
