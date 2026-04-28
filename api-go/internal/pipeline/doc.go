// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package pipeline owns the shared per-step gRPC logic used by both the
// granular journey handlers (internal/server) and the shortcut handlers
// (/pipeline/plan-coverage, /pipeline/resume — added in Plan 12-02).
//
// Design rule (SHC-03): the granular handlers MUST call pipeline.Steps.X
// and the shortcut handlers MUST call pipeline.Pipeline.RunAll / RunFrom.
// Nothing outside this package may call f2cclient.F2CClient directly for
// a step operation — that is the single source of truth guarantee.
//
// Layout:
//
//	doc.go       — this file
//	result.go    — PipelineResult accumulator + input/output structs
//	steps.go     — Steps struct: one method per pipeline step, OpenAPI-in / OpenAPI-out
//	pipeline.go  — Pipeline struct: RunAll + RunFrom over Steps
//
// The package imports internal/api (generated OpenAPI types),
// internal/convert (paired converters), and internal/f2cclient (generated
// gRPC client). It does NOT import internal/journey — envelope building
// is a response-shaping concern that lives in the handler layer.
package pipeline
