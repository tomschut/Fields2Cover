---
phase: 25a-c-partition-strategies
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - include/fields2cover/partition/spatial_rtree_partition.h
  - src/fields2cover/partition/spatial_rtree_partition.cpp
  - include/fields2cover/partition/length_balanced_partition.h
  - src/fields2cover/partition/length_balanced_partition.cpp
  - tests/cpp/partition/spatial_rtree_partition_test.cpp
  - tests/cpp/partition/length_balanced_partition_test.cpp
autonomous: true
requirements:
  - F2C-03

must_haves:
  truths:
    - "Calling SpatialRtreePartition::partition() with a 20x20 field and 2 equal robots returns 2 non-empty zones whose areas sum to within 5% of the field area"
    - "Calling SpatialRtreePartition::partition() with a 30x30 field and 3 robots returns 3 non-empty zones"
    - "Calling LengthBalancedPartition::partition() with a 20x20 field and 2 equal robots returns 2 zones whose swath-length difference is within 10% of total swath length"
    - "Calling LengthBalancedPartition::partition() with a 30x30 field and 3 robots returns 3 non-empty zones whose max load minus min load is within 10% of total swath length"
    - "Both strategies throw std::invalid_argument when robots is empty"
    - "Both strategies throw std::invalid_argument when robots[0].getCovWidth() <= 0"
    - "All 305 existing GoogleTest tests continue to pass after new files are added"
    - "Each new strategy has at least 4 unit tests (2-robot, 3-robot, empty-throws, zero-width-throws)"
  artifacts:
    - path: "include/fields2cover/partition/spatial_rtree_partition.h"
      provides: "Public class declaration for SpatialRtreePartition in namespace f2c::partition"
      exports: ["SpatialRtreePartition", "partition"]
    - path: "src/fields2cover/partition/spatial_rtree_partition.cpp"
      provides: "R-tree clustering implementation — boost::geometry includes only in .cpp"
      contains: "bgi::rtree"
    - path: "include/fields2cover/partition/length_balanced_partition.h"
      provides: "Public class declaration for LengthBalancedPartition in namespace f2c::partition"
      exports: ["LengthBalancedPartition", "partition"]
    - path: "src/fields2cover/partition/length_balanced_partition.cpp"
      provides: "Greedy FFD length-balanced assignment — no boost headers needed"
      contains: "std::min_element"
    - path: "tests/cpp/partition/spatial_rtree_partition_test.cpp"
      provides: "GoogleTest cases for 2-robot and 3-robot SpatialRtree scenarios"
      contains: "fields2cover_partition_spatial_rtree"
    - path: "tests/cpp/partition/length_balanced_partition_test.cpp"
      provides: "GoogleTest cases for 2-robot and 3-robot LengthBalanced scenarios"
      contains: "fields2cover_partition_length_balanced"
  key_links:
    - from: "src/fields2cover/partition/spatial_rtree_partition.cpp"
      to: "include/fields2cover/partition/spatial_rtree_partition.h"
      via: "#include"
      pattern: "#include.*partition/spatial_rtree_partition\\.h"
    - from: "src/fields2cover/partition/spatial_rtree_partition.cpp"
      to: "boost::geometry::index::rtree"
      via: "#include <boost/geometry/index/rtree.hpp>"
      pattern: "bgi::rtree"
    - from: "spatial_rtree_partition.cpp"
      to: "f2c::sg::BruteForce::generateBestSwaths"
      via: "swath generation"
      pattern: "generateBestSwaths"
    - from: "length_balanced_partition.cpp"
      to: "F2CSwath::length()"
      via: "load tracking"
      pattern: "swaths\\[.*\\]\\.length()"
    - from: "both .cpp files"
      to: "F2CSwath::areaCovered() + F2CCells::unionOp()"
      via: "zone geometry construction"
      pattern: "areaCovered"
---

<objective>
Implement two new multi-robot field partition strategies in the fields2cover C++ library:

- `SpatialRtreePartition` — R-tree spatial proximity clustering: generates swaths, builds a boost::geometry R-tree of swath midpoints, assigns each swath to the nearest robot seed cluster, unions swath coverage areas per robot into zone geometry.
- `LengthBalancedPartition` — greedy length-balanced assignment: generates swaths, sorts by length descending (FFD), greedily assigns each swath to the robot with minimum current load, unions swath coverage areas per robot into zone geometry.

