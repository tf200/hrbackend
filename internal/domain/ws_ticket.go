package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidWebSocketTicket     = errors.New("invalid websocket ticket")
	ErrExpiredWebSocketTicket     = errors.New("expired websocket ticket")
	ErrWebSocketTicketUnavailable = errors.New("websocket ticket manager unavailable")
)

type WebSocketTicketPayload struct {
	UserID     uuid.UUID
	EmployeeID uuid.UUID
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

type IssueWebSocketTicketResult struct {
	Ticket    string
	ExpiresAt time.Time
}

type WebSocketTicketStore interface {
	Issue(ctx context.Context, payload WebSocketTicketPayload, ttl time.Duration) (string, error)
	Consume(ctx context.Context, ticket string) (*WebSocketTicketPayload, error)
	Close() error
}

type WebSocketAuthService interface {
	IssueTicket(ctx context.Context, payload *TokenPayload) (*IssueWebSocketTicketResult, error)
	ConsumeTicket(ctx context.Context, ticket string) (*WebSocketTicketPayload, error)
}
