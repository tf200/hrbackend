package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrSessionNotFound     = errors.New("session not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrTooManyAttempts     = errors.New("too many attempts")
	ErrTwoFaAlreadyEnabled = errors.New("two-factor authentication already enabled")
	ErrInvalidTwoFACode    = errors.New("invalid two-factor authentication code")
)

type LoginParams struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken   string
	RefreshToken  string
	RequiresTwoFA bool
	TempToken     string
}

type RefreshTokenParams struct {
	RefreshToken string
}

type RefreshTokenResult struct {
	AccessToken string
}

type LogoutParams struct {
	SessionID uuid.UUID
}

type Setup2FAResponse struct {
	QRCode string `json:"qr_code_base64"`
	Secret string `json:"secret"`
}

type Enable2FAResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type AuthUser struct {
	ID                  uuid.UUID
	Password            string
	Email               string
	IsActive            bool
	EmployeeID          uuid.UUID
	LastLogin           time.Time
	DateJoined          time.Time
	ProfilePicture      *string
	TwoFactorEnabled    bool
	TwoFactorSecret     *string
	TwoFactorSecretTemp *string
	RecoveryCodes       []string
}

type CreateSessionParams struct {
	ID           uuid.UUID
	RefreshToken string
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UserID       uuid.UUID
}

type AuthSession struct {
	ID           uuid.UUID
	RefreshToken string
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UserID       uuid.UUID
}

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*AuthUser, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*AuthUser, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (*AuthSession, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*AuthSession, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	CreateTemp2FaSecret(ctx context.Context, userID uuid.UUID, secret *string) (int64, error)
	Enable2Fa(
		ctx context.Context,
		userID uuid.UUID,
		secret *string,
		recoveryCodes []string,
	) (int64, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error
}

type AuthService interface {
	Login(
		ctx context.Context,
		params LoginParams,
		clientIP string,
		userAgent string,
	) (*LoginResult, error)
	RefreshToken(ctx context.Context, params RefreshTokenParams) (*RefreshTokenResult, error)
	Logout(ctx context.Context, params LogoutParams) error
	Verify2FA(
		ctx context.Context,
		code string,
		tempToken string,
		clientIP string,
		userAgent string,
	) (*LoginResult, error)
	Setup2FA(
		ctx context.Context,
		userID uuid.UUID,
		currentPassword string,
	) (*Setup2FAResponse, error)
	Enable2FA(ctx context.Context, userID uuid.UUID, code string) (*Enable2FAResponse, error)
	ChangePassword(
		ctx context.Context,
		userID uuid.UUID,
		oldPassword string,
		newPassword string,
	) error
}
