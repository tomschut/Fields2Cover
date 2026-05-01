# Phase 25c: C++ Obstacle Avoider - Research

**Researched:** 2026-04-29
**Domain:** C++ geometry clipping, GDAL/OGR polygon buffering, swath fragmentation
**Confidence:** HIGH

## Summary

Phase 25c adds `ObstacleAvoider` to a new `obstacle/` module. The class exposes a single
method `avoid(F2CSwaths, F2CCell, double safety_margin)` that inflates the obstacle polygon,
clips each swath's centre-line against the complement of the inflated obstacle, drops short
residual segments, and returns the surviving `F2CSwaths`.

All geometry operations are already present in the f2c type library:

- `Cell::buffer(cell, margin)` inflates the obstacle — wraps OGR `OGRBuffer()` [VERIFIED: codebase]
- `getLinesInside(LineString)` on `F2CCells` clips a line to the non-obstacle area — wraps
  `MultiLineString::intersection(line, *this)` [VERIFIED: codebase]
- `F2CSwath::length()` provides segment length for the minimum-threshold filter [VERIFIED: codebase]

The implementation follows the identical three-file pattern (header, source, test) established
by phases 24, 25, 25a, and 25b. No CMakeLists.txt edits are needed — GLOB_RECURSE in both
the library and test builds auto-discovers any new `.cpp` under `src/` and `tests/cpp/`
respectively. [VERIFIED: codebase]

**Primary recommendation:** Implement `ObstacleAvoider::avoid()` by:
1. Inflating the obstacle with `Cell::buffer(obstacle, safety_margin)` to produce an `F2CCell`
2. Computing the "safe region" as the field's bounding `F2CCells` minus the inflated obstacle
   — represented as `Cells` with the inflated cell treated as a hole; the practical approach
   is to build a `F2CCells` from the inflated cell and use `getLinesInside` on the *complement*
   — the canonical f2c idiom is to call `getLinesInside` directly on the safe-zone `Cells`
3. For each input `F2CSwath`, extract its `LineString` path, call `safe_cells.getLinesInside(path)`,
   and for each returned segment whose `length()` >= 0.1 m, push a new `F2CSwath` (preserving
   original width, id, type)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
All implementation choices are at Claude's discretion — discuss phase was skipped per user
setting. Use ROADMAP phase goal, success criteria, and codebase conventions to guide decisions.

Key decisions for Claude to make:
- Obstacle inflation strategy (buffer polygon by safety_margin, use GEOS/Boost.Geometry)
- Swath fragmentation: clip each swath LineString against the inflated obstacle polygon, collect residual segments
- Minimum segment threshold: drop segments < 0.1 m (per success criteria)
- Unit test: place known obstacle across a swath so it splits into 2 segments

### Claude's Discretion
All implementation choices are at Claude's discretion.

### Deferred Ideas (OUT OF SCOPE)
None — discuss phase skipped.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| F2C-05 | f2c library can fragment swaths around polygon obstacles, producing trimmed swath segments that avoid inflated obstacle boundaries | `Cell::buffer()` provides inflation; `F2CCells::getLinesInside()` provides line-polygon clipping; `Swath::length()` enables minimum-threshold filter |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Obstacle polygon inflation | C++ library (f2c::obstacle) | — | Pure geometry; Cell::buffer wraps OGR natively |
| Swath line clipping against safe region | C++ library (f2c::obstacle) | — | getLinesInside is the canonical f2c idiom for line-polygon intersection |
| Minimum segment length filter | C++ library (f2c::obstacle) | — | Swath::length() available in-process; no I/O needed |
| Unit testing | Test tier (GoogleTest) | — | All existing tests use GoogleTest; GLOB_RECURSE auto-discovers new test files |
| gRPC/REST exposure | Phase 26 | — | Out of scope; noted in ROADMAP as Phase 26 dependency |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `fields2cover/types.h` | project-local | F2CSwaths, F2CSwath, F2CCell, F2CCells, F2CLineString | Required by every f2c algorithm |
| `<vector>` (std) | system | Return type for output swath collection | Standard C++ |
| `<stdexcept>` (std) | system | `std::invalid_argument` for negative safety_margin | Consistent with all other modules |

