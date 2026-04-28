// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Shortcut handlers: POST /pipeline/plan-coverage and POST /pipeline/resume.
// Both go through the shared *pipeline.Pipeline so they cannot diverge
// from the granular endpoints' per-step logic (SHC-03).

package server

import (
	"encoding/json"
	"net/http"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/convert"
	"github.com/Fields2Cover/fields2cover/api-go/internal/journey"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
)

// PlanSwaths handles POST /pipeline/plan-swaths. Runs the cheap
// prefix (parse -> headland -> swaths -> sort) and stops. Clients use
// this to preview the swath layout, then call /pipeline/resume with
// start_step=generate-route to commit to the expensive tail.
func (s *Server) PlanSwaths(w http.ResponseWriter, r *http.Request) {
	var req api.PlanCoverageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	in := pipeline.PipelineInput{
		Geojson:          req.Geojson,
		TargetCRS:        req.TargetCrs,
		Robot:            req.Robot,
		HeadlandWidthM:   req.HeadlandWidthM,
		SwathWidthM:      req.SwathWidthM,
		TurningAlgorithm: planCoverageTurningToPlanPath(req.TurningAlgorithm),
	}
	if req.HeadlandCount != nil {
		in.HeadlandCount = int32(*req.HeadlandCount)
	}
	in.SwathAngleRad = req.SwathAngleRad
	if req.SortAlgorithm != nil {
		a := api.SortSwathsRequestAlgorithm(*req.SortAlgorithm)
		in.SortAlgorithm = &a
	}
	if req.SortVariant != nil {
		v := *req.SortVariant
		in.SortVariant = &v
	}
	if req.SortStartPoint != nil {
		sp := []float64(*req.SortStartPoint)
		in.SortStartPoint = &sp
	}

	res, err := s.pipeline.RunThroughSort(r.Context(), in)
	if err != nil {
		s.writeStepError(w, err)
		return
	}

	resp := api.PlanSwathsResponse{
		Journey: journey.AfterSort(),
	}
	if res.FieldWithHeadlands != nil {
		resp.FieldWithHeadlands = *res.FieldWithHeadlands
	}
	if res.SortedSwaths != nil {
		resp.SortedSwaths = *res.SortedSwaths
	}
	convert.ProjectSwathsResponseToWGS84(&resp)
	s.writeJSON(w, http.StatusOK, resp)
}

// PlanCoverage handles POST /pipeline/plan-coverage.
func (s *Server) PlanCoverage(w http.ResponseWriter, r *http.Request) {
	var req api.PlanCoverageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	in := pipeline.PipelineInput{
		Geojson:          req.Geojson,
		TargetCRS:        req.TargetCrs,
		Robot:            req.Robot,
		HeadlandWidthM:   req.HeadlandWidthM,
		SwathWidthM:      req.SwathWidthM,
		TurningAlgorithm: planCoverageTurningToPlanPath(req.TurningAlgorithm),
	}
	if req.HeadlandCount != nil {
		in.HeadlandCount = int32(*req.HeadlandCount)
	}
	in.SwathAngleRad = req.SwathAngleRad
	if req.SortAlgorithm != nil {
		a := api.SortSwathsRequestAlgorithm(*req.SortAlgorithm)
		in.SortAlgorithm = &a
	}
	if req.SortVariant != nil {
		v := *req.SortVariant
		in.SortVariant = &v
	}
	if req.SortStartPoint != nil {
		sp := []float64(*req.SortStartPoint)
		in.SortStartPoint = &sp
	}

	res, err := s.pipeline.RunAll(r.Context(), in)
	if err != nil {
		s.writeStepError(w, err)
		return
	}

	resp := buildCoverageResponse(res)
	convert.ProjectResponseToWGS84(&resp)
	s.writeJSON(w, http.StatusOK, resp)
}

// ResumePipeline handles POST /pipeline/resume.
func (s *Server) ResumePipeline(w http.ResponseWriter, r *http.Request) {
	var req api.ResumePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	// The OpenAPI enum does not include "parse-field" at all, but we
	// still defensively catch any client that passed the string through
	// (e.g. a hand-crafted JSON body) before it reaches the pipeline.
	if string(req.StartStep) == string(pipeline.StepParseField) {
		s.writeError(w, http.StatusBadRequest, "INVALID_START_STEP",
			"start_step=parse-field is not a resume target; use /pipeline/plan-coverage")
		return
	}

	start, payload, validationErr := buildResumePayload(req)
	if validationErr != nil {
		s.writeError(w, http.StatusBadRequest, "MISSING_PAYLOAD", validationErr.Error())
		return
	}

	res, err := s.pipeline.RunFrom(r.Context(), start, payload)
	if err != nil {
		s.writeStepError(w, err)
		return
	}

	resumeResp := buildCoverageResponse(res)
	convert.ProjectResponseToWGS84(&resumeResp)
	s.writeJSON(w, http.StatusOK, resumeResp)
}

