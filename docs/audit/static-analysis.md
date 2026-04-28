# Static Analysis Findings — Phase 6 Plan 01

**Generated:** 2026-04-11
**Tools:** clang-tidy (LLVM 18.1.3), cppcheck 2.13.0
**Scope:** src/fields2cover/ and include/fields2cover/ (all modules)

## Configuration

- `.clang-tidy` check families: `bugprone-*`, `clang-analyzer-*`, `cppcoreguidelines-*`, `modernize-*`, `performance-*`, `readability-*`
- Excluded (noise filters — drown the signal):
  - `modernize-use-trailing-return-type` — stylistic preference
  - `readability-identifier-length` — not actionable at library scale
  - `readability-magic-numbers` / `cppcoreguidelines-avoid-magic-numbers` — library is math-heavy
  - `cppcoreguidelines-pro-bounds-array-to-pointer-decay` / `cppcoreguidelines-pro-bounds-pointer-arithmetic` — fires on every C API interop
- Compile DB: `build/compile_commands.json` generated with `BUILD_TESTS=OFF BUILD_TUTORIALS=OFF BUILD_PYTHON=OFF` (library TUs only)
- cppcheck: `--enable=all --std=c++17 --inline-suppr --suppress=missingIncludeSystem --suppress=unusedFunction`

## Severity Rules

| Severity | clang-tidy checks | cppcheck severities |
|---|---|---|
| **Critical** | `clang-analyzer-core.*`, `clang-analyzer-cplusplus.NewDelete*`, `bugprone-use-after-move`, `bugprone-dangling-handle`, `clang-diagnostic-infinite-recursion` | `error:` |
| **High** | `clang-analyzer-security.*`, `clang-analyzer-deadcode.*`, `bugprone-*` (non-critical), `cppcoreguidelines-owning-memory`, `cppcoreguidelines-slicing` | `warning:` |
| **Medium** | `performance-*`, `cppcoreguidelines-pro-type-*`, `modernize-pass-by-value`, `cppcoreguidelines-narrowing-conversions` (alias) | `performance:` |
| **Low** | `readability-*`, other `modernize-*`, remaining `cppcoreguidelines-*` | `style:` |

Stable IDs **F-001..F-NNN** are assigned below. They are referenced from plans 06-04 (low-risk fixes) and 06-05 (high-risk triage) for traceability.

## Summary

| Severity | Count |
|---|---|
| Critical | 1 |
| High | 117 |
| Medium | 45 |
| Low | 467 |
| **Total unique** | **630** |

## Findings by Module

### swath_generator

*Total findings: 9*

#### Critical

- **F-001** `src/fields2cover/swath_generator/swath_generator_base.cpp:75` `[clang-diagnostic-infinite-recursion]` — all paths through this function will call itself

#### High

- **F-028** `src/fields2cover/swath_generator/swath_generator_base.cpp:43` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'generateSwaths' of similar type ('double') are easily swapped by mistake

#### Medium

None.

#### Low

- **F-232** `include/fields2cover/swath_generator/brute_force.h:22` `[modernize-use-nodiscard]` — function 'getStepAngle' should be marked [[nodiscard]]
- **F-233** `include/fields2cover/swath_generator/swath_generator_base.h:17` `[cppcoreguidelines-special-member-functions]` — class 'SwathGeneratorBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-234** `include/fields2cover/swath_generator/swath_generator_base.h:19` `[modernize-use-nodiscard]` — function 'getAllowOverlap' should be marked [[nodiscard]]
- **F-235** `include/fields2cover/swath_generator/swath_generator_base.h:20` `[readability-named-parameter]` — all parameters should be named in a function
- **F-236** `include/fields2cover/swath_generator/swath_generator_base.h:43` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'allow_overlap' has protected visibility
- **F-237** `src/fields2cover/swath_generator/swath_generator_base.cpp:28` `[useStlAlgorithm]` — Consider using std::transform algorithm instead of a raw loop.
- **F-238** `src/fields2cover/swath_generator/swath_generator_base.cpp:37` `[useStlAlgorithm]` — Consider using std::transform algorithm instead of a raw loop.

### headland_generator

*Total findings: 1*

#### Critical

None.

#### High

None.

#### Medium

None.

#### Low

