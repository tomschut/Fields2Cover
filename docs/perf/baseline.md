# Performance Baseline — Phase 7 (PRF-01)

**Plan:** 07-07
**Generated:** 2026-04-11
**Build:** `cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=ON -DBUILD_PYTHON=OFF -DBUILD_TUTORIALS=OFF`
**Library commit at measurement time:** `fd90c7b` (post Task 1 of Plan 07-07)
**Host:** AMD Ryzen 9 9950X 16-Core / 32 logical CPUs, 61 GiB RAM, Linux 6.17.0-19-generic x86_64
**Compiler:** g++ (Ubuntu 13.3.0-6ubuntu2~24.04.1) 13.3.0
**Sanitizers:** off

## Methodology

All harnesses are standalone `std::chrono::high_resolution_clock` timers
(no google-benchmark dependency) matching the style of the existing
`tests/cpp/route_planning/benchmarks/ortools_benchmark` from Phase 5. Each
bench reports median and sample stddev over N repeats per input size:

- **bench_swath**  — `f2c::sg::BruteForce::generateBestSwaths` with
  `f2c::obj::NSwath` on a mainland produced by `ConstHL::generateHeadlandArea`
  of a rectangular field. Inputs: 10 / 100 / 1000 linear swath count. Repeats:
  5, 5, 3.
- **bench_headland** — `f2c::hg::ConstHL::generateHeadlandArea` on the same
  rectangular fields. Inputs: 10 / 100 / 1000. Repeats: 5, 5, 3.
- **bench_path_interp** — `F2CPath::populate(N)` on a Dubins-planned seed
  path (20-swath field, `planPath(robot, swaths, dubins)` produces a
  1677.9 m path with 8956 raw states). Inputs: 1k / 10k / 100k sample points.
  Repeats: 5 each.
- **ortools_benchmark** — Phase 5 SLV-01 harness, reused unmodified.
  `f2c::rp::RoutePlannerBase::genRoute` with OR-Tools defaults on 10 / 50 /
  100 / 500 parallel swaths. Repeats: 5 each.

Reproducer: `BUILD_DIR=build ./scripts/run-benchmarks.sh`. Outputs land
in `docs/perf/*.csv`.

## Results

### Swath generation (`BM_SwathGen`)

| Input (linear swath count) | Median (ms) | Stddev (ms) | Swaths produced |
| -------------------------: | ----------: | ----------: | --------------: |
|                         10 |      20.364 |       0.398 |               6 |
|                        100 |     577.976 |       4.285 |              96 |
|                       1000 |  31 243.911 |       7.438 |             996 |
|                            |             |             |                 |

Scaling 10 → 100: **28.4×** (expected 10× linear). 100 → 1000: **54.1×**.
Super-linear: ~N^1.45 on this input family. This is the dominant hot path.

### Headland generation (`BM_Headland`)

| Input | Median (ms) | Stddev (ms) |
| ----: | ----------: | ----------: |
|    10 |       0.007 |       0.049 |
|   100 |       0.005 |       0.000 |
|  1000 |       0.005 |       0.000 |

Effectively flat at ~5 µs across all sizes. `generateHeadlandArea` for
rectangular fields is not a hot path — the GEOS buffer operation amortises
the geometry scan away. Non-goal for Plan 07-08.

### Path interpolation (`BM_PathInterp`)

| Input (sample points) | Median (ms) | Stddev (ms) | Final path states |
| --------------------: | ----------: | ----------: | ----------------: |
|                  1000 |       1.283 |       0.075 |              1000 |
|                10 000 |       3.130 |       0.088 |            10 000 |
|               100 000 |      23.118 |       0.984 |           100 000 |

Near-linear (~2.4× per 10× input) — dominated by spline evaluation + state
copy. Mid-value optimisation target.

### Route planning / TSP (`ortools_benchmark` — Phase 5 reuse)

