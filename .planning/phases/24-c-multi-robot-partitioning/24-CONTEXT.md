# Phase 24: C++ Multi-Robot Partitioning - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

The f2c library gains a new `MultiRobotPartition` class in namespace `f2c::partition` that partitions a field (`F2CCells`) into N non-overlapping zones whose areas are proportional to each robot's work rate (`getCovWidth() × getCruiseVel()`). Pure C++ library work — no gRPC, no Go, no frontend changes in this phase.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase. The plan already specifies axis-aligned strip partitioning via GEOS intersection. Follow the plan exactly as written in 24-01-PLAN.md.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `F2CCells::intersection(const F2CCell&)` — GEOS-backed polygon intersection already available
- `F2CCells::getDimMinX/MaxX/MinY/MaxY()` — bounding box helpers already available
- `F2CRobot::getCovWidth()`, `F2CRobot::getCruiseVel()` — work-rate inputs already available
- `F2CLinearRing{F2CPoint(...), ...}` — brace-init ring construction

### Established Patterns
- Module header: `include/fields2cover/<module>/<class>.h` with `#pragma once` + `#ifndef` guard
- Module source: `src/fields2cover/<module>/<class>.cpp` (auto-picked up by GLOB_RECURSE)
- Tests: `tests/cpp/<module>/<suite>_test.cpp` with `TEST(suite_name, case_name)` format
- Namespace: `f2c::<module>` (e.g., `f2c::partition`)
- License header: BSD-3 Wageningen University block at top of each file

### Integration Points
- Build: `file(GLOB_RECURSE ... src/*.cpp)` auto-discovers new .cpp — no CMakeLists.txt edits needed
- Tests: `file(GLOB_RECURSE TEST_SOURCES ... cpp/*/*.cpp)` auto-discovers new test — no edits needed
- Phase 26 will consume the `f2c::partition::MultiRobotPartition::partition()` API via gRPC

</code_context>

<specifics>
## Specific Ideas

See 24-01-PLAN.md for the exact implementation (header, source, tests fully spelled out). Follow it as written.

</specifics>

<deferred>
## Deferred Ideas

None — infrastructure phase. No out-of-scope ideas surfaced.

</deferred>
