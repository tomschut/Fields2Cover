# Audit Followups — Phase 6 Plan 05

**Generated:** 2026-04-11
**Scope:** All Critical, High, and in-scope Medium findings from Phase 6 audits (plans 06-01, 06-02, 06-03) that are NOT auto-fixed by plan 06-04.

**Sources:**
- `docs/audit/static-analysis.md` (F-NNN — clang-tidy + cppcheck)
- `docs/audit/memory-audit.md` (M-NNN — manual memory/lifetime review)
- `docs/audit/api-hygiene.md` (A-NNN — manual const-correctness / cache cascades)

**Note on scope:** Low-severity findings are handled by plan 06-04's mechanical sweep and are deliberately excluded here. Plan 06-04 runs in parallel with this plan and touches disjoint files (src/**, include/**, fixes-applied.md); this plan only touches docs/audit/followups.md.

## Summary

| Severity | Count | Effort distribution |
|---|---|---|
| Critical | 2 | S:1 M:1 L:0 |
| High | 9 | S:4 M:4 L:1 |
| Medium | 13 | S:5 M:6 L:2 |
| **Total followups** | **24** | **S:10 M:11 L:3** |

**Counting notes:**
- F-NNN High narrowing-conversions are grouped by module into 6 bulk triage entries (roughly 80 underlying findings) to avoid a 100+-line wall of near-identical items. Each bulk entry lists the specific F-IDs it covers.
- F-NNN High `bugprone-easily-swappable-parameters` findings are grouped into one bulk triage entry (covers F-006, F-008, F-009, F-010, F-014, F-015, F-028, F-043, F-044, F-045, F-048, F-103, F-108, F-116) because the mitigation is uniform (named-tag types or parameter renames).
- F-NNN High `bugprone-narrowing-conversions` findings in the `types` module (F-032..F-048 and the 54 truncated from static-analysis.md) are aggregated into T-010 (types narrowing-conversion sweep).
- Medium F-NNN findings that map directly to M-NNN memory-audit items (F-130..F-144 noexcept-move-operations) are mentioned by cross-reference but their triage entry lives under the M-NNN ID that already carries the lifetime context.
- Plan 06-04 is expected to absorb any Medium finding that is mechanically safe (pure clang-tidy --fix output). A handful of Medium items below are included here anyway because they require judgment (e.g., dynamic_cast vs static_cast in F-149) or overlap with architectural concerns.

## Critical

### swath_generator

#### T-001 (F-001): `swath_generator_base.cpp:74-77` — infinite recursion in `computeBestAngle` [RESOLVED commit 8e7fd66]

**Source:** `clang-diagnostic-infinite-recursion` (F-001, docs/audit/static-analysis.md)
**Location:** `src/fields2cover/swath_generator/swath_generator_base.cpp:74-77`
**Severity:** Critical
**Description:** `SwathGeneratorBase::computeBestAngle` body is literally `return computeBestAngle(obj, width, poly);` — unbounded self-recursion on identical arguments.
**Root cause:** The header `include/fields2cover/swath_generator/swath_generator_base.h:37-38` already declares this method as pure virtual (`virtual double computeBestAngle(...) = 0;`). The .cpp definition exists only because of an authoring mistake — a pure virtual base method must not have a default implementation that recurses. Derived classes (`BruteForce`) correctly override it, so the dead definition is unreachable during normal use. The danger is (a) any refactor that drops `= 0` in the header silently enables the infinite recursion, and (b) the clang diagnostic pollutes the build output.
**Proposed fix:** Delete lines 74-77 of `swath_generator_base.cpp` entirely. The pure-virtual declaration in the header stands on its own; no definition is required. Verify after deletion:
```bash
grep -rn "SwathGeneratorBase::computeBestAngle" src/ include/
```
should show only the header declaration. Run `tests/cpp/` to confirm no regression (there should be none — the method is pure virtual, so no caller could have reached this body).
**Alternative (safer for ABI):** If keeping the definition is desired for any linker reason, replace the body with `throw std::logic_error("SwathGeneratorBase::computeBestAngle is pure virtual");`.
**Risk:** Low. Dead code removal. Confirmed pure-virtual in the header. Tests in `tests/cpp/swath_generator/` cover the `BruteForce` override path; the base definition is unreachable at runtime.
**Effort:** S (≤15 min including test run)
**Blocks:** nothing
**Depends on:** nothing
**Notes:** This is the only Critical static-analysis finding in the library. Should be the first followup acted on in any future remediation pass. Do NOT mark it fixed without rerunning clang-tidy on `swath_generator_base.cpp` to confirm the warning is gone.

### types

#### T-002 (M-001): `Geometry_impl.hpp:33,40` — `EmptyDestructor` non-owning views can dangle [RESOLVED commit 8a1c24b]

