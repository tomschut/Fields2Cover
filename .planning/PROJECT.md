# Fields2Cover API — Interactive Demo + Journey-First Go API

## What This Is

Fields2Cover is a C++17 coverage path planning library for agricultural robots. The project has shipped three milestones: v1.0 hardened the Python/connexion API; v2.0 rewrote the API layer in Go with a journey-first surface (Go↔f2c via gRPC shim, clients in wodan and farmmaps, Python API retired); v3.0 ships a React front-end — an interactive demo served from the Go container with Leaflet map, PDOK parcel import, full pipeline UI, and start-point selection. **v4.0 extends f2c with multi-robot field division and follower coordination**, wired through gRPC → Go API → React UI.

## Current Milestone: v4.0 Multi-Robot Planning

**Goal:** Extend the fields2cover library with multi-robot field partitioning and follower path coordination, expose both through new gRPC RPCs and Go API endpoints, and add fleet configuration + multi-path visualization to the React frontend.

**Target features:**
- **Multi-robot division** — N robots with heterogeneous specs; field partitioned by work rate (width × speed); each robot gets a full independent coverage plan; all paths shown in distinct colors
- **Follower planning** — configure a trailing follower (capacity, unload time, speed); system computes follower path alongside primary robot route, marks headland rendezvous points; both visualized as toggleable overlays
- **New C++ in f2c** — multi-robot partitioning algorithm + follower coordination algorithm
- **New gRPC RPCs** + Go API endpoints (`/pipeline/plan-multi-robot`, `/pipeline/plan-follower`) + frontend robot fleet and follower panels

## Core Value

Every API call either produces a directly-usable result or produces a typed handoff that the next call in the journey accepts without adaptation.

## Current State (after v3.0)

**Shipped:** 2026-04-29
**Stack:** Go API (chi, oapi-codegen) + f2c gRPC shim (C++) + React 18 / Vite 8 / TypeScript frontend, served via go:embed
**Frontend:** ~3,059 TS/TSX LOC across `frontend/src/`
**Go API:** ~11,381 LOC in `api-go/`
**Image:** Single multi-stage Docker container, Go + C++ gRPC shim supervised
**Branch:** `feat/api-full`

## Requirements

### Validated

- ✓ REST API exposing coverage path planning pipeline — existing
- ✓ Field import (GML→GeoJSON), headland generation, swath generation, route ordering, path planning endpoints — existing
- ✓ OpenAPI spec + go-chi routing — v2.0
- ✓ Integration test suite — v1.0 / v2.0
- ✓ C++ unit tests for core library (GoogleTest) — existing
- ✓ Journey-first OpenAPI surface (field→headland→swath→sort→route→path) — v2.0
- ✓ One-shot `/pipeline/plan-coverage` shortcut endpoint — v2.0
- ✓ wodan + farmmaps Go clients (oapi-codegen generated) — v2.0
- ✓ Python API retired — v2.0
- ✓ React app (`frontend/`) served statically from Go API at `/` — v3.0 Phase 17
- ✓ Leaflet map with draw tools (geoman), file upload, GeoJSON paste, tile toggle — v3.0 Phase 18
- ✓ Full pipeline run from UI — sidebar form, one-click run, 4-layer result overlays — v3.0 Phase 19
- ✓ Integration guide page with curl/JS/Go code examples — v3.0 Phase 20
- ✓ Swagger UI API docs page at `/api-docs` — v3.0 Phase 20
- ✓ PDOK gewaspercelen crop parcel import — zoom overlay, click to select — v3.0 Phase 21
- ✓ Start-point selection — map click → backend-native sort_start_point → C++ variant loop — v3.0 Phases 22+23

### Active (next milestone candidates)

