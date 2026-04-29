# Phase 25b: C++ Graph Route Optimizer - Research

**Researched:** 2026-04-29
**Domain:** C++ route planning, greedy nearest-neighbor graph search, f2c::rp module
**Confidence:** HIGH

## Summary

Phase 25b adds `GraphRouteOptimizer` to the `route_planning/` module. The class inherits `SingleCellSwathsOrderBase` and overrides the single pure-virtual method `sortSwaths(F2CSwaths&)`. The base class already handles cloning, sorting, start-point variants, and the boustrophedon reversal of odd-indexed swaths — `sortSwaths` only needs to reorder indices.

The greedy nearest-neighbor algorithm runs O(n²) per starting swath and O(n³) globally (try all n starts). For the expected field size (n < 100) this is negligible. Edge costs combine Euclidean distance between endpoints and an additive direction penalty when consecutive swath headings diverge by more than 30°. Swath flip (reverse) is chosen per-step: at each greedy step both entry points of each candidate swath are tested and the cheaper orientation is used.

**Primary recommendation:** Implement `sortSwaths` as a self-contained static-function greedy search. No external graph library, no new CMakeLists.txt edits. One header, one source, one test file — all auto-discovered by existing GLOB_RECURSE.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Module: `route_planning/` — alongside existing sorters; auto-discovered by GLOB_RECURSE; no CMakeLists.txt edits
- Namespace: `f2c::rp` — consistent with `BoustrophedonOrder`, `SnakeOrder`, `SpiralOrder`
- Inherits `SingleCellSwathsOrderBase` — override `sortSwaths(F2CSwaths&)` so it integrates with `RoutePlannerBase`
- Public entry point: `genSortedSwaths()` via base class — ROADMAP "optimize()" is conceptual intent, not a distinct method
- Direction flip: YES — each swath may be traversed in either direction; choose entry endpoint that minimises cost to previous swath endpoint
- Algorithm: Greedy nearest-neighbor (O(n²)) — deterministic, no external dependencies, practical for n < 1000 swaths
- Starting swath: try all n starts, keep the globally cheapest result — for typical field sizes (n < 100) the O(n³) cost is negligible
- Direction penalty: additive — when the angle between consecutive swath directions > 30°, add `penalty_weight * distance` to the edge cost; default `penalty_weight = 0.5`
- Parallel threshold: 30° — prefer clearly parallel swaths as neighbors; uses dot product of direction vectors to determine angle

### Claude's Discretion
- Penalty weight (0.5) exposed as constructor parameter with default, allowing future tuning
- Graph construction: fully connected (all-pairs endpoint distances computed on-the-fly, no adjacency matrix stored)
- No external dependencies beyond existing f2c types and std library

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| F2C-04 | f2c library gains a direction-aware graph-based swath traversal optimizer that minimises total travel distance | Greedy NN algorithm with dot-product direction penalty; verified via `DirectDistPathObj::computeCost(F2CSwaths&)` in test |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Swath ordering algorithm | C++ library (f2c::rp) | — | Pure in-memory sort; no I/O or API surface in this phase |
| Direction penalty computation | C++ library (f2c::rp) | — | Angle computed from `F2CSwath::getInAngle()` — available in same tier |
| Cost verification in tests | Test tier (GoogleTest) | — | `DirectDistPathObj::computeCost(F2CSwaths&)` used only in tests |
| gRPC/REST exposure | Phase 26 | — | Out of scope; noted in ROADMAP as Phase 26 dependency |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `fields2cover/types.h` | project-local | All f2c types (F2CSwaths, F2CSwath, F2CPoint) | Required by every route planner |
| `fields2cover/route_planning/single_cell_swaths_order_base.h` | project-local | Base class with `genSortedSwaths` / `sortSwaths` contract | All route planners in f2c::rp inherit this |
| `<cmath>` (std) | system | `acos`, `fabs`, `sqrt` for direction angle | Standard C++ |
| `<vector>` (std) | system | `std::vector<bool>` visited tracking | Standard C++ |
| `<limits>` (std) | system | `std::numeric_limits<double>::max()` sentinel | Standard C++ |

