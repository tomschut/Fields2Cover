# Phase 25: C++ Follower Coordination - Research

**Researched:** 2026-04-29
**Domain:** C++ library extension — fields2cover follower/cart coordination algorithm
**Confidence:** HIGH (primary source: codebase inspection)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
None — all implementation choices are at Claude's discretion (infrastructure phase).

### Claude's Discretion
- Namespace: `f2c::follower` (follow Phase 24 pattern)
- Rendezvous triggering: accumulate robot route length until follower tank would be full → trigger rendezvous at nearest headland point
- Follower path: connect rendezvous points with straight-line travel segments
- Output: follower path geometry (`F2CLineString`) + list of rendezvous coordinates (`std::vector<F2CPoint>`)
- Class name: `FollowerCoordination` or similar

### Deferred Ideas
None.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| F2C-02 | f2c library exposes a follower coordination algorithm that, given a robot's coverage route and follower specs (tank capacity, unload time, travel speed), computes the cart's travel path and the headland rendezvous points where unloading occurs | F2CPath is the correct input; PathState.len + PathSectionType::HL_SWATH enable distance accumulation and headland detection; output is F2CLineString + std::vector<F2CPoint> |
</phase_requirements>

---

## Summary

Phase 25 adds a single new C++ class — `FollowerCoordination` in `f2c::follower` — to the fields2cover library. The class takes a robot coverage path (`F2CPath`) and follower specs (tank capacity in kg/m, unload time in seconds, follower speed in m/s) and returns two outputs: a `F2CLineString` representing the follower's travel path (connecting rendezvous points), and a `std::vector<F2CPoint>` listing the rendezvous coordinates.

The algorithm is distance-accumulation based. `F2CPath` is a sequence of `PathState` records. Each `PathState` has a `.len` field (segment length in metres) and a `.type` field (`PathSectionType::SWATH`, `PathSectionType::TURN`, or `PathSectionType::HL_SWATH`). The robot accumulates covered distance (from SWATH states only — what the robot is harvesting). When accumulated distance × yield_per_metre would fill the follower tank, a rendezvous is triggered at the nearest point of type `HL_SWATH` (headland swath = the robot is in the headland where the cart can meet it).

The entire module follows the pattern established in Phase 24 (`f2c::partition::MultiRobotPartition`): one header, one source, one test file — all auto-discovered by GLOB_RECURSE.

**Primary recommendation:** Use `F2CPath` as input (not `F2CRoute`). Iterate `PathState` records; accumulate `len` from SWATH states; trigger rendezvous when accumulated yield crosses tank capacity; snap rendezvous to the start-point of the next `HL_SWATH` state.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Distance accumulation along robot path | C++ library | — | Pure computation on F2CPath data |
| Rendezvous point identification | C++ library | — | Requires PathState::type discrimination |
| Follower path geometry construction | C++ library | — | Connects rendezvous points as LineString |
| gRPC exposure of this API | Phase 26 | — | Out of scope for Phase 25 |
| Frontend display of results | Phase 28 | — | Out of scope for Phase 25 |

---

## Standard Stack

### Core
| Type | Header | Purpose | Notes |
|------|--------|---------|-------|
| `F2CPath` (= `f2c::types::Path`) | `fields2cover/types.h` | Robot coverage path — the INPUT | Sequence of PathState; has `.length()` and `.size()` |
| `F2CPathState` (= `f2c::types::PathState`) | `fields2cover/types/PathState.h` | Single path segment | `.point`, `.len`, `.type`, `.velocity` |
| `PathSectionType::HL_SWATH` | `fields2cover/types/PathState.h` | Headland swath segment enum value | Use to identify headland segments |
| `PathSectionType::SWATH` | `fields2cover/types/PathState.h` | Coverage swath segment enum value | Accumulate distance from these |
| `F2CPoint` | `fields2cover/types.h` | Rendezvous coordinate | Output list element |
| `F2CLineString` | `fields2cover/types.h` | Follower travel path geometry | Output geometry |
| `F2CRobot` (optional) | `fields2cover/types.h` | May carry `getCovWidth()` for yield rate | Can pass yield_per_metre directly instead |

### Build
| Mechanism | Detail |
|-----------|--------|
| Header location | `include/fields2cover/follower/follower_coordination.h` |
| Source location | `src/fields2cover/follower/follower_coordination.cpp` |
| Test location | `tests/cpp/follower/follower_coordination_test.cpp` |
| CMake discovery | GLOB_RECURSE in top-level `CMakeLists.txt` and `tests/CMakeLists.txt` — **no CMakeLists edits needed** |

