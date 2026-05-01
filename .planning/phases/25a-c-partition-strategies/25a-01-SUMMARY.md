---
phase: 25a-c-partition-strategies
plan: "01"
subsystem: fields2cover-cpp
tags: [cpp, partition, spatial, rtree, boost-geometry, length-balanced, googletest]
dependency_graph:
  requires: [phase-24-multi-robot-partition]
  provides: [SpatialRtreePartition, LengthBalancedPartition, f2c::partition namespace]
  affects: [phase-26-grpc-rpcs]
tech_stack:
  added: [boost::geometry R-tree (header-only, no new link deps)]
  patterns: [swath-cluster-zone pipeline, FFD greedy bin-packing, GLOB_RECURSE auto-discovery]
key_files:
  created:
    - include/fields2cover/partition/spatial_rtree_partition.h
    - src/fields2cover/partition/spatial_rtree_partition.cpp
    - include/fields2cover/partition/length_balanced_partition.h
    - src/fields2cover/partition/length_balanced_partition.cpp
    - tests/cpp/partition/spatial_rtree_partition_test.cpp
    - tests/cpp/partition/length_balanced_partition_test.cpp
  modified: []
decisions:
  - "Use robots[0].getCovWidth() for internal swath generation (uniform-width fleet assumption; documented in headers)"
  - "Field must be single-cell (field.size()==1); throws invalid_argument otherwise; Phase 26 must pre-validate"
  - "boost::geometry R-tree headers confined to .cpp only — no header contamination"
  - "zero_width_throws tests use setCovWidth(0.0) after construction because F2CRobot(0.0) itself throws"
metrics:
  duration_minutes: 15
  completed: "2026-04-29"
  tasks_completed: 2
  files_created: 6
requirements: [F2C-03]
---

# Phase 25a Plan 01: SpatialRtreePartition and LengthBalancedPartition Summary

**One-liner:** Two new `f2c::partition` strategies — R-tree spatial clustering and FFD greedy length-balanced assignment — both returning `std::vector<F2CCells>` from swath-level field decomposition.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement SpatialRtreePartition and LengthBalancedPartition headers and sources | fa3a640 | 4 files created |
| 2 | Write GoogleTest unit tests for both partition strategies | 7d59f0c | 2 test files created |

## What Was Built

### SpatialRtreePartition (`include/fields2cover/partition/spatial_rtree_partition.h` + `.cpp`)
- Generates swaths from the single-cell field via `f2c::sg::BruteForce`
- Builds a `boost::geometry::index::rtree<quadratic<16>>` over all swath midpoints
- Selects N evenly-spaced seed swaths (one per robot)
- Assigns remaining swaths to the robot with the nearest current cluster centroid
- Unions each robot's swath `areaCovered()` polygons into zone geometry
- boost::geometry includes confined to `.cpp` only (no downstream header contamination)

### LengthBalancedPartition (`include/fields2cover/partition/length_balanced_partition.h` + `.cpp`)
- Generates swaths from the single-cell field via `f2c::sg::BruteForce`
- Sorts swaths by `length()` descending (first-fit decreasing)
- Greedily assigns each swath to the robot with minimum current load via `std::min_element`
- Unions each robot's swath `areaCovered()` polygons into zone geometry
- No external dependencies beyond `<algorithm>` and `<numeric>`

### Both classes
- Namespace: `f2c::partition` (consistent with Phase 24 `MultiRobotPartition`)
- Guard pattern: `#pragma once` + `#ifndef FIELDS2COVER_PARTITION_*_H_`
- BSD-3 Wageningen University license header
- Input validation: throws `std::invalid_argument` for empty robots, zero/negative width, multi-cell field
- No CMakeLists.txt edits — GLOB_RECURSE auto-discovers all new files

### Tests (8 new GoogleTest cases)
- `fields2cover_partition_spatial_rtree`: two_robots, three_robots, empty_robots_throws, zero_width_throws
- `fields2cover_partition_length_balanced`: two_robots, three_robots, empty_robots_throws, zero_width_throws

## Test Results

```
[  PASSED  ] 313 tests.   (305 pre-existing + 8 new)
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] F2CRobot(0.0) constructor throws before partition() guard**
- **Found during:** Task 2, running zero_width_throws tests
- **Issue:** `F2CRobot r_zero(0.0)` itself throws `"Robot widths have to be greater than 0."` — the test body never reaches `part.partition()`, causing EXPECT_THROW to fail (wrong exception origin)
- **Fix:** Changed test construction to `F2CRobot r_zero(1.0); r_zero.setCovWidth(0.0);` — mirrors the Phase 24 `multi_robot_partition_test.cpp` zero_work_rate_throws pattern exactly
- **Files modified:** `tests/cpp/partition/spatial_rtree_partition_test.cpp`, `tests/cpp/partition/length_balanced_partition_test.cpp`
- **Commit:** 7d59f0c

## Known Stubs

None — both strategies produce real zone geometry from actual swath generation and GEOS union operations.

## Threat Flags

No new network endpoints, auth paths, file access patterns, or schema changes introduced. Pure C++ library computation — no trust boundary surface added in this plan. The threat model in the plan (T-25a-01 through T-25a-06) is fully mitigated by input validation guards.

## Self-Check: PASSED

- `include/fields2cover/partition/spatial_rtree_partition.h` — FOUND
- `src/fields2cover/partition/spatial_rtree_partition.cpp` — FOUND
- `include/fields2cover/partition/length_balanced_partition.h` — FOUND
- `src/fields2cover/partition/length_balanced_partition.cpp` — FOUND
- `tests/cpp/partition/spatial_rtree_partition_test.cpp` — FOUND
- `tests/cpp/partition/length_balanced_partition_test.cpp` — FOUND
- Commit fa3a640 — FOUND
- Commit 7d59f0c — FOUND
- 313 tests PASSED, 0 failures — VERIFIED
