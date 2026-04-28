// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthz_AlwaysOK(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	Healthz(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestReadyz_NilProbe_ReturnsStarting(t *testing.T) {
	t.Parallel()
	h := NewReadyz(nil, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"starting"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestReadyz_ProbeOK(t *testing.T) {
	t.Parallel()
	probe := func(_ context.Context) error { return nil }
	h := NewReadyz(probe, 100*time.Millisecond)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ready"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestReadyz_ProbeError(t *testing.T) {
	t.Parallel()
	probe := func(_ context.Context) error { return errors.New("shim down") }
	h := NewReadyz(probe, 100*time.Millisecond)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"not_ready"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "shim down") {
		t.Errorf("body missing probe error: %s", rec.Body.String())
	}
}

func TestReadyz_ProbeTimeout(t *testing.T) {
	t.Parallel()
	probe := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	}
	h := NewReadyz(probe, 20*time.Millisecond)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestReadyz_ZeroTimeoutUsesDefault(t *testing.T) {
	t.Parallel()
	probe := func(_ context.Context) error { return nil }
	h := NewReadyz(probe, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
