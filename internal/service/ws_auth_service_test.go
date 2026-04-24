package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestWebSocketAuthServiceIssueTicketCapsExpiryToEarlierOfTTLAndPayload(t *testing.T) {
	tests := []struct {
		name              string
		ttl               time.Duration
		payloadExpiryFrom time.Duration
		expectSource      string
	}{
		{
			name:              "payload expiry earlier than ttl",
			ttl:               2 * time.Minute,
			payloadExpiryFrom: 30 * time.Second,
			expectSource:      "payload",
		},
		{
			name:              "ttl earlier than payload expiry",
			ttl:               20 * time.Second,
			payloadExpiryFrom: 2 * time.Minute,
			expectSource:      "ttl",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeWebSocketTicketStore{issueTicket: "ws-ticket"}
			service := &WebSocketAuthService{store: store, ttl: tc.ttl}

			start := time.Now()
			payload := &domain.TokenPayload{
				UserID:     uuid.New(),
				EmployeeID: uuid.New(),
				ExpiresAt:  start.Add(tc.payloadExpiryFrom),
			}

			result, err := service.IssueTicket(context.Background(), payload)
			if err != nil {
				t.Fatalf("IssueTicket returned error: %v", err)
			}
			if result == nil {
				t.Fatalf("expected non-nil result")
			}
			if result.Ticket != "ws-ticket" {
				t.Fatalf("expected ticket ws-ticket, got %q", result.Ticket)
			}

			if store.issueCalled != 1 {
				t.Fatalf("expected Issue to be called once, got %d", store.issueCalled)
			}

			if tc.expectSource == "payload" {
				if !result.ExpiresAt.Equal(payload.ExpiresAt) {
					t.Fatalf("expected expires_at to match payload expiry, got %v want %v", result.ExpiresAt, payload.ExpiresAt)
				}
			} else {
				if result.ExpiresAt.After(start.Add(tc.ttl).Add(150 * time.Millisecond)) {
					t.Fatalf("expected expires_at to be capped by ttl, got %v", result.ExpiresAt)
				}
				if !result.ExpiresAt.Before(payload.ExpiresAt) {
					t.Fatalf("expected expires_at to be before payload expiry when ttl is smaller")
				}
			}

			if !store.issuePayload.ExpiresAt.Equal(result.ExpiresAt) {
				t.Fatalf("expected store payload expiry %v, got %v", result.ExpiresAt, store.issuePayload.ExpiresAt)
			}
			if store.issuePayload.UserID != payload.UserID || store.issuePayload.EmployeeID != payload.EmployeeID {
				t.Fatalf("expected user/employee IDs to be forwarded to store")
			}
			if store.issueTTL <= 0 {
				t.Fatalf("expected positive ttl passed to store, got %v", store.issueTTL)
			}
		})
	}
}

func TestWebSocketAuthServiceIssueTicketReturnsExpiredWhenPayloadAlreadyExpired(t *testing.T) {
	store := &fakeWebSocketTicketStore{issueTicket: "ws-ticket"}
	service := &WebSocketAuthService{store: store, ttl: time.Minute}

	payload := &domain.TokenPayload{
		UserID:     uuid.New(),
		EmployeeID: uuid.New(),
		ExpiresAt:  time.Now().Add(-1 * time.Second),
	}

	_, err := service.IssueTicket(context.Background(), payload)
	if !errors.Is(err, domain.ErrExpiredWebSocketTicket) {
		t.Fatalf("expected ErrExpiredWebSocketTicket, got %v", err)
	}
	if store.issueCalled != 0 {
		t.Fatalf("expected Issue not to be called, got %d", store.issueCalled)
	}
}

func TestWebSocketAuthServiceConsumeTicketRejectsBlankTicket(t *testing.T) {
	store := &fakeWebSocketTicketStore{}
	service := &WebSocketAuthService{store: store, ttl: time.Minute}

	_, err := service.ConsumeTicket(context.Background(), "   ")
	if !errors.Is(err, domain.ErrInvalidWebSocketTicket) {
		t.Fatalf("expected ErrInvalidWebSocketTicket, got %v", err)
	}
	if store.consumeCalled != 0 {
		t.Fatalf("expected Consume not to be called, got %d", store.consumeCalled)
	}
}

func TestWebSocketAuthServiceConsumeTicketDelegatesStoreSuccess(t *testing.T) {
	expected := &domain.WebSocketTicketPayload{
		UserID:     uuid.New(),
		EmployeeID: uuid.New(),
		IssuedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(time.Minute),
	}
	store := &fakeWebSocketTicketStore{consumePayload: expected}
	service := &WebSocketAuthService{store: store, ttl: time.Minute}

	result, err := service.ConsumeTicket(context.Background(), "ticket-123")
	if err != nil {
		t.Fatalf("ConsumeTicket returned error: %v", err)
	}
	if result != expected {
		t.Fatalf("expected store payload pointer to be returned")
	}
	if store.consumeCalled != 1 {
		t.Fatalf("expected Consume to be called once, got %d", store.consumeCalled)
	}
	if store.consumeTicket != "ticket-123" {
		t.Fatalf("expected ticket to be forwarded, got %q", store.consumeTicket)
	}
}

type fakeWebSocketTicketStore struct {
	issueTicket  string
	issueErr     error
	issueCalled  int
	issuePayload domain.WebSocketTicketPayload
	issueTTL     time.Duration

	consumePayload *domain.WebSocketTicketPayload
	consumeErr     error
	consumeCalled  int
	consumeTicket  string
}

func (f *fakeWebSocketTicketStore) Issue(
	_ context.Context,
	payload domain.WebSocketTicketPayload,
	ttl time.Duration,
) (string, error) {
	f.issueCalled++
	f.issuePayload = payload
	f.issueTTL = ttl

	if f.issueErr != nil {
		return "", f.issueErr
	}

	return f.issueTicket, nil
}

func (f *fakeWebSocketTicketStore) Consume(_ context.Context, ticket string) (*domain.WebSocketTicketPayload, error) {
	f.consumeCalled++
	f.consumeTicket = ticket

	if f.consumeErr != nil {
		return nil, f.consumeErr
	}

	return f.consumePayload, nil
}

func (f *fakeWebSocketTicketStore) Close() error {
	return nil
}