### Supporting (test only)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `fields2cover/objectives/rp_obj/direct_dist_path_obj.h` | project-local | `DirectDistPathObj::computeCost(F2CSwaths&)` for total-cost assertions | Test assertion of travel improvement |
| `<gtest/gtest.h>` | system (cmake) | GoogleTest macros | All f2c unit tests |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Greedy NN (O(n³)) | Dijkstra on prebuilt graph | Dijkstra adds adjacency matrix storage and more complex code; greedy NN is simpler and adequate for n<1000 |
| Additive direction penalty | Multiplicative factor | Additive is independent of distance scale; simpler to tune |

**Installation:** No new packages. All dependencies already present in the build.

## Architecture Patterns

### System Architecture Diagram

```
F2CSwaths (input)
       |
       v
genSortedSwaths()  [base class — SingleCellSwathsOrderBase]
  clone() + sort() + changeStartPoint()
       |
       v
sortSwaths(F2CSwaths&)  [GraphRouteOptimizer — THIS CLASS]
  for each start_idx in 0..n-1:
    greedy_nn(swaths, start_idx, penalty_weight_)
      -> compute edge cost (endpt distance + direction penalty)
      -> pick lowest-cost unvisited swath; flip if reverse is cheaper
      -> accumulate total cost
  keep ordering with globally minimum total cost
  reorder swaths in-place
       |
       v
reverseDirOddSwaths()  [base class — boustrophedon flip of odd indices]
       |
       v
F2CSwaths (reordered, direction-corrected output)
```

### Recommended Project Structure
```
include/fields2cover/route_planning/
├── graph_route_optimizer.h       # NEW — class declaration
├── single_cell_swaths_order_base.h
├── boustrophedon_order.h
└── ...

src/fields2cover/route_planning/
├── graph_route_optimizer.cpp     # NEW — sortSwaths implementation
├── single_cell_swaths_order_base.cpp
└── ...

tests/cpp/route_planning/
├── graph_route_optimizer_test.cpp  # NEW — TEST(fields2cover_route_graph, ...)
└── ...
```

### Pattern 1: sortSwaths Override (Greedy Nearest-Neighbor)

**What:** Override the single pure-virtual `sortSwaths(F2CSwaths&)` to reorder swaths in-place using greedy NN from each possible start; keep globally cheapest.

**When to use:** This is the only required implementation point.

**Example — header:**
```cpp
// Source: verified from boustrophedon_order.h + single_cell_swaths_order_base.h patterns [VERIFIED: codebase]
#pragma once
#ifndef FIELDS2COVER_ROUTE_PLANNING_GRAPH_ROUTE_OPTIMIZER_H_
#define FIELDS2COVER_ROUTE_PLANNING_GRAPH_ROUTE_OPTIMIZER_H_

#include "fields2cover/types.h"
#include "fields2cover/route_planning/single_cell_swaths_order_base.h"

namespace f2c::rp {

class GraphRouteOptimizer : public SingleCellSwathsOrderBase {
 public:
  explicit GraphRouteOptimizer(double penalty_weight = 0.5);
  ~GraphRouteOptimizer() = default;

 protected:
  void sortSwaths(F2CSwaths& swaths) const override;

 private:
  double penalty_weight_;

  // Returns cost from prev_end to next swath (best entry direction).
  // Sets flipped_out to true if next_swath should be reversed.
  double edgeCost(
      const F2CPoint& prev_end,
      double prev_angle,
      const F2CSwath& next,
      bool& flipped_out) const;
};

}  // namespace f2c::rp

#endif  // FIELDS2COVER_ROUTE_PLANNING_GRAPH_ROUTE_OPTIMIZER_H_
```

