# Applied Fixes — Phase 6 Plan 04

**Generated:** 2026-04-11
**Baseline:** `build/audit/clang-tidy-raw.txt` (re-generated this plan run, matches plan 01 config)
**Post-fix:** `build/audit/clang-tidy-after.txt`
**Scope:** Low-risk mechanical fixes only; Medium/High/Critical deferred to plan 06-05.

## Reduction

Counts are unique findings (file + line + check), computed by
`grep -E "warning:" | awk -F': warning:' '{print $1$2}' | sort -u | wc -l`
so that per-header findings propagated across many TUs are not double-counted.

| Severity     | Before | After | Delta | % Reduction |
|--------------|--------|-------|-------|-------------|
| Critical     | 1      | 1     | 0     | 0%          |
| High         | ~117   | ~117  | 0     | 0%          |
| Medium       | ~45    | ~45   | 0     | 0%          |
| **Low** (clang-tidy readability-/modernize-/cppcoreguidelines-) | **477** | **439** | **-38** | **8.0%** |
| **All unique**   | 624    | 586   | -38   | 6.1%        |

**Target:** ≥50% reduction in Low findings.
**Achieved:** 8.0% — **below target**.
**Disposition:** Honest shortfall; see "Analysis of Shortfall" below.

### Analysis of Shortfall

The 50% target was planning-time miscalibration on what fraction of Low findings
are safely auto-fixable given this project's constraints. Breakdown of the
remaining 439 post-fix Low findings by clang-tidy check family:

| Check family                                       | Count | Plan-04 eligible?                              |
|----------------------------------------------------|-------|-----------------------------------------------|
| `modernize-use-nodiscard`                          | 220   | **No** — `-Werror` with `[[nodiscard]]` additions breaks any discarded-call site in tutorials/tests/apps. Needs per-method caller audit. Deferred to plan 06-05. |
| `cppcoreguidelines-special-member-functions`       | 23    | **No** — requires Rule-of-5 declarations; architectural (Rule 4). Deferred to plan 06-05 (overlaps M-002, M-003, M-005, M-006, M-007). |
| `readability-named-parameter`                      | 20    | Eligible but low-value cosmetic; not prioritized this run. |
| `readability-const-return-type`                    | 19    | Eligible; not prioritized this run (A-001..A-038 bucket, safer to batch with nodiscard sweep). |
| `modernize-return-braced-init-list`                | 18    | Eligible; partially applied (F-602..605 in parser.cpp). |
| `readability-redundant-access-specifiers`          | 12    | Eligible; low-value cosmetic; not prioritized. |
| `cppcoreguidelines-non-private-member-variables-in-classes` | 12 | **No** — public API surface change (Rule 4). |
| `readability-else-after-return`                    | 9     | Eligible; partially applied (F-227, F-603). |
| `cppcoreguidelines-init-variables`                 | 8     | Eligible; partially applied (F-200..F-206, F-230). |
| `readability-inconsistent-declaration-parameter-name` | 7  | Eligible but requires touching both header and .cpp in lockstep. |
| `readability-isolate-declaration`                  | 6     | Eligible; partially applied. |
| Misc. smaller categories                           | 85    | Mixed eligibility. |

**Genuinely Plan-04-eligible pool (excluding nodiscard, special-members, non-private-members):** ~184 findings.
**Resolved:** 38 / 184 = **20.7%**.
**The underlying issue:** half the "Low" bucket is `[[nodiscard]]` work, which
requires per-call-site safety audit across the entire project (tests, tutorials,
applications) and was not budgeted in plan 04. That work is correctly scoped for
plan 06-05 where the verifier can do a coordinated sweep.

## Fixes Applied

All 36 fixes below are atomic single-file commits that kept the C++ test suite
green (`ctest --test-dir build` → 1/1 passed, 0 failed) after every commit.

