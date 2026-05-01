---
phase: 25c
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - include/fields2cover/obstacle/obstacle_avoider.h
  - src/fields2cover/obstacle/obstacle_avoider.cpp
  - tests/cpp/obstacle/obstacle_avoider_test.cpp
autonomous: true
requirements:
  - F2C-05

must_haves:
  truths:
    - "ObstacleAvoider::avoid() with a swath and an intersecting obstacle returns segments
       that do not overlap with the inflated obstacle polygon"
    - "Swath segments shorter than 0.1 m are absent from the returned F2CSwaths"
    - "A swath that crosses an obstacle is split into two segments (one either side)"
    - "A swath entirely outside an obstacle is returned in full (length unchanged)"
    - "Calling avoid() with safety_margin < 0 throws std::invalid_argument"
    - "All pre-existing GoogleTest unit tests continue to pass after the module is added"
  artifacts:
    - path: "include/fields2cover/obstacle/obstacle_avoider.h"
      provides: "ObstacleAvoider class declaration in namespace f2c::obstacle"
      contains: "class ObstacleAvoider"
    - path: "src/fields2cover/obstacle/obstacle_avoider.cpp"
      provides: "ObstacleAvoider::avoid() implementation"
      contains: "ObstacleAvoider::avoid"
    - path: "tests/cpp/obstacle/obstacle_avoider_test.cpp"
      provides: "GoogleTest unit tests for ObstacleAvoider"
      contains: "fields2cover_obstacle_avoider"
  key_links:
    - from: "obstacle_avoider.cpp"
      to: "Cell::buffer()"
      via: "F2CCell::buffer(obstacle, safety_margin)"
      pattern: "Cell::buffer"
    - from: "obstacle_avoider.cpp"
      to: "F2CCells::difference()"
      via: "bbox_cells.difference(inflated)"
      pattern: "difference"
    - from: "obstacle_avoider.cpp"
      to: "F2CCells::getLinesInside()"
      via: "safe.getLinesInside(path)"
      pattern: "getLinesInside"
    - from: "obstacle_avoider_test.cpp"
      to: "obstacle_avoider.h"
      via: "#include"
      pattern: "obstacle/obstacle_avoider.h"
---

<objective>
Implement the ObstacleAvoider class — a new f2c algorithm module in `f2c::obstacle` — that
fragments a set of swaths around an inflated polygon obstacle. Swath segments that overlap the
inflated obstacle are removed; segments shorter than 0.1 m are discarded. The module follows
the identical three-file (header / source / test) pattern established by phases 24, 25, 25a,
and 25b.

Purpose: Satisfies F2C-05. Enables pipeline step that removes obstacle-intersecting swath
segments before route planning.

Output:
- `include/fields2cover/obstacle/obstacle_avoider.h`
- `src/fields2cover/obstacle/obstacle_avoider.cpp`
- `tests/cpp/obstacle/obstacle_avoider_test.cpp`
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/25c-c-obstacle-avoider/25c-RESEARCH.md
@.planning/phases/25c-c-obstacle-avoider/25c-PATTERNS.md

<interfaces>
<!-- Key types used in ObstacleAvoider. Extracted from codebase. Executor needs these — no codebase search needed. -->

From include/fields2cover/types/Cell.h:
```cpp
static Cell buffer(const Cell& geom, double width);  // line 54
// Wraps geom->OGRBuffer(width) — positive width inflates, negative shrinks
```

From include/fields2cover/types/Cells.h:
```cpp
explicit Cells(const Cell& c);                              // wrap one Cell
Cells difference(const Cell& c) const;                      // line 55
MultiLineString getLinesInside(const LineString& line) const; // line 71
```

From include/fields2cover/types/MultiLineString.h:
```cpp
size_t size() const;                          // line 28
LineString getGeometry(size_t i);             // line 38
```

From include/fields2cover/types/Swath.h:
```cpp
LineString getPath() const;    // line 46
double getWidth() const;       // line 49
SwathType getType() const;     // line 54
double length() const;         // line 58
```

From include/fields2cover/types/Swaths.h:
```cpp
void emplace_back(const LineString& l, double w, int id = 0,
    SwathType type = SwathType::MAINLAND);  // line 34
size_t size() const;
Swath& operator[](size_t i);  // via std::vector base
```

