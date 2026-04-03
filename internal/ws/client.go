package ws

import (
	"bytes"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub *Hub

	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(hub *Hub, userID uuid.UUID, conn *websocket.Conn) *Client {
	return &Client{
		hub:    hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("Client %d disconnected (readPump exit)", c.userID)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("error reading message for user %d: %v", c.userID, err)
			} else {
				log.Printf("websocket closed for user %d: %v", c.userID, err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		log.Printf("Received message from user %d: %s", c.userID, message)
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("Client %d disconnected (writePump exit)", c.userID)
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("Hub closed channel for user %d. Closing connection.", c.userID)
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting next writer for user %d: %v", c.userID, err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log.Printf("Error writing message for user %d: %v", c.userID, err)
				_ = w.Close()
				return
			}

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for user %d: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error writing ping for user %d: %v", c.userID, err)
				return
			}
		}
	}
}

func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
	log.Printf("Client pumps started for user %d", c.userID)
}
