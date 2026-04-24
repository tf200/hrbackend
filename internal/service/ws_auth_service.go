package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
)

type WebSocketAuthService struct {
	store  domain.WebSocketTicketStore
	logger domain.Logger
	ttl    time.Duration
}

func NewWebSocketAuthService(
	store domain.WebSocketTicketStore,
	logger domain.Logger,
	ttl time.Duration,
) domain.WebSocketAuthService {
	return &WebSocketAuthService{store: store, logger: logger, ttl: ttl}
}

func (s *WebSocketAuthService) IssueTicket(
	ctx context.Context,
	payload *domain.TokenPayload,
) (*domain.IssueWebSocketTicketResult, error) {
	if s.store == nil {
		return nil, domain.ErrWebSocketTicketUnavailable
	}
	if payload == nil {
		return nil, domain.ErrUnauthorized
	}

	now := time.Now()
	expiresAt := now.Add(s.ttl)
	if payload.ExpiresAt.Before(expiresAt) {
		expiresAt = payload.ExpiresAt
	}
	if !expiresAt.After(now) {
		return nil, domain.ErrExpiredWebSocketTicket
	}

	ticketPayload := domain.WebSocketTicketPayload{
		UserID:     payload.UserID,
		EmployeeID: payload.EmployeeID,
		IssuedAt:   now,
		ExpiresAt:  expiresAt,
	}

	ticket, err := s.store.Issue(ctx, ticketPayload, time.Until(expiresAt))
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"WebSocketAuthService.IssueTicket",
				"failed to issue websocket ticket",
				err,
				zap.String("user_id", payload.UserID.String()),
			)
		}
		return nil, fmt.Errorf("issue websocket ticket: %w", err)
	}

	if s.logger != nil {
		s.logger.LogInfo(
			ctx,
			"WebSocketAuthService.IssueTicket",
			"websocket ticket issued",
			zap.String("user_id", payload.UserID.String()),
			zap.Time("expires_at", expiresAt),
		)
	}

	return &domain.IssueWebSocketTicketResult{Ticket: ticket, ExpiresAt: expiresAt}, nil
}

func (s *WebSocketAuthService) ConsumeTicket(
	ctx context.Context,
	ticket string,
) (*domain.WebSocketTicketPayload, error) {
	if s.store == nil {
		return nil, domain.ErrWebSocketTicketUnavailable
	}
	if strings.TrimSpace(ticket) == "" {
		return nil, domain.ErrInvalidWebSocketTicket
	}

	payload, err := s.store.Consume(ctx, ticket)
	if err != nil {
		if s.logger != nil {
			s.logger.LogWarn(
				ctx,
				"WebSocketAuthService.ConsumeTicket",
				"failed to consume websocket ticket",
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.LogInfo(
			ctx,
			"WebSocketAuthService.ConsumeTicket",
			"websocket ticket consumed",
			zap.String("user_id", payload.UserID.String()),
		)
	}

	return payload, nil
}