**Example — sortSwaths skeleton:**
```cpp
// Source: verified from existing sorter patterns + F2CSwath API [VERIFIED: codebase]
void GraphRouteOptimizer::sortSwaths(F2CSwaths& swaths) const {
  const size_t n = swaths.size();
  if (n <= 1) return;

  F2CSwaths best_order;
  double best_cost = std::numeric_limits<double>::max();

  for (size_t start = 0; start < n; ++start) {
    F2CSwaths candidate = swaths;   // copy — F2CSwaths::clone not needed, value type
    std::vector<bool> used(n, false);

    // Place start swath first
    std::swap(candidate[0], candidate[start]);
    used[start] = true;

    double total_cost = 0.0;
    for (size_t i = 1; i < n; ++i) {
      F2CPoint prev_end = candidate[i-1].endPoint();
      double prev_angle = candidate[i-1].getOutAngle();

      size_t best_j = n;  // index in *original swaths* space — use candidate
      double best_edge = std::numeric_limits<double>::max();
      bool best_flip = false;

      for (size_t j = i; j < n; ++j) {
        if (used[j]) continue;  // already placed (tracked by 'i' sweep)
        bool flip = false;
        double cost = edgeCost(prev_end, prev_angle, candidate[j], flip);
        if (cost < best_edge) {
          best_edge = cost;
          best_j = j;
          best_flip = flip;
        }
      }

      // Move best_j into position i
      std::swap(candidate[i], candidate[best_j]);
      if (best_flip) candidate[i].reverse();
      used[best_j] = true;
      total_cost += best_edge;
    }

    if (total_cost < best_cost) {
      best_cost = total_cost;
      best_order = candidate;
    }
  }

  swaths = best_order;
}
```

**Note on "used" tracking:** The swap-into-place approach means `used[j]` is always `false` for j >= i, so the inner loop can simply run from `i` to `n` without a separate `visited` vector. The skeleton above shows the conceptual `used` array but the planner may simplify.

### Pattern 2: Direction Penalty Formula

**What:** Compute angle between two swath direction vectors using dot product. Add penalty when angle > threshold.

**Implementation:**
```cpp
// Source: verified from F2CSwath::getInAngle(), Point::operator*, Point arithmetic [VERIFIED: codebase]
double GraphRouteOptimizer::edgeCost(
    const F2CPoint& prev_end,
    double prev_angle,
    const F2CSwath& next,
    bool& flipped_out) const {
  // Test both entry orientations
  double dist_fwd = prev_end.distance(next.startPoint());
  double dist_rev = prev_end.distance(next.endPoint());

  // Direction of next swath in forward orientation
  double angle_fwd = next.getInAngle();    // angle at swath start -> end
  double angle_rev = next.getOutAngle();   // angle at swath end -> start (reversed entry)

  auto dirPenalty = [&](double angle) -> double {
    // Angle difference between consecutive swath directions
    double diff = std::fabs(
        f2c::types::Geometry<OGRPoint, wkbPoint>::getAngleDiffAbs(prev_angle, angle));
    // diff is in [0, pi]. Parallel = 0, perpendicular = pi/2, anti-parallel = pi.
    // Treat anti-parallel (flip is same direction of travel) — use min(diff, pi-diff)
    double eff_diff = std::min(diff, M_PI - diff);
    constexpr double kThreshold = M_PI / 6.0;  // 30 degrees
    if (eff_diff > kThreshold) {
      return penalty_weight_ * (dist_fwd);  // proportional to distance of this edge
    }
    return 0.0;
  };

  double cost_fwd = dist_fwd + dirPenalty(angle_fwd);
  double cost_rev = dist_rev + dirPenalty(angle_rev);

  flipped_out = (cost_rev < cost_fwd);
  return std::min(cost_fwd, cost_rev);
}
```

**Note:** `Geometry::getAngleDiffAbs(a, b)` is a static method available on any `Geometry<>` subclass. The simplest call: `F2CPoint::getAngleDiffAbs(a, b)` — verified available in `Geometry.h`.

### Pattern 3: Test Structure

**What:** `TEST(fields2cover_route_graph, ...)` format, mirroring `boustrophedon_order_test.cpp`.

