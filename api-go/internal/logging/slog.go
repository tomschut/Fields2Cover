// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package logging wires log/slog with a JSON handler and provides
// context-scoped logger propagation for the chi request pipeline.

package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

type ctxKey int

const loggerKey ctxKey = 0

// NewJSONLogger returns a *slog.Logger that writes JSON-formatted
// records to w at the requested level. format is "json" (default) or
// "text" — anything else falls back to JSON. level is "debug",
// "info", "warn", or "error" (case-insensitive); unknown values
// default to info.
func NewJSONLogger(w io.Writer, format, level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	if strings.ToLower(format) == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}
	return slog.New(h)
}

// WithLogger returns a derived context carrying the supplied logger.
// Handlers extract it via LoggerFromContext.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext returns the request-scoped logger if one was
// installed by the middleware, or slog.Default() as a fallback.
// A nil context returns the default logger.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
