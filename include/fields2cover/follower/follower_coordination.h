//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_FOLLOWER_FOLLOWER_COORDINATION_H_
#define FIELDS2COVER_FOLLOWER_FOLLOWER_COORDINATION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::follower {

/// @brief Compute a follower cart's travel path and headland rendezvous points.
///
/// The algorithm iterates the robot coverage path (F2CPath), accumulating
/// distance from SWATH states only (what the robot is harvesting). When the
/// accumulated distance reaches tank_capacity_swath_m, the next HL_SWATH state
/// is taken as the rendezvous location. The follower path is a F2CLineString
/// connecting the rendezvous points in order.
///
/// Tank capacity units: swath-metres (accumulated SWATH segment .len).
/// Phase 26 performs kg <-> swath-metre conversion before calling this function.
class FollowerCoordination {
 public:
  /// @brief Follower cart specifications.
  struct FollowerSpec {
    double tank_capacity_swath_m;  ///< Swath-metres the tank holds before unloading.
    double unload_time_s;          ///< Seconds spent at each rendezvous (informational; unused in geometry).
    double follower_speed_mps;     ///< Cart travel speed in m/s (informational; unused in geometry).
  };

  /// @brief Output of compute().
  struct Result {
    F2CLineString path;                   ///< Follower travel path (polyline connecting rendezvous points).
    std::vector<F2CPoint> rendezvous_pts; ///< Rendezvous coordinates, one per unload event.
  };

  /// @brief Compute follower path and rendezvous points from a robot path.
  ///
  /// @param robot_path  Robot's coverage path (output of path planning).
  /// @param spec        Follower specifications.
  /// @return            Result with follower path geometry and rendezvous list.
  /// @throws std::invalid_argument if spec.tank_capacity_swath_m <= 0.
  Result compute(const F2CPath& robot_path, const FollowerSpec& spec) const;
};

}  // namespace f2c::follower

#endif  // FIELDS2COVER_FOLLOWER_FOLLOWER_COORDINATION_H_
