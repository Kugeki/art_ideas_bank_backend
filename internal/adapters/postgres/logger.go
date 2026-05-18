package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/tracelog"
	"log/slog"
)

type logger struct {
	log *slog.Logger
}

func newLogger(log *slog.Logger) *logger {
	return &logger{log: log}
}

func (l *logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	args := make([]any, 0, len(data))

	for k, v := range data {
		args = append(args, slog.Any(k, v))
	}

	l.log.Log(ctx, toSlogLevel(level), msg, args...)
}

func getSlogLevel(ctx context.Context, log *slog.Logger) slog.Level {
	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

	for _, level := range levels {
		if log.Enabled(ctx, level) {
			return level
		}
	}

	return DefaultSlogLevel
}

var DefaultSlogLevel = slog.LevelInfo

func toSlogLevel(level tracelog.LogLevel) slog.Level {
	switch level {
	case tracelog.LogLevelNone:
		return DefaultSlogLevel
	case tracelog.LogLevelError:
		return slog.LevelError
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelInfo:
		return slog.LevelInfo
	case tracelog.LogLevelDebug:
		return slog.LevelDebug
	case tracelog.LogLevelTrace:
		return slog.LevelDebug
	default:
		return DefaultSlogLevel
	}
}

var DefaultTraceLogLevel = tracelog.LogLevelInfo

func fromSlogLevel(level slog.Level) tracelog.LogLevel {
	switch level {
	case slog.LevelDebug:
		return tracelog.LogLevelDebug
	case slog.LevelInfo:
		return tracelog.LogLevelInfo
	case slog.LevelWarn:
		return tracelog.LogLevelWarn
	case slog.LevelError:
		return tracelog.LogLevelError
	default:
		return DefaultTraceLogLevel
	}
}
