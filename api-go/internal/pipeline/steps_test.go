// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package pipeline

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// --- Shared mock harness ----------------------------------------------

// mockF2CClient mirrors the pattern from internal/server/server_test.go:
// one optional closure per method; unset closures return Unimplemented.
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

var _ f2cclient.F2CClient = (*mockF2CClient)(nil)

// --- Fixtures (minimal valid proto / api types) ------------------------

func fxProtoField(id string) *f2cclient.Field {
	return &f2cclient.Field{
		Id:          id,
		Crs:         &f2cclient.CRS{Epsg: 32631, UtmZone: "31N"},
		GeometryWkt: []byte(`{"type":"FeatureCollection","features":[]}`),
	}
}

func fxProtoFWH(id string) *f2cclient.FieldWithHeadlands {
	return &f2cclient.FieldWithHeadlands{
		Field:          fxProtoField(id),
		HeadlandsWkt:   []byte(`{"type":"FeatureCollection","features":[]}`),
		InnerFieldWkt:  []byte(`{"type":"FeatureCollection","features":[]}`),
		HeadlandWidthM: 3.0,
	}
}

func fxProtoSwaths(fieldID string, ids ...int32) *f2cclient.Swaths {
	items := make([]*f2cclient.Swath, 0, len(ids))
	for _, id := range ids {
		items = append(items, &f2cclient.Swath{
			Id:     id,
			WidthM: 2.7,
			Type:   f2cclient.SwathType_SWATH_TYPE_MAINLAND,
		})
	}
	return &f2cclient.Swaths{
		Field: fxProtoFWH(fieldID),
		Items: items,
	}
}

func fxProtoRoute(fieldID string) *f2cclient.Route {
	return &f2cclient.Route{
		Field:  fxProtoFWH(fieldID),
		Swaths: fxProtoSwaths(fieldID, 0, 1),
	}
}

func fxProtoPath() *f2cclient.Path {
	return &f2cclient.Path{
		Field:     fxProtoFWH("fld_test_001"),
		LengthM:   42.0,
		TaskTimeS: 7.5,
	}
}

func fxAPIField(id string) api.Field {
	return api.Field{
		Id:  id,
		Crs: api.CRS{Epsg: 32631},
		Geometry: api.GeoJSONFeatureCollection{
			Type: api.FeatureCollection,
		},
	}
}

func fxAPIRobot() api.Robot {
	return api.Robot{WidthM: 3.0, CovWidthM: 2.7}
}

func fxAPIFWH(id string) api.FieldWithHeadlands {
	return api.FieldWithHeadlands{
		Field:          fxAPIField(id),
		Headlands:      api.GeoJSONFeatureCollection{Type: api.FeatureCollection},
		InnerField:     api.GeoJSONFeatureCollection{Type: api.FeatureCollection},
		HeadlandWidthM: 3.0,
	}
}

func fxAPISwaths(fieldID string, ids ...int) api.Swaths {
	items := make([]api.Swath, 0, len(ids))
	for _, id := range ids {
		var s api.Swath
		s.Id = id
		s.WidthM = 2.7
		s.Type = "MAINLAND"
		s.Path.Type = "LineString"
		s.Path.Coordinates = [][]float64{{0, 0}, {10, 0}}
		items = append(items, s)
	}
	return api.Swaths{
		Field: fxAPIFWH(fieldID),
		Items: items,
	}
}

func fxAPIRoute(fieldID string) api.Route {
	return api.Route{
		Field:       fxAPIFWH(fieldID),
		Swaths:      fxAPISwaths(fieldID, 0, 1),
		Connections: []api.RouteConnection{},
	}
}

func fxAPIGeoJSON() api.GeoJSONFeatureCollection {
	return api.GeoJSONFeatureCollection{Type: api.FeatureCollection}
}

// --- Steps unit tests --------------------------------------------------

func TestSteps_ParseField_Happy(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.ParseGeoJSONRequest
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, req *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			captured = req
			return &f2cclient.ParseGeoJSONResponse{Field: fxProtoField("fld_test_001")}, nil
		},
	}
	steps := NewSteps(mock)
	field, err := steps.ParseField(context.Background(), fxAPIGeoJSON(), nil)
	if err != nil {
		t.Fatalf("ParseField: %v", err)
	}
	if field == nil || field.Id != "fld_test_001" {
		t.Fatalf("unexpected field: %+v", field)
	}
	if captured == nil || captured.Geojson == "" {
		t.Fatalf("mock did not capture geojson request")
	}
}

