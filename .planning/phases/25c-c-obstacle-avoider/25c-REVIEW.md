---
phase: 25c-c-obstacle-avoider
reviewed: 2026-04-29T00:00:00Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - include/fields2cover/obstacle/obstacle_avoider.h
  - src/fields2cover/obstacle/obstacle_avoider.cpp
  - tests/cpp/obstacle/obstacle_avoider_test.cpp
findings:
  critical: 0
  warning: 3
  info: 3
  total: 6
status: issues_found
---

# Phase 25c: Code Review Report

**Reviewed:** 2026-04-29
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Reviewed the new `ObstacleAvoider` class (`avoid` method), its header, and the
accompanying test suite. The implementation is clean and well-commented. All
public API types resolve correctly against the existing F2C type layer — no raw
GEOS/OGR pointers, no hand-rolled geometry math, consistent with the codebase
conventions observed in `Cell.h`, `Cells.h`, `Swath.h`, and `Swaths.h`.

Three warnings concern logic correctness and a silent failure mode; three info
items concern minor style or missing coverage.

---

## Warnings

### WR-01: `F2CCell::buffer` returns `F2CCell`, not `F2CCells` — type mismatch silently accepted by constructor

**File:** `src/fields2cover/obstacle/obstacle_avoider.cpp:23`

**Issue:** `F2CCell::buffer(const Cell&, double)` (declared `Cell.h:54`) returns
`F2CCell`. The result is stored as `F2CCell inflated`, which is correct.
However, at line 46 the call `bbox_cells.difference(inflated)` uses the
overload `Cells::difference(const Cell& c)` (`Cells.h:55`). This is correct.

The concern is subtler: `F2CCell::buffer` with a non-zero `safety_margin` will
wrap the result in an OGR polygon whose ring representation may be a
**circle-approximated** polygon (OGR default 30-segment approximation). No
validation that the returned geometry is non-empty before passing it to
`difference`. If `obstacle` is an empty/degenerate `F2CCell`, `buffer` will
return an empty polygon, `difference` will return `bbox_cells` unchanged, and
the obstacle is silently ignored rather than flagged.

**Fix:** Guard against a degenerate obstacle and add an assertion or early-return:

```cpp
if (obstacle.isEmpty()) {
  return swaths;  // nothing to clip against
}
F2CCell inflated = F2CCell::buffer(obstacle, safety_margin);
```

---

### WR-02: Bounding-box approach breaks for vertical or near-vertical paths with zero width

**File:** `src/fields2cover/obstacle/obstacle_avoider.cpp:36-45`

**Issue:** When a `F2CLineString` is a single point or has identical X or Y
extents (e.g. a perfectly vertical line has `getDimMinX() == getDimMaxX()`),
the computed bounding box degenerates:

```
x0 = x1 - padding  (correct)
x1 = x0 + padding  (correct)  // wait — these are the same value
```

For a vertical segment at x=5, `getDimMinX()` and `getDimMaxX()` both return 5.
The bounding box still has non-zero width because `padding = safety_margin + 1.0`
is added to both sides, so the box becomes `[5 - padding, 5 + padding]`. This is
actually correct.

However if the path has **zero length** (a single point stored as a
`LineString`), `getDimMinX() == getDimMaxX()` AND `getDimMinY() == getDimMaxY()`,
making the bounding box a square of side `2 * padding` centred on that point.
Subsequent `getLinesInside` on a point-geometry `F2CLineString` will return an
empty `F2CMultiLineString` — the point is silently discarded. If the caller
intends to represent a waiting point as a zero-length swath this would be
dropped without warning.

More practically: a path produced by upstream code that stores only one point
could reach `avoid()` and disappear silently.

**Fix:** Add a guard at the top of the per-swath loop:

```cpp
if (path.size() < 2) {
  // Zero-length paths produce no geometry; skip gracefully.
  continue;
}
```

---

### WR-03: `out_id` counter type is `int` but `F2CSwath` constructor accepts `int id`; overflow is theoretical but the counter is never reset across calls

**File:** `src/fields2cover/obstacle/obstacle_avoider.cpp:26,58`

**Issue:** `out_id` is a local `int` initialised to 0 for each `avoid()` call,
so there is no cross-call accumulation. Within a single call, IDs are
reassigned starting at 0 regardless of the IDs the input swaths carried. This
means **the original swath IDs are silently dropped**, which can break any
caller that later uses `getId()` to correlate output swaths back to input
swaths (e.g., when merging route segments or displaying diagnostics).

**Fix:** Preserve the original ID in the common case (no split), and use a
derived ID scheme when a swath is split:

```cpp
// For a single residual, keep original ID:
if (residuals.size() == 1 && seg matches original length) {
  result.emplace_back(seg, sw.getWidth(), sw.getId(), sw.getType());
} else {
  // Split: annotate with original ID + sub-index if needed
  result.emplace_back(seg, sw.getWidth(), out_id++, sw.getType());
}
```

Or at minimum document the contract in the header that IDs are reassigned.

---

## Info

### IN-01: Header guard and `#pragma once` are redundant

**File:** `include/fields2cover/obstacle/obstacle_avoider.h:7-9`

**Issue:** Both `#pragma once` (line 7) and the traditional `#ifndef` / `#define`
guard (lines 8-9) are present. Using both is not wrong, but it is redundant. All
other headers in the codebase (`Cell.h`, `Cells.h`, `Swath.h`, etc.) use the
same dual-guard pattern, so this is consistent with project style. No action
required unless the project decides to standardise.

**Fix:** Leave as-is given project consistency; or adopt either `#pragma once`
alone or the `#ifndef` guard alone project-wide.

---

### IN-02: Test `short_segments_dropped` relies on unstated geometric assumption

**File:** `tests/cpp/obstacle/obstacle_avoider_test.cpp:80-91`

**Issue:** The comment says residuals become "~0.05 m" with `safety_margin=0.45`,
but this assumes OGR's `difference` clips exactly at the inflated square edge.
OGR/GEOS clip operations on axis-aligned squares produce exact results for
simple geometry, so the assumption holds in practice. However, if the OGR
version changes floating-point handling the test could become order-sensitive.
The test asserts `result.size() == 0u` with no tolerance or message — a failure
gives no diagnostic context.

**Fix:** Add a descriptive failure message:

```cpp
EXPECT_EQ(result.size(), 0u)
    << "Expected all short segments (<0.1 m) to be dropped; got "
    << result.size() << " segment(s)";
```

---

### IN-03: No test for multiple input swaths or for swath with zero width

**File:** `tests/cpp/obstacle/obstacle_avoider_test.cpp`

**Issue:** All five tests use a single swath. The implementation iterates over
`swaths.size()` swaths and reassigns `out_id` across all of them, so a
multi-swath scenario exercises the ID-accumulation logic and the swath-ordering
guarantee. A missing test here means the cross-swath ID assignment (WR-03) is
not observable from the test suite.

**Fix:** Add a test using two input swaths with the obstacle intersecting only
one of them, asserting that the untouched swath is returned whole and the split
swath produces the expected number of segments.

---

_Reviewed: 2026-04-29_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