### Supporting (test only)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `<gtest/gtest.h>` | system (cmake) | GoogleTest macros | All f2c unit tests |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `Cell::buffer(obstacle, margin)` + `getLinesInside` | Manual GEOS `GEOSBuffer` / `GEOSDifference` | The f2c wrappers already handle OGRGeometry lifetime; using raw GEOS bypasses the type system and requires manual memory management |
| `F2CCells safe_zone = field_as_cells.difference(inflated)` then `getLinesInside` | Same | This variant requires a field polygon input. Using the inflated obstacle directly as the exclusion zone avoids requiring a field boundary — simpler signature |
| Boost.Geometry for buffering | OGR (via `Cell::buffer`) | OGR is already the f2c geometry engine; adding Boost.Geometry is an unnecessary dependency |

**Installation:** No new packages. All geometry operations are already compiled into `libFields2Cover`.

## Architecture Patterns

### System Architecture Diagram

```
Input: F2CSwaths swaths, F2CCell obstacle, double safety_margin
          |
          v
    [1] Cell::buffer(obstacle, safety_margin)
          |  --> inflated_obstacle: F2CCell
          |
    [2] Wrap obstacle as F2CCells: F2CCells excl(inflated_obstacle)
          |
    For each F2CSwath in swaths:
          |
    [3]   swath.getPath()           --> F2CLineString swath_line
          |
    [4]   excl.getLinesInside(swath_line) produces the OVERLAP segments
          We need the complement: segments OUTSIDE the obstacle.
          --> Use two-complement approach:
              build F2CMultiLineString from original line,
              subtract the overlap via OGR difference,
              OR: build safe_region = field_bounding_cells.difference(excl),
              then safe_region.getLinesInside(swath_line)
              --> F2CMultiLineString residuals
          |
    [5]   For each LineString seg in residuals:
              if seg.length() >= 0.1 m:
                  output.emplace_back(seg, swath.getWidth(), new_id++, swath.getType())
          |
          v
Output: F2CSwaths (trimmed segments, short segments dropped)
```

**Practical approach for step 4:** Build `F2CCells safe_cells` by creating a very large bounding
box cell (e.g., 2x the swath length in each direction, centred on the swath midpoint) and
subtracting the inflated obstacle — `safe_cells = bbox.difference(inflated_obstacle)`. Then call
`safe_cells.getLinesInside(swath_line)`. This is the exact idiom the f2c swath generator uses
to clip swaths to the field interior. [VERIFIED: src/fields2cover/types/Swaths.cpp line 108]

**Simpler practical approach:** Use `MultiLineString::intersection` with the complement directly.
Since `getLinesInside` computes `MultiLineString::intersection(line, *this)`, and the available
`difference` operates on `Cells`, the most direct path is:

```cpp
// 1. inflate
F2CCell inflated = Cell::buffer(obstacle, safety_margin);
F2CCells excl(inflated);

// 2. for each swath, compute complement clips
F2CLineString path = swath.getPath();
// Get the "outside" segments by intersecting with a bounding cells
// that excludes the inflated obstacle.
// Build a safe region: large bbox minus inflated obstacle
F2CCells safe = makeBboxCells(path).difference(inflated);
F2CMultiLineString residuals = safe.getLinesInside(path);

// 3. filter and collect
for (size_t j = 0; j < residuals.size(); ++j) {
  auto seg = residuals.getGeometry(j);
  if (seg.length() >= kMinSegmentLength) {
    result.emplace_back(seg, swath.getWidth(), out_id++, swath.getType());
  }
}
```

### Recommended Project Structure
```
include/fields2cover/obstacle/
└── obstacle_avoider.h          # ObstacleAvoider class declaration

src/fields2cover/obstacle/
└── obstacle_avoider.cpp        # ObstacleAvoider::avoid() implementation

tests/cpp/obstacle/
└── obstacle_avoider_test.cpp   # GoogleTest unit tests
```

