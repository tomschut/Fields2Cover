// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package pipeline

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

func f64p(v float64) *float64 { return &v }

func fxPipelineInput() PipelineInput {
	return PipelineInput{
		Geojson:          fxAPIGeoJSON(),
		Robot:            fxAPIRobot(),
		HeadlandWidthM:   3.0,
		HeadlandCount:    2,
		SwathAngleRad:    f64p(1.5),
		SwathWidthM:      2.7,
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	}
}

func TestPipeline_RunAll_Happy(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			calls.Add(1)
			return &f2cclient.ParseGeoJSONResponse{Field: fxProtoField("fld_test_001")}, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: fxProtoFWH("fld_test_001")}, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0, 1)}, nil
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0, 1)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunAll(context.Background(), fxPipelineInput())
	if err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	if calls.Load() != 6 {
		t.Errorf("call count = %d, want 6", calls.Load())
	}
	if res == nil || res.Field == nil || res.FieldWithHeadlands == nil || res.Swaths == nil ||
		res.SortedSwaths == nil || res.Route == nil || res.Path == nil {
		t.Fatalf("incomplete result: %+v", res)
	}
}

func TestPipeline_RunAll_ShortCircuitsOnGenerateSwaths(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			calls.Add(1)
			return &f2cclient.ParseGeoJSONResponse{Field: fxProtoField("fld")}, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: fxProtoFWH("fld")}, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			calls.Add(1)
			return nil, status.Error(codes.InvalidArgument, "bad angle")
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			t.Fatalf("SortSwaths must not be called after short-circuit")
			return nil, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			t.Fatalf("GenerateRoute must not be called after short-circuit")
			return nil, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			t.Fatalf("PlanPath must not be called after short-circuit")
			return nil, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunAll(context.Background(), fxPipelineInput())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result, got %+v", res)
	}
	if !strings.Contains(err.Error(), "generate-swaths") {
		t.Errorf("err = %v, want to mention generate-swaths", err)
	}
	if calls.Load() != 3 {
		t.Errorf("call count = %d, want 3 (parse + headland + swaths-attempt)", calls.Load())
	}
	// Verify the gRPC status is still reachable via errors.As/status.FromError.
	unwrapped := err
	for unwrapped != nil {
		if st, ok := status.FromError(unwrapped); ok && st.Code() == codes.InvalidArgument {
			break
		}
		u, ok := unwrapped.(interface{ Unwrap() error })
		if !ok {
			t.Errorf("could not locate gRPC status in wrapped err: %v", err)
			break
		}
		unwrapped = u.Unwrap()
	}
}

func TestPipeline_RunFrom_ParseFieldIsInvalid(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{} // everything returns Unimplemented
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepParseField, ResumePayload{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result, got %+v", res)
	}
	if !strings.Contains(err.Error(), "not a resume target") {
		t.Errorf("err = %v", err)
	}
}

func TestPipeline_RunFrom_SortSwaths_SentinelPreserved(t *testing.T) {
	t.Parallel()
	sentinelSwaths := fxAPISwaths("fld_test_001", 9999)

	var calls atomic.Int32
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			t.Fatalf("ParseGeoJSON must not be called when resuming from sort")
			return nil, nil
		},
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			t.Fatalf("GenerateHeadland must not be called when resuming from sort")
			return nil, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			t.Fatalf("GenerateSwaths must not be called when resuming from sort")
			return nil, nil
		},
		sortSwaths: func(_ context.Context, req *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			calls.Add(1)
			// Return a distinct sorted swath (id 42) so we can verify
			// SortedSwaths != Swaths-input.
			if len(req.Swaths.GetItems()) == 0 || req.Swaths.Items[0].Id != 9999 {
				t.Errorf("mock: incoming swaths.items[0].id = %d, want 9999", req.Swaths.Items[0].Id)
			}
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 42)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepSortSwaths, ResumePayload{
		Swaths:           &sentinelSwaths,
		Robot:            fxAPIRobot(),
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	})
	if err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("call count = %d, want 3 (sort + route + plan)", calls.Load())
	}
	if res.Field != nil {
		t.Errorf("Field should be nil (not run), got %+v", res.Field)
	}
	if res.FieldWithHeadlands != nil {
		t.Errorf("FieldWithHeadlands should be nil (not run)")
	}
	if res.Swaths == nil || len(res.Swaths.Items) != 1 || res.Swaths.Items[0].Id != 9999 {
		t.Errorf("sentinel Swaths not preserved: %+v", res.Swaths)
	}
	if res.SortedSwaths == nil || len(res.SortedSwaths.Items) == 0 || res.SortedSwaths.Items[0].Id != 42 {
		t.Errorf("SortedSwaths wrong: %+v", res.SortedSwaths)
	}
	if res.Route == nil || res.Path == nil {
		t.Errorf("Route/Path not populated")
	}
}

func TestPipeline_RunFrom_PlanPathOnly(t *testing.T) {
	t.Parallel()
	seedRoute := fxAPIRoute("fld_test_001")
	var calls atomic.Int32
	mock := &mockF2CClient{
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepPlanPath, ResumePayload{
		Route:            &seedRoute,
		Robot:            fxAPIRobot(),
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	})
	if err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if calls.Load() != 1 {
		t.Errorf("call count = %d, want 1", calls.Load())
	}
	if res.Route != &seedRoute {
		t.Errorf("Route pointer not preserved")
	}
	if res.Path == nil {
		t.Errorf("Path not populated")
	}
}

