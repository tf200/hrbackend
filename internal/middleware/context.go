package middleware

import (
	"context"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
)

type contextKey string

const (
	authPayloadKey       contextKey = "authorization_payload"
	actorRolesKey        contextKey = "actor_roles"
	employeeIDContextKey contextKey = "employee_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return domain.WithRequestID(ctx, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	return domain.RequestIDFromContext(ctx)
}

func WithAuthPayload(ctx context.Context, payload *domain.TokenPayload) context.Context {
	return context.WithValue(ctx, authPayloadKey, payload)
}

func AuthPayloadFromContext(ctx context.Context) (*domain.TokenPayload, bool) {
	value, ok := ctx.Value(authPayloadKey).(*domain.TokenPayload)
	return value, ok
}

func WithActorRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, actorRolesKey, roles)
}

func ActorRolesFromContext(ctx context.Context) ([]string, bool) {
	value, ok := ctx.Value(actorRolesKey).([]string)
	return value, ok
}

func WithEmployeeID(ctx context.Context, employeeID uuid.UUID) context.Context {
	return context.WithValue(ctx, employeeIDContextKey, employeeID)
}

func EmployeeIDFromContext(ctx context.Context) uuid.UUID {
	if value, ok := ctx.Value(employeeIDContextKey).(uuid.UUID); ok {
		return value
	}

	payload, ok := AuthPayloadFromContext(ctx)
	if !ok || payload == nil {
		return uuid.Nil
	}

	return payload.EmployeeID
}
