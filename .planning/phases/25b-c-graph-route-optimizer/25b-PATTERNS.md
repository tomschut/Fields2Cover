# Phase 25b: C++ Graph Route Optimizer - Pattern Map

**Mapped:** 2026-04-29
**Files analyzed:** 3
**Analogs found:** 3 / 3

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `include/fields2cover/route_planning/graph_route_optimizer.h` | class header | request-response | `include/fields2cover/route_planning/spiral_order.h` | exact — same base class, constructor with param, private member, `sortSwaths` override |
| `src/fields2cover/route_planning/graph_route_optimizer.cpp` | service / algorithm | transform | `src/fields2cover/route_planning/snake_order.cpp` + `spiral_order.cpp` | role-match — same base class `sortSwaths` override, in-place reorder of `F2CSwaths&` |
| `tests/cpp/route_planning/graph_route_optimizer_test.cpp` | test | CRUD | `tests/cpp/route_planning/boustrophedon_order_test.cpp` | exact — same fixture pattern, `DirectDistPathObj`, `genSortedSwaths` call |

---

## Pattern Assignments

### `include/fields2cover/route_planning/graph_route_optimizer.h` (class header)

**Analog:** `include/fields2cover/route_planning/spiral_order.h`

**License + pragma + include guard pattern** (spiral_order.h lines 1-4, boustrophedon_order.h lines 1-9):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_ROUTE_PLANNING_GRAPH_ROUTE_OPTIMIZER_H_
#define FIELDS2COVER_ROUTE_PLANNING_GRAPH_ROUTE_OPTIMIZER_H_
```

**Includes pattern** (boustrophedon_order.h lines 11-12):
```cpp
#include "fields2cover/types.h"
#include "fields2cover/route_planning/single_cell_swaths_order_base.h"
```

**Class declaration with constructor param + private member** (spiral_order.h lines 8-26):
```cpp
namespace f2c::rp {

class SpiralOrder : public SingleCellSwathsOrderBase {
 public:
  explicit SpiralOrder(size_t sp_size = 2);
  ~SpiralOrder();

 protected:
  void sortSwaths(F2CSwaths& swaths) const override;

 private:
  size_t spiral_size;
  void spiral(F2CSwaths& swaths, size_t offset, size_t size) const;
};

}  // namespace f2c::rp

#endif  // FIELDS2COVER_ROUTE_PLANNING_SPIRAL_ORDER_H_
```

**Copy for GraphRouteOptimizer** — adapt class name, param type (`double penalty_weight = 0.5`), private member (`double penalty_weight_`), and add private helper `edgeCost(...)`. Keep `~GraphRouteOptimizer() = default;` (no destructor body needed unlike SpiralOrder). The `sortSwaths` override goes under `protected:`.

---

### `src/fields2cover/route_planning/graph_route_optimizer.cpp` (algorithm implementation)

**Analog:** `src/fields2cover/route_planning/spiral_order.cpp` (constructor + sortSwaths) and `src/fields2cover/route_planning/snake_order.cpp` (in-place swaths reorder)

**License + single include pattern** (boustrophedon_order.cpp lines 1-8, spiral_order.cpp lines 1-2):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#include "fields2cover/route_planning/graph_route_optimizer.h"
```

**Constructor body pattern** (spiral_order.cpp lines 6-8):
```cpp
SpiralOrder::SpiralOrder(size_t sp_size) {
  setSpiralSize(static_cast<size_t>(std::max(2, static_cast<int>(sp_size))));
}
```
For GraphRouteOptimizer: constructor stores `penalty_weight_` directly — no clamping needed, simpler than SpiralOrder.

**sortSwaths signature** (boustrophedon_order.cpp line 11, snake_order.cpp line 11):
```cpp
void BoustrophedonOrder::sortSwaths(F2CSwaths& swaths) const {
```
```cpp
void SnakeOrder::sortSwaths(F2CSwaths& swaths) const {
```
Both take `F2CSwaths&` by non-const reference and modify in-place. Copy this signature exactly.

**In-place reorder via swap pattern** (snake_order.cpp lines 12-19 — uses `std::rotate`; for greedy NN use `std::swap`):
```cpp
// snake_order.cpp — shows in-place manipulation of swaths vector
for (; i < (swaths.size() - 1) / 2 + 1; ++i) {
  std::rotate(swaths.begin() + i, swaths.begin() + i + 1, swaths.end());
}
```
For GraphRouteOptimizer's greedy loop use `std::swap(candidate[i], candidate[best_j])` — same STL vector manipulation idiom.

**Swath geometry API to use** (verified from `include/fields2cover/types/Swath.h` and `include/fields2cover/types/Point.h`):
```cpp
F2CPoint start = swaths[i].startPoint();   // first point
F2CPoint end   = swaths[i].endPoint();     // last point
double in_angle  = swaths[i].getInAngle(); // heading [0,2pi) at start->end
double out_angle = swaths[i].getOutAngle();// heading at end->start (reversed entry)
swaths[i].reverse();                       // flip path in-place
double dist = p1.distance(p2);             // OGR Euclidean distance
double diff = F2CPoint::getAngleDiffAbs(a, b); // returns [0, pi], handles wraparound
```

**Namespace wrapper pattern** (spiral_order.cpp lines 3, 36):
```cpp
namespace f2c::rp {

// ... all class method definitions ...

}  // namespace f2c::rp
```

**std includes needed** — add after the project header:
```cpp
#include <cmath>
#include <limits>
#include <vector>
```

---