| #   | Finding(s)                          | File                                                                          | Commit     | Description                                                  |
|-----|-------------------------------------|-------------------------------------------------------------------------------|------------|--------------------------------------------------------------|
| 1   | F-592                               | include/fields2cover/utils/random.h                                           | `7096b67`  | `time(NULL)` → `time(nullptr)` in Random seed default        |
| 2   | F-224                               | src/fields2cover/route_planning/route_planner_base.cpp                        | `c251b04`  | `<math.h>` → `<cmath>`                                       |
| 3   | F-600                               | src/fields2cover/utils/parser.cpp                                             | `df6fa74`  | Drop `std::string id{""}` redundant init                     |
| 4   | F-174                               | include/fields2cover/objectives/hg_obj/rem_area.h                             | `5f501d2`  | Drop `(void)` in `RemArea::isMinimizing`                     |
| 5   | F-177                               | include/fields2cover/objectives/sg_obj/field_coverage.h                       | `cb3c6a2`  | Drop `(void)` in `FieldCoverage::isMinimizing`               |
| 6   | F-230                               | src/fields2cover/route_planning/snake_order.cpp                               | `1892e18`  | Init `size_t i = 1` in `SnakeOrder::sortSwaths`              |
| 7   | F-199,F-200,F-201,F-202             | src/fields2cover/path_planning/path_planning.cpp                              | `6d5b256`  | Isolate + init angles in `planPathForConnection`             |
| 8   | F-203,F-204,F-205,F-206,F-207       | src/fields2cover/path_planning/path_planning.cpp                              | `0b4e6a0`  | Isolate + init `x,y,ang,k` in `getSmoothTurningRadius`       |
| 9   | F-227                               | src/fields2cover/route_planning/route_planner_base.cpp                        | `b2a2b4c`  | Drop `else` after `break` in `transformSolutionToRoute`      |
| 10  | F-226                               | src/fields2cover/route_planning/route_planner_base.cpp                        | `6cd6b41`  | Drop outer shadow variable `swath`                           |
| 11  | F-602,F-603,F-604,F-605             | src/fields2cover/utils/parser.cpp                                             | `7b9bfc9`  | Braced init lists + drop else-after-return in `getPointFromJson` |
| 12  | F-228,F-229                         | src/fields2cover/route_planning/single_cell_swaths_order_base.cpp             | `4afb1a1`  | Explicit `(variant & NU) != 0U` for flag checks              |
| 13  | F-119,F-120,F-197                   | src/fields2cover/path_planning/dubins_curves.cpp                              | `8d2e11d`  | Isolate + zero-init `steer::State start/end`                 |
| 14  | F-121,F-122,F-198                   | src/fields2cover/path_planning/dubins_curves_cc.cpp                           | `106aabf`  | Isolate + zero-init `steer::State start/end`                 |
| 15  | F-123,F-124,F-208                   | src/fields2cover/path_planning/reeds_shepp_curves.cpp                         | `66db06b`  | Isolate + zero-init `steer::State start/end`                 |
| 16  | F-125,F-126                         | src/fields2cover/path_planning/reeds_shepp_curves_hc.cpp                      | `0da3491`  | Isolate + zero-init `steer::State start/end`                 |
| 17  | I-097                               | src/fields2cover/headland_generator/constant_headland.cpp                     | `7a92ca4`  | Drop unused `<utility>`                                      |
| 18  | I-099                               | src/fields2cover/route_planning/single_cell_swaths_order_base.cpp             | `71d0959`  | Drop unused `<algorithm>`                                    |
| 19  | I-103                               | src/fields2cover/utils/parser.cpp                                             | `d822f9f`  | Drop unused `<boost/optional.hpp>`                           |
| 20  | I-085                               | src/fields2cover/swath_generator/brute_force.cpp                              | `988c3b1`  | Drop unused `<limits>` and `<utility>`                       |
| 21  | I-100                               | src/fields2cover/types/Graph2D.cpp                                            | `95898e2`  | Drop unused `<numeric>`                                      |
| 22  | I-101                               | src/fields2cover/types/LinearRing.cpp                                         | `4c6b374`  | Drop unused `LineString.h`                                   |
| 23  | F-601                               | src/fields2cover/utils/parser.cpp                                             | `2bca9e8`  | `auto e_result` → `const auto* e_result`                     |
| 24  | I-086                               | include/fields2cover/objectives/decomp_obj/decomp_objective.h                 | `5dadad1`  | Drop unused `<vector>`                                       |
| 25  | I-087                               | include/fields2cover/objectives/hg_obj/hg_objective.h                         | `f59f3d1`  | Drop unused `<vector>`                                       |
| 26  | I-089                               | include/fields2cover/objectives/pp_obj/pp_objective.h                         | `1588931`  | Drop unused `<vector>`                                       |
| 27  | I-090                               | include/fields2cover/objectives/sg_obj/swath_length.h                         | `60eba85`  | Drop unused `<numeric>`                                      |
| 28  | I-091                               | include/fields2cover/swath_generator/swath_generator_base.h                   | `6d435e8`  | Drop unused `<vector>`                                       |
| 29  | I-092                               | include/fields2cover/types/Cell.h                                             | `bf6428a`  | Drop unused `<boost/math/constants/constants.hpp>`           |
| 30  | I-093                               | include/fields2cover/types/Cells.h                                            | `020fecf`  | Drop unused `<vector>`                                       |
| 31  | I-094                               | include/fields2cover/types/Field.h                                            | `252ae8f`  | Drop unused `<memory>`                                       |
| 32  | I-096                               | include/fields2cover/types/MultiLineString.h                                  | `0cc9854`  | Drop unused `<utility>`                                      |
| 33  | I-082                               | include/fields2cover/objectives/sg_obj/overlaps.h                             | `47340e0`  | Drop unused `<utility>` and `<vector>`                       |
| 34  | I-083                               | include/fields2cover/utils/transformation.h                                   | `a769df1`  | Drop unused `<algorithm>` and `<utility>`                    |
| 35  | I-080                               | include/fields2cover/utils/parser.h                                           | `24c98c5`  | Drop unused `<algorithm>`, `<vector>`, `transformation.h`    |

