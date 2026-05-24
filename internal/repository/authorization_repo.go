package repository

import (
	"context"
	"errors"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthorizationRepository struct {
	store *db.Store
}

func NewAuthorizationRepository(store *db.Store) domain.AuthorizationRepository {
	return &AuthorizationRepository{store: store}
}

func (r *AuthorizationRepository) ListAuthorizations(
	ctx context.Context,
) ([]domain.Authorization, error) {
	rows, err := r.store.ListAuthorizations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Authorization, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainAuthorization(row))
	}

	return result, nil
}

func (r *AuthorizationRepository) CreateAuthorization(
	ctx context.Context,
	params domain.CreateAuthorizationParams,
) (*domain.Authorization, error) {
	row, err := r.store.CreateAuthorization(ctx, db.CreateAuthorizationParams{
		Name:           params.Name,
		Description:    params.Description,
		Category:       params.Category,
		RequiresExpiry: params.RequiresExpiry,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainAuthorization(row)
	return &result, nil
}

func (r *AuthorizationRepository) UpdateAuthorization(
	ctx context.Context,
	id uuid.UUID,
	params domain.CreateAuthorizationParams,
) (*domain.Authorization, error) {
	row, err := r.store.UpdateAuthorization(ctx, db.UpdateAuthorizationParams{
		ID:             id,
		Name:           params.Name,
		Description:    params.Description,
		Category:       params.Category,
		RequiresExpiry: params.RequiresExpiry,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAuthorizationNotFound
		}
		return nil, err
	}

	result := toDomainAuthorization(row)
	return &result, nil
}

func toDomainAuthorization(row db.Authorization) domain.Authorization {
	return domain.Authorization{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		Category:       row.Category,
		RequiresExpiry: row.RequiresExpiry,
		IsActive:       row.IsActive,
		CreatedAt:      conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:      conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