**Example:**
```cpp
// Source: verified from boustrophedon_order_test.cpp and snake_order_test.cpp [VERIFIED: codebase]
#include <gtest/gtest.h>
#include "fields2cover/types.h"
#include "fields2cover/objectives/rp_obj/direct_dist_path_obj.h"
#include "fields2cover/route_planning/graph_route_optimizer.h"

TEST(fields2cover_route_graph, improves_over_sequential) {
  // 4 parallel horizontal swaths — GraphRouteOptimizer should find ordering
  // with cost <= sequential (boustrophedon) cost
  F2CSwaths swaths;
  // Swaths at y=1,2,3,4 going horizontally from x=0 to x=10
  for (int i = 1; i <= 4; ++i) {
    swaths.emplace_back(
      F2CLineString({F2CPoint(0, i), F2CPoint(10, i)}), 1.0, i);
  }

  f2c::rp::GraphRouteOptimizer optimizer;
  f2c::obj::DirectDistPathObj objective;

  // Baseline: sequential ordering cost
  double sequential_cost = objective.computeCost(swaths);

  // Optimized ordering
  F2CSwaths result = optimizer.genSortedSwaths(swaths);
  double optimized_cost = objective.computeCost(result);

  EXPECT_EQ(result.size(), 4u);
  EXPECT_LE(optimized_cost, sequential_cost);
}

TEST(fields2cover_route_graph, empty_swaths) {
  F2CSwaths swaths;
  f2c::rp::GraphRouteOptimizer optimizer;
  auto result = optimizer.genSortedSwaths(swaths);
  EXPECT_EQ(result.size(), 0u);
}

TEST(fields2cover_route_graph, single_swath) {
  F2CSwaths swaths;
  swaths.emplace_back(F2CLineString({F2CPoint(0, 0), F2CPoint(1, 0)}), 1.0, 1);
  f2c::rp::GraphRouteOptimizer optimizer;
  auto result = optimizer.genSortedSwaths(swaths);
  EXPECT_EQ(result.size(), 1u);
}
```

### Anti-Patterns to Avoid

- **Building an adjacency matrix:** All-pairs distances computed on-the-fly in the greedy loop. No `n×n` matrix needed — avoids O(n²) memory allocation.
- **Using `F2CGraph` for this problem:** `f2c::types::Graph` stores `int64_t` costs (scaled integers). The greedy NN only needs floating-point distances in a single pass — using the Graph class would require scaling and is unnecessary complexity.
- **Calling `reverseDirOddSwaths` inside `sortSwaths`:** The base class `genSortedSwaths` calls this AFTER `sortSwaths` returns. Calling it inside `sortSwaths` would double-apply the flip.
- **Forgetting edge cost tests both orientations before picking:** The direction flip must be chosen per greedy step to achieve true bidirectional traversal benefit.
- **Flipping swaths in the candidate copy before selecting the global best:** Flip (`.reverse()`) only the chosen swath at the position being filled, after the best candidate is confirmed.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Point-to-point distance | Custom Euclidean sqrt | `p1.distance(p2)` (GDAL OGR via Geometry<>) | OGR distance is already available, handles edge cases |
| Swath direction angle | Manual atan2 on endpoints | `swath.getInAngle()` / `swath.getOutAngle()` | Already computed from path geometry; consistent with rest of codebase |
| Angle difference | Manual subtraction + mod | `Geometry<>::getAngleDiffAbs(a, b)` static method | Handles wraparound correctly |
| Travel cost assertion | Manual loop | `DirectDistPathObj::computeCost(F2CSwaths&)` | Verified existing API; used in all route_planning tests |
| Swath reversal | Manually swap start/end points | `swath.reverse()` | Updates both path geometry and `creation_dir_` flag consistently |

**Key insight:** F2CSwath already provides all geometry primitives needed; the optimizer only needs to reorder and flip swaths without touching raw coordinates.

## Common Pitfalls