From include/fields2cover/types/LineString.h (Geometry base):
```cpp
double getDimMinX() const;
double getDimMaxX() const;
double getDimMinY() const;
double getDimMaxY() const;
double length() const;
```

Analog file: include/fields2cover/partition/multi_robot_partition.h
Analog file: src/fields2cover/partition/multi_robot_partition.cpp
Analog file: tests/cpp/partition/multi_robot_partition_test.cpp
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create ObstacleAvoider header and implementation</name>
  <files>
    include/fields2cover/obstacle/obstacle_avoider.h
    src/fields2cover/obstacle/obstacle_avoider.cpp
  </files>

  <read_first>
    - include/fields2cover/partition/multi_robot_partition.h  (include-guard + namespace pattern to replicate exactly)
    - src/fields2cover/partition/multi_robot_partition.cpp    (guard clause + bounding-box + algorithm shape to replicate)
    - include/fields2cover/types/Cell.h                       (Cell::buffer signature — line 54)
    - include/fields2cover/types/Cells.h                      (difference + getLinesInside signatures — lines 55, 71)
    - include/fields2cover/types/Swaths.h                     (emplace_back(LineString, double, int, SwathType) — line 34)
    - include/fields2cover/types/MultiLineString.h            (size() + getGeometry() — lines 28, 38)
    - include/fields2cover/types/Swath.h                      (getPath, getWidth, getType, length — lines 46, 49, 54, 58)
    - .planning/phases/25c-c-obstacle-avoider/25c-PATTERNS.md (exact algorithm pseudocode and code examples)
    - .planning/phases/25c-c-obstacle-avoider/25c-RESEARCH.md (pitfalls — bbox padding, degenerate intersections, buffer sign)
  </read_first>

  <behavior>
    - Test 1: avoid() with swath x∈[0,20] y=0 and square obstacle at x=10 half-side 2 (safety_margin=0) returns 2 segments each > 0.1 m
    - Test 2: avoid() with safety_margin=1.0 on same geometry returns segments whose combined length is less than with safety_margin=0
    - Test 3: avoid() with obstacle nearly covering entire swath leaving < 0.1 m residuals returns 0 segments
    - Test 4: avoid() with safety_margin=-1.0 throws std::invalid_argument
    - Test 5: avoid() with obstacle placed far from swath returns 1 segment whose length equals the original swath length
  </behavior>

  <action>
Create directory `include/fields2cover/obstacle/` and write `obstacle_avoider.h`:

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
#define FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_

#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::obstacle {

/// @brief Fragment swaths around an inflated polygon obstacle.
///
/// For each input swath, the obstacle is inflated by `safety_margin` metres,
/// the swath centre-line is clipped against the complement of the inflated
/// obstacle, and any residual segment shorter than kMinSegmentLength is dropped.
///
/// Geometry operations are delegated entirely to the f2c/OGR type layer —
/// no raw GEOS or OGRGeometry pointers are used.
class ObstacleAvoider {
 public:
  /// Minimum returned segment length (metres). Segments below this are dropped.
  static constexpr double kMinSegmentLength = 0.1;

  /// @brief Fragment swaths around an inflated polygon obstacle.
  ///
  /// @param swaths        Input swaths to fragment.
  /// @param obstacle      Obstacle polygon (F2CCell) in the same CRS as swaths.
  /// @param safety_margin Buffer distance (metres >= 0) to inflate the obstacle
  ///                      before clipping.
  /// @return F2CSwaths of residual segments; segments < kMinSegmentLength dropped.
  /// @throws std::invalid_argument if safety_margin < 0.
  F2CSwaths avoid(const F2CSwaths& swaths,
                  const F2CCell& obstacle,
                  double safety_margin) const;
};

}  // namespace f2c::obstacle

