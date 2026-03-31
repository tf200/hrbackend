package app

import (
	"time"

	"hrbackend/internal/domain"
	pkgjwt "hrbackend/pkg/jwt"

	"github.com/google/uuid"
)

type tokenMakerAdapter struct {
	maker *pkgjwt.Maker
}

func (a *tokenMakerAdapter) CreateToken(userID, employeeID uuid.UUID, duration time.Duration, tokenType domain.TokenType) (string, *domain.TokenPayload, error) {
	token, payload, err := a.maker.CreateToken(userID, employeeID, duration, toPkgTokenType(tokenType))
	if err != nil {
		return "", nil, err
	}
	return token, toDomainTokenPayload(payload), nil
}

func (a *tokenMakerAdapter) CreateTokenWithSessionID(userID, employeeID uuid.UUID, duration time.Duration, tokenType domain.TokenType, sessionID uuid.UUID) (string, *domain.TokenPayload, error) {
	token, payload, err := a.maker.CreateTokenWithSessionID(userID, employeeID, duration, toPkgTokenType(tokenType), sessionID)
	if err != nil {
		return "", nil, err
	}
	return token, toDomainTokenPayload(payload), nil
}

func (a *tokenMakerAdapter) VerifyToken(token string) (*domain.TokenPayload, error) {
	payload, err := a.maker.VerifyToken(token)
	if err != nil {
		switch err {
		case pkgjwt.ErrExpiredToken:
			return nil, domain.ErrExpiredToken
		case pkgjwt.ErrInvalidToken:
			return nil, domain.ErrInvalidToken
		default:
			return nil, err
		}
	}

	return toDomainTokenPayload(payload), nil
}

func toPkgTokenType(tokenType domain.TokenType) pkgjwt.TokenType {
	switch tokenType {
	case domain.RefreshTokenType:
		return pkgjwt.RefreshToken
	case domain.TwoFATokenType:
		return pkgjwt.TwoFAToken
	default:
		return pkgjwt.AccessToken
	}
}

func toDomainTokenPayload(payload *pkgjwt.Payload) *domain.TokenPayload {
	if payload == nil {
		return nil
	}

	return &domain.TokenPayload{
		ID:         payload.ID,
		SessionID:  payload.SessionID,
		UserID:     payload.UserID,
		EmployeeID: payload.EmployeeID,
		TokenType:  domain.TokenType(payload.TokenType),
		IssuedAt:   payload.IssuedAt,
		ExpiresAt:  payload.ExpiresAt,
	}
}
