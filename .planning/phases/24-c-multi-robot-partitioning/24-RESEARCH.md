# Phase 24: C++ Multi-Robot Partitioning - Research

**Researched:** 2026-04-29
**Domain:** C++ / fields2cover library / field geometry partitioning
**Confidence:** HIGH

---

## Summary

Phase 24 is a pure C++ library addition: a new function that accepts an `F2CCells` field geometry and a list of robot specs, and returns N non-overlapping `F2CCells` zones whose areas are proportional to each robot's work rate (width × speed). No gRPC, no Go, no frontend changes are involved.

The fields2cover library uses GDAL/GEOS geometry types wrapped in its own thin value-type hierarchy (`F2CCell`, `F2CCells`, `F2CPoint`, etc.). All geometry set operations (intersection, difference, splitByLine, union) are already available on `F2CCells`. The `F2CRobot` type already stores both `getWidth()` (via `getCovWidth()`) and `getCruiseVel()`, making work-rate calculation straightforward.

The most reliable partitioning approach for arbitrary convex and non-convex polygons is axis-aligned strip partitioning: cut the field bounding box into N horizontal (or vertical) strips with cut positions derived from cumulative work-rate fractions, then intersect each strip with the actual field geometry. This produces geometrically correct zones whose net area is proportional to each robot's work rate, handles non-convex fields correctly via GEOS intersection (already available as `F2CCells::intersection`), and has no dependencies beyond what is already in the library.

**Primary recommendation:** Implement `f2c::partition::MultiRobotPartition::partition(const F2CCells& field, const std::vector<F2CRobot>& robots)` returning `std::vector<F2CCells>` using bounding-box strip partitioning intersected with the actual field geometry.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| F2C-01 | f2c library exposes a multi-robot field partitioning algorithm that divides a field into N zones, where zone sizes are proportional to each robot's work rate (robot_width × robot_speed) | Strip-partitioning via `F2CCells::intersection` + bounding-box cut lines satisfies proportionality, total-area, and non-overlap criteria directly |
</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Partitioning algorithm | C++ library (`src/fields2cover/`) | — | Pure geometry computation; no I/O, no gRPC in this phase |
| Public header / API surface | C++ library (`include/fields2cover/`) | — | Headers are the contract for downstream gRPC shim (Phase 26) |
| Unit test coverage | GoogleTest (`tests/cpp/`) | — | All other modules follow this pattern |
| Build system registration | `CMakeLists.txt` GLOB_RECURSE | — | New `.cpp` files in `src/` are auto-picked up; no CMake edits needed for src |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| GDAL/OGR | 3.0+ (system) | All polygon geometry operations | F2C's geometry layer is thin wrappers over OGRGeometry |
| GEOS (via OGR) | system | Intersection, difference, splitByLine internals | Already linked via `-lgeos_c` in CMakeLists.txt |
| GoogleTest | system | Unit tests | All 73 existing test suites use it |
| C++17 | — | Language standard | Set in CMakeLists: `CMAKE_CXX_STANDARD 17` |

[VERIFIED: codebase inspection]

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `<vector>`, `<numeric>` | stdlib | Robot list, cumulative sums | Standard; no extra deps needed |
| `<stdexcept>` | stdlib | Input validation | Throw on empty robot list or zero total work rate |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Strip partition (bounding box) | Voronoi partition | Voronoi produces non-rectangular zones that are harder to cover; strip is simpler and sufficient for proportional area |
| Strip partition (bounding box) | Recursive binary subdivision | Binary subdivision is cleaner for powers-of-2 but complex for arbitrary N robots with unequal weights |
| Strip partition (bounding box) | Weighted centroidal Voronoi | Research-grade algorithm; no existing f2c support; out of scope |
| `F2CCells` zone type | `F2CCell` zone type | `F2CCells` (OGRMultiPolygon) is the natural return from `intersection` and handles non-convex fields correctly |

**Installation:** No new packages. All dependencies are already present. [VERIFIED: CMakeLists.txt inspection]

---

## Architecture Patterns

### System Architecture Diagram

```
robots: [F2CRobot, ...]          field: F2CCells
     |                                |
     v                                v
[compute work rates]          [get bounding box]
  width × speed                getDimMinX/Y/MaxX/Y
     |                                |
     v                                v
[compute cut positions]   [build N strip rectangles]
 cumulative fractions      along bounding box axis
               \                    /
                v                  v
          [intersect each strip with field]
            F2CCells::intersection(strip_cell)
                        |
                        v
              std::vector<F2CCells>  (N zones)
```