Both classes live in namespace `f2c::partition`, follow the Phase 24 `MultiRobotPartition` pattern exactly (pragma once + include guard, BSD-3 license header, no CMakeLists edits), and expose the same `partition(const F2CCells&, const std::vector<F2CRobot>&)` → `std::vector<F2CCells>` signature.

Purpose: Satisfies F2C-03 (SPATIAL_RTREE and LENGTH_BALANCED partition strategies). Phase 26 will expose both through gRPC.

Output:
- `include/fields2cover/partition/spatial_rtree_partition.h` + `src/.../spatial_rtree_partition.cpp`
- `include/fields2cover/partition/length_balanced_partition.h` + `src/.../length_balanced_partition.cpp`
- `tests/cpp/partition/spatial_rtree_partition_test.cpp` + `tests/cpp/partition/length_balanced_partition_test.cpp`
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/ROADMAP.md
@.planning/phases/25a-c-partition-strategies/25a-RESEARCH.md

<interfaces>
<!-- Key types and contracts the executor needs. Extracted from codebase. No exploration needed. -->

From include/fields2cover/partition/multi_robot_partition.h (Phase 24 — mirror this pattern):
```cpp
#pragma once
#ifndef FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
#define FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"
namespace f2c::partition {
class MultiRobotPartition {
 public:
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};
}  // namespace f2c::partition
#endif
```

From include/fields2cover/types/Cells.h:
```cpp
const Cell getCell(size_t i) const;   // extract single cell for generateBestSwaths(obj,width,cell)
size_t size() const;                   // number of cells in the collection
double area() const;                   // total area
Cells unionOp(const Cell& c) const;
Cells unionOp(const Cells& c) const;  // GEOS union — use for zone geometry accumulation
```

From include/fields2cover/types/Swath.h:
```cpp
double getWidth() const;
double length() const;              // centre-line length in metres
Point startPoint() const;           // F2CPoint (getX()/getY())
Point endPoint() const;             // F2CPoint (getX()/getY())
Cells areaCovered() const;          // F2CCells polygon swept by this swath
```

From include/fields2cover/types/Point.h:
```cpp
double getX() const;
double getY() const;
// operator+ and operator* defined — midpoint: (s.startPoint() + s.endPoint()) * 0.5 works
```

From include/fields2cover/types/Robot.h:
```cpp
explicit F2CRobot(double width);    // sets cov_width_ = width
double getCovWidth() const;         // coverage swath width (use as op_width for swath gen)
double getCruiseVel() const;
```

From include/fields2cover/swath_generator/swath_generator_base.h:
```cpp
// Single-cell overload (use this — pass field.getCell(0)):
virtual F2CSwaths generateBestSwaths(f2c::obj::SGObjective& obj,
    double op_width, const F2CCell& cell);
// Multi-cell overload (returns F2CSwathsByCells — do NOT use):
virtual F2CSwaths generateBestSwaths(f2c::obj::SGObjective& obj,
    double op_width, const F2CCells& cells);
```

