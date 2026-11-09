package config

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// NewLogger builds a text slog handler from a level name (debug, info, warn, error).
func NewLogger(level string) *slog.Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	return slog.New(textHandler(lvl))
}

func textHandler(lvl slog.Level) slog.Handler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339Nano))
			}
			return a
		},
	})
}
