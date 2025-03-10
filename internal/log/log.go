package log

import (
	"log/slog"
	"os"
)

func Debug(message string, fields ...any) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Debug(message, fields...)
}

func Info(message string, fields ...any) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info(message, fields...)
}

func Warn(message string, fields ...any) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Warn(message, fields...)
}

func Error(message string, fields ...any) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Error(message, fields...)
}