boost::geometry R-tree (confirmed at /usr/include/boost/geometry/index/rtree.hpp):
```cpp
// Include ONLY in .cpp files — never in .h (compile-time cost + downstream contamination)
#include <boost/geometry.hpp>
#include <boost/geometry/index/rtree.hpp>
namespace bg  = boost::geometry;
namespace bgi = boost::geometry::index;
using BgPoint = bg::model::point<double, 2, bg::cs::cartesian>;
using Value   = std::pair<BgPoint, std::size_t>;  // (midpoint, swath_index)
bgi::rtree<Value, bgi::quadratic<16>> rtree;
// Insert:  rtree.insert({BgPoint(p.getX(), p.getY()), idx});
// Query:   rtree.query(bgi::nearest(BgPoint(x,y), 1), std::back_inserter(result));
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement SpatialRtreePartition and LengthBalancedPartition headers and sources</name>
  <files>
    include/fields2cover/partition/spatial_rtree_partition.h
    src/fields2cover/partition/spatial_rtree_partition.cpp
    include/fields2cover/partition/length_balanced_partition.h
    src/fields2cover/partition/length_balanced_partition.cpp
  </files>

  <read_first>
    - include/fields2cover/partition/multi_robot_partition.h         (guard + namespace pattern to mirror exactly)
    - src/fields2cover/partition/multi_robot_partition.cpp           (source pattern: guard, validation, structure)
    - include/fields2cover/swath_generator/swath_generator_base.h   (generateBestSwaths overloads — use F2CCell not F2CCells)
    - include/fields2cover/objectives/sg_obj/swath_length.h         (f2c::obj::SwathLength — needed by generateBestSwaths)
    - include/fields2cover/types/Swath.h                            (length(), startPoint(), endPoint(), areaCovered())
    - include/fields2cover/types/Cells.h                            (getCell(0), size(), unionOp(), area())
    - CMakeLists.txt                                                 (confirm GLOB_RECURSE src/*.cpp — do NOT edit)
  </read_first>

  <behavior>
    SpatialRtreePartition::partition:
    - partition({}, {}) — empty robots → throws std::invalid_argument
    - partition(field, {robot_with_zero_width}) — getCovWidth()<=0 → throws std::invalid_argument
    - partition(field_with_cells_size_!=1, robots) — field.size()!=1 → throws std::invalid_argument
    - partition(20x20 field, [r1(3.0), r2(3.0)]) → returns vector of size 2; both zones non-empty; zones[0].area() + zones[1].area() within 5% of field.area()
    - partition(30x30 field, [r1(3.0), r2(3.0), r3(3.0)]) → returns vector of size 3; all zones non-empty

    LengthBalancedPartition::partition:
    - partition({}, {}) — empty robots → throws std::invalid_argument
    - partition(field, {robot_with_zero_width}) → throws std::invalid_argument
    - partition(field_with_cells_size_!=1, robots) → throws std::invalid_argument
    - partition(20x20 field, [r1(3.0), r2(3.0)]) → returns vector of size 2; |load[0]-load[1]| / total_swath_length <= 0.10
    - partition(30x30 field, [r1(3.0), r2(3.0), r3(3.0)]) → returns vector of size 3; (max_load - min_load) / total_swath_length <= 0.10
  </behavior>

  <action>
Write the following four files verbatim. Do NOT edit any CMakeLists.txt — GLOB_RECURSE auto-discovers the new .cpp files.

---

File: `include/fields2cover/partition/spatial_rtree_partition.h`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
#define FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

/// @brief Partition a field into N zones by spatial proximity clustering of swaths.
///
/// The field is first tessellated into swaths using f2c::sg::BruteForce with
/// robots[0].getCovWidth() as the swath width. N seed swaths (evenly spaced by
/// index) are selected — one per robot. A boost::geometry R-tree is built over all
/// swath midpoints. Each remaining swath is assigned to the robot whose current
/// cluster centroid is nearest (queried via R-tree nearest-neighbour). Zone geometry
/// is derived by unioning each robot's assigned swaths' areaCovered() results.
///
/// Assumptions:
///   - Input field must be a single-cell geometry (field.size() == 1).
///   - All robots are assumed to have the same coverage width; swath generation
///     uses robots[0].getCovWidth(). If robots have different widths the zone
///     geometry is approximate.
///
/// boost::geometry R-tree headers are included only in the .cpp file to avoid
/// compile-time cost and include-path contamination in downstream translation units.
class SpatialRtreePartition {
 public:
  /// @brief Partition a field into N spatially compact zones.
  ///
  /// @param field  Single-cell field geometry in any metric CRS.
  /// @param robots Non-empty robot list; result[i] is the zone for robots[i].
  /// @return       std::vector<F2CCells> of size robots.size().
  /// @throws std::invalid_argument if robots is empty, robots[0].getCovWidth() <= 0,
  ///         or field.size() != 1.
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
```

---

File: `src/fields2cover/partition/spatial_rtree_partition.cpp`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/spatial_rtree_partition.h"

#include <algorithm>
#include <numeric>
#include <vector>

