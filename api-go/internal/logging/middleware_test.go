// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLogger_GeneratesRequestID(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	base := NewJSONLogger(&buf, "json", "info")

	mw := RequestLogger(base)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Errorf("X-Request-Id header missing")
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if !strings.Contains(buf.String(), "http_request") {
		t.Errorf("log output missing http_request line: %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"status":418`) {
		t.Errorf("log output missing status=418: %s", buf.String())
	}
}

func TestRequestLogger_ReusesInboundRequestID(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	base := NewJSONLogger(&buf, "json", "info")

	mw := RequestLogger(base)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := LoggerFromContext(r.Context())
		if l == nil {
			t.Error("LoggerFromContext returned nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("X-Request-Id", "fixed-id-123")
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got != "fixed-id-123" {
		t.Errorf("X-Request-Id = %q, want %q", got, "fixed-id-123")
	}

	var entry map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			if entry["request_id"] == "fixed-id-123" {
				return
			}
		}
	}
	t.Errorf("no log line carried request_id=fixed-id-123, buf=%s", buf.String())
}

func TestNewJSONLogger_LevelParsing(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, "json", "warn")
	l.Info("should be filtered")
	l.Warn("should appear")
	if strings.Contains(buf.String(), "should be filtered") {
		t.Errorf("info line leaked through warn filter: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("warn line missing: %s", buf.String())
	}
}

func TestNewJSONLogger_TextFormat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, "text", "info")
	l.Info("hello")
	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Errorf("text logger output = %q", out)
	}
	// Text handler does not produce JSON braces around the message.
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Errorf("expected text format, got JSON: %q", out)
	}
}

func TestLoggerFromContext_Fallback(t *testing.T) {
	t.Parallel()
	if got := LoggerFromContext(context.Background()); got == nil {
		t.Error("LoggerFromContext returned nil, want non-nil fallback")
	}
	if got := LoggerFromContext(nil); got == nil { //nolint:staticcheck
		t.Error("LoggerFromContext(nil) returned nil, want non-nil fallback")
	}
	_ = slog.Default()
}

func TestRequestLogger_NilBaseFallsBack(t *testing.T) {
	t.Parallel()
	mw := RequestLogger(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}