- **F-169** `include/fields2cover/headland_generator/headland_generator_base.h:17` `[cppcoreguidelines-special-member-functions]` — class 'HeadlandGeneratorBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator

### route_planning

*Total findings: 34*

#### Critical

None.

#### High

- **F-017** `src/fields2cover/route_planning/custom_order.cpp:18` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-018** `src/fields2cover/route_planning/custom_order.cpp:20` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-019** `src/fields2cover/route_planning/custom_order.cpp:20` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'value_type' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-020** `src/fields2cover/route_planning/route_planner_base.cpp:204` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-021** `src/fields2cover/route_planning/snake_order.cpp:14` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined
- **F-022** `src/fields2cover/route_planning/snake_order.cpp:16` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined
- **F-023** `src/fields2cover/route_planning/snake_order.cpp:18` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined
- **F-024** `src/fields2cover/route_planning/spiral_order.cpp:22` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-025** `src/fields2cover/route_planning/spiral_order.cpp:30` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined
- **F-026** `src/fields2cover/route_planning/spiral_order.cpp:31` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined
- **F-027** `src/fields2cover/route_planning/spiral_order.cpp:32` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'difference_type' (aka 'long') is implementation-defined

#### Medium

- **F-127** `src/fields2cover/route_planning/custom_order.cpp:22` `[performance-move-const-arg]` — passing result of std::move() as a const reference argument; no move will actually happen
- **F-128** `src/fields2cover/route_planning/route_planner_base.cpp:128` `[performance-unnecessary-value-param]` — the parameter 'deposit' is copied for each invocation but only used as a const reference; consider making it a const reference
- **F-129** `src/fields2cover/route_planning/spiral_order.cpp:6` `[cppcoreguidelines-pro-type-member-init]` — constructor does not initialize these fields: spiral_size

#### Low

- **F-212** `include/fields2cover/route_planning/custom_order.h:15` `[cppcoreguidelines-special-member-functions]` — class 'CustomOrder' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-213** `include/fields2cover/route_planning/custom_order.h:19` `[cppcoreguidelines-explicit-virtual-functions]` (also: modernize-use-override) — annotate this function with 'override' or (rarely) 'final'
- **F-214** `include/fields2cover/route_planning/route_planner_base.h:23` `[cppcoreguidelines-special-member-functions]` — class 'RoutePlannerBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-215** `include/fields2cover/route_planning/route_planner_base.h:48` `[modernize-use-nodiscard]` — function 'createShortestGraph' should be marked [[nodiscard]]
- **F-216** `include/fields2cover/route_planning/route_planner_base.h:72` `[modernize-use-nodiscard]` — function 'computeBestRoute' should be marked [[nodiscard]]
- **F-217** `include/fields2cover/route_planning/route_planner_base.h:81` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-218** `include/fields2cover/route_planning/route_planner_base.h:82` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'r_start' has protected visibility
- **F-219** `include/fields2cover/route_planning/route_planner_base.h:83` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'r_end' has protected visibility
- **F-220** `include/fields2cover/route_planning/single_cell_swaths_order_base.h:15` `[cppcoreguidelines-special-member-functions]` — class 'SingleCellSwathsOrderBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-221** `include/fields2cover/route_planning/single_cell_swaths_order_base.h:17` `[modernize-use-nodiscard]` — function 'genSortedSwaths' should be marked [[nodiscard]]
- **F-222** `include/fields2cover/route_planning/spiral_order.h:10` `[cppcoreguidelines-special-member-functions]` — class 'SpiralOrder' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-223** `include/fields2cover/route_planning/spiral_order.h:13` `[cppcoreguidelines-explicit-virtual-functions]` (also: modernize-use-override) — annotate this function with 'override' or (rarely) 'final'
- **F-224** `src/fields2cover/route_planning/route_planner_base.cpp:11` `[modernize-deprecated-headers]` — inclusion of deprecated C++ header 'math.h'; consider using 'cmath' instead
- **F-225** `src/fields2cover/route_planning/route_planner_base.cpp:64` `[readability-function-cognitive-complexity]` — function 'createShortestGraph' has cognitive complexity of 31 (threshold 25)
- **F-226** `src/fields2cover/route_planning/route_planner_base.cpp:255` `[shadowVariable]` — Local variable 'swath' shadows outer variable
- **F-227** `src/fields2cover/route_planning/route_planner_base.cpp:263` `[readability-else-after-return]` — do not use 'else' after 'break'
- **F-228** `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp:27` `[readability-implicit-bool-conversion]` — implicit conversion 'uint32_t' (aka 'unsigned int') -> 'bool'
- **F-229** `src/fields2cover/route_planning/single_cell_swaths_order_base.cpp:30` `[readability-implicit-bool-conversion]` — implicit conversion 'uint32_t' (aka 'unsigned int') -> 'bool'
- **F-230** `src/fields2cover/route_planning/snake_order.cpp:12` `[cppcoreguidelines-init-variables]` — variable 'i' is not initialized
- **F-231** `src/fields2cover/route_planning/spiral_order.cpp:28` `[readability-convert-member-functions-to-static]` — method 'spiral' can be made static

