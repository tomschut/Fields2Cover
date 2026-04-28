//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#include <algorithm>
#ifdef ALLOW_PARALLELIZATION
#include <execution>
#endif
#include <vector>
#include "fields2cover/swath_generator/brute_force.h"

namespace f2c::sg {

double BruteForce::getStepAngle() const {
  return this->step_angle;
}

void BruteForce::setStepAngle(double d) {
  this->step_angle = d;
}

double BruteForce::computeBestAngle(f2c::obj::SGObjective& obj,
    double op_width, const F2CCell& poly) {
  // Angles `a` and `a + π` produce the *same* set of parallel swath lines
  // (just traversed in opposite direction). Every SGObjective in the library
  // depends only on the geometric set of swaths through `poly`, not on the
  // direction of travel — so candidate angles in [π, 2π) are guaranteed to
  // tie with their counterparts in [0, π) and we can skip them outright.
  // This halves the number of computeCostOfAngle calls (the per-angle
  // cost dominates BM_SwathGen at large n_cells).
  int n = static_cast<int>(
      boost::math::constants::pi<double>() / step_angle);
  std::vector<double> costs(n);
  std::vector<int> ids(n);
  std::iota(ids.begin(), ids.end(), 0);

  auto getCostSwaths = [this, op_width, &poly, &obj] (const int& i) {
    return computeCostOfAngle(obj, i * step_angle, op_width, poly);
  };

  #ifdef ALLOW_PARALLELIZATION
    std::transform(std::execution::par_unseq, ids.begin(), ids.end(),
        costs.begin(), getCostSwaths);
  #else
    std::transform(ids.begin(), ids.end(), costs.begin(), getCostSwaths);
  #endif

  return ids[std::min_element(
      costs.begin(), costs.end()) - costs.begin()] * step_angle;
}

}  // namespace f2c::sg

