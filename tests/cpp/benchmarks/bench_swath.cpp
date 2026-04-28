//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 PRF-01: swath generation benchmark (Plan 07-07)
//
//    Measures f2c::sg::BruteForce::generateSwaths on rectangular mainlands
//    whose size scales with n_cells (10 / 100 / 1000 linear swath count).
//    Standalone std::chrono harness (no google-benchmark dependency) to
//    match the style of tests/cpp/route_planning/benchmarks/ortools_benchmark.
//=============================================================================

#include <algorithm>
#include <chrono>
#include <cmath>
#include <cstdio>
#include <numeric>
#include <string>
#include <vector>

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/n_swath.h"
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

struct Row {
  int n_cells = 0;
  double median_ms = 0.0;
  double stddev_ms = 0.0;
  size_t n_swaths = 0;
};

Row runOne(int n_cells, int repeats) {
  F2CCells mainland = f2c_bench::makeMainland(n_cells);

  std::vector<double> times;
  times.reserve(repeats);
  size_t last_count = 0;

  for (int i = 0; i < repeats; ++i) {
    f2c::sg::BruteForce bf;
    f2c::obj::NSwath obj;
    const auto t0 = std::chrono::high_resolution_clock::now();
    auto swaths = bf.generateBestSwaths(obj, 3.0, mainland);
    const auto t1 = std::chrono::high_resolution_clock::now();
    times.push_back(std::chrono::duration<double, std::milli>(t1 - t0).count());
    last_count = swaths.sizeTotal();
  }

  return Row{n_cells, median(times), stddev(times), last_count};
}

}  // namespace

int main() {
  std::printf("benchmark,input,median_ms,stddev_ms,n_swaths\n");
  const std::vector<int> sizes = {10, 100, 1000};
  for (int n : sizes) {
    // Scale repeats down for 1000: it's already expensive.
    const int repeats = (n <= 100) ? 5 : 3;
    Row r = runOne(n, repeats);
    std::printf("BM_SwathGen,%d,%.3f,%.3f,%zu\n",
        r.n_cells, r.median_ms, r.stddev_ms, r.n_swaths);
  }
  return 0;
}