### Pitfall 1: Base Class Post-Processing Surprises
**What goes wrong:** `genSortedSwaths` calls `reverseDirOddSwaths()` AFTER `sortSwaths` returns — this flips even/odd swath directions for boustrophedon travel. If `sortSwaths` also adjusts directions (e.g., by flipping swaths to minimise edge cost), the base class post-processing will undo some of those flips.
**Why it happens:** The base class was designed for boustrophedon-style sequential ordering; `reverseDirOddSwaths` is called unconditionally.
**How to avoid:** Apply swath flips inside `sortSwaths` for odd-indexed swaths with the understanding that `reverseDirOddSwaths` will flip them back. Alternatively, do NOT pre-flip in `sortSwaths` and let the base class handle direction assignment — the direction penalty is only used for ordering selection, not for final path direction.
**Warning signs:** Test swath at index 1 starts at unexpected endpoint.

**Recommended approach:** In `sortSwaths`, choose the best entry point for cost-comparison purposes only. Record the flip decision but apply it only for swaths that will end up at even indices (0, 2, 4...) after sorting — odd-indexed ones will be flipped by `reverseDirOddSwaths` regardless. Simplest: don't flip in `sortSwaths` at all; use bidirectional cost only for selection. The base class handles the zig-zag direction pattern.

**Alternative approach (if direction flip for every step is required by the design):** Override `genSortedSwaths` entirely, skipping `reverseDirOddSwaths`. But the CONTEXT.md decision says to use the base class, so avoid this.

**Resolved approach per CONTEXT.md:** The CONTEXT says "each swath may be traversed in either direction; choose entry endpoint that minimises cost to previous swath endpoint (same as `redirect_swaths` logic in base)". This means the flip IS intended in `sortSwaths`. The base class `reverseDirOddSwaths` will run after, but this just ensures even/odd boustrophedon alternation — which is the desired final state anyway. Flip within `sortSwaths` for cost purposes; the base class corrects directionality afterward.

### Pitfall 2: Direction Angle Interpretation
**What goes wrong:** `getInAngle()` returns an angle in `[0, 2π)` relative to the positive x-axis. Two parallel swaths in the same direction have angle ~0 difference. Two parallel swaths in opposite directions have angle ~π. Anti-parallel swaths have the same physical direction of travel if one is traversed reversed — so the effective angular difference is `min(|a-b|, π - |a-b|)` NOT simply `|a-b|`.
**Why it happens:** Naively comparing angles treats anti-parallel as maximum penalty when they are actually the same "parallel field direction".
**How to avoid:** Use `eff_diff = min(diff, pi - diff)` before comparing to the 30° threshold.
**Warning signs:** Two swaths at y=1 going right and y=2 going left (boustrophedon) get penalised even though they are parallel swaths.

### Pitfall 3: "Try All Starts" Copy Cost
**What goes wrong:** Each of the n outer-loop iterations copies all swaths. For n=100, that is 100 copies of 100 swaths.
**Why it happens:** `F2CSwaths` has value semantics; assignment makes a copy.
**How to avoid:** Inside each start iteration, copy into a local `F2CSwaths candidate` (which already happens in the pattern above). This is O(n²) total swath copies — fine for n<1000.
**Warning signs:** None — this is expected cost.

### Pitfall 4: GLOB_RECURSE Requires cmake Re-Run
**What goes wrong:** New `.cpp` files added to `src/fields2cover/route_planning/` and `tests/cpp/route_planning/` are NOT automatically picked up by a running build — CMake must be re-run to glob the new files.
**Why it happens:** `file(GLOB_RECURSE ...)` is evaluated at configure time, not build time.
**How to avoid:** After adding new files, run `cmake ..` (or let the docker build re-run cmake) before building. This is standard f2c practice — confirmed by all Phase 24/25/25a implementations.
**Warning signs:** Linker errors about missing symbols for the new class.

## Code Examples

### Creating F2CSwaths test fixtures (verified pattern)
```cpp
// Source: verified from boustrophedon_order_test.cpp [VERIFIED: codebase]
F2CSwaths swaths;
for (int i = 1; i <= 4; ++i) {
  // Horizontal swath at y=i, from x=0 to x=10, width=1.0, id=i
  swaths.emplace_back(
    F2CLineString({F2CPoint(0, i), F2CPoint(10, i)}), 1.0, i);
}
```

