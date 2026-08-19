// Package logger wraps uber-go/zap to provide structured, levelled logging.
// All services obtain their *zap.Logger from here — never call zap.New directly.
//
// Usage:
//
//	log, sync := logger.New("user-service", "info")
//	defer sync()
//	log.Info("server started", zap.String("port", ":8080"))
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a production-ready zap logger.
// serviceName is added as a constant field so every log line identifies its origin.
// level is one of: debug, info, warn, error.
// Returns the logger and a flush function that must be deferred by the caller.
func New(serviceName, level string) (*zap.Logger, func()) {
	zapLevel := parseLevel(level)

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	log, err := cfg.Build(
		zap.Fields(zap.String("service", serviceName)),
		zap.AddCallerSkip(0),
	)
	if err != nil {
		panic(fmt.Sprintf("logger: failed to initialize: %v", err))
	}

	sync := func() {
		// Sync flushes any buffered log entries. Error is intentionally ignored
		// because some environments (e.g. Docker stdout) don't support Sync.
		_ = log.Sync()
	}

	return log, sync
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
