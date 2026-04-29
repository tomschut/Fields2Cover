---
phase: 25c-c-obstacle-avoider
plan: 01
subsystem: cpp-library
tags: [cpp, ogr, gdal, geometry, swath, obstacle, gtest, fields2cover]

# Dependency graph
requires:
  - phase: 25a-spatial-and-length-partitioning
    provides: "f2c::partition module pattern (three-file header/source/test)"
  - phase: 25b-graph-route-optimizer
    provides: "obstacle/ module namespace pattern"
provides:
  - "f2c::obstacle::ObstacleAvoider class — avoid() fragments swaths around inflated polygon obstacle"
  - "5 GoogleTest unit tests covering split, margin enlargement, short-segment filter, throw, and no-overlap cases"
affects:
  - phase-26-grpc-rpcs
  - phase-27-multi-robot-frontend

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Cell::buffer + Cells::difference + Cells::getLinesInside idiom for safe-zone clipping"
    - "Bounding-box padding formula: safety_margin + 1.0 (1 m floor ensures bbox is never swallowed)"
    - "Sequential ID reassignment when one swath splits into multiple output segments"

key-files:
  created:
    - include/fields2cover/obstacle/obstacle_avoider.h
    - src/fields2cover/obstacle/obstacle_avoider.cpp
    - tests/cpp/obstacle/obstacle_avoider_test.cpp
  modified: []

key-decisions:
  - "Bbox padding = safety_margin + 1.0: 1 m floor ensures bbox is not swallowed by obstacle even when safety_margin=0"
  - "Sequential out_id counter reassigns IDs after swath splits; original width and type are always preserved"
  - "All geometry ops via f2c type layer (Cell::buffer, Cells::difference, getLinesInside) — no raw OGRGeometry* usage"
  - "Single-obstacle form (not std::vector<F2CCell>): caller can compose by invoking avoid() multiple times"
  - "Worktree requires its own cmake build (-DBUILD_TESTING=ON) since main repo build does not pick up worktree source files"

patterns-established:
  - "Pattern: obstacle module layout mirrors partition module exactly — include/fields2cover/{module}/, src/fields2cover/{module}/, tests/cpp/{module}/"
  - "Pattern: safe-zone = bounding-box minus inflated obstacle via Cells::difference, then getLinesInside clips each swath"

requirements-completed: [F2C-05]

# Metrics
duration: 25min
completed: 2026-04-29
---

# Phase 25c Plan 01: ObstacleAvoider Summary

**Cell::buffer + Cells::difference + getLinesInside idiom fragments swaths around inflated obstacles, with 0.1 m minimum-segment filter and negative-margin guard — 5/5 tests pass, 318 total green**

## Performance

- **Duration:** 25 min
- **Started:** 2026-04-29T08:30:00Z
- **Completed:** 2026-04-29T08:55:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Implemented `f2c::obstacle::ObstacleAvoider::avoid()` — inflates obstacle with `Cell::buffer`, builds safe region via `Cells::difference`, clips each swath via `getLinesInside`, drops segments < 0.1 m
- Added 5 GoogleTest unit tests covering all plan success criteria (split confirmation, margin enlargement, short-segment filter, throw on negative margin, full-swath pass-through)
- 318 total tests pass — 313 pre-existing + 5 new obstacle tests; zero regressions

## Task Commits

Each task was committed atomically:

1. **Task 1: Create ObstacleAvoider header and implementation** - `fbea55a` (feat)
2. **Task 2: Write GoogleTest unit tests for ObstacleAvoider** - `8d4504e` (test)

**Plan metadata:** committed with SUMMARY

## Files Created/Modified

- `include/fields2cover/obstacle/obstacle_avoider.h` — ObstacleAvoider class declaration in namespace f2c::obstacle with kMinSegmentLength=0.1 constexpr
- `src/fields2cover/obstacle/obstacle_avoider.cpp` — avoid() implementation using Cell::buffer, Cells::difference, getLinesInside; throws std::invalid_argument for negative safety_margin
- `tests/cpp/obstacle/obstacle_avoider_test.cpp` — 5 GoogleTest tests with helper functions makeStraightSwaths and makeSquareObstacle

## Decisions Made

- **Bbox padding formula:** `safety_margin + 1.0` — the 1 m floor ensures the bounding box is never enclosed by the inflated obstacle even when safety_margin=0 (prevents `difference()` returning empty Cells)
- **Sequential ID reassignment:** `out_id` counter increments per output segment; original swath `width` and `type` are always preserved when emitting split segments
- **Single-obstacle API:** `avoid(swaths, obstacle, margin)` takes one obstacle per call — composability (call avoid() multiple times for multiple obstacles) avoids API complexity
- **No raw OGR pointers:** All geometry operations go through f2c wrapper types (F2CCell, F2CCells, F2CMultiLineString) — consistent with all other f2c modules
- **Worktree build:** Required separate `cmake -S . -B build -DBUILD_TESTING=ON` in the worktree directory because the main repo build directory targets `/home/tom/devenv/fields2cover/src/` not the worktree source

## Deviations from Plan

None — plan executed exactly as written. The worktree build setup (creating a local build directory with BUILD_TESTING=ON) was a necessary environment action, not a code deviation.

## Issues Encountered

- **Worktree cmake isolation:** The existing build at `/home/tom/devenv/fields2cover/build` uses GLOB_RECURSE pointing to the main repo, not the worktree. Created `/home/tom/devenv/fields2cover/.claude/worktrees/agent-aa217292/build` with `-DBUILD_TESTING=ON` to compile and test the worktree files. Build completed without errors; all 318 tests pass.
- **Pre-existing linker error in main repo build** (`f2c-grpc-utils-tests`: undefined reference to vtable for F2CServiceImpl): this error is pre-existing and out of scope — not caused by this plan. The worktree build did not reproduce it.

## User Setup Required

None — no external service configuration required. C++ library change only.

## Next Phase Readiness

- `ObstacleAvoider::avoid()` is ready for Phase 26 gRPC exposure — method signature `avoid(F2CSwaths, F2CCell, double)` is stable
- F2C-05 fully satisfied: split verification, short-segment filter, and no-regression criteria all green
- No blockers

---
*Phase: 25c-c-obstacle-avoider*
*Completed: 2026-04-29*
