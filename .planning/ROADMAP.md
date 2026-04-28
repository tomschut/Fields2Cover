# Roadmap: Fields2Cover

## Milestones

- ✅ **v1.0 Quality & Test** — Phases 1–8 (shipped 2026-04-11)
- ✅ **v2.0 Journey-First Go Rewrite** — Phases 9–16 (shipped 2026-04-15)
- ✅ **v3.0 Front-End** — Phases 17–23 (shipped 2026-04-29)
- 🔄 **v4.0 Multi-Robot Planning** — Phases 24–28 (in progress)

## Phases

<details>
<summary>✅ v1.0 Quality & Test (Phases 1–8) — SHIPPED 2026-04-11</summary>

8 phases: API code fixes, connexion 3.x upgrade, API unit test suite, route planner crash fixes, solver evaluation, f2c code quality audit, hardening (ASan/UBSan/clang-tidy/libFuzzer), GeoJSON parser. See git history.

</details>

<details>
<summary>✅ v2.0 Journey-First Go Rewrite (Phases 9–16) — SHIPPED 2026-04-15</summary>

8 phases: OpenAPI + proto contract design, f2c gRPC shim (C++), Go API server + journey handlers, shortcut endpoints + E2E, Dockerization, wodan client integration, farmmaps client integration, Python API retirement. See `.planning/archive/`.

</details>

<details>
<summary>✅ v3.0 Front-End (Phases 17–23) — SHIPPED 2026-04-29</summary>

- [x] Phase 17: Infrastructure Foundation (3/3 plans) — completed 2026-04-16
- [x] Phase 18: Map Shell & Field Input (2/2 plans) — completed 2026-04-16
- [x] Phase 19: Pipeline Integration (2/2 plans) — completed 2026-04-16
- [x] Phase 20: Docs Pages (2/2 plans) — completed 2026-04-19
- [x] Phase 21: PDOK Gewaspercelen Import (1/1 plan) — completed 2026-04-20
- [x] Phase 22: Start/End Point Selection (2/2 plans) — completed 2026-04-28
- [x] Phase 23: Start-Point Native Backend Support (2/2 plans) — completed 2026-04-28

Full archive: `.planning/milestones/v3.0-ROADMAP.md`

</details>

### v4.0 Multi-Robot Planning

- [ ] **Phase 24: C++ Multi-Robot Partitioning** — New f2c algorithm divides a field into N zones weighted by robot work rate
- [ ] **Phase 25: C++ Grain Cart Coordination** — New f2c algorithm computes cart travel path and headland rendezvous points alongside a robot route
- [ ] **Phase 26: gRPC RPCs + Go API Endpoints** — Wire both new algorithms through proto → Go API as two new REST endpoints
- [ ] **Phase 27: Multi-Robot Frontend** — Fleet configuration panel + multi-path map overlay in the React UI
- [ ] **Phase 28: Grain Cart Frontend** — Cart configuration panel + cart path and rendezvous overlays in the React UI

## Phase Details

### Phase 24: C++ Multi-Robot Partitioning
**Goal**: The f2c library can partition a field into N zones weighted by each robot's work rate
**Depends on**: Nothing (pure C++ library work, no API layer changes)
**Requirements**: F2C-01
**Success Criteria** (what must be TRUE):
  1. A call to the new partitioning function with a field polygon and a list of (width, speed) robot specs returns N non-overlapping zone geometries whose areas are proportional to each robot's work rate (width × speed)
  2. The total area of all returned zones equals the input field area (no coverage gap, no overlap)
  3. All existing GoogleTest unit tests continue to pass after the algorithm is added
  4. The new partitioning function has its own unit test covering at least a 2-robot and a 3-robot case
**Plans**: TBD

### Phase 25: C++ Grain Cart Coordination
**Goal**: The f2c library can compute a grain cart's travel path and headland rendezvous points from a robot route and cart specs
**Depends on**: Phase 24 (shares C++ library build context; establishes patterns for new algorithm additions)
**Requirements**: F2C-02
**Success Criteria** (what must be TRUE):
  1. A call to the new cart coordination function with a robot coverage route and cart specs (tank capacity, unload time, travel speed) returns a cart path geometry and a list of rendezvous point coordinates
  2. The number of rendezvous points is consistent with the robot's route length and the cart's tank capacity (more rendezvous when capacity is smaller)
  3. All existing GoogleTest unit tests continue to pass after the algorithm is added
  4. The new coordination function has its own unit test covering at least one rendezvous-triggering scenario and one no-rendezvous (large capacity) scenario