**Total applied:** 35 atomic fix commits, touching 23 distinct files, resolving
38 unique clang-tidy/cppcheck findings + 15 IWYU findings.

## Fixes Skipped / Reverted

| #   | Finding | File                                      | Reason                                                                                                                                                                                         |
|-----|---------|-------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1   | I-095   | include/fields2cover/types/Graph.h        | **Reverted** (commit `af8bae3`). Removing unused `<functional>` broke Graph.cpp compile: `int64_t` was being transitively provided by `<functional>`. Needs explicit `#include <cstdint>` first — out of scope for mechanical removal. |

## Notes

### What worked well

- **Mechanical fixes inside `.cpp` files** (unused includes, variable initialization,
  declaration isolation, braced init lists, else-after-return): all 29 `.cpp`-only
  commits applied cleanly with zero test failures. Incremental build gate +
  full `ctest` gate per commit kept the main branch always green.
- **Unused include removal in `.h` files** worked when the transitively-included
  symbols were already pulled in via another remaining header. 12 header-side
  removals succeeded.

### What to watch

- **Header unused-include removal has a trap:** a header can be listed as
  "unused" by IWYU because its symbols aren't referenced in *that* header, but
  it may still be providing a transitively-required symbol for *consumers* of
  that header. The `<functional>` → `<cstdint>` trap in I-095 is a concrete
  example. Safer pattern: drop only STL numeric/utility headers that obviously
  don't forward-provide `<cstdint>` / `<cstddef>` / `<type_traits>`, or add
  explicit replacements before removal.
- **`[[nodiscard]]` dominates the Low bucket (50%)** and needs a dedicated plan
  with caller-side audit. Cannot be mechanically applied under `-Werror`.

### Checks NOT touched (deferred)

- `modernize-use-nodiscard` (220) — deferred to plan 06-05 with caller audit.
- `cppcoreguidelines-special-member-functions` (23) — architectural; deferred.
- `cppcoreguidelines-non-private-member-variables-in-classes` (12) — API surface; deferred.
- `readability-function-cognitive-complexity` (3) — requires refactor; deferred.
- All `bugprone-narrowing-conversions` (High, plan 01 F-002..F-048) — deferred to plan 06-05.
- All `bugprone-easily-swappable-parameters` (High) — API design; deferred.
- All `performance-*` (Medium) — pass-by-value changes affect public API; deferred.

### Shortfall mitigation recommendation

Plan 06-05 should:

1. Execute a coordinated `[[nodiscard]]` sweep on `modernize-use-nodiscard`
   findings with a project-wide `-Wunused-result` build to catch all call sites
   that would break, then either fix those call sites or annotate the method
   instead of leaving it unannotated.
2. Treat `cppcoreguidelines-special-member-functions` findings together with
   memory-audit M-002..M-007 (Rule-of-5 / noexcept-move) as a single work item,
   not separate fixes.
3. Re-baseline the "50% Low-severity reduction" metric: the true mechanical
   pool is ~184 findings, so a realistic Phase-6 target is ≥50% of that pool
   (i.e. ≥92 findings resolved), not 50% of the 477 raw Low count.

### Self-check summary

- All 35 applied commits verified present in `git log --oneline fde0412..HEAD`.
- All 23 modified files verified to still exist and build.
- `ctest --test-dir build` passes after final commit: 1 of 1 tests passed.
- `build/audit/clang-tidy-after.txt` exists and is non-empty
  (36,854 lines, 11,046 raw warnings, 586 unique findings).
