//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 5 SLV-01: OR-Tools baseline benchmark
//
//    Standalone main() that measures f2c::rp::RoutePlannerBase::genRoute
//    wall-clock time, total route length, and peak resident set size on
//    rectangular fields with 10/50/100/500 parallel swaths. Reports a CSV
//    row per input size with median and sample stddev of the run times.
//
//    NOT a unit test — this is a benchmark harness. Run manually:
//        ./build/tests/cpp/route_planning/benchmarks/route_planner_benchmark \
//            [optional-output.csv]
//=============================================================================

#include <algorithm>
#include <chrono>
#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <fstream>
#include <numeric>
#include <string>
#include <vector>

#include <sys/resource.h>

#include "fields2cover/route_planning/route_planner_base.h"
#include "bench_fields.h"

namespace {

struct RunStats {
  int n_swaths = 0;
  double median_ms = 0.0;
  double stddev_ms = 0.0;
  double route_length_m = 0.0;
  long peak_rss_kb = 0;
};

// Median of a copy of the input. Sorts in-place (by value).
double median(std::vector<double> xs) {
  std::sort(xs.begin(), xs.end());
  const size_t n = xs.size();
  if (n == 0) {
    return 0.0;
  }
  if (n % 2 == 1) {
    return xs[n / 2];
  }
  return 0.5 * (xs[n / 2 - 1] + xs[n / 2]);
}

// Sample standard deviation (ddof = 1).
double stddev(const std::vector<double>& xs) {
  if (xs.size() < 2) {
    return 0.0;
  }
  const double mean =
      std::accumulate(xs.begin(), xs.end(), 0.0) /
      static_cast<double>(xs.size());
  double acc = 0.0;
  for (double x : xs) {
    const double d = x - mean;
    acc += d * d;
  }
  return std::sqrt(acc / static_cast<double>(xs.size() - 1));
}

// Peak resident set size in kilobytes. On Linux, `ru_maxrss` is reported
// in KiB already; on macOS it is in bytes (not our target platform here).
long peak_rss_kb() {
  struct rusage ru{};
  getrusage(RUSAGE_SELF, &ru);
  return ru.ru_maxrss;
}

RunStats runOne(int n_swaths, int repeats) {
  auto field = f2c_bench::makeRectField(n_swaths, /*seed=*/42);

  std::vector<double> times_ms;
  times_ms.reserve(static_cast<size_t>(repeats));

  double last_length = 0.0;
  for (int i = 0; i < repeats; ++i) {
    f2c::rp::RoutePlannerBase planner;

    const auto t0 = std::chrono::high_resolution_clock::now();
    F2CRoute route = planner.genRoute(field.headland_cells, field.swaths);
    const auto t1 = std::chrono::high_resolution_clock::now();

    const double ms =
        std::chrono::duration<double, std::milli>(t1 - t0).count();
    times_ms.push_back(ms);
    last_length = route.length();
  }

  return RunStats{
      n_swaths,
      median(times_ms),
      stddev(times_ms),
      last_length,
      peak_rss_kb()
  };
}

void writeHeader(std::FILE* sink) {
  std::fprintf(sink,
      "n_swaths,median_ms,stddev_ms,route_length_m,peak_rss_kb\n");
}

void writeRow(std::FILE* sink, const RunStats& s) {
  std::fprintf(sink, "%d,%.3f,%.3f,%.3f,%ld\n",
      s.n_swaths, s.median_ms, s.stddev_ms, s.route_length_m, s.peak_rss_kb);
}

}  // namespace

int main(int argc, char** argv) {
  // Input sizes per plan 05-01 / SLV-01.
  const std::vector<int> sizes = {10, 50, 100, 500};
  const int repeats = 5;

  writeHeader(stdout);

  // Optional CSV output path.
  std::string out_path;
  if (argc > 1) {
    out_path = argv[1];
  }
  std::ofstream out;
  if (!out_path.empty()) {
    out.open(out_path);
    if (out.is_open()) {
      out << "n_swaths,median_ms,stddev_ms,route_length_m,peak_rss_kb\n";
    } else {
      std::fprintf(stderr,
          "warning: could not open %s for writing; stdout only\n",
          out_path.c_str());
    }
  }

  for (int n : sizes) {
    const RunStats s = runOne(n, repeats);
    writeRow(stdout, s);
    if (out.is_open()) {
      out << s.n_swaths << ',' << s.median_ms << ',' << s.stddev_ms
          << ',' << s.route_length_m << ',' << s.peak_rss_kb << '\n';
    }
  }

  return 0;
}
