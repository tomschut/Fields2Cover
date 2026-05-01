# Phase 25a: C++ Partition Strategies - Research

**Researched:** 2026-04-29
**Domain:** C++ / fields2cover library / swath-based field partitioning strategies
**Confidence:** HIGH

---

## Summary

Phase 25a adds two new `f2c::partition` classes alongside `MultiRobotPartition`: a spatial R-tree
clustering partitioner (`SpatialRtreePartition`) and a length-balanced greedy assignment partitioner
(`LengthBalancedPartition`). Both share the exact same public signature as Phase 24's
`MultiRobotPartition::partition(const F2CCells&, const std::vector<F2CRobot>&)` and return
`std::vector<F2CCells>`. They operate at the swath level rather than the raw-geometry level —
the field is first tessellated into swaths, then swaths are assigned to robots, then assigned
swath midpoint buffers are unioned into a zone geometry.

**Key design insight:** The Phase 24 partitioner works on raw geometry (bounding-box strips
intersected with the field). The two new strategies must work on swaths — they decompose the
field into swaths internally (using the existing `f2c::sg::BruteForce` swath generator from
`fields2cover.h`), cluster those swaths by spatial proximity or length balance, and then derive
zone geometries by taking the `F2CCell::buffer` of each swath's `areaCovered()` area and unioning
the result. This is the standard f2c swath-to-coverage pipeline in reverse.

boost::geometry R-tree is available (libboost-all-dev 1.83.0 installed, header
`/usr/include/boost/geometry/index/rtree.hpp` confirmed). boost::geometry is header-only — no
CMakeLists.txt changes are needed for linking. boost::math is already used in the codebase.

**Primary recommendation:** Implement both strategies as standalone classes in
`include/fields2cover/partition/` + `src/fields2cover/partition/`, following the Phase 24 pattern
exactly (pragma once + include guard, BSD-3 license header, `f2c::partition` namespace). No
CMakeLists.txt edits required — GLOB_RECURSE auto-discovers new files.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- All implementation choices are at Claude's discretion — discuss phase was skipped.
- Namespace: `f2c::partition` (consistent with Phase 24)
- Module directories: `include/fields2cover/partition/` and `src/fields2cover/partition/`
- Tests: `tests/cpp/partition/` (auto-discovered by GLOB_RECURSE)
- For SpatialRtreePartition: use boost::geometry R-tree or simple centroid-distance clustering if boost unavailable
- For LengthBalancedPartition: compute swath lengths, sort, greedily assign to robots to balance total load

### Claude's Discretion
All implementation choices — class internals, swath generation parameters, zone geometry construction.

### Deferred Ideas (OUT OF SCOPE)
None.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| F2C-03 | f2c library exposes SPATIAL_RTREE and LENGTH_BALANCED multi-robot partition strategies | Both strategies implemented as `f2c::partition` classes following Phase 24 patterns; swath generation via existing `f2c::sg::BruteForce`; zone geometry via `Swath::areaCovered()` + union |
</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Swath clustering / assignment logic | C++ library (`src/fields2cover/partition/`) | — | Pure geometry computation; no I/O, no gRPC in this phase |
| Public header / API surface | C++ library (`include/fields2cover/partition/`) | — | Contract for Phase 26 gRPC shim |
| Swath generation (internal use) | `f2c::sg::BruteForce` (existing) | — | Existing generator produces `F2CSwaths` from `F2CCells`; no need to reimplement |
| Zone geometry construction | `Swath::areaCovered()` + `Cells::unionOp()` (existing) | — | areaCovered() returns the polygon swept by a swath; unioning per-robot swaths gives zone |
| Unit tests | GoogleTest (`tests/cpp/partition/`) | — | Pattern established by Phase 24 |
| Build discovery | CMakeLists GLOB_RECURSE | — | No CMakeLists edits needed |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| GDAL/OGR | 3.0+ (system) | F2CCells geometry, intersection, union | F2C geometry layer wraps OGRGeometry |
| GEOS (via OGR) | system | unionOp, intersection internals | Already linked via `-lgeos_c` |
| boost::geometry | 1.83.0 (system, header-only) | R-tree spatial index for SpatialRtreePartition | Confirmed at `/usr/include/boost/geometry/index/rtree.hpp` |
| GoogleTest | system | Unit tests | All partition tests use it |
| C++17 | — | Language standard | `CMAKE_CXX_STANDARD 17` in CMakeLists |

