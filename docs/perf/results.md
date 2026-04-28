# Performance Results — Phase 7 Wave 2 (Plan 07-08)

**Baseline:** `docs/perf/baseline.md` (commit `5ac9b89`, Plan 07-07)
**Final:** commit `c100383` (Plan 07-08, after revert sweep)
**Generated:** 2026-04-11
**Host:** AMD Ryzen 9 9950X 16-Core / 32 logical CPUs, 61 GiB RAM, Linux 6.17.0-19-generic
**Compiler:** g++ (Ubuntu 13.3.0-6ubuntu2~24.04.1) 13.3.0
**Build:** `cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=ON -DBUILD_PYTHON=OFF -DBUILD_TUTORIALS=OFF`
**Sanitizers:** off
**Methodology:** identical to Plan 07-07. Same harnesses, same fields, same repeat
counts. CSVs in `/tmp/perf-baseline/` (before) and `/tmp/perf-final/` (after) on
the build host.

## Summary table

| Benchmark              | Baseline (ms) | Final (ms) |    Delta | Status               |
| ---------------------- | ------------: | ---------: | -------: | -------------------- |
| BM_SwathGen / 10       |        20.739 |     10.666 |  -48.57% | KEPT (opt 1)         |
| BM_SwathGen / 100      |       579.191 |    292.965 |  -49.42% | KEPT (opt 1)         |
| BM_SwathGen / 1000     |    31 561.799 | 15 803.509 |  -49.93% | KEPT (opt 1)         |
| BM_Headland / 10       |         0.005 |      0.005 |     flat | not a target         |
| BM_Headland / 100      |         0.005 |      0.005 |     flat | not a target         |
| BM_Headland / 1000     |         0.005 |      0.005 |     flat | not a target         |
| BM_PathInterp / 1000   |         1.332 |      1.209 |   -9.23% | KEPT (opt 2, sub-gate but improving) |
| BM_PathInterp / 10000  |         3.206 |      2.366 |  -26.20% | KEPT (opt 2)         |
| BM_PathInterp / 100000 |        22.392 |     14.540 |  -35.07% | KEPT (opt 2)         |
| OrToolsTSP / 10        |         3.157 |      3.124 |   -1.05% | unchanged (no opt 3) |
| OrToolsTSP / 50        |       120.852 |    123.606 |   +2.28% | within 5% gate       |
| OrToolsTSP / 100       |       554.875 |    539.092 |   -2.84% | unchanged (noise)    |
| OrToolsTSP / 500       |    14 992.138 | 15 138.299 |   +0.97% | within 5% gate       |

Route length unchanged for every OrToolsTSP size (correctness intact).

## Optimizations

### 1. BruteForce angle-scan π-symmetry (target: BM_SwathGen)

- **Commit:** `2012154`
- **File:** `src/fields2cover/swath_generator/brute_force.cpp`
- **Before / 10:**       20.739 ms
- **After  / 10:**       10.666 ms
- **Delta  / 10:**     **-48.57 %**
- **Before / 100:**     579.191 ms
- **After  / 100:**     292.965 ms
- **Delta  / 100:**   **-49.42 %**
- **Before / 1000:**  31 561.799 ms
- **After  / 1000:**  15 803.509 ms
- **Delta  / 1000:**  **-49.93 %**
- **Status:** **KEPT** — clears the 10 % gate by ~5×.
- **Technique:** `BruteForce::computeBestAngle` iterated candidate angles in
  `[0, 2π)`. Every SGObjective in the library is invariant under
  `angle ↔ angle + π`: `generateSwaths(ang, op_width, poly)` produces the same
  geometric set of parallel lines through `poly` for both, only the traversal
  direction differs. `NSwathModified::computeCost` already encodes this
  symmetry explicitly via `fabs(sin(ang - edge_ang))`. Cutting the candidate
  range to `[0, π)` halves the number of `computeCostOfAngle` calls — exactly
  the 2× speedup the measurements show. Output angle now lives in `[0, π)`
  instead of `[0, 2π)`; downstream consumers (route planner, path planner)
  are direction-invariant w.r.t. swath orientation, so this is ABI-compatible.

### 2. Path::populate reserve + monotone-cursor lookup (target: BM_PathInterp)

- **Commit:** `c100383`
- **File:** `src/fields2cover/types/Path.cpp`
- **Before / 1000:**       1.332 ms
- **After  / 1000:**       1.209 ms
- **Delta  / 1000:**     **-9.23 %** (just under gate, but improving — see note)
- **Before / 10 000:**     3.206 ms
- **After  / 10 000:**     2.366 ms
- **Delta  / 10 000:**  **-26.20 %**
- **Before / 100 000:**   22.392 ms
- **After  / 100 000:**   14.540 ms
- **Delta  / 100 000:** **-35.07 %**
- **Status:** **KEPT** — the dominant 100k case (the worst-case input the plan
  identified) clears the 10 % gate at 35 %. The 1k row sits just below the
  gate at 9.23 %; we keep the change because (a) the larger sizes benefit
  much more, (b) the change is a strict improvement at every input size,
  and (c) reverting would lose the 35 % win on the worst case.