func TestSteps_ParseField_GrpcError(t *testing.T) {
	t.Parallel()
	wantErr := status.Error(codes.InvalidArgument, "empty feature collection")
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return nil, wantErr
		},
	}
	steps := NewSteps(mock)
	_, err := steps.ParseField(context.Background(), fxAPIGeoJSON(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("error = %v, want gRPC InvalidArgument", err)
	}
}

func TestSteps_ParseField_EmptyResponse(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return &f2cclient.ParseGeoJSONResponse{Field: nil}, nil
		},
	}
	steps := NewSteps(mock)
	_, err := steps.ParseField(context.Background(), fxAPIGeoJSON(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var empty *EmptyResponseError
	if !errors.As(err, &empty) {
		t.Fatalf("error = %v, want *EmptyResponseError", err)
	}
	if empty.Op != "ParseGeoJSON" {
		t.Errorf("empty.Op = %q, want ParseGeoJSON", empty.Op)
	}
}

func TestSteps_ParseField_WithTargetCRS(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.ParseGeoJSONRequest
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, req *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			captured = req
			return &f2cclient.ParseGeoJSONResponse{Field: fxProtoField("fld_x")}, nil
		},
	}
	crs := api.CRS{Epsg: 32631}
	steps := NewSteps(mock)
	_, err := steps.ParseField(context.Background(), fxAPIGeoJSON(), &crs)
	if err != nil {
		t.Fatalf("ParseField: %v", err)
	}
	if captured == nil || captured.TargetCrs == nil || captured.TargetCrs.Epsg != 32631 {
		t.Fatalf("target_crs not propagated: %+v", captured)
	}
}

func TestSteps_GenerateHeadland_DefaultCount(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.GenerateHeadlandRequest
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, req *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			captured = req
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: fxProtoFWH("fld_test_001")}, nil
		},
	}
	steps := NewSteps(mock)
	_, err := steps.GenerateHeadland(context.Background(), fxAPIField("fld_test_001"), fxAPIRobot(), 3.0, 0)
	if err != nil {
		t.Fatalf("GenerateHeadland: %v", err)
	}
	if captured.Count != 3 {
		t.Errorf("Count = %d, want default 3", captured.Count)
	}
}

func TestSteps_GenerateHeadland_ExplicitCount(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.GenerateHeadlandRequest
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, req *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			captured = req
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: fxProtoFWH("fld_test_001")}, nil
		},
	}
	steps := NewSteps(mock)
	_, err := steps.GenerateHeadland(context.Background(), fxAPIField("fld_test_001"), fxAPIRobot(), 3.0, 5)
	if err != nil {
		t.Fatalf("GenerateHeadland: %v", err)
	}
	if captured.Count != 5 {
		t.Errorf("Count = %d, want 5", captured.Count)
	}
}

func TestSteps_GenerateHeadland_EmptyResponse(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			return &f2cclient.GenerateHeadlandResponse{}, nil
		},
	}
	steps := NewSteps(mock)
	_, err := steps.GenerateHeadland(context.Background(), fxAPIField("fld"), fxAPIRobot(), 3.0, 3)
	var empty *EmptyResponseError
	if !errors.As(err, &empty) {
		t.Fatalf("error = %v, want *EmptyResponseError", err)
	}
}

func TestSteps_GenerateSwaths_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateSwaths: func(_ context.Context, req *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			if req.AngleRad == nil || *req.AngleRad != 1.5 || req.WidthM != 2.7 {
				t.Errorf("request fields wrong: %+v", req)
			}
			return &f2cclient.GenerateSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0, 1)}, nil
		},
	}
	steps := NewSteps(mock)
	angle := 1.5
	out, err := steps.GenerateSwaths(context.Background(), fxAPIFWH("fld_test_001"), fxAPIRobot(), &angle, 2.7)
	if err != nil {
		t.Fatalf("GenerateSwaths: %v", err)
	}
	if out == nil || len(out.Items) != 2 {
		t.Fatalf("unexpected swaths: %+v", out)
	}
}