**Installation:** No new libraries. Uses only existing fields2cover types and the C++ standard library.

---

## Architecture Patterns

### System Architecture Diagram

```
F2CPath (robot coverage path)          FollowerSpec struct
  │  [PathState, PathState, ...]          │  tank_capacity_m   (kg/m)
  │  type: SWATH / TURN / HL_SWATH       │  unload_time_s
  │  len: segment length in metres       │  follower_speed_mps
  └─────────────────┬─────────────────────┘
                    │
          FollowerCoordination::compute()
                    │
     ┌──────────────┴──────────────┐
     │                             │
     ↓                             ↓
F2CLineString                std::vector<F2CPoint>
(follower travel path —      (rendezvous coordinates —
 connects rendezvous pts     one per unload event)
 in order)
```

**Data flow through algorithm:**
```
for each PathState in path:
  if type == SWATH:
    accumulated_distance += state.len
  if accumulated_distance * yield_per_metre >= tank_capacity_m:
    scan forward for next HL_SWATH state
    rendezvous_point = that state's .point
    append to rendezvous list
    append to LineString
    reset accumulated_distance = 0
```

### Recommended Project Structure
```
include/fields2cover/follower/
└── follower_coordination.h       # class FollowerCoordination

src/fields2cover/follower/
└── follower_coordination.cpp     # implementation

tests/cpp/follower/
└── follower_coordination_test.cpp  # GoogleTest suite
```

### Pattern 1: F2CPath iteration
`F2CPath` stores `std::vector<PathState> states_` internally. Public interface:

```cpp
// Source: verified from include/fields2cover/types/Path.h
path.size()                          // number of PathState records
path[i]                              // operator[] access to PathState
path.begin() / path.end()            // range-for support
path.length()                        // total path length (all states)

// PathState fields (from include/fields2cover/types/PathState.h):
state.point    // F2CPoint — start of this segment
state.len      // double — segment length in metres (>= 0)
state.type     // PathSectionType: SWATH, TURN, or HL_SWATH
state.velocity // double — robot speed on this segment
state.angle    // double — heading

// PathState::atEnd() returns the endpoint of the segment
state.atEnd()  // F2CPoint computed from point + angle + len
```

### Pattern 2: FollowerSpec struct (plain struct, no RTTI)
```cpp
// Source: [ASSUMED] — follows Phase 24 style (no external deps, simple aggregates)
struct FollowerSpec {
  double tank_capacity_m;   // tank capacity in kg per metre of swath width
                             // (or treat as dimensionless "units of coverage")
  double unload_time_s;     // time in seconds the follower spends at rendezvous
  double follower_speed_mps; // follower travel speed in m/s
};
```

**Design note on units:** The CONTEXT.md says "tank capacity" without specifying kg vs. metres. The simplest approach that avoids unit confusion: treat `tank_capacity_m` as the number of **swath-metres** the follower can carry (i.e. accumulated SWATH `.len`). This is unit-consistent — no yield density constant needed, and tests are easy to reason about.

### Pattern 3: Rendezvous triggering
```cpp
// Source: [ASSUMED] — derived from algorithm description in CONTEXT.md
double accumulated = 0.0;
for (size_t i = 0; i < path.size(); ++i) {
  const auto& state = path[i];
  if (state.type == f2c::types::PathSectionType::SWATH) {
    accumulated += state.len;
  }
  if (accumulated >= tank_capacity_m) {
    // Find next HL_SWATH to snap rendezvous point onto headland
    for (size_t j = i + 1; j < path.size(); ++j) {
      if (path[j].type == f2c::types::PathSectionType::HL_SWATH) {
        rendezvous_pts.push_back(path[j].point);
        follower_ls.addPoint(path[j].point);
        accumulated = 0.0;
        i = j;  // advance outer loop past the rendezvous
        break;
      }
    }
  }
}
```

### Pattern 4: Class signature (mirrors Phase 24)
```cpp
// Source: [ASSUMED] — mirrors f2c::partition::MultiRobotPartition pattern
namespace f2c::follower {

class FollowerCoordination {
 public:
  struct FollowerSpec {
    double tank_capacity_m;    // swath-metres before unload needed
    double unload_time_s;      // seconds spent at each rendezvous
    double follower_speed_mps; // follower travel speed m/s
  };

  struct Result {
    F2CLineString path;                    // follower travel geometry
    std::vector<F2CPoint> rendezvous_pts;  // one point per unload event
  };

  Result compute(const F2CPath& robot_path, const FollowerSpec& spec) const;
};

}  // namespace f2c::follower
```

