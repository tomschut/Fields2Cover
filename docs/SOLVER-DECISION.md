# Solver Decision — Route Planning TSP

**Phase:** 5 — Solver Evaluation
**Requirement:** SLV-03
**Status:** Provisional (pending benchmark numbers — see gating conditions below)
**Date:** 2026-04-11

## Verdict

> **KEEP** OR-Tools as the primary TSP solver for `RoutePlannerBase::computeBestRoute`,
> with a Phase 7 follow-up to add a BSD-3-licensed nearest-neighbor + 2-opt fallback
> that engages when OR-Tools returns no feasible solution. This is a **provisional**
> verdict: Plan 05-01's benchmark harness could not be run on the planning host
> (gnuplot is absent, and the top-level `CMakeLists.txt:169` `BUILD_TESTS AND
> GNUPLOT_FOUND` guard skips the entire `tests/` subtree, so `route_planner_benchmark`
> is never emitted). The verdict stands on the literature and on the current
> integration surface; it must be re-confirmed on the first build-host run of
> `route_planner_benchmark` before Phase 7 freezes its direction. If the measured
> p50 at n=500 exceeds 1500 ms, or if route quality regresses >5% vs. the current
> configuration, the verdict flips to **AUGMENT** (see gating conditions in the
> "If KEEP" section).

## Context

`f2c::rp::RoutePlannerBase::computeBestRoute` (see
[`src/fields2cover/route_planning/route_planner_base.cpp`](../src/fields2cover/route_planning/route_planner_base.cpp),
lines 199–242) currently uses Google OR-Tools' `RoutingModel` with the
`FirstSolutionStrategy::AUTOMATIC` first-solution strategy, the
`LocalSearchMetaheuristic::AUTOMATIC` local-search metaheuristic,
`use_full_propagation(false)`, and a 1-second wall-clock time limit. The solver is
called once per `genRoute()` invocation with a single vehicle, a single depot
(`numNodes - 1`), and an arc-cost evaluator that reads edge weights directly from
an `F2CGraph2D` shortest-path graph. Phase 4 (plan 04-01) hardened this call site
to throw `std::runtime_error` instead of dereferencing a null `Assignment*` when no
feasible solution exists.

This is the TSP-solver surface area under evaluation.

Evaluation inputs:

- Benchmark: [`perf/ortools-baseline.md`](perf/ortools-baseline.md)
- Alternatives survey: [`../.planning/research/solvers-comparison.md`](../.planning/research/solvers-comparison.md)
- Phase 4 integration hardening: `src/fields2cover/route_planning/route_planner_base.cpp:227-231`

## Current-state benchmarks

> **STATUS: pending.** Plan 05-01 delivered the harness
> (`tests/cpp/route_planning/benchmarks/ortools_benchmark.cpp`,
> `bench_fields.h`, and a CMake target `route_planner_benchmark`) and a methodology
> document but could not produce numbers on the planning host. The build host's
> `cmake -S . -B build-bench -DBUILD_TESTS=ON -DBUILD_PYTHON=OFF` configured
> cleanly (OR-Tools, GDAL, Eigen all resolved) but
> `cmake --build build-bench --target route_planner_benchmark` failed with
> `gmake: *** No rule to make target 'route_planner_benchmark'. Stop.` because
> `CMakeLists.txt:169` gates the whole `tests/` subdirectory behind
> `BUILD_TESTS AND GNUPLOT_FOUND`, and gnuplot is not installed. The one-line fix
> is `apt-get install gnuplot`; see the diagnostics block in
> [`perf/ortools-baseline.md`](perf/ortools-baseline.md) for the verbatim cmake
> output.

| n_swaths | median_ms | stddev_ms | route_length_m | peak_rss_kb |
| -------: | --------: | --------: | -------------: | ----------: |
|       10 |   pending |   pending |        pending |     pending |
|       50 |   pending |   pending |        pending |     pending |
|      100 |   pending |   pending |        pending |     pending |
|      500 |   pending |   pending |        pending |     pending |