### path_planning

*Total findings: 39*

#### Critical

None.

#### High

- **F-009** `src/fields2cover/path_planning/dubins_curves.cpp:14` `[bugprone-easily-swappable-parameters]` — 3 adjacent parameters of 'createSimpleTurn' of similar type ('double') are easily swapped by mistake
- **F-010** `src/fields2cover/path_planning/dubins_curves_cc.cpp:14` `[bugprone-easily-swappable-parameters]` — 3 adjacent parameters of 'createSimpleTurn' of similar type ('double') are easily swapped by mistake
- **F-011** `src/fields2cover/path_planning/path_planning.cpp:33` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-012** `src/fields2cover/path_planning/path_planning.cpp:35` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-013** `src/fields2cover/path_planning/path_planning.cpp:36` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-014** `src/fields2cover/path_planning/reeds_shepp_curves.cpp:14` `[bugprone-easily-swappable-parameters]` — 3 adjacent parameters of 'createSimpleTurn' of similar type ('double') are easily swapped by mistake
- **F-015** `src/fields2cover/path_planning/reeds_shepp_curves_hc.cpp:14` `[bugprone-easily-swappable-parameters]` — 3 adjacent parameters of 'createSimpleTurn' of similar type ('double') are easily swapped by mistake
- **F-016** `src/fields2cover/path_planning/turning_base.cpp:51` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'bool'

#### Medium

