# HRD-04 raw pointer sweep

**Generated:** 2026-04-11
**Phase:** 07-f2c-performance-optimization
**Plan:** 07-04

Tracks the library-wide sweep of raw owning pointers in `src/fields2cover/`
and `include/fields2cover/`. Per the Phase 6 memory audit
(`docs/audit/memory-audit.md`), only one raw `delete` existed (already
resolved in Plan 07-01 via T-011) and only a handful of raw `new` sites
remained. This is a short pass; the primary target is T-021 (PImpl
construction in `Geometries_impl.hpp`).

## Before

### Raw `new` sites (potential owning allocations)

Searched `src/fields2cover/` and `include/fields2cover/` for `\bnew\b[^a-zA-Z_0-9]`,
filtering out matches in comments/documentation:

```
include/fields2cover/types/Geometries_impl.hpp:68:      m_poPrivate(new Private()) {
include/fields2cover/types/Geometries_impl.hpp:113:        int nPos) : m_poPrivate(new Private()) {
src/fields2cover/types/Path.cpp:277:  ss.imbue(std::locale(std::locale(), new DecimalSeparator<char>('.')));
src/fields2cover/utils/transformation.cpp:246:      new OGRSpatialReference(), [](OGRSpatialReference* ref) {
```

Total: **4 raw `new` expressions**.

### Raw pointer member declarations (`T* m_foo;`)

Searched both include/ and src/ for member patterns (e.g.
`^\s+[A-Za-z_][A-Za-z0-9_:<>,]*\*\s+[a-zA-Z_][a-zA-Z0-9_]*_\s*(;|=)`):

```
(none)
```

Total: **0 raw pointer member declarations**.

### Raw pointer function returns

Non-OGR returns owning a `T*` would appear as top-level signatures
`^\s*T\*\s+foo(`. None found outside the OGR C-API boundary (which is
out of scope — OGR handle returns are governed by
`OGRGeometryFactory`/`destroyGeometry` discipline, already audited in
plan 07-01).

## Classification

| # | Site | Line | Category | Action |
|---|---|---|---|---|
| 1 | `include/fields2cover/types/Geometries_impl.hpp` | 68  | **OWNING**    | Task 2: `std::make_unique<Private>()` (T-021) |
| 2 | `include/fields2cover/types/Geometries_impl.hpp` | 113 | **OWNING**    | Task 2: `std::make_unique<Private>()` (T-021) |
| 3 | `src/fields2cover/types/Path.cpp`                | 277 | **API SHAPE** | `std::locale` ctor demands a raw `Facet*` and *takes ownership* per the standard library contract (see `[locale.cons]`). No smart-pointer form exists. Annotate only. |
| 4 | `src/fields2cover/utils/transformation.cpp`      | 246 | **ALREADY SMART** | Construction is already inside a `std::unique_ptr<OGRSpatialReference, void(*)(...)>` ctor with custom deleter. `make_unique` cannot take a deleter, so the raw-`new` expression is structurally required. Annotate only. |

**OWNING sites to convert:** 2 (both T-021).
**VIEW sites:** 0 — no raw pointer members exist.
**API SHAPE / ALREADY SMART (leave, annotate):** 2.

This matches the Phase 6 memory-audit characterization: "only **one**
manual `new` and **one** manual `delete` exist in the entire library"
plus the two PImpl construction sites flagged in M-008/T-021. Total of
**4 raw-`new` expressions**; **2** are OWNING and reachable for
conversion.

## After

### Raw `new` sites (post-sweep)

```
src/fields2cover/types/Path.cpp:277:  ss.imbue(std::locale(std::locale(), new DecimalSeparator<char>('.')));
src/fields2cover/utils/transformation.cpp:246:      new OGRSpatialReference(), [](OGRSpatialReference* ref) {
```

Both remaining sites are structurally required by standard-library /
custom-deleter API shape and cannot be converted to `make_unique` /
`make_shared`:

- `Path.cpp:277` — `std::locale` ctor takes ownership of a facet via
  raw `Facet*` per the standard (`[locale.cons]`). No smart-pointer
  overload exists. Classified **API SHAPE**.
- `transformation.cpp:246` — construction is the argument to
  `std::unique_ptr<OGRSpatialReference, void(*)(OGRSpatialReference*)>`
  with a custom `DestroySpatialReference` deleter. `std::make_unique`
  does not accept a deleter, and `OGRSpatialReference::DestroySpatialReference`
  must be used (not plain `delete`) to match OGR's allocator. Classified
  **ALREADY SMART**.

### Counts

| Metric                         | Before | After | Delta |
|--------------------------------|--------|-------|-------|
| Raw `new` in library           | 4      | 2     | -2    |
| Raw `new` (OWNING only)        | 2      | 0     | -2    |
| Raw pointer members (OWNING)   | 0      | 0     | 0     |
| Raw `delete` in library        | 0      | 0     | 0     |

**HRD-04 outcome:** every raw owning pointer in `src/fields2cover/` and
`include/fields2cover/` is now either inside a smart-pointer ctor or
annotated as non-owning. The library has **zero** raw `delete` (since
Plan 07-01 T-011) and **zero** raw owning `new` (since this plan's
T-021 conversion). The remaining two `new` expressions are classified
and documented above.

### Verification

Per-commit test gate on both builds:

- `ctest --test-dir build --output-on-failure` — 100% pass (1/1 ctest
  suite, full gtest suite green).
- `GTEST_FILTER=<T-015 filter> ASAN_OPTIONS=detect_leaks=1:halt_on_error=1
  build-asan/tests/unittests` — 290/290 pass.

The `T-015` filter excludes 4 pre-existing OR-Tools `RoutingIndexManager`
SEGVs in `fields2cover_rp_route_plan_base.*` and
`fields2cover_utils_visualizer.save_Route_and_Path`, documented as
deferred in `docs/audit/followups.md:505-553`. These failures reproduce
at the 07-04 merge base (commit `ad5d1b2`) and are unrelated to the
`make_unique` conversion in this plan.

### T-021 status

RESOLVED. See `docs/audit/followups.md` § T-021.
