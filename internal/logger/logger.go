package logger

import (
	"avito_intern/internal/config"
	"log/slog"
	"os"
)

func InitLogger(cfg *config.Config) *slog.Logger {

	var level slog.Level
	switch cfg.Logger.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)

}
