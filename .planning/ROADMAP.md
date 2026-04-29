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

- [x] **Phase 24: C++ Multi-Robot Partitioning** — New f2c algorithm divides a field into N zones weighted by robot work rate (completed 2026-04-29)
- [x] **Phase 25: C++ Follower Coordination** — New f2c algorithm computes cart travel path and headland rendezvous points alongside a robot route (completed 2026-04-29)
- [x] **Phase 25a: C++ Partition Strategies** — SPATIAL_RTREE and LENGTH_BALANCED multi-robot division strategies (inspired by farmtrax Divy) (completed 2026-04-29)
- [ ] **Phase 25b: C++ Graph Route Optimizer** — Direction-aware Dijkstra swath ordering to minimise travel distance (inspired by farmtrax Nety)
- [x] **Phase 25c: C++ Obstacle Avoider** — Fragment swaths around inflated polygon obstacles (inspired by farmtrax ObstacleAvoider) (completed 2026-04-29)
- [ ] **Phase 26: gRPC RPCs + Go API Endpoints** — Wire all five new C++ algorithms through proto → Go API as REST endpoints
- [ ] **Phase 27: Multi-Robot Frontend** — Fleet configuration panel + multi-path map overlay in the React UI
- [ ] **Phase 28: Follower Frontend** — Cart configuration panel + follower path and rendezvous overlays in the React UI

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
**Plans**: 1 plan
Plans:
- [x] 24-01-PLAN.md — Implement MultiRobotPartition class (header + source) and unit tests

### Phase 25: C++ Follower Coordination
**Goal**: The f2c library can compute a follower's travel path and headland rendezvous points from a robot route and follower specs
**Depends on**: Phase 24 (shares C++ library build context; establishes patterns for new algorithm additions)
**Requirements**: F2C-02
**Success Criteria** (what must be TRUE):
  1. A call to the new cart coordination function with a robot coverage route and follower specs (tank capacity, unload time, travel speed) returns a follower path geometry and a list of rendezvous point coordinates
  2. The number of rendezvous points is consistent with the robot's route length and the cart's tank capacity (more rendezvous when capacity is smaller)
  3. All existing GoogleTest unit tests continue to pass after the algorithm is added
  4. The new coordination function has its own unit test covering at least one rendezvous-triggering scenario and one no-rendezvous (large capacity) scenario
**Plans**: 1 plan
Plans:
- [x] 25-01-PLAN.md — Implement FollowerCoordination class (header + source) and unit tests

### Phase 25a: C++ Partition Strategies
**Goal**: The f2c library gains two additional multi-robot field division strategies beyond the existing strip-based partitioner — spatial proximity clustering (SPATIAL_RTREE) and workload-balanced assignment (LENGTH_BALANCED)
**Depends on**: Phase 24 (extends MultiRobotPartition patterns)
**Requirements**: F2C-03
**Success Criteria** (what must be TRUE):
  1. A call to `SpatialRtreePartition::partition()` returns N zones where each zone is a spatially contiguous cluster of swaths assigned to one robot (geographically compact, not scattered strips)
  2. A call to `LengthBalancedPartition::partition()` returns N zones where total swath length per robot is balanced within 10% of each other across all robots
  3. Both strategies accept the same `(F2CCells field, std::vector<F2CRobot> robots)` signature as the Phase 24 partitioner
  4. All existing GoogleTest unit tests continue to pass; each new strategy has its own unit test covering a 2-robot and 3-robot case
**Plans**: 1 plan
Plans:
- [ ] 25a-01-PLAN.md — Implement SpatialRtreePartition and LengthBalancedPartition (headers + sources + unit tests)

### Phase 25b: C++ Graph Route Optimizer
**Goal**: The f2c library gains a Nety-style graph-based swath traversal optimizer that minimizes total travel distance between swaths using direction-aware Dijkstra scoring
**Depends on**: Phase 25 (builds on C++ library patterns)
**Requirements**: F2C-04
**Success Criteria** (what must be TRUE):
  1. A call to `GraphRouteOptimizer::optimize()` with a set of swaths returns a reordered swath sequence whose total endpoint-to-endpoint travel distance is less than or equal to the naive sequential ordering
  2. The optimizer applies direction penalties (parallel connections score lower than crossing connections) so naturally parallel swaths are preferred as neighbors
  3. All existing GoogleTest unit tests continue to pass; the optimizer has its own unit test on a ≥4-swath field demonstrating improved ordering over sequential