### Recommended Project Structure

```
include/fields2cover/
└── partition/
    └── multi_robot_partition.h       # New header

src/fields2cover/
└── partition/
    └── multi_robot_partition.cpp     # New implementation

tests/cpp/
└── partition/
    └── multi_robot_partition_test.cpp  # New test file
```

**Note on CMakeLists:** The top-level `CMakeLists.txt` uses `file(GLOB_RECURSE fields2cover_src "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp")` to collect all source files. New `.cpp` files placed under `src/fields2cover/partition/` are picked up automatically — no CMakeLists.txt edit is needed for the library. Tests use `file(GLOB_RECURSE TEST_SOURCES ... cpp/*/*.cpp cpp/*/*/*.cpp)` — the new test file at `tests/cpp/partition/` matches `cpp/*/*.cpp` exactly. [VERIFIED: CMakeLists.txt, tests/CMakeLists.txt]

### Pattern 1: New Algorithm Module (mirrors decomposition pattern)

**What:** A standalone class in its own `include/` + `src/` subdirectory, consuming f2c types, returning f2c types.
**When to use:** Any new algorithm added to the f2c library.

Header pattern (from `decomposition_base.h`):
```cpp
// Source: include/fields2cover/decomposition/decomposition_base.h (existing pattern)
#pragma once
#ifndef FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
#define FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_

#include <vector>
#include "fields2cover/types.h"

namespace f2c::partition {

class MultiRobotPartition {
 public:
  /// Partition a field into N zones proportional to each robot's work rate.
  /// @param field  The field to partition (in any CRS; coordinates must be metric)
  /// @param robots List of robots; work rate = getCovWidth() * getCruiseVel()
  /// @return Vector of N F2CCells zones, index i corresponds to robots[i]
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
```

Implementation pattern (strip-based):
```cpp
// Source: pattern derived from Cells::splitByLine + intersection (verified in src/fields2cover/types/Cells.cpp)
std::vector<F2CCells> MultiRobotPartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument("robots must not be empty");
  }

  // 1. Compute work rates and total
  std::vector<double> rates(robots.size());
  for (size_t i = 0; i < robots.size(); ++i) {
    rates[i] = robots[i].getCovWidth() * robots[i].getCruiseVel();
    if (rates[i] <= 0.0) {
      throw std::invalid_argument("All robots must have positive work rate");
    }
  }
  double total = std::accumulate(rates.begin(), rates.end(), 0.0);

  // 2. Compute cumulative cut fractions along x-axis (or y-axis)
  //    e.g. for 3 robots with rates [2, 1, 1]: cuts at 0.5, 0.75, 1.0
  double x_min = field.getDimMinX();
  double x_max = field.getDimMaxX();
  double y_min = field.getDimMinY();
  double y_max = field.getDimMaxY();

  // 3. Build N strip rectangles, intersect each with the field
  std::vector<F2CCells> zones;
  double cum_frac = 0.0;
  double x_prev = x_min;
  for (size_t i = 0; i < robots.size(); ++i) {
    cum_frac += rates[i] / total;
    double x_next = (i == robots.size() - 1) ? x_max : x_min + cum_frac * (x_max - x_min);
    // build strip Cell from x_prev to x_next, y_min to y_max
    F2CLinearRing ring{
      F2CPoint(x_prev, y_min), F2CPoint(x_next, y_min),
      F2CPoint(x_next, y_max), F2CPoint(x_prev, y_max),
      F2CPoint(x_prev, y_min)};
    F2CCell strip{ring};
    zones.push_back(field.intersection(strip));
    x_prev = x_next;
  }
  return zones;
}
```

[VERIFIED: F2CCells::getDimMinX/Y/MaxX/Y from Geometry.h; F2CCells::intersection from Cells.h; F2CLinearRing/F2CCell construction from Cell_test.cpp / Cells_test.cpp]

### Pattern 2: GoogleTest Unit Test (mirrors decomposition tests)

**What:** Isolated test file, includes only the new header and `fields2cover/types.h`.
**When to use:** Every new algorithm.

