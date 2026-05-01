//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/length_balanced_partition.h"

#include <algorithm>
#include <numeric>
#include <vector>

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/swath_length.h"

namespace f2c::partition {

std::vector<F2CCells> LengthBalancedPartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: robots must not be empty");
  }
  if (robots[0].getCovWidth() <= 0.0) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: robots[0] must have positive coverage width");
  }
  if (field.size() != 1) {
    throw std::invalid_argument(
        "LengthBalancedPartition::partition: field must be a single-cell geometry (field.size() == 1)");
  }

  const std::size_t N = robots.size();

  // 1. Generate swaths from the single-cell field.
  f2c::sg::BruteForce sw_gen;
  f2c::obj::SwathLength obj;
  F2CSwaths swaths = sw_gen.generateBestSwaths(obj, robots[0].getCovWidth(), field.getCell(0));

  // 2. Sort swath indices by length() descending (first-fit decreasing).
  std::vector<std::size_t> order(swaths.size());
  std::iota(order.begin(), order.end(), 0);
  std::sort(order.begin(), order.end(), [&](std::size_t a, std::size_t b) {
    return swaths[a].length() > swaths[b].length();
  });

  // 3. Greedy min-load assignment.
  std::vector<double> load(N, 0.0);
  std::vector<std::vector<std::size_t>> assigned(N);

  for (std::size_t idx : order) {
    auto it = std::min_element(load.begin(), load.end());
    std::size_t robot_idx = static_cast<std::size_t>(
        std::distance(load.begin(), it));
    assigned[robot_idx].push_back(idx);
    load[robot_idx] += swaths[idx].length();
  }

  // 4. Build zone geometry: union each robot's swaths' areaCovered().
  std::vector<F2CCells> zones;
  zones.reserve(N);
  for (std::size_t r = 0; r < N; ++r) {
    F2CCells zone;
    for (std::size_t idx : assigned[r]) {
      F2CCells covered = swaths[idx].areaCovered();
      zone = zone.unionOp(covered);
    }
    zones.push_back(zone);
  }

  return zones;
}

}  // namespace f2c::partition