**Plans**: TBD

### Phase 26: gRPC RPCs + Go API Endpoints
**Goal**: Both new algorithms are reachable via REST — proto definitions, C++ gRPC server stubs, Go client, and typed OpenAPI endpoints are all wired and tested
**Depends on**: Phase 25 (both C++ algorithms must exist before they can be exposed via gRPC)
**Requirements**: API-01, API-02, API-03, API-04
**Success Criteria** (what must be TRUE):
  1. `POST /pipeline/plan-multi-robot` accepts a valid request (field geometry + N robot specs) and returns N zone geometries plus N coverage plans with HTTP 200
  2. `POST /pipeline/plan-cart` accepts a valid request (robot path + cart specs) and returns a cart path geometry plus a rendezvous point list with HTTP 200
  3. Both endpoints are declared in `api-go/openapi.yaml` with typed request and response schemas; `npm run generate` in `frontend/` produces updated TypeScript types without errors
  4. Integration tests covering both happy-path calls pass in CI (`docker compose build && docker compose up -d`)
**Plans**: TBD

### Phase 27: Multi-Robot Frontend
**Goal**: Users can configure a fleet of N robots and see all robot coverage paths in distinct colors on the map
**Depends on**: Phase 26 (requires the `/pipeline/plan-multi-robot` endpoint to exist)
**Requirements**: MRD-01, MRD-02, MRD-03, MRD-04, MRD-05
**Success Criteria** (what must be TRUE):
  1. User can add robots to a fleet panel, set each robot's working width, speed, headland width, route planner, and path planner independently
  2. User clicks a single "Run" button and the UI calls `POST /pipeline/plan-multi-robot`; all N robot paths appear on the map without a page reload
  3. Each robot's path is rendered in a distinct color that is visually distinguishable from all other robots and from existing single-robot result layers
  4. Each robot's zone boundary is shown as a distinct map overlay alongside its path
  5. The multi-robot result layers can be toggled independently of each other and of the single-robot pipeline layers
**Plans**: TBD
**UI hint**: yes

### Phase 28: Grain Cart Frontend
**Goal**: Users can configure a grain cart and see the cart's computed travel path and rendezvous markers as toggleable overlays on the map
**Depends on**: Phase 27 (cart planning is presented alongside multi-robot results; shares UI patterns established in Phase 27)
**Requirements**: GC-01, GC-02, GC-03, GC-04
**Success Criteria** (what must be TRUE):
  1. User can enter grain cart configuration (tank capacity in kg, unload time in seconds, cart travel speed in m/s) in a dedicated panel
  2. After a plan run that includes the cart, the cart's full travel path appears on the map as a distinct overlay with its own color
  3. Each headland rendezvous/unload point is marked on the map (e.g., a circle or pin marker at the computed coordinate)
  4. The cart path overlay and the rendezvous markers layer are each independently toggleable without affecting any robot path layers
**Plans**: TBD
**UI hint**: yes

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1–8. Quality & Test | v1.0 | 8/8 | Complete | 2026-04-11 |
| 9–16. Go Rewrite | v2.0 | 8/8 | Complete | 2026-04-15 |
| 17. Infrastructure Foundation | v3.0 | 3/3 | Complete | 2026-04-16 |
| 18. Map Shell & Field Input | v3.0 | 2/2 | Complete | 2026-04-16 |
| 19. Pipeline Integration | v3.0 | 2/2 | Complete | 2026-04-16 |
| 20. Docs Pages | v3.0 | 2/2 | Complete | 2026-04-19 |
| 21. PDOK Gewaspercelen Import | v3.0 | 1/1 | Complete | 2026-04-20 |
| 22. Start/End Point Selection | v3.0 | 2/2 | Complete | 2026-04-28 |
| 23. Start-Point Native Backend | v3.0 | 2/2 | Complete | 2026-04-28 |
| 24. C++ Multi-Robot Partitioning | v4.0 | 0/? | Not started | - |
| 25. C++ Grain Cart Coordination | v4.0 | 0/? | Not started | - |
| 26. gRPC RPCs + Go API Endpoints | v4.0 | 0/? | Not started | - |
| 27. Multi-Robot Frontend | v4.0 | 0/? | Not started | - |
| 28. Grain Cart Frontend | v4.0 | 0/? | Not started | - |