**Observations (from the literature, to be confirmed by measurement):**

- OR-Tools' Guided Local Search (the effective metaheuristic once AUTOMATIC
  resolves) is reported in Google's routing manual to land within 0–2% of optimal
  on random Euclidean TSP at n ≤ 500, and it typically converges in tens of
  milliseconds at that scale. The current 1-second time limit is therefore
  likely to be "wasted budget" — the solver finishes early but the harness
  should confirm whether wall-clock approaches the 1000 ms cap at n=500.
- The coverage graph in fields2cover is highly structured (parallel swaths with
  short headland transitions), which is *easier* than random Euclidean TSP for
  any reasonable local-search heuristic; expect route quality to be near-optimal
  regardless of which sensible solver is chosen.
- Memory growth should be O(n²) on the cost-matrix side (n = numNodes), dominated
  by OR-Tools' internal constraint-graph representation. Peak RSS at n=500 is
  expected to stay well under a few hundred MB.
- Stddev / median ratio will reveal whether OR-Tools' metaheuristic
  nondeterminism or measurement noise dominates — if stddev is <5% of median,
  the deterministic seed path is working.

These observations are the basis of the provisional verdict. They must be
replaced with measured numbers as soon as a build-capable host runs the harness.

## Alternatives considered

Full details in
[`../.planning/research/solvers-comparison.md`](../.planning/research/solvers-comparison.md).
Short summary:

| Solver                       | BSD-3 OK                 | Integration           | Quality @ n≤500                     | Carried forward?                                       |
| ---------------------------- | ------------------------ | --------------------- | ----------------------------------- | ------------------------------------------------------ |
| OR-Tools (baseline)          | yes (Apache-2.0)         | (current)             | excellent (~0–2% gap, literature)   | yes — baseline, kept                                   |
| LKH-3                        | **no (research-only)**   | high                  | best-in-class (<1% gap)             | **no — license**                                       |
| Concorde                     | **no (research-only)**   | high (needs LP)       | optimal (0% gap)                    | **no — license + LP backend also license-blocked**     |
| Custom Lin-Kernighan (BSD-3) | yes (ours)               | medium (1–2 weeks)    | good (~2–5% gap, estimate)          | deferred — Phase 7 prototype behind CMake option       |
| Nearest-neighbor + 2-opt     | yes (ours)               | low (1–2 days)        | acceptable (<5% gap on coverage)    | **yes — as Phase 7 fallback for OR-Tools infeasible**  |

## Rationale

Walking the decision framework from
[`05-03-PLAN.md`](../.planning/phases/05-solver-evaluation/05-03-PLAN.md):

- **License gate.** LKH-3 and Concorde are both disqualified. LKH-3's canonical
  distribution page carries non-OSI "research use only, commercial use by
  author permission" language; Concorde's README at
  `math.uwaterloo.ca/tsp/concorde.html` is explicit about academic-only use, and
  every viable Concorde LP backend (QSopt, QSopt_ex, CPLEX) is itself license-toxic.
  Neither can be linked into a BSD-3 library that invites commercial reuse.
  This leaves OR-Tools, custom Lin-Kernighan, and NN+2-opt as the only
  candidates. See the license deep-dive in
  [`solvers-comparison.md`](../.planning/research/solvers-comparison.md) §
  "License deep-dive on LKH-3 and Concorde".
- **Solve-time budget.** Target: median `computeBestRoute` wall-clock <500 ms at
  n=500. OR-Tools with GLS typically finishes in tens of milliseconds at this
  scale according to published routing benchmarks, so it is expected to be
  well inside budget. Numbers are pending; if they exceed 1500 ms at n=500
  the verdict flips.