func TestPipeline_RunFrom_MismatchedPayload(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		step    Step
		payload ResumePayload
		wantMsg string
	}{
		{"sort-nil-swaths", StepSortSwaths, ResumePayload{}, "payload.Swaths is required"},
		{"headland-nil-field", StepGenerateHeadland, ResumePayload{}, "payload.Field is required"},
		{"swaths-nil-fwh", StepGenerateSwaths, ResumePayload{}, "payload.FieldWithHeadlands is required"},
		{"route-nil-sorted", StepGenerateRoute, ResumePayload{}, "payload.SortedSwaths is required"},
		{"plan-nil-route", StepPlanPath, ResumePayload{}, "payload.Route is required"},
	}
	mock := &mockF2CClient{}
	p := NewFromClient(mock)
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := p.RunFrom(context.Background(), c.step, c.payload)
			if err == nil || !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("err = %v, want contains %q", err, c.wantMsg)
			}
		})
	}
}

func TestPipeline_RunFrom_UnknownStep(t *testing.T) {
	t.Parallel()
	p := NewFromClient(&mockF2CClient{})
	_, err := p.RunFrom(context.Background(), Step("banana"), ResumePayload{})
	if err == nil || !strings.Contains(err.Error(), "unknown step") {
		t.Errorf("err = %v, want 'unknown step'", err)
	}
}

func TestPipeline_RunFrom_GenerateRoute_Happy(t *testing.T) {
	t.Parallel()
	seedSorted := fxAPISwaths("fld_test_001", 0, 1)
	var calls atomic.Int32
	mock := &mockF2CClient{
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepGenerateRoute, ResumePayload{
		SortedSwaths:     &seedSorted,
		Robot:            fxAPIRobot(),
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	})
	if err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("call count = %d, want 2", calls.Load())
	}
	if res.SortedSwaths != &seedSorted {
		t.Errorf("SortedSwaths pointer not preserved")
	}
	if res.Route == nil || res.Path == nil {
		t.Errorf("Route/Path not populated")
	}
}

func TestPipeline_RunFrom_GenerateHeadland_RunsFullTail(t *testing.T) {
	t.Parallel()
	seedField := fxAPIField("fld_test_001")
	var calls atomic.Int32
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateHeadlandResponse{FieldWithHeadlands: fxProtoFWH("fld_test_001")}, nil
		},
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0, 1)}, nil
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0, 1)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepGenerateHeadland, ResumePayload{
		Field:            &seedField,
		Robot:            fxAPIRobot(),
		HeadlandWidthM:   3.0,
		HeadlandCount:    2,
		SwathAngleRad:    f64p(1.5),
		SwathWidthM:      2.7,
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	})
	if err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if calls.Load() != 5 {
		t.Errorf("call count = %d, want 5 (headland+swaths+sort+route+plan)", calls.Load())
	}
	if res.Field != &seedField {
		t.Errorf("Field pointer not preserved")
	}
	if res.FieldWithHeadlands == nil || res.Swaths == nil || res.SortedSwaths == nil || res.Route == nil || res.Path == nil {
		t.Fatalf("incomplete result: %+v", res)
	}
}

func TestPipeline_RunFrom_GenerateSwaths_RunsSwathsAndTail(t *testing.T) {
	t.Parallel()
	seedFWH := fxAPIFWH("fld_test_001")
	var calls atomic.Int32
	mock := &mockF2CClient{
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0)}, nil
		},
		sortSwaths: func(_ context.Context, _ *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			calls.Add(1)
			return &f2cclient.SortSwathsResponse{Swaths: fxProtoSwaths("fld_test_001", 0)}, nil
		},
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			calls.Add(1)
			return &f2cclient.GenerateRouteResponse{Route: fxProtoRoute("fld_test_001")}, nil
		},
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			calls.Add(1)
			return &f2cclient.PlanPathResponse{Path: fxProtoPath()}, nil
		},
	}
	p := NewFromClient(mock)
	res, err := p.RunFrom(context.Background(), StepGenerateSwaths, ResumePayload{
		FieldWithHeadlands: &seedFWH,
		Robot:              fxAPIRobot(),
		SwathAngleRad:      f64p(1.5),
		SwathWidthM:        2.7,
		TurningAlgorithm:   api.PlanPathRequestTurningAlgorithmDUBINS,
	})
	if err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if calls.Load() != 4 {
		t.Errorf("call count = %d, want 4", calls.Load())
	}
	if res.FieldWithHeadlands != &seedFWH {
		t.Errorf("FieldWithHeadlands pointer not preserved")
	}
}

func TestEmptyResponseError_ErrorString(t *testing.T) {
	t.Parallel()
	e := &EmptyResponseError{Op: "Whatever"}
	if !strings.Contains(e.Error(), "Whatever") {
		t.Errorf("Error() = %q, want to contain op name", e.Error())
	}
}

func TestPipeline_Steps_Accessor(t *testing.T) {
	t.Parallel()
	steps := NewSteps(&mockF2CClient{})
	p := NewPipeline(steps)
	if p.Steps() != steps {
		t.Errorf("Steps() did not return the wrapped pointer")
	}
}
