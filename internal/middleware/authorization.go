package middleware

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/domain/permission"
	"hrbackend/internal/httpapi"
	db "hrbackend/internal/repository/db"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PermissionMiddleware struct {
	queries db.Querier
	logger  domain.Logger
}

func NewPermissionMiddleware(queries db.Querier, logger domain.Logger) *PermissionMiddleware {
	return &PermissionMiddleware{
		queries: queries,
		logger:  logger,
	}
}

func (m *PermissionMiddleware) Require(p permission.Permission) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, ok := AuthPayloadFromContext(ctx.Request.Context())
		if !ok || payload == nil {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpapi.Fail(domain.ErrUnauthorized.Error(), ""),
			)
			return
		}

		allowed, err := m.queries.CheckUserPermission(
			ctx.Request.Context(),
			db.CheckUserPermissionParams{
				UserID: payload.UserID,
				Name:   string(p),
			},
		)
		if err != nil {
			if m.logger != nil {
				m.logger.LogError(
					ctx.Request.Context(),
					"PermissionMiddleware.Require",
					"failed to verify permission",
					err,
					zap.String("permission", string(p)),
					zap.String("user_id", payload.UserID.String()),
				)
			}
			ctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				httpapi.Fail("failed to verify permission", ""),
			)
			return
		}

		if !allowed {
			if m.logger != nil {
				m.logger.LogWarn(
					ctx.Request.Context(),
					"PermissionMiddleware.Require",
					"permission denied",
					zap.String("permission", string(p)),
					zap.String("user_id", payload.UserID.String()),
				)
			}
			ctx.AbortWithStatusJSON(
				http.StatusForbidden,
				httpapi.Fail(errors.New("forbidden").Error(), ""),
			)
			return
		}

		ctx.Next()
	}
}