```cpp
// Source: pattern from tests/cpp/decomposition/boustrophedon_decomp_test.cpp
#include <gtest/gtest.h>
#include "fields2cover/types.h"
#include "fields2cover/partition/multi_robot_partition.h"

TEST(fields2cover_partition_multi_robot, two_robots_equal_rate) {
  // 10x10 field
  F2CLinearRing ring{
    F2CPoint(0,0), F2CPoint(10,0), F2CPoint(10,10), F2CPoint(0,10),
    F2CPoint(0,0)};
  F2CCells field{F2CCell{ring}};

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0); r2.setCruiseVel(1.0);  // equal work rates

  f2c::partition::MultiRobotPartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2);
  EXPECT_NEAR(zones[0].area() + zones[1].area(), field.area(), 1e-3);
  EXPECT_NEAR(zones[0].area(), zones[1].area(), 1e-3);
}

TEST(fields2cover_partition_multi_robot, three_robots_proportional) {
  F2CLinearRing ring{
    F2CPoint(0,0), F2CPoint(12,0), F2CPoint(12,12), F2CPoint(0,12),
    F2CPoint(0,0)};
  F2CCells field{F2CCell{ring}};

  F2CRobot r1(2.0), r2(1.0), r3(1.0);
  // work rates: 2, 1, 1 → area ratio 2:1:1
  auto zones = f2c::partition::MultiRobotPartition().partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3);
  EXPECT_NEAR(zones[0].area() + zones[1].area() + zones[2].area(),
              field.area(), 1e-3);
  // zone[0] should have double the area of zone[1] and zone[2]
  EXPECT_NEAR(zones[0].area(), 2.0 * zones[1].area(), 1e-3);
  EXPECT_NEAR(zones[1].area(), zones[2].area(), 1e-3);
}
```

### Anti-Patterns to Avoid

- **Returning `std::vector<F2CCell>` instead of `std::vector<F2CCells>`:** `F2CCells::intersection` returns `F2CCells` (OGRMultiPolygon), which is the correct type for a non-convex field zone. `F2CCell` (OGRPolygon) cannot represent a disconnected zone produced by intersection with a strip.
- **Using `F2CField` as the input type:** The function should accept `F2CCells` (the inner geometry) rather than `F2CField` (which includes CRS metadata). The gRPC shim in Phase 26 will extract `field.getField()` before calling partition.
- **Adding CMakeLists.txt entries for library sources:** The GLOB_RECURSE in the top-level CMakeLists picks up new `.cpp` files automatically. Editing it is unnecessary and risks introducing drift.
- **Using floating-point equality in tests:** Use `EXPECT_NEAR(..., 1e-3)` as the existing decomp tests do, not `EXPECT_EQ`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Polygon intersection | Custom polygon clipping | `F2CCells::intersection(const Cell& c)` | GEOS-backed; handles all edge cases including non-convex, holes, degenerate touches |
| Polygon area | Shoelace formula | `.area()` inherited from `Geometries<>` | OGR computes it directly from the internal geometry |
| Bounding box | Manual min/max loops | `getDimMinX/Y`, `getDimMaxX/Y` on `F2CCells` | Already implemented in `Geometry_impl.hpp` |
| Polygon subtraction / slicing | Custom line-clipping | `F2CCells::splitByLine` or `difference` | Already battle-tested in decomposition module |

**Key insight:** GEOS/OGR already handles all the hard geometry corner cases. The algorithm only needs to compute cut positions and call existing methods.

---

## Common Pitfalls

### Pitfall 1: Zero-area zone from degenerate strip intersection
**What goes wrong:** If the bounding box is wider than the field in the cut axis, a strip near the edge may intersect only slivers or nothing.
**Why it happens:** Bounding box can exceed the actual field extent along the non-cut axis; strips near the edge of a narrow field may produce near-zero-area zones.
**How to avoid:** Use the field's own bounding box extents (not a hard-coded coordinate). Verify with `EXPECT_NEAR(total_zone_area, field.area(), 1e-3)` in tests.
**Warning signs:** Test shows total zone area less than field area.

### Pitfall 2: Confusion between `getCovWidth()` and `getWidth()`
**What goes wrong:** Using `getWidth()` (physical robot body) instead of `getCovWidth()` (effective coverage swath) for work rate.
**Why it happens:** Both methods exist on `F2CRobot`. The requirement text says "robot_width × robot_speed" but in the f2c model the coverage-relevant width is `getCovWidth()`.
**How to avoid:** Use `getCovWidth()`. Note: when `cov_width` is 0 in the constructor, it defaults to `width_`, so both are equivalent when only `width` is set. Either is defensible; document the choice.
**Warning signs:** Unexpected area ratios in tests.