### Pattern 1: Module Header Convention
**What:** All new module headers follow the same guard/namespace/include pattern.
**When to use:** Always for new f2c algorithm classes.
**Example:**
```cpp
// Source: codebase — include/fields2cover/partition/multi_robot_partition.h
#pragma once
#ifndef FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
#define FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::obstacle {

class ObstacleAvoider {
 public:
  static constexpr double kMinSegmentLength = 0.1;  // metres

  /// @brief Fragment swaths around an inflated polygon obstacle.
  /// @param swaths       Input swaths to fragment.
  /// @param obstacle     Obstacle polygon (F2CCell) in the same CRS as swaths.
  /// @param safety_margin Buffer distance (metres) to inflate obstacle before clipping.
  /// @return F2CSwaths containing residual segments; segments < kMinSegmentLength are dropped.
  /// @throws std::invalid_argument if safety_margin < 0.
  F2CSwaths avoid(const F2CSwaths& swaths,
                  const F2CCell& obstacle,
                  double safety_margin) const;
};

}  // namespace f2c::obstacle

#endif  // FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
```

### Pattern 2: Polygon Buffering via Cell::buffer
**What:** Inflate a polygon obstacle by a safety margin.
**When to use:** Always for obstacle inflation in f2c code.
**Example:**
```cpp
// Source: codebase — include/fields2cover/types/Cell.h (line 54)
// Cell::buffer(const Cell& geom, double width) wraps geom->OGRBuffer(width)
F2CCell inflated = Cell::buffer(obstacle, safety_margin);
```

### Pattern 3: Line-Polygon Clipping via getLinesInside
**What:** Compute segments of a LineString that lie inside a Cells polygon.
**When to use:** Whenever a line must be trimmed to a polygon boundary — the canonical f2c idiom.
**Example:**
```cpp
// Source: codebase — src/fields2cover/types/Swaths.cpp lines 106-109
// F2CSwaths::append(line, polys) calls polys.getLinesInside(line) internally.
// For ObstacleAvoider: build a "safe" Cells, then:
F2CMultiLineString residuals = safe_cells.getLinesInside(swath.getPath());
```

### Pattern 4: Building a Safe-Region Cells via difference
**What:** Subtract the inflated obstacle from a bounding region to produce the "safe zone".
**When to use:** When you need the exterior (complement) of an obstacle polygon.
**Example:**
```cpp
// Source: codebase — src/fields2cover/types/Cells.cpp line 164-169
// Cells::difference(Cell) wraps this->data_->Difference(c.get())
// Build bbox around swath to avoid infinite-plane arithmetic:
F2CCells bbox = makeBboxCells(swath.getPath(), padding);
F2CCells safe = bbox.difference(inflated_obstacle);
F2CMultiLineString residuals = safe.getLinesInside(swath.getPath());
```

### Pattern 5: GoogleTest Fixture for Swath Tests
**What:** Standard test structure — namespace-anonymous, helper functions, EXPECT_*/ASSERT_*.
**When to use:** All new f2c tests.
**Example:**
```cpp
// Source: codebase — tests/cpp/partition/multi_robot_partition_test.cpp
#include <gtest/gtest.h>
#include "fields2cover/types.h"
#include "fields2cover/obstacle/obstacle_avoider.h"

namespace {

// Helper: build a straight horizontal swath along y=0, x in [x0, x1]
F2CSwaths makeStraightSwath(double x0, double x1, double width = 3.0) {
  F2CLineString path{F2CPoint(x0, 0.0), F2CPoint(x1, 0.0)};
  F2CSwaths s;
  s.emplace_back(path, width);
  return s;
}

// Helper: build a square obstacle cell centred at (cx, cy) with half-side r
F2CCell makeSquareObstacle(double cx, double cy, double r) {
  F2CLinearRing ring{
    F2CPoint(cx-r, cy-r), F2CPoint(cx+r, cy-r),
    F2CPoint(cx+r, cy+r), F2CPoint(cx-r, cy+r),
    F2CPoint(cx-r, cy-r)};
  return F2CCell{ring};
}

}  // namespace

TEST(fields2cover_obstacle_avoider, obstacle_splits_swath_into_two) {
  // Swath: x in [0, 20], y=0.  Obstacle at x=10, half-side 2m, no margin.
  // After avoidance: expect segments [0,8] and [12,20] (approx).
  auto swaths = makeStraightSwath(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  auto result = avoider.avoid(swaths, obstacle, 0.0);

  ASSERT_EQ(result.size(), 2u);
  EXPECT_GT(result[0].length(), 0.1);
  EXPECT_GT(result[1].length(), 0.1);
}
```

