// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package pipeline

import (
	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// Step identifies a single pipeline step. Values match the OpenAPI
// Journey.step_name enum prefix (e.g. "parse-field" matches the
// fields:parse step). Used as the start-point for Pipeline.RunFrom.
type Step string

const (
	StepParseField       Step = "parse-field"
	StepGenerateHeadland Step = "generate-headland"
	StepGenerateSwaths   Step = "generate-swaths"
	StepSortSwaths       Step = "sort-swaths"
	StepGenerateRoute    Step = "generate-route"
	StepPlanPath         Step = "plan-path"
)

// AllSteps is the canonical ordering used by Pipeline.RunAll.
var AllSteps = []Step{
	StepParseField,
	StepGenerateHeadland,
	StepGenerateSwaths,
	StepSortSwaths,
	StepGenerateRoute,
	StepPlanPath,
}

// PipelineInput is the Go-native input to Pipeline.RunAll. Plan 12-02's
// /pipeline/plan-coverage handler builds this from the OpenAPI
// PlanCoverageRequest schema.
type PipelineInput struct {
	Geojson          api.GeoJSONFeatureCollection
	TargetCRS        *api.CRS // optional; nil = server picks UTM
	Robot            api.Robot
	HeadlandWidthM   float64
	HeadlandCount    int32 // 0 = default (3)
	SwathAngleRad    *float64 // nil = auto-optimize (minimises swath count)
	SwathWidthM      float64
	SortAlgorithm    *api.SortSwathsRequestAlgorithm // nil = UNSPECIFIED (shim picks BOUSTROPHEDON)
	SortVariant      *int
	SortStartPoint   *[]float64 // WGS84 [lng, lat]; nil = use SortVariant
	TurningAlgorithm api.PlanPathRequestTurningAlgorithm
}

// PipelineResult accumulates every intermediate payload a run produces.
// A step that did not execute (RunFrom started later) keeps the caller-
// supplied seed in the matching slot. The field pointers are all
// populated on a successful RunAll; partial nils signal an aborted run.
type PipelineResult struct {
	Field              *api.Field
	FieldWithHeadlands *api.FieldWithHeadlands
	Swaths             *api.Swaths
	SortedSwaths       *api.Swaths
	Route              *api.Route
	Path               *api.Path
}

// ResumePayload carries the caller-supplied intermediate that
// Pipeline.RunFrom uses as its starting point. Exactly one field must
// be non-nil, and it must match the Step passed to RunFrom:
//
//	StepGenerateHeadland -> Field
//	StepGenerateSwaths   -> FieldWithHeadlands
//	StepSortSwaths       -> Swaths
//	StepGenerateRoute    -> SortedSwaths
//	StepPlanPath         -> Route
//
// StepParseField cannot be a resume target (there is nothing before it
// to resume from) — use Pipeline.RunAll for that case.
type ResumePayload struct {
	Field              *api.Field
	FieldWithHeadlands *api.FieldWithHeadlands
	Swaths             *api.Swaths
	SortedSwaths       *api.Swaths
	Route              *api.Route

	// Shared across every resume: the full-pipeline parameters needed
	// by the remaining steps (robot, widths, algorithms).
	Robot            api.Robot
	HeadlandWidthM   float64
	HeadlandCount    int32
	SwathAngleRad    *float64 // nil = auto-optimize (minimises swath count)
	SwathWidthM      float64
	SortAlgorithm    *api.SortSwathsRequestAlgorithm
	SortVariant      *int
	TurningAlgorithm api.PlanPathRequestTurningAlgorithm
}
