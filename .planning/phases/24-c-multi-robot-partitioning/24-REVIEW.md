---
phase: 24-c-multi-robot-partitioning
reviewed: 2026-04-29T00:00:00Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - include/fields2cover/partition/multi_robot_partition.h
  - src/fields2cover/partition/multi_robot_partition.cpp
  - tests/cpp/partition/multi_robot_partition_test.cpp
findings:
  critical: 0
  warning: 3
  info: 3
  total: 6
status: issues_found
---

# Phase 24: Code Review Report

**Reviewed:** 2026-04-29
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Reviewed the `MultiRobotPartition` implementation and its accompanying tests. The
core algorithm — axis-aligned strip partitioning weighted by work rate — is correct
for well-formed inputs. The arithmetic is sound, the last-strip boundary clamping
prevents floating-point gaps, and the `F2CCells::intersection` call correctly uses
the `const Cell&` overload.

Three warnings concern robustness gaps for degenerate inputs that are not guarded
against: an empty field geometry triggers GDAL undefined behaviour, a zero-span
(point/line) field silently produces all-empty zones, and the test suite omits the
N=1 single-robot case as well as non-rectangular field shapes.

---

## Warnings

### WR-01: Empty field causes undefined behaviour via uninitialized OGREnvelope

**File:** `src/fields2cover/partition/multi_robot_partition.cpp:36-41`

**Issue:** `getDimMinX()`, `getDimMaxX()`, `getDimMinY()`, and `getDimMaxY()` each
call GDAL's `OGRGeometry::getEnvelope()`. When the geometry is empty, GDAL does not
initialise the `OGREnvelope` fields — they contain whatever was on the stack.
Reading those uninitialized values is undefined behaviour in C++. A caller that
passes a default-constructed or empty `F2CCells` will trigger this silently.

**Fix:** Add an emptiness guard before reading the bounding box:

```cpp
if (field.isEmpty() || field.size() == 0) {
  throw std::invalid_argument(
      "MultiRobotPartition::partition: field must not be empty");
}
```

Place this check immediately after the `robots.empty()` guard (around line 21).

---

### WR-02: Zero-span (degenerate) field returns all-empty zones without error

**File:** `src/fields2cover/partition/multi_robot_partition.cpp:40-41, 56`

**Issue:** When the field's bounding box has `width == 0` and `height == 0` (a point
geometry), or when the longer axis has zero length (a perpendicular line), `span`
becomes 0. Every strip is then zero-width — the intersection with the field yields
an empty `F2CCells`. The function returns `robots.size()` empty zones with status
`ok`, giving the caller no indication that the partition is meaningless.

**Fix:** After computing `width` and `height`, validate that `span` is positive:

```cpp
const double span = cut_along_x ? width : height;
if (span <= 0.0) {
  throw std::invalid_argument(
      "MultiRobotPartition::partition: field bounding box has zero extent "
      "along the partition axis");
}
```

---

### WR-03: Single-robot case not tested; non-rectangular field not tested

**File:** `tests/cpp/partition/multi_robot_partition_test.cpp`

**Issue:** The test suite covers two equal-rate robots, three proportional robots,
empty-robots throw, and zero-work-rate throw. Missing:

1. **N=1**: A single robot should receive the whole field (area equality). Without
   this test, regressions in the last-strip clamping or the `cum_frac` computation
   could go undetected.
2. **Non-convex/concave field**: All tests use `makeRect`, so GEOS intersection
   correctness on real-world shaped fields is untested. A concave polygon (L-shape,
   U-shape) would exercise the intersection code path meaningfully.

**Fix:** Add the following tests:

```cpp
TEST(fields2cover_partition_multi_robot, single_robot_gets_whole_field) {
  F2CCells field = makeRect(10.0, 8.0);
  F2CRobot r(3.0);
  r.setCruiseVel(2.0);
  auto zones = f2c::partition::MultiRobotPartition().partition(field, {r});
  ASSERT_EQ(zones.size(), 1u);
  EXPECT_NEAR(zones[0].area(), field.area(), 1e-3);
}
```

For the concave-field test, construct an L-shaped `F2CCells` (union of two
rectangles) and verify that the summed zone areas equal the total field area.

---

## Info

### IN-01: Redundant `setCruiseVel(1.0)` calls in tests

**File:** `tests/cpp/partition/multi_robot_partition_test.cpp:29-30, 46-48`

**Issue:** `Robot::cruise_speed_` is default-initialised to `1.0` (see
`Robot.h:69`). The explicit `setCruiseVel(1.0)` calls in both tests are no-ops and
add noise that can mislead a reader into thinking the default might differ.

**Fix:** Remove the redundant calls, or change the test to use a non-default velocity
(e.g. `2.0`) to make the test less dependent on internal defaults:

```cpp
r1.setCruiseVel(2.0);
r2.setCruiseVel(2.0);  // still equal → equal zones, but not relying on default
```

---

### IN-02: Negative cruise velocity not rejected by Robot; partition error message is misleading

**File:** `src/fields2cover/partition/multi_robot_partition.cpp:26-30`

**Issue:** `Robot::setCruiseVel()` does not validate its argument, so a caller can
set a negative velocity. If one robot has `getCruiseVel() < 0` while another has
`getCruiseVel() > 0` and the product `getCovWidth() * getCruiseVel()` is still
negative, partition throws correctly. However if all rates are somehow individually
positive yet one was achieved via two negatives (`cov_width < 0`, `cruise_vel < 0`),
the partition proceeds silently with a physically nonsensical work rate. The current
error message also says "positive work rate" without mentioning negative velocity as
a concrete cause.

**Fix (partition):** No change required in the partition guard itself — `<= 0.0`
already catches the common case. Improve the error message to be more concrete:

```cpp
throw std::invalid_argument(
    "MultiRobotPartition::partition: robot[" + std::to_string(i) +
    "] has non-positive work rate (getCovWidth() * getCruiseVel() = " +
    std::to_string(rates[i]) + "); both must be positive");
```

**Fix (Robot):** Add a `>= 0` check in `setCruiseVel`:

```cpp
void Robot::setCruiseVel(double v) {
  if (v <= 0.0) {
    throw std::out_of_range("Robot cruise velocity must be positive.");
  }
  this->cruise_speed_ = v;
}
```

---

### IN-03: `partition()` is not declared `[[nodiscard]]`

**File:** `include/fields2cover/partition/multi_robot_partition.h:32-34`

**Issue:** The return value of `partition()` is the entire point of calling the
function. Marking it `[[nodiscard]]` causes the compiler to warn if a caller ignores
the result, catching accidental call-without-capture mistakes.

**Fix:**

```cpp
[[nodiscard]] std::vector<F2CCells> partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const;
```

---

_Reviewed: 2026-04-29_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
