// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Hand-rolled F2CClient mock used by every per-handler test file.
// No external mock framework — each method holds an optional override
// closure; if unset the default returns codes.Unimplemented.

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

type mockF2CClient struct {
	parseGeoJSON     func(ctx context.Context, req *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error)
	transformToUTM   func(ctx context.Context, req *f2cclient.TransformToUTMRequest) (*f2cclient.TransformToUTMResponse, error)
	cloneField       func(ctx context.Context, req *f2cclient.CloneFieldRequest) (*f2cclient.CloneFieldResponse, error)
	cloneRobot       func(ctx context.Context, req *f2cclient.CloneRobotRequest) (*f2cclient.CloneRobotResponse, error)
	generateHeadland func(ctx context.Context, req *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error)
	generateSwaths   func(ctx context.Context, req *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error)
	sortSwaths       func(ctx context.Context, req *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error)
	generateRoute    func(ctx context.Context, req *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error)
	planPath         func(ctx context.Context, req *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error)
}

func unimplemented(method string) error {
	return status.Errorf(codes.Unimplemented, "mock: %s not configured", method)
}

func (m *mockF2CClient) ParseGeoJSON(ctx context.Context, in *f2cclient.ParseGeoJSONRequest, _ ...grpc.CallOption) (*f2cclient.ParseGeoJSONResponse, error) {
	if m.parseGeoJSON != nil {
		return m.parseGeoJSON(ctx, in)
	}
	return nil, unimplemented("ParseGeoJSON")
}
func (m *mockF2CClient) TransformToUTM(ctx context.Context, in *f2cclient.TransformToUTMRequest, _ ...grpc.CallOption) (*f2cclient.TransformToUTMResponse, error) {
	if m.transformToUTM != nil {
		return m.transformToUTM(ctx, in)
	}
	return nil, unimplemented("TransformToUTM")
}
func (m *mockF2CClient) CloneField(ctx context.Context, in *f2cclient.CloneFieldRequest, _ ...grpc.CallOption) (*f2cclient.CloneFieldResponse, error) {
	if m.cloneField != nil {
		return m.cloneField(ctx, in)
	}
	return nil, unimplemented("CloneField")
}
func (m *mockF2CClient) CloneRobot(ctx context.Context, in *f2cclient.CloneRobotRequest, _ ...grpc.CallOption) (*f2cclient.CloneRobotResponse, error) {
	if m.cloneRobot != nil {
		return m.cloneRobot(ctx, in)
	}
	return nil, unimplemented("CloneRobot")
}
func (m *mockF2CClient) GenerateHeadland(ctx context.Context, in *f2cclient.GenerateHeadlandRequest, _ ...grpc.CallOption) (*f2cclient.GenerateHeadlandResponse, error) {
	if m.generateHeadland != nil {
		return m.generateHeadland(ctx, in)
	}
	return nil, unimplemented("GenerateHeadland")
}
func (m *mockF2CClient) GenerateSwaths(ctx context.Context, in *f2cclient.GenerateSwathsRequest, _ ...grpc.CallOption) (*f2cclient.GenerateSwathsResponse, error) {
	if m.generateSwaths != nil {
		return m.generateSwaths(ctx, in)
	}
	return nil, unimplemented("GenerateSwaths")
}
func (m *mockF2CClient) SortSwaths(ctx context.Context, in *f2cclient.SortSwathsRequest, _ ...grpc.CallOption) (*f2cclient.SortSwathsResponse, error) {
	if m.sortSwaths != nil {
		return m.sortSwaths(ctx, in)
	}
	return nil, unimplemented("SortSwaths")
}
func (m *mockF2CClient) GenerateRoute(ctx context.Context, in *f2cclient.GenerateRouteRequest, _ ...grpc.CallOption) (*f2cclient.GenerateRouteResponse, error) {
	if m.generateRoute != nil {
		return m.generateRoute(ctx, in)
	}
	return nil, unimplemented("GenerateRoute")
}
func (m *mockF2CClient) PlanPath(ctx context.Context, in *f2cclient.PlanPathRequest, _ ...grpc.CallOption) (*f2cclient.PlanPathResponse, error) {
	if m.planPath != nil {
		return m.planPath(ctx, in)
	}
	return nil, unimplemented("PlanPath")
}

// Compile-time check that the mock satisfies the interface.
var _ f2cclient.F2CClient = (*mockF2CClient)(nil)

// doRequest is a small helper that runs handler against (method, path, body) and returns the recorder.
func doRequest(t *testing.T, handler http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler(rec, req)
	return rec
}

// validFeatureCollectionJSON is a tiny but spec-valid GeoJSON
// FeatureCollection used as the geometry payload across multiple
// handler tests.
const validFeatureCollectionJSON = `{
  "type": "FeatureCollection",
  "features": [
    {"type":"Feature","properties":null,"geometry":{"type":"Polygon","coordinates":[[[4.71,52.34],[4.72,52.34],[4.72,52.35],[4.71,52.35],[4.71,52.34]]]}}
  ]
}`
