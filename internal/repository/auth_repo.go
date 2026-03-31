package repository

import (
	"context"
	"database/sql"
	"errors"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
)

type AuthRepository struct {
	queries db.Querier
}

func NewAuthRepository(queries db.Querier) domain.AuthRepository {
	return &AuthRepository{queries: queries}
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.AuthUser, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return toDomainAuthUserFromEmailRow(row), nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.AuthUser, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return toDomainAuthUserFromIDRow(row), nil
}

func (r *AuthRepository) CreateTemp2FaSecret(ctx context.Context, userID uuid.UUID, secret *string) (int64, error) {
	params := db.CreateTemp2FaSecretParams{
		ID:                  userID,
		TwoFactorSecretTemp: secret,
	}
	return r.queries.CreateTemp2FaSecret(ctx, params)
}

func (r *AuthRepository) Enable2Fa(ctx context.Context, userID uuid.UUID, secret *string, recoveryCodes []string) (int64, error) {
	params := db.Enable2FaParams{
		ID:              userID,
		TwoFactorSecret: secret,
		RecoveryCodes:   recoveryCodes,
	}
	return r.queries.Enable2Fa(ctx, params)
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	params := db.UpdatePasswordParams{
		ID:       userID,
		Password: password,
	}
	return r.queries.UpdatePassword(ctx, params)
}

func (r *AuthRepository) CreateSession(ctx context.Context, params domain.CreateSessionParams) (*domain.AuthSession, error) {
	row, err := r.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:           params.ID,
		RefreshToken: params.RefreshToken,
		UserAgent:    params.UserAgent,
		ClientIp:     params.ClientIP,
		IsBlocked:    params.IsBlocked,
		ExpiresAt:    conv.PgTimestamptzFromTime(params.ExpiresAt),
		CreatedAt:    conv.PgTimestamptzFromTime(params.CreatedAt),
		UserID:       params.UserID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainAuthSession(row), nil
}

func (r *AuthRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.AuthSession, error) {
	row, err := r.queries.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	return toDomainAuthSession(row), nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteSession(ctx, id)
}

func toDomainAuthUserFromEmailRow(row db.GetUserByEmailRow) *domain.AuthUser {
	return &domain.AuthUser{
		ID:                  row.ID,
		Password:            row.Password,
		Email:               row.Email,
		IsActive:            row.IsActive,
		EmployeeID:          row.EmployeeID,
		LastLogin:           conv.TimeFromPgTimestamptz(row.LastLogin),
		DateJoined:          conv.TimeFromPgTimestamptz(row.DateJoined),
		ProfilePicture:      row.ProfilePicture,
		TwoFactorEnabled:    row.TwoFactorEnabled,
		TwoFactorSecret:     row.TwoFactorSecret,
		TwoFactorSecretTemp: row.TwoFactorSecretTemp,
		RecoveryCodes:       row.RecoveryCodes,
	}
}

func toDomainAuthUserFromIDRow(row db.GetUserByIDRow) *domain.AuthUser {
	return &domain.AuthUser{
		ID:                  row.ID,
		Password:            row.Password,
		Email:               row.Email,
		IsActive:            row.IsActive,
		EmployeeID:          row.EmployeeID,
		LastLogin:           conv.TimeFromPgTimestamptz(row.LastLogin),
		DateJoined:          conv.TimeFromPgTimestamptz(row.DateJoined),
		ProfilePicture:      row.ProfilePicture,
		TwoFactorEnabled:    row.TwoFactorEnabled,
		TwoFactorSecret:     row.TwoFactorSecret,
		TwoFactorSecretTemp: row.TwoFactorSecretTemp,
		RecoveryCodes:       row.RecoveryCodes,
	}
}

func toDomainAuthSession(row db.Session) *domain.AuthSession {
	return &domain.AuthSession{
		ID:           row.ID,
		RefreshToken: row.RefreshToken,
		UserAgent:    row.UserAgent,
		ClientIP:     row.ClientIp,
		IsBlocked:    row.IsBlocked,
		ExpiresAt:    conv.TimeFromPgTimestamptz(row.ExpiresAt),
		CreatedAt:    conv.TimeFromPgTimestamptz(row.CreatedAt),
		UserID:       row.UserID,
	}
}

var _ domain.AuthRepository = (*AuthRepository)(nil)