### No New Build Dependencies
boost::geometry is header-only — no `find_package(Boost)` or `target_link_libraries` change needed.
boost::math (already used in `include/fields2cover/types/Geometry.h`) proves boost headers are on the
include path. `#include <boost/geometry/index/rtree.hpp>` works today.

---

## Architecture Patterns

### How the Two Strategies Differ from Phase 24

Phase 24 (`MultiRobotPartition`) works on raw geometry:
```
F2CCells (field) → bounding-box strips → intersection with field → std::vector<F2CCells>
```

Phase 25a strategies work on swaths:
```
F2CCells (field) → generate swaths → cluster swaths per robot → union swath area → std::vector<F2CCells>
```

The zone geometry is derived by taking the union of each robot's assigned swaths' coverage areas,
not by intersecting geometric strips with the field boundary.

### Swath Generation Inside Partition

Both new classes need to generate swaths from the input `F2CCells`. The standard approach:
```cpp
// Source: include/fields2cover/swath_generator/swath_generator_base.h [VERIFIED: codebase]
#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/swath_length.h"

f2c::sg::BruteForce sw_gen;
f2c::obj::SwathLength obj;
// Pick swath width from the first robot's coverage width
double op_width = robots[0].getCovWidth();  // or average
F2CSwaths swaths = sw_gen.generateBestSwaths(obj, op_width, field.getCell(0));
```

`generateBestSwaths` returns `F2CSwaths` (a `std::vector<Swath>` wrapper). Each `F2CSwath` has:
- `swath.startPoint()` — `F2CPoint`
- `swath.endPoint()` — `F2CPoint`
- `swath.length()` — `double` (metres along the swath centre line)
- `swath.areaCovered()` — `F2CCells` (polygon swept by the swath at its width)
- `swath.getWidth()` — `double`

### Zone Geometry Construction from Assigned Swaths

Once swaths are assigned to robots, compute zone geometry:
```cpp
// For each robot i, union all assigned swath coverage areas
// Source: include/fields2cover/types/Cells.h [VERIFIED: codebase]
F2CCells zone;
for (const F2CSwath& s : assigned_swaths[i]) {
    F2CCells covered = s.areaCovered();
    zone = zone.unionOp(covered);
}
zones.push_back(zone);
```

### SpatialRtreePartition — Centroid-based R-tree Clustering

The algorithm:
1. Generate swaths from the input field.
2. Compute each swath's midpoint: `mid = (swath.startPoint() + swath.endPoint()) * 0.5`
3. Select N seed swaths (evenly spaced or one per quadrant) — one per robot.
4. Build a boost::geometry R-tree of all swath midpoints.
5. For each remaining swath, query R-tree for nearest already-assigned swath centroid, assign to that
   robot's cluster. Grow clusters round-robin until all swaths are assigned.
6. After full assignment, union each robot's swaths' `areaCovered()` into a zone `F2CCells`.

**Simpler alternative (if R-tree proves complex):** K-means clustering on swath midpoints (2D,
fixed N iterations). This is also header-only and requires no additional dependencies. The key
invariant (spatially contiguous zones) is satisfied by either approach. Research recommends the
R-tree approach since the CONTEXT.md names it explicitly and boost::geometry is confirmed available.

#### boost::geometry R-tree Minimal Pattern
```cpp
// Source: /usr/include/boost/geometry/index/rtree.hpp [VERIFIED: file exists]
#include <boost/geometry.hpp>
#include <boost/geometry/index/rtree.hpp>

namespace bg  = boost::geometry;
namespace bgi = boost::geometry::index;

// 2D point type for the R-tree
using BgPoint = bg::model::point<double, 2, bg::cs::cartesian>;
using Value   = std::pair<BgPoint, std::size_t>;  // (midpoint, swath_index)

bgi::rtree<Value, bgi::quadratic<16>> rtree;

// Insert all swath midpoints
for (std::size_t i = 0; i < swaths.size(); ++i) {
    auto mid = (swaths[i].startPoint() + swaths[i].endPoint()) * 0.5;
    rtree.insert({BgPoint(mid.getX(), mid.getY()), i});
}

// Nearest-neighbour query for a target point
BgPoint query_pt(target.getX(), target.getY());
std::vector<Value> result;
rtree.query(bgi::nearest(query_pt, 1), std::back_inserter(result));
```

