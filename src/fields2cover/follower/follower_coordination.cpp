//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/follower/follower_coordination.h"

namespace f2c::follower {

FollowerCoordination::Result FollowerCoordination::compute(
    const F2CPath& robot_path, const FollowerSpec& spec) const {
  if (spec.tank_capacity_swath_m <= 0.0) {
    throw std::invalid_argument(
        "FollowerCoordination::compute: tank_capacity_swath_m must be > 0");
  }

  Result result;

  if (robot_path.size() == 0) {
    return result;  // empty path -> empty result, no crash
  }

  double accumulated = 0.0;

  for (size_t i = 0; i < robot_path.size(); ++i) {
    const auto& state = robot_path[i];

    if (state.type == f2c::types::PathSectionType::SWATH) {
      accumulated += state.len;
    }

    if (accumulated >= spec.tank_capacity_swath_m) {
      // Scan forward for the next HL_SWATH to snap the rendezvous point.
      bool found = false;
      for (size_t j = i + 1; j < robot_path.size(); ++j) {
        if (robot_path[j].type == f2c::types::PathSectionType::HL_SWATH) {
          const F2CPoint& rv = robot_path[j].point;
          result.rendezvous_pts.push_back(rv);
          result.path.addPoint(rv);
          accumulated = 0.0;
          i = j;  // advance outer loop past rendezvous (for-loop will +1)
          found = true;
          break;
        }
      }
      if (!found) {
        // No HL_SWATH remains -- use the last path point as fallback rendezvous.
        const F2CPoint& rv = robot_path.back().point;
        result.rendezvous_pts.push_back(rv);
        result.path.addPoint(rv);
        accumulated = 0.0;
        // No index advance needed; the outer loop will exhaust naturally.
        break;
      }
    }
  }

  return result;
}

}  // namespace f2c::follower