- **F-119** `src/fields2cover/path_planning/dubins_curves.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'end'
- **F-120** `src/fields2cover/path_planning/dubins_curves.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'start'
- **F-121** `src/fields2cover/path_planning/dubins_curves_cc.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'end'
- **F-122** `src/fields2cover/path_planning/dubins_curves_cc.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'start'
- **F-123** `src/fields2cover/path_planning/reeds_shepp_curves.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'end'
- **F-124** `src/fields2cover/path_planning/reeds_shepp_curves.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'start'
- **F-125** `src/fields2cover/path_planning/reeds_shepp_curves_hc.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'end'
- **F-126** `src/fields2cover/path_planning/reeds_shepp_curves_hc.cpp:15` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'start'

#### Low

- **F-189** `include/fields2cover/path_planning/steer_to_path.hpp:12` `[missingInclude]` — Include file: "steering_functions/steering_functions.hpp" not found.
- **F-190** `include/fields2cover/path_planning/turning_base.h:22` `[cppcoreguidelines-special-member-functions]` — class 'TurningBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-191** `include/fields2cover/path_planning/turning_base.h:42` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::pp::TurningBase::createTurnIfNotCached' has a definition with different parameter names
- **F-192** `include/fields2cover/path_planning/turning_base.h:70` `[modernize-use-nodiscard]` — function 'getDiscretization' should be marked [[nodiscard]]
- **F-193** `include/fields2cover/path_planning/turning_base.h:75` `[modernize-use-nodiscard]` — function 'getUsingCache' should be marked [[nodiscard]]
- **F-194** `include/fields2cover/path_planning/turning_base.h:91` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'path_cache_' has protected visibility
- **F-195** `include/fields2cover/path_planning/turning_base.h:92` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'discretization' has protected visibility
- **F-196** `include/fields2cover/path_planning/turning_base.h:93` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'using_cache' has protected visibility
- **F-197** `src/fields2cover/path_planning/dubins_curves.cpp:15` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- **F-198** `src/fields2cover/path_planning/dubins_curves_cc.cpp:15` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- **F-199** `src/fields2cover/path_planning/path_planning.cpp:51` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- **F-200** `src/fields2cover/path_planning/path_planning.cpp:52` `[cppcoreguidelines-init-variables]` — variable 'ang1' is not initialized
- **F-201** `src/fields2cover/path_planning/path_planning.cpp:52` `[cppcoreguidelines-init-variables]` — variable 'ang2' is not initialized
- **F-202** `src/fields2cover/path_planning/path_planning.cpp:52` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- **F-203** `src/fields2cover/path_planning/path_planning.cpp:95` `[cppcoreguidelines-init-variables]` — variable 'x' is not initialized
- **F-204** `src/fields2cover/path_planning/path_planning.cpp:95` `[cppcoreguidelines-init-variables]` — variable 'y' is not initialized
- **F-205** `src/fields2cover/path_planning/path_planning.cpp:95` `[cppcoreguidelines-init-variables]` — variable 'ang' is not initialized
- **F-206** `src/fields2cover/path_planning/path_planning.cpp:95` `[cppcoreguidelines-init-variables]` — variable 'k' is not initialized
- **F-207** `src/fields2cover/path_planning/path_planning.cpp:95` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- **F-208** `src/fields2cover/path_planning/reeds_shepp_curves.cpp:15` `[readability-isolate-declaration]` — multiple declarations in a single statement reduces readability
- *(... 3 more low findings in path_planning truncated — see `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt`)*

### objectives

*Total findings: 26*

#### Critical

None.

#### High

- **F-002** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:103` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-003** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:114` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-004** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:115` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-005** `src/fields2cover/objectives/sg_obj/n_swath.cpp:16` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to 'double'
- **F-006** `src/fields2cover/objectives/sg_obj/n_swath_modified.cpp:12` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'computeCost' of similar type ('double') are easily swapped by mistake
- **F-007** `src/fields2cover/objectives/sg_obj/overlaps.cpp:18` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-008** `src/fields2cover/objectives/sg_obj/sg_objective.cpp:38` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'computeCost' of similar type ('double') are easily swapped by mistake

#### Medium

None.

#### Low

- **F-170** `include/fields2cover/objectives/base_objective.h:17` `[cppcoreguidelines-special-member-functions]` — class 'BaseObjective' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-171** `include/fields2cover/objectives/base_objective.h:48` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-172** `include/fields2cover/objectives/base_objective.h:50` `[modernize-use-nodiscard]` — function 'isMinimizing' should be marked [[nodiscard]]
- **F-173** `include/fields2cover/objectives/base_objective.h:52` `[modernize-use-nodiscard]` — function 'isMaximizing' should be marked [[nodiscard]]
- **F-174** `include/fields2cover/objectives/hg_obj/rem_area.h:20` `[modernize-redundant-void-arg]` — redundant void argument list in function declaration
- **F-175** `include/fields2cover/objectives/hg_obj/rem_area.h:20` `[modernize-use-nodiscard]` — function 'isMinimizing' should be marked [[nodiscard]]
- **F-176** `include/fields2cover/objectives/sg_obj/field_coverage.h:26` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-177** `include/fields2cover/objectives/sg_obj/field_coverage.h:27` `[modernize-redundant-void-arg]` — redundant void argument list in function declaration
- **F-178** `include/fields2cover/objectives/sg_obj/field_coverage.h:27` `[modernize-use-nodiscard]` — function 'isMinimizing' should be marked [[nodiscard]]
- **F-179** `include/fields2cover/objectives/sg_obj/n_swath_modified.h:23` `[modernize-use-nodiscard]` — function 'isFastCompAvailable' should be marked [[nodiscard]]
- **F-180** `include/fields2cover/objectives/sg_obj/sg_objective.h:23` `[modernize-use-nodiscard]` — function 'isFastCompAvailable' should be marked [[nodiscard]]
- **F-181** `include/fields2cover/objectives/sg_obj/sg_objective.h:26` `[readability-named-parameter]` — all parameters should be named in a function
- **F-182** `include/fields2cover/objectives/sg_obj/sg_objective.h:28` `[readability-named-parameter]` — all parameters should be named in a function
- **F-183** `include/fields2cover/objectives/sg_obj/sg_objective.h:32` `[readability-named-parameter]` — all parameters should be named in a function
- **F-184** `include/fields2cover/objectives/sg_obj/sg_objective.h:34` `[readability-named-parameter]` — all parameters should be named in a function
- **F-185** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:89` `[readability-implicit-bool-conversion]` — implicit conversion 'OGRBoolean' (aka 'int') -> 'bool'
- **F-186** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:123` `[useStlAlgorithm]` — Consider using std::accumulate algorithm instead of a raw loop.
- **F-187** `src/fields2cover/objectives/rp_obj/rp_objective.cpp:126` `[useStlAlgorithm]` — Consider using std::accumulate algorithm instead of a raw loop.
- **F-188** `src/fields2cover/objectives/sg_obj/sg_objective.cpp:11` `[readability-named-parameter]` — all parameters should be named in a function