**Plans**: 1 plan
Plans:
- [ ] 25b-01-PLAN.md — Implement GraphRouteOptimizer (header + source + unit tests)

### Phase 25c: C++ Obstacle Avoider
**Goal**: The f2c library can fragment swaths around polygon obstacles, producing trimmed swath segments that avoid inflated obstacle boundaries
**Depends on**: Phase 25 (builds on C++ library patterns)
**Requirements**: F2C-05
**Success Criteria** (what must be TRUE):
  1. A call to `ObstacleAvoider::avoid()` with a list of swaths and a polygon obstacle returns swath segments with no overlap with the inflated obstacle polygon (inflation = safety margin parameter)
  2. Swath segments shorter than a minimum threshold (0.1 m) are dropped rather than returned
  3. All existing GoogleTest unit tests continue to pass; the avoider has its own unit test with a known obstacle placement that splits at least one swath into two segments
**Plans**: 1 plan
Plans:
- [x] 25c-01-PLAN.md — Implement ObstacleAvoider (header + source + unit tests)

### Phase 26: gRPC RPCs + Go API Endpoints
**Goal**: All five new C++ algorithms are reachable via REST — proto definitions, C++ gRPC server stubs, Go client, and typed OpenAPI endpoints are all wired and tested
**Depends on**: Phase 25c (all C++ algorithms must exist before gRPC exposure)
**Requirements**: API-01, API-02, API-03, API-04, API-05, API-06, API-07
**Success Criteria** (what must be TRUE):
  1. `POST /pipeline/plan-multi-robot` accepts a valid request (field geometry + N robot specs + optional partition strategy) and returns N zone geometries plus N coverage plans with HTTP 200
  2. `POST /pipeline/plan-follower` accepts a valid request (robot path + follower specs) and returns a follower path geometry plus a rendezvous point list with HTTP 200
  3. `POST /pipeline/optimize-route` accepts swaths and returns an optimized swath ordering with HTTP 200
  4. `POST /pipeline/avoid-obstacles` accepts swaths + obstacle polygons + safety margin and returns trimmed swath segments with HTTP 200
  5. All endpoints are declared in `api-go/openapi.yaml` with typed schemas; `npm run generate` produces updated TypeScript types without errors
  6. Integration tests covering all happy-path calls pass in CI (`docker compose build && docker compose up -d`)
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

### Phase 28: Follower Frontend
**Goal**: Users can configure a follower and see the follower's computed travel path and rendezvous markers as toggleable overlays on the map
**Depends on**: Phase 27 (cart planning is presented alongside multi-robot results; shares UI patterns established in Phase 27)
**Requirements**: FOL-01, FOL-02, FOL-03, FOL-04
**Success Criteria** (what must be TRUE):
  1. User can enter follower configuration (tank capacity in kg, unload time in seconds, follower travel speed in m/s) in a dedicated panel
  2. After a plan run that includes the cart, the cart's full travel path appears on the map as a distinct overlay with its own color
  3. Each headland rendezvous/unload point is marked on the map (e.g., a circle or pin marker at the computed coordinate)
  4. The follower path overlay and the rendezvous markers layer are each independently toggleable without affecting any robot path layers
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
| 24. C++ Multi-Robot Partitioning | v4.0 | 1/1 | Complete | 2026-04-29 |
| 25. C++ Follower Coordination | v4.0 | 1/1 | Complete   | 2026-04-29 |
| 25a. C++ Partition Strategies | v4.0 | 1/1 | Complete | 2026-04-29 |
| 25b. C++ Graph Route Optimizer | v4.0 | 0/1 | Not started | - |
| 25c. C++ Obstacle Avoider | v4.0 | 1/1 | Complete   | 2026-04-29 |
| 26. gRPC RPCs + Go API Endpoints | v4.0 | 0/? | Not started | - |
| 27. Multi-Robot Frontend | v4.0 | 0/? | Not started | - |
| 28. Follower Frontend | v4.0 | 0/? | Not started | - |
