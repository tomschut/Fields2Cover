// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// --- Shared fixture helpers ------------------------------------------

// cannedProtoField returns a minimally-valid proto Field the mock can
// hand back from ParseGeoJSON.
func cannedProtoField(id string) *f2cclient.Field {
	return &f2cclient.Field{
		Id:          id,
		Crs:         &f2cclient.CRS{Epsg: 32631, UtmZone: "31N"},
		GeometryWkt: []byte(`{"type":"FeatureCollection","features":[]}`),
	}
}

func cannedProtoFWH(id string) *f2cclient.FieldWithHeadlands {
	return &f2cclient.FieldWithHeadlands{
		Field:          cannedProtoField(id),
		HeadlandWidthM: 3.0,
		HeadlandsWkt:   []byte(`{"type":"FeatureCollection","features":[]}`),
		InnerFieldWkt:  []byte(`{"type":"FeatureCollection","features":[]}`),
	}
}

func cannedProtoSwaths(id string, ids ...int32) *f2cclient.Swaths {
	items := make([]*f2cclient.Swath, 0, len(ids))
	for _, sid := range ids {
		items = append(items, &f2cclient.Swath{
			Id:     sid,
			WidthM: 2.7,
			Type:   f2cclient.SwathType_SWATH_TYPE_MAINLAND,
		})
	}
	return &f2cclient.Swaths{
		Field: cannedProtoFWH(id),
		Items: items,
	}
}

func cannedProtoRoute(id string) *f2cclient.Route {
	return &f2cclient.Route{
		Field:  cannedProtoFWH(id),
		Swaths: cannedProtoSwaths(id),
		Connections: []*f2cclient.RouteConnection{
			{FromSwathId: 0, ToSwathId: 1, PointsWkt: []byte(`{"type":"MultiPoint","coordinates":[[10,0],[10,3]]}`)},
		},
	}
}

func cannedProtoPath(id string) *f2cclient.Path {
	return &f2cclient.Path{
		Field:     cannedProtoFWH(id),
		LengthM:   42.0,
		TaskTimeS: 18.0,
		States: []*f2cclient.PathState{
			{
				Point:       &f2cclient.Point{X: 0, Y: 0},
				AngleRad:    0,
				LengthM:     1,
				Velocity:    1.5,
				Direction:   f2cclient.PathDirection_PATH_DIRECTION_FORWARD,
				SectionType: f2cclient.PathSectionType_PATH_SECTION_TYPE_SWATH,
			},
		},
	}
}

// fullPipelineMock returns a mockF2CClient that happily answers every
// one of the six pipeline RPCs with canned data. The returned counters
// let tests assert exactly which RPCs ran.
type pipelineCallCounters struct {
	parseGeoJSON     atomic.Int32
	generateHeadland atomic.Int32
	generateSwaths   atomic.Int32
	sortSwaths       atomic.Int32
	generateRoute    atomic.Int32
	planPath         atomic.Int32
}

func fullPipelineMock(_ *testing.T, id string, counters *pipelineCallCounters) *mockF2CClient {
	return &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			counters.parseGeoJSON.Add(1)
			return &f2cclient.ParseGeoJSONResponse{Field: cannedProtoField(id)}, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			counters.generateHeadland.Add(1)
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: cannedProtoFWH(id)}, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			counters.generateSwaths.Add(1)
			return &f2cclient.GenerateSwathsResponse{Swaths: cannedProtoSwaths(id, 0, 1)}, nil
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			counters.sortSwaths.Add(1)
			return &f2cclient.SortSwathsResponse{Swaths: cannedProtoSwaths(id, 42, 43)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			counters.generateRoute.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: cannedProtoRoute(id)}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			counters.planPath.Add(1)
			return &f2cclient.PlanPathResponse{Path: cannedProtoPath(id)}, nil
		},
	}
}

// --- /pipeline/plan-coverage tests -----------------------------------

const validPlanCoverageBody = `{
  "geojson": ` + validFeatureCollectionJSON + `,
  "robot": {"width_m": 3.0, "cov_width_m": 2.7, "min_turning_radius_m": 4.5},
  "headland_width_m": 3.0,
  "headland_count": 2,
  "swath_angle_rad": 0.0,
  "swath_width_m": 2.7,
  "sort_algorithm": "BOUSTROPHEDON",
  "turning_algorithm": "DUBINS"
}`