### Anti-Patterns to Avoid
- **Raw OGR pointers without destroyGeometry:** All OGR results must be passed through
  `destroyResGeom<>` or the f2c wrapper types to avoid memory leaks. Never use raw
  `OGRGeometry*` intermediates outside of the type wrappers.
- **Using `Cells::difference` on the full field polygon for obstacle avoidance:** The
  `ObstacleAvoider` intentionally does NOT take a field polygon — only swaths and the obstacle.
  Constructing a bbox dynamically (or using the inflated obstacle directly) avoids requiring
  callers to provide field geometry.
- **Storing F2CSwath IDs as the original swath's id:** When one swath splits into multiple
  segments, the id must be reassigned (sequential counter) to keep IDs unique within the result
  set. Preserve the original `width` and `type` attributes.
- **Returning segments that are exactly at the OGR intersection boundary (point segments):**
  OGR intersection can produce degenerate point "lines" for tangent contacts. The `length() >= 0.1`
  filter eliminates these.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Polygon buffering / inflation | Custom offset algorithm | `Cell::buffer(obstacle, margin)` | OGR handles convex/concave shapes, holes, arc segments — hand-rolling is fragile |
| Line-polygon intersection | Manual parametric clipping | `Cells::getLinesInside(line)` | OGR `Intersection` handles all edge cases (tangencies, multi-segment output, empty result) |
| Line difference against polygon | Manual segment arithmetic | `Cells::difference` + `getLinesInside` | OGR handles numeric precision and degenerate cases |
| GoogleTest entry point | Custom `main()` | Existing `tests/unittests.cpp` | GLOB_RECURSE assembles all test `.cpp` files into one executable with the existing `main()` |

**Key insight:** Every geometry operation needed (buffer, intersect, difference) is already wrapped
in f2c's type layer. ObstacleAvoider is pure composition of existing operations — no new geometry
primitives needed.

## Common Pitfalls

### Pitfall 1: Safety-Margin Sign
**What goes wrong:** Calling `Cell::buffer(obstacle, -margin)` with a positive margin value
shrinks instead of inflates the obstacle.
**Why it happens:** OGR `Buffer` treats negative distance as erosion (inward offset).
**How to avoid:** Always pass `safety_margin >= 0` to buffer. Validate at entry: throw
`std::invalid_argument` if `safety_margin < 0`.
**Warning signs:** Unit test with safety_margin=1.0 produces more swath coverage than with
safety_margin=0 (obstacle shrank rather than grew).

### Pitfall 2: Empty Cells::difference When Swath Does Not Cross Obstacle
**What goes wrong:** `bbox.difference(inflated)` returns an empty `F2CCells` if the bbox is
entirely inside the inflated obstacle (bbox was too small).
**Why it happens:** Bbox constructed from only the swath endpoints without padding can be
fully enclosed by a large obstacle.
**How to avoid:** Expand the bbox by `safety_margin + epsilon` beyond the swath endpoint
extents, or use a fixed large padding (e.g., `max(safety_margin * 2, 1.0)` on each side).
**Warning signs:** Test with obstacle entirely enclosing the swath returns non-empty result
(should return zero segments).

