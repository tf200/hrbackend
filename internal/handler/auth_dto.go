package handler

import (
	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"testemail@gmail.com"`
	Password string `json:"password" binding:"required" example:"t2aha000"`
}

type loginResponse struct {
	RefreshToken  string `json:"refresh" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	AccessToken   string `json:"access" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RequiresTwoFA bool   `json:"requires_2fa" example:"false"`
	TempToken     string `json:"temp_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type logoutResponse struct {
	SessionID uuid.UUID `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type verify2FARequest struct {
	Code      string `json:"validation_code" binding:"required"`
	TempToken string `json:"temp_token" binding:"required"`
}

type setup2FARequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
}

type setup2FAResponse struct {
	QRCode string `json:"qr_code_base64" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
	Secret string `json:"secret" example:"JBSWY3DPEHPK3PXP"`
}

type enable2FARequest struct {
	ValidationCode string `json:"validation_code" binding:"required"`
}

type enable2FAResponse struct {
	RecoveryCodes []string `json:"recovery_codes" example:"[\"code1\", \"code2\"]"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Mappers

func toLoginParams(req loginRequest) domain.LoginParams {
	return domain.LoginParams{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toLoginResponse(result *domain.LoginResult) loginResponse {
	if result == nil {
		return loginResponse{}
	}

	return loginResponse{
		RefreshToken:  result.RefreshToken,
		AccessToken:   result.AccessToken,
		RequiresTwoFA: result.RequiresTwoFA,
		TempToken:     result.TempToken,
	}
}

func toRefreshTokenParams(req refreshTokenRequest) domain.RefreshTokenParams {
	return domain.RefreshTokenParams{
		RefreshToken: req.RefreshToken,
	}
}

func toRefreshTokenResponse(result *domain.RefreshTokenResult) refreshTokenResponse {
	if result == nil {
		return refreshTokenResponse{}
	}

	return refreshTokenResponse{
		AccessToken: result.AccessToken,
	}
}

func toLogoutParams(payload *domain.TokenPayload) domain.LogoutParams {
	sessionID := payload.SessionID
	if sessionID == uuid.Nil {
		sessionID = payload.ID
	}

	return domain.LogoutParams{SessionID: sessionID}
}

func toLogoutResponse(payload *domain.TokenPayload) logoutResponse {
	if payload == nil {
		return logoutResponse{}
	}

	return logoutResponse{SessionID: toLogoutParams(payload).SessionID}
}

func toSetup2FAResponse(result *domain.Setup2FAResponse) setup2FAResponse {
	if result == nil {
		return setup2FAResponse{}
	}

	return setup2FAResponse{
		QRCode: result.QRCode,
		Secret: result.Secret,
	}
}

func toEnable2FAResponse(result *domain.Enable2FAResponse) enable2FAResponse {
	if result == nil {
		return enable2FAResponse{}
	}

	return enable2FAResponse{
		RecoveryCodes: result.RecoveryCodes,
	}
}
