package middleware

import (
	"net/url"
	"time"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RequestLoggingMiddleware struct {
	logger      domain.Logger
	environment string
}

func NewRequestLoggingMiddleware(logger domain.Logger, environment string) *RequestLoggingMiddleware {
	return &RequestLoggingMiddleware{
		logger:      logger,
		environment: environment,
	}
}

func (m *RequestLoggingMiddleware) Handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		path := ctx.Request.URL.Path
		rawQuery := ctx.Request.URL.RawQuery

		ctx.Next()

		if m.logger == nil {
			return
		}

		statusCode := ctx.Writer.Status()
		duration := time.Since(startTime)
		requestID, _ := RequestIDFromContext(ctx.Request.Context())

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", duration),
			zap.Int64("latency_ms", duration.Milliseconds()),
			zap.String("client_ip", ctx.ClientIP()),
			zap.Int("response_size", ctx.Writer.Size()),
		}

		if rawQuery != "" && m.environment != "production" {
			fields = append(fields, zap.String("query", sanitizeQuery(rawQuery)))
		}

		if userID, exists := ctx.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				fields = append(fields, zap.String("user_id", userIDStr))
			}
		}

		switch requestLogLevel(statusCode) {
		case zapcore.ErrorLevel:
			m.logger.LogError(ctx.Request.Context(), "RequestLoggingMiddleware", "HTTP request", nil, fields...)
		case zapcore.WarnLevel:
			m.logger.LogWarn(ctx.Request.Context(), "RequestLoggingMiddleware", "HTTP request", fields...)
		default:
			if shouldLogRequest(m.environment, statusCode) {
				m.logger.LogInfo(ctx.Request.Context(), "RequestLoggingMiddleware", "HTTP request", fields...)
			}
		}

		if duration > time.Second {
			m.logger.LogWarn(ctx.Request.Context(), "RequestLoggingMiddleware", "slow HTTP request",
				append(fields, zap.Bool("slow_request", true))...)
		}
	}
}

func shouldLogRequest(environment string, statusCode int) bool {
	if environment == "production" {
		return statusCode >= 400
	}
	return true
}

func requestLogLevel(statusCode int) zapcore.Level {
	switch {
	case statusCode >= 500:
		return zapcore.ErrorLevel
	case statusCode >= 400:
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}

func sanitizeQuery(raw string) string {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return ""
	}

	for _, key := range []string{"ticket", "access_token", "token"} {
		if _, exists := values[key]; exists {
			values.Set(key, "[REDACTED]")
		}
	}

	return values.Encode()
}
