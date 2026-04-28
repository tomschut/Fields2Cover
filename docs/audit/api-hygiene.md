# API Hygiene Audit — Phase 6 Plan 03

**Generated:** 2026-04-11
**Tools:** include-what-you-use 0.21 (Ubuntu clang 17.0.6); manual const-correctness review
**Scope:** `include/fields2cover/**/*.h`, `src/fields2cover/**/*.cpp`
**Raw artifacts (gitignored):** `build/audit/iwyu-raw.txt`, `build/audit/const-review.md`

## Method

1. **IWYU** was executed per-TU by extracting each compile command from `build/compile_commands.json` (generated via `cmake -DCMAKE_EXPORT_COMPILE_COMMANDS=ON`) and replacing the compiler binary with `include-what-you-use -Xiwyu --no_fwd_decls`. 54 translation units analyzed. Raw output: `build/audit/iwyu-raw.txt` (2140 lines).
2. **Const-correctness** was reviewed manually across every public header in `include/fields2cover/`, with implementation-file cross-checks where constness was ambiguous (e.g., memoizing query methods such as `Graph::shortestPath`).
3. Findings use two disjoint ID namespaces:
   - **I-NNN** — include-what-you-use findings (this plan)
   - **A-NNN** — api-hygiene manual findings (this plan)
   These do **not** collide with `F-NNN` (plan 01 static analysis), `M-NNN` (plan 02 memory audit), or `T-NNN` (future plans).

## Summary

| Category                                             | Count | Severity   |
| ---------------------------------------------------- | ----- | ---------- |
| IWYU — files with missing direct includes            |    69 | Low        |
| IWYU — individual missing #include/fwd-decl lines    |   281 | Low        |
| IWYU — files with unused includes                    |    35 | Low        |
| IWYU — individual unused #include lines              |    76 | Low        |
| Const: `const T` return-by-value                     |    21 | Low        |
| Const: missing `[[nodiscard]]` (grouped)             |  ~40 groups (~200 methods) | Low |
| Const: missing `const` on query methods (mutable-cache candidates) | 7 | Medium |
| Const: non-const virtual where subclass is pure-compute | 5   | Medium     |
| Cosmetic: unnamed parameter / `(void)` signature     |     4 | Low        |
| Non-const overload with no mutation purpose          |     7 | Low        |

**Total findings:** 104 IWYU (I-001..I-104, per-file aggregated) + 83 const-correctness (A-001..A-083).

### Hottest headers (top offenders)

| File                                               | Missing #includes | Unused #includes | const issues |
| -------------------------------------------------- | ----------------- | ---------------- | ------------ |
| `include/fields2cover/types/Robot.h`               | 0                 | 9                | 2 groups     |
| `include/fields2cover/route_planning/route_planner_base.h` | 1       | 6                | 2 (medium)   |
| `include/fields2cover/types/Cells.h`               | 4                 | 1                | 7            |
| `include/fields2cover/types/Cell.h`                | 4                 | 1                | 4            |
| `include/fields2cover/utils/random.h`              | 2                 | 5                | 0 (legit)    |
| `include/fields2cover/types/Swath.h`               | 2                 | 4                | 3            |
| `include/fields2cover/types/Path.h`                | 1                 | 3                | 3            |
| `include/fields2cover/path_planning/turning_base.h`| 0                 | 4                | 1 (medium)   |
| `src/fields2cover/utils/transformation.cpp`        | 19                | 0                | n/a          |
| `src/fields2cover/utils/visualizer.cpp`            | 17                | 1                | n/a          |
| `src/fields2cover/route_planning/route_planner_base.cpp` | 14          | 2                | n/a          |

## Findings

### Include-What-You-Use (I-001..I-104)

Each I-NNN entry aggregates all missing or unused includes for a single translation unit. Full per-line detail is available in the raw artifact `build/audit/iwyu-raw.txt` (not committed — gitignored under `build/`). The tables below are authoritative for triage and fix planning.