### Anti-Patterns to Avoid
- **Do not use F2CRoute as input.** `F2CRoute` is the intermediate representation (Swaths + connections) before path planning. `F2CPath` is the final output of path planning and is what gets sent to the gRPC shim. Phase 26 will pass a serialized path — it should be `F2CPath`.
- **Do not accumulate TURN or HL_SWATH distance.** Only `PathSectionType::SWATH` represents actual harvesting coverage. Turns and headland traversals do not fill the tank.
- **Do not omit the "no HL_SWATH found" fallback.** If the robot path ends before a headland is reached after threshold, handle gracefully (either use the last path point or skip that rendezvous).
- **Do not use F2CCells for the follower path.** The follower path is a polyline (sequence of line segments), not a polygon. Use `F2CLineString`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Segment length | Manual Euclidean distance | `PathState::len` field directly | Already stored per segment |
| Path total length | Manual sum | `F2CPath::length()` | Sums all state `.len` values |
| Headland detection | Geometry buffer/intersection | `PathSectionType::HL_SWATH` enum check | The path planner already tags headland states |
| Point coordinates | Trigonometry | `PathState::point` + `PathState::atEnd()` | Start and end points of each segment already computed |

---

## Common Pitfalls

### Pitfall 1: HL_SWATH not present in path
**What goes wrong:** If the robot path was generated without headland swaths, no `HL_SWATH` states exist and the scan-forward loop finds nothing — rendezvous never triggers even when tank is full.
**Why it happens:** Some path planners or test setups generate paths with only SWATH + TURN states.
**How to avoid:** After the scan-forward loop, if no HL_SWATH was found before `path.size()`, use the current state's `.point` as a fallback rendezvous (or add a flag to the Result indicating an incomplete rendezvous).
**Warning signs:** Test passes with 0 rendezvous points even when tank_capacity_m is tiny.

### Pitfall 2: Double-counting distance after rendezvous
**What goes wrong:** After a rendezvous, the outer loop index `i` continues from where it left off, but the accumulated distance reset means the segment where the rendezvous was found gets counted again on the next pass.
**Why it happens:** Off-by-one in the loop advance after rendezvous.
**How to avoid:** After finding a rendezvous at index `j`, set `i = j` (the for-loop will increment to `j+1`). Do not re-add `path[j].len` after reset.

### Pitfall 3: Empty path input
**What goes wrong:** Segfault or UB if `robot_path.size() == 0`.
**How to avoid:** Guard at function entry: if path is empty, return empty Result.

### Pitfall 4: tank_capacity_m of 0 or negative
**What goes wrong:** Infinite rendezvous loop (every segment triggers).
**How to avoid:** Validate `spec.tank_capacity_m > 0` at function entry; throw `std::invalid_argument`.

### Pitfall 5: Headland vs. headland swath confusion
**What goes wrong:** Confusing `PathSectionType::HL_SWATH` (a swath mowed on the headland area — the robot traverses it) with the headland geometry itself. HL_SWATH is where the follower can safely meet the robot.
**Why it matters:** Only `HL_SWATH` states (not `TURN` states) are suitable rendezvous locations because the robot is moving slowly and predictably on a known line.

---

## Code Examples

### Iterating PathStates and checking type
```cpp
// Source: verified from include/fields2cover/types/Path.h and PathState.h
for (const auto& state : robot_path) {
  if (state.type == f2c::types::PathSectionType::SWATH) {
    accumulated += state.len;
  } else if (state.type == f2c::types::PathSectionType::HL_SWATH) {
    // potential rendezvous location
  }
}
```

### Building F2CLineString from points
```cpp
// Source: verified from include/fields2cover/types/LineString.h
F2CLineString follower_path;
follower_path.addPoint(rendezvous_a);
follower_path.addPoint(rendezvous_b);
// or initializer: F2CLineString{pt1, pt2, pt3}
```

