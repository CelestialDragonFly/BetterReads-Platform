package log

import (
	"context"
	"io"
	"log"
	"log/slog"
)

type Logger struct {
	Logger *slog.Logger
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{
		Logger: slog.New(slog.NewJSONHandler(w, nil)),
	}
}

func (l *Logger) NewErrorLogger() *log.Logger {
	return slog.NewLogLogger(l.Logger.Handler(), slog.LevelError)
}

type Fields map[string]any

func getAttr(fields Fields) []slog.Attr {
	attr := make([]slog.Attr, len(fields))
	for k, v := range fields {
		attr = append(attr, slog.Any(k, v))
	}
	return attr
}

func (l *Logger) Info(msg string, fields Fields) {
	l.Logger.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		msg,
		getAttr(fields)...)
}

func (l *Logger) Warn(msg string, fields Fields) {
	l.Logger.LogAttrs(
		context.Background(),
		slog.LevelWarn,
		msg,
		getAttr(fields)...)
}

func (l *Logger) Error(msg string, fields Fields) {
	l.Logger.LogAttrs(
		context.Background(),
		slog.LevelError,
		msg,
		getAttr(fields)...)
}
