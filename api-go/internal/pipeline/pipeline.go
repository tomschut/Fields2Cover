// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package pipeline

import (
	"context"
	"fmt"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// Pipeline composes the per-step helpers into a full-pipeline runner.
// Both shortcut endpoints use it:
//
//	POST /pipeline/plan-coverage -> Pipeline.RunAll
//	POST /pipeline/resume        -> Pipeline.RunFrom
//
// Pipeline is stateless; a single instance can serve concurrent
// requests as long as the underlying F2CClient is (the generated
// client is).
type Pipeline struct {
	steps *Steps
}

// NewPipeline wraps a Steps instance. Call NewFromClient for the
// common case of building from a bare F2CClient.
func NewPipeline(steps *Steps) *Pipeline { return &Pipeline{steps: steps} }

// NewFromClient constructs a Pipeline directly from an F2CClient,
// building a fresh Steps internally.
func NewFromClient(c f2cclient.F2CClient) *Pipeline {
	return &Pipeline{steps: NewSteps(c)}
}

// Steps returns the underlying Steps helper. Granular handlers
// (internal/server) use this to drive a single step without running
// the full pipeline.
func (p *Pipeline) Steps() *Steps { return p.steps }

// RunThroughSort executes the cheap prefix of the pipeline --
// parse -> headland -> swaths -> sort -- and stops. Returns a
// PipelineResult with Field, FieldWithHeadlands, Swaths, and
// SortedSwaths populated; Route and Path are nil. Used by
// /pipeline/plan-swaths so clients can preview the swath layout
// before committing to the expensive route+plan tail.
func (p *Pipeline) RunThroughSort(ctx context.Context, in PipelineInput) (*PipelineResult, error) {
	res := &PipelineResult{}

	field, err := p.steps.ParseField(ctx, in.Geojson, in.TargetCRS)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunThroughSort[parse-field]: %w", err)
	}
	res.Field = field

	fwh, err := p.steps.GenerateHeadland(ctx, *field, in.Robot, in.HeadlandWidthM, in.HeadlandCount)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunThroughSort[generate-headland]: %w", err)
	}
	res.FieldWithHeadlands = fwh

	swaths, err := p.steps.GenerateSwaths(ctx, *fwh, in.Robot, in.SwathAngleRad, in.SwathWidthM)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunThroughSort[generate-swaths]: %w", err)
	}
	res.Swaths = swaths

	sorted, err := p.steps.SortSwaths(ctx, *swaths, in.SortAlgorithm, in.SortVariant, in.SortStartPoint)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunThroughSort[sort-swaths]: %w", err)
	}
	res.SortedSwaths = sorted

	return res, nil
}

// RunAll executes every pipeline step in canonical order and returns
// the accumulated intermediates plus the terminal Path. Errors from
// any step short-circuit the run; earlier intermediates are lost (the
// returned *PipelineResult is nil on error).
func (p *Pipeline) RunAll(ctx context.Context, in PipelineInput) (*PipelineResult, error) {
	res := &PipelineResult{}

	field, err := p.steps.ParseField(ctx, in.Geojson, in.TargetCRS)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[parse-field]: %w", err)
	}
	res.Field = field

	fwh, err := p.steps.GenerateHeadland(ctx, *field, in.Robot, in.HeadlandWidthM, in.HeadlandCount)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[generate-headland]: %w", err)
	}
	res.FieldWithHeadlands = fwh

	swaths, err := p.steps.GenerateSwaths(ctx, *fwh, in.Robot, in.SwathAngleRad, in.SwathWidthM)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[generate-swaths]: %w", err)
	}
	res.Swaths = swaths

	sorted, err := p.steps.SortSwaths(ctx, *swaths, in.SortAlgorithm, in.SortVariant, in.SortStartPoint)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[sort-swaths]: %w", err)
	}
	res.SortedSwaths = sorted

	route, err := p.steps.GenerateRoute(ctx, *sorted)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[generate-route]: %w", err)
	}
	res.Route = route

	path, err := p.steps.PlanPath(ctx, *route, in.Robot, in.TurningAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunAll[plan-path]: %w", err)
	}
	res.Path = path

	return res, nil
}