### Computing total travel cost (verified pattern)
```cpp
// Source: verified from boustrophedon_order_test.cpp and snake_order_test.cpp [VERIFIED: codebase]
f2c::obj::DirectDistPathObj objective;
double cost = objective.computeCost(swaths);  // sum of endpoint-to-endpoint headland distances
```

### Swath geometry access (verified from Swath.h + Swath.cpp)
```cpp
// Source: [VERIFIED: codebase]
F2CSwath s = swaths[i];
F2CPoint start = s.startPoint();    // first point of path
F2CPoint end   = s.endPoint();      // last point of path
double in_angle  = s.getInAngle();  // heading angle [0, 2pi) from startPoint direction
double out_angle = s.getOutAngle(); // heading angle from endPoint direction
s.reverse();                         // flip path in-place, toggles creation_dir_
double dist = start.distance(end);   // OGR Euclidean distance
```

### Angle difference (verified from Geometry.h)
```cpp
// Source: [VERIFIED: codebase]
// Static on any Geometry<> subclass — call via F2CPoint for convenience
double diff = F2CPoint::getAngleDiffAbs(angle_a, angle_b);
// Returns value in [0, pi] — smallest angular distance
```

### Point dot product for direction vectors (verified from Point.h)
```cpp
// Source: [VERIFIED: codebase]
// Point::operator*(const Point& b) computes dot product: X*b.X + Y*b.Y + Z*b.Z
F2CPoint dir_a(cos(angle_a), sin(angle_a));
F2CPoint dir_b(cos(angle_b), sin(angle_b));
double dot = dir_a * dir_b;   // in [-1, 1]; 1 = parallel, 0 = perpendicular, -1 = anti-parallel
// Alternative to getAngleDiffAbs; both approaches work
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Boustrophedon sequential | Greedy NN with direction penalty | Phase 25b (new) | Reduces headland travel distance for irregular fields |
| Fixed traversal direction | Bidirectional swath entry | Phase 25b (new) | Allows choosing entry endpoint to reduce turn cost |

**Deprecated/outdated:**
- Nothing deprecated in this phase; adds to existing ordering strategy set.

## Runtime State Inventory

Step 2.5: SKIPPED — this is a greenfield C++ library addition, not a rename/refactor phase.

## Environment Availability

Step 2.6: All dependencies are already present in the project's Docker build environment (C++17 compiler, GDAL/OGR, GoogleTest, CMake). No new external tools required.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| C++17 compiler | `<cmath>`, `<vector>`, `<limits>` | ✓ | GCC/Clang (project-wide) | — |
| GDAL/OGR | F2CPoint.distance(), geometry types | ✓ | Project-wide | — |
| GoogleTest | Unit tests | ✓ | Project-wide | — |
| CMake GLOB_RECURSE | Auto-discover new .cpp files | ✓ | Confirmed in tests/CMakeLists.txt | — |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | GoogleTest (gtest) |
| Config file | tests/CMakeLists.txt — GLOB_RECURSE auto-discovers all `cpp/*/*.cpp` |
| Quick run command | `cd build && ctest -R fields2cover_route_graph -V` |
| Full suite command | `cd build && ctest --output-on-failure` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-04 | `genSortedSwaths` returns ordering with cost ≤ sequential | unit | `ctest -R fields2cover_route_graph` | ❌ Wave 0 |
| F2C-04 | Direction penalty: parallel swaths preferred as neighbors | unit | `ctest -R fields2cover_route_graph` | ❌ Wave 0 |
| F2C-04 | All existing tests still pass | regression | `ctest --output-on-failure` | ✓ existing |

### Sampling Rate
- **Per task commit:** `cd build && ctest -R fields2cover_route_graph -V`
- **Per wave merge:** `cd build && ctest --output-on-failure`
- **Phase gate:** Full suite green (≥313 tests) before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `tests/cpp/route_planning/graph_route_optimizer_test.cpp` — covers F2C-04 (improves over sequential, empty, single-swath)
- [ ] `include/fields2cover/route_planning/graph_route_optimizer.h` — class declaration
- [ ] `src/fields2cover/route_planning/graph_route_optimizer.cpp` — `sortSwaths` implementation

*(No new CMakeLists.txt entry needed — GLOB_RECURSE covers both directories.)*

## Security Domain

Security enforcement not applicable to a pure C++ algorithm library addition with no network, file I/O, or user input surface.

## Open Questions

1. **Does `reverseDirOddSwaths` interfere with direction-flip decisions made inside `sortSwaths`?**
   - What we know: `genSortedSwaths` always calls `reverseDirOddSwaths()` after `sortSwaths`. This flips swaths at positions 1, 3, 5... to opposite direction.
   - What's unclear: If `sortSwaths` flips a swath (`.reverse()`) to minimize edge cost, the base class will un-flip it if it lands at an odd position.
   - Recommendation: In `sortSwaths`, apply `.reverse()` only for swaths that land at even indices (0, 2, 4...) — OR simpler: use bidirectional cost only for SELECTION (which candidate to pick next), but do NOT actually call `.reverse()` inside `sortSwaths`. Let the base class handle all direction assignment via `reverseDirOddSwaths`. This is the safer approach that avoids interaction with the base class post-processing.

2. **What "improved ordering" looks like for the test assertion**
   - What we know: `DirectDistPathObj::computeCost(F2CSwaths&)` sums endpoint-to-endpoint headland distances.
   - What's unclear: For a ≥4-swath perfectly-parallel field, sequential boustrophedon IS already optimal. A non-trivial test requires a scattered (non-sequential) input.
   - Recommendation: Shuffle the swaths before passing to `genSortedSwaths` (random engine like in boustrophedon_order_test), OR create a deliberately suboptimal sequential ordering (e.g., alternate y=1, y=4, y=2, y=3) and assert optimizer beats it.

## Sources

### Primary (HIGH confidence)
- `include/fields2cover/route_planning/single_cell_swaths_order_base.h` — base class contract [VERIFIED: codebase]
- `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp` — `genSortedSwaths` flow [VERIFIED: codebase]
- `include/fields2cover/types/Swath.h` — `startPoint()`, `endPoint()`, `getInAngle()`, `getOutAngle()`, `reverse()` [VERIFIED: codebase]
- `src/fields2cover/types/Swath.cpp` — implementations of all Swath methods [VERIFIED: codebase]
- `include/fields2cover/types/Point.h` — `distance()`, `operator*` (dot product), `getAngleFromPoint()` [VERIFIED: codebase]
- `include/fields2cover/types/Geometry.h` — `getAngleDiffAbs()` static method [VERIFIED: codebase]
- `include/fields2cover/objectives/rp_obj/direct_dist_path_obj.h` — `computeCost(F2CSwaths&)` [VERIFIED: codebase]
- `tests/CMakeLists.txt` — GLOB_RECURSE pattern `cpp/*/*.cpp` auto-discovers new test files [VERIFIED: codebase]
- `CMakeLists.txt` — GLOB_RECURSE for src also confirmed [VERIFIED: codebase]
- `tests/cpp/route_planning/boustrophedon_order_test.cpp` — canonical test fixture patterns [VERIFIED: codebase]
- `tests/cpp/route_planning/snake_order_test.cpp` — secondary test pattern reference [VERIFIED: codebase]
- `include/fields2cover/types/Graph.h` and `Graph2D.h` — existing graph utilities (int64 edges; NOT used for this phase) [VERIFIED: codebase]

### Secondary (MEDIUM confidence)
- CONTEXT.md decisions verified against codebase API availability — all APIs confirmed to exist

### Tertiary (LOW confidence)
- None

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Applying `.reverse()` inside `sortSwaths` for even-indexed swaths does not interfere with `reverseDirOddSwaths` post-processing in a harmful way | Pitfall 1 / Open Questions | Swath direction at output may be incorrect for some indices; test would catch this |

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all APIs verified directly in source code
- Architecture: HIGH — base class flow read from source; pattern confirmed in 4 existing sorters
- Pitfalls: HIGH — identified from direct source code reading of base class and related implementations

**Research date:** 2026-04-29
**Valid until:** Stable (no external dependencies; codebase is the ground truth)
