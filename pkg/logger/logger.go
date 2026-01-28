package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level string `yaml:"level"`
}

func SetDefaults(cfg *Config) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	slog.SetDefault(slog.New(handler))
}
