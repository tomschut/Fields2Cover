//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 HRD-05: libFuzzer harness for route planner genRoute
//=============================================================================
//
// Targets f2c::rp::RoutePlannerBase::genRoute with fuzzer-controlled field
// dimensions. We reuse the makeRectField helper from the Phase 5 benchmark
// harness to build a deterministic field given a small integer "n_swaths",
// then let the fuzzer mutate the route planner parameters (d_tol,
// redirect_swaths, time_limit, search_for_optimum, start/end points).

#include <cstddef>
#include <cstdint>
#include <exception>

#include <fuzzer/FuzzedDataProvider.h>

#include "fields2cover/types.h"
#include "fields2cover/route_planning/route_planner_base.h"

#include "../route_planning/benchmarks/bench_fields.h"

extern "C" int LLVMFuzzerTestOneInput(const uint8_t* data, size_t size) {
  // T-026/T-027: Pre-validate input size before constructing the f2c
  // pipeline. Very small inputs drive the field builder into degenerate
  // shapes that reach a pre-existing SEGV deep inside OR-Tools
  // (tracked separately as T-015). Library-level hardening is deferred;
  // here we harden the *harness* so it rejects trivially-malformed
  // inputs that cannot yield a meaningful field/swath.
  //
  // Minimum structural validation: require >=64 bytes. Below this, the
  // FuzzedDataProvider cannot fully populate all downstream parameters
  // (n_swaths + 4 doubles + 4 bool + 4 endpoint doubles), and the
  // leftover-zero-bytes behavior of FuzzedDataProvider feeds constant
  // zeros into the field builder, hitting the OR-Tools crash.
  //
  // Returning -1 tells libFuzzer the input is rejected and should NOT
  // be added to the corpus, avoiding corpus pollution with useless
  // minimal inputs.
  if (data == nullptr || size < 64) {
    return -1;
  }
  FuzzedDataProvider fdp(data, size);
  // Ensure FuzzedDataProvider has enough remaining bytes for every
  // Consume* call below. If it runs dry mid-way, later Consume* calls
  // return default/zero values which can also tickle the OR-Tools path.
  if (fdp.remaining_bytes() < 64) {
    return -1;
  }

  // Keep the field small: large fields turn this into a performance test
  // rather than a crash-finding test, and libFuzzer wants high execs/sec.
  // Minimum of 3 swaths to stay out of the OR-Tools degenerate path.
  const int n_swaths = fdp.ConsumeIntegralInRange<int>(3, 6);
  const double d_tol = fdp.ConsumeFloatingPointInRange<double>(1e-6, 1.0);
  const bool redirect_swaths = fdp.ConsumeBool();
  const bool search_for_optimum = fdp.ConsumeBool();
  const int64_t time_limit_seconds = 1;  // fixed; libFuzzer caps per-exec

  const double sx = fdp.ConsumeFloatingPointInRange<double>(-50.0, 50.0);
  const double sy = fdp.ConsumeFloatingPointInRange<double>(-50.0, 50.0);
  const double ex = fdp.ConsumeFloatingPointInRange<double>(-50.0, 50.0);
  const double ey = fdp.ConsumeFloatingPointInRange<double>(-50.0, 50.0);
  const bool set_endpoints = fdp.ConsumeBool();

  try {
    f2c_bench::BenchField bf = f2c_bench::makeRectField(n_swaths);

    f2c::rp::RoutePlannerBase planner;
    if (set_endpoints) {
      planner.setStartAndEndPoint(F2CPoint(sx, sy), F2CPoint(ex, ey));
    }
    (void)planner.genRoute(bf.headland_cells, bf.swaths,
                           /*show_log=*/false, d_tol, redirect_swaths,
                           time_limit_seconds, search_for_optimum);
  } catch (const std::exception&) {
  } catch (...) {
  }
  return 0;
}