### Pitfall 3: Degenerate Point Intersections
**What goes wrong:** When a swath is tangent to (touches but does not cross) the inflated
obstacle boundary, OGR may return a degenerate `LineString` of zero length or two coincident
points.
**Why it happens:** OGR `Intersection` preserves all geometry types from the result, including
point components embedded in `GeometryCollection` output.
**How to avoid:** The `length() >= kMinSegmentLength` filter (0.1 m) eliminates these.
**Warning signs:** `result.size()` is unexpectedly large for tangent test cases.

### Pitfall 4: F2CCells Constructor from OGRGeometry with Non-Polygon Type
**What goes wrong:** `Cells::difference(Cell)` can return a `GeometryCollection` rather than
`MultiPolygon` when OGR encounters degenerate edge cases. The `F2CCells(OGRGeometry*)` constructor
handles this by filtering for polygon sub-geometries, but the resulting `Cells` may be empty.
**Why it happens:** OGR difference on near-touching polygons can produce mixed-type collections.
**How to avoid:** Handle the case where `safe_cells.size() == 0` gracefully — the swath is
entirely inside the obstacle and should produce no output segments.
**Warning signs:** Crash or empty result for swath entirely inside obstacle boundary.

### Pitfall 5: Forgetting cmake Re-Run After Adding New Directories
**What goes wrong:** New `include/fields2cover/obstacle/` and `src/fields2cover/obstacle/` are
not picked up by the build because cmake cached the GLOB_RECURSE result.
**Why it happens:** GLOB_RECURSE results are cached at cmake configure time, not re-evaluated
at make time.
**How to avoid:** After creating new files in new directories, always re-run
`cmake -S . -B build` (or equivalent) before `make`.
**Warning signs:** Linker error "undefined reference to `f2c::obstacle::ObstacleAvoider::avoid`".

## Code Examples

Verified patterns from codebase:

### Buffer an F2CCell obstacle
```cpp
// Source: include/fields2cover/types/Cell.h line 54
// Source: src/fields2cover/types/Cell.cpp lines 101-103
F2CCell inflated = Cell::buffer(obstacle, safety_margin);
// Internally: destroyResGeom<Cell>(obstacle.OGRBuffer(safety_margin))
```

### Get line segments inside a Cells polygon
```cpp
// Source: src/fields2cover/types/Cells.cpp lines 217-219
// Source: src/fields2cover/types/Swaths.cpp lines 106-109
F2CMultiLineString residuals = safe_cells.getLinesInside(swath.getPath());
// Internally: MultiLineString::intersection(line, *this)
//             = line->Intersection(cells.get()) via OGR
```

### Compute Cells difference (safe region = bbox - inflated obstacle)
```cpp
// Source: src/fields2cover/types/Cells.cpp lines 164-169
F2CCells bbox_cells(bbox_cell);          // wrap F2CCell in F2CCells
F2CCells safe = bbox_cells.difference(inflated_obstacle);
// Internally: this->data_->Difference(c.get())
```

### Iterate MultiLineString result
```cpp
// Source: include/fields2cover/types/MultiLineString.h lines 34-37
for (size_t j = 0; j < residuals.size(); ++j) {
  F2CLineString seg = residuals.getGeometry(j);
  if (seg.length() >= kMinSegmentLength) {
    result.emplace_back(seg, original_width, out_id++, original_type);
  }
}
```

