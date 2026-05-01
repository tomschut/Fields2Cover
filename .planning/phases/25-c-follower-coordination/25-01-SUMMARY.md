---
phase: 25-c-follower-coordination
plan: 01
subsystem: cpp-library
tags: [cpp, googletest, fields2cover, follower-coordination, distance-accumulation]

# Dependency graph
requires:
  - phase: 24-c-multi-robot-partitioning
    provides: "f2c::partition::MultiRobotPartition pattern (header/source/test structure, namespace, guard format)"

provides:
  - "f2c::follower::FollowerCoordination class with compute() in include/fields2cover/follower/"
  - "Greedy distance-accumulation algorithm: SWATH distance -> HL_SWATH rendezvous snap"
  - "7 GoogleTest cases covering empty path, zero/negative capacity, large tank, small tank, scaling, headland discrimination"
  - "F2C-02 satisfied: library computes follower cart travel path geometry and headland rendezvous points"

affects:
  - 26-grpc-rpcs-go-api  # will call FollowerCoordination::compute via gRPC
  - 28-follower-frontend  # will visualize follower path and rendezvous points

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "f2c::follower namespace — new module pattern in include/fields2cover/follower/ + src/fields2cover/follower/"
    - "FollowerSpec inner struct for algorithm inputs (mirrors Phase 24 MultiRobotPartition param style)"
    - "Result inner struct with F2CLineString path + std::vector<F2CPoint> rendezvous_pts"
    - "Greedy outer-loop i=j advance after rendezvous to prevent off-by-one double-counting"

key-files:
  created:
    - include/fields2cover/follower/follower_coordination.h
    - src/fields2cover/follower/follower_coordination.cpp
    - tests/cpp/follower/follower_coordination_test.cpp
  modified: []

key-decisions:
  - "Tank capacity in swath-metres (not kg): unit-consistent with F2CPath.len; Phase 26 converts kg via yield_per_metre"
  - "Rendezvous snap to HL_SWATH .point (start of headland swath segment), not .atEnd()"
  - "No-HL_SWATH fallback: use last path point as rendezvous and break outer loop to avoid O(n^2)"
  - "i = j advance inside inner scan-forward loop; for-loop increment brings to j+1 (off-by-one guard)"

patterns-established:
  - "Pattern: follower module follows exact structure of partition module (one header, one source, one test, GLOB_RECURSE auto-discovery)"
  - "Pattern: cmake re-run required when new test directories are created (GLOB_RECURSE does not auto-detect mid-session)"

requirements-completed: [F2C-02]

# Metrics
duration: 2min
completed: 2026-04-29
---

# Phase 25 Plan 01: C++ Follower Coordination Summary

**FollowerCoordination class in f2c::follower namespace: greedy SWATH-distance accumulation with HL_SWATH rendezvous snapping, 7 GoogleTest cases, 305/305 tests green**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-29T07:59:18Z
- **Completed:** 2026-04-29T08:01:08Z
- **Tasks:** 2
- **Files created:** 3

## Accomplishments

- Implemented `FollowerCoordination::compute()` — iterates F2CPath, accumulates SWATH `.len`, triggers rendezvous at next `HL_SWATH` state when threshold crossed, validates inputs, handles empty path and no-HL_SWATH fallback
- Wrote 7 GoogleTest cases covering all behavioral specifications: empty path, zero/negative capacity guard, large tank (no rendezvous), small tank (2 rendezvous), capacity scaling, headland vs. turn discrimination
- All 305 tests pass (298 pre-existing + 7 new), zero regressions; F2C-02 requirement satisfied

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement FollowerCoordination header and source** - `fb7c0f1` (feat)
2. **Task 2: Write GoogleTest unit tests for FollowerCoordination** - `d61fd57` (test)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified

- `include/fields2cover/follower/follower_coordination.h` — class FollowerCoordination in namespace f2c::follower with FollowerSpec and Result inner structs, compute() declaration
- `src/fields2cover/follower/follower_coordination.cpp` — greedy distance-accumulation implementation: SWATH len accumulator, HL_SWATH snap, empty-path guard, invalid_argument throw on capacity <= 0, no-HL_SWATH fallback to last point
- `tests/cpp/follower/follower_coordination_test.cpp` — 7 GoogleTest cases; buildCyclicPath() helper builds SWATH+TURN+HL_SWATH cyclic synthetic paths

## Decisions Made

- **Tank capacity units:** Used `tank_capacity_swath_m` (pure distance threshold in swath-metres) — avoids needing a yield density constant in C++; Phase 26 can do kg -> swath-metre conversion at the gRPC layer before calling compute()
- **Rendezvous snap:** Use `HL_SWATH .point` (start of headland segment), not `.atEnd()` — follower arrives as robot enters headland
- **No-HL_SWATH fallback:** Fall back to last path point and break outer loop — prevents O(n^2) rescan of all remaining states after every SWATH
- **Loop advance:** Set `i = j` inside inner scan-forward loop; for-loop `++i` brings outer index to `j+1` — prevents double-counting the rendezvous segment distance

## Deviations from Plan

None — plan executed exactly as written.

Note: cmake re-run was needed after creating the new `tests/cpp/follower/` directory because GLOB_RECURSE evaluates at configure time. This is expected cmake behavior (documented in Phase 24 research notes) — not a deviation.

## Issues Encountered

- **cmake GLOB_RECURSE not auto-detecting new test directory:** After writing the test file and building, `--gtest_filter=fields2cover_follower_coordination*` matched 0 tests. Root cause: cmake GLOB_RECURSE runs at configure time; new directories require `cmake ..` before they appear. Fixed by running `cmake .. && make unittests` — tests appeared and all 7 passed. No code changes were required.

## Known Stubs

None — all functionality is fully implemented. No hardcoded empty values, no placeholder text.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes. This is a pure in-process C++ computation — no external trust boundary in this phase.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Phase 26 (gRPC RPCs + Go API Endpoints) can call `f2c::follower::FollowerCoordination::compute()` via the gRPC shim
- Input type is `F2CPath` (already serialized in existing gRPC endpoints); output is `F2CLineString` + `std::vector<F2CPoint>`
- Phase 26 should perform kg -> swath-metre conversion using `yield_per_metre_kg_m` before calling compute()

---
*Phase: 25-c-follower-coordination*
*Completed: 2026-04-29*
