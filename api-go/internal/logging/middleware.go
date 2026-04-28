// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package logging

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-Id"

// statusRecorder wraps http.ResponseWriter to capture the response
// status code for access logging. The default status (when WriteHeader
// is never called) is 200, matching net/http semantics.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// RequestLogger returns a chi-compatible middleware that:
//   - Generates a request_id (or reuses X-Request-Id from the inbound headers)
//   - Builds a request-scoped slog.Logger with method, path, request_id
//   - Stores the logger on the request context via WithLogger
//   - Echoes the request id back to the client via X-Request-Id
//   - Logs one structured access line per request (status, duration_ms, bytes)
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	if base == nil {
		base = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID := r.Header.Get(requestIDHeader)
			if reqID == "" {
				reqID = uuid.NewString()
			}
			w.Header().Set(requestIDHeader, reqID)

			logger := base.With(
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
			ctx := WithLogger(r.Context(), logger)

			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r.WithContext(ctx))

			if rec.status == 0 {
				rec.status = http.StatusOK
			}

			logger.LogAttrs(ctx, slog.LevelInfo, "http_request",
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.Int("bytes", rec.bytes),
			)
		})
	}
}