### Test helper: building a synthetic F2CPath
```cpp
// Source: [ASSUMED] — derived from Path.h addState signature
F2CPath buildTestPath(double swath_len, int n_swaths) {
  F2CPath p;
  for (int i = 0; i < n_swaths; ++i) {
    // add swath segment
    p.addState(F2CPoint(i * swath_len, 0), 0.0, swath_len,
        f2c::types::PathDirection::FORWARD,
        f2c::types::PathSectionType::SWATH, 1.0);
    // add turn
    p.addState(F2CPoint((i + 1) * swath_len, 0), M_PI_2, 3.0,
        f2c::types::PathDirection::FORWARD,
        f2c::types::PathSectionType::TURN, 0.5);
    // add headland swath
    p.addState(F2CPoint((i + 1) * swath_len, 3.0), M_PI, 2.0,
        f2c::types::PathDirection::FORWARD,
        f2c::types::PathSectionType::HL_SWATH, 0.8);
  }
  return p;
}
```

---

## Validation Architecture

### What F2C-02 Requires

F2C-02 states: "given a robot's coverage route and follower specs (tank capacity, unload time, travel speed), computes the cart's travel path and the headland rendezvous points."

The ROADMAP success criteria for Phase 25 translates to four verifiable behaviors:

| Criterion | Verifiable by | Test type |
|-----------|--------------|-----------|
| `compute()` returns a Result with non-null path and rendezvous list | Automated | Unit |
| Rendezvous count scales inversely with tank capacity (smaller tank → more rendezvous) | Automated | Unit (two calls, compare counts) |
| With tank_capacity_m > total route SWATH length → zero rendezvous | Automated | Unit |
| All 298 existing tests still pass | Automated | GoogleTest full suite |

### Test Framework

| Property | Value |
|----------|-------|
| Framework | GoogleTest (already installed; version bundled with fields2cover) |
| Config | `tests/CMakeLists.txt` — GLOB_RECURSE auto-discovers `tests/cpp/follower/follower_coordination_test.cpp` |
| Quick run | `./build/tests/unittests --gtest_filter="fields2cover_follower_coordination*"` |
| Full suite | `./build/tests/unittests` |
| Build | `make -C build unittests -j$(nproc)` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-02a | Small tank triggers at least 1 rendezvous | unit | `--gtest_filter=*small_tank*` | Wave 0 |
| F2C-02b | Large tank (> total length) triggers 0 rendezvous | unit | `--gtest_filter=*large_tank*` | Wave 0 |
| F2C-02c | Rendezvous count increases as capacity halves | unit | `--gtest_filter=*capacity_scaling*` | Wave 0 |
| F2C-02d | Empty path returns empty Result without crash | unit | `--gtest_filter=*empty_path*` | Wave 0 |
| F2C-02e | Zero capacity throws std::invalid_argument | unit | `--gtest_filter=*zero_capacity*` | Wave 0 |
| F2C-02f | Rendezvous points land on HL_SWATH states (not TURN) | unit | `--gtest_filter=*rendezvous_on_headland*` | Wave 0 |
| F2C-02g | No regression in existing 298 tests | regression | `./build/tests/unittests` | Exists (298 pass) |

### Sampling Rate
- Per task commit: `make -C build unittests -j$(nproc) && ./build/tests/unittests --gtest_filter="fields2cover_follower_coordination*"`
- Per wave merge: `./build/tests/unittests` (full suite — must be 298 + new tests, 0 failures)
- Phase gate: Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `tests/cpp/follower/follower_coordination_test.cpp` — covers F2C-02a through F2C-02f
- [ ] `include/fields2cover/follower/follower_coordination.h` — class under test
- [ ] `src/fields2cover/follower/follower_coordination.cpp` — implementation

*(No framework changes needed — GLOB_RECURSE already covers `cpp/*/*.cpp`)*

---

## State of the Art

| Concern | Approach Used Here | Rationale |
|---------|-------------------|-----------|
| Input type | `F2CPath` (not `F2CRoute`) | F2CPath is post-path-planning — has all segment metadata including HL_SWATH tags |
| Headland detection | `PathSectionType::HL_SWATH` enum | Already encoded in path; no geometry re-analysis needed |
| Follower path | `F2CLineString` (straight lines between rendezvous) | Matches CONTEXT.md decision; follower travel is unconstrained (no swath width constraints) |
| Rendezvous output | `std::vector<F2CPoint>` | Simplest typed list; consistent with Phase 26 gRPC serialization target |
| Algorithm | Greedy distance accumulation | Correct for single-cart case; multi-cart deferred to FOL-F01 |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `FollowerSpec` is best as an inner struct of the class (not a standalone type) | Architecture Patterns | Low — inner struct is easy to promote; Phase 26 will wrap it in proto anyway |
| A2 | Tank capacity modelled as swath-metres (accumulated SWATH `.len`) rather than kg requires yield_per_metre constant | Architecture Patterns | Medium — if API consumers expect kg, a `yield_per_metre` param is needed; but CONTEXT.md says "kg" so best to accept kg directly and require caller to supply a yield density constant, OR treat the capacity parameter as a pure distance threshold named clearly |
| A3 | No HL_SWATH found → fallback to current point | Common Pitfalls | Low — fallback behaviour can be adjusted in planning/implementation |
| A4 | `PathState::atEnd()` returns the endpoint of the segment correctly | Code Examples | Low — verified method exists in header; behaviour confirmed by name |

