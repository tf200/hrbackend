package handler

import (
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type authorizationResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	Category       string    `json:"category"`
	RequiresExpiry bool      `json:"requires_expiry"`
	CreatedAt      time.Time `json:"created_at"`
}

func toAuthorizationResponse(a *domain.Authorization) authorizationResponse {
	return authorizationResponse{
		ID:             a.ID,
		Name:           a.Name,
		Description:    a.Description,
		Category:       a.Category,
		RequiresExpiry: a.RequiresExpiry,
		CreatedAt:      a.CreatedAt,
	}
}

type createAuthorizationRequest struct {
	Name           string  `json:"name"           binding:"required"`
	Category       string  `json:"category"       binding:"required"`
	Description    *string `json:"description"`
	RequiresExpiry bool    `json:"requires_expiry"`
}
