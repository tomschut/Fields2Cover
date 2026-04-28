//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 PRF-01: path interpolation benchmark (Plan 07-07)
//
//    Measures F2CPath::populate(N) — the state sampler used throughout the
//    library to discretise a planned path. Three input sizes (1k / 10k /
//    100k sample points) against a fixed path built from a Dubins plan.
//=============================================================================

#include <algorithm>
#include <chrono>
#include <cmath>
#include <cstdio>
#include <numeric>
#include <vector>

#include "fields2cover/types.h"
#include "fields2cover/path_planning/path_planning.h"
#include "fields2cover/path_planning/dubins_curves.h"
#include "common.h"

namespace {

double median(std::vector<double> xs) {
  std::sort(xs.begin(), xs.end());
  if (xs.empty()) return 0.0;
  const size_t n = xs.size();
  if (n % 2 == 1) return xs[n / 2];
  return 0.5 * (xs[n / 2 - 1] + xs[n / 2]);
}

double stddev(const std::vector<double>& xs) {
  if (xs.size() < 2) return 0.0;
  const double mean =
      std::accumulate(xs.begin(), xs.end(), 0.0) /
      static_cast<double>(xs.size());
  double acc = 0.0;
  for (double x : xs) { const double d = x - mean; acc += d * d; }
  return std::sqrt(acc / static_cast<double>(xs.size() - 1));
}

// Build a single realistic F2CPath via Dubins plan on a small rectangular
// field. This is the one-time setup; the benchmark then repeatedly
// re-populates a copy of it to isolate F2CPath::populate cost.
F2CPath buildSeedPath() {
  F2CRobot robot(2.0, 6.0);
  robot.setMinTurningRadius(2.0);
  robot.setMaxDiffCurv(0.1);
  F2CSwaths swaths = f2c_bench::makeSwaths(20);
  f2c::pp::PathPlanning path_planner;
  f2c::pp::DubinsCurves dubins;
  return path_planner.planPath(robot, swaths, dubins);
}

struct Row {
  int n_points = 0;
  double median_ms = 0.0;
  double stddev_ms = 0.0;
  size_t path_states = 0;
};

Row runOne(const F2CPath& seed, int n_points, int repeats) {
  std::vector<double> times;
  times.reserve(repeats);
  size_t last_states = 0;

  for (int i = 0; i < repeats; ++i) {
    F2CPath copy = seed;  // fresh state every run
    const auto t0 = std::chrono::high_resolution_clock::now();
    copy.populate(n_points);
    const auto t1 = std::chrono::high_resolution_clock::now();
    times.push_back(std::chrono::duration<double, std::milli>(t1 - t0).count());
    last_states = copy.size();
  }

  return Row{n_points, median(times), stddev(times), last_states};
}

}  // namespace

int main() {
  F2CPath seed = buildSeedPath();
  std::fprintf(stderr, "seed path length=%.3f m, states=%zu\n",
      seed.length(), seed.size());

  std::printf("benchmark,input,median_ms,stddev_ms,path_states\n");
  const std::vector<int> sizes = {1000, 10000, 100000};
  for (int n : sizes) {
    Row r = runOne(seed, n, 5);
    std::printf("BM_PathInterp,%d,%.3f,%.3f,%zu\n",
        r.n_points, r.median_ms, r.stddev_ms, r.path_states);
  }
  return 0;
}
