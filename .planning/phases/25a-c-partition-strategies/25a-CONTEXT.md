# Phase 25a: C++ Partition Strategies - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped via workflow.skip_discuss)

<domain>
## Phase Boundary

The f2c library gains two additional multi-robot field division strategies beyond the existing strip-based partitioner:
- `SpatialRtreePartition::partition()` — spatial proximity clustering: zones are geographically compact swath clusters
- `LengthBalancedPartition::partition()` — workload-balanced assignment: total swath length per robot balanced within 10%

Both strategies share the same `(F2CCells field, std::vector<F2CRobot> robots)` signature as the Phase 24 `MultiRobotPartition` partitioner. Pure C++ library work — no gRPC, no Go, no frontend changes.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — discuss phase was skipped per user setting. Use ROADMAP phase goal, success criteria, and codebase conventions to guide decisions.

Key guidance from ROADMAP:
- Follow Phase 24 `MultiRobotPartition` patterns for namespace, guard macros, module layout
- Namespace: `f2c::partition` (consistent with Phase 24)
- Module directories: `include/fields2cover/partition/` and `src/fields2cover/partition/`
- Tests: `tests/cpp/partition/` (auto-discovered by GLOB_RECURSE)
- For SpatialRtreePartition: use boost::geometry R-tree or a simple centroid-distance clustering approach if boost is unavailable
- For LengthBalancedPartition: compute swath lengths, sort, greedily assign to robots to balance total load

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `f2c::partition::MultiRobotPartition` — existing Phase 24 class as direct pattern to follow
- `F2CCells`, `F2CCell`, `F2CSwaths`, `F2CRobot` — types available via `fields2cover/types.h`
- Phase 24 GLOB_RECURSE pattern auto-discovers new files in `src/fields2cover/partition/`

### Established Patterns
- Module header: `include/fields2cover/partition/<class>.h` with `#pragma once` + `#ifndef` guard
- Module source: `src/fields2cover/partition/<class>.cpp` (auto-discovered by GLOB_RECURSE)
- Tests: `tests/cpp/partition/<suite>_test.cpp` with `TEST(suite_name, case_name)` format
- License header: BSD-3 Wageningen University block at top of each file
- No CMakeLists.txt edits needed

### Integration Points
- Phase 26 will expose both new strategies via gRPC
- Both must accept same signature as `MultiRobotPartition::partition(F2CCells, std::vector<F2CRobot>)`

</code_context>

<specifics>
## Specific Ideas

No specific requirements — discuss phase skipped. Refer to ROADMAP phase description and success criteria.

</specifics>

<deferred>
## Deferred Ideas

None — discuss phase skipped.

</deferred>