#### Missing direct includes / forward decls — per-file

Each I-NNN covers all missing items in a single translation unit. Full details in `build/audit/iwyu-raw.txt`.

| ID | File | # missing | Top items |
|---|---|---|---|
| I-001 | `src/fields2cover/utils/transformation.cpp` | 19 | <math.h>, <ogr_core.h>, <ogr_geometry.h>, +16 more |
| I-002 | `src/fields2cover/utils/visualizer.cpp` | 17 | <algorithm>, <array>, <memory>, +14 more |
| I-003 | `src/fields2cover/route_planning/route_planner_base.cpp` | 14 | <google/protobuf/duration.pb.h>, <ortools/constraint_solver/constraint_solver.h>, <ortools/constraint_solver/routing_parameters.pb.h>, +11 more |
| I-004 | `src/fields2cover/utils/parser.cpp` | 12 | <ogr_core.h>, <ogr_geometry.h>, <ogr_spatialref.h>, +9 more |
| I-005 | `src/fields2cover/types/Cell.cpp` | 10 | <math.h>, <ogr_geometry.h>, <algorithm>, +7 more |
| I-006 | `src/fields2cover/objectives/rp_obj/rp_objective.cpp` | 8 | <ogr_geometry.h>, <cstddef>, <memory>, +5 more |
| I-007 | `src/fields2cover/path_planning/path_planning.cpp` | 8 | <math.h>, <algorithm>, <cstddef>, +5 more |
| I-008 | `src/fields2cover/path_planning/turning_base.cpp` | 8 | <math.h>, <algorithm>, <boost/math/constants/constants.hpp>, +5 more |
| I-009 | `src/fields2cover/types/Cells.cpp` | 8 | <math.h>, <ogr_geometry.h>, <memory>, +5 more |
| I-010 | `include/fields2cover/types/MultiLineString.h` | 7 | <ogr_core.h>, <cstddef>, <initializer_list>, +4 more |
| I-011 | `src/fields2cover/swath_generator/swath_generator_base.cpp` | 7 | <ogr_core.h>, "fields2cover/types/Cell.h", "fields2cover/types/Geometries_impl.hpp", +4 more |
| I-012 | `src/fields2cover/types/LineString.cpp` | 7 | <ogr_geometry.h>, <algorithm>, <memory>, +4 more |
| I-013 | `src/fields2cover/types/LinearRing.cpp` | 7 | <ogr_geometry.h>, <algorithm>, <memory>, +4 more |
| I-014 | `src/fields2cover/types/MultiPoint.cpp` | 7 | <ogr_geometry.h>, <memory>, <stdexcept>, +4 more |
| I-015 | `src/fields2cover/utils/random.cpp` | 7 | <assert.h>, <math.h>, <boost/math/constants/constants.hpp>, +4 more |
| I-016 | `src/fields2cover/objectives/sg_obj/n_swath_modified.cpp` | 6 | <math.h>, <cstddef>, "fields2cover/types/Geometries_impl.hpp", +3 more |
| I-017 | `src/fields2cover/objectives/sg_obj/overlaps.cpp` | 6 | <ogr_core.h>, <vector>, "fields2cover/types/Cells.h", +3 more |
| I-018 | `src/fields2cover/types/Path.cpp` | 6 | <algorithm>, <cmath>, <fstream>, +3 more |
| I-019 | `src/fields2cover/types/Swath.cpp` | 6 | <ogr_geometry.h>, <memory>, <stdexcept>, +3 more |
| I-020 | `src/fields2cover/decomposition/boustrophedon_decomp.cpp` | 5 | <math.h>, "fields2cover/types/Cell.h", "fields2cover/types/Geometries_impl.hpp", +2 more |
| I-021 | `src/fields2cover/route_planning/custom_order.cpp` | 5 | <algorithm>, <stdexcept>, <string>, +2 more |
| I-022 | `src/fields2cover/types/Field.cpp` | 5 | <ogr_core.h>, <memory>, "fields2cover/types/Geometries_impl.hpp", +2 more |
| I-023 | `include/fields2cover/types/Cell.h` | 4 | <ogr_core.h>, <cstddef>, <sstream>, +1 more |
| I-024 | `include/fields2cover/types/Cells.h` | 4 | <ogr_core.h>, <cstddef>, "fields2cover/types/LinearRing.h", +1 more |
| I-025 | `include/fields2cover/types/Point.h` | 4 | <ogr_core.h>, <cstddef>, <string_view>, +1 more |
| I-026 | `src/fields2cover/decomposition/trapezoidal_decomp.cpp` | 4 | <math.h>, "fields2cover/types/Cell.h", "fields2cover/types/Geometries_impl.hpp", +1 more |
| I-027 | `src/fields2cover/objectives/sg_obj/field_coverage.cpp` | 4 | <vector>, "fields2cover/types/Cells.h", "fields2cover/types/Geometries_impl.hpp", +1 more |
| I-028 | `src/fields2cover/types/MultiLineString.cpp` | 4 | <ogr_geometry.h>, <stdexcept>, <string>, +1 more |
| I-029 | `src/fields2cover/types/Point.cpp` | 4 | <math.h>, <ogr_geometry.h>, <algorithm>, +1 more |
| I-030 | `include/fields2cover/types/LineString.h` | 3 | <ogr_core.h>, <cstddef>, <initializer_list> |
| I-031 | `include/fields2cover/types/LinearRing.h` | 3 | <ogr_core.h>, <cstddef>, <initializer_list> |
| I-032 | `include/fields2cover/types/MultiPoint.h` | 3 | <ogr_core.h>, <cstddef>, <initializer_list> |
| I-033 | `include/fields2cover/types/Swaths.h` | 3 | <cstddef>, <initializer_list>, "fields2cover/types/Point.h" |
| I-034 | `include/fields2cover/utils/visualizer.h` | 3 | <cstddef>, <iterator>, <numeric> |
| I-035 | `src/fields2cover/objectives/sg_obj/sg_objective.cpp` | 3 | <numeric>, <vector>, "fields2cover/types/Geometries_impl.hpp" |
| I-036 | `src/fields2cover/route_planning/snake_order.cpp` | 3 | <algorithm>, <cstddef>, <vector> |
| I-037 | `src/fields2cover/types/Route.cpp` | 3 | <ogr_core.h>, <algorithm>, "fields2cover/types/Geometries_impl.hpp" |
| I-038 | `include/fields2cover/swath_generator/brute_force.h` | 2 | <boost/math/constants/constants.hpp>, "fields2cover/objectives/sg_obj/sg_objective.h" |
| I-039 | `include/fields2cover/types/Graph.h` | 2 | <stddef.h>, <stdint.h> |
| I-040 | `include/fields2cover/types/Graph2D.h` | 2 | <stdint.h>, <cstddef> |
| I-041 | `include/fields2cover/types/Route.h` | 2 | <cstddef>, "fields2cover/types/Point.h" |
| I-042 | `include/fields2cover/types/Swath.h` | 2 | <cstddef>, "fields2cover/types/Point.h" |
| I-043 | `include/fields2cover/types/SwathsByCells.h` | 2 | <cstddef>, <initializer_list> |
| I-044 | `include/fields2cover/utils/random.h` | 2 | <stdint.h>, <cstddef> |
| I-045 | `src/fields2cover/headland_generator/constant_headland.cpp` | 2 | <ogr_core.h>, "fields2cover/types/Cells.h" |
| I-046 | `src/fields2cover/objectives/hg_obj/hg_objective.cpp` | 2 | "fields2cover/types/Cells.h", "fields2cover/types/Geometries_impl.hpp" |
| I-047 | `src/fields2cover/route_planning/spiral_order.cpp` | 2 | <algorithm>, <vector> |
| I-048 | `src/fields2cover/types/Graph2D.cpp` | 2 | <utility>, "fields2cover/types/Geometry_impl.hpp" |
| I-049 | `src/fields2cover/types/Robot.cpp` | 2 | <math.h>, <stdexcept> |
| I-050 | `include/fields2cover/decomposition/boustrophedon_decomp.h` | 1 | "fields2cover/objectives/decomp_obj/decomp_objective.h" |
| I-051 | `include/fields2cover/decomposition/trapezoidal_decomp.h` | 1 | "fields2cover/objectives/decomp_obj/decomp_objective.h" |
| I-052 | `include/fields2cover/objectives/sg_obj/n_swath_modified.h` | 1 | "fields2cover/objectives/sg_obj/sg_objective.h" |
| I-053 | `include/fields2cover/route_planning/custom_order.h` | 1 | <cstddef> |
| I-054 | `include/fields2cover/route_planning/route_planner_base.h` | 1 | <stdint.h> |
| I-055 | `include/fields2cover/route_planning/single_cell_swaths_order_base.h` | 1 | <stdint.h> |
| I-056 | `include/fields2cover/route_planning/spiral_order.h` | 1 | <cstddef> |
| I-057 | `include/fields2cover/types/Path.h` | 1 | <cstddef> |
| I-058 | `include/fields2cover/utils/spline.h` | 1 | <stddef.h> |
| I-059 | `src/fields2cover/objectives/rp_obj/direct_dist_path_obj.cpp` | 1 | "fields2cover/types/Geometry_impl.hpp" |
| I-060 | `src/fields2cover/path_planning/dubins_curves.cpp` | 1 | "steering_functions/steering_functions.hpp" |
| I-061 | `src/fields2cover/path_planning/dubins_curves_cc.cpp` | 1 | "steering_functions/steering_functions.hpp" |
| I-062 | `src/fields2cover/path_planning/reeds_shepp_curves.cpp` | 1 | "steering_functions/steering_functions.hpp" |
| I-063 | `src/fields2cover/path_planning/reeds_shepp_curves_hc.cpp` | 1 | "steering_functions/steering_functions.hpp" |
| I-064 | `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp` | 1 | "fields2cover/types/Swath.h" |
| I-065 | `src/fields2cover/swath_generator/brute_force.cpp` | 1 | <numeric> |
| I-066 | `src/fields2cover/types/Graph.cpp` | 1 | <algorithm> |
| I-067 | `src/fields2cover/types/Strip.cpp` | 1 | "fields2cover/types/Geometries_impl.hpp" |
| I-068 | `src/fields2cover/types/Swaths.cpp` | 1 | "fields2cover/types/Geometries_impl.hpp" |
| I-069 | `src/fields2cover/utils/spline.cpp` | 1 | <algorithm> |