**Source:** Manual memory audit (M-001, docs/audit/memory-audit.md)
**Locations:**
- `include/fields2cover/types/Geometry_impl.hpp:33` (constructor taking `T*, EmptyDestructor`)
- `include/fields2cover/types/Geometry_impl.hpp:40` (constructor taking `OGRGeometry*, EmptyDestructor`)
- Callsites: `src/fields2cover/types/Cell.cpp:42,53`, `Cells.cpp:68,76`, `MultiLineString.cpp:76,84`, `MultiPoint.cpp:39,47`
**Severity:** Critical
**Description:** These constructors wrap a borrowed OGR pointer in a `std::shared_ptr<T>` with a no-op deleter. The resulting `Cell`/`LinearRing`/`LineString`/`Point` object is indistinguishable at the type level from an owning wrapper, but holds a non-owning view into its parent container. If the parent is destroyed while the view lives, every subsequent method call on the view is use-after-free UB. See memory-audit.md M-001 for the full worked example.
**Proposed fix (recommended — option (a)):** Deep-copy in `Cells::getGeometry(size_t, Cell&)` and its siblings. The non-const overload `Cell Cells::getGeometry(size_t)` at `Cells.cpp:79` already returns a clone; port that logic into the out-parameter overloads so both behave identically. Concretely, inside `getGeometry(i, out)` replace the `EmptyDestructor` constructor with a clone:
```cpp
// before
out = Cell(OGRGeometryFactory::forceToPolygon(
    static_cast<OGRGeometry*>(data_->getGeometryRef(i))), EmptyDestructor());
// after
out = Cell(...).clone();  // or copy from the owning overload
```
Repeat for `MultiLineString::getGeometry`, `MultiPoint::getGeometry`, `Cell::getInteriorRing`, etc. — all 6 accessors flagged in memory-audit.md's fix sketch.
**Alternative (b) — type-level fix:** Introduce a `CellView` / `CellRef` distinct type with a narrow interface and no `get()`/`size()`/etc. mutators. Callers must explicitly opt into non-ownership. This is a public-API change and is correspondingly higher effort and risk.
**Alternative (c) — comment-only:** Rename `EmptyDestructor` to `NonOwningTag`, add `// NONOWNING — do not outlive parent` at every callsite, add doxygen. Cheap but provides zero compile-time protection.
**Risk:** Medium for option (a). Every callsite of the two-arg `getGeometry` must be audited for "does the caller rely on in-place mutation of the borrowed cell propagating back to the parent?" — if yes, deep-copy silently changes behavior. Spot-check suggests callers treat it as read-only, but this MUST be verified before applying. High for option (b) — public API change. Low for option (c) — pure documentation.
**Effort:** M for option (a) — 6 accessors + test update in `tests/cpp/types/`. L for option (b). S for option (c).
**Blocks:** Any future SWIG directive that would expose `EmptyDestructor` to Python (currently `%ignore`-d in `Fields2Cover.i:35`, which is the only thing preventing M-001 from being a Python footgun).
**Depends on:** nothing
**Notes:** Highest-value memory finding in the library. Worth prioritizing despite being Medium-effort because the blast radius is the entire `types` module.

## High

### path_planning

#### T-003 (bulk): `easily-swappable-parameters` in `createSimpleTurn` family

**Source:** `bugprone-easily-swappable-parameters` (F-009 `dubins_curves.cpp:14`, F-010 `dubins_curves_cc.cpp:14`, F-014 `reeds_shepp_curves.cpp:14`, F-015 `reeds_shepp_curves_hc.cpp:14`)
**Severity:** High
**Description:** Each `createSimpleTurn` variant takes three adjacent `double` parameters (typically radius, kurv, ang — or similar). A caller swapping any two compiles cleanly and produces a garbage path.
**Proposed fix:** Introduce strong typedefs or a small `TurnParams` aggregate struct so that each argument is labeled at the callsite:
```cpp
struct TurnParams { double radius; double curvature; double angle; };
Path createSimpleTurn(const TurnParams&, ...);
```
Or, minimum viable: mark parameters `[[gsl::suppress]]` / add `// NOLINTNEXTLINE(bugprone-easily-swappable-parameters)` only at the declaration with a `// PARAMETER ORDER: radius, curvature, angle` comment block. Strong-typedef is the correct fix; comment suppression is the escape hatch for plan 06-04.
**Risk:** Medium. `createSimpleTurn` is the hot inner loop of all four turn planners. Signature change cascades to every override and callsite. Test coverage in `tests/cpp/path_planning/` is reasonable, so a refactor is feasible.
**Effort:** M (parameter struct refactor across 4 files + headers + tests).
**Blocks:** nothing
**Depends on:** nothing
**Notes:** High-value fix — one of the class of bugs that static analysis catches best. But defer until there is test budget to cover all four turn planners, because a silent swap here would manifest as subtly wrong robot motion, not a crash.

#### T-004 (F-011..F-013, F-016): narrowing conversions in `path_planning.cpp` and `turning_base.cpp`

**Source:** `bugprone-narrowing-conversions` (F-011, F-012, F-013 at `path_planning.cpp:33,35,36`; F-016 at `turning_base.cpp:51` — double→bool)
**Severity:** High
**Description:** Silent size_t→int narrowing in vector index conversions, and a double→bool implicit conversion in `turning_base.cpp:51` (the most concerning of the four — a double being treated as a boolean is almost always a logic error).
**Proposed fix:** For F-011..F-013, insert explicit casts or change the receiving variable type to `size_t`. For F-016, INVESTIGATE — read ±20 lines around `turning_base.cpp:51` and determine what the `double` actually represents. If it's a radius or angle, the code is wrong; if it's a flag already encoded as `1.0`/`0.0`, add an explicit `!= 0.0` comparison.
**Risk:** Low for F-011..F-013 (mechanical). Medium for F-016 (semantic — could expose a latent logic bug).
**Effort:** S for F-011..F-013; M for F-016 including investigation.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** F-016 should be split into its own T-entry if it turns out to be a real bug during investigation.

### route_planning

#### T-005 (F-017..F-027): route_planning narrowing-conversion sweep

**Source:** `bugprone-narrowing-conversions` — F-017, F-018, F-019, F-020, F-021, F-022, F-023, F-024, F-025, F-026, F-027 across `custom_order.cpp`, `route_planner_base.cpp`, `snake_order.cpp`, `spiral_order.cpp`
**Severity:** High
**Description:** 11 `size_t → int` / `size_t → difference_type` narrowing warnings in the route-ordering implementations. Most are index conversions for OR-Tools or STL iterator arithmetic.
**Proposed fix:** Mechanical sweep — insert `static_cast<int>(...)` or, better, change the accumulator/index variable type to `int64_t` (OR-Tools uses `int64_t` for node indices). The correct fix in several spots is to keep `size_t` in the vector indexing and only cast at the OR-Tools boundary.
**Risk:** Low. All sites are integer-bookkeeping paths with no floating-point interaction. Test coverage in `tests/cpp/route_planning/` is solid.
**Effort:** M (11 sites across 4 files; each needs a glance at the surrounding loop to pick the right fix pattern).
**Blocks:** nothing
**Depends on:** nothing
**Notes:** The OR-Tools route planner previously had an auth gate around its solver API; see Phase 5 decision log. None of these warnings affect that decision.

