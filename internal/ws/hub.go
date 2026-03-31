package ws

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type UserMessage struct {
	UserID  uuid.UUID
	Message []byte
}

type Hub struct {
	clients map[uuid.UUID]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	sendToUser chan *UserMessage

	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
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
			log.Printf("Client registered via channel for user %d. Total connections for user: %d", client.userID, len(userClients))

		case client := <-h.unregister:
			userClients, ok := h.clients[client.userID]
			if ok {
				if _, clientExists := userClients[client]; clientExists {
					close(client.send)
					delete(userClients, client)
					log.Printf("Client unregistered via channel for user %d. Remaining connections for user: %d", client.userID, len(userClients))
					if len(userClients) == 0 {
						delete(h.clients, client.userID)
						log.Printf("User %d has no more connections. Removed user entry.", client.userID)
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
						log.Printf("Client send buffer full for user %d. Forcing unregister.", client.userID)
						close(client.send)
						delete(userClients, client)
						if len(userClients) == 0 {
							delete(h.clients, client.userID)
							log.Printf("User %d has no more connections after forced unregister. Removed user entry.", client.userID)
						}
					}
				}
				if activeClients == 0 && len(userClients) > 0 {
					log.Printf("Warning: No active clients could receive message for user %d, but %d clients were registered.", userMessage.UserID, len(userClients))
				} else if ok {
					log.Printf("Message sent to %d active connections for user %d", activeClients, userMessage.UserID)
				}
			}

		case <-h.shutdown:
			log.Println("Hub shutting down...")
			for userID, userClients := range h.clients {
				log.Printf("Closing %d connections for user %d", len(userClients), userID)
				for client := range userClients {
					close(client.send)
					_ = client.conn.WriteMessage(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"),
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
		log.Printf("Client for user %d queued for registration", client.userID)
	default:
		log.Printf("CRITICAL: Hub register channel blocked. Cannot register client for user %d. Closing client.", client.userID)
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
		log.Printf("Warning: Hub's sendToUser channel is blocked or Hub is not running. Message for user %d dropped.", userID)
	}
}

func (h *Hub) Shutdown() {
	h.shutdownOnce.Do(func() {
		log.Println("Signaling Hub shutdown...")
		close(h.shutdown)
	})
}
