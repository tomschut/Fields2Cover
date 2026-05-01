# Phase 25c: C++ Obstacle Avoider - Pattern Map

**Mapped:** 2026-04-29
**Files analyzed:** 3 (header, source, test)
**Analogs found:** 3 / 3

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `include/fields2cover/obstacle/obstacle_avoider.h` | header | request-response | `include/fields2cover/partition/multi_robot_partition.h` | exact |
| `src/fields2cover/obstacle/obstacle_avoider.cpp` | service | transform | `src/fields2cover/partition/multi_robot_partition.cpp` | exact |
| `tests/cpp/obstacle/obstacle_avoider_test.cpp` | test | CRUD | `tests/cpp/partition/multi_robot_partition_test.cpp` | exact |

## Pattern Assignments

### `include/fields2cover/obstacle/obstacle_avoider.h` (header, request-response)

**Analog:** `include/fields2cover/partition/multi_robot_partition.h`

**Imports + include-guard pattern** (lines 1-13):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
#define FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"
```

**Namespace + class declaration pattern** (lines 15-37):
```cpp
namespace f2c::partition {

/// @brief Partition a field into N zones proportional to each robot's work rate.
class MultiRobotPartition {
 public:
  /// @brief <doc comment>
  ///
  /// @param field  Field geometry in any metric CRS; ...
  /// @param robots Non-empty list of robots; ...
  /// @return       std::vector<F2CCells> of size robots.size().
  /// @throws std::invalid_argument if ...
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
```

**Adaptation for ObstacleAvoider:**
- Replace guard macro: `FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_`
- Replace namespace: `f2c::obstacle`
- Replace class name: `ObstacleAvoider`
- Method signature:
  ```cpp
  static constexpr double kMinSegmentLength = 0.1;  // metres

