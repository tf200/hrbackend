package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, handler *AuthHandler, auth gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	authGroup.POST("/token", handler.Login)
	authGroup.POST("/refresh", handler.RefreshToken)
	authGroup.POST("/verify_2fa", handler.Verify2FA)
	authGroup.POST("/logout", auth, handler.Logout)
	authGroup.POST("/setup_2fa", auth, handler.Setup2FA)
	authGroup.POST("/enable_2fa", auth, handler.Enable2FA)
	authGroup.POST("/change_password", auth, handler.ChangePassword)
}

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.Login(
		ctx.Request.Context(),
		toLoginParams(req),
		ctx.ClientIP(),
		ctx.Request.UserAgent(),
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTooManyAttempts):
			ctx.JSON(http.StatusTooManyRequests, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrUserNotFound),
			errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to login", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toLoginResponse(result), "login successful"))
}

func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	var req refreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.RefreshToken(ctx.Request.Context(), toRefreshTokenParams(req))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrUnauthorized),
			errors.Is(err, domain.ErrInvalidToken),
			errors.Is(err, domain.ErrExpiredToken):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to refresh token", ""))
		}
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toRefreshTokenResponse(result), "access token refreshed successfully"),
	)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("authorization payload not found", ""))
		return
	}

	if err := h.service.Logout(ctx.Request.Context(), toLogoutParams(payload)); err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to logout", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toLogoutResponse(payload), "logout successful"))
}

func (h *AuthHandler) Verify2FA(ctx *gin.Context) {
	var req verify2FARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.Verify2FA(
		ctx.Request.Context(),
		req.Code,
		req.TempToken,
		ctx.ClientIP(),
		ctx.Request.UserAgent(),
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTooManyAttempts):
			ctx.JSON(http.StatusTooManyRequests, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrInvalidTwoFACode),
			errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrSessionNotFound),
			errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to verify 2FA", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toLoginResponse(result), "2FA verification successful"))
}

func (h *AuthHandler) Setup2FA(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("authorization payload not found", ""))
		return
	}

	var req setup2FARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.Setup2FA(ctx.Request.Context(), payload.UserID, req.CurrentPassword)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrTwoFaAlreadyEnabled),
			errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to setup 2FA", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toSetup2FAResponse(result), "2FA setup successful"))
}

func (h *AuthHandler) Enable2FA(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("authorization payload not found", ""))
		return
	}

	var req enable2FARequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.Enable2FA(ctx.Request.Context(), payload.UserID, req.ValidationCode)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrTwoFaAlreadyEnabled),
			errors.Is(err, domain.ErrInvalidTwoFACode):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to enable 2FA", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEnable2FAResponse(result), "2FA enabled successfully"))
}

func (h *AuthHandler) ChangePassword(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("authorization payload not found", ""))
		return
	}

	var req changePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	err := h.service.ChangePassword(
		ctx.Request.Context(),
		payload.UserID,
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound),
			errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to change password", ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK[any](nil, "password changed successfully"))
}
