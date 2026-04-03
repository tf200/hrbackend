package middleware

import (
	"net/http"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RequestContextMiddleware struct {
	logger domain.Logger
}

func NewRequestContextMiddleware(logger domain.Logger) *RequestContextMiddleware {
	return &RequestContextMiddleware{logger: logger}
}

func (m *RequestContextMiddleware) Handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx.Header(RequestIDHeader, requestID)
		ctx.Set(string(requestIDContextKey), requestID)

		requestCtx := WithRequestID(ctx.Request.Context(), requestID)
		ctx.Request = ctx.Request.WithContext(requestCtx)

		if m.logger != nil {
			m.logger.LogInfo(
				ctx.Request.Context(),
				"RequestContextMiddleware",
				"request context initialized",
				zap.String("request_id", requestID),
				zap.String("path", ctx.Request.URL.Path),
				zap.String("method", ctx.Request.Method),
			)
		}

		ctx.Next()

		if ctx.Writer.Status() == http.StatusNotFound && m.logger != nil {
			m.logger.LogWarn(ctx.Request.Context(), "RequestContextMiddleware", "route not found",
				zap.String("request_id", requestID),
				zap.String("path", ctx.Request.URL.Path),
				zap.String("method", ctx.Request.Method),
			)
		}
	}
}
