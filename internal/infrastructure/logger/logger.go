package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a structured logger. In production you might use JSON handler.
func New(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Keep output clean: use short level names
			if a.Key == slog.LevelKey {
				if l, ok := a.Value.Any().(slog.Level); ok {
					switch l {
					case slog.LevelDebug:
						return slog.String(slog.LevelKey, "DEBUG")
					case slog.LevelInfo:
						return slog.String(slog.LevelKey, "INFO")
					case slog.LevelWarn:
						return slog.String(slog.LevelKey, "WARN")
					case slog.LevelError:
						return slog.String(slog.LevelKey, "ERROR")
					}
				}
			}
			return a
		},
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

// LevelFromString returns slog.Level for "debug", "info", "warn", "error" (default info).
func LevelFromString(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithEventID returns a logger with event_id attribute.
func WithEventID(log *slog.Logger, eventID string) *slog.Logger {
	return log.With("event_id", eventID)
}