- **Dep weight.** OR-Tools pulls a heavy transitive set (protobuf, abseil, glog,
  gflags, re2, eigen; author's estimate ~50 MB). This is a real pain point —
  it is the most common fields2cover install complaint — but the sunk cost is
  already paid. Replacing OR-Tools to save build weight is a multi-week project
  (custom LK) whose quality impact is unknown. Not justified at this stage
  without benchmark evidence that OR-Tools is also slow or buggy.
- **Crash history.** Phase 4 fixed the known OR-Tools crash
  (`solution == nullptr` dereference) in `route_planner_base.cpp:227-231` and
  added regression tests in `tests/cpp/route_planning/route_planner_base_test.cpp`.
  No other OR-Tools crash is currently open. OR-Tools *does* occasionally report
  infeasibility on trivially Hamiltonian small instances (n ≤ 3); this is the
  concrete motivation for the Phase 7 NN+2-opt fallback rather than for an
  outright replacement.
- **Maintenance burden.** OR-Tools is Google-maintained with multiple releases
  per year, a current pin of `v9.9.3963` (see `cmake/F2CUtils.cmake:34-41`),
  excellent CI, and a healthy issue tracker. Maintenance risk is negligible.
  The custom-LK alternative carries permanent first-party maintenance cost
  with no upstream to fall back on.
- **Primary driver.** With license-blocked candidates removed and crash history
  addressed, the only concrete case for replacement would be a benchmark-proven
  speed or quality regression at n=500. Until the harness produces numbers,
  the null hypothesis ("OR-Tools is fine") stands.

## If KEEP

**Conditions for revisit (any one of these flips the verdict to AUGMENT or REPLACE):**

- Measured p50 wall-clock of `genRoute` exceeds **1500 ms at n=500** on a
  representative build host. (The current 1-second OR-Tools time limit means
  anything above ~1100 ms is the solver hitting the wall — it *found* a
  solution but did not converge within budget, and route length is likely
  plateaued.)
- Measured route length regresses by **>5%** vs. the current v9.9.3963
  baseline after a future OR-Tools upgrade.
- Swath counts per field routinely exceed **2000** on real customer data
  (well beyond the current 500-node ceiling).
- OR-Tools' transitive-dependency surface (protobuf, abseil, glog, etc.)
  causes a supply-chain audit failure for a downstream consumer, or a
  protobuf ODR conflict with another library loaded into the same process.
- A new OR-Tools-triggered crash is found in production that cannot be
  fixed inside `computeBestRoute` with a null-guard (Phase 4 pattern).

**Phase 7 action on the solver itself:** parameter tuning only. No
replacement. Specifically:

- The 1-second time limit
  (`searchParameters.mutable_time_limit()->set_seconds(1)`,
  `route_planner_base.cpp:223`) is likely generous by an order of magnitude
  at n ≤ 500. Consider reducing to 100–250 ms once benchmark numbers
  confirm GLS converges well before 1000 ms.
- Consider pinning `FirstSolutionStrategy` to `PATH_CHEAPEST_ARC` (or
  `SAVINGS`) and `LocalSearchMetaheuristic` to `GUIDED_LOCAL_SEARCH`
  explicitly, rather than leaving both on `AUTOMATIC`. This gives
  reproducible behaviour without relying on OR-Tools' internal selection
  heuristic.
- File-level scope for Phase 7 tuning:
  `src/fields2cover/route_planning/route_planner_base.cpp:214-223`.

**Phase 7 action not on the primary solver — the NN+2-opt fallback
(no-regret safety improvement, regardless of benchmark outcome):**

- **New file:** `src/fields2cover/route_planning/nn_2opt_solver.cpp` (with a
  header `include/fields2cover/route_planning/nn_2opt_solver.h`).
- **Public API:** unchanged. `RoutePlannerBase::computeBestRoute` catches
  its own `std::runtime_error` from the OR-Tools path and delegates to
  `nn_2opt::solveTSP(cov_graph, depot_id)` as a fallback. Callers of
  `genRoute` see a tour in cases that currently throw.
- **Gating condition:** the fallback engages only when OR-Tools returns
  `solution == nullptr` or throws the "no feasible solution" runtime error
  from `route_planner_base.cpp:227-231`. On the happy path, OR-Tools
  remains the sole solver.
- **Effort estimate:** 1–2 person-days for the solver + unit tests
  (`tests/cpp/route_planning/nn_2opt_solver_test.cpp`) + one end-to-end
  regression that constructs a scenario known to make OR-Tools report
  infeasibility and verifies the fallback returns a non-empty route.
- **Phase 5 benchmark extension:** add a `nn_2opt_benchmark` sibling
  executable under `tests/cpp/route_planning/benchmarks/` that runs the
  same four input sizes; both solvers' numbers go into an updated
  `perf/ortools-baseline.md` (which can then be renamed `perf/solver-baseline.md`).

**Phase 7 exploratory work — custom LK prototype behind a CMake option:**

This is not triggered by the KEEP verdict; it is an *open door*. If the
benchmark numbers show that OR-Tools' build weight is unjustified (for example
a custom LK prototype matches OR-Tools' quality within ≤2% while cutting the
transitive-dependency surface to zero), Phase 7 or Phase 8 can add
`F2C_TSP_SOLVER={ortools|lk|nn2opt}` as a CMake option and flip the default.
This is *not* committed to by the current verdict; it is explicitly optional
follow-up work justified only by a future benchmark result.

