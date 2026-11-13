package logger

import (
	"io"
	"log/slog"
	"os"
)

func New(service string, output io.Writer) *slog.Logger {
	if output == nil {
		output = os.Stdout
	}

	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler).With(
		"service", service,
	)
}
