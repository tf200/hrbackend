package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const minSecretKeySize = 32

type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
	TwoFAToken   TokenType = "2fa_token"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Payload struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	UserID     uuid.UUID
	EmployeeID uuid.UUID
	TokenType  TokenType
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

type Maker struct {
	accessTokenKey  string
	refreshTokenKey string
	twoFATokenKey   string
}

type claims struct {
	ID         uuid.UUID `json:"id"`
	SessionID  uuid.UUID `json:"session_id"`
	UserID     uuid.UUID `json:"user_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	TokenType  TokenType `json:"token_type"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (c *claims) Valid() error {
	if time.Now().After(c.ExpiresAt) {
		return ErrExpiredToken
	}
	return nil
}

func New(accessTokenKey, refreshTokenKey, twoFATokenKey string) (*Maker, error) {
	if len(accessTokenKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	if len(refreshTokenKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	if len(twoFATokenKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &Maker{
		accessTokenKey:  accessTokenKey,
		refreshTokenKey: refreshTokenKey,
		twoFATokenKey:   twoFATokenKey,
	}, nil
}

func (m *Maker) CreateToken(
	userID, employeeID uuid.UUID,
	duration time.Duration,
	tokenType TokenType,
) (string, *Payload, error) {
	return m.CreateTokenWithSessionID(userID, employeeID, duration, tokenType, uuid.Nil)
}

func (m *Maker) CreateTokenWithSessionID(
	userID, employeeID uuid.UUID,
	duration time.Duration,
	tokenType TokenType,
	sessionID uuid.UUID,
) (string, *Payload, error) {
	payload, err := newPayload(userID, employeeID, duration, tokenType)
	if err != nil {
		return "", nil, err
	}

	if sessionID != uuid.Nil {
		payload.SessionID = sessionID
	}

	secretKey, err := m.secretKeyForType(tokenType)
	if err != nil {
		return "", nil, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payloadToClaims(payload))
	token, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create token: %w", err)
	}

	return token, payload, nil
}

func (m *Maker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}

		claims, ok := token.Claims.(*claims)
		if !ok {
			return nil, ErrInvalidToken
		}

		secretKey, err := m.secretKeyForType(claims.TokenType)
		if err != nil {
			return nil, err
		}
		return []byte(secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &claims{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	parsedClaims, ok := jwtToken.Claims.(*claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &Payload{
		ID:         parsedClaims.ID,
		SessionID:  parsedClaims.SessionID,
		UserID:     parsedClaims.UserID,
		EmployeeID: parsedClaims.EmployeeID,
		TokenType:  parsedClaims.TokenType,
		IssuedAt:   parsedClaims.IssuedAt,
		ExpiresAt:  parsedClaims.ExpiresAt,
	}, nil
}

func (m *Maker) secretKeyForType(tokenType TokenType) (string, error) {
	switch tokenType {
	case AccessToken:
		return m.accessTokenKey, nil
	case RefreshToken:
		return m.refreshTokenKey, nil
	case TwoFAToken:
		return m.twoFATokenKey, nil
	default:
		return "", fmt.Errorf("unknown token type: %v", tokenType)
	}
}

func newPayload(
	userID, employeeID uuid.UUID,
	duration time.Duration,
	tokenType TokenType,
) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Payload{
		ID:         tokenID,
		SessionID:  tokenID,
		UserID:     userID,
		EmployeeID: employeeID,
		TokenType:  tokenType,
		IssuedAt:   now,
		ExpiresAt:  now.Add(duration),
	}, nil
}

func payloadToClaims(payload *Payload) *claims {
	return &claims{
		ID:         payload.ID,
		SessionID:  payload.SessionID,
		UserID:     payload.UserID,
		EmployeeID: payload.EmployeeID,
		TokenType:  payload.TokenType,
		IssuedAt:   payload.IssuedAt,
		ExpiresAt:  payload.ExpiresAt,
	}
}
