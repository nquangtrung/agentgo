package fsm

import (
	"log/slog"
)

// Package-private logger initialized with package-specific context
var logger = slog.Default().With(
	slog.String("package", "fsm"),
)

// SetLogger allows the main application to inject a custom logger
func SetLogger(l *slog.Logger) {
	logger = l.With(slog.String("package", "fsm"))
}