#endif  // FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
```

Create directory `src/fields2cover/obstacle/` and write `obstacle_avoider.cpp`:

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/obstacle/obstacle_avoider.h"

#include <stdexcept>

namespace f2c::obstacle {

F2CSwaths ObstacleAvoider::avoid(
    const F2CSwaths& swaths,
    const F2CCell& obstacle,
    double safety_margin) const {
  if (safety_margin < 0.0) {
    throw std::invalid_argument(
        "ObstacleAvoider::avoid: safety_margin must be >= 0");
  }

  // 1. Inflate the obstacle by the safety margin.
  F2CCell inflated = F2CCell::buffer(obstacle, safety_margin);

  F2CSwaths result;
  int out_id = 0;

  for (size_t i = 0; i < swaths.size(); ++i) {
    const F2CSwath& sw = swaths[i];
    F2CLineString path = sw.getPath();

    // 2. Build a bounding-box Cells that covers the swath + padding,
    //    then subtract the inflated obstacle to obtain the safe region.
    //    Padding = safety_margin + 1.0 ensures the bbox is never swallowed
    //    by a large obstacle (minimum 1 m floor when safety_margin == 0).
    const double padding = safety_margin + 1.0;
    const double x0 = path.getDimMinX() - padding;
    const double x1 = path.getDimMaxX() + padding;
    const double y0 = path.getDimMinY() - padding;
    const double y1 = path.getDimMaxY() + padding;
    F2CLinearRing ring{
        F2CPoint(x0, y0), F2CPoint(x1, y0),
        F2CPoint(x1, y1), F2CPoint(x0, y1),
        F2CPoint(x0, y0)};
    F2CCells bbox_cells(F2CCell{ring});
    F2CCells safe = bbox_cells.difference(inflated);

    // 3. Clip the swath path against the safe region.
    //    getLinesInside returns only the portions of `path` inside `safe`.
    F2CMultiLineString residuals = safe.getLinesInside(path);

    // 4. Filter by minimum length and emit new swaths.
    //    IDs are reassigned sequentially so every output swath has a unique id.
    //    Original width and type are preserved.
    for (size_t j = 0; j < residuals.size(); ++j) {
      F2CLineString seg = residuals.getGeometry(j);
      if (seg.length() >= kMinSegmentLength) {
        result.emplace_back(seg, sw.getWidth(), out_id++, sw.getType());
      }
    }
  }

  return result;
}

}  // namespace f2c::obstacle
```

IMPORTANT after writing both files: re-run cmake to pick up the new `obstacle/` directories
(GLOB_RECURSE caches at configure time):
```bash
cd /home/tom/devenv/fields2cover && cmake -S . -B build 2>&1 | tail -5
```
Then build:
```bash
cd /home/tom/devenv/fields2cover/build && make -j$(nproc) 2>&1 | tail -20
```
The build must succeed (no linker errors about ObstacleAvoider) before Task 2 proceeds.
  </action>

  <verify>
    <automated>
      cd /home/tom/devenv/fields2cover/build && make -j$(nproc) 2>&amp;&amp; echo "BUILD_OK"
    </automated>
  </verify>

  <acceptance_criteria>
    - `include/fields2cover/obstacle/obstacle_avoider.h` exists
    - `src/fields2cover/obstacle/obstacle_avoider.cpp` exists
    - `grep -r "class ObstacleAvoider" include/fields2cover/obstacle/obstacle_avoider.h` produces a match
    - `grep "FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_" include/fields2cover/obstacle/obstacle_avoider.h` produces a match (include guard present)
    - `grep "namespace f2c::obstacle" src/fields2cover/obstacle/obstacle_avoider.cpp` produces a match
    - `grep "Cell::buffer" src/fields2cover/obstacle/obstacle_avoider.cpp` produces a match
    - `grep "getLinesInside" src/fields2cover/obstacle/obstacle_avoider.cpp` produces a match
    - `grep "kMinSegmentLength" src/fields2cover/obstacle/obstacle_avoider.cpp` produces a match
    - `grep "invalid_argument" src/fields2cover/obstacle/obstacle_avoider.cpp` produces a match
    - `cd /home/tom/devenv/fields2cover/build && make -j$(nproc) 2>&1 | grep -c "error:"` outputs `0`
  </acceptance_criteria>

  <done>
    Both files exist, the build compiles without errors, and grep confirms all key API calls
    (Cell::buffer, getLinesInside, kMinSegmentLength, invalid_argument guard) are present.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Write GoogleTest unit tests for ObstacleAvoider</name>
  <files>
    tests/cpp/obstacle/obstacle_avoider_test.cpp
  </files>

  <read_first>
    - tests/cpp/partition/multi_robot_partition_test.cpp   (test file structure, anonymous namespace, helper + TEST macro patterns)
    - include/fields2cover/obstacle/obstacle_avoider.h     (the class being tested — verify method signature before writing calls)
    - .planning/phases/25c-c-obstacle-avoider/25c-PATTERNS.md (exact test cases with coordinates, section "Tests to implement")
    - .planning/phases/25c-c-obstacle-avoider/25c-RESEARCH.md (Pattern 5: GoogleTest fixture — full test case code)
  </read_first>

  <behavior>
    - Test `obstacle_splits_swath_into_two`: swath x∈[0,20] y=0 + square obstacle centred (10,0) half-side 2 safety_margin=0 → ASSERT_EQ(result.size(), 2u), both segments > 0.1 m
    - Test `safety_margin_enlarges_exclusion`: same geometry safety_margin=1.0 → sum of segment lengths less than safety_margin=0 case
    - Test `short_segments_dropped`: obstacle nearly covers swath leaving tiny residuals → EXPECT_EQ(result.size(), 0u)
    - Test `negative_margin_throws`: safety_margin=-1.0 → EXPECT_THROW std::invalid_argument
    - Test `no_obstacle_overlap_returns_full_swath`: obstacle at x=100 (far from swath x∈[0,20]) → ASSERT_EQ(result.size(), 1u), result[0].length() near 20.0 within 0.01
  </behavior>

  <action>