### types

*Total findings: 453*

#### Critical

None.

#### High

- **F-029** `include/fields2cover/types/Geometry_impl.hpp:329` `[cppcoreguidelines-owning-memory]` — deleting a pointer through a type that is not marked 'gsl::owner<>'; consider using a smart pointer instead
- **F-030** `include/fields2cover/types/Graph.h:20` `[bugprone-reserved-identifier]` — declaration uses identifier 'pair_vec_size__int', which is a reserved identifier
- **F-031** `include/fields2cover/types/Graph2D.h:40` `[duplInheritedMember]` — The class 'Graph2D' defines member function with name 'numNodes' also defined in its parent class 'Graph'.
- **F-032** `src/fields2cover/types/Cell.cpp:42` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-033** `src/fields2cover/types/Cell.cpp:53` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-034** `src/fields2cover/types/Cell.cpp:137` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-035** `src/fields2cover/types/Cells.cpp:68` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-036** `src/fields2cover/types/Cells.cpp:76` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-037** `src/fields2cover/types/Cells.cpp:84` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-038** `src/fields2cover/types/Cells.cpp:92` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-039** `src/fields2cover/types/Cells.cpp:118` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-040** `src/fields2cover/types/Cells.cpp:122` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-041** `src/fields2cover/types/Cells.cpp:123` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-042** `src/fields2cover/types/Cells.cpp:131` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-043** `src/fields2cover/types/Field.cpp:99` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'getUTMCoordSystem' of similar type ('const std::string &') are easily swapped by mistake
- **F-044** `src/fields2cover/types/Field.cpp:109` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'getUTMDatum' of similar type ('const std::string &') are easily swapped by mistake
- **F-045** `src/fields2cover/types/Graph.cpp:60` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'getCostFromEdge' of convertible types are easily swapped by mistake
- **F-046** `src/fields2cover/types/Graph.cpp:87` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_type' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-047** `src/fields2cover/types/Graph.cpp:110` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'unsigned long' to signed type 'value_type' (aka 'long') is implementation-defined
- **F-048** `src/fields2cover/types/Graph.cpp:153` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'shortestPath' of convertible types are easily swapped by mistake
- *(... 54 more high findings in types truncated — see `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt`)*

#### Medium