#### Unused includes — per-file

| ID | File | # unused | Items to remove |
|---|---|---|---|
| I-070 | `include/fields2cover/types/Robot.h` | 9 | <gdal/ogr_geometry.h>, <utility>, <vector>, "fields2cover/types/Cell.h", +5 more |
| I-071 | `include/fields2cover/route_planning/route_planner_base.h` | 6 | <limits>, <map>, <utility>, "fields2cover/objectives/rp_obj/direct_dist_path_obj.h", +2 more |
| I-072 | `include/fields2cover/utils/random.h` | 5 | <boost/math/constants/constants.hpp>, <limits>, "fields2cover/types/LineString.h", "fields2cover/types/LinearRing.h", +1 more |
| I-073 | `include/fields2cover/path_planning/turning_base.h` | 4 | <functional>, <limits>, <memory>, "fields2cover/utils/random.h" |
| I-074 | `include/fields2cover/swath_generator/brute_force.h` | 4 | <limits>, <memory>, <numeric>, <utility> |
| I-075 | `include/fields2cover/types/Swath.h` | 4 | <gdal/ogr_geometry.h>, <algorithm>, <memory>, <utility> |
| I-076 | `include/fields2cover/route_planning/custom_order.h` | 3 | <set>, <stdexcept>, <string> |
| I-077 | `include/fields2cover/types/Path.h` | 3 | <gdal/ogr_geometry.h>, <fstream>, "fields2cover/types/MultiLineString.h" |
| I-078 | `include/fields2cover/types/Point.h` | 3 | <functional>, <memory>, <utility> |
| I-079 | `include/fields2cover/types/Route.h` | 3 | <gdal/ogr_geometry.h>, <numeric>, <optional> |
| I-080 | `include/fields2cover/utils/parser.h` | 3 | <algorithm>, <vector>, "fields2cover/utils/transformation.h" |
| I-081 | `include/fields2cover/objectives/sg_obj/field_coverage.h` | 2 | <memory>, <utility> |
| I-082 | `include/fields2cover/objectives/sg_obj/overlaps.h` | 2 | <utility>, <vector> |
| I-083 | `include/fields2cover/utils/transformation.h` | 2 | <algorithm>, <utility> |
| I-084 | `src/fields2cover/route_planning/route_planner_base.cpp` | 2 | <math.h>, <limits> |
| I-085 | `src/fields2cover/swath_generator/brute_force.cpp` | 2 | <limits>, <utility> |
| I-086 | `include/fields2cover/objectives/decomp_obj/decomp_objective.h` | 1 | <vector> |
| I-087 | `include/fields2cover/objectives/hg_obj/hg_objective.h` | 1 | <vector> |
| I-088 | `include/fields2cover/objectives/hg_obj/rem_area.h` | 1 | "fields2cover/types.h" |
| I-089 | `include/fields2cover/objectives/pp_obj/pp_objective.h` | 1 | <vector> |
| I-090 | `include/fields2cover/objectives/sg_obj/swath_length.h` | 1 | <numeric> |
| I-091 | `include/fields2cover/swath_generator/swath_generator_base.h` | 1 | <vector> |
| I-092 | `include/fields2cover/types/Cell.h` | 1 | <boost/math/constants/constants.hpp> |
| I-093 | `include/fields2cover/types/Cells.h` | 1 | <vector> |
| I-094 | `include/fields2cover/types/Field.h` | 1 | <memory> |
| I-095 | `include/fields2cover/types/Graph.h` | 1 | <functional> |
| I-096 | `include/fields2cover/types/MultiLineString.h` | 1 | <utility> |
| I-097 | `src/fields2cover/headland_generator/constant_headland.cpp` | 1 | <utility> |
| I-098 | `src/fields2cover/objectives/pp_obj/path_length.cpp` | 1 | "fields2cover/objectives/pp_obj/pp_objective.h" |
| I-099 | `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp` | 1 | <algorithm> |
| I-100 | `src/fields2cover/types/Graph2D.cpp` | 1 | <numeric> |
| I-101 | `src/fields2cover/types/LinearRing.cpp` | 1 | "fields2cover/types/LineString.h" |
| I-102 | `src/fields2cover/types/Path.cpp` | 1 | <steering_functions/utilities/utilities.hpp> |
| I-103 | `src/fields2cover/utils/parser.cpp` | 1 | <boost/optional.hpp> |
| I-104 | `src/fields2cover/utils/visualizer.cpp` | 1 | <matplot/matplot.h> |

