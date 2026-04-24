package app

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"hrbackend/config"
	"hrbackend/internal/domain"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

const wsTicketKeyPrefix = "ws_ticket:"

type redisWebSocketTicketStore struct {
	client    *redis.Client
	keyPrefix string
}

func newWebSocketTicketStore(cfg config.Config) domain.WebSocketTicketStore {
	if cfg.RedisHost == "" {
		return nil
	}

	var tlsConfig *tls.Config
	if cfg.Remote {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(&redis.Options{
		Addr:      cfg.RedisHost,
		Password:  cfg.RedisPassword,
		TLSConfig: tlsConfig,
	})

	return &redisWebSocketTicketStore{client: client, keyPrefix: wsTicketKeyPrefix}
}

func (s *redisWebSocketTicketStore) Issue(
	ctx context.Context,
	payload domain.WebSocketTicketPayload,
	ttl time.Duration,
) (string, error) {
	if s == nil || s.client == nil {
		return "", domain.ErrWebSocketTicketUnavailable
	}
	if ttl <= 0 {
		return "", domain.ErrExpiredWebSocketTicket
	}

	recordBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal websocket ticket payload: %w", err)
	}

	for range 3 {
		ticket, err := generateWebSocketTicket(32)
		if err != nil {
			return "", fmt.Errorf("generate websocket ticket: %w", err)
		}

		stored, err := s.client.SetNX(ctx, s.keyPrefix+ticket, recordBytes, ttl).Result()
		if err != nil {
			return "", fmt.Errorf("store websocket ticket: %w", err)
		}
		if stored {
			return ticket, nil
		}
	}

	return "", fmt.Errorf("failed to store websocket ticket")
}

func (s *redisWebSocketTicketStore) Consume(
	ctx context.Context,
	ticket string,
) (*domain.WebSocketTicketPayload, error) {
	if s == nil || s.client == nil {
		return nil, domain.ErrWebSocketTicketUnavailable
	}

	recordRaw, err := s.client.GetDel(ctx, s.keyPrefix+ticket).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrInvalidWebSocketTicket
		}
		return nil, fmt.Errorf("consume websocket ticket: %w", err)
	}

	var payload domain.WebSocketTicketPayload
	if err := json.Unmarshal(recordRaw, &payload); err != nil {
		return nil, domain.ErrInvalidWebSocketTicket
	}

	if time.Now().After(payload.ExpiresAt) {
		return nil, domain.ErrExpiredWebSocketTicket
	}

	return &payload, nil
}

func (s *redisWebSocketTicketStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func generateWebSocketTicket(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