  F2CSwaths avoid(const F2CSwaths& swaths,
                  const F2CCell& obstacle,
                  double safety_margin) const;
  ```
- Imports: `<vector>` and `<stdexcept>` (same); `"fields2cover/types.h"` (same)

---

### `src/fields2cover/obstacle/obstacle_avoider.cpp` (service, transform)

**Analog:** `src/fields2cover/partition/multi_robot_partition.cpp`

**File header + include pattern** (lines 1-12):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/multi_robot_partition.h"

#include <numeric>
#include <stdexcept>

namespace f2c::partition {
```

**Guard clause (validation) pattern** (lines 17-21):
```cpp
  if (robots.empty()) {
    throw std::invalid_argument(
        "MultiRobotPartition::partition: robots must not be empty");
  }
```

**Bounding-box construction pattern** (lines 35-39):
```cpp
  const double x_min = field.getDimMinX();
  const double x_max = field.getDimMaxX();
  const double y_min = field.getDimMinY();
  const double y_max = field.getDimMaxY();
```

**F2CLinearRing + F2CCell construction pattern** (lines 66-72):
```cpp
    F2CLinearRing ring;
    ring = F2CLinearRing{
        F2CPoint(prev, lo), F2CPoint(next, lo),
        F2CPoint(next, hi), F2CPoint(prev, hi),
        F2CPoint(prev, lo)};
    zones.push_back(field.intersection(F2CCell{ring}));
```

**Full algorithm shape for ObstacleAvoider::avoid():**
```cpp
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

  // 1. Inflate the obstacle polygon.
  F2CCell inflated = F2CCell::buffer(obstacle, safety_margin);

  F2CSwaths result;
  int out_id = 0;

  for (size_t i = 0; i < swaths.size(); ++i) {
    const F2CSwath& sw = swaths[i];
    F2CLineString path = sw.getPath();

    // 2. Build a bounding-box Cells that covers the swath with padding,
    //    then subtract the inflated obstacle to get the safe region.
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
    F2CMultiLineString residuals = safe.getLinesInside(path);

    // 4. Filter by minimum length and emit new swaths.
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

**Key API calls verified in codebase:**
- `F2CCell::buffer(obstacle, margin)` — `include/fields2cover/types/Cell.h` line 54
- `F2CCells::difference(F2CCell)` — `include/fields2cover/types/Cells.h` line 55
- `F2CCells::getLinesInside(F2CLineString)` — `include/fields2cover/types/Cells.h` line 71
- `F2CMultiLineString::getGeometry(size_t)` — `include/fields2cover/types/MultiLineString.h` line 38
- `F2CLineString::length()` — via `Geometry` base; `F2CSwath::length()` at `include/fields2cover/types/Swath.h` line 58
- `F2CSwaths::emplace_back(LineString, width, id, type)` — `include/fields2cover/types/Swaths.h` line 34
- `F2CLineString::getDimMinX/getDimMaxX/getDimMinY/getDimMaxY()` — `Geometry` base (same interface used on `F2CCells` in `multi_robot_partition.cpp` lines 36-39)

---

### `tests/cpp/obstacle/obstacle_avoider_test.cpp` (test)

**Analog:** `tests/cpp/partition/multi_robot_partition_test.cpp`

**File header + includes pattern** (lines 1-11):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/partition/multi_robot_partition.h"
```

**Anonymous namespace + helper pattern** (lines 12-22):
```cpp
namespace {

/// Helper: create a rectangular F2CCells with corners (0,0)-(w,h)
F2CCells makeRect(double w, double h) {
  F2CLinearRing ring{
    F2CPoint(0, 0), F2CPoint(w, 0),
    F2CPoint(w, h), F2CPoint(0, h),
    F2CPoint(0, 0)};
  return F2CCells{F2CCell{ring}};
}

}  // namespace
```

**TEST macro pattern** (lines 25-38):
```cpp
TEST(fields2cover_partition_multi_robot, two_robots_equal_rate) {
  F2CCells field = makeRect(10.0, 10.0);
  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);

  f2c::partition::MultiRobotPartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_NEAR(zones[0].area() + zones[1].area(), field.area(), 1e-3);
}
```

**EXPECT_THROW pattern** (lines 59-63):
```cpp
TEST(fields2cover_partition_multi_robot, empty_robots_throws) {
  F2CCells field = makeRect(10.0, 10.0);
  f2c::partition::MultiRobotPartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}
```

**Adaptation for obstacle_avoider_test.cpp:**

Test name prefix: `fields2cover_obstacle_avoider`

Helper functions to write:
```cpp
namespace {

// Build a straight horizontal swath along y=0, x in [x0, x1]
F2CSwaths makeStraightSwaths(double x0, double x1, double width = 3.0) {
  F2CLineString path{F2CPoint(x0, 0.0), F2CPoint(x1, 0.0)};
  F2CSwaths s;
  s.emplace_back(path, width);
  return s;
}

// Build a square F2CCell obstacle centred at (cx, cy) with half-side r
F2CCell makeSquareObstacle(double cx, double cy, double r) {
  F2CLinearRing ring{
    F2CPoint(cx-r, cy-r), F2CPoint(cx+r, cy-r),
    F2CPoint(cx+r, cy+r), F2CPoint(cx-r, cy+r),
    F2CPoint(cx-r, cy-r)};
  return F2CCell{ring};
}

}  // namespace
```

Tests to implement:
1. `obstacle_splits_swath_into_two` — swath x∈[0,20], y=0; square obstacle at x=10 half-side 2; safety_margin=0 → ASSERT_EQ(result.size(), 2u)
2. `safety_margin_enlarges_exclusion` — same geometry; safety_margin=1.0 → result segments shorter than no-margin case
3. `short_segments_dropped` — obstacle nearly covers entire swath leaving only tiny residuals → EXPECT_EQ(result.size(), 0u)
4. `negative_margin_throws` — EXPECT_THROW with std::invalid_argument
5. `no_obstacle_overlap_returns_full_swath` — obstacle far from swath → result.size()==1 and result[0].length() near original

---

## Shared Patterns

### License Header
**Apply to:** All three new files.
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================
```

### Include Guard Convention
**Source:** `include/fields2cover/partition/multi_robot_partition.h` lines 7-8
**Apply to:** `obstacle_avoider.h`

Pattern: `#pragma once` followed by `#ifndef FIELDS2COVER_<MODULE>_<CLASS>_H_`

For ObstacleAvoider:
```cpp
#pragma once
#ifndef FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
#define FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
// ...
#endif  // FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
```

### Namespace Convention
**Source:** All partition headers/sources.
**Apply to:** Header and source.

- Header declares: `namespace f2c::obstacle { ... }  // namespace f2c::obstacle`
- Source opens:    `namespace f2c::obstacle { ... }  // namespace f2c::obstacle`

### Error Handling (std::invalid_argument)
**Source:** `src/fields2cover/partition/multi_robot_partition.cpp` lines 17-20
**Apply to:** `obstacle_avoider.cpp`

```cpp
  if (safety_margin < 0.0) {
    throw std::invalid_argument(
        "ObstacleAvoider::avoid: safety_margin must be >= 0");
  }
```

### F2CLinearRing Constructor Syntax
**Source:** `tests/cpp/partition/multi_robot_partition_test.cpp` lines 16-21 and `src/fields2cover/partition/multi_robot_partition.cpp` lines 66-72
**Apply to:** Both source (bbox construction) and test (helper functions).

Initializer-list syntax with closing point repeating the first:
```cpp
F2CLinearRing ring{
    F2CPoint(x0, y0), F2CPoint(x1, y0),
    F2CPoint(x1, y1), F2CPoint(x0, y1),
    F2CPoint(x0, y0)};   // closing point = first point
F2CCell cell{ring};
F2CCells cells(cell);    // or F2CCells{cell}
```

### Test Suite Naming
**Source:** `tests/cpp/partition/multi_robot_partition_test.cpp` line 25
**Apply to:** `obstacle_avoider_test.cpp`

Pattern: `TEST(fields2cover_<module>_<class_snake>, <scenario_snake>)`

For ObstacleAvoider: `TEST(fields2cover_obstacle_avoider, <scenario>)`

### CMakeLists.txt Auto-Discovery (no edits needed)
**Source:** `CMakeLists.txt` line 68-70 and `tests/CMakeLists.txt` lines 8-9
**Verified pattern:**
```cmake
# Library: auto-discovers all src/*.cpp recursively
file(GLOB_RECURSE fields2cover_src "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp")

# Tests: auto-discovers cpp/*/*.cpp and cpp/*/*/*.cpp
file(GLOB_RECURSE TEST_SOURCES LIST_DIRECTORIES false unittests.cpp cpp/*/*.cpp cpp/*/*/*.cpp)
```

New files in `src/fields2cover/obstacle/` and `tests/cpp/obstacle/` match these glob patterns automatically. **No CMakeLists.txt edits are needed.** However, cmake must be re-run after creating new directories (GLOB_RECURSE is cached at configure time, not re-evaluated at make time).

## No Analog Found

None — all three files have exact analogs in the `partition/` module which follows the identical three-file pattern (header, source, test).

## Metadata

**Analog search scope:** `include/fields2cover/`, `src/fields2cover/`, `tests/cpp/`
**Files scanned:** 8 source files read in full; 2 CMakeLists.txt read for glob patterns
**Pattern extraction date:** 2026-04-29
