package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access_token"
	RefreshTokenType TokenType = "refresh_token"
	TwoFATokenType   TokenType = "2fa_token"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenPayload struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	UserID     uuid.UUID
	EmployeeID uuid.UUID
	TokenType  TokenType
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

type TokenMaker interface {
	CreateToken(userID, employeeID uuid.UUID, duration time.Duration, tokenType TokenType) (string, *TokenPayload, error)
	CreateTokenWithSessionID(userID, employeeID uuid.UUID, duration time.Duration, tokenType TokenType, sessionID uuid.UUID) (string, *TokenPayload, error)
	VerifyToken(token string) (*TokenPayload, error)
}

type TokenVerifier = TokenMaker