**Note on A2:** The CONTEXT.md says "tank capacity" in kg. The path only carries swath lengths (metres). To bridge them, the function signature should accept either: (a) `tank_capacity_m` as a distance threshold (simple, testable), or (b) `tank_capacity_kg` + `yield_per_metre_kg_m` as two params. Option (a) is recommended for Phase 25 because Phase 26's gRPC contract can always convert units on the server side. The planner should pick one and state it clearly in the plan.

---

## Open Questions

1. **Tank capacity units: distance vs. kg**
   - What we know: CONTEXT.md says "kg"; F2CPath has only distances
   - What's unclear: Does the algorithm need a yield density parameter, or should Phase 25 use distance-based capacity and let Phase 26 do unit conversion?
   - Recommendation: Use `double tank_capacity_swath_m` (pure distance threshold) in Phase 25. Document the unit clearly. Phase 26 adds `yield_per_metre_kg_m` to the gRPC request and converts before calling.

2. **Rendezvous snap: use `state.point` or `state.atEnd()`?**
   - What we know: `PathState::point` is the start of the segment; `atEnd()` computes the end
   - What's unclear: Is the start or end of the HL_SWATH the better meeting point?
   - Recommendation: Use `state.point` (start of the HL_SWATH segment) — the follower arrives as the robot enters the headland, not after.

---

## Environment Availability

Step 2.6: SKIPPED — pure C++ library change, no external tool dependencies beyond the existing build system.

**Build verified:** `./build/tests/unittests` currently reports 298 tests passing (verified from STATE.md: "298 tests green" after Phase 24). Build system confirmed to be cmake with GLOB_RECURSE auto-discovery.

---

## Sources

### Primary (HIGH confidence — codebase inspection)
- `include/fields2cover/types/Path.h` — F2CPath interface: size(), operator[], begin/end, length(), addState(), PathState fields
- `include/fields2cover/types/PathState.h` — PathSectionType enum: SWATH=1, TURN=2, HL_SWATH=3; PathState struct fields
- `include/fields2cover/types/LineString.h` — F2CLineString: addPoint(), length(), size(), startPoint(), endPoint()
- `include/fields2cover/types/Point.h` — F2CPoint: X(), Y(), getX/Y/Z(), setPoint(), getPointFromAngle()
- `include/fields2cover/types.h` — alias map: F2CPath, F2CPathState, F2CLineString, F2CPoint all confirmed
- `include/fields2cover/partition/multi_robot_partition.h` — Phase 24 naming/guard/namespace pattern to mirror
- `tests/CMakeLists.txt` — GLOB_RECURSE pattern: `cpp/*/*.cpp` covers `tests/cpp/follower/*.cpp` automatically
- `CMakeLists.txt` (top-level) — GLOB_RECURSE `src/*.cpp` covers `src/fields2cover/follower/*.cpp` automatically

### Secondary (MEDIUM confidence)
- `.planning/phases/24-c-multi-robot-partitioning/24-01-PLAN.md` — confirmed file layout, guard format, namespace, license header, CMakeLists no-edit rule
- `.planning/phases/25-c-follower-coordination/25-CONTEXT.md` — confirmed algorithm choices (distance accumulation, HL rendezvous, F2CLineString output)

### Tertiary (LOW — not independently verified)
- A2 and A3 in Assumptions Log above

---

## Metadata

**Confidence breakdown:**
- Type system (F2CPath, PathState, PathSectionType): HIGH — verified by reading actual headers
- Module structure / CMake discovery: HIGH — verified by reading CMakeLists.txt and Phase 24 plan
- Algorithm logic (accumulate → HL_SWATH snap): MEDIUM — design is consistent with type system; exact loop boundary conditions need implementation-time verification
- Output types (F2CLineString + std::vector<F2CPoint>): HIGH — confirmed available and appropriate

**Research date:** 2026-04-29
**Valid until:** Stable — no external dependencies; valid until fields2cover type system changes
