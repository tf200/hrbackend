package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *zap.Logger
}

func Setup(environment string) (domain.Logger, error) {
	var config zap.Config
	if environment == "production" {
		config = zap.NewProductionConfig()
		config.DisableCaller = true
		config.DisableStacktrace = true
		config.OutputPaths = []string{"stderr"}
	} else {
		config = zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"stderr"}
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	if environment == "production" {
		logDir := "/var/log/maicare"
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		allLogsWriter := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "app.log"),
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		}

		errorLogsWriter := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "error.log"),
			MaxSize:    50,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		}

		allLogsCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(allLogsWriter),
			zap.InfoLevel,
		)

		errorLogsCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(errorLogsWriter),
			zap.ErrorLevel,
		)

		zapLogger = zap.New(zapcore.NewTee(
			zapLogger.Core(),
			allLogsCore,
			errorLogsCore,
		))
	}

	return &Logger{logger: zapLogger}, nil
}

func (l *Logger) Sync() error {
	if l == nil || l.logger == nil {
		return nil
	}
	return l.logger.Sync()
}

func (l *Logger) LogError(ctx context.Context, operation, message string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	l.log(ctx, zap.ErrorLevel, operation, message, fields...)
}

func (l *Logger) LogWarn(ctx context.Context, operation, message string, fields ...zap.Field) {
	l.log(ctx, zap.WarnLevel, operation, message, fields...)
}

func (l *Logger) LogInfo(ctx context.Context, operation, message string, fields ...zap.Field) {
	l.log(ctx, zap.InfoLevel, operation, message, fields...)
}

func (l *Logger) log(ctx context.Context, level zapcore.Level, operation, message string, fields ...zap.Field) {
	requestID := "unknown"
	if v, ok := ctx.Value("request_id").(string); ok {
		requestID = v
	}

	commonFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("operation", operation),
		zap.String("service", "maicare-api"),
		zap.Int64("timestamp", time.Now().Unix()),
	}

	allFields := append(commonFields, fields...)

	switch level {
	case zap.ErrorLevel:
		l.logger.Error(message, allFields...)
	case zap.WarnLevel:
		l.logger.Warn(message, allFields...)
	default:
		l.logger.Info(message, allFields...)
	}
}