**Totals:** I-001..I-069 (missing includes, 69 files, 281 individual lines); I-070..I-104 (unused includes, 35 files, 76 individual lines).

#### IWYU caveats

- The run emits a spurious `-Winjected-class-name` error on `include/fields2cover/types/Geometry_impl.hpp:57,61` for every TU that transitively includes `types.h`. This is an IWYU-compiler compatibility artifact (IWYU uses an older clang frontend than the project's `-Werror`); it does not indicate a real bug and does not prevent IWYU from completing analysis. The same construct (`typename Geometry<T,R>::Geometry&`) will also be flagged by modern compilers eventually and is worth a one-line rewrite (`Geometry<T,R>&` return type). **Flagged separately as A-084** below.
- `steering_functions/steering_functions.hpp` appears as a missing direct include on all four turn-planner TUs (I-060..I-063) — currently pulled in transitively via internal headers. Low priority.
- The `-Xiwyu --no_fwd_decls` flag was set to avoid proposing forward declarations (they're verbose and rarely improve build times for this codebase). A handful of "Add forward decl" suggestions still leaked through in the raw output but were excluded from the per-file tables above.

### Const-Correctness (A-001..A-084)

Full per-finding detail (82 individual A-NNN entries with file:line, what/why/fix/effort) is available in `build/audit/const-review.md` (gitignored intermediate). The following summarizes the categories and highlights the Medium-severity findings that drive plan 06-04.

#### Missing `const` on accessors (by count of queries without `const`)

Zero findings — every get/size/empty-style accessor in the public API is correctly const-qualified. The library is disciplined about const member methods on pure accessors.

#### `const T` return-by-value (low, cross-refs F-NNN via `readability-const-return-type`)

21 findings: A-001..A-005, A-008, A-009, A-012..A-014, A-017, A-018, A-021, A-022, A-025, A-028, A-029, A-038. All in `types/*.h` on `getGeometry`, `getCell*`, `startPoint`, `endPoint`, `operator[]`, `at`, `back`. Fix: drop the top-level `const` on the return type.

#### Missing `const` on parameters

Zero systemic findings. Parameters that should be const-ref are const-ref throughout the reviewed surface. A-051 (`Graph::DFS`) flags a non-const int& output parameter as intentionally non-const (bookkeeping counter) — no change.

#### Missing `[[nodiscard]]`

**Library-wide sweep finding.** Zero uses of `[[nodiscard]]` in `include/fields2cover/` today. Every pure query should have it: getters, `size()`, `empty()`, `length()`, `area()`, `clone()`, arithmetic operators on value types, pure functional static helpers in `Transform`/`Parser`/`Visualizer`.

Enumerated per-header in `build/audit/const-review.md`: A-007, A-011, A-015, A-020, A-023, A-026, A-030, A-032, A-034, A-037, A-040..A-043, A-046..A-049, A-052, A-056, A-057, A-067..A-070, A-074, A-076, A-080..A-082 — ~40 grouped findings covering ~200 individual methods.

Recommended fix strategy: a single mechanical sweep in plan 06-04, applied after the `const T` return-by-value fixes land.

#### Missing `override`

Zero findings. Every virtual override in the reviewed surface uses `override`. (Plan 01's clang-tidy run will catch anything this manual sweep missed.)

#### Medium-severity: query methods that should be const with mutable cache

These are the only const-correctness findings that require non-mechanical work.

- **A-050** `Graph::shortestPathsAndCosts`, `shortestPath`, `shortestPathCost` — memoize into `shortest_paths_`. Fix: mark methods `const`, declare `shortest_paths_` as `mutable` in header, touch impl file.
- **A-053** `Graph2D::shortestPath`, `shortestPathCost` — cascade from A-050.
- **A-066** `TurningBase::createTurn`, `createTurnIfNotCached` — memoize into `path_cache_`. Same fix pattern.
- **A-058** `HeadlandGeneratorBase::generateHeadlands`, `generateHeadlandArea`, `generateHeadlandSwaths` — pure compute (verified against `ConstHL` impl). Fix: declare const in base + all derivatives.
- **A-060/A-062** `RoutePlannerBase::genRoute`, `genShortestRoute` — per-subclass investigation; `computeBestRoute` and `transformSolutionToRoute` are already const, suggesting the top-level `genRoute` could be too.
- **A-054** `SwathGeneratorBase::generateBestSwaths`, `generateSwaths`, `computeCostOfAngle`, `computeBestAngle` — base + overriders synchronized change.
- **A-072/A-073** All `Objective::computeCost` overloads across `SGObjective`, `HGObjective`, `PPObjective`, `RPObjective`, `DecompObjective` hierarchies. Objective functions are by definition pure and should be const; this is the single largest-impact Medium finding because it cascades to `BaseObjective::computeCostWithMinimizingSign` (A-071) and every caller holding a `const Objective&`.
- **A-077** `DecompositionBase::decompose`, `split`, `genSplitLines`, `merge` — pure transforms on input cells. Declare const in base + `BoustrophedonDecomp`, `TrapezoidalDecomp` overrides.

#### Cosmetic / readability

- **A-035** `Path::length(void) const` — drop C-style `(void)`.
- **A-075** `RemArea::isMinimizing(void) const override` — drop C-style `(void)`.
- **A-045** `Robot` setters have unnamed parameters (`void setWidth(double);`).
- **A-055** `SwathGeneratorBase::setAllowOverlap(bool)` unnamed parameter.
- **A-036** `Path::moveTo(const Point&)`, `rotateFromPoint(const Point&, double ang)` — unnamed first parameter.

#### Non-const overloads with no mutation purpose

**A-006, A-010, A-019, A-024, A-027, A-031, A-039** — Every container-like type in `types/*.h` has paired `void getGeometry(size_t i, T& out);` and `void getGeometry(size_t i, T& out) const;` overloads. The non-const overload in each case performs identical work (verified by inspection). These likely exist for SWIG binding compatibility; document the reason or remove.

#### Additional finding from IWYU-triggered inspection

- **A-084** [low] `include/fields2cover/types/Geometry_impl.hpp:57,61` — `typename Geometry<T, R>::Geometry& Geometry<T, R>::operator=(...)` triggers `-Winjected-class-name` on newer clang. Although the project compiles this header with the older system clang, IWYU's embedded parser rejects it. Future clang upgrades will break the build. Fix: rewrite as `Geometry<T, R>& Geometry<T, R>::operator=(...)`.

## Cross-references

- **I-NNN** = include-what-you-use findings (this plan)
- **A-NNN** = api-hygiene manual findings (this plan)
- **F-NNN** = static analysis findings (plan 06-01) — overlaps expected on:
  - `readability-const-return-type` → every A-NNN in the "`const T` return-by-value" bucket
  - `modernize-use-override` → (no matches here)
  - `readability-named-parameter` → A-036, A-045, A-055
  - `modernize-redundant-void-arg` → A-035, A-075
  - `misc-include-cleaner` / `modernize-deprecated-headers` → may overlap with some I-NNN unused/missing include findings
- **M-NNN** = memory audit findings (plan 06-02) — no overlap expected
- **T-NNN** = reserved for future plans

**Triage rule for plan 06-04:** process plan 01 clang-tidy auto-fixes first; the clang-tidy `--fix` pass will resolve the vast majority of `const T` return-value findings and all cosmetic nits. Afterward, apply:

1. A blanket `[[nodiscard]]` sweep across public headers (covers ~40 grouped findings).
2. The 7 Medium `mutable-cache` / pure-virtual const-ification cascades (A-050, A-053, A-058, A-066, A-072, A-073, A-077) — these are the only non-mechanical items and should each land as independent commits with `make test` gating.
3. The IWYU unused-include cleanups (I-070..I-104) — safe mechanical removals.
4. The IWYU missing-include additions (I-001..I-069) — add transitively-available headers explicitly. Low risk but touches many files; consider batching per-subdirectory.

## Statistics

- Total unique source files touched by findings: 73 (of 110 C++ files in the library)
- Mean missing-includes per src TU: 5.2
- Mean unused-includes per public header: 2.2
- Files with 5+ missing includes: 21
- Files with 4+ unused includes: 5

---

*Generated by plan 06-03 as input to plan 06-04 (mechanical fixes) and plan 06-05 (triage).*
