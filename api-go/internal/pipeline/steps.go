// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/convert"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// Steps owns the per-step gRPC calls. Each method is OpenAPI-in /
// OpenAPI-out. All proto conversion is done here so the rest of the
// code base never touches f2cclient types for pipeline operations.
//
// Steps is safe for concurrent use iff the underlying F2CClient is
// (the generated client is).
type Steps struct {
	f2c f2cclient.F2CClient
}

// NewSteps constructs a Steps bound to the supplied gRPC client.
func NewSteps(client f2cclient.F2CClient) *Steps {
	return &Steps{f2c: client}
}

// Client returns the underlying F2CClient. Used by code that needs to
// invoke non-pipeline RPCs (e.g. CloneRobot for /readyz probes). NOT a
// back door for pipeline ops — those must go through the step methods
// below.
func (s *Steps) Client() f2cclient.F2CClient { return s.f2c }

// --- Per-step helpers --------------------------------------------------

// ParseField calls ParseGeoJSON. geojson is the raw inline
// FeatureCollection; the helper marshals it to the string expected by
// the proto.
func (s *Steps) ParseField(
	ctx context.Context,
	geojson api.GeoJSONFeatureCollection,
	targetCRS *api.CRS,
) (*api.Field, error) {
	geojsonBytes, err := json.Marshal(geojson)
	if err != nil {
		return nil, fmt.Errorf("pipeline.ParseField: marshal geojson: %w", err)
	}
	grpcReq := &f2cclient.ParseGeoJSONRequest{
		Geojson: string(geojsonBytes),
	}
	if targetCRS != nil {
		grpcReq.TargetCrs = convert.CRSToProto(*targetCRS)
	}
	resp, err := s.f2c.ParseGeoJSON(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Field == nil {
		return nil, errEmptyResponse("ParseGeoJSON")
	}
	field := convert.FieldFromProto(resp.Field)
	return &field, nil
}

// GenerateHeadland calls GenerateHeadland. count<=0 uses the default 3.
func (s *Steps) GenerateHeadland(
	ctx context.Context,
	field api.Field,
	robot api.Robot,
	widthM float64,
	count int32,
) (*api.FieldWithHeadlands, error) {
	if count <= 0 {
		count = 3
	}
	grpcReq := &f2cclient.GenerateHeadlandRequest{
		Field:  convert.FieldToProto(field),
		Robot:  convert.RobotToProto(robot),
		WidthM: widthM,
		Count:  count,
	}
	resp, err := s.f2c.GenerateHeadland(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.FieldWithHeadlands == nil {
		return nil, errEmptyResponse("GenerateHeadland")
	}
	out := convert.FieldWithHeadlandsFromProto(resp.FieldWithHeadlands)
	return &out, nil
}

// GenerateSwaths calls GenerateSwaths.
// angleRad is optional: pass nil to let the server auto-select the heading
// that minimises swath count (brute-force over [0, π)).
func (s *Steps) GenerateSwaths(
	ctx context.Context,
	fwh api.FieldWithHeadlands,
	robot api.Robot,
	angleRad *float64,
	widthM float64,
) (*api.Swaths, error) {
	grpcReq := &f2cclient.GenerateSwathsRequest{
		FieldWithHeadlands: convert.FieldWithHeadlandsToProto(fwh),
		Robot:              convert.RobotToProto(robot),
		WidthM:             widthM,
	}
	if angleRad != nil {
		grpcReq.AngleRad = angleRad
	}
	resp, err := s.f2c.GenerateSwaths(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Swaths == nil {
		return nil, errEmptyResponse("GenerateSwaths")
	}
	out := convert.SwathsFromProto(resp.Swaths)
	return &out, nil
}

// SortSwaths calls SortSwaths. algorithm=nil -> UNSPECIFIED (shim picks
// BOUSTROPHEDON); variant=nil -> 0. When startPoint is non-nil and the
// field CRS is a recognised UTM zone, the WGS84 [lng, lat] is projected
// to local metric and forwarded as SortSwathsRequest.start_point so the
// C++ server can auto-select the nearest variant.
func (s *Steps) SortSwaths(
	ctx context.Context,
	swaths api.Swaths,
	algorithm *api.SortSwathsRequestAlgorithm,
	variant *int,
	startPoint *[]float64,
) (*api.Swaths, error) {
	algo := f2cclient.SortAlgorithm_SORT_ALGORITHM_UNSPECIFIED
	if algorithm != nil {
		algo = SortAlgorithmToProto(*algorithm)
	}
	v := int32(0)
	if variant != nil {
		v = int32(*variant)
	}
	grpcReq := &f2cclient.SortSwathsRequest{
		Swaths:    convert.SwathsToProto(swaths),
		Algorithm: algo,
		Variant:   v,
	}
	if startPoint != nil && len(*startPoint) >= 2 {
		sp := *startPoint
		field := swaths.Field.Field
		if field.RefPoint != nil && len(*field.RefPoint) >= 2 {
			rp := *field.RefPoint
			lx, ly, ok := convert.WGS84ToLocal(sp[0], sp[1], rp[0], rp[1], field.Crs.Epsg)
			if ok {
				grpcReq.StartPoint = &f2cclient.Point{X: lx, Y: ly}
			}
		}
	}
	resp, err := s.f2c.SortSwaths(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Swaths == nil {
		return nil, errEmptyResponse("SortSwaths")
	}
	out := convert.SwathsFromProto(resp.Swaths)
	return &out, nil
}

// GenerateRoute calls GenerateRoute.
func (s *Steps) GenerateRoute(
	ctx context.Context,
	sortedSwaths api.Swaths,
) (*api.Route, error) {
	grpcReq := &f2cclient.GenerateRouteRequest{
		Swaths: convert.SwathsToProto(sortedSwaths),
	}
	resp, err := s.f2c.GenerateRoute(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Route == nil {
		return nil, errEmptyResponse("GenerateRoute")
	}
	out := convert.RouteFromProto(resp.Route)
	return &out, nil
}

// PlanPath calls PlanPath (terminal step).
func (s *Steps) PlanPath(
	ctx context.Context,
	route api.Route,
	robot api.Robot,
	turningAlgorithm api.PlanPathRequestTurningAlgorithm,
) (*api.Path, error) {
	grpcReq := &f2cclient.PlanPathRequest{
		Route:            convert.RouteToProto(route),
		Robot:            convert.RobotToProto(robot),
		TurningAlgorithm: TurningAlgorithmToProto(turningAlgorithm),
	}
	resp, err := s.f2c.PlanPath(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Path == nil {
		return nil, errEmptyResponse("PlanPath")
	}
	out := convert.PathFromProto(resp.Path)
	return &out, nil
}

// --- Enum bridges (moved from internal/server/server.go) ---------------

// SortAlgorithmToProto maps OpenAPI SortSwathsRequest.Algorithm enum
// to the proto SortAlgorithm enum. Exported so the refactored server
// handlers (and future code) reuse it instead of re-declaring.
func SortAlgorithmToProto(a api.SortSwathsRequestAlgorithm) f2cclient.SortAlgorithm {
	switch a {
	case api.BOUSTROPHEDON:
		return f2cclient.SortAlgorithm_SORT_ALGORITHM_BOUSTROPHEDON
	case api.SNAKE:
		return f2cclient.SortAlgorithm_SORT_ALGORITHM_SNAKE
	case api.SPIRAL:
		return f2cclient.SortAlgorithm_SORT_ALGORITHM_SPIRAL
	default:
		return f2cclient.SortAlgorithm_SORT_ALGORITHM_UNSPECIFIED
	}
}

// TurningAlgorithmToProto maps OpenAPI PlanPathRequest.TurningAlgorithm
// enum to the proto TurningAlgorithm enum.
func TurningAlgorithmToProto(a api.PlanPathRequestTurningAlgorithm) f2cclient.TurningAlgorithm {
	switch a {
	case api.PlanPathRequestTurningAlgorithmDUBINS:
		return f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS
	case api.PlanPathRequestTurningAlgorithmDUBINSCC:
		return f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS_CC
	case api.PlanPathRequestTurningAlgorithmREEDSSHEPP:
		return f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP
	case api.PlanPathRequestTurningAlgorithmREEDSSHEPPHC:
		return f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP_HC
	default:
		return f2cclient.TurningAlgorithm_TURNING_ALGORITHM_UNSPECIFIED
	}
}

// EmptyResponseError is returned when a shim RPC succeeded but did not
// populate its payload — a Rule 2 hardening guard inherited from Phase
// 11-05. Handlers map this to 500 EMPTY_RESPONSE.
type EmptyResponseError struct{ Op string }

func (e *EmptyResponseError) Error() string {
	return fmt.Sprintf("pipeline: shim returned empty response from %s", e.Op)
}

func errEmptyResponse(op string) error { return &EmptyResponseError{Op: op} }