### LengthBalancedPartition — Greedy Length-Balanced Assignment

The algorithm (no external dependencies beyond `<algorithm>` and `<numeric>`):
1. Generate swaths from the input field.
2. Sort swaths by `length()` descending (largest first — improves balance by first-fit decreasing).
3. Maintain a `std::vector<double> load(N, 0.0)` — total swath length assigned to each robot.
4. For each swath, assign to the robot with the minimum current load (standard greedy bin-packing).
5. After full assignment, union each robot's swaths into zone geometry.

**Balance guarantee:** First-fit decreasing (FFD) on swath lengths achieves near-optimal balance.
With uniform swath lengths (typical for a constant-heading coverage pattern), the maximum imbalance
is at most one swath's length across robots. For a 10x10 field with 1 m swath width, each swath is
~10 m; with 10 swaths total and 2 robots, the worst imbalance is 10 m out of 50 m = 20%. To achieve
the stated ≤10% balance requirement, the partition should also ensure no single robot gets all
adjacent swaths from one end (spatial clustering may break balance if field is narrow). A post-pass
can swap boundary swaths to improve spatial locality, but FFD alone satisfies the length-balance
criterion.

```cpp
// Source: <algorithm> / <numeric> standard library [VERIFIED: C++17]
// Sort swaths by length descending
std::vector<std::size_t> order(swaths.size());
std::iota(order.begin(), order.end(), 0);
std::sort(order.begin(), order.end(), [&](std::size_t a, std::size_t b) {
    return swaths[a].length() > swaths[b].length();
});

// Greedy min-load assignment
std::vector<double> load(N, 0.0);
std::vector<std::vector<std::size_t>> assigned(N);
for (std::size_t idx : order) {
    auto it = std::min_element(load.begin(), load.end());
    std::size_t robot_idx = std::distance(load.begin(), it);
    assigned[robot_idx].push_back(idx);
    load[robot_idx] += swaths[idx].length();
}
```

### Swath Generation: Which Width to Use

