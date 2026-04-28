//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 5 SLV-01: benchmark field generator
//=============================================================================

#pragma once

#include <cmath>
#include <cstdint>

#include "fields2cover/types.h"
#include "fields2cover/headland_generator/constant_headland.h"
#include "fields2cover/swath_generator/brute_force.h"

namespace f2c_bench {

struct BenchField {
  F2CCells headland_cells;     // inner headland ring used as `cells` arg to genRoute
  F2CSwathsByCells swaths;     // covers the inner area
};

// Build a rectangular field with approximately n_swaths parallel swaths.
// Deterministic given the seed: field dimensions and swath width are fixed
// by n_swaths (we scale the rectangle so that the inner area has exactly
// n_swaths swaths of width 3.0 m).
//
// The seed parameter is reserved for future perturbation of corner points;
// unused in v1 but accepted to document the intent and keep the API stable.
inline BenchField makeRectField(int n_swaths, uint32_t /*seed*/ = 42) {
  const double swath_w = 3.0;
  const double headland_w = 3.0;

  // Inner usable width = n_swaths * swath_w. Add 2*headland_w on each side.
  const double inner_w = static_cast<double>(n_swaths) * swath_w;
  const double inner_h = inner_w;  // square inner area
  const double W = inner_w + 2.0 * headland_w;
  const double H = inner_h + 2.0 * headland_w;

  F2CCells cells {
    F2CCell(F2CLinearRing({
      F2CPoint(0, 0), F2CPoint(W, 0),
      F2CPoint(W, H), F2CPoint(0, H), F2CPoint(0, 0)
    }))
  };

  f2c::hg::ConstHL const_hl;
  F2CCells no_hl = const_hl.generateHeadlandArea(cells, swath_w, 3);
  auto hl_swaths = const_hl.generateHeadlandSwaths(cells, swath_w, 3, false);

  f2c::sg::BruteForce bf;
  F2CSwathsByCells swaths = bf.generateSwaths(M_PI / 2.0, swath_w, no_hl);

  // hl_swaths[1] is the inner headland ring used as the travel surface for
  // genRoute (matches the pattern in route_planner_base_test.cpp, which
  // requests 3 headland rings so index [1] is always valid).
  return BenchField{hl_swaths[1], swaths};
}

}  // namespace f2c_bench