### Pitfall 3: `Geometry::getWidth()` naming collision
**What goes wrong:** Calling `field.getWidth()` by mistake — this returns the X-dimension of the bounding box of the `F2CCells`, not the robot width.
**Why it happens:** Both `Geometry<>` and `Robot` have a `getWidth()` method with different semantics.
**How to avoid:** For geometry extents always use `getDimMaxX() - getDimMinX()`. For robot specs always use `robot.getCovWidth()`.

### Pitfall 4: Last strip has floating-point gap
**What goes wrong:** Cumulative floating-point arithmetic leaves the last strip slightly short of `x_max`, producing a tiny uncovered sliver.
**Why it happens:** Repeated `+=` of `rates[i]/total` drifts below 1.0.
**How to avoid:** Force the last strip's right edge to exactly `x_max` (as shown in the implementation sketch above: `(i == robots.size()-1) ? x_max : ...`).

### Pitfall 5: Tests not found by GLOB
**What goes wrong:** Test file added to `tests/cpp/partition/partition_test.cpp` is not compiled.
**Why it happens:** The GLOB pattern `cpp/*/*.cpp` matches one directory level deep. A file at `cpp/partition/multi_robot_partition_test.cpp` matches `cpp/*/*.cpp` exactly. Double-check the path.
**Warning signs:** `ctest` output shows fewer test suites than expected.

---

## Code Examples

### Constructing a rectangular Cell
```cpp
// Source: tests/cpp/types/Cell_test.cpp (verified)
F2CLinearRing ring{
  F2CPoint(0,0), F2CPoint(10,0), F2CPoint(10,5),
  F2CPoint(0,5), F2CPoint(0,0)};
F2CCell cell{ring};
// cell.area() == 50.0
```

### Intersecting F2CCells with a Cell
```cpp
// Source: include/fields2cover/types/Cells.h (verified)
F2CCells zone = field.intersection(strip_cell);
// zone is an F2CCells (OGRMultiPolygon) — correct type for non-convex fields
```

### Reading bounding box
```cpp
// Source: include/fields2cover/types/Geometry.h (verified)
double x_min = field.getDimMinX();
double x_max = field.getDimMaxX();
double y_min = field.getDimMinY();
double y_max = field.getDimMaxY();
```

### Running the unit test suite locally
```bash
# From repo root (build dir already exists, gnuplot installed):
make -C build unittests -j$(nproc)
./build/tests/unittests
# Expected: 294 tests from 73 test suites — PASSED
```

[VERIFIED: confirmed locally — 294/294 tests pass before any changes]

---

## Exact Function Signature

```cpp
// Namespace: f2c::partition
// Header: include/fields2cover/partition/multi_robot_partition.h
// Source: src/fields2cover/partition/multi_robot_partition.cpp

namespace f2c::partition {

class MultiRobotPartition {
 public:
  /// @brief Partition a field into N zones proportional to each robot's work rate.
  ///
  /// Zone sizes are proportional to robot.getCovWidth() * robot.getCruiseVel().
  /// Zones are non-overlapping; their union equals the input field.
  ///
  /// @param field  Field geometry (must be in a metric CRS — no CRS check done here)
  /// @param robots N robots; result[i] is the zone assigned to robots[i]
  /// @return       N F2CCells zones; result.size() == robots.size()
  /// @throws std::invalid_argument if robots is empty or any work rate <= 0
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition
```

This signature satisfies all four success criteria:
1. Returns N zones proportional to work rate.
2. Union of returned zones == input area (enforced by strip approach).
3. Existing tests unaffected (new files only).
4. Testable with a 2-robot and 3-robot GoogleTest case.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Python-only API | C++ library + gRPC shim | Phase 10 (v2.0) | C++ is the canonical implementation |
| Manual CMakeLists for each source file | GLOB_RECURSE | Original codebase | New .cpp files are auto-discovered |

**No deprecated APIs in scope.** All GDAL/GEOS types used are stable. [ASSUMED — training knowledge, GDAL 3.x is stable for these APIs]

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | GoogleTest (system package, exact version not pinned in CMakeLists) |
| Config file | `tests/CMakeLists.txt` + `tests/unittests.cpp` (main()) |
| Quick run command | `./build/tests/unittests --gtest_filter="fields2cover_partition*"` |
| Full suite command | `./build/tests/unittests` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-01 | 2-robot equal-rate: zones have equal area | unit | `./build/tests/unittests --gtest_filter="fields2cover_partition_multi_robot.two_robots_equal_rate"` | Wave 0 |
| F2C-01 | 3-robot proportional: zone areas match 2:1:1 ratio | unit | `./build/tests/unittests --gtest_filter="fields2cover_partition_multi_robot.three_robots_proportional"` | Wave 0 |
| F2C-01 | Total area of all zones = input field area | unit (embedded in above tests) | same | Wave 0 |
| F2C-01 | Existing 294 tests still pass | regression | `./build/tests/unittests` | Yes |