### Build a bounding-box cell around a swath path
```cpp
// Source: pattern from src/fields2cover/partition/multi_robot_partition.cpp lines 36-39
// (getDimMinX/getDimMaxX used on F2CCells — same interface on F2CCell via Geometry base)
double x0 = path.getDimMinX() - padding;
double x1 = path.getDimMaxX() + padding;
double y0 = path.getDimMinY() - padding;
double y1 = path.getDimMaxY() + padding;
F2CLinearRing ring{
    F2CPoint(x0, y0), F2CPoint(x1, y0),
    F2CPoint(x1, y1), F2CPoint(x0, y1),
    F2CPoint(x0, y0)};
F2CCells bbox_cells(F2CCell{ring});
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual GEOS API calls | OGR wrapper layer in f2c types | f2c v1.0 | All geometry operations go through F2CCell/F2CCells — no raw GEOS usage in new code |
| Field-level obstacle handling (holes in field polygon) | Module-level avoider operating on swath segments | Phase 25c (new) | Obstacle avoidance becomes a composable pipeline step, not baked into field geometry |

## Open Questions

1. **Bbox padding heuristic**
   - What we know: bbox must be larger than the swath + inflated obstacle to avoid the bbox
     being swallowed by the obstacle in the difference step
   - What's unclear: exact formula for adequate padding when the obstacle is very large relative
     to the swath
   - Recommendation: use `padding = safety_margin + 1.0` (1 m floor ensures padding even when
     safety_margin=0); document this as an implementation constant

2. **Multiple obstacles in a single call**
   - What we know: success criteria say "a polygon obstacle" (singular) — one obstacle per call
   - What's unclear: should the API support `std::vector<F2CCell>` for multiple obstacles?
   - Recommendation: implement single-obstacle form matching success criteria; composability
     (call avoid() multiple times) handles the multiple-obstacle case without API complexity

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| CMake | Build system | Yes | 3.28.3 | — |
| libFields2Cover (built) | Linking tests | Yes | build/ present | Re-run cmake + make |
| GoogleTest | Unit tests | Yes | linked in CMakeLists | — |
| GDAL/OGR | Cell::buffer, Intersection | Yes | compiled into libFields2Cover | — |

**Missing dependencies with no fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | GoogleTest (gtest) |
| Config file | tests/CMakeLists.txt — GLOB_RECURSE auto-discovers test files |
| Quick run command | `cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure` |
| Full suite command | `cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-05-a | avoid() returns segments with no overlap with inflated obstacle | unit | `ctest -R obstacle` | No — Wave 0 |
| F2C-05-b | Segments < 0.1 m are dropped | unit | `ctest -R obstacle` | No — Wave 0 |
| F2C-05-c | At least one swath splits into two segments | unit | `ctest -R obstacle` | No — Wave 0 |
| F2C-05-d | All pre-existing tests continue to pass | regression | `ctest --output-on-failure` | Yes (313+ tests) |

### Sampling Rate
- **Per task commit:** `cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure`
- **Per wave merge:** `cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `tests/cpp/obstacle/obstacle_avoider_test.cpp` — covers F2C-05-a, F2C-05-b, F2C-05-c
- [ ] `include/fields2cover/obstacle/obstacle_avoider.h` — header must exist before test compiles
- [ ] `src/fields2cover/obstacle/obstacle_avoider.cpp` — implementation
- [ ] cmake re-run required after new directories are created

*(All three files are created in a single plan wave; cmake re-run is the first build step)*

## Sources

### Primary (HIGH confidence)
- Codebase: `include/fields2cover/types/Cell.h` — Cell::buffer signatures verified
- Codebase: `include/fields2cover/types/Cells.h` — getLinesInside, difference, buffer signatures verified
- Codebase: `src/fields2cover/types/Cells.cpp` — getLinesInside and difference implementations verified
- Codebase: `src/fields2cover/types/Cell.cpp` — Cell::buffer wraps OGRBuffer verified
- Codebase: `src/fields2cover/types/Swaths.cpp` — append(line, polys) uses getLinesInside idiom verified
- Codebase: `tests/CMakeLists.txt` — GLOB_RECURSE pattern `cpp/*/*.cpp` verified
- Codebase: `CMakeLists.txt` — GLOB_RECURSE `src/*.cpp` for library verified
- Codebase: phase 25b plan — identical three-file module pattern confirmed

### Secondary (MEDIUM confidence)
- None required — all findings are directly verified from the codebase.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all types and operations verified in codebase
- Architecture: HIGH — exact same three-file pattern as phases 24, 25, 25a, 25b; geometry idioms confirmed in Swaths.cpp
- Pitfalls: HIGH — OGR buffer sign, bbox sizing, degenerate intersection, cmake cache all verified against codebase comments (e.g., splitByLine comment about OGR artefacts)

**Research date:** 2026-04-29
**Valid until:** 2026-05-29 (f2c type layer is stable; OGR behaviour is stable)