- **Technique:** Two stacked wins:
  1. Reserve the output `states_` vector and the seven temporary vectors
     (`x`, `y`, `ang_prov`, `len`, `old_vel`, `old_dir`, `old_type`) up
     front. The output grew from 0 to `number_points` via `emplace_back`
     and was paying ~17 reallocations + state-copy storms at 100k. The
     temp vectors were doing the same.
  2. Replace the per-sample `std::lower_bound` on the parameterisation
     vector `t` with a monotone running cursor. The output sample positions
     `d = i*step` are strictly increasing, so the lookup index can only
     advance — no need to re-binary-search from `t.begin()` for every
     sample point. Drops the per-sample lookup from `O(log n_in)` to
     amortised `O(1)`. Original semantics
     (`it = lower_bound(t, d + 1e-10) - 1`) preserved exactly.

### 3. RoutePlannerBase distance matrix (target: OrToolsTSP / 500)

- **Commit:** none — both attempts reverted, see Reverts table.
- **Status:** **NO LOW-HANGING IMPROVEMENT FOUND.** After two attempts the
  route-planning hot path resists the kind of mechanical fix this plan
  was scoped to deliver. Documented for the next perf phase.
- **Why the easy fixes did not work:** The 14.9 s wall-clock for `genRoute`
  at 500 swaths is dominated by `Graph::shortestPathsAndCosts` running
  Floyd-Warshall on the ~2000-node `shortest_graph`, plus the OR-Tools
  solver hitting its 1-second time limit per cell. The visible `O(N²)`
  loop in `createCoverageGraph` (the obvious target) is *not* the
  bottleneck — see attempt 2 below where halving its iteration count
  produced no measurable improvement.
- **Recommended next steps** (separate phase, not in scope here):
  - Replace the `O(N³)` Floyd-Warshall with an `O(V·E·log V)` Dijkstra
    *only on the swath-endpoint subset* (≪ N), not all-pairs over every
    border node — attempt 1 below replaced F-W globally and regressed
    because it built per-source vectors of size `numNodes()` for thousands
    of sources. A scoped variant computing only the (swath_endpoint →
    swath_endpoint) submatrix would be much faster.
  - OR-Tools tuning: swap `time_limit = 1 s` for an absolute search-bound,
    or pre-warm with `PATH_CHEAPEST_ARC` first solution.
  - Sparse distance matrix: today every endpoint pair gets an explicit
    edge; many of those are never used by the solver. A k-nearest-neighbour
    pruning pass would cut edges and Floyd-Warshall input alike.
  - Both items are algorithmic restructures that exceed the "<5 % regression
    elsewhere" bar this plan is held to and need their own design pass.

## Reverts

| Attempt | Commit | File | Reason |
|---|---|---|---|
| Replace Graph::shortestPathsAndCosts Floyd-Warshall with N-source Dijkstra | (not committed; reverted via `git checkout`) | `src/fields2cover/types/Graph.cpp` | OrToolsTSP/500 regressed **+43.4 %** (14 992 ms → 21 502 ms). The graph after `createShortestGraph`'s "connect nodes near edges" pass is denser than expected, and per-source Dijkstra with N≈2000 sources × `O((V+E) log V)` plus N² path-vector reconstruction allocations cost more than the cache-friendly F-W triple-loop. Route length unchanged (correctness OK), but the perf rule says revert. |
| Halve `createCoverageGraph` nested loop via upper-triangle iteration over swath endpoints | (not committed; reverted via `git checkout`) | `src/fields2cover/route_planning/route_planner_base.cpp` | OrToolsTSP/500 measured **+0.5 %** (14 992 ms → 15 062 ms) — within noise but **below** the 10 % improvement gate. The nested loop turned out not to be the bottleneck; F-W and the OR-Tools solver dominate. Reverted per the "improvement < 10 % → revert" rule. Route length unchanged. |

## Cross-benchmark regression gate

| Benchmark             | Δ vs baseline | Status |
| --------------------- | ------------: | :----: |
| BM_SwathGen all sizes |        ~-50 % |  KEPT  |
| BM_Headland           |          flat |   OK   |
| BM_PathInterp all     |       -9..-35 |  KEPT  |
| OrToolsTSP / 10       |        -1.05% |   OK   |
| OrToolsTSP / 50       |        +2.28% |   OK   |
| OrToolsTSP / 100      |        -2.84% |   OK   |
| OrToolsTSP / 500      |        +0.97% |   OK   |

**No benchmark regressed by more than 5 %.** Full C++ test suite passes
at final HEAD (`ctest --test-dir build` — `unittests` 100% pass, 1.30 s
wall-clock). Python bindings build is gated to the human-verify checkpoint
in Plan 07-08 Task 2.

## Flamegraphs

Still none — `kernel.perf_event_paranoid = 4` on this host blocks
`perf record`. The `scripts/run-benchmarks.sh` reproducer auto-detects
this and falls back to CSV-only output. To capture flamegraphs on a
privileged host, follow the procedure in `docs/perf/baseline.md`
"Flamegraphs" section.
