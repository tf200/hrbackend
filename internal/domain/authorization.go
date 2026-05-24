package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAuthorizationNotFound = errors.New("authorization not found")

type Authorization struct {
	ID             uuid.UUID
	Name           string
	Description    *string
	Category       string
	RequiresExpiry bool
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateAuthorizationParams struct {
	Name           string
	Description    *string
	Category       string
	RequiresExpiry bool
}

type AuthorizationRepository interface {
	ListAuthorizations(ctx context.Context) ([]Authorization, error)
	CreateAuthorization(ctx context.Context, params CreateAuthorizationParams) (*Authorization, error)
	UpdateAuthorization(ctx context.Context, id uuid.UUID, params CreateAuthorizationParams) (*Authorization, error)
}

type AuthorizationService interface {
	ListAuthorizations(ctx context.Context) ([]Authorization, error)
	CreateAuthorization(ctx context.Context, params CreateAuthorizationParams) (*Authorization, error)
	UpdateAuthorization(ctx context.Context, id uuid.UUID, params CreateAuthorizationParams) (*Authorization, error)
}