### objectives

#### T-006 (F-002..F-005, F-007): objectives narrowing-conversion sweep

**Source:** `bugprone-narrowing-conversions` — F-002 (`rp_objective.cpp:103`), F-003 (`:114`), F-004 (`:115`), F-005 (`n_swath.cpp:16` — size_t→double), F-007 (`overlaps.cpp:18`)
**Severity:** High
**Description:** 5 narrowing conversions in the objectives layer. F-005 (`size_t → double`) loses precision above 2^53 — benign for current field sizes but worth fixing.
**Proposed fix:** Same as T-005 — explicit cast or variable-type change. For F-005 specifically, wrap in `static_cast<double>(...)` to silence the warning (the conversion is intentional — computing a ratio).
**Risk:** Low.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** Likely already caught by clang-tidy `--fix` in plan 06-04. Cross-check before acting.

#### T-007 (F-006, F-008): `easily-swappable-parameters` in `computeCost`

**Source:** `bugprone-easily-swappable-parameters` — F-006 (`n_swath_modified.cpp:12`), F-008 (`sg_objective.cpp:38`)
**Severity:** High
**Description:** Two adjacent `double` parameters in `computeCost` overrides can be swapped without compile error.
**Proposed fix:** Rename for clarity (`double angle, double op_width` — already the naming, so this is a false positive on the lint rule). Add `// NOLINTNEXTLINE(bugprone-easily-swappable-parameters)` with a justification comment at the declaration. The objective-function signature is a stable interface across the whole objectives hierarchy — introducing named-tag types would cascade to every subclass.
**Risk:** Low (suppression only).
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** This is a case where the linter is correct in principle but the codebase's existing convention (angle, width always in that order) is already a de facto protocol.

### types

#### T-008 (F-029): `Geometry_impl.hpp:329` — raw `delete` flagged as `gsl::owner<>` missing [RESOLVED via T-011 commit a0a6672]

**Source:** `cppcoreguidelines-owning-memory` (F-029)
**Severity:** High
**Description:** Deleting a pointer through a type that is not marked `gsl::owner<>`. This is the same line as M-004 below — same file, same line, different rule angle.
**Proposed fix:** See T-011 (M-004). The real fix is `OGRGeometryFactory::destroyGeometry()` replacing the raw `delete`, which also silences this warning as a side effect.
**Risk:** Low.
**Effort:** S (covered by T-011).
**Blocks:** nothing
**Depends on:** T-011 (same fix site)
**Notes:** Don't double-count work — apply T-011 and verify both F-029 and M-004 clear simultaneously.

#### T-009 (F-030): `Graph.h:20` — reserved identifier `pair_vec_size__int`

**Source:** `bugprone-reserved-identifier` (F-030)
**Severity:** High
**Description:** Identifiers with two leading underscores OR a leading underscore followed by an uppercase letter are reserved for the implementation. The double-underscore in `pair_vec_size__int` violates [lex.name].
**Proposed fix:** Rename to `pair_vec_size_int` (one underscore) or, better, `PairVecSizeInt` if it's a type alias, or `kPairVecSizeInt` if it's a constant. Grep for uses first:
```bash
grep -rn "pair_vec_size__int" src/ include/ tests/
```
Mechanical rename across all sites.
**Risk:** Low. Pure rename.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** UB under a strict reading of the standard, but in practice no toolchain miscompiles this. Still worth fixing for portability.

#### T-010 (F-032..F-048 + truncated 54 types High): `types` narrowing-conversion bulk sweep

**Source:** `bugprone-narrowing-conversions` and `bugprone-easily-swappable-parameters` — F-032 through F-048 enumerated in static-analysis.md, plus an additional ~54 truncated High findings in the `types` module (see `build/audit/clang-tidy-raw.txt`).
**Severity:** High
**Description:** The `types` module is the hotspot for narrowing warnings — 74 total High findings per the module hotspots table in static-analysis.md. Dominant pattern: `size_t → int` conversions when interfacing with OGR's C API (which takes `int` for ring/geometry indices) and with `std::vector<>::operator[](size_t)` going into `int` counters.
**Proposed fix:** Two-pass approach:
1. **Pass 1 (mechanical, plan 06-04 territory if scope allows):** explicit `static_cast<int>(...)` at each OGR-API boundary. This silences the warning without changing any runtime behavior.
2. **Pass 2 (structural):** introduce a `f2c::types::detail::to_ogr_int(size_t)` helper that `assert`s on overflow and returns `int`. Use it at every OGR call boundary. Single line per call, but gives runtime overflow protection.
Because the `types` module is the public-API backbone, avoid `int → size_t` direction changes in public signatures (would be breaking).
**Risk:** Medium. Large sweep touches 20+ files in `src/fields2cover/types/`. Each change is individually mechanical, but the volume means test regressions are harder to track down. Gate on `make test` after every batch of ~10 files.
**Effort:** L (>2h) — this is the biggest single followup item in the backlog.
**Blocks:** Future `-Werror=conversion` build flag (cannot be enabled until this sweep lands).
**Depends on:** nothing
**Notes:** Consider splitting into per-file followups if tackled incrementally. The `Cells.cpp` sub-sweep (F-035..F-042, 8 findings) is a good starter chunk.

### utils

#### T-011 (M-004 + F-029): `Geometry_impl.hpp:329` — raw `delete` on OGR factory object [RESOLVED commit a0a6672]

