// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package server hosts the Server type that implements the generated
// api.ServerInterface. Each method follows the same template:
//
//  1. Decode the OpenAPI request from JSON
//  2. Validate path-id consistency (where applicable)
//  3. Call the shared pipeline.Steps helper (single source of truth for
//     proto conversion + gRPC dispatch — SHC-03)
//  4. Attach a journey envelope from internal/journey
//  5. Encode JSON to the response writer
//
// Errors from the step helper short-circuit through writeStepError,
// which distinguishes *pipeline.EmptyResponseError (500 EMPTY_RESPONSE)
// from gRPC status errors (delegated to writeGrpcError).
//
// Layout note: this package lives under internal/server (not
// internal/api) because internal/convert already imports internal/api
// for the generated OpenAPI types — putting the handlers in
// internal/api would create an import cycle.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/journey"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
)

// Server implements api.ServerInterface. It is intentionally stateless
// — no maps, no caches, no in-memory sessions (GOS-04). The only
// field is the shared pipeline.Steps helper (injected at startup or
// shared with the Pipeline used by Plan 12-02's shortcut handlers).
type Server struct {
	steps    *pipeline.Steps
	pipeline *pipeline.Pipeline
}

// NewServer constructs a Server bound to the supplied F2CClient. Both
// the granular handlers and the Plan 12-02 shortcut handlers share the
// single Pipeline's Steps — the SHC-03 contract is enforced here.
func NewServer(client f2cclient.F2CClient) *Server {
	p := pipeline.NewFromClient(client)
	return &Server{steps: p.Steps(), pipeline: p}
}

// NewServerFromPipeline lets callers inject a shared Pipeline. Used by
// cmd/api/main.go to wire a single Pipeline through startup wiring and
// by integration tests that want to stub the Pipeline.
func NewServerFromPipeline(p *pipeline.Pipeline) *Server {
	return &Server{steps: p.Steps(), pipeline: p}
}

// NewServerFromSteps preserves the Plan 12-01 constructor signature.
// It wraps the supplied Steps in a fresh Pipeline so the shortcut
// handlers still have something to dispatch through.
func NewServerFromSteps(s *pipeline.Steps) *Server {
	return &Server{steps: s, pipeline: pipeline.NewPipeline(s)}
}

// Compile-time check.
var _ api.ServerInterface = (*Server)(nil)

// writeJSON writes an HTTP response with Content-Type application/json.
// Encoding errors are intentionally swallowed — once we have started
// writing the body there is no useful recovery action.
func (s *Server) writeJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

// ParseField handles POST /fields:parse.
func (s *Server) ParseField(w http.ResponseWriter, r *http.Request) {
	var req api.ParseFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	field, err := s.steps.ParseField(r.Context(), req.Geojson, req.TargetCrs)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.ParseFieldResponse{
		Field:   *field,
		Journey: journey.AfterParse(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// GenerateHeadlands handles POST /headlands:generate.
func (s *Server) GenerateHeadlands(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateHeadlandsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	count := int32(0) // pipeline.Steps.GenerateHeadland applies the default
	if req.Count != nil {
		count = int32(*req.Count)
	}
	fwh, err := s.steps.GenerateHeadland(r.Context(), req.Field, req.Robot, req.WidthM, count)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.GenerateHeadlandsResponse{
		FieldWithHeadlands: *fwh,
		Journey:            journey.AfterHeadlands(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// GenerateSwaths handles POST /swaths:generate.
func (s *Server) GenerateSwaths(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateSwathsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	swaths, err := s.steps.GenerateSwaths(r.Context(), req.FieldWithHeadlands, req.Robot, req.AngleRad, req.WidthM)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.GenerateSwathsResponse{
		Swaths:  *swaths,
		Journey: journey.AfterSwaths(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// SortSwaths handles POST /swaths:sort.
func (s *Server) SortSwaths(w http.ResponseWriter, r *http.Request) {
	var req api.SortSwathsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	out, err := s.steps.SortSwaths(r.Context(), req.Swaths, req.Algorithm, req.Variant, nil)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.SortSwathsResponse{
		Swaths:  *out,
		Journey: journey.AfterSort(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// GenerateRoute handles POST /routes:generate.
func (s *Server) GenerateRoute(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	route, err := s.steps.GenerateRoute(r.Context(), req.Swaths)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.GenerateRouteResponse{
		Route:   *route,
		Journey: journey.AfterRoute(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// PlanPath handles POST /paths:plan (terminal step).
func (s *Server) PlanPath(w http.ResponseWriter, r *http.Request) {
	var req api.PlanPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	path, err := s.steps.PlanPath(r.Context(), req.Route, req.Robot, req.TurningAlgorithm)
	if err != nil {
		s.writeStepError(w, err)
		return
	}
	resp := api.PlanPathResponse{
		Path:    *path,
		Journey: journey.AfterPlan(),
	}
	s.writeJSON(w, http.StatusOK, resp)
}
