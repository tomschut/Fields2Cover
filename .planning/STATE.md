---
gsd_state_version: 1.0
milestone: v4.0
milestone_name: Multi-Robot Planning
status: planning
last_updated: "2026-04-29T00:00:00.000Z"
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Deferred Items

Items acknowledged and deferred at milestone close on 2026-04-29:

| Category | Item | Status |
|----------|------|--------|
| verification | Phase 17: 17-VERIFICATION.md | human_needed (UAT done — 13/13 browser tests passed) |
| verification | Phase 18: 18-VERIFICATION.md | human_needed (UAT done — 13/13 browser tests passed) |
| verification | Phase 19: 19-VERIFICATION.md | human_needed (UAT done — 13/13 browser tests passed) |
| requirements | DOCS-01: Integration guide page | Pending in REQUIREMENTS.md — implemented and wired per integration checker |
| requirements | DOCS-02: Swagger UI API docs page | Pending in REQUIREMENTS.md — implemented and wired per integration checker |
| requirements | PDOK-01: PDOK parcel overlay on map | Pending in REQUIREMENTS.md — implemented and wired per integration checker |
| requirements | PDOK-02: Click parcel to select as field | Pending in REQUIREMENTS.md — implemented and wired per integration checker |
| requirements | START-01: Start-point selection | Missing from REQUIREMENTS.md — implemented in phases 22+23 |

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-29 after v3.0 milestone)

**Core value:** Every API call either produces a directly-usable result or produces a typed handoff that the next call in the journey accepts without adaptation.
**Current focus:** v4.0 Multi-Robot Planning — Phase 24 (C++ Multi-Robot Partitioning)

## Current Position

Milestone: v4.0 Multi-Robot Planning
Phase: 24 — NOT STARTED
Plan: -

```
[                                        ] 0%
Phase 24 of 28 — PLANNING
```

## Phase Status (v4.0)

| Phase | Name | Status | Notes |
|-------|------|--------|-------|
| 24 | C++ Multi-Robot Partitioning | Not started | - |
| 25 | C++ Follower Coordination | Not started | - |
| 26 | gRPC RPCs + Go API Endpoints | Not started | - |
| 27 | Multi-Robot Frontend | Not started | - |
| 28 | Follower Frontend | Not started | - |

## Phase Status (v3.0 — complete)

| Phase | Name | Status | Notes |
|-------|------|--------|-------|
| 17 | Infrastructure Foundation | Complete | 2026-04-16 |
| 18 | Map Shell & Field Input | Complete | 2026-04-16 |
| 19 | Pipeline Integration | Complete | 2026-04-16 |
| 20 | Docs Pages | Complete | 2026-04-19 |
| 21 | PDOK Gewaspercelen Import | Complete | 2026-04-20 |
| 22 | Start/End Point Selection | Complete | 2026-04-28 |
| 23 | Start-Point Native Backend | Complete | 2026-04-28 |

## Phase Status (v2.0 — complete)

| Phase | Name | Status | Notes |
|-------|------|--------|-------|
| 9 | Contract Design (OpenAPI + proto) | DONE | 6 journey endpoints, 9 RPCs, both lint-clean |
| 10 | f2c gRPC Shim (C++) | DONE | 9 RPCs, 18 integration tests passing |
| 11 | Go API Server Scaffold + Journey | DONE | 6 handlers, 88.8% coverage, graceful shutdown, CI |
| 12 | Shortcut Endpoints + E2E | DONE | 18/18 integration tests pass, SHC-03 contract check in CI |
| 13 | Dockerization + Smoke | DONE | Multi-stage Dockerfile, 359 MB image, compose up works |
| 14 | wodan Client Integration | DONE | Migrated to v2 /pipeline/plan-coverage, 30/30 jest tests |
| 15 | farmmaps Client Integration | DONE | Fields2CoverService.planCoverage() migrated, ng build clean |
| 16 | Python API Retirement | DONE | fields2cover-api/ deleted, legacy Dockerfile replaced |

## Key Architecture Decisions (v4.0)

- **New C++ algorithms** go in `include/` + `src/fields2cover/` — not wrappers around existing f2c APIs
- **Proto contract:** `proto/f2c.proto` → generates Go client (`api-go/internal/f2cclient/`) and C++ server (`f2c-grpc/`)
- **Go API:** `api-go/openapi.yaml` → `oapi-codegen` → typed handlers; frontend uses `npm run generate` for TypeScript types
- **Frontend generated types:** Always run `npm run generate` in `frontend/` after OpenAPI changes — never hand-write API types

## Key Architecture Decisions (v3.0)

- **Stack:** Vite 8 + React 18.3 + TypeScript — React 19 blocked by leaflet-geoman ecosystem
- **Draw library:** `@geoman-io/leaflet-geoman-free` via `useEffect` + `useMap()` — NOT react-leaflet-draw (abandoned)
- **Map:** `react-leaflet@4.2.1` + `leaflet@1.9.4` — leaflet 2 still alpha
- **HTTP client:** `ky@1.7.x` — ESM-native, no Redux/TanStack needed for this scope
- **SPA handler:** `statusRecorder` pattern in chi — intercepts file-server 404 before writing to real ResponseWriter
- **go:embed:** `//go:embed dist` in `api-go/internal/static/embed.go`; placeholder `dist/.gitkeep` + `dist/index.html` committed so `go build` works clean
- **PDOK CRS:** Always request `?crs=http://www.opengis.net/def/crs/OGC/1.3/CRS84` — default is EPSG:28992 (RD New) which displaces polygons off the Netherlands
- **No Redux/Zustand:** Three pages, `useState` + `useContext` is sufficient
- **No Tailwind:** PostCSS overhead not justified for a small demo

## Critical Pitfalls to Remember

1. Missing `leaflet/dist/leaflet.css` import in `main.tsx` — controls and icons break silently
2. `react-leaflet-draw` is abandoned — use `@geoman-io/leaflet-geoman-free` exclusively
3. PDOK returns EPSG:28992 by default — always append `?crs=...CRS84` query param
4. `go:embed` requires non-empty target dir at compile time — commit placeholder files
5. SPA `r.NotFound(serveIndex)` poisons API 404s — use `statusRecorder` + `r.Get("/*", spaHandler)` after `api.HandlerFromMux`
6. Frontend API types must always be generated (`npm run generate`) — never hand-written

## Accumulated Context (from v3.0)

- f2c C++ core stable, ASan-clean
- Go API on port 8080, chi router, `api.HandlerFromMux` entry point
- `POST /pipeline/plan-coverage` is the one-shot endpoint used by the UI
- Existing API paths: `/fields:parse`, `/headlands:generate`, `/swaths:generate`, `/swaths:sort`, `/routes:plan`, `/paths:plan`, `/pipeline/plan-coverage`, `/pipeline/resume`, `/healthz`, `/readyz`, `/openapi.yaml`, `/docs`
- Docker image: multi-stage, ubuntu:22.04 runtime, 359 MB, supervised bash entrypoint

## How to Resume

1. `/clear` (if continuing in same session)
2. `/gsd-resume-work` — loads this STATE.md and routes to next action
3. Or run directly: `/gsd-plan-phase 24`

## Session Continuity

Last session: 2026-04-29T00:00:00.000Z
Stopped at: roadmap created for v4.0
Resume file: None

## Next Step

`/gsd-plan-phase 24` to begin C++ Multi-Robot Partitioning.