func TestPlanCoverage_Happy(t *testing.T) {
	t.Parallel()
	var counters pipelineCallCounters
	srv := NewServer(fullPipelineMock(t, "fld_test_001", &counters))
	rec := doRequest(t, srv.PlanCoverage, http.MethodPost, "/pipeline/plan-coverage", validPlanCoverageBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.PlanCoverageResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Path.LengthM != 42.0 {
		t.Errorf("Path.LengthM = %v, want 42.0", resp.Path.LengthM)
	}
	if resp.Intermediates.Field == nil {
		t.Error("Intermediates.Field is nil")
	}
	if resp.Intermediates.FieldWithHeadlands == nil {
		t.Error("Intermediates.FieldWithHeadlands is nil")
	}
	if resp.Intermediates.Swaths == nil {
		t.Error("Intermediates.Swaths is nil")
	}
	if resp.Intermediates.SortedSwaths == nil {
		t.Error("Intermediates.SortedSwaths is nil")
	}
	if resp.Intermediates.Route == nil {
		t.Error("Intermediates.Route is nil")
	}
	// All 6 RPCs should have fired exactly once.
	if counters.parseGeoJSON.Load() != 1 || counters.generateHeadland.Load() != 1 ||
		counters.generateSwaths.Load() != 1 || counters.sortSwaths.Load() != 1 ||
		counters.generateRoute.Load() != 1 || counters.planPath.Load() != 1 {
		t.Errorf("call counts = parse:%d headland:%d swaths:%d sort:%d route:%d plan:%d (want 1 each)",
			counters.parseGeoJSON.Load(), counters.generateHeadland.Load(),
			counters.generateSwaths.Load(), counters.sortSwaths.Load(),
			counters.generateRoute.Load(), counters.planPath.Load())
	}
	if string(resp.Journey.StepName) != "paths:plan" {
		t.Errorf("Journey.StepName = %q, want paths:plan", resp.Journey.StepName)
	}
	if resp.Journey.Next != nil {
		t.Errorf("Journey.Next = %+v, want nil (terminal)", resp.Journey.Next)
	}
}

func TestPlanCoverage_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.PlanCoverage, http.MethodPost, "/pipeline/plan-coverage", "{not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_JSON") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestPlanCoverage_ParseFailurePropagatesAs400(t *testing.T) {
	t.Parallel()
	var headlandCalls atomic.Int32
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "bad geojson")
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			headlandCalls.Add(1)
			return nil, nil
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.PlanCoverage, http.MethodPost, "/pipeline/plan-coverage", validPlanCoverageBody)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bad geojson") {
		t.Errorf("body = %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "InvalidArgument") {
		t.Errorf("expected InvalidArgument code in body = %s", rec.Body.String())
	}
	if headlandCalls.Load() != 0 {
		t.Errorf("generateHeadland called %d times, want 0 (short-circuit)", headlandCalls.Load())
	}
}

// --- /pipeline/resume tests ------------------------------------------

const validSwathsBodyFragment = `{
  "field": {
    "field": {
      "id": "fld_test_001",
      "crs": {"epsg": 32631},
      "geometry": ` + validFeatureCollectionJSON + `
    },
    "headlands": ` + validFeatureCollectionJSON + `,
    "inner_field": ` + validFeatureCollectionJSON + `,
    "headland_width_m": 3.0
  },
  "items": [
    {"id": 9999, "width_m": 2.7, "type": "MAINLAND", "path": {"type":"LineString","coordinates":[[0,0],[10,0]]}}
  ]
}`

const validRouteBodyFragment = `{
  "field": {
    "field": {
      "id": "fld_test_001",
      "crs": {"epsg": 32631},
      "geometry": ` + validFeatureCollectionJSON + `
    },
    "headlands": ` + validFeatureCollectionJSON + `,
    "inner_field": ` + validFeatureCollectionJSON + `,
    "headland_width_m": 3.0
  },
  "swaths": {
    "field": {
      "field": {
        "id": "fld_test_001",
        "crs": {"epsg": 32631},
        "geometry": ` + validFeatureCollectionJSON + `
      },
      "headlands": ` + validFeatureCollectionJSON + `,
      "inner_field": ` + validFeatureCollectionJSON + `,
      "headland_width_m": 3.0
    },
    "items": []
  },
  "connections": []
}`

func TestResumePipeline_RejectsParseFieldStart(t *testing.T) {
	t.Parallel()
	var counters pipelineCallCounters
	srv := NewServer(fullPipelineMock(t, "fld_test_001", &counters))
	// start_step="parse-field" is not in the OpenAPI enum, but the handler
	// must reject it defensively even if a client passes the raw string.
	body := `{
	  "start_step": "parse-field",
	  "robot": {"width_m": 3.0, "cov_width_m": 2.7},
	  "turning_algorithm": "DUBINS"
	}`
	rec := doRequest(t, srv.ResumePipeline, http.MethodPost, "/pipeline/resume", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_START_STEP") {
		t.Errorf("body = %s", rec.Body.String())
	}
	if counters.parseGeoJSON.Load()+counters.generateHeadland.Load()+counters.generateSwaths.Load()+
		counters.sortSwaths.Load()+counters.generateRoute.Load()+counters.planPath.Load() != 0 {
		t.Errorf("expected zero RPC calls, got parse:%d headland:%d swaths:%d sort:%d route:%d plan:%d",
			counters.parseGeoJSON.Load(), counters.generateHeadland.Load(),
			counters.generateSwaths.Load(), counters.sortSwaths.Load(),
			counters.generateRoute.Load(), counters.planPath.Load())
	}
}

