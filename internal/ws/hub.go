package ws

import (
	"context"
	"sync"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type UserMessage struct {
	UserID  uuid.UUID
	Message []byte
}

type Hub struct {
	clients map[uuid.UUID]map[*Client]bool
	logger  domain.Logger

	register   chan *Client
	unregister chan *Client
	sendToUser chan *UserMessage

	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func NewHub(logger domain.Logger) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		logger:     logger,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sendToUser: make(chan *UserMessage),
		shutdown:   make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			userClients, ok := h.clients[client.userID]
			if !ok {
				userClients = make(map[*Client]bool)
				h.clients[client.userID] = userClients
			}
			userClients[client] = true
			h.logInfo(
				"WebSocketHub.Run",
				"client registered",
				zap.String("user_id", client.userID.String()),
				zap.Int("connections", len(userClients)),
			)

		case client := <-h.unregister:
			userClients, ok := h.clients[client.userID]
			if ok {
				if _, clientExists := userClients[client]; clientExists {
					close(client.send)
					delete(userClients, client)
					h.logInfo(
						"WebSocketHub.Run",
						"client unregistered",
						zap.String("user_id", client.userID.String()),
						zap.Int("connections", len(userClients)),
					)
					if len(userClients) == 0 {
						delete(h.clients, client.userID)
						h.logInfo(
							"WebSocketHub.Run",
							"user has no active websocket connections",
							zap.String("user_id", client.userID.String()),
						)
					}
				}
			}

		case userMessage := <-h.sendToUser:
			userClients, ok := h.clients[userMessage.UserID]
			if ok {
				activeClients := 0
				for client := range userClients {
					select {
					case client.send <- userMessage.Message:
						activeClients++
					default:
						h.logWarn(
							"WebSocketHub.Run",
							"client send buffer full; forcing unregister",
							zap.String("user_id", client.userID.String()),
						)
						close(client.send)
						delete(userClients, client)
						if len(userClients) == 0 {
							delete(h.clients, client.userID)
							h.logInfo(
								"WebSocketHub.Run",
								"user has no active websocket connections after forced unregister",
								zap.String("user_id", client.userID.String()),
							)
						}
					}
				}
				if activeClients == 0 && len(userClients) > 0 {
					h.logWarn(
						"WebSocketHub.Run",
						"no active clients could receive websocket message",
						zap.String("user_id", userMessage.UserID.String()),
						zap.Int("connections", len(userClients)),
					)
				} else {
					h.logInfo(
						"WebSocketHub.Run",
						"message sent to active websocket connections",
						zap.String("user_id", userMessage.UserID.String()),
						zap.Int("active_connections", activeClients),
					)
				}
			}

		case <-h.shutdown:
			h.logInfo("WebSocketHub.Run", "hub shutting down")
			for userID, userClients := range h.clients {
				h.logInfo(
					"WebSocketHub.Run",
					"closing user websocket connections",
					zap.String("user_id", userID.String()),
					zap.Int("connections", len(userClients)),
				)
				for client := range userClients {
					close(client.send)
					_ = client.conn.WriteMessage(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(
							websocket.CloseGoingAway,
							"Server shutting down",
						),
					)
					_ = client.conn.Close()
				}
				delete(h.clients, userID)
			}
			h.clients = make(map[uuid.UUID]map[*Client]bool)
			return
		}
	}
}

func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
		h.logInfo("WebSocketHub.Register", "client queued for registration", zap.String("user_id", client.userID.String()))
	default:
		h.logWarn("WebSocketHub.Register", "hub register channel blocked; closing client", zap.String("user_id", client.userID.String()))
		_ = client.conn.Close()
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	msg := &UserMessage{
		UserID:  userID,
		Message: message,
	}
	select {
	case h.sendToUser <- msg:
	default:
		h.logWarn("WebSocketHub.SendToUser", "hub send channel blocked; message dropped", zap.String("user_id", userID.String()))
	}
}

func (h *Hub) Shutdown() {
	h.shutdownOnce.Do(func() {
		h.logInfo("WebSocketHub.Shutdown", "signaling hub shutdown")
		close(h.shutdown)
	})
}

func (h *Hub) logInfo(operation, message string, fields ...zap.Field) {
	if h.logger != nil {
		h.logger.LogInfo(context.Background(), operation, message, fields...)
	}
}

func (h *Hub) logWarn(operation, message string, fields ...zap.Field) {
	if h.logger != nil {
		h.logger.LogWarn(context.Background(), operation, message, fields...)
	}
}