// buildCoverageResponse converts a *PipelineResult into the OpenAPI
// PlanCoverageResponse envelope. Every nil slot becomes a nil pointer
// in the response (OpenAPI `nullable: true`).
func buildCoverageResponse(res *pipeline.PipelineResult) api.PlanCoverageResponse {
	inter := api.FieldIntermediates{
		Field:              res.Field,
		FieldWithHeadlands: res.FieldWithHeadlands,
		Swaths:             res.Swaths,
		SortedSwaths:       res.SortedSwaths,
		Route:              res.Route,
	}
	resp := api.PlanCoverageResponse{
		Intermediates: inter,
		Journey:       journey.AfterPlan(), // terminal — same envelope as /paths:plan
	}
	if res.Path != nil {
		resp.Path = *res.Path
	}
	return resp
}

// buildResumePayload validates the resume request and converts it to
// a (Step, ResumePayload) pair. Returns a descriptive error if the
// payload does not match the start_step.
func buildResumePayload(req api.ResumePipelineRequest) (pipeline.Step, pipeline.ResumePayload, error) {
	rp := pipeline.ResumePayload{
		Field:              req.Field,
		FieldWithHeadlands: req.FieldWithHeadlands,
		Swaths:             req.Swaths,
		SortedSwaths:       req.SortedSwaths,
		Route:              req.Route,
		Robot:              req.Robot,
		TurningAlgorithm:   resumeTurningToPlanPath(req.TurningAlgorithm),
	}
	if req.HeadlandWidthM != nil {
		rp.HeadlandWidthM = *req.HeadlandWidthM
	}
	if req.HeadlandCount != nil {
		rp.HeadlandCount = int32(*req.HeadlandCount)
	}
	rp.SwathAngleRad = req.SwathAngleRad
	if req.SwathWidthM != nil {
		rp.SwathWidthM = *req.SwathWidthM
	}
	if req.SortAlgorithm != nil {
		a := api.SortSwathsRequestAlgorithm(*req.SortAlgorithm)
		rp.SortAlgorithm = &a
	}
	if req.SortVariant != nil {
		v := *req.SortVariant
		rp.SortVariant = &v
	}

	switch req.StartStep {
	case api.GenerateHeadland:
		if rp.Field == nil {
			return "", rp, errMissingPayload("generate-headland", "field")
		}
		return pipeline.StepGenerateHeadland, rp, nil
	case api.GenerateSwaths:
		if rp.FieldWithHeadlands == nil {
			return "", rp, errMissingPayload("generate-swaths", "field_with_headlands")
		}
		return pipeline.StepGenerateSwaths, rp, nil
	case api.SortSwaths:
		if rp.Swaths == nil {
			return "", rp, errMissingPayload("sort-swaths", "swaths")
		}
		return pipeline.StepSortSwaths, rp, nil
	case api.GenerateRoute:
		if rp.SortedSwaths == nil {
			return "", rp, errMissingPayload("generate-route", "sorted_swaths")
		}
		return pipeline.StepGenerateRoute, rp, nil
	case api.PlanPath:
		if rp.Route == nil {
			return "", rp, errMissingPayload("plan-path", "route")
		}
		return pipeline.StepPlanPath, rp, nil
	default:
		return "", rp, errUnknownStartStep(string(req.StartStep))
	}
}

// planCoverageTurningToPlanPath converts the per-request turning-algo
// enum produced by oapi-codegen for PlanCoverageRequest back into the
// canonical PlanPathRequestTurningAlgorithm that the pipeline helpers
// consume. Both types have identical string values so the conversion
// is total.
func planCoverageTurningToPlanPath(a api.PlanCoverageRequestTurningAlgorithm) api.PlanPathRequestTurningAlgorithm {
	return api.PlanPathRequestTurningAlgorithm(a)
}

// resumeTurningToPlanPath is the same idempotent conversion for the
// ResumePipelineRequest's turning-algo enum.
func resumeTurningToPlanPath(a api.ResumePipelineRequestTurningAlgorithm) api.PlanPathRequestTurningAlgorithm {
	return api.PlanPathRequestTurningAlgorithm(a)
}

type resumeValidationError struct{ msg string }

func (e *resumeValidationError) Error() string { return e.msg }

func errMissingPayload(step, field string) error {
	return &resumeValidationError{
		msg: "start_step=" + step + " requires payload." + field,
	}
}

func errUnknownStartStep(s string) error {
	return &resumeValidationError{msg: "unknown start_step: " + s}
}
