//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
#define FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

/// @brief Partition a field into N zones proportional to each robot's work rate.
///
/// Uses axis-aligned strip partitioning: the field's bounding box is divided
/// along its longer axis at cumulative work-rate fraction cut points. Each strip
/// is intersected with the actual field geometry (via GEOS) to produce the zone.
///
/// Work rate = robot.getCovWidth() * robot.getCruiseVel()
class MultiRobotPartition {
 public:
  /// @brief Partition a field into N zones proportional to each robot's work rate.
  ///
  /// @param field  Field geometry in any metric CRS; no CRS check is performed.
  /// @param robots Non-empty list of robots; result[i] is the zone for robots[i].
  /// @return       std::vector<F2CCells> of size robots.size().
  /// @throws std::invalid_argument if robots is empty or any work rate is <= 0.
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_MULTI_ROBOT_PARTITION_H_
