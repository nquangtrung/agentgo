package graph

import (
	"log/slog"
)

// Package-private logger initialized with package-specific context
var logger = slog.Default().With(
	slog.String("package", "graph"),
)

// SetLogger allows the main application to inject a custom logger into the graph package, enabling consistent logging across the application.
func SetLogger(l *slog.Logger) {
	logger = l.With(slog.String("package", "auth"))
}
