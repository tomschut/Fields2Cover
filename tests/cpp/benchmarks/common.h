//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 PRF-01: benchmark helpers for Plan 07-07
//=============================================================================

#pragma once

#include <cmath>
#include <cstdint>

#include "fields2cover/types.h"
#include "fields2cover/headland_generator/constant_headland.h"
#include "fields2cover/swath_generator/brute_force.h"

namespace f2c_bench {

// Build a rectangular outer field with a deterministic size that scales
// roughly with `n_cells`. Because Fields2Cover's swath/headland pipeline
// operates on a single F2CCells container, we use `n_cells` to control the
// size of the mainland area (and thus the number of internal swaths the
// brute-force generator must create). This is the same "bigger input ->
// more work" knob used by the Phase 5 ortools benchmark.
inline F2CCells makeRectCells(int n_cells) {
  const double swath_w = 3.0;
  const double headland_w = 3.0;
  // n_cells acts as the linear swath count per side of the square field.
  const double inner = static_cast<double>(n_cells) * swath_w;
  const double W = inner + 2.0 * headland_w;
  const double H = inner + 2.0 * headland_w;
  return F2CCells{
    F2CCell(F2CLinearRing({
      F2CPoint(0, 0), F2CPoint(W, 0),
      F2CPoint(W, H), F2CPoint(0, H), F2CPoint(0, 0)
    }))
  };
}

// Pre-built mainland (post-headland) area for swath benchmarks that want to
// measure only the swath step.
inline F2CCells makeMainland(int n_cells) {
  F2CCells cells = makeRectCells(n_cells);
  f2c::hg::ConstHL const_hl;
  return const_hl.generateHeadlandArea(cells, 3.0, 3);
}

// Pre-built swaths (single-cell brute force) for path-interp benchmarks.
inline F2CSwaths makeSwaths(int n_cells) {
  F2CCells mainland = makeMainland(n_cells);
  f2c::sg::BruteForce bf;
  F2CSwathsByCells by_cells = bf.generateSwaths(M_PI / 2.0, 3.0, mainland);
  return by_cells.flatten();
}

}  // namespace f2c_bench