- [ ] Step-by-step stepper — each of the 6 journey endpoints triggered individually, intermediate results shown after each step
- [ ] Export path as GPX or KML for field navigation systems
- [ ] Address search via PDOK Locatieserver — autocomplete to location, then pick a parcel
- [ ] Multi-field selection — select multiple PDOK parcels and merge into one field polygon
- [ ] Mobile-responsive layout — map and sidebar stack vertically on small screens
- [ ] Share URL encoding field + parameters in query string for reproducible demos
- [ ] `/pipeline/resume` start-point support — ResumePipelineRequest currently excludes sort_start_point

### Out of Scope

| Feature | Reason |
|---------|--------|
| API key authentication | Demo tool — no accounts or saved sessions |
| `transform_to_prev_crs` endpoint | Permanently removed from spec; CRS tracking not scoped |
| Visualizer endpoint implementation | Models exist but no immediate use case |
| GDAL/Docker base image upgrade | Separate infrastructure concern |
| Headland coverage, obstacle support | Upstream library features, not API layer |
| Redoc (read-only docs) | Swagger UI try-it-out is more useful at same origin |
| Mobile app | Web-first; React app sufficient |
| Multi-language i18n | English only |

## Context

Brownfield project. v1.0 stabilized the Python API; v2.0 replaced it with Go; v3.0 added the frontend. The React app is embedded in the Go binary via `go:embed` — no separate static file server. The PDOK gewaspercelen API requires explicit CRS84 parameter; the default (EPSG:28992) displaces all polygons off the Netherlands. The SPA fallback uses a `statusRecorder` pattern to avoid poisoning API 404 responses.

**Known tech debt after v3.0:**
- `planSwaths` + `pickBestVariant` exported from `pipeline.ts` — unused in production (only in tests); dead exports from Phase 22 pre-backend
- Stale `planSwaths` mock in `MapPage.test.tsx` line 39
- `/pipeline/resume` does not support `sort_start_point` — by design, known limitation

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Reject out-of-range variant with 400 | Integration test already expects this; explicit errors better than silent clamping | ✓ Good |
| Remove transform/to-previous-crs endpoint | Permanently 501; misleading client expectations — cleaner to remove | ✓ Good |
| pytest + Flask test client for API tests | In-process, fast, CI-friendly; avoids server lifecycle complexity | ✓ Good |
| Upgrade to connexion 3.x | 2.x unmaintained, no security patches; breaking change isolated to own phase | ✓ Good |
| Sandbox filePath to configured base dir | Path traversal critical security fix | ✓ Good |
| Go for API layer | Kills SWIG maintenance tax; compile-time type safety; static binary; smaller image | ✓ Good |
| gRPC over Unix socket in-container | Clean process isolation — f2c SEGV returns 500 instead of killing API | ✓ Good |
| Not cgo | Debuggability across FFI boundary is poor | ✓ Good |
| Vite 8 + React 18.3 (not 19) | React 19 blocked by leaflet-geoman ecosystem | ✓ Good |
| leaflet-geoman (not react-leaflet-draw) | react-leaflet-draw is abandoned | ✓ Good |
| ky@1.7.x HTTP client | ESM-native; no Redux/TanStack needed for this scope | ✓ Good |
| statusRecorder SPA fallback pattern | `r.NotFound` pattern poisons API 404s | ✓ Good |
| Always request CRS84 from PDOK | Default EPSG:28992 displaces polygons off Netherlands | ✓ Good |
| No Redux/Zustand | Three pages; useState + useContext sufficient | ✓ Good |
| Native backend start-point (Phase 23) | Eliminated 4 parallel planSwaths round-trips; cleaner contract | ✓ Good |

## Constraints

- **Go API:** chi router, oapi-codegen-generated handlers, gRPC client to f2c shim on Unix socket
- **Frontend:** Vite dev proxy → :8080; go:embed for production; React 18.3 (not 19 — ecosystem constraint)
- **Docker:** Multi-stage build; Node stage builds frontend; Go stage embeds dist/; single container
- **PDOK:** CRS84 must be requested explicitly; rate-limit with AbortController on rapid pans

---
*Last updated: 2026-04-29 after v3.0 milestone*
