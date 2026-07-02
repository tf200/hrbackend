package ws

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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
		c.hub.logInfo(
			"WebSocketClient.readPump",
			"client disconnected",
			zap.String("user_id", c.userID.String()),
		)
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
				c.hub.logWarn(
					"WebSocketClient.readPump",
					"error reading websocket message",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
			} else {
				c.hub.logInfo(
					"WebSocketClient.readPump",
					"websocket closed",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
			}
			break
		}

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		c.hub.logInfo(
			"WebSocketClient.readPump",
			"websocket message received",
			zap.String("user_id", c.userID.String()),
			zap.Int("message_size", len(message)),
		)
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.hub.logInfo(
			"WebSocketClient.writePump",
			"client write pump stopped",
			zap.String("user_id", c.userID.String()),
		)
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.hub.logInfo(
					"WebSocketClient.writePump",
					"hub closed client channel",
					zap.String("user_id", c.userID.String()),
				)
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.logWarn(
					"WebSocketClient.writePump",
					"failed to get websocket writer",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				c.hub.logWarn(
					"WebSocketClient.writePump",
					"failed to write websocket message",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
				_ = w.Close()
				return
			}

			if err := w.Close(); err != nil {
				c.hub.logWarn(
					"WebSocketClient.writePump",
					"failed to close websocket writer",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.logWarn(
					"WebSocketClient.writePump",
					"failed to write websocket ping",
					zap.String("user_id", c.userID.String()),
					zap.Error(err),
				)
				return
			}
		}
	}
}

func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
	c.hub.logInfo(
		"WebSocketClient.Start",
		"client pumps started",
		zap.String("user_id", c.userID.String()),
	)
}
