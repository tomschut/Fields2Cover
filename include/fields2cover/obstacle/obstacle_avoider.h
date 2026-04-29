//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#pragma once
#ifndef FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
#define FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_

#include <stdexcept>
#include "fields2cover/types.h"

namespace f2c::obstacle {

/// @brief Fragment swaths around an inflated polygon obstacle.
///
/// For each input swath, the obstacle is inflated by `safety_margin` metres,
/// the swath centre-line is clipped against the complement of the inflated
/// obstacle, and any residual segment shorter than kMinSegmentLength is dropped.
///
/// Geometry operations are delegated entirely to the f2c/OGR type layer —
/// no raw GEOS or OGRGeometry pointers are used.
class ObstacleAvoider {
 public:
  /// Minimum returned segment length (metres). Segments below this are dropped.
  static constexpr double kMinSegmentLength = 0.1;

  /// @brief Fragment swaths around an inflated polygon obstacle.
  ///
  /// @param swaths        Input swaths to fragment.
  /// @param obstacle      Obstacle polygon (F2CCell) in the same CRS as swaths.
  /// @param safety_margin Buffer distance (metres >= 0) to inflate the obstacle
  ///                      before clipping.
  /// @return F2CSwaths of residual segments; segments < kMinSegmentLength dropped.
  /// @throws std::invalid_argument if safety_margin < 0.
  F2CSwaths avoid(const F2CSwaths& swaths,
                  const F2CCell& obstacle,
                  double safety_margin) const;
};

}  // namespace f2c::obstacle

#endif  // FIELDS2COVER_OBSTACLE_OBSTACLE_AVOIDER_H_
