//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
#define FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_

#include <vector>
#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::partition {

/// @brief Partition a field into N zones by spatial proximity clustering of swaths.
///
/// The field is first tessellated into swaths using f2c::sg::BruteForce with
/// robots[0].getCovWidth() as the swath width. N seed swaths (evenly spaced by
/// index) are selected — one per robot. Each remaining swath is assigned to the
/// robot whose current cluster centroid is nearest. Zone geometry is derived by
/// unioning each robot's assigned swaths' areaCovered() results.
///
/// Assumptions:
///   - Input field must be a single-cell geometry (field.size() == 1).
///   - All robots are assumed to have the same coverage width; swath generation
///     uses robots[0].getCovWidth().
///
/// boost::geometry R-tree headers are included only in the .cpp file to avoid
/// compile-time cost and include-path contamination in downstream translation units.
class SpatialRtreePartition {
 public:
  /// @brief Partition a field into N spatially compact zones.
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

#endif  // FIELDS2COVER_PARTITION_SPATIAL_RTREE_PARTITION_H_