## Non-goals

- Benchmarks against real customer field data. No data access; see the
  deferred list in
  [`05-CONTEXT.md`](../.planning/phases/05-solver-evaluation/05-CONTEXT.md).
- Runtime solver-switching config (deferred until after this decision
  lands and the NN+2-opt fallback is implemented).
- Replacing OR-Tools. Explicitly out of scope for this verdict. Any
  replacement requires a separate decision document citing benchmark
  numbers, not literature estimates.

## Appendix A — OR-Tools integration surface

The current touch points are narrow — a future replacement (or the NN+2-opt
fallback) only needs to match this interface:

- **Input:** `const F2CGraph2D& cov_graph` — a directed weighted graph with
  `N` nodes, where `N - 1` is the depot. The solver must solve a
  Hamiltonian path / TSP on this graph starting and ending at the depot
  node. Edge weights are read via `cov_graph.getCostFromEdge(from, to)`.
- **Output:** `std::vector<int64_t>` of node indices in visit order,
  excluding the depot (which is implicit at both ends of the tour).
- **Error signal:** throw `std::runtime_error` if no feasible solution
  exists — matches the Phase 4 fix at
  `src/fields2cover/route_planning/route_planner_base.cpp:227-231`. A
  fallback solver in the "If KEEP" section intercepts this by catching the
  exception and returning its own tour.
- **Construction overhead:** one solve per `genRoute()` call; no
  amortisation across calls. Construction + first-solution + local-search
  budget must fit within the same wall-clock window as the current
  1-second OR-Tools configuration.

## Appendix B — Plan 05-01 build-host fix

Future runs of the Plan 05-01 benchmark harness (required to un-gate this
provisional verdict) need:

```bash
# 1. Install gnuplot so the tests/ tree is not skipped
sudo apt-get install -y gnuplot

# 2. Configure + build the benchmark target
cmake -S . -B build -DBUILD_TESTS=ON -DBUILD_PYTHON=OFF
cmake --build build --target route_planner_benchmark -j

# 3. Run it, writing CSV output alongside the baseline doc
./build/tests/cpp/route_planning/benchmarks/route_planner_benchmark \
    docs/perf/ortools-baseline.csv
```

Then paste the four CSV rows into the results table in
[`perf/ortools-baseline.md`](perf/ortools-baseline.md), fill in the Environment
table (CPU, cores, RAM, kernel, compiler, OR-Tools version), flip that doc's
STATUS from `pending` to `captured YYYY-MM-DD`, and return to this document to
change the Status line from "Provisional" to "Final". If the measured numbers
trip any of the "Conditions for revisit" bullets in the "If KEEP" section,
reopen the verdict instead of finalising it.