| n_swaths | Median (ms) | Stddev (ms) | Route length (m) | Peak RSS (kB) |
| -------: | ----------: | ----------: | ---------------: | ------------: |
|       10 |       3.163 |       0.738 |              168 |        47 616 |
|       50 |     120.168 |       2.066 |            6 888 |        53 504 |
|      100 |     544.003 |       4.976 |           28 788 |        86 016 |
|      500 |  14 901.794 |     353.152 |          743 988 |       982 052 |

Matches the envelope observed in Phase 5 `ortools-baseline.md`. 500-swath
run hits the 1-second OR-Tools time limit internally but total wall-clock
is dominated by graph construction + distance matrix population, not the
solver loop. Confirmed hot path.

## Flamegraphs

`perf record` requires `kernel.perf_event_paranoid ≤ 2` or `CAP_PERFMON`;
on the build host used for this baseline the setting is `4`, so perf
sampling returns empty `*.data` files. As a consequence **no flamegraphs
were captured** for this baseline. The `scripts/run-benchmarks.sh`
reproducer will auto-enable flamegraph capture on any host where a
throwaway `perf record /bin/true` produces non-empty output AND the
`$FLAMEGRAPH` environment variable points at a FlameGraph checkout
(`stackcollapse-perf.pl`, `flamegraph.pl`).

To produce flamegraphs later (privileged host or container):

```bash
# 1. Lower paranoid OR run as root inside a container
sudo sysctl -w kernel.perf_event_paranoid=1
# 2. Point at FlameGraph
export FLAMEGRAPH=/opt/FlameGraph
# 3. Re-run
BUILD_DIR=build ./scripts/run-benchmarks.sh
```

The resulting SVGs will be written to `docs/perf/flamegraphs/<bench>.svg`
and the observed-hot-paths section below should be refreshed against
them.

## Observed hot paths (candidates for Plan 07-08)

Derived from the CSV tables above (no flamegraph — self-time inferred
from scaling and total cost):

1. **`f2c::sg::BruteForce::computeBestAngle` on 1000-cell fields**
   — 31.2 s wall-clock for a single call is the single biggest number in
   the entire baseline. The brute-force angle search is O(π / step_angle)
   = 180 candidate angles × O(swath intersection cost) per call. Plan
   07-08 target: cut the inner angle loop via early-rejection on the
   NSwath objective (best-so-far pruning), or vectorise the intersection
   tests. Budget: >10% improvement on BM_SwathGen/1000 with <5%
   regression elsewhere.
2. **`f2c::rp::RoutePlannerBase::genRoute` distance-matrix construction
   on 500-swath routes** — 14.9 s wall-clock, with `peak_rss_kb`
   approaching 1 GiB. The Phase 5 audit flagged the pairwise-distance
   pass as O(n²) in swath count. Plan 07-08 target: cache/deduplicate
   pairwise-distance calls, or switch to a banded distance matrix for
   locally-connected swath sets. Budget: same >10% / <5%.
3. **`F2CPath::populate` state-copy overhead at 100k samples** —
   23.1 ms for 100k states is dominated by the final `std::vector<State>`
   growth + per-state spline evaluation. Plan 07-08 target: reserve
   capacity up-front and amortise the spline segment lookup with a
   running index. Budget: same >10% / <5%.

These three are the targets for Plan 07-08 optimisations. `BM_Headland`
is explicitly NOT a target (already ≈5 µs on all sizes).

## Known limitations

- **No flamegraphs** on this baseline — host has perf locked down. See
  the "Flamegraphs" section for the refresh procedure.
- **Single-host numbers** — the Ryzen 9 9950X is a fast desktop part.
  Absolute numbers will be higher on CI hardware. Plan 07-08's regression
  gate should be defined as a percentage against a fresh baseline run on
  the CI host, not against the absolute values in this table.
- **`bench_swath` at n=1000 takes ≈31 s** — intentional (we want the
  hotspot visible). CI smoke runs should use `n=10,100` only.
- The headland timings are suspiciously flat; if Plan 07-08 investigates
  the headland pipeline it should first add an input family that actually
  exercises it (e.g. obstacle-laden polygons with curved boundaries).