// RunFrom resumes a pipeline at `start`, feeding it the intermediate
// payload the caller already has, and runs every subsequent step.
// The returned PipelineResult has the caller-supplied intermediate(s)
// preserved in the matching slot(s) so the shortcut endpoint can echo
// them back unchanged (proving to tests that earlier steps did not
// re-run).
//
// start=StepParseField is invalid — use RunAll instead.
func (p *Pipeline) RunFrom(ctx context.Context, start Step, in ResumePayload) (*PipelineResult, error) {
	res := &PipelineResult{}

	switch start {
	case StepParseField:
		return nil, fmt.Errorf("pipeline.RunFrom: start=%q is not a resume target; use RunAll", start)

	case StepGenerateHeadland:
		if in.Field == nil {
			return nil, fmt.Errorf("pipeline.RunFrom[%s]: payload.Field is required", start)
		}
		res.Field = in.Field
		fwh, err := p.steps.GenerateHeadland(ctx, *in.Field, in.Robot, in.HeadlandWidthM, in.HeadlandCount)
		if err != nil {
			return nil, fmt.Errorf("pipeline.RunFrom[generate-headland]: %w", err)
		}
		res.FieldWithHeadlands = fwh
		return p.tailFromSwaths(ctx, res, in, *fwh)

	case StepGenerateSwaths:
		if in.FieldWithHeadlands == nil {
			return nil, fmt.Errorf("pipeline.RunFrom[%s]: payload.FieldWithHeadlands is required", start)
		}
		res.FieldWithHeadlands = in.FieldWithHeadlands
		return p.tailFromSwaths(ctx, res, in, *in.FieldWithHeadlands)

	case StepSortSwaths:
		if in.Swaths == nil {
			return nil, fmt.Errorf("pipeline.RunFrom[%s]: payload.Swaths is required", start)
		}
		res.Swaths = in.Swaths
		return p.tailFromSort(ctx, res, in, *in.Swaths)

	case StepGenerateRoute:
		if in.SortedSwaths == nil {
			return nil, fmt.Errorf("pipeline.RunFrom[%s]: payload.SortedSwaths is required", start)
		}
		res.SortedSwaths = in.SortedSwaths
		return p.tailFromRoute(ctx, res, in, *in.SortedSwaths)

	case StepPlanPath:
		if in.Route == nil {
			return nil, fmt.Errorf("pipeline.RunFrom[%s]: payload.Route is required", start)
		}
		res.Route = in.Route
		path, err := p.steps.PlanPath(ctx, *in.Route, in.Robot, in.TurningAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("pipeline.RunFrom[plan-path]: %w", err)
		}
		res.Path = path
		return res, nil

	default:
		return nil, fmt.Errorf("pipeline.RunFrom: unknown step %q", start)
	}
}

// tailFromSwaths runs swaths -> sort -> route -> plan starting from a
// FieldWithHeadlands, updating res in place.
func (p *Pipeline) tailFromSwaths(ctx context.Context, res *PipelineResult, in ResumePayload, fwh api.FieldWithHeadlands) (*PipelineResult, error) {
	swaths, err := p.steps.GenerateSwaths(ctx, fwh, in.Robot, in.SwathAngleRad, in.SwathWidthM)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunFrom[generate-swaths]: %w", err)
	}
	res.Swaths = swaths
	return p.tailFromSort(ctx, res, in, *swaths)
}

// tailFromSort runs sort -> route -> plan starting from a Swaths.
func (p *Pipeline) tailFromSort(ctx context.Context, res *PipelineResult, in ResumePayload, swaths api.Swaths) (*PipelineResult, error) {
	sorted, err := p.steps.SortSwaths(ctx, swaths, in.SortAlgorithm, in.SortVariant, nil)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunFrom[sort-swaths]: %w", err)
	}
	res.SortedSwaths = sorted
	return p.tailFromRoute(ctx, res, in, *sorted)
}

// tailFromRoute runs route -> plan starting from sorted Swaths.
func (p *Pipeline) tailFromRoute(ctx context.Context, res *PipelineResult, in ResumePayload, sorted api.Swaths) (*PipelineResult, error) {
	route, err := p.steps.GenerateRoute(ctx, sorted)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunFrom[generate-route]: %w", err)
	}
	res.Route = route

	path, err := p.steps.PlanPath(ctx, *route, in.Robot, in.TurningAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("pipeline.RunFrom[plan-path]: %w", err)
	}
	res.Path = path
	return res, nil
}