### `tests/cpp/route_planning/graph_route_optimizer_test.cpp` (GoogleTest file)

**Analog:** `tests/cpp/route_planning/boustrophedon_order_test.cpp` (primary) + `tests/cpp/route_planning/snake_order_test.cpp` (secondary)

**File header + includes pattern** (boustrophedon_order_test.cpp lines 1-11):
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <random>
#include "fields2cover/types.h"
#include "fields2cover/objectives/rp_obj/direct_dist_path_obj.h"
#include "fields2cover/route_planning/boustrophedon_order.h"
```
For GraphRouteOptimizer: replace last include with `"fields2cover/route_planning/graph_route_optimizer.h"`.

**Swath fixture construction pattern** (boustrophedon_order_test.cpp lines 14-19):
```cpp
const int n = 5;
F2CSwaths swaths;
for (int i = 1; i < n; ++i) {
  swaths.emplace_back(F2CLineString({F2CPoint(i, 0), F2CPoint(i, 1)}), i, i);
}
```
Constructor: `F2CSwath(F2CLineString path, double width, int id)`. Coordinates matter for directional tests.

**Shuffle for non-trivial input** (boustrophedon_order_test.cpp lines 21-22):
```cpp
auto rng = std::default_random_engine {};
std::shuffle(swaths.begin(), swaths.end(), rng);
```

**genSortedSwaths call + cost assertion pattern** (boustrophedon_order_test.cpp lines 24-27):
```cpp
f2c::rp::BoustrophedonOrder swath_sorter;
f2c::obj::DirectDistPathObj objective;

swaths = swath_sorter.genSortedSwaths(swaths);
EXPECT_EQ(swaths.size(), n - 1);
EXPECT_EQ(objective.computeCost(swaths), 2*(n-1)-1);
```

**Empty swaths edge-case test** (boustrophedon_order_test.cpp lines 53-59):
```cpp
TEST(fields2cover_route_boustrophedon, genSortedSwaths_empty_swaths) {
  F2CSwaths swaths;
  f2c::rp::BoustrophedonOrder swath_sorter;
  auto new_swaths = swath_sorter.genSortedSwaths(swaths);
  EXPECT_EQ(swaths.size(), 0);
  EXPECT_EQ(new_swaths.size(), 0);
}
```

**Test name suite prefix convention:**
- `fields2cover_route_boustrophedon` for BoustrophedonOrder
- `fields2cover_route_snake` for SnakeOrder
- `fields2cover_route_spiral` for SpiralOrder
- Use `fields2cover_route_graph` for GraphRouteOptimizer — consistent snake_case suffix.

**Constructor-with-param test pattern** (spiral_order_test.cpp lines 5-13):
```cpp
TEST(fields2cover_route_spiral, genSortedSwaths_even) {
  // ...
  f2c::rp::SpiralOrder swath_sorter(size);   // explicit param
  swaths = swath_sorter.genSortedSwaths(swaths);
  // ...
}
```
Copy for GraphRouteOptimizer default and custom `penalty_weight` variants.

---

## Shared Patterns

### License Header
**Source:** `include/fields2cover/route_planning/boustrophedon_order.h` lines 1-5, `src/fields2cover/route_planning/boustrophedon_order.cpp` lines 1-5
**Apply to:** All three new files (header, implementation, test)
```cpp
//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================
```

### Namespace Wrapper
**Source:** `src/fields2cover/route_planning/spiral_order.cpp` lines 3 and 36, `include/fields2cover/route_planning/boustrophedon_order.h` lines 14 and 22
**Apply to:** Header and implementation files
```cpp
namespace f2c::rp {
// ...
}  // namespace f2c::rp
```

### genSortedSwaths Call Convention
**Source:** `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp` lines 11-22
**Apply to:** Tests only — callers use `genSortedSwaths()`, never `sortSwaths()` directly
```cpp
F2CSwaths SingleCellSwathsOrderBase::genSortedSwaths(
    const F2CSwaths& swaths, uint32_t variant) const {
  F2CSwaths new_swaths = swaths.clone();
  if (new_swaths.size() > 1) {
    new_swaths.sort();
    this->changeStartPoint(new_swaths, variant);
    this->sortSwaths(new_swaths);
    new_swaths.reverseDirOddSwaths();   // <-- runs AFTER sortSwaths
  }
  return new_swaths;
}
```
**Critical:** `reverseDirOddSwaths()` runs unconditionally after `sortSwaths`. Any `.reverse()` calls inside `sortSwaths` for odd-indexed swaths will be undone. Use `.reverse()` only for the purpose of finding the best entry endpoint during cost evaluation in `sortSwaths`; whether to actually flip in `sortSwaths` or rely on the base class post-processing is the key design decision for the implementation wave.

### DirectDistPathObj Cost Assertion
**Source:** `tests/cpp/route_planning/boustrophedon_order_test.cpp` lines 10, 32, `tests/cpp/route_planning/snake_order_test.cpp` lines 9, 29
**Apply to:** `graph_route_optimizer_test.cpp`
```cpp
#include "fields2cover/objectives/rp_obj/direct_dist_path_obj.h"
// ...
f2c::obj::DirectDistPathObj objective;
double cost = objective.computeCost(swaths);
```

---

## No Analog Found

None — all three files have exact or role-match analogs in the codebase.

---

## Metadata

**Analog search scope:** `include/fields2cover/route_planning/`, `src/fields2cover/route_planning/`, `tests/cpp/route_planning/`
**Files scanned:** 12 (6 headers, 6 source/test files read directly)
**Pattern extraction date:** 2026-04-29
