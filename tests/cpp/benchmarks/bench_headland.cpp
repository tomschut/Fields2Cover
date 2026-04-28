//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 PRF-01: headland generation benchmark (Plan 07-07)
//
//    Measures f2c::hg::ConstHL::generateHeadlandArea on rectangular fields
//    whose size scales with n_cells (10 / 100 / 1000). Standalone
//    std::chrono harness.
//=============================================================================

#include <algorithm>
#include <chrono>
#include <cmath>
#include <cstdio>
#include <numeric>
#include <vector>

#include "fields2cover/headland_generator/constant_headland.h"
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
};

Row runOne(int n_cells, int repeats) {
  F2CCells cells = f2c_bench::makeRectCells(n_cells);

  std::vector<double> times;
  times.reserve(repeats);

  for (int i = 0; i < repeats; ++i) {
    f2c::hg::ConstHL hl;
    const auto t0 = std::chrono::high_resolution_clock::now();
    auto mainland = hl.generateHeadlandArea(cells, 3.0, 3);
    const auto t1 = std::chrono::high_resolution_clock::now();
    (void) mainland;
    times.push_back(std::chrono::duration<double, std::milli>(t1 - t0).count());
  }

  return Row{n_cells, median(times), stddev(times)};
}

}  // namespace

int main() {
  std::printf("benchmark,input,median_ms,stddev_ms\n");
  const std::vector<int> sizes = {10, 100, 1000};
  for (int n : sizes) {
    const int repeats = (n <= 100) ? 5 : 3;
    Row r = runOne(n, repeats);
    std::printf("BM_Headland,%d,%.3f,%.3f\n",
        r.n_cells, r.median_ms, r.stddev_ms);
  }
  return 0;
}