- **F-130** `include/fields2cover/types/Field.h:24` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-131** `include/fields2cover/types/Field.h:25` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-132** `include/fields2cover/types/Geometry.h:37` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-133** `include/fields2cover/types/Geometry.h:38` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-134** `include/fields2cover/types/PathState.h:15` `[performance-enum-size]` — enum 'PathSectionType' uses a larger base type ('int', size: 4 bytes) than necessary for its value set, consider using 'std::uint8_t' (1 byte) as the base type to reduce its size
- **F-135** `include/fields2cover/types/PathState.h:21` `[performance-enum-size]` — enum 'PathDirection' uses a larger base type ('int', size: 4 bytes) than necessary for its value set, consider using 'std::int8_t' (1 byte) as the base type to reduce its size
- **F-136** `include/fields2cover/types/Point.h:27` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-137** `include/fields2cover/types/Point.h:30` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-138** `include/fields2cover/types/Robot.h:32` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-139** `include/fields2cover/types/Robot.h:34` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-140** `include/fields2cover/types/Swath.h:21` `[performance-enum-size]` — enum 'SwathType' uses a larger base type ('int', size: 4 bytes) than necessary for its value set, consider using 'std::uint8_t' (1 byte) as the base type to reduce its size
- **F-141** `include/fields2cover/types/Swath.h:32` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-142** `src/fields2cover/types/Field.cpp:13` `[modernize-pass-by-value]` — pass by value and use std::move
- **F-143** `src/fields2cover/types/Field.cpp:22` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-144** `src/fields2cover/types/Field.cpp:23` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-145** `src/fields2cover/types/Graph.cpp:74` `[performance-unnecessary-value-param]` — the const qualified parameter 'x' is copied for each invocation; consider making it a reference
- **F-146** `src/fields2cover/types/Graph.cpp:74` `[passedByValue]` — Function parameter 'x' should be passed by const reference.
- **F-147** `src/fields2cover/types/Graph2D.cpp:64` `[performance-inefficient-vector-operation]` — 'emplace_back' is called inside a loop; consider pre-allocating the container capacity before the loop
- **F-148** `src/fields2cover/types/Graph2D.cpp:96` `[performance-inefficient-vector-operation]` — 'emplace_back' is called inside a loop; consider pre-allocating the container capacity before the loop
- **F-149** `src/fields2cover/types/LineString.cpp:13` `[cppcoreguidelines-pro-type-static-cast-downcast]` — do not use static_cast to downcast from a base to a derived class; use dynamic_cast instead
- *(... 7 more medium findings in types truncated — see `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt`)*

#### Low

- **F-239** `include/fields2cover/types/Cell.h:24` `[cppcoreguidelines-missing-std-forward]` — forwarding reference parameter 'args' is never forwarded inside the function body
- **F-240** `include/fields2cover/types/Cell.h:32` `[cppcoreguidelines-special-member-functions]` — class 'Cell' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-241** `include/fields2cover/types/Cell.h:45` `[modernize-use-nodiscard]` — function 'getGeometry' should be marked [[nodiscard]]
- **F-242** `include/fields2cover/types/Cell.h:47` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::types::Cell::setGeometry' has a definition with different parameter names
- **F-243** `include/fields2cover/types/Cell.h:49` `[modernize-use-nodiscard]` — function 'size' should be marked [[nodiscard]]
- **F-244** `include/fields2cover/types/Cell.h:59` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::types::Cell::buffer' has a definition with different parameter names
- **F-245** `include/fields2cover/types/Cell.h:63` `[modernize-use-nodiscard]` — function 'convexHull' should be marked [[nodiscard]]
- **F-246** `include/fields2cover/types/Cell.h:68` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::types::Cell::addRing' has a definition with different parameter names
- **F-247** `include/fields2cover/types/Cell.h:71` `[modernize-use-nodiscard]` — function 'getExteriorRing' should be marked [[nodiscard]]
- **F-248** `include/fields2cover/types/Cell.h:72` `[modernize-use-nodiscard]` — function 'getInteriorRing' should be marked [[nodiscard]]
- **F-249** `include/fields2cover/types/Cell.h:75` `[modernize-use-nodiscard]` — function 'isConvex' should be marked [[nodiscard]]
- **F-250** `include/fields2cover/types/Cell.h:79` `[modernize-use-nodiscard]` — function 'createSemiLongLine' should be marked [[nodiscard]]
- **F-251** `include/fields2cover/types/Cell.h:83` `[modernize-use-nodiscard]` — function 'createStraightLongLine' should be marked [[nodiscard]]
- **F-252** `include/fields2cover/types/Cell.h:86` `[modernize-use-nodiscard]` — function 'getLinesInside' should be marked [[nodiscard]]
- **F-253** `include/fields2cover/types/Cell.h:89` `[modernize-use-nodiscard]` — function 'getLinesInside' should be marked [[nodiscard]]
- **F-254** `include/fields2cover/types/Cell.h:92` `[modernize-use-nodiscard]` — function 'isPointInBorder' should be marked [[nodiscard]]
- **F-255** `include/fields2cover/types/Cell.h:95` `[modernize-use-nodiscard]` — function 'isPointIn' should be marked [[nodiscard]]
- **F-256** `include/fields2cover/types/Cell.h:98` `[modernize-use-nodiscard]` — function 'createLineUntilBorder' should be marked [[nodiscard]]
- **F-257** `include/fields2cover/types/Cell.h:101` `[modernize-use-nodiscard]` — function 'closestPointOnBorderTo' should be marked [[nodiscard]]
- **F-258** `include/fields2cover/types/Cells.h:21` `[cppcoreguidelines-special-member-functions]` — class 'Cells' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- *(... 332 more low findings in types truncated — see `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt`)*

