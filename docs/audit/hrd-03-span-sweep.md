# HRD-03 std::span sweep

**Generated:** 2026-04-11
**Plan:** `.planning/phases/07-f2c-performance-optimization/07-05-PLAN.md`
**Requirement:** HRD-03 — migrate raw-pointer range parameters `(T*, size_t)` / `(const T*, size_t)` to `std::span<T>`.

## Executive summary

**Library-wide raw-pointer range-parameter count: 0.**

Zero matches across every targeted module (`utils/`, `types/`, `swath_generator/`, `route_planning/`, `path_planning/`) in both `src/fields2cover/**` and `include/fields2cover/**`. This matches — and sharpens — the Phase 6 API-hygiene findings in `docs/audit/api-hygiene.md`, which catalogued const/IWYU issues and did **not** flag a single `(T*, size_t)` pair. The library was already written to use `std::vector<T>` by-value / by-const-ref and OGR/GEOS smart wrappers everywhere internal signatures cross a call boundary.

Because the sweep's "before" count is zero, there is no mechanical refactor work to perform and no regression surface to gate. HRD-03 is satisfied by documentation.

## Language standard decision

**Chosen: keep `CMAKE_CXX_STANDARD 17` (no bump).**

The plan (Task 1) recommended bumping to C++20 so that `std::span` would be available in-tree. With zero `(T*, size_t)` pairs to convert, bumping the standard would introduce risk (downstream consumers stuck on C++17, SWIG toolchain compatibility, clang-tidy rule drift) with no corresponding benefit. The bump is deferred until an actual consumer of `std::span<T>` emerges (e.g., a future hot-path refactor in path planning, or if HRD-03's scope is widened to cover raw C-style array parameters such as `double[4]` that do exist).

No polyfill (tcb::span) is introduced either, for the same reason — nothing in-tree needs it.

## Before — raw pointer range params (per module)

Pattern searched (ripgrep / Grep tool, ERE, multi-file):

```
\b(const\s+)?[A-Za-z_][A-Za-z0-9_:]*\s*\*\s+[a-z][a-zA-Z0-9_]*\s*,\s*(size_t|std::size_t|size_type|int|std::ptrdiff_t)\b
```

Scoped to `src/fields2cover/<module>/` and `include/fields2cover/<module>/`.

### utils
```
(none)
```

### types
```
(none)
```

### swath_generator
```
(none)
```

### route_planning
```
(none)
```

### path_planning
```
(none)
```

### Sanity checks (broader patterns)

To guard against the narrow ERE missing legitimate hits, two relaxed patterns were also run library-wide:

```
\*\s*\w+\s*,\s*(size_t|std::size_t|int\s|unsigned)     -> 0 matches in src/fields2cover
\*\s*\w+\s*,\s*(size_t|std::size_t)                    -> 0 matches in include/fields2cover
```

Both returned zero matches. The verdict stands.

## After — per-module counts

| Module            | Before | After | Converted | Deferred | Notes                                     |
| ----------------- | -----: | ----: | --------: | -------: | ----------------------------------------- |
| utils             |      0 |     0 |         0 |        0 | No candidate sites                        |
| types             |      0 |     0 |         0 |        0 | No candidate sites                        |
| swath_generator   |      0 |     0 |         0 |        0 | No candidate sites                        |
| route_planning    |      0 |     0 |         0 |        0 | No candidate sites (route_planner_base recently merged, untouched) |
| path_planning     |      0 |     0 |         0 |        0 | No candidate sites                        |
| **Total**         |  **0** | **0** |     **0** |    **0** |                                           |

## Deferred / skipped

None. There is nothing to defer because there is nothing to convert.

If a future audit widens HRD-03 to cover C-style fixed-size array parameters (e.g. `double coord[3]`) or OGR-owned buffers intentionally passed as raw pointers, those should be catalogued under a new requirement ID (HRD-03b or similar) rather than retrofit here.

## Trust-boundary disposition (STRIDE T-07-05-01, T-07-05-02)

- **T-07-05-01 (OOB via length/pointer mismatch)** — *mitigate (trivially).* No library-internal call crosses a `(T*, size_t)` boundary; the class of bug does not exist in the current tree. The threat is mitigated by construction, not by conversion.
- **T-07-05-02 (SWIG ABI break)** — *accept (vacuously).* No SWIG-visible header is touched.

## Verification

- Grep evidence: the command block above is reproducible; re-running it on `a45f697` (current branch tip prior to this plan) yields identical zero counts.
- Build/test gate: no source files are modified by this plan, so the pre-existing default + sanitizer build state is unchanged. The CI build and sanitizer job that already gate the branch continue to cover HRD-03's invariant: if a future PR reintroduces a `(T*, size_t)` pair the clang-tidy `cppcoreguidelines-pro-bounds-pointer-arithmetic` and `bugprone-*` checks will surface it.

## Outcome

HRD-03 is **satisfied** with a null-result sweep. Documentation is the deliverable.