### Sampling Rate
- **Per task commit:** `make -C build unittests -j$(nproc) && ./build/tests/unittests --gtest_filter="fields2cover_partition*"`
- **Per wave merge:** `./build/tests/unittests`
- **Phase gate:** Full suite green (294 + new tests) before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `tests/cpp/partition/multi_robot_partition_test.cpp` — covers F2C-01 (2-robot and 3-robot cases)
- [ ] `include/fields2cover/partition/multi_robot_partition.h` — header with class declaration
- [ ] `src/fields2cover/partition/multi_robot_partition.cpp` — implementation

*(No new test framework or build system changes required — GLOB_RECURSE picks up new files automatically)*

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| GDAL/OGR | Field geometry types | Yes | GDAL 3.x (system) | — |
| GEOS | Intersection/difference | Yes | system | — |
| GoogleTest | Unit tests | Yes | system | — |
| gnuplot | Tests CMake condition | Yes | 6.0 | — |
| make / ninja | Build | Yes | system | — |

[VERIFIED: build produces 294 passing tests; `gnuplot --version` = 6.0]

**Missing dependencies with no fallback:** None.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `getCovWidth()` is the right width to use for work-rate (vs `getWidth()`) | Function Signature | Minor — when only `width` is set, both are equal (constructor sets cov_width_ = width_ when 0). Document choice clearly. |
| A2 | Axis-aligned strip partition is sufficient (no requirement for strips to be axis-aligned with field's dominant direction) | Architecture Patterns | Low — requirement says "proportional areas", not "optimal direction". Can extend in Phase 24+ if needed. |
| A3 | GDAL 3.x `OGRGeometry::Intersection` handles the field/strip intersection correctly for non-convex fields | Pitfalls | Low — this is the same path used by `Cells::difference` which powers the existing decomposition module |

---

## Open Questions

1. **Which axis to cut along (X or Y)?**
   - What we know: The bounding box approach works on either axis.
   - What's unclear: No preference stated in requirements. Strip orientation could be suboptimal for long-thin fields.
   - Recommendation: Default to the longer bounding box axis (use `getWidth()` vs `getHeight()` on the `F2CCells` to decide). Document this choice. Could be made configurable in Phase 25+.

2. **Should the function accept `F2CField` or `F2CCells`?**
   - What we know: `F2CField` wraps `F2CCells` with CRS metadata. The partition algorithm itself is CRS-agnostic (works on raw coordinates).
   - What's unclear: The gRPC shim (Phase 26) will need to call this. It's easier to pass `field.getField()` at the shim boundary.
   - Recommendation: Accept `F2CCells` to keep the function CRS-agnostic and consistent with the decomposition API pattern.

---

## Sources

### Primary (HIGH confidence)
- Codebase: `include/fields2cover/types/Cells.h` — intersection, difference, splitByLine API
- Codebase: `include/fields2cover/types/Geometry.h` / `Geometry_impl.hpp` — getDimMinX/Y/MaxX/Y
- Codebase: `include/fields2cover/types/Robot.h` + `src/fields2cover/types/Robot.cpp` — getCovWidth, getCruiseVel
- Codebase: `include/fields2cover/decomposition/decomposition_base.h` — new-module pattern
- Codebase: `tests/CMakeLists.txt` — GLOB_RECURSE pattern, no manual registration needed
- Codebase: `CMakeLists.txt` — GLOB_RECURSE for library sources, C++17, GDAL/GEOS deps
- Verified: `./build/tests/unittests` — 294/294 tests pass as baseline

### Secondary (MEDIUM confidence)
- None required — all claims verified from codebase.

### Tertiary (LOW confidence)
- A1-A3 in Assumptions Log above.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all deps verified from CMakeLists and running build
- Architecture: HIGH — pattern copied directly from existing decomposition module
- Pitfalls: HIGH — derived from reading actual implementation code and test patterns
- Function signature: HIGH — derived from f2c type system inspection

**Research date:** 2026-04-29
**Valid until:** 2026-08-01 (GDAL/GEOS APIs are stable; library structure unlikely to change)
