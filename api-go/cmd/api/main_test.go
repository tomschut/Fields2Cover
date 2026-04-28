// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/logging"
	"github.com/Fields2Cover/fields2cover/api-go/internal/server"
)

// stubF2CClient is the minimum implementation needed to wire main's
// buildRouter + readyz probe in a unit test. The internal/server
// package has its own (more complete) mockF2CClient — we don't import
// that here because cmd/api should not depend on server's test
// helpers.
type stubF2CClient struct {
	cloneRobotErr error
}

func (s *stubF2CClient) ParseGeoJSON(context.Context, *f2cclient.ParseGeoJSONRequest, ...grpc.CallOption) (*f2cclient.ParseGeoJSONResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) TransformToUTM(context.Context, *f2cclient.TransformToUTMRequest, ...grpc.CallOption) (*f2cclient.TransformToUTMResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) CloneField(context.Context, *f2cclient.CloneFieldRequest, ...grpc.CallOption) (*f2cclient.CloneFieldResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) CloneRobot(_ context.Context, in *f2cclient.CloneRobotRequest, _ ...grpc.CallOption) (*f2cclient.CloneRobotResponse, error) {
	if s.cloneRobotErr != nil {
		return nil, s.cloneRobotErr
	}
	return &f2cclient.CloneRobotResponse{Robot: in.Robot}, nil
}
func (s *stubF2CClient) GenerateHeadland(context.Context, *f2cclient.GenerateHeadlandRequest, ...grpc.CallOption) (*f2cclient.GenerateHeadlandResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) GenerateSwaths(context.Context, *f2cclient.GenerateSwathsRequest, ...grpc.CallOption) (*f2cclient.GenerateSwathsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) SortSwaths(context.Context, *f2cclient.SortSwathsRequest, ...grpc.CallOption) (*f2cclient.SortSwathsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) GenerateRoute(context.Context, *f2cclient.GenerateRouteRequest, ...grpc.CallOption) (*f2cclient.GenerateRouteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
func (s *stubF2CClient) PlanPath(context.Context, *f2cclient.PlanPathRequest, ...grpc.CallOption) (*f2cclient.PlanPathResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

var _ f2cclient.F2CClient = (*stubF2CClient)(nil)

func newTestRouter(t *testing.T, client f2cclient.F2CClient) http.Handler {
	t.Helper()
	var buf bytes.Buffer
	logger := logging.NewJSONLogger(&buf, "json", "error")
	srv := server.NewServer(client)
	return buildRouter(logger, srv, client, "/nonexistent/openapi.yaml")
}

func TestBuildRouter_Healthz(t *testing.T) {
	t.Parallel()
	router := newTestRouter(t, &stubF2CClient{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestBuildRouter_Readyz_NoShim(t *testing.T) {
	t.Parallel()
	client := &stubF2CClient{
		cloneRobotErr: status.Error(codes.Unavailable, "shim not connected"),
	}
	router := newTestRouter(t, client)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"not_ready"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestBuildRouter_Readyz_OK(t *testing.T) {
	t.Parallel()
	router := newTestRouter(t, &stubF2CClient{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"ready"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Don't run in parallel — mutates env.
	t.Setenv("API_PORT", "")
	t.Setenv("F2C_GRPC_SOCKET", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	cfg := loadConfig()
	if cfg.port != "8080" {
		t.Errorf("port = %q, want 8080", cfg.port)
	}
	if cfg.socketPath != "/tmp/f2c.sock" {
		t.Errorf("socketPath = %q, want /tmp/f2c.sock", cfg.socketPath)
	}
	if cfg.shutdownTimeout.Seconds() != 10 {
		t.Errorf("shutdownTimeout = %v, want 10s", cfg.shutdownTimeout)
	}
}

func TestLoadConfig_Overrides(t *testing.T) {
	t.Setenv("API_PORT", "9090")
	t.Setenv("F2C_GRPC_SOCKET", "/var/run/f2c.sock")
	t.Setenv("SHUTDOWN_TIMEOUT", "30")
	cfg := loadConfig()
	if cfg.port != "9090" {
		t.Errorf("port = %q, want 9090", cfg.port)
	}
	if cfg.socketPath != "/var/run/f2c.sock" {
		t.Errorf("socketPath = %q", cfg.socketPath)
	}
	if cfg.shutdownTimeout.Seconds() != 30 {
		t.Errorf("shutdownTimeout = %v, want 30s", cfg.shutdownTimeout)
	}
}