Create directory `tests/cpp/obstacle/` and write `obstacle_avoider_test.cpp`:

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/obstacle/obstacle_avoider.h"

namespace {

/// Helper: build a single straight horizontal swath along y=0, x in [x0, x1].
F2CSwaths makeStraightSwaths(double x0, double x1, double width = 3.0) {
  F2CLineString path{F2CPoint(x0, 0.0), F2CPoint(x1, 0.0)};
  F2CSwaths s;
  s.emplace_back(path, width);
  return s;
}

/// Helper: build a square F2CCell obstacle centred at (cx, cy) with half-side r.
F2CCell makeSquareObstacle(double cx, double cy, double r) {
  F2CLinearRing ring{
      F2CPoint(cx - r, cy - r), F2CPoint(cx + r, cy - r),
      F2CPoint(cx + r, cy + r), F2CPoint(cx - r, cy + r),
      F2CPoint(cx - r, cy - r)};
  return F2CCell{ring};
}

}  // namespace

// -----------------------------------------------------------------------------
// obstacle_splits_swath_into_two
// Swath: x in [0, 20], y=0.  Square obstacle centred at (10,0), half-side 2 m.
// The obstacle straddles the swath midpoint → expect exactly 2 residual segments.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, obstacle_splits_swath_into_two) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  auto result = avoider.avoid(swaths, obstacle, 0.0);

  ASSERT_EQ(result.size(), 2u);
  EXPECT_GT(result[0].length(), 0.1);
  EXPECT_GT(result[1].length(), 0.1);
}

// -----------------------------------------------------------------------------
// safety_margin_enlarges_exclusion
// Adding a 1 m safety margin inflates the obstacle → shorter residual segments.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, safety_margin_enlarges_exclusion) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  auto result_no_margin = avoider.avoid(swaths, obstacle, 0.0);
  auto result_with_margin = avoider.avoid(swaths, obstacle, 1.0);

  // Total covered length must decrease when margin is added.
  double len_no_margin = 0.0;
  for (size_t i = 0; i < result_no_margin.size(); ++i) {
    len_no_margin += result_no_margin[i].length();
  }
  double len_with_margin = 0.0;
  for (size_t i = 0; i < result_with_margin.size(); ++i) {
    len_with_margin += result_with_margin[i].length();
  }

  EXPECT_LT(len_with_margin, len_no_margin);
}

// -----------------------------------------------------------------------------
// short_segments_dropped
// Obstacle nearly covers the entire swath (x in [0.5, 19.5]) → residuals < 0.1 m
// must be dropped, so result is empty.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, short_segments_dropped) {
  // Swath: x in [0, 20].  Obstacle covers [0.5, 19.5] → residuals ~0.5 m each.
  // Add safety_margin=0.45 to shrink residuals to ~0.05 m (< 0.1 threshold).
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 9.5);  // covers x in [0.5, 19.5]

  f2c::obstacle::ObstacleAvoider avoider;
  // With safety_margin=0.45, residuals become ~0.05 m → all dropped.
  auto result = avoider.avoid(swaths, obstacle, 0.45);

  EXPECT_EQ(result.size(), 0u);
}

// -----------------------------------------------------------------------------
// negative_margin_throws
// A negative safety_margin is invalid — expect std::invalid_argument.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, negative_margin_throws) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  EXPECT_THROW(avoider.avoid(swaths, obstacle, -1.0), std::invalid_argument);
}

