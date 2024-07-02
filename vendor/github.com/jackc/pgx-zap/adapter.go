// Package zap provides a logger that writes to a go.uber.org/zap.Logger.
package zap

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{logger: logger.WithOptions(zap.AddCallerSkip(1))}
}

func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, len(data))
	i := 0
	for k, v := range data {
		fields[i] = zap.Any(k, v)
		i++
	}

	switch level {
	case tracelog.LogLevelTrace:
		pl.logger.Debug(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case tracelog.LogLevelDebug:
		pl.logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		pl.logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		pl.logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		pl.logger.Error(msg, fields...)
	default:
		pl.logger.Error(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}