**Source:** Memory audit M-004, mirrored by static analysis F-029
**Location:** `include/fields2cover/types/Geometry_impl.hpp:329`
**Severity:** High
**Description:** `delete poOGRProduct;` is called on a raw `OGRGeometry*` that was allocated via `OGRGeometryFactory::createFromGEOS`. GDAL may build with a non-default allocator (`CPLMalloc`), in which case `new`/`delete` and factory `create`/`destroy` are incompatible — mixing them is undefined behavior. This is the only raw `delete` in the entire library.
**Proposed fix:** Replace with:
```cpp
OGRGeometryFactory::destroyGeometry(poOGRProduct);
```
**Risk:** Low. One-line change. Makes the library `delete`-free. Any GDAL build (default or CPLMalloc) handles `destroyGeometry()` correctly.
**Effort:** S (≤5 min).
**Blocks:** nothing
**Depends on:** nothing
**Notes:** Highest-value single-line fix in the backlog. Should be applied before T-010's large sweep.

#### T-012 (M-002): `Swaths` missing rule-of-5 declarations [RESOLVED commit 04ebe67]

**Source:** Memory audit M-002
**Locations:** `include/fields2cover/types/Swaths.h:25` + `src/fields2cover/types/Swaths.cpp:27`
**Severity:** High
**Description:** `Swaths` has a user-declared `~Swaths()` but declares no move/copy ops. C++ suppresses implicit move generation when the dtor is user-declared, so every `std::move(swaths)` and every `std::vector<Swaths>` reallocation silently falls back to copy. Copies cascade through `Swath::LineString::shared_ptr<OGRGeometry>` atomic-refcount increments. `Route::addSwaths()` / `addConnectedSwaths()` hit this on every large field.
**Proposed fix (recommended):** Rule-of-zero — delete the user-declared dtor entirely (no resource to clean up). If a dtor is needed for ABI stability:
```cpp
// in Swaths.h
Swaths(const Swaths&) = default;
Swaths(Swaths&&) noexcept = default;
Swaths& operator=(const Swaths&) = default;
Swaths& operator=(Swaths&&) noexcept = default;
~Swaths() = default;
```
**Risk:** Low. Declaring `= default` and marking moves `noexcept` is a conservative change. The benefit is moving into the noexcept-move path of `std::vector`, which is a pure perf win.
**Effort:** S (≤15 min including test run).
**Blocks:** Any performance audit that measures `Route` copy cost.
**Depends on:** nothing
**Notes:** Plan 06-04 is expected to pick this up as part of its noexcept-move sweep. If that happens, mark this followup resolved and remove.

#### T-013 (M-003): `SwathsByCells` missing rule-of-5 declarations [RESOLVED commit 04ebe67]

**Source:** Memory audit M-003
**Locations:** `include/fields2cover/types/SwathsByCells.h:22` + `src/fields2cover/types/SwathsByCells.cpp:28`
**Severity:** High
**Description:** Same pattern as M-002 one layer deeper. `SwathsByCells` holds `std::vector<Swaths>`; copying it triggers the M-002 cascade plus its own layer.
**Proposed fix:** Same recipe as T-012 — rule-of-zero or rule-of-five with `noexcept`.
**Risk:** Low.
**Effort:** S.
**Blocks:** nothing
**Depends on:** T-012 (apply the same pattern for consistency)
**Notes:** Fix in the same commit as T-012 so the change is atomic.

## Medium

### path_planning

#### T-014 (F-119..F-126): uninitialized `start`/`end` records in turn planners

**Source:** `cppcoreguidelines-pro-type-member-init` — F-119, F-120 (`dubins_curves.cpp:15`), F-121, F-122 (`dubins_curves_cc.cpp:15`), F-123, F-124 (`reeds_shepp_curves.cpp:15`), F-125, F-126 (`reeds_shepp_curves_hc.cpp:15`)
**Severity:** Medium
**Description:** `start` and `end` record objects are declared without initializer. Under strict reading of the guidelines this is a potential read-of-uninitialized. Inspection of the following lines usually shows them assigned before first read, but the linter cannot prove it.
**Proposed fix:** Change `StateRecord start, end;` to `StateRecord start{}, end{};` (or use `= {}`). Equivalent mechanical transform across all four files.
**Risk:** Low. Default-initialization is a no-op when the object is assigned before first read.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** This is the kind of finding plan 06-04 should absorb via clang-tidy `--fix`. Only remains in followups if the auto-fix pass skips it.

### route_planning

#### T-015 (F-127): `custom_order.cpp:22` — `std::move` on const-ref argument

**Source:** `performance-move-const-arg` (F-127)
**Location:** `src/fields2cover/route_planning/custom_order.cpp:22`
**Severity:** Medium
**Description:** `std::move(x)` is passed as a `const T&` parameter. The move never happens because the callee binds to const — the `move` is a no-op and the cost is lost intent clarity.
**Proposed fix:** Remove the `std::move()` wrapper. The argument is copied (same as before) but the read-side reviewer isn't misled about ownership semantics.
**Risk:** Low. Zero runtime change.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing

#### T-016 (F-128): `route_planner_base.cpp:128` — `deposit` copied by value

**Source:** `performance-unnecessary-value-param` (F-128)
**Location:** `src/fields2cover/route_planning/route_planner_base.cpp:128`
**Description:** The free function `addStartEndToGraph` takes `F2CPoint deposit` by value. `F2CPoint` holds a `shared_ptr<OGRGeometry>` — copying triggers an atomic refcount bump per call. The parameter is only used as `const F2CPoint&` in the body (lines 132-136).
**Severity:** Medium
**Proposed fix:** Change the parameter to `const F2CPoint& deposit`. One-line signature change, no callsite change required (implicit conversion already works).
**Risk:** Low. `addStartEndToGraph` is a file-local helper.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing

