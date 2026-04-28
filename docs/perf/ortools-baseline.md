# OR-Tools Route Planner — Baseline Benchmark

**Phase:** 5 — Solver Evaluation
**Requirement:** SLV-01
**Harness:** `tests/cpp/route_planning/benchmarks/route_planner_benchmark`
**Source:** `tests/cpp/route_planning/benchmarks/ortools_benchmark.cpp`

## Methodology

- **Fields:** Square rectangular fields with 10, 50, 100, 500 parallel swaths.
  Swath width 3.0 m, headland width 3.0 m, square inner area.
  Generator: `f2c_bench::makeRectField(n, seed=42)` in `bench_fields.h`.
- **Solver under test:** `f2c::rp::RoutePlannerBase::genRoute` with defaults
  (OR-Tools `AUTOMATIC` first-solution strategy, `AUTOMATIC` metaheuristic,
  1-second time limit — see `src/fields2cover/route_planning/route_planner_base.cpp`
  `computeBestRoute`).
- **Repeats:** 5 runs per input size; report median ± standard deviation of
  wall-clock time (sample stddev, ddof = 1).
- **Metrics:**
  - `median_ms` — median wall-clock of `genRoute()` in milliseconds
  - `stddev_ms` — sample stddev of the 5 runs, ms
  - `route_length_m` — `F2CRoute::length()` on the final run, metres
  - `peak_rss_kb` — `getrusage(RUSAGE_SELF).ru_maxrss` after all runs, kilobytes
- **Hardware:** record CPU model, core count, RAM, kernel, compiler version,
  build type on whichever host actually runs the harness.

## Environment

| Field            | Value                    |
| ---------------- | ------------------------ |
| Host CPU         | _fill in_                |
| Cores            | _fill in_                |
| RAM              | _fill in_                |
| Kernel           | _fill in_                |
| Compiler         | _fill in_                |
| CMAKE_BUILD_TYPE | Release                  |
| OR-Tools version | _fill in from cmake log_ |

## Results

> **STATUS:** `pending — collect on a build host where Fields2Cover tests compile`
>
> The planning host configured cmake successfully (OR-Tools, GDAL, Eigen all
> resolved) but the `tests/` subdirectory is gated on `BUILD_TESTS AND
> GNUPLOT_FOUND` in the top-level `CMakeLists.txt` (line 169), and gnuplot is
> not installed on this host. As a result, `route_planner_benchmark` is not
> emitted as a build target on this host. See the diagnostics block below for
> the exact cmake output and the one-line fix.

| n_swaths | median_ms | stddev_ms | route_length_m | peak_rss_kb |
| -------: | --------: | --------: | -------------: | ----------: |
|       10 |       TBD |       TBD |            TBD |         TBD |
|       50 |       TBD |       TBD |            TBD |         TBD |
|      100 |       TBD |       TBD |            TBD |         TBD |
|      500 |       TBD |       TBD |            TBD |         TBD |

## Observations

_Populate once numbers are collected. Look for:_

- Linear / super-linear scaling with `n_swaths` (coverage graph edge count
  grows roughly as O(n²), so 10× more swaths should not be 10× the time —
  expect super-linear).
- Does the 1-second OR-Tools time limit bind at 500 swaths? Compare
  `median_ms` at n=500 to 1000 ms. If solver hits the wall, route length
  will also plateau and Phase 5 has evidence for the "augment/replace"
  verdict on large inputs.
- RSS growth pattern across n=10 → n=500.
- Stddev as a fraction of median — large stddev suggests OR-Tools
  metaheuristic nondeterminism rather than measurement noise.

## Reproducing

```bash
# 1. Install gnuplot (required by the f2c tests build guard)
sudo apt-get install -y gnuplot

# 2. Configure + build just the benchmark target
cmake -S . -B build -DBUILD_TESTS=ON -DBUILD_PYTHON=OFF
cmake --build build --target route_planner_benchmark -j

# 3. Run it, writing CSV output alongside this doc
./build/tests/cpp/route_planning/benchmarks/route_planner_benchmark \
    docs/perf/ortools-baseline.csv
```

The binary prints the same CSV to stdout; the positional argument just
mirrors it to a file.

## Build host notes (results pending)

<details>
<summary>cmake configure output from the planning host (2026-04-11)</summary>

```
-- Could NOT find Gnuplot (missing: GNUPLOT_EXECUTABLE)
...
-- Configuring done
-- Generating done
-- Build files have been written to: build-bench

$ cmake --build build-bench --target route_planner_benchmark -j
gmake: *** No rule to make target 'route_planner_benchmark'.  Stop.
```

**Diagnosis:** `CMakeLists.txt:169` guards the `add_subdirectory(tests)` call
with `if(BUILD_TESTS AND GNUPLOT_FOUND)`. Without gnuplot installed the whole
`tests/` tree is skipped, so neither `unittests` nor `route_planner_benchmark`
is emitted. Install gnuplot and reconfigure — no source changes needed.

The harness source itself was not built on this host. Before trusting any
numbers collected elsewhere, run a smoke check:

```bash
./build/tests/cpp/route_planning/benchmarks/route_planner_benchmark /tmp/smoke.csv
head /tmp/smoke.csv
# Expect 5 lines: header + n=10, n=50, n=100, n=500
```

</details>