func TestSteps_SortSwaths_NilAlgorithm_UsesUnspecified(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.SortSwathsRequest
	mock := &mockF2CClient{
		sortSwaths: func(_ context.Context, req *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			captured = req
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0)}, nil
		},
	}
	steps := NewSteps(mock)
	_, err := steps.SortSwaths(context.Background(), fxAPISwaths("fld_test_001", 0), nil, nil, nil)
	if err != nil {
		t.Fatalf("SortSwaths: %v", err)
	}
	if captured.Algorithm != f2cclient.SortAlgorithm_SORT_ALGORITHM_UNSPECIFIED {
		t.Errorf("algorithm = %v, want UNSPECIFIED", captured.Algorithm)
	}
	if captured.Variant != 0 {
		t.Errorf("variant = %d, want 0", captured.Variant)
	}
}

func TestSteps_SortSwaths_PropagatesAlgorithmAndVariant(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.SortSwathsRequest
	mock := &mockF2CClient{
		sortSwaths: func(_ context.Context, req *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			captured = req
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld", 0)}, nil
		},
	}
	steps := NewSteps(mock)
	algo := api.SNAKE
	variant := 2
	_, err := steps.SortSwaths(context.Background(), fxAPISwaths("fld", 0), &algo, &variant, nil)
	if err != nil {
		t.Fatalf("SortSwaths: %v", err)
	}
	if captured.Algorithm != f2cclient.SortAlgorithm_SORT_ALGORITHM_SNAKE {
		t.Errorf("algorithm = %v, want SNAKE", captured.Algorithm)
	}
	if captured.Variant != 2 {
		t.Errorf("variant = %d, want 2", captured.Variant)
	}
}

func TestSteps_GenerateRoute_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
	}
	steps := NewSteps(mock)
	route, err := steps.GenerateRoute(context.Background(), fxAPISwaths("fld_test_001", 0, 1))
	if err != nil {
		t.Fatalf("GenerateRoute: %v", err)
	}
	if route == nil {
		t.Fatal("route nil")
	}
}

func TestSteps_PlanPath_Happy(t *testing.T) {
	t.Parallel()
	var captured *f2cclient.PlanPathRequest
	mock := &mockF2CClient{
		planPath: func(_ context.Context, req *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			captured = req
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	steps := NewSteps(mock)
	path, err := steps.PlanPath(context.Background(), fxAPIRoute("fld_test_001"), fxAPIRobot(), api.PlanPathRequestTurningAlgorithmDUBINS)
	if err != nil {
		t.Fatalf("PlanPath: %v", err)
	}
	if path == nil {
		t.Fatal("path nil")
	}
	if captured.TurningAlgorithm != f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS {
		t.Errorf("turning algorithm = %v, want DUBINS", captured.TurningAlgorithm)
	}
}

func TestSteps_Client_ReturnsUnderlyingClient(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{}
	steps := NewSteps(mock)
	if steps.Client() == nil {
		t.Fatal("Client() returned nil")
	}
}

func TestSortAlgorithmToProto_Table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   api.SortSwathsRequestAlgorithm
		want f2cclient.SortAlgorithm
	}{
		{api.BOUSTROPHEDON, f2cclient.SortAlgorithm_SORT_ALGORITHM_BOUSTROPHEDON},
		{api.SNAKE, f2cclient.SortAlgorithm_SORT_ALGORITHM_SNAKE},
		{api.SPIRAL, f2cclient.SortAlgorithm_SORT_ALGORITHM_SPIRAL},
		{api.SortSwathsRequestAlgorithm("GARBAGE"), f2cclient.SortAlgorithm_SORT_ALGORITHM_UNSPECIFIED},
	}
	for _, c := range cases {
		if got := SortAlgorithmToProto(c.in); got != c.want {
			t.Errorf("SortAlgorithmToProto(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestTurningAlgorithmToProto_Table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   api.PlanPathRequestTurningAlgorithm
		want f2cclient.TurningAlgorithm
	}{
		{api.PlanPathRequestTurningAlgorithmDUBINS, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS},
		{api.PlanPathRequestTurningAlgorithmDUBINSCC, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS_CC},
		{api.PlanPathRequestTurningAlgorithmREEDSSHEPP, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP},
		{api.PlanPathRequestTurningAlgorithmREEDSSHEPPHC, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP_HC},
		{api.PlanPathRequestTurningAlgorithm("GARBAGE"), f2cclient.TurningAlgorithm_TURNING_ALGORITHM_UNSPECIFIED},
	}
	for _, c := range cases {
		if got := TurningAlgorithmToProto(c.in); got != c.want {
			t.Errorf("TurningAlgorithmToProto(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