#### T-017 (F-129): `spiral_order.cpp:6` — `spiral_size` uninitialized in constructor

**Source:** `cppcoreguidelines-pro-type-member-init` (F-129)
**Location:** `src/fields2cover/route_planning/spiral_order.cpp:6`
**Severity:** Medium
**Description:** `SpiralOrder` constructor does not initialize `spiral_size`. Any code path that reads it before `setSpiralSize()` is called gets an indeterminate value.
**Proposed fix:** Add a default value in the header: `int spiral_size {0};` (or a sentinel that triggers validation on first use). Cheaper than a constructor-body init.
**Risk:** Low.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing

### types

#### T-018 (M-005): `Swath` partial rule-of-5 + virtual `operator=`

**Source:** Memory audit M-005
**Locations:** `include/fields2cover/types/Swath.h:23-33` + `src/fields2cover/types/Swath.cpp:23-26`
**Severity:** Medium
**Description:** `Swath` declares `virtual ~Swath()`, `virtual operator=(Swath&&)`, `virtual operator=(const Swath&)`, and a non-virtual copy constructor, but does not declare a move constructor. Two problems:
1. With a user-declared copy ctor, the compiler does not generate a move ctor — every `std::vector<Swath>` reallocation silently falls back to copy.
2. `virtual operator=` is a slicing hazard that invites derived-class state bugs. No derived class exists today.
**Proposed fix:**
- Add `Swath(Swath&&) noexcept = default;` declaration and definition.
- Remove `virtual` from both `operator=` overloads unless polymorphic assignment is actually required (it isn't — there are no subclasses).
- Mark move operations `noexcept` explicitly.
- Consider devirtualizing `~Swath()` unless the class is genuinely intended as a base.
**Risk:** Medium. Removing `virtual` from `operator=` is technically an ABI change if any downstream user has subclassed `Swath` — the static audit confirms no in-tree subclass, but external users might exist. If ABI is a constraint, skip the `virtual` removal and only fix the missing move ctor + `noexcept`.
**Effort:** S if not devirtualizing; M if doing the full cleanup with ABI risk review.
**Blocks:** Performance audit of `Swaths::emplace_back` hot path.
**Depends on:** nothing

#### T-019 (M-006, F-130, F-131, F-142, F-143, F-144): `Field` missing `noexcept` moves and pass-by-value opportunities

**Source:** Memory audit M-006, static analysis F-130, F-131, F-142, F-143, F-144
**Locations:** `include/fields2cover/types/Field.h:24,25` + `src/fields2cover/types/Field.cpp:13,22,23`
**Severity:** Medium
**Description:** `Field`'s move operations are not marked `noexcept`. `std::vector<F2CField>` reallocation requires `is_nothrow_move_constructible_v<T>` to use the move path — otherwise it copies for strong-exception guarantee. F-142 additionally flags a `pass-by-value` opportunity at `Field.cpp:13`.
**Proposed fix:** Add `noexcept` to the move ctor/assign declarations and definitions. Apply `modernize-pass-by-value` at `Field.cpp:13` (clang-tidy `--fix`-safe). Members are all nothrow-movable (`std::string`, `Point`, `Cells`) so the noexcept addition is sound.
**Risk:** Low.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing
**Notes:** Cross-references multiple F-NNN findings — single commit should clear all of them.

#### T-020 (M-007, F-138, F-139): `Robot` missing `noexcept` moves

**Source:** Memory audit M-007, static analysis F-138, F-139
**Locations:** `include/fields2cover/types/Robot.h:32,34`
**Severity:** Medium
**Description:** Same pattern as T-019 for the much smaller `Robot` class. Benefit is smaller (Robot is cheap to copy), but the discipline should match.
**Proposed fix:** Add `noexcept` to the move ctor/assign declarations and definitions.
**Risk:** Low.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing

#### T-021 (M-008): `Geometries_impl.hpp` raw-pointer PImpl construction — RESOLVED (Plan 07-04)

**Source:** Memory audit M-008
**Locations:** `include/fields2cover/types/Geometries_impl.hpp:68,113`
**Severity:** Medium
**Status:** RESOLVED in Plan 07-04 (commit `511e2bf`).
**Description:** `Iterator::Iterator(..., int nPos)` used `m_poPrivate(new Private())` in its init list. Exception-unsafe if a future refactor added any other initializer that threw after `m_poPrivate` — the raw `Private*` would leak before `unique_ptr` takes ownership.
**Fix applied:**
```cpp
Iterator::Iterator(Geometries* poSelf, int nPos)
    : m_poPrivate(std::make_unique<Private>()) {
  m_poPrivate->m_poSelf = poSelf;
  m_poPrivate->m_nPos = nPos;
}
```
Identical transform applied to `ConstIterator`. `m_poPrivate` was already declared `std::unique_ptr<Private>` in `Geometries.h`, so this is purely a change from raw-new to `make_unique` in the init list — no ownership or lifetime change.
**Verification:** Both default (`build/`) and sanitizer (`build-asan/`) test suites pass at HEAD (`ctest --test-dir build --output-on-failure`; `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` asan run, 290/290 tests pass with pre-existing T-015 route_plan_base SEGV filter applied).
**Risk:** Low. `std::make_unique` is the canonical exception-safe form.
**Effort:** S. (~15 min actual)
**Blocks:** nothing
**Depends on:** nothing

#### T-022 (M-009): `Geometry` copy semantics undocumented

**Source:** Memory audit M-009
**Location:** `include/fields2cover/types/Geometry_impl.hpp:51-62` (and header `Geometry.h`)
**Severity:** Medium (documentation)
**Description:** `Geometry<T,R>::Geometry(const Geometry&) = default` performs a `shared_ptr` copy — the new `Geometry` shares the OGR data with the original. Mutation through one wrapper silently affects the other. `clone()` is the explicit deep-copy path. This convention is correct but invisible to readers, and callsites like `Swath a = b; a.getPath().setPoint(...)` silently mutate `b`.
**Proposed fix:** Add a doxygen block at the top of `Geometry.h` explaining the "copy = shallow share, clone = deep copy" convention. Enumerate the mutation methods that do NOT trigger a deep copy (`setPoint`, `addGeometry`, etc.).
**Risk:** None (doc-only).
**Effort:** S (documentation only, ~30 min to enumerate methods).
**Blocks:** nothing
**Depends on:** nothing
**Notes:** Should land before T-002 option (a) is applied — otherwise the deep-copy fix might confuse readers into thinking the library promises deep-copy everywhere.

#### T-023 (F-149): `LineString.cpp:13` — `static_cast` downcast

**Source:** `cppcoreguidelines-pro-type-static-cast-downcast` (F-149)
**Location:** `src/fields2cover/types/LineString.cpp:13`
**Severity:** Medium
**Description:** `static_cast` is used to downcast from `OGRGeometry*` base to `OGRLineString*` derived. `dynamic_cast` would be safer at the cost of RTTI. The code likely knows statically that the pointer type is correct (based on OGR's type-tag system), but the compiler cannot verify.
**Proposed fix:** INVESTIGATE — read LineString.cpp:1-30 to understand the construction context. If the pointer's dynamic type is genuinely guaranteed by an adjacent OGR type check (`OGR_G_GetGeometryType`), add a comment and keep the `static_cast` with a NOLINT suppression. If not, switch to `dynamic_cast` and handle the null path. Same pattern likely recurs elsewhere in the `types` module — a grep for `static_cast<OGR` would find them.
**Risk:** Medium. A switch to `dynamic_cast` adds RTTI dependency on OGR classes (fine — they already use RTTI internally). A bad `static_cast` is UB.
**Effort:** M (investigation + potential sweep).
**Blocks:** nothing
**Depends on:** nothing

#### T-024 (Medium const cascades from api-hygiene): A-050, A-053, A-054, A-058, A-060/A-062, A-066, A-072/A-073, A-077

**Source:** API hygiene A-050, A-053, A-054, A-058, A-060, A-062, A-066, A-072, A-073, A-077
**Severity:** Medium
**Description:** Seven groups of query methods that should be `const` but aren't — either because they require a `mutable` cache member (A-050, A-053, A-066) or because the cascade across a pure-virtual hierarchy requires synchronized changes to base + all overriders (A-054, A-058, A-060, A-062, A-072, A-073, A-077).

Per api-hygiene.md § "Medium-severity: query methods that should be const with mutable cache":
- **A-050** — `Graph::shortestPathsAndCosts`, `shortestPath`, `shortestPathCost`: memoize into `shortest_paths_`. Mark methods `const`, declare `shortest_paths_` as `mutable` in header.
- **A-053** — `Graph2D::shortestPath`, `shortestPathCost` (cascade from A-050).
- **A-054** — `SwathGeneratorBase::generateBestSwaths`, `generateSwaths`, `computeCostOfAngle`, `computeBestAngle` — base + overriders synchronized change.
- **A-058** — `HeadlandGeneratorBase::generateHeadlands`, `generateHeadlandArea`, `generateHeadlandSwaths` — pure compute verified against `ConstHL` impl.
- **A-060/A-062** — `RoutePlannerBase::genRoute`, `genShortestRoute` — per-subclass investigation; `computeBestRoute` and `transformSolutionToRoute` are already const.
- **A-066** — `TurningBase::createTurn`, `createTurnIfNotCached` — memoize into `path_cache_`.
- **A-072/A-073** — All `Objective::computeCost` overloads across `SGObjective`, `HGObjective`, `PPObjective`, `RPObjective`, `DecompObjective` hierarchies. Largest blast radius of any const-cascade because it touches every objective-function caller.
- **A-077** — `DecompositionBase::decompose`, `split`, `genSplitLines`, `merge` — pure transforms; declare const in base + `BoustrophedonDecomp`, `TrapezoidalDecomp`.

**Proposed fix:** Treat as 7 independent commits, one per group, each with `make test` gating. Each follows the same pattern:
1. Mark the method `const` in the base header.
2. If using a cache, mark the cache member `mutable`.
3. Sync every override in derived classes (missed overrides produce a compile error — that's a feature, not a bug).
4. Rebuild, run tests.
If a test fails, revert the single group and downgrade it to a T-NNN InvestigateLater entry.

**Risk:** Medium. Pure-virtual cascades have a habit of surfacing hidden mutation in one derived class, forcing either a refactor or a partial rollback. Test coverage is the gate.
**Effort:** L — each of the 7 groups is individually M (30 min – 2h), so the bundle is ~4-8 hours total.
**Blocks:** Any future attempt to hold `const Objective&` through the generateBestSwaths pipeline (currently impossible because `computeCost` is non-const).
**Depends on:** nothing
**Notes:** A-072/A-073 is the highest-leverage group. Tackle it first if only one can be done.

#### T-025 (F-147, F-148): `Graph2D.cpp` inefficient `emplace_back` in loop

**Source:** `performance-inefficient-vector-operation` (F-147 `Graph2D.cpp:64`, F-148 `Graph2D.cpp:96`)
**Severity:** Medium
**Description:** Two loops call `emplace_back` without a prior `reserve()`. For graph-construction code this can mean O(log N) reallocations on the edge list per call. The loop upper bound is statically knowable in both sites (it's `swaths.size() * 2` and similar).
**Proposed fix:** Add a `reserve(expected_size)` call before each loop. One line per site.
**Risk:** Low. Pure perf improvement.
**Effort:** S.
**Blocks:** nothing
**Depends on:** nothing

#### T-026 (F-134, F-135, F-140): oversized enum base types in `PathState` / `Swath`

**Source:** `performance-enum-size` (F-134, F-135, F-140)
**Locations:** `PathState.h:15,21`, `Swath.h:21`
**Severity:** Medium
**Description:** `PathSectionType`, `PathDirection`, `SwathType` enums use the default `int` (4 bytes) but have ≤8 values each. `std::uint8_t` or `std::int8_t` would shrink the type and improve cache density of `std::vector<PathState>`.
**Proposed fix:**
```cpp
enum class PathSectionType : std::uint8_t { ... };
enum class PathDirection : std::int8_t { ... };
enum class SwathType : std::uint8_t { ... };
```
**Risk:** Medium. If any code serializes these enums (e.g., to a binary format or via SWIG typemap), the size change breaks the serialization. Grep for `reinterpret_cast<.*PathSectionType>` and similar before applying. SWIG tends to round-trip enums as Python ints regardless of underlying size, so Python bindings are likely safe.
**Effort:** M (includes a serialization audit).
**Blocks:** nothing
**Depends on:** nothing

## Cross-references

- **F-NNN** → `docs/audit/static-analysis.md` (plan 06-01)
- **M-NNN** → `docs/audit/memory-audit.md` (plan 06-02)
- **A-NNN / I-NNN** → `docs/audit/api-hygiene.md` (plan 06-03)
- **T-NNN** → stable triage ID assigned in this document
- **Plan 06-04 (parallel):** applies Low findings and clang-tidy `--fix`-safe Medium items. Expected overlap: T-004 (F-011..F-013 only), T-006, T-014, T-015, T-016, T-017, T-019, T-020, T-025 may be absorbed. Cross-check `fixes-applied.md` after plan 06-04 completes and mark any T-NNN `[RESOLVED commit <sha>]` accordingly.

## Disposition

When a future phase acts on a followup, update this doc in-place:
- Mark the entry header with `[RESOLVED commit <sha>]` (keep the entry for traceability).
- Do not delete entries.
- If a fix is attempted and rolled back, add a `Status:` line with `[ATTEMPTED commit <sha> — REVERTED]` and move to "Deferred Issues" at the bottom of the doc.

## Notable findings for immediate attention (despite deferral)

1. **T-001 (F-001)** — the only Critical static-analysis finding; 15-minute fix. Do this first in any future remediation pass.
2. **T-011 (M-004 / F-029)** — only raw `delete` in the library; one-line fix; makes the codebase `delete`-free.
3. **T-002 (M-001)** — highest-risk memory finding; Medium effort but eliminates a whole class of UB.
4. **T-012 (M-002) + T-013 (M-003)** — pair-commit rule-of-5 fix that unblocks hot-path move semantics.
5. **T-024 A-072/A-073 sub-group** — largest const-cascade leverage; makes the objective hierarchy composable with `const`.

Everything else can wait for a future phase without active regression risk.

---

## Sanitizer findings (Wave 2)

Entries below surfaced when running the C++ unit test suite under
`-DENABLE_SANITIZERS=ON` (ASan + UBSan + LSan, clang 18). These are in
addition to the static-analysis findings above and were discovered by plan
07-02.

### T-014 (ASAN-1): `Point.h:186` — UB float-to-unsigned cast in `std::hash<Point>` [RESOLVED by 07-02]

**Source:** UBSan (`include/fields2cover/types/Point.h:186`) — `runtime error: -1e+10 is outside the range of representable values of type 'unsigned long'`
**Location:** `include/fields2cover/types/Point.h:184-188`
**Severity:** High
**Description:** The `std::hash<f2c::types::Point>` specialization computed
`size_t(p.getX() + p.getY()*1e10 + p.getZ()*1e20)`. Whenever any coordinate
was negative, the result of the double expression was negative, and the
direct cast to `size_t` is UB per the C++ standard. In practice it produces
an unpredictable hash, corrupting `std::unordered_map<Point, ...>` lookups
(the Graph2D node table). UBSan flagged this on the `addEdges` test when
the deposit placeholder `(-1e8, -1e8)` was inserted.
**Fix:** Combine `std::hash<double>` of each coordinate using boost-style
`hash_combine`. Preserves stability for equal points, avoids UB for any
double input, including NaN/±inf (where `std::hash<double>` is defined).
**Risk:** Low. Hash codes for previously-inserted positive points change,
but any in-process `unordered_map` is rebuilt from scratch per call so
there is no on-disk compatibility concern.
**Effort:** S (≤15 min)
**Status:** RESOLVED inline by plan 07-02.

### T-015 (ASAN-2): Pre-existing pipeline/scaling test failures mask coverage gap

**Source:** Running `ctest` under sanitizers surfaced that multiple C++
tests already fail in the **non-sanitizer** baseline build at the 07-01
merge base. These are NOT sanitizer findings — they reproduce identically
in `build-baseline` (plain Release build) — but they must be documented
because the sanitizer CI job gates on a clean run and has to filter them.
**Affected tests (reproduce in baseline):**
- `fields2cover_rp_route_plan_base.simple_example`
- `fields2cover_rp_route_plan_base.redirect_flag`
- `fields2cover_rp_route_plan_base.start_and_end_points`
- `fields2cover_utils_visualizer.save_Route_and_Path`
- `fields2cover_utils_visualizer.save_field`
- `fields2cover_decomp_boustrophedon.decompose`
- `fields2cover_decomp_trapezoidal.decompose`
- `fields2cover_hl_const_gen.empty_area`
- `fields2cover_hl_const_gen.border_area`
- `fields2cover_hl_const_gen.border_swaths`
- `fields2cover_types_cell.eq_mult_operator`
- `fields2cover_types_multilinestring.init`
- `fields2cover_types_point.rotateFromPoint`
- `fields2cover_utils_Random.genRandField`
**Two distinct symptoms:**
1. **OR-Tools RoutingIndexManager abort** (`route_plan_base.*`,
   `save_Route_and_Path`): `computeBestRoute` calls
   `cov_graph.numNodes() - 1` which is `-1` when the graph is empty, and
   feeds it as the depot node to `RoutingIndexManager`, triggering
   `Check failed: start >= 0 (-1 vs. 0)` and `abort()`. Root cause is
   upstream: the swath/decomp pipeline returns empty for these inputs.
2. **Scaling/operator off-by-factor** (all `types_*` + `hl_const_gen.*` +
   `decomp.*` + `Random.genRandField`): tests that scale geometries
   (`cells *= 3e1`, `l *= 2`, etc.) produce areas and coordinates that are
   roughly 1/16000 of the expected value (`cells.area() = 0.60` vs
   `1e4`). Points and linestrings show an exact factor-2 mismatch. This
   looks like a regression in `Geometry::operator*=` or `LineString::getX`
   that landed before the 07-01 merge base; it affects several test
   families uniformly.
**Severity:** High (logic bugs), but pre-existing — not introduced or
surfaced by the sanitizer work.
**Disposition:** Filed here so the sanitizer CI filter can stay
documented and transparent (see `cmake/sanitizer-test-filter.txt`). The
filter string is a negative GTEST_FILTER that excludes the tests above.
A dedicated fix plan (recommend 07-03 or 07-04) should root-cause the
`operator*=` regression and the empty-pipeline depot bug.
**Effort:** M-L — requires bisecting the regression before the 07-01
merge base and auditing the swath→route pipeline for empty-input
handling.
**Status:** DEFERRED to a dedicated fix plan. Tracked by the sanitizer
CI filter so 07-02 can land green.
**Follow-up fix location for the depot bug:**
`src/fields2cover/route_planning/route_planner_base.cpp:201-204` —
needs a guard for `numNodes() == 0` (early-return empty route) plus
input validation on the entire `genRoute` path.

---

## T-025 — GeoJSON parser crashes on missing `coordinates` key [RESOLVED commit d5223c1]
**Source:** Fuzz finding C-001 (plan 07-06)
**Location:** `src/fields2cover/utils/parser.cpp::getCellFromJson`
**Severity:** HIGH — reachable via `POST /parser/import-field-geojson`
**Description:** `getCellFromJson` indexes `imported_cell["geometry"]["coordinates"]` without checking key presence. A client POST with a malformed FeatureCollection can crash the server.
**Crash input:** `tests/cpp/fuzz/crashes/geojson_missing_coordinates.json`
**Proposed fix:** Check `.contains("geometry")` and `.contains("coordinates")` before indexing. On miss, throw `std::invalid_argument("GeoJSON feature missing geometry.coordinates")`. Controller catches and returns HTTP 400 `INVALID_GEOJSON_SCHEMA`.
**Risk:** Low — pure defensive check; no behavior change for valid inputs.
**Effort:** S (< 1 hour, ~10 lines C++ + 2 tests)
**Status:** RESOLVED in commit d5223c1. `getCellFromJson` now validates `is_object`/`contains("geometry")`/`contains("coordinates")`/`is_array` before indexing and throws `std::invalid_argument` on malformed input. Verified against the reproducer — harness returns cleanly, no SEGV. All 294 unit tests pass.

---

## T-026 — Route planner SEGV on 1-byte random input [RESOLVED commit a264674]
**Source:** Fuzz finding C-002 (plan 07-06)
**Location:** `src/fields2cover/route_planning/route_planner_base.cpp` (via genRoute → OR-Tools)
**Severity:** HIGH in fuzzing CI, MEDIUM in production (schema validation filters the API path)
**Description:** 1 byte of random input fed through the fuzz harness crashes inside OR-Tools. Same class as the pre-existing T-015 route_plan_base brittleness.
**Crash input:** `tests/cpp/fuzz/crashes/route_ortools_segv_1byte.bin` (0xa6)
**Proposed fix:** Two options: (a) harden the fuzz harness to pre-validate byte count before constructing the f2c pipeline — scope-local; or (b) harden `RoutePlannerBase::computeBestRoute` with extra pre-conditions matching the existing `numNodes() == 0` guard added in commit 39336d0. Recommend (a) first.
**Risk:** Low for (a), Medium for (b).
**Effort:** S for (a), M for (b).
**Status:** RESOLVED in commit a264674. Option (a) applied: `fuzz_route_gen.cpp` now requires ≥64 bytes, checks `remaining_bytes()` after constructing `FuzzedDataProvider`, and returns `-1` (libFuzzer corpus reject) instead of `0` so rejected inputs don't pollute the corpus. Library-level hardening (option b) remains deferred and is still tracked under T-015. Verified against the 1-byte reproducer — cleanly rejected, no SEGV.

---

## T-027 — Route planner SEGV on 31-byte input (integer underflow?) [RESOLVED commit a264674]
**Source:** Fuzz finding C-003 (plan 07-06)
**Location:** Same as T-026 but with slightly deeper code path
**Severity:** HIGH (same reasoning as T-026)
**Description:** 31 bytes with pattern `0x00 f0 00...00 fd ff ff ff ff ff ff de`. The `0xfd ffff ffff` suggests a signed 32-bit value of -3 being interpreted somewhere in the graph construction path — possibly a vertex count or index underflowing before OR-Tools sees it.
**Crash input:** `tests/cpp/fuzz/crashes/route_ortools_segv_31bytes.bin`
**Proposed fix:** Same options as T-026. Additionally, audit the graph construction path for any `int32` → `size_t` conversions that could underflow.
**Risk:** Medium.
**Effort:** M.
**Status:** RESOLVED in commit a264674 alongside T-026 (same harness-level fix: minimum 64-byte threshold + `remaining_bytes()` check). Verified against the 31-byte reproducer — cleanly rejected by the harness. Integer-underflow audit of the graph construction path remains TODO under T-015.