### utils

*Total findings: 63*

#### Critical

None.

#### High

- **F-103** `src/fields2cover/utils/random.cpp:43` `[bugprone-easily-swappable-parameters]` — 3 adjacent parameters of 'generateRandCell' of convertible types are easily swapped by mistake
- **F-104** `src/fields2cover/utils/random.cpp:54` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to 'double'
- **F-105** `src/fields2cover/utils/random.cpp:72` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-106** `src/fields2cover/utils/spline.cpp:92` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-107** `src/fields2cover/utils/spline.cpp:139` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to signed type 'int' is implementation-defined
- **F-108** `src/fields2cover/utils/visualizer.cpp:16` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'plot' of similar type ('const std::vector<double> &') are easily swapped by mistake
- **F-109** `src/fields2cover/utils/visualizer.cpp:19` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'float'
- **F-110** `src/fields2cover/utils/visualizer.cpp:28` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'value_type' (aka 'double') to 'value_type' (aka 'float')
- **F-111** `src/fields2cover/utils/visualizer.cpp:49` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'float'
- **F-112** `src/fields2cover/utils/visualizer.cpp:94` `[bugprone-implicit-widening-of-multiplication-result]` — performing an implicit widening conversion to type 'size_t' (aka 'unsigned long') of a multiplication performed in type 'int'
- **F-113** `src/fields2cover/utils/visualizer.cpp:107` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'float'
- **F-114** `src/fields2cover/utils/visualizer.cpp:192` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'float'
- **F-115** `src/fields2cover/utils/visualizer.cpp:246` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'double' to 'float'
- **F-116** `src/fields2cover/utils/visualizer.cpp:260` `[bugprone-easily-swappable-parameters]` — 2 adjacent parameters of 'figure_size' of similar type ('const unsigned int') are easily swapped by mistake
- **F-117** `src/fields2cover/utils/visualizer.cpp:299` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to 'double'
- **F-118** `src/fields2cover/utils/visualizer.cpp:302` `[bugprone-narrowing-conversions]` (also: cppcoreguidelines-narrowing-conversions) — narrowing conversion from 'size_t' (aka 'unsigned long') to 'double'

#### Medium

- **F-157** `include/fields2cover/utils/random.h:30` `[performance-trivially-destructible]` — class 'Random' can be made trivially destructible by defaulting the destructor on its first declaration
- **F-158** `include/fields2cover/utils/random.h:33` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move constructors should be marked noexcept
- **F-159** `include/fields2cover/utils/random.h:34` `[cppcoreguidelines-noexcept-move-operations]` (also: performance-noexcept-move-constructor) — move assignment operators should be marked noexcept
- **F-160** `src/fields2cover/utils/parser.cpp:59` `[performance-unnecessary-value-param]` — the parameter 'toSearch' is copied for each invocation but only used as a const reference; consider making it a const reference
- **F-161** `src/fields2cover/utils/parser.cpp:60` `[performance-unnecessary-value-param]` — the parameter 'replaceStr' is copied for each invocation but only used as a const reference; consider making it a const reference
- **F-162** `src/fields2cover/utils/parser.cpp:169` `[performance-avoid-endl]` — do not use 'std::endl' with streams; use '\n' instead
- **F-163** `src/fields2cover/utils/visualizer.cpp:26` `[cppcoreguidelines-pro-type-member-init]` — uninitialized record type: 'fc'

#### Low

