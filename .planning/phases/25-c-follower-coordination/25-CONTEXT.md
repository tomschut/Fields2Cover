# Phase 25: C++ Follower Coordination - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

The f2c library gains a new `FollowerCoordination` class (or similar name, at Claude's discretion) that computes a follower's travel path and headland rendezvous points given a robot coverage route and follower specs (tank capacity, unload time, travel speed). Pure C++ library work — no gRPC, no Go, no frontend changes in this phase.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase. Key concerns to address during planning:
- Namespace: follow Phase 24 pattern → `f2c::follower` or `f2c::coordination`
- Rendezvous triggering: accumulate robot route length until follower tank would be full → trigger rendezvous at nearest headland point
- Follower path: connect rendezvous points with straight-line travel segments at follower speed
- Output: follower path geometry (`F2CCells` or `F2CLineString`) + list of rendezvous coordinates (`std::vector<F2CPoint>`)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `F2CPath` — robot coverage route output type (what Phase 25 takes as input)
- `F2CRobot` — robot specs (getCovWidth, getCruiseVel already used in Phase 24)
- `F2CPoint` — coordinate type for rendezvous points
- `F2CCell`, `F2CCells` — geometry types
- Phase 24 pattern: `f2c::partition::MultiRobotPartition` as the module/class naming model

### Established Patterns
- Module header: `include/fields2cover/<module>/<class>.h` with `#pragma once` + `#ifndef` guard
- Module source: `src/fields2cover/<module>/<class>.cpp` (auto-picked up by GLOB_RECURSE)
- Tests: `tests/cpp/<module>/<suite>_test.cpp` with `TEST(suite_name, case_name)` format
- Namespace: `f2c::<module>` (e.g., `f2c::follower`)
- License header: BSD-3 Wageningen University block at top of each file

### Integration Points
- Build: `file(GLOB_RECURSE ... src/*.cpp)` auto-discovers new .cpp — no CMakeLists.txt edits needed
- Tests: `file(GLOB_RECURSE TEST_SOURCES ... cpp/*/*.cpp)` auto-discovers new test — no edits needed
- Phase 26 will consume the new coordination API via gRPC

</code_context>

<specifics>
## Specific Ideas

Research the f2c type system carefully before planning — understand F2CPath structure, how route segments are stored, and how to compute cumulative distance along the robot route. The rendezvous logic depends on this.

</specifics>

<deferred>
## Deferred Ideas

None — infrastructure phase. No out-of-scope ideas surfaced.

</deferred>