func TestResumePipeline_MissingPayload(t *testing.T) {
	t.Parallel()
	var counters pipelineCallCounters
	srv := NewServer(fullPipelineMock(t, "fld_test_001", &counters))
	// start_step=sort-swaths but no swaths field.
	body := `{
	  "start_step": "sort-swaths",
	  "robot": {"width_m": 3.0, "cov_width_m": 2.7},
	  "swath_width_m": 2.7,
	  "turning_algorithm": "DUBINS"
	}`
	rec := doRequest(t, srv.ResumePipeline, http.MethodPost, "/pipeline/resume", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "MISSING_PAYLOAD") {
		t.Errorf("body = %s", rec.Body.String())
	}
	if counters.sortSwaths.Load() != 0 {
		t.Errorf("sortSwaths called despite validation error")
	}
}

func TestResumePipeline_SortSwaths_SentinelPreserved(t *testing.T) {
	t.Parallel()
	var counters pipelineCallCounters
	mock := &mockF2CClient{
		// These must NOT be called for a sort-swaths resume.
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			t.Fatalf("parseGeoJSON should not be called on sort-swaths resume")
			return nil, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			t.Fatalf("generateHeadland should not be called on sort-swaths resume")
			return nil, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			t.Fatalf("generateSwaths should not be called on sort-swaths resume")
			return nil, nil
		},
		// These do run.
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			counters.sortSwaths.Add(1)
			return &f2cclient.SortSwathsResponse{Swaths: cannedProtoSwaths("fld_test_001", 1, 2)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			counters.generateRoute.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: cannedProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			counters.planPath.Add(1)
			return &f2cclient.PlanPathResponse{Path: cannedProtoPath("fld_test_001")}, nil
		},
	}
	srv := NewServer(mock)
	body := `{
	  "start_step": "sort-swaths",
	  "robot": {"width_m": 3.0, "cov_width_m": 2.7, "min_turning_radius_m": 4.5},
	  "swath_width_m": 2.7,
	  "turning_algorithm": "DUBINS",
	  "swaths": ` + validSwathsBodyFragment + `
	}`
	rec := doRequest(t, srv.ResumePipeline, http.MethodPost, "/pipeline/resume", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.PlanCoverageResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Intermediates.Swaths == nil || len(resp.Intermediates.Swaths.Items) != 1 ||
		resp.Intermediates.Swaths.Items[0].Id != 9999 {
		t.Errorf("sentinel swath id=9999 not preserved in intermediates: %+v", resp.Intermediates.Swaths)
	}
	if resp.Intermediates.Field != nil {
		t.Errorf("Intermediates.Field should be nil (parse-field did not run), got %+v", resp.Intermediates.Field)
	}
	if resp.Intermediates.FieldWithHeadlands != nil {
		t.Errorf("Intermediates.FieldWithHeadlands should be nil")
	}
	if counters.sortSwaths.Load() != 1 || counters.generateRoute.Load() != 1 || counters.planPath.Load() != 1 {
		t.Errorf("call counts sort:%d route:%d plan:%d (want 1 each)",
			counters.sortSwaths.Load(), counters.generateRoute.Load(), counters.planPath.Load())
	}
}

func TestResumePipeline_PlanPathOnly(t *testing.T) {
	t.Parallel()
	var planCalls atomic.Int32
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			t.Fatalf("parseGeoJSON should not be called")
			return nil, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			t.Fatalf("generateHeadland should not be called")
			return nil, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			t.Fatalf("generateSwaths should not be called")
			return nil, nil
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			t.Fatalf("sortSwaths should not be called")
			return nil, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			t.Fatalf("generateRoute should not be called")
			return nil, nil
		},
		planPath: func(_ context.Context, req *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			planCalls.Add(1)
			if req.TurningAlgorithm != f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS {
				t.Errorf("mock: turning algorithm = %v, want DUBINS", req.TurningAlgorithm)
			}
			return &f2cclient.PlanPathResponse{Path: cannedProtoPath("fld_test_001")}, nil
		},
	}
	srv := NewServer(mock)
	body := `{
	  "start_step": "plan-path",
	  "robot": {"width_m": 3.0, "cov_width_m": 2.7, "min_turning_radius_m": 4.5},
	  "swath_width_m": 2.7,
	  "turning_algorithm": "DUBINS",
	  "route": ` + validRouteBodyFragment + `
	}`
	rec := doRequest(t, srv.ResumePipeline, http.MethodPost, "/pipeline/resume", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if planCalls.Load() != 1 {
		t.Errorf("planPath calls = %d, want 1", planCalls.Load())
	}
	var resp api.PlanCoverageResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Path.LengthM != 42.0 {
		t.Errorf("Path.LengthM = %v, want 42.0", resp.Path.LengthM)
	}
	if resp.Intermediates.Route == nil {
		t.Error("Intermediates.Route should be the echoed-back caller payload")
	}
}

func TestResumePipeline_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.ResumePipeline, http.MethodPost, "/pipeline/resume", "{not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_JSON") {
		t.Errorf("body = %s", rec.Body.String())
	}
}
