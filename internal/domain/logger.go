package domain

import (
	"context"

	"go.uber.org/zap"
)

type Logger interface {
	LogError(ctx context.Context, operation, message string, err error, fields ...zap.Field)
	LogWarn(ctx context.Context, operation, message string, fields ...zap.Field)
	LogInfo(ctx context.Context, operation, message string, fields ...zap.Field)
}
