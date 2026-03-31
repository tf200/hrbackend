package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/pkg/password"
	"hrbackend/pkg/twofa"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService struct {
	repository           domain.AuthRepository
	tokenMaker           domain.TokenMaker
	logger               domain.Logger
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	twoFATokenDuration   time.Duration

	mu            sync.Mutex
	loginAttempts map[string]attemptState
	twoFAAttempts map[string]attemptState
}

func NewAuthService(
	repository domain.AuthRepository,
	tokenMaker domain.TokenMaker,
	logger domain.Logger,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	twoFATokenDuration time.Duration,
) domain.AuthService {
	return &AuthService{
		repository:           repository,
		tokenMaker:           tokenMaker,
		logger:               logger,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		twoFATokenDuration:   twoFATokenDuration,
		loginAttempts:        make(map[string]attemptState),
		twoFAAttempts:        make(map[string]attemptState),
	}
}

type attemptState struct {
	Count       int
	FirstFailed time.Time
	LockedUntil time.Time
}

func (s *AuthService) Login(ctx context.Context, params domain.LoginParams, clientIP string, userAgent string) (*domain.LoginResult, error) {
	email := strings.ToLower(params.Email)
	now := time.Now()
	emailKey := loginAttemptKeyForEmail(email)
	ipKey := loginAttemptKeyForIP(clientIP)

	if err := s.checkAttemptAllowed(s.loginAttempts, emailKey, now); err != nil {
		s.logger.LogWarn(ctx, "Login", "login blocked due to too many attempts", zap.String("email", email))
		return nil, domain.ErrTooManyAttempts
	}
	if err := s.checkAttemptAllowed(s.loginAttempts, ipKey, now); err != nil {
		s.logger.LogWarn(ctx, "Login", "login blocked due to too many attempts", zap.String("client_ip", clientIP))
		return nil, domain.ErrTooManyAttempts
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.recordAttemptFailure(s.loginAttempts, emailKey, now)
			s.recordAttemptFailure(s.loginAttempts, ipKey, now)
			s.logger.LogWarn(ctx, "Login", "failed login attempt: user not found",
				zap.String("email", email), zap.String("client_ip", clientIP), zap.String("user_agent", userAgent))
			return nil, domain.ErrInvalidCredentials
		}
		s.logger.LogError(ctx, "Login", "database error during login", err, zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := password.CheckPassword(params.Password, user.Password); err != nil {
		s.recordAttemptFailure(s.loginAttempts, emailKey, now)
		s.recordAttemptFailure(s.loginAttempts, ipKey, now)
		s.logger.LogWarn(ctx, "Login", "failed login attempt: incorrect password",
			zap.String("email", email), zap.String("client_ip", clientIP), zap.String("user_agent", userAgent))
		return nil, domain.ErrInvalidCredentials
	}

	s.clearAttemptState(s.loginAttempts, emailKey)
	s.clearAttemptState(s.loginAttempts, ipKey)

	if user.TwoFactorEnabled {
		tempToken, tempPayload, err := s.tokenMaker.CreateToken(user.ID, user.EmployeeID, s.twoFATokenDuration, domain.TwoFATokenType)
		if err != nil {
			s.logger.LogError(ctx, "Login", "failed to create 2FA token", err, zap.String("email", email))
			return nil, fmt.Errorf("failed to create 2FA token: %w", err)
		}

		_, err = s.repository.CreateSession(ctx, domain.CreateSessionParams{
			ID:           tempPayload.ID,
			RefreshToken: tempToken,
			UserAgent:    userAgent,
			ClientIP:     clientIP,
			IsBlocked:    false,
			ExpiresAt:    tempPayload.ExpiresAt,
			CreatedAt:    time.Now(),
			UserID:       tempPayload.UserID,
		})
		if err != nil {
			s.logger.LogError(ctx, "Login", "failed to create temporary 2FA challenge", err, zap.String("email", email))
			return nil, fmt.Errorf("failed to create temporary 2FA challenge: %w", err)
		}

		s.logger.LogInfo(ctx, "Login", "2FA required for user",
			zap.String("email", email), zap.String("client_ip", clientIP), zap.String("user_agent", userAgent))

		return &domain.LoginResult{
			RequiresTwoFA: true,
			TempToken:     tempToken,
		}, nil
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.EmployeeID, s.refreshTokenDuration, domain.RefreshTokenType)
	if err != nil {
		s.logger.LogError(ctx, "Login", "failed to create refresh token", err, zap.String("email", email))
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	accessToken, _, err := s.tokenMaker.CreateTokenWithSessionID(user.ID, user.EmployeeID, s.accessTokenDuration, domain.AccessTokenType, refreshPayload.ID)
	if err != nil {
		s.logger.LogError(ctx, "Login", "failed to create access token", err, zap.String("email", email))
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	session, err := s.repository.CreateSession(ctx, domain.CreateSessionParams{
		ID:           refreshPayload.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
		CreatedAt:    time.Now(),
		UserID:       refreshPayload.UserID,
	})
	if err != nil {
		s.logger.LogError(ctx, "Login", "database error during session creation", err, zap.String("email", email))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.LogInfo(ctx, "Login", "user logged in successfully",
		zap.String("email", email), zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent), zap.String("session_id", session.ID.String()))

	return &domain.LoginResult{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		RequiresTwoFA: false,
		TempToken:     "",
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, params domain.RefreshTokenParams) (*domain.RefreshTokenResult, error) {
	payload, err := s.tokenMaker.VerifyToken(params.RefreshToken)
	if err != nil {
		s.logger.LogWarn(ctx, "RefreshToken", "invalid refresh token", zap.Error(err))
		return nil, domain.ErrInvalidCredentials
	}

	if payload.TokenType != domain.RefreshTokenType {
		s.logger.LogWarn(ctx, "RefreshToken", "token type is not refresh token",
			zap.String("user_id", payload.UserID.String()))
		return nil, domain.ErrUnauthorized
	}

	session, err := s.repository.GetSessionByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.logger.LogWarn(ctx, "RefreshToken", "session not found",
				zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
			return nil, domain.ErrSessionNotFound
		}
		s.logger.LogError(ctx, "RefreshToken", "database error during session retrieval", err,
			zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.IsBlocked {
		s.logger.LogWarn(ctx, "RefreshToken", "blocked session attempt",
			zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if session.UserID != payload.UserID {
		s.logger.LogWarn(ctx, "RefreshToken", "session user mismatch",
			zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if subtle.ConstantTimeCompare([]byte(session.RefreshToken), []byte(params.RefreshToken)) != 1 {
		s.logger.LogWarn(ctx, "RefreshToken", "refresh token mismatch",
			zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) {
		s.logger.LogWarn(ctx, "RefreshToken", "expired session attempt",
			zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	accessToken, _, err := s.tokenMaker.CreateTokenWithSessionID(
		payload.UserID,
		payload.EmployeeID,
		s.accessTokenDuration,
		domain.AccessTokenType,
		payload.ID,
	)
	if err != nil {
		s.logger.LogError(ctx, "RefreshToken", "failed to create access token", err,
			zap.String("user_id", payload.UserID.String()))
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	s.logger.LogInfo(ctx, "RefreshToken", "access token refreshed successfully",
		zap.String("user_id", payload.UserID.String()), zap.String("session_id", payload.ID.String()))

	return &domain.RefreshTokenResult{AccessToken: accessToken}, nil
}

func (s *AuthService) Logout(ctx context.Context, params domain.LogoutParams) error {
	if params.SessionID == uuid.Nil {
		return domain.ErrUnauthorized
	}

	if err := s.repository.DeleteSession(ctx, params.SessionID); err != nil {
		s.logger.LogError(ctx, "Logout", "database error during session deletion", err,
			zap.String("session_id", params.SessionID.String()))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.LogInfo(ctx, "Logout", "user logged out successfully",
		zap.String("session_id", params.SessionID.String()))

	return nil
}

func (s *AuthService) Verify2FA(ctx context.Context, code string, tempToken string, clientIP string, userAgent string) (*domain.LoginResult, error) {
	now := time.Now()
	ipKey := loginAttemptKeyForIP(clientIP)
	emailKey := "" // will be set after user fetch

	// Rate limit by IP
	if err := s.checkAttemptAllowed(s.twoFAAttempts, ipKey, now); err != nil {
		s.logger.LogWarn(ctx, "Verify2FA", "2FA verification blocked due to too many attempts", zap.String("client_ip", clientIP))
		return nil, domain.ErrTooManyAttempts
	}

	// Verify temp token
	payload, err := s.tokenMaker.VerifyToken(tempToken)
	if err != nil {
		s.recordAttemptFailure(s.twoFAAttempts, ipKey, now)
		s.logger.LogWarn(ctx, "Verify2FA", "invalid temp token", zap.Error(err))
		return nil, domain.ErrInvalidCredentials
	}

	if payload.TokenType != domain.TwoFATokenType {
		s.logger.LogWarn(ctx, "Verify2FA", "token type is not 2fa token", zap.String("user_id", payload.UserID.String()))
		return nil, domain.ErrUnauthorized
	}

	// Get session
	tempSession, err := s.repository.GetSessionByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.logger.LogWarn(ctx, "Verify2FA", "session not found", zap.String("session_id", payload.ID.String()))
			return nil, domain.ErrSessionNotFound
		}
		s.logger.LogError(ctx, "Verify2FA", "database error during session retrieval", err, zap.String("session_id", payload.ID.String()))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if tempSession.IsBlocked {
		s.logger.LogWarn(ctx, "Verify2FA", "blocked session attempt", zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if tempSession.UserID != payload.UserID {
		s.logger.LogWarn(ctx, "Verify2FA", "session user mismatch", zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(tempSession.ExpiresAt) {
		s.logger.LogWarn(ctx, "Verify2FA", "expired session attempt", zap.String("session_id", payload.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	// Get user
	user, err := s.repository.GetUserByID(ctx, payload.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.LogWarn(ctx, "Verify2FA", "user not found", zap.String("user_id", payload.UserID.String()))
			return nil, domain.ErrUserNotFound
		}
		s.logger.LogError(ctx, "Verify2FA", "database error during user retrieval", err, zap.String("user_id", payload.UserID.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	emailKey = loginAttemptKeyForEmail(user.Email)

	// Validate 2FA code
	if user.TwoFactorSecret == nil || *user.TwoFactorSecret == "" {
		s.logger.LogWarn(ctx, "Verify2FA", "2FA secret not set", zap.String("user_id", user.ID.String()))
		return nil, domain.ErrUnauthorized
	}

	if !twofa.ValidateCode(*user.TwoFactorSecret, code) {
		s.recordAttemptFailure(s.twoFAAttempts, ipKey, now)
		s.recordAttemptFailure(s.twoFAAttempts, emailKey, now)
		s.logger.LogWarn(ctx, "Verify2FA", "invalid 2FA code", zap.String("user_id", user.ID.String()))
		return nil, domain.ErrInvalidTwoFACode
	}

	// Clear 2FA attempts
	s.clearAttemptState(s.twoFAAttempts, ipKey)
	s.clearAttemptState(s.twoFAAttempts, emailKey)

	// Delete temp session (the 2fa session)
	if err := s.repository.DeleteSession(ctx, payload.ID); err != nil {
		s.logger.LogError(ctx, "Verify2FA", "failed to delete temp session", err, zap.String("session_id", payload.ID.String()))
		// continue anyway
	}

	// Create new access and refresh tokens
	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.EmployeeID, s.refreshTokenDuration, domain.RefreshTokenType)
	if err != nil {
		s.logger.LogError(ctx, "Verify2FA", "failed to create refresh token", err, zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	accessToken, _, err := s.tokenMaker.CreateTokenWithSessionID(user.ID, user.EmployeeID, s.accessTokenDuration, domain.AccessTokenType, refreshPayload.ID)
	if err != nil {
		s.logger.LogError(ctx, "Verify2FA", "failed to create access token", err, zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create final session
	session, err := s.repository.CreateSession(ctx, domain.CreateSessionParams{
		ID:           refreshPayload.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
		CreatedAt:    time.Now(),
		UserID:       refreshPayload.UserID,
	})
	if err != nil {
		s.logger.LogError(ctx, "Verify2FA", "database error during session creation", err, zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.LogInfo(ctx, "Verify2FA", "2FA verification successful",
		zap.String("user_id", user.ID.String()), zap.String("session_id", session.ID.String()),
		zap.String("client_ip", clientIP), zap.String("user_agent", userAgent))

	return &domain.LoginResult{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		RequiresTwoFA: false,
		TempToken:     "",
	}, nil
}

func (s *AuthService) Setup2FA(ctx context.Context, userID uuid.UUID, currentPassword string) (*domain.Setup2FAResponse, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.LogWarn(ctx, "Setup2FA", "user not found for 2FA setup", zap.String("user_id", userID.String()))
			return nil, domain.ErrUserNotFound
		}
		s.logger.LogError(ctx, "Setup2FA", "database error during user retrieval", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		s.logger.LogWarn(ctx, "Setup2FA", "inactive user attempting 2FA setup", zap.String("user_id", userID.String()))
		return nil, domain.ErrUnauthorized
	}

	if user.TwoFactorEnabled {
		s.logger.LogWarn(ctx, "Setup2FA", "2FA already enabled for user", zap.String("user_id", userID.String()))
		return nil, domain.ErrTwoFaAlreadyEnabled
	}

	if err := password.CheckPassword(currentPassword, user.Password); err != nil {
		s.logger.LogWarn(ctx, "Setup2FA", "invalid password for 2FA setup", zap.String("user_id", userID.String()))
		return nil, domain.ErrInvalidCredentials
	}

	secret, err := twofa.GenerateOTPSecret("Maicare", user.Email)
	if err != nil {
		s.logger.LogError(ctx, "Setup2FA", "error generating 2FA secret", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	qrCodeBase64, err := twofa.GenerateQRCode(fmt.Sprintf("otpauth://totp/Maicare:%s?secret=%s&issuer=Maicare", user.Email, secret))
	if err != nil {
		s.logger.LogError(ctx, "Setup2FA", "error generating QR code", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Store temporary secret
	secretPtr := &secret
	_, err = s.repository.CreateTemp2FaSecret(ctx, user.ID, secretPtr)
	if err != nil {
		s.logger.LogError(ctx, "Setup2FA", "database error saving temp 2FA secret", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to save temp 2FA secret: %w", err)
	}

	s.logger.LogInfo(ctx, "Setup2FA", "2FA setup initiated", zap.String("user_id", userID.String()))

	return &domain.Setup2FAResponse{
		QRCode: qrCodeBase64,
		Secret: secret,
	}, nil
}

func (s *AuthService) Enable2FA(ctx context.Context, userID uuid.UUID, code string) (*domain.Enable2FAResponse, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.LogWarn(ctx, "Enable2FA", "user not found for 2FA enable", zap.String("user_id", userID.String()))
			return nil, domain.ErrUserNotFound
		}
		s.logger.LogError(ctx, "Enable2FA", "database error during user retrieval", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.TwoFactorEnabled {
		s.logger.LogWarn(ctx, "Enable2FA", "2FA already enabled for user", zap.String("user_id", userID.String()))
		return nil, domain.ErrTwoFaAlreadyEnabled
	}

	if user.TwoFactorSecretTemp == nil || *user.TwoFactorSecretTemp == "" {
		s.logger.LogWarn(ctx, "Enable2FA", "no temp 2FA secret found for user", zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("no temp 2FA secret found")
	}

	if !twofa.ValidateCode(*user.TwoFactorSecretTemp, code) {
		s.logger.LogWarn(ctx, "Enable2FA", "invalid 2FA validation code", zap.String("user_id", userID.String()))
		return nil, domain.ErrInvalidTwoFACode
	}

	recoveryCodes := twofa.GenerateRecoveryCodes(10)

	_, err = s.repository.Enable2Fa(ctx, user.ID, user.TwoFactorSecretTemp, recoveryCodes)
	if err != nil {
		s.logger.LogError(ctx, "Enable2FA", "database error enabling 2FA", err, zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to enable 2FA: %w", err)
	}

	s.logger.LogInfo(ctx, "Enable2FA", "2FA enabled successfully", zap.String("user_id", userID.String()))

	return &domain.Enable2FAResponse{
		RecoveryCodes: recoveryCodes,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.LogWarn(ctx, "ChangePassword", "user not found", zap.String("user_id", userID.String()))
			return domain.ErrUserNotFound
		}
		s.logger.LogError(ctx, "ChangePassword", "database error during user retrieval", err, zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		s.logger.LogWarn(ctx, "ChangePassword", "inactive user attempting password change", zap.String("user_id", userID.String()))
		return domain.ErrUnauthorized
	}

	if err := password.CheckPassword(oldPassword, user.Password); err != nil {
		s.logger.LogWarn(ctx, "ChangePassword", "invalid old password", zap.String("user_id", userID.String()))
		return domain.ErrInvalidCredentials
	}

	hashedNewPassword, err := password.HashPassword(newPassword)
	if err != nil {
		s.logger.LogError(ctx, "ChangePassword", "failed to hash new password", err, zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.repository.UpdatePassword(ctx, user.ID, hashedNewPassword); err != nil {
		s.logger.LogError(ctx, "ChangePassword", "database error updating password", err, zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.LogInfo(ctx, "ChangePassword", "password changed successfully", zap.String("user_id", userID.String()))

	return nil
}

func (s *AuthService) checkAttemptAllowed(store map[string]attemptState, key string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := store[key]
	if !ok {
		return nil
	}

	if !state.LockedUntil.IsZero() && now.Before(state.LockedUntil) {
		return domain.ErrTooManyAttempts
	}
	if state.FirstFailed.IsZero() || now.Sub(state.FirstFailed) > 15*time.Minute {
		delete(store, key)
		return nil
	}
	if state.Count >= 5 {
		state.LockedUntil = now.Add(15 * time.Minute)
		store[key] = state
		return domain.ErrTooManyAttempts
	}
	return nil
}

func (s *AuthService) recordAttemptFailure(store map[string]attemptState, key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := store[key]
	if state.FirstFailed.IsZero() || now.Sub(state.FirstFailed) > 15*time.Minute {
		state = attemptState{Count: 1, FirstFailed: now}
	} else {
		state.Count++
	}
	store[key] = state
}

func (s *AuthService) clearAttemptState(store map[string]attemptState, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(store, key)
}

func loginAttemptKeyForEmail(email string) string {
	return "email:" + strings.TrimSpace(strings.ToLower(email))
}

func loginAttemptKeyForIP(ip string) string {
	return "ip:" + strings.TrimSpace(ip)
}

var _ domain.AuthService = (*AuthService)(nil)
