//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_
#define FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

/// @brief Partition a field into N zones by greedy length-balanced swath assignment.
///
/// The field is tessellated into swaths using f2c::sg::BruteForce with
/// robots[0].getCovWidth() as the swath width. Swaths are sorted by length()
/// descending (first-fit decreasing). Each swath is greedily assigned to the
/// robot with the minimum current load (total swath-metres assigned so far).
/// Zone geometry is derived by unioning each robot's assigned swaths' areaCovered().
///
/// Assumptions:
///   - Input field must be a single-cell geometry (field.size() == 1).
///   - Swath width uses robots[0].getCovWidth().
class LengthBalancedPartition {
 public:
  /// @brief Partition a field into N length-balanced zones.
  ///
  /// @param field  Single-cell field geometry in any metric CRS.
  /// @param robots Non-empty robot list; result[i] is the zone for robots[i].
  /// @return       std::vector<F2CCells> of size robots.size().
  /// @throws std::invalid_argument if robots is empty, robots[0].getCovWidth() <= 0,
  ///         or field.size() != 1.
  std::vector<F2CCells> partition(
      const F2CCells& field,
      const std::vector<F2CRobot>& robots) const;
};

}  // namespace f2c::partition

#endif  // FIELDS2COVER_PARTITION_LENGTH_BALANCED_PARTITION_H_