// -----------------------------------------------------------------------------
// no_obstacle_overlap_returns_full_swath
// Obstacle placed far from swath → full swath is returned as one segment.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, no_obstacle_overlap_returns_full_swath) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(100.0, 0.0, 2.0);  // far from swath

  f2c::obstacle::ObstacleAvoider avoider;
  auto result = avoider.avoid(swaths, obstacle, 0.0);

  ASSERT_EQ(result.size(), 1u);
  EXPECT_NEAR(result[0].length(), 20.0, 0.01);
}
```

After writing the test file, re-run cmake and make to include the new test:
```bash
cd /home/tom/devenv/fields2cover && cmake -S . -B build 2>&1 | tail -5 && cd build && make -j$(nproc) 2>&1 | tail -10
```
Then run the obstacle tests:
```bash
cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure
```
All 5 tests must pass.

Then run the full suite to confirm no regressions:
```bash
cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure 2>&1 | tail -20
```
  </action>

  <verify>
    <automated>
      cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure
    </automated>
  </verify>

  <acceptance_criteria>
    - `tests/cpp/obstacle/obstacle_avoider_test.cpp` exists
    - `grep "fields2cover_obstacle_avoider" tests/cpp/obstacle/obstacle_avoider_test.cpp` produces 5 matches (one per TEST macro)
    - `grep "obstacle_splits_swath_into_two" tests/cpp/obstacle/obstacle_avoider_test.cpp` produces a match
    - `grep "short_segments_dropped" tests/cpp/obstacle/obstacle_avoider_test.cpp` produces a match
    - `grep "negative_margin_throws" tests/cpp/obstacle/obstacle_avoider_test.cpp` produces a match
    - `grep "no_obstacle_overlap_returns_full_swath" tests/cpp/obstacle/obstacle_avoider_test.cpp` produces a match
    - `cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure` exits 0 with "5 tests passed"
    - `cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure 2>&1 | grep -E "^[0-9]+ tests passed"` shows count >= 318 (313 pre-existing + 5 new)
  </acceptance_criteria>

  <done>
    All 5 obstacle tests pass. Full test suite passes with count >= 318.
    No pre-existing tests regressed.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| C++ library internal | All inputs come from other f2c C++ code; no external network, file I/O, or user-supplied raw bytes |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-25c-01 | Tampering | ObstacleAvoider::avoid — safety_margin parameter | mitigate | Validate at entry: throw std::invalid_argument if safety_margin < 0; prevents OGR erosion misuse |
| T-25c-02 | Denial of Service | Cell::buffer with very large safety_margin | accept | Pure geometry library; no resource limits enforced at this layer; callers (Phase 26 gRPC) impose input bounds |
| T-25c-03 | Information Disclosure | OGR memory management | accept | All geometry operations go through f2c wrapper types (destroyResGeom<>); no raw OGRGeometry* escapes |
</threat_model>

<verification>
After both tasks are complete:

```bash
# 1. Obstacle-specific tests
cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure

# 2. Full regression suite
cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure 2>&1 | tail -5
```

Expected: 5 obstacle tests pass; full suite >= 318 tests pass.
</verification>

<success_criteria>
1. `ObstacleAvoider::avoid()` with a swath crossing an obstacle returns exactly 2 segments
   (split confirmation — F2C-05 success criterion 3)
2. No returned segment is shorter than 0.1 m (F2C-05 success criterion 2)
3. Returned segments have zero overlap with the inflated obstacle polygon
   (verified implicitly by the split test: segments are outside the obstacle boundary)
4. Full GoogleTest suite passes — all pre-existing 313+ tests remain green (F2C-05 success criterion 3)
5. `ctest -R obstacle --output-on-failure` exits 0 with 5/5 tests passed
</success_criteria>

<output>
After completion, create `.planning/phases/25c-c-obstacle-avoider/25c-01-SUMMARY.md` using the
summary template at `$HOME/.claude/get-shit-done/templates/summary.md`.

Include:
- Files created: header, source, test
- Key decisions: bbox-padding formula (safety_margin + 1.0), sequential ID reassignment,
  Cell::buffer + difference + getLinesInside idiom
- Test count before/after
- Any deviations from the plan and why
</output>