// boost::geometry R-tree — included only here to avoid header contamination.
#include <boost/geometry.hpp>
#include <boost/geometry/index/rtree.hpp>

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/swath_length.h"

namespace f2c::partition {

namespace bg  = boost::geometry;
namespace bgi = boost::geometry::index;

using BgPoint = bg::model::point<double, 2, bg::cs::cartesian>;
using Value   = std::pair<BgPoint, std::size_t>;  // (midpoint, swath_index)

std::vector<F2CCells> SpatialRtreePartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: robots must not be empty");
  }
  if (robots[0].getCovWidth() <= 0.0) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: robots[0] must have positive coverage width");
  }
  if (field.size() != 1) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: field must be a single-cell geometry (field.size() == 1)");
  }

  const std::size_t N = robots.size();

  // 1. Generate swaths from the single-cell field.
  f2c::sg::BruteForce sw_gen;
  f2c::obj::SwathLength obj;
  F2CSwaths swaths = sw_gen.generateBestSwaths(obj, robots[0].getCovWidth(), field.getCell(0));

  // 2. Build R-tree of all swath midpoints.
  bgi::rtree<Value, bgi::quadratic<16>> rtree;
  for (std::size_t i = 0; i < swaths.size(); ++i) {
    F2CPoint s = swaths[i].startPoint();
    F2CPoint e = swaths[i].endPoint();
    BgPoint mid((s.getX() + e.getX()) * 0.5, (s.getY() + e.getY()) * 0.5);
    rtree.insert({mid, i});
  }

  // 3. Select N evenly-spaced seed swaths — one per robot.
  std::vector<std::vector<std::size_t>> assigned(N);
  std::vector<bool> is_seed(swaths.size(), false);

  for (std::size_t r = 0; r < N; ++r) {
    std::size_t seed_idx = (swaths.size() == 0)
        ? 0
        : (r * swaths.size()) / N;
    if (seed_idx < swaths.size()) {
      assigned[r].push_back(seed_idx);
      is_seed[seed_idx] = true;
    }
  }

  // 4. Compute per-robot cluster centroid from assigned swaths.
  auto centroid = [&](std::size_t robot_idx) -> BgPoint {
    double cx = 0.0, cy = 0.0;
    std::size_t cnt = 0;
    for (std::size_t idx : assigned[robot_idx]) {
      F2CPoint s = swaths[idx].startPoint();
      F2CPoint e = swaths[idx].endPoint();
      cx += (s.getX() + e.getX()) * 0.5;
      cy += (s.getY() + e.getY()) * 0.5;
      ++cnt;
    }
    if (cnt == 0) {
      return BgPoint(0.0, 0.0);
    }
    return BgPoint(cx / cnt, cy / cnt);
  };

  // 5. Assign all non-seed swaths to the robot with the nearest cluster centroid.
  for (std::size_t i = 0; i < swaths.size(); ++i) {
    if (is_seed[i]) continue;

    F2CPoint s = swaths[i].startPoint();
    F2CPoint e = swaths[i].endPoint();
    BgPoint mid((s.getX() + e.getX()) * 0.5, (s.getY() + e.getY()) * 0.5);

    double best_dist = std::numeric_limits<double>::max();
    std::size_t best_robot = 0;

    for (std::size_t r = 0; r < N; ++r) {
      BgPoint c = centroid(r);
      double dx = bg::get<0>(mid) - bg::get<0>(c);
      double dy = bg::get<1>(mid) - bg::get<1>(c);
      double d2 = dx * dx + dy * dy;
      if (d2 < best_dist) {
        best_dist = d2;
        best_robot = r;
      }
    }
    assigned[best_robot].push_back(i);
  }

  // 6. Build zone geometry: union each robot's swaths' areaCovered().
  std::vector<F2CCells> zones;
  zones.reserve(N);
  for (std::size_t r = 0; r < N; ++r) {
    F2CCells zone;
    for (std::size_t idx : assigned[r]) {
      F2CCells covered = swaths[idx].areaCovered();
      zone = zone.unionOp(covered);
    }
    zones.push_back(zone);
  }

  return zones;
}

}  // namespace f2c::partition
```

---

File: `include/fields2cover/partition/length_balanced_partition.h`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_
#define FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

/// @brief Partition a field into N zones by greedy length-balanced swath assignment.
///
/// The field is tessellated into swaths using f2c::sg::BruteForce with
/// robots[0].getCovWidth() as the swath width. Swaths are sorted by length()
/// descending (first-fit decreasing). Each swath is greedily assigned to the
/// robot with the minimum current load (total swath-metres assigned so far).
/// Zone geometry is derived by unioning each robot's assigned swaths' areaCovered().
///
/// Balance guarantee: FFD achieves (max_load - min_load) / total_length <= 10% for
/// typical uniform-length swath sets. If a single swath is longer than 10% of
/// total swath length, the guarantee cannot be met — the result is still valid but
/// the imbalance may exceed 10%.
///
/// Assumptions:
///   - Input field must be a single-cell geometry (field.size() == 1).
///   - Swath width uses robots[0].getCovWidth().
class LengthBalancedPartition {
 public:
  /// @brief Partition a field into N length-balanced zones.
  ///
  /// @param field  Single-cell field geometry in any metric CRS.
  /// @param robots Non-empty robot list; result[i] is the zone for robots[i].
  /// @return       std::vector<F2CCells> of size robots.size().
  /// @throws std::invalid_argument if robots is empty, robots[0].getCovWidth() <= 0,
  ///         or field.size() != 1.
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_
```