Both strategies need to pick a swath width for the internal generator. Options:
- Use `robots[0].getCovWidth()` (first robot's width) — simple, consistent.
- Use `max` across all robots — fewer swaths, faster.
- Use `min` across all robots — more swaths, finer granularity, better balance.

**Recommended:** Use `robots[0].getCovWidth()`. Rationale: the same field would be given to all
robots in a typical multi-robot scenario where robots have the same model. The partition divides
swaths between robots, not the geometry directly. If all robots have the same coverage width (common
case), this is exact. If robots have different widths, the caller should pre-generate swaths
externally and the internal swath generation is approximate — this is acceptable since the phase
goal is zone-level assignment, not sub-swath accuracy. Document this assumption in the class header.

### Recommended Project Structure

No changes to directory structure — existing partition layout is already correct:
```
include/fields2cover/partition/
├── multi_robot_partition.h           # Phase 24 (exists)
├── spatial_rtree_partition.h         # New (this phase)
└── length_balanced_partition.h       # New (this phase)

src/fields2cover/partition/
├── multi_robot_partition.cpp         # Phase 24 (exists)
├── spatial_rtree_partition.cpp       # New (this phase)
└── length_balanced_partition.cpp     # New (this phase)

tests/cpp/partition/
├── multi_robot_partition_test.cpp    # Phase 24 (exists)
├── spatial_rtree_partition_test.cpp  # New (this phase)
└── length_balanced_partition_test.cpp # New (this phase)
```

### Anti-Patterns to Avoid

- **Using `F2CCells::splitByLine` for swath-based strategies:** The Phase 24 approach splits
  geometry directly. The new strategies split at the swath level — `splitByLine` is not the right
  API here.
- **Assuming `F2CCells` has a `flatten()` method for swaths:** `F2CCells` stores geometry polygons,
  not swaths. `F2CSwathsByCells` is for per-cell swath collections. For a single-cell field, use
  `f2c::sg::BruteForce::generateBestSwaths(obj, width, field.getCell(0))` returning `F2CSwaths`.
- **Not handling the empty-union case:** If a robot gets zero swaths assigned (possible when
  N > total swaths), push an empty `F2CCells{}` as its zone rather than crashing.
- **boost::geometry point type mismatch:** The R-tree uses `bg::model::point<double, 2, cartesian>`,
  not `F2CPoint`. Convert: `BgPoint(p.getX(), p.getY())`. Do not use `OGRPoint` directly with
  boost::geometry index.
- **Including `<boost/geometry.hpp>` umbrella in headers:** It pulls in substantial compile time.
  Prefer including only `<boost/geometry/index/rtree.hpp>` and the minimal model headers in the
  `.cpp` file, not in the `.h` file. The header declares only the partition method — R-tree details
  are implementation-specific and belong in `.cpp`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Swath generation | Custom line-sweep | `f2c::sg::BruteForce::generateBestSwaths` | Already handles angle optimization, overlap control |
| Zone union geometry | Manual polygon merge | `F2CCells::unionOp()` | GEOS handles self-touching polygons, holes, etc. |
| Spatial nearest-neighbour | Custom distance loop O(n²) | `boost::geometry::index::rtree` | R-tree is O(log n) per query; already available |
| Swath coverage polygon | Buffer/sweep manually | `F2CSwath::areaCovered()` | Already returns correct `F2CCells` for a swath |

---

## Common Pitfalls

### Pitfall 1: F2CCells vs F2CCell for swath generation
**What goes wrong:** `BruteForce::generateBestSwaths` has two overloads — one for `F2CCell` (single
polygon) and one for `F2CCells` (multi-polygon, returns `F2CSwathsByCells`). Using the wrong
overload returns a per-cell collection instead of a flat `F2CSwaths`.
**Why it happens:** A field with a single polygon is stored as `F2CCells` with one `F2CCell` inside.
**How to avoid:** For a single-cell field (common case), call
`sw_gen.generateBestSwaths(obj, width, field.getCell(0))` to get `F2CSwaths`. For multi-cell
fields (decomposed fields), iterate cells.
**Warning signs:** Compiler error matching `F2CSwathsByCells` where `F2CSwaths` expected.

### Pitfall 2: Empty zone when a robot gets no swaths
**What goes wrong:** If `N > swaths.size()`, some robots get no swaths assigned. Calling
`unionOp` on zero swaths may return an empty/default-constructed `F2CCells`. Callers in Phase 26
may not handle empty zones gracefully.
**How to avoid:** Validate that `robots.size() <= swaths.size()` before assignment, or throw
`std::invalid_argument` when N exceeds available swaths. Document in the header.
**Warning signs:** Zone with `area() == 0.0` in the result vector.

### Pitfall 3: boost::geometry R-tree header in public header
**What goes wrong:** If `#include <boost/geometry/index/rtree.hpp>` appears in the public `.h`
file, all downstream translation units (including the gRPC shim in Phase 26) pay its compile-time
cost and must have boost on their include path.
**How to avoid:** Put all boost::geometry includes in `spatial_rtree_partition.cpp` only. The
header only needs `#include "fields2cover/types.h"` and the method declaration.
**Warning signs:** Compile errors in `f2c-grpc/` or Python swig after including the partition header.

### Pitfall 4: cmake re-run needed before new test is discovered
**What goes wrong:** After adding a new `*_test.cpp` file, running `make -C build unittests`
without re-running cmake may not compile the new test (stale GLOB_RECURSE result).
**How to avoid:** Always run `cmake -S . -B build` (or `cmake --build build`) after adding new
source or test files. CMake 3.12+ detects glob changes and re-runs automatically when using
`CONFIGURE_DEPENDS`, but the project does not set that flag — manual re-run is safer.
**Warning signs:** New test cases are not present in `./build/tests/unittests --gtest_list_tests`.

### Pitfall 5: `Swath::areaCovered()` returns zero area for swaths with zero width
**What goes wrong:** `F2CSwath::areaCovered()` computes area using the swath's stored width. If
swaths were generated with a zero coverage width, `areaCovered()` returns an empty `F2CCells`.
**How to avoid:** Validate `robots[0].getCovWidth() > 0` at the start of `partition()` (mirror
the check in Phase 24's MultiRobotPartition). Throw `std::invalid_argument` if not.

---

## Code Examples

### Phase 24 header pattern to replicate
```cpp
// Source: include/fields2cover/partition/multi_robot_partition.h [VERIFIED: codebase]
#pragma once
#ifndef FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
#define FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

class SpatialRtreePartition {
 public:
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition
#endif  // FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
```

### Phase 24 source pattern to replicate
```cpp
// Source: src/fields2cover/partition/multi_robot_partition.cpp [VERIFIED: codebase]
#include "fields2cover/partition/spatial_rtree_partition.h"
// ... additional includes needed for implementation (boost, swath gen) ...

namespace f2c::partition {

std::vector<F2CCells> SpatialRtreePartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: robots must not be empty");
  }
  if (robots[0].getCovWidth() <= 0.0) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: first robot must have positive coverage width");
  }
  // ... implementation
}
}  // namespace f2c::partition
```

### Test pattern to replicate
```cpp
// Source: tests/cpp/partition/multi_robot_partition_test.cpp [VERIFIED: codebase]
#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/partition/spatial_rtree_partition.h"

namespace {
F2CCells makeRect(double w, double h) {
  F2CLinearRing ring{
    F2CPoint(0, 0), F2CPoint(w, 0),
    F2CPoint(w, h), F2CPoint(0, h),
    F2CPoint(0, 0)};
  return F2CCells{F2CCell{ring}};
}
}  // namespace

TEST(fields2cover_partition_spatial_rtree, two_robots) {
  F2CCells field = makeRect(20.0, 20.0);
  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  f2c::partition::SpatialRtreePartition part;
  auto zones = part.partition(field, {r1, r2});
  ASSERT_EQ(zones.size(), 2u);
  EXPECT_NEAR(zones[0].area() + zones[1].area(), field.area(), field.area() * 0.05);
}
```

### Total area tolerance note
Unlike Phase 24 (strip intersection always covers 100% of area), the new strategies cover the field
via swath `areaCovered()` unions. Small gaps may exist between swaths if the swath generator leaves
thin unswathed strips at field boundaries. Tests should allow ~5% area tolerance (swath boundary
effects), not the 1e-3 absolute used by Phase 24.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | GoogleTest (system package) |
| Config file | `tests/CMakeLists.txt` + `tests/unittests.cpp` |
| Quick run command | `make -C build unittests -j$(nproc) && ./build/tests/unittests --gtest_filter="fields2cover_partition_spatial*:fields2cover_partition_length*"` |
| Full suite command | `./build/tests/unittests` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-03 | SpatialRtreePartition returns 2 zones on 2-robot call | unit | `--gtest_filter="fields2cover_partition_spatial_rtree*"` | No — Wave 0 |
| F2C-03 | SpatialRtreePartition returns 3 zones on 3-robot call | unit | `--gtest_filter="fields2cover_partition_spatial_rtree*"` | No — Wave 0 |
| F2C-03 | LengthBalancedPartition returns 2 zones, balanced within 10% | unit | `--gtest_filter="fields2cover_partition_length_balanced*"` | No — Wave 0 |
| F2C-03 | LengthBalancedPartition returns 3 zones, balanced within 10% | unit | `--gtest_filter="fields2cover_partition_length_balanced*"` | No — Wave 0 |
| F2C-03 | All existing tests still pass | regression | `./build/tests/unittests` | Yes (305 tests) |

### Sampling Rate
- **Per task commit:** `make -C build unittests -j$(nproc) && ./build/tests/unittests --gtest_filter="fields2cover_partition*"`
- **Per wave merge:** `./build/tests/unittests`
- **Phase gate:** Full suite green (305 + new tests) before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `tests/cpp/partition/spatial_rtree_partition_test.cpp` — covers F2C-03 (SpatialRtree)
- [ ] `tests/cpp/partition/length_balanced_partition_test.cpp` — covers F2C-03 (LengthBalanced)
- [ ] cmake re-run required after adding new files — no framework install needed

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| boost::geometry | SpatialRtreePartition R-tree | Yes | 1.83.0 | Centroid-distance O(n²) loop |
| boost::math | Already used in Geometry.h | Yes | 1.83.0 | — |
| GDAL/OGR | All geometry operations | Yes | system | — |
| GEOS | `unionOp`, `intersection` | Yes | system | — |
| GoogleTest | Unit tests | Yes | system | — |
| `f2c::sg::BruteForce` | Swath generation inside partition | Yes (library) | — | — |

**Missing dependencies with no fallback:** None.

**boost::geometry linkage:** Header-only. No `find_package(Boost COMPONENTS geometry)` needed.
Current CMakeLists already gives the compiler access to boost headers via the system include path.

---

## Open Questions

1. **Swath generation for multi-cell fields**
   - What we know: `BruteForce::generateBestSwaths` has a `F2CCells` overload returning
     `F2CSwathsByCells` (per-cell). Flattening via `swaths_by_cells.flatten()` gives all swaths.
   - What's unclear: Whether the partition strategies should handle multi-cell fields (disjoint
     field geometries) or just single-cell fields (the common case from the gRPC pipeline).
   - Recommendation: Support single-cell fields in the initial implementation (call
     `field.getCell(0)`). Add a guard that throws if `field.size() != 1`. Phase 26 can relax later.

2. **Spatial contiguity guarantee for SpatialRtreePartition**
   - What we know: K-nearest clustering on swath midpoints produces spatially compact clusters
     but does not formally guarantee convex or simply-connected zones.
   - What's unclear: Whether the Phase 26 gRPC client or the frontend needs formally connected zones.
   - Recommendation: For now, "spatially compact" (low intra-cluster centroid variance) satisfies
     success criterion 1 ("geographically compact, not scattered strips"). Document the limitation.

3. **Which robot's `getCovWidth()` to use for internal swath generation**
   - Recommendation stated above: use `robots[0].getCovWidth()`. [ASSUMED] — no explicit project
     spec for heterogeneous-width fleets. If robots have different widths, swath assignment is
     approximate.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `robots[0].getCovWidth()` is the right swath width for internal swath generation in both strategies | Architecture Patterns | If robots have very different widths, generated swaths may not match any robot's actual coverage pattern; zone areas will be approximate |
| A2 | Single-cell input is the common case; multi-cell fields can be deferred | Open Questions | If gRPC callers send decomposed multi-cell fields, both new classes will throw; Phase 26 must pre-process |
| A3 | 5% total area tolerance is sufficient for swath-union zone geometry | Code Examples / Tests | If areaCovered() gaps are larger than 5%, test assertions using relative tolerance may still fail on non-rectangular fields |

---

## Sources

### Primary (HIGH confidence)
- `include/fields2cover/partition/multi_robot_partition.h` — Phase 24 pattern verified in codebase
- `src/fields2cover/partition/multi_robot_partition.cpp` — Phase 24 implementation verified
- `tests/cpp/partition/multi_robot_partition_test.cpp` — test pattern verified
- `include/fields2cover/types/Swath.h` — `length()`, `startPoint()`, `endPoint()`, `areaCovered()` verified
- `include/fields2cover/types/Swaths.h` — `F2CSwaths` container verified
- `include/fields2cover/types/Cells.h` — `unionOp()`, `intersection()`, `getCell()` verified
- `include/fields2cover/swath_generator/swath_generator_base.h` — `generateBestSwaths` signature verified
- `include/fields2cover/types/Geometry.h` — `distance()` template verified
- `/usr/include/boost/geometry/index/rtree.hpp` — confirmed present (libboost-all-dev 1.83.0)
- `CMakeLists.txt` — GLOB_RECURSE auto-discovery confirmed; boost not explicitly linked (header-only confirmed)
- `tests/CMakeLists.txt` — `cpp/*/*.cpp` glob pattern confirmed

### Secondary (MEDIUM confidence)
- boost::geometry R-tree quadratic packing parameter `<16>` — standard default for small datasets, from boost::geometry documentation conventions
