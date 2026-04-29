# Phase 25b: C++ Graph Route Optimizer - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning

<domain>
## Phase Boundary

The f2c library gains a Nety-style graph-based swath traversal optimizer (`GraphRouteOptimizer`) that minimizes total travel distance between swaths using direction-aware greedy nearest-neighbor scoring.

Placed in the `route_planning/` module, namespace `f2c::rp`, inheriting from `SingleCellSwathsOrderBase` — integrates with `RoutePlannerBase` and follows the `BoustrophedonOrder`/`SnakeOrder` pattern. Pure C++ library work — no gRPC, no Go, no frontend changes.

</domain>

<decisions>
## Implementation Decisions

### Module Placement & Interface
- Module: `route_planning/` — alongside existing sorters; auto-discovered by GLOB_RECURSE; no CMakeLists.txt edits
- Namespace: `f2c::rp` — consistent with `BoustrophedonOrder`, `SnakeOrder`, `SpiralOrder`
- Inherits `SingleCellSwathsOrderBase` — override `sortSwaths(F2CSwaths&)` so it integrates with `RoutePlannerBase`
- Public entry point: `genSortedSwaths()` via base class — ROADMAP "optimize()" is conceptual intent, not a distinct method
- Direction flip: YES — each swath may be traversed in either direction; choose entry endpoint that minimises cost to previous swath endpoint (same as `redirect_swaths` logic in base)

### Core Algorithm
- Algorithm: Greedy nearest-neighbor (O(n²)) — deterministic, no external dependencies, practical for n < 1000 swaths
- Starting swath: try all n starts, keep the globally cheapest result — for typical field sizes (n < 100) the O(n³) cost is negligible
- Direction penalty: additive — when the angle between consecutive swath directions > 30°, add `penalty_weight * distance` to the edge cost; default `penalty_weight = 0.5`
- Parallel threshold: 30° — prefer clearly parallel swaths as neighbors; uses dot product of direction vectors to determine angle

### Claude's Discretion
- Penalty weight (0.5) exposed as constructor parameter with default, allowing future tuning
- Graph construction: fully connected (all-pairs endpoint distances computed on-the-fly, no adjacency matrix stored)
- No external dependencies beyond existing f2c types and std library

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `SingleCellSwathsOrderBase` — base class to inherit; `genSortedSwaths()` handles variant loop + `changeStartPoint()`; only `sortSwaths(F2CSwaths&)` needs implementation
- `F2CSwaths`, `F2CPoint`, `F2CLineString` — geometry types via `fields2cover/types.h`
- `f2c::obj::DirectDistPathObj` — available for cost verification in tests; `computeCost(F2CSwaths&)` returns total endpoint-to-endpoint distance
- License header: BSD-3 Wageningen University block (identical to all other route_planning files)

### Established Patterns
- Header: `include/fields2cover/route_planning/graph_route_optimizer.h` with `#pragma once` + `#ifndef` guard
- Source: `src/fields2cover/route_planning/graph_route_optimizer.cpp` (auto-discovered by GLOB_RECURSE)
- Test: `tests/cpp/route_planning/graph_route_optimizer_test.cpp` with `TEST(fields2cover_route_graph, ...)` format
- No CMakeLists.txt edits needed — GLOB_RECURSE already covers both directories

### Integration Points
- `RoutePlannerBase::genRoute()` accepts a `SingleCellSwathsOrderBase*` — `GraphRouteOptimizer` becomes a drop-in ordering strategy
- Phase 26 will expose this optimizer through gRPC when wiring all algorithms to the API

</code_context>

<specifics>
## Specific Ideas

- `sortSwaths()` implementation: build all-pairs edge costs (endpoint distance + direction penalty), then greedy nearest-neighbor from each possible start, return the globally best ordering
- Direction penalty: `angle = acos(|dot(dir_i, dir_j)|)` where `dir` is the unit vector along each swath; penalty applied when `angle > 30° (π/6 radians)`
- Swath direction flip: for each candidate next swath, test both entry points and pick the cheaper entry; flip the swath in-place if the reverse direction is chosen

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>