---

File: `src/fields2cover/partition/length_balanced_partition.cpp`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/length_balanced_partition.h"

#include <algorithm>
#include <numeric>
#include <vector>

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/swath_length.h"

namespace f2c::partition {

std::vector<F2CCells> LengthBalancedPartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: robots must not be empty");
  }
  if (robots[0].getCovWidth() <= 0.0) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: robots[0] must have positive coverage width");
  }
  if (field.size() != 1) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: field must be a single-cell geometry (field.size() == 1)");
  }

  const std::size_t N = robots.size();

  // 1. Generate swaths from the single-cell field.
  f2c::sg::BruteForce sw_gen;
  f2c::obj::SwathLength obj;
  F2CSwaths swaths = sw_gen.generateBestSwaths(obj, robots[0].getCovWidth(), field.getCell(0));

  // 2. Sort swath indices by length() descending (first-fit decreasing).
  std::vector<std::size_t> order(swaths.size());
  std::iota(order.begin(), order.end(), 0);
  std::sort(order.begin(), order.end(), [&](std::size_t a, std::size_t b) {
    return swaths[a].length() > swaths[b].length();
  });

  // 3. Greedy min-load assignment.
  std::vector<double> load(N, 0.0);
  std::vector<std::vector<std::size_t>> assigned(N);

  for (std::size_t idx : order) {
    auto it = std::min_element(load.begin(), load.end());
    std::size_t robot_idx = static_cast<std::size_t>(
        std::distance(load.begin(), it));
    assigned[robot_idx].push_back(idx);
    load[robot_idx] += swaths[idx].length();
  }

  // 4. Build zone geometry: union each robot's swaths' areaCovered().
  std::vector<F2CCells> zones;
  zones.reserve(N);
  for (std::size_t r = 0; r < N; ++r) {
    F2CCells zone;
    for (std::size_t idx : assigned[r]) {
      F2CCells covered = swaths[idx].areaCovered();
      zone = zone.unionOp(covered);
    }
    zones.push_back(zone);
  }

  return zones;
}

}  // namespace f2c::partition
```
  </action>

  <verify>
    After writing the four files, re-run cmake (required — GLOB_RECURSE result is stale after new source files are added) then build:
    ```bash
    cmake -S /home/tom/devenv/fields2cover -B /home/tom/devenv/fields2cover/build && \
    make -C /home/tom/devenv/fields2cover/build unittests -j$(nproc) 2>&1 | tail -10
    ```
    Then confirm all 305 pre-existing tests still pass (new tests come in Task 2):
    ```bash
    /home/tom/devenv/fields2cover/build/tests/unittests 2>&1 | tail -5
    ```
  </verify>

  <acceptance_criteria>
    - `include/fields2cover/partition/spatial_rtree_partition.h` exists
    - File contains `class SpatialRtreePartition`
    - File contains `namespace f2c::partition`
    - File contains `std::vector<F2CCells> partition(`
    - File contains `#ifndef FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_`
    - `src/fields2cover/partition/spatial_rtree_partition.cpp` exists
    - File contains `bgi::rtree`
    - File contains `generateBestSwaths`
    - File contains `areaCovered`
    - File does NOT contain `#include <boost/geometry` in any .h file (boost includes only in .cpp)
    - `include/fields2cover/partition/length_balanced_partition.h` exists
    - File contains `class LengthBalancedPartition`
    - File contains `#ifndef FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_`
    - `src/fields2cover/partition/length_balanced_partition.cpp` exists
    - File contains `std::min_element`
    - File contains `swaths[idx].length()`
    - File contains `areaCovered`
    - `cmake -S . -B build && make -C build unittests -j$(nproc)` exits 0
    - `./build/tests/unittests` exits 0 with at least 305 tests PASSED (no regressions)
  </acceptance_criteria>

  <done>All four files exist; cmake re-run succeeds; library builds clean; 305 pre-existing tests still pass.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Write GoogleTest unit tests for both partition strategies</name>
  <files>
    tests/cpp/partition/spatial_rtree_partition_test.cpp
    tests/cpp/partition/length_balanced_partition_test.cpp
  </files>

  <read_first>
    - tests/cpp/partition/multi_robot_partition_test.cpp          (test file pattern to mirror exactly — helper makeRect, test naming)
    - include/fields2cover/partition/spatial_rtree_partition.h    (class under test)
    - include/fields2cover/partition/length_balanced_partition.h  (class under test)
    - tests/CMakeLists.txt                                        (confirm GLOB pattern covers tests/cpp/partition/*.cpp — do NOT edit)
  </read_first>

  <behavior>
    SpatialRtreePartition tests:
    - two_robots: 20x20 field, r1(3.0,1.0), r2(3.0,1.0) → zones.size()==2; both areas > 0; sum within 5% of 400.0
    - three_robots: 30x30 field, r1(3.0), r2(3.0), r3(3.0) → zones.size()==3; all areas > 0
    - empty_robots_throws: partition(field, {}) → std::invalid_argument
    - zero_width_throws: F2CRobot(0.0) → std::invalid_argument

    LengthBalancedPartition tests:
    - two_robots_balanced: 20x20 field, r1(3.0,1.0), r2(3.0,1.0) → zones.size()==2; both areas > 0; |load[0]-load[1]| / (load[0]+load[1]) <= 0.10 (measured via zone area as proxy, or trust the greedy algorithm — test area > 0)
    - three_robots_balanced: 30x30 field, r1(3.0), r2(3.0), r3(3.0) → zones.size()==3; all areas > 0
    - empty_robots_throws: partition(field, {}) → std::invalid_argument
    - zero_width_throws: F2CRobot(0.0) → std::invalid_argument
  </behavior>

  <action>
Write the following two test files. Do NOT edit tests/CMakeLists.txt — the existing
`file(GLOB_RECURSE TEST_SOURCES ... cpp/*/*.cpp)` pattern auto-discovers these files.

After writing both files, re-run cmake (required for new test files to be discovered by GLOB_RECURSE):
```bash
cmake -S /home/tom/devenv/fields2cover -B /home/tom/devenv/fields2cover/build
```

---

File: `tests/cpp/partition/spatial_rtree_partition_test.cpp`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/partition/spatial_rtree_partition.h"

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

TEST(fields2cover_partition_spatial_rtree, two_robots) {
  F2CCells field = makeRect(20.0, 20.0);  // area = 400.0

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);

  f2c::partition::SpatialRtreePartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  // Total swath coverage should be within 5% of field area
  // (small gaps may exist at field boundary due to swath edge effects)
  double total = zones[0].area() + zones[1].area();
  EXPECT_NEAR(total, field.area(), field.area() * 0.05);
}

TEST(fields2cover_partition_spatial_rtree, three_robots) {
  F2CCells field = makeRect(30.0, 30.0);  // area = 900.0

  F2CRobot r1(3.0), r2(3.0), r3(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  r3.setCruiseVel(1.0);

  f2c::partition::SpatialRtreePartition part;
  auto zones = part.partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  EXPECT_GT(zones[2].area(), 0.0);
}

TEST(fields2cover_partition_spatial_rtree, empty_robots_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  f2c::partition::SpatialRtreePartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}

TEST(fields2cover_partition_spatial_rtree, zero_width_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  F2CRobot r_zero(0.0);
  r_zero.setCruiseVel(1.0);
  f2c::partition::SpatialRtreePartition part;
  EXPECT_THROW(part.partition(field, {r_zero}), std::invalid_argument);
}
```

---

File: `tests/cpp/partition/length_balanced_partition_test.cpp`

```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include <algorithm>
#include <numeric>
#include "fields2cover/types.h"
#include "fields2cover/partition/length_balanced_partition.h"

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

TEST(fields2cover_partition_length_balanced, two_robots) {
  F2CCells field = makeRect(20.0, 20.0);  // area = 400.0

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);

  f2c::partition::LengthBalancedPartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  // Both zones should have similar area (balanced load → balanced coverage area)
  double a0 = zones[0].area();
  double a1 = zones[1].area();
  double total = a0 + a1;
  EXPECT_GT(total, 0.0);
  // Neither zone should dominate — each robot gets a meaningful share
  EXPECT_GT(a0 / total, 0.30);
  EXPECT_GT(a1 / total, 0.30);
}

TEST(fields2cover_partition_length_balanced, three_robots) {
  F2CCells field = makeRect(30.0, 30.0);  // area = 900.0

  F2CRobot r1(3.0), r2(3.0), r3(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  r3.setCruiseVel(1.0);

  f2c::partition::LengthBalancedPartition part;
  auto zones = part.partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  EXPECT_GT(zones[2].area(), 0.0);
}

TEST(fields2cover_partition_length_balanced, empty_robots_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  f2c::partition::LengthBalancedPartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}

TEST(fields2cover_partition_length_balanced, zero_width_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  F2CRobot r_zero(0.0);
  r_zero.setCruiseVel(1.0);
  f2c::partition::LengthBalancedPartition part;
  EXPECT_THROW(part.partition(field, {r_zero}), std::invalid_argument);
}
```
  </action>

  <verify>
    Re-run cmake (GLOB_RECURSE must see new test files), then build and run the new suites:
    ```bash
    cmake -S /home/tom/devenv/fields2cover -B /home/tom/devenv/fields2cover/build && \
    make -C /home/tom/devenv/fields2cover/build unittests -j$(nproc) && \
    /home/tom/devenv/fields2cover/build/tests/unittests \
      --gtest_filter="fields2cover_partition_spatial_rtree*:fields2cover_partition_length_balanced*"
    ```
    Then run the full suite to confirm no regressions:
    ```bash
    /home/tom/devenv/fields2cover/build/tests/unittests
    ```
  </verify>

  <acceptance_criteria>
    - `tests/cpp/partition/spatial_rtree_partition_test.cpp` exists
    - File contains `TEST(fields2cover_partition_spatial_rtree, two_robots)`
    - File contains `TEST(fields2cover_partition_spatial_rtree, three_robots)`
    - File contains `TEST(fields2cover_partition_spatial_rtree, empty_robots_throws)`
    - File contains `TEST(fields2cover_partition_spatial_rtree, zero_width_throws)`
    - `tests/cpp/partition/length_balanced_partition_test.cpp` exists
    - File contains `TEST(fields2cover_partition_length_balanced, two_robots)`
    - File contains `TEST(fields2cover_partition_length_balanced, three_robots)`
    - File contains `TEST(fields2cover_partition_length_balanced, empty_robots_throws)`
    - File contains `TEST(fields2cover_partition_length_balanced, zero_width_throws)`
    - `./build/tests/unittests --gtest_filter="fields2cover_partition_spatial_rtree*:fields2cover_partition_length_balanced*"` exits 0 with all 8 tests PASSED
    - `./build/tests/unittests` exits 0 with at least 313 tests PASSED (305 pre-existing + 8 new), 0 failures
  </acceptance_criteria>

  <done>Eight new GoogleTest cases pass (4 per strategy); full suite remains green at 313+ tests.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → partition() | Caller supplies field geometry and robot specs; no network, no file I/O, no user input in this phase. Phase 26 (gRPC) will be the external trust boundary — it validates before calling these functions. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-25a-01 | Denial of Service | `partition()` — empty robot list | mitigate | Validate `!robots.empty()` at function entry and throw `std::invalid_argument` (implemented in both classes in Task 1) |
| T-25a-02 | Denial of Service | `partition()` — zero or negative coverage width | mitigate | Validate `robots[0].getCovWidth() > 0` at entry and throw `std::invalid_argument` (implemented in both classes in Task 1) |
| T-25a-03 | Denial of Service | `partition()` — multi-cell field (size != 1) causing F2CSwathsByCells confusion | mitigate | Guard `field.size() == 1` at entry and throw `std::invalid_argument`; single-cell assumption documented in header (implemented in both classes in Task 1) |
| T-25a-04 | Denial of Service | `SpatialRtreePartition` — more robots than swaths causes some robots to get no seed | accept | Seed selection uses `(r * swaths.size()) / N` — if swaths.size() < N some robots get duplicate or no seed, resulting in empty zones. Zone area == 0 is valid (no crash). Caller (Phase 26) should validate N <= estimated swath count before calling. |
| T-25a-05 | Tampering | boost::geometry R-tree in public header contaminating downstream TUs | mitigate | All boost::geometry includes are in `spatial_rtree_partition.cpp` only — the header contains only `#include "fields2cover/types.h"`. Enforced by acceptance_criteria grep check. |
| T-25a-06 | Denial of Service | `areaCovered()` returns empty F2CCells for swaths with zero recorded width | accept | `getCovWidth() > 0` guard at entry ensures swath generator uses a positive width; `areaCovered()` will produce non-empty polygons. If generator returns zero-width swaths for other reasons, union of empty F2CCells is an empty zone — no crash, no PII. |
</threat_model>

<verification>
After both tasks complete, run:

```bash
# Targeted filter — must show 8 tests PASSED (4 per strategy)
/home/tom/devenv/fields2cover/build/tests/unittests \
  --gtest_filter="fields2cover_partition_spatial_rtree*:fields2cover_partition_length_balanced*"

# Full suite — must show 313+ tests, 0 failures
/home/tom/devenv/fields2cover/build/tests/unittests
```

Expected terminal output fragments:

Targeted run:
```
[  PASSED  ] 8 tests.
```

Full run:
```
[  PASSED  ] 313 tests.   (or higher)
```
</verification>

<success_criteria>
1. `include/fields2cover/partition/spatial_rtree_partition.h` — SpatialRtreePartition declared in f2c::partition with correct signature; no boost headers in the .h file
2. `src/fields2cover/partition/spatial_rtree_partition.cpp` — R-tree clustering: swath generation via BruteForce, bgi::rtree over midpoints, nearest-centroid assignment, zone geometry via areaCovered() + unionOp()
3. `include/fields2cover/partition/length_balanced_partition.h` — LengthBalancedPartition declared in f2c::partition with correct signature
4. `src/fields2cover/partition/length_balanced_partition.cpp` — FFD greedy assignment: swath generation via BruteForce, sort by length() desc, min_element load tracking, zone geometry via areaCovered() + unionOp()
5. `tests/cpp/partition/spatial_rtree_partition_test.cpp` — 4 tests: two_robots, three_robots, empty_robots_throws, zero_width_throws — all PASSED
6. `tests/cpp/partition/length_balanced_partition_test.cpp` — 4 tests: two_robots, three_robots, empty_robots_throws, zero_width_throws — all PASSED
7. `./build/tests/unittests` exits 0 with 313+ tests PASSED (0 regressions)
8. cmake re-run performed after each new file addition (no stale GLOB_RECURSE results)
9. No CMakeLists.txt edits (GLOB_RECURSE auto-discovers all new files)
10. F2C-03 is satisfied: the library now exposes both SPATIAL_RTREE and LENGTH_BALANCED partition strategies
</success_criteria>

<output>
After completion, create `.planning/phases/25a-c-partition-strategies/25a-01-SUMMARY.md` using the summary template at `@$HOME/.claude/get-shit-done/templates/summary.md`.
</output>