- **F-591** `include/fields2cover/utils/parser.h:38` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::Parser::importCellJsonFromString' has a definition with different parameter names
- **F-592** `include/fields2cover/utils/random.h:27` `[modernize-use-nullptr]` — use nullptr
- **F-593** `include/fields2cover/utils/random.h:37` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-594** `include/fields2cover/utils/spline.h:44` `[cppcoreguidelines-special-member-functions]` — class 'CubicSpline' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-595** `include/fields2cover/utils/spline.h:44` `[cppcoreguidelines-special-member-functions]` — class 'CubicSpline' defines a destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-596** `include/fields2cover/utils/transformation.h:44` `[readability-inconsistent-declaration-parameter-name]` — function 'f2c::Transform::transformToPrevCRS' has a definition with different parameter names
- **F-597** `include/fields2cover/utils/visualizer.h:121` `[modernize-type-traits]` — use c++17 style variable templates
- **F-598** `include/fields2cover/utils/visualizer.h:137` `[modernize-type-traits]` — use c++17 style variable templates
- **F-599** `include/fields2cover/utils/visualizer.h:140` `[modernize-type-traits]` — use c++17 style variable templates
- **F-600** `src/fields2cover/utils/parser.cpp:40` `[readability-redundant-string-init]` — redundant string initialization
- **F-601** `src/fields2cover/utils/parser.cpp:42` `[readability-qualified-auto]` — 'auto e_result' can be declared as 'const auto *e_result'
- **F-602** `src/fields2cover/utils/parser.cpp:85` `[modernize-return-braced-init-list]` — avoid repeating the return type from the declaration; use a braced initializer list instead
- **F-603** `src/fields2cover/utils/parser.cpp:86` `[readability-else-after-return]` — do not use 'else' after 'return'
- **F-604** `src/fields2cover/utils/parser.cpp:87` `[modernize-return-braced-init-list]` — avoid repeating the return type from the declaration; use a braced initializer list instead
- **F-605** `src/fields2cover/utils/parser.cpp:89` `[modernize-return-braced-init-list]` — avoid repeating the return type from the declaration; use a braced initializer list instead
- **F-606** `src/fields2cover/utils/parser.cpp:102` `[modernize-loop-convert]` — use range-based for loop instead
- **F-607** `src/fields2cover/utils/parser.cpp:109` `[cppcoreguidelines-avoid-c-arrays]` (also: modernize-avoid-c-arrays) — do not declare C-style arrays, use std::array<> instead
- **F-608** `src/fields2cover/utils/parser.cpp:117` `[cppcoreguidelines-pro-bounds-constant-array-index]` — do not use array subscript when the index is not an integer constant expression
- **F-609** `src/fields2cover/utils/parser.cpp:129` `[modernize-use-emplace]` — unnecessary temporary object created while calling emplace_back
- **F-610** `src/fields2cover/utils/random.cpp:71` `[cppcoreguidelines-avoid-do-while]` — avoid do-while loops
- *(... 20 more low findings in utils truncated — see `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt`)*

### decomposition

*Total findings: 5*

#### Critical

None.

#### High

None.

#### Medium

None.

#### Low

- **F-164** `include/fields2cover/decomposition/decomposition_base.h:21` `[cppcoreguidelines-special-member-functions]` — class 'DecompositionBase' defines a default destructor but does not define a copy constructor, a copy assignment operator, a move constructor or a move assignment operator
- **F-165** `include/fields2cover/decomposition/decomposition_base.h:31` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-166** `include/fields2cover/decomposition/trapezoidal_decomp.h:23` `[modernize-use-nodiscard]` — function 'getSplitAngle' should be marked [[nodiscard]]
- **F-167** `include/fields2cover/decomposition/trapezoidal_decomp.h:27` `[readability-redundant-access-specifiers]` — redundant access specifier has the same accessibility as the previous access specifier
- **F-168** `include/fields2cover/decomposition/trapezoidal_decomp.h:35` `[cppcoreguidelines-non-private-member-variables-in-classes]` — member variable 'split_angle' has protected visibility

## Module Hotspots

| Module | Critical | High | Medium | Low | Total |
|---|---|---|---|---|---|
| swath_generator | 1 | 1 | 0 | 7 | 9 |
| headland_generator | 0 | 0 | 0 | 1 | 1 |
| route_planning | 0 | 11 | 3 | 20 | 34 |
| path_planning | 0 | 8 | 8 | 23 | 39 |
| objectives | 0 | 7 | 0 | 19 | 26 |
| types | 0 | 74 | 27 | 352 | 453 |
| utils | 0 | 16 | 7 | 40 | 63 |
| decomposition | 0 | 0 | 0 | 5 | 5 |

## Raw Output

See `build/audit/clang-tidy-raw.txt` and `build/audit/cppcheck-raw.txt` (not committed — regenerate via plan 06-01 Task 2 commands).
