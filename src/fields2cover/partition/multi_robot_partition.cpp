//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/multi_robot_partition.h"

#include <numeric>
#include <stdexcept>

namespace f2c::partition {

std::vector<F2CCells> MultiRobotPartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "MultiRobotPartition::partition: robots must not be empty");
  }

  // 1. Compute work rate for each robot: getCovWidth() * getCruiseVel()
  std::vector<double> rates(robots.size());
  for (std::size_t i = 0; i < robots.size(); ++i) {
    rates[i] = robots[i].getCovWidth() * robots[i].getCruiseVel();
    if (rates[i] <= 0.0) {
      throw std::invalid_argument(
          "MultiRobotPartition::partition: all robots must have positive work rate "
          "(getCovWidth() * getCruiseVel() > 0)");
    }
  }
  const double total =
      std::accumulate(rates.begin(), rates.end(), 0.0);

  // 2. Read field bounding box
  const double x_min = field.getDimMinX();
  const double x_max = field.getDimMaxX();
  const double y_min = field.getDimMinY();
  const double y_max = field.getDimMaxY();
  const double width  = x_max - x_min;
  const double height = y_max - y_min;

  // 3. Cut along the longer bounding-box axis
  //    cut_along_x == true  → cut along X (vertical strips)
  //    cut_along_x == false → cut along Y (horizontal strips)
  const bool cut_along_x = (width >= height);

  // 4. Build N strips and intersect each with the field geometry
  std::vector<F2CCells> zones;
  zones.reserve(robots.size());

  double cum_frac = 0.0;
  double prev = cut_along_x ? x_min : y_min;
  const double lo = cut_along_x ? y_min : x_min;
  const double hi = cut_along_x ? y_max : x_max;
  const double span = cut_along_x ? width : height;

  for (std::size_t i = 0; i < robots.size(); ++i) {
    cum_frac += rates[i] / total;
    // Force the last edge to the exact boundary to avoid floating-point gaps
    const double next = (i == robots.size() - 1)
        ? (cut_along_x ? x_max : y_max)
        : (cut_along_x ? x_min : y_min) + cum_frac * span;

    // Build a rectangular strip cell
    F2CLinearRing ring;
    if (cut_along_x) {
      ring = F2CLinearRing{
          F2CPoint(prev, lo), F2CPoint(next, lo),
          F2CPoint(next, hi), F2CPoint(prev, hi),
          F2CPoint(prev, lo)};
    } else {
      ring = F2CLinearRing{
          F2CPoint(lo, prev), F2CPoint(hi, prev),
          F2CPoint(hi, next), F2CPoint(lo, next),
          F2CPoint(lo, prev)};
    }

    zones.push_back(field.intersection(F2CCell{ring}));
    prev = next;
  }

  return zones;
}

}  // namespace f2c::partition
