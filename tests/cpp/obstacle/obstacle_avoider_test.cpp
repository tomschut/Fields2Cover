//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/obstacle/obstacle_avoider.h"

namespace {

/// Helper: build a single straight horizontal swath along y=0, x in [x0, x1].
F2CSwaths makeStraightSwaths(double x0, double x1, double width = 3.0) {
  F2CLineString path{F2CPoint(x0, 0.0), F2CPoint(x1, 0.0)};
  F2CSwaths s;
  s.emplace_back(path, width);
  return s;
}

/// Helper: build a square F2CCell obstacle centred at (cx, cy) with half-side r.
F2CCell makeSquareObstacle(double cx, double cy, double r) {
  F2CLinearRing ring{
      F2CPoint(cx - r, cy - r), F2CPoint(cx + r, cy - r),
      F2CPoint(cx + r, cy + r), F2CPoint(cx - r, cy + r),
      F2CPoint(cx - r, cy - r)};
  return F2CCell{ring};
}

}  // namespace

// -----------------------------------------------------------------------------
// obstacle_splits_swath_into_two
// Swath: x in [0, 20], y=0.  Square obstacle centred at (10,0), half-side 2 m.
// The obstacle straddles the swath midpoint -> expect exactly 2 residual segments.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, obstacle_splits_swath_into_two) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  auto result = avoider.avoid(swaths, obstacle, 0.0);

  ASSERT_EQ(result.size(), 2u);
  EXPECT_GT(result[0].length(), 0.1);
  EXPECT_GT(result[1].length(), 0.1);
}

// -----------------------------------------------------------------------------
// safety_margin_enlarges_exclusion
// Adding a 1 m safety margin inflates the obstacle -> shorter residual segments.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, safety_margin_enlarges_exclusion) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  auto result_no_margin = avoider.avoid(swaths, obstacle, 0.0);
  auto result_with_margin = avoider.avoid(swaths, obstacle, 1.0);

  // Total covered length must decrease when margin is added.
  double len_no_margin = 0.0;
  for (size_t i = 0; i < result_no_margin.size(); ++i) {
    len_no_margin += result_no_margin[i].length();
  }
  double len_with_margin = 0.0;
  for (size_t i = 0; i < result_with_margin.size(); ++i) {
    len_with_margin += result_with_margin[i].length();
  }

  EXPECT_LT(len_with_margin, len_no_margin);
}

// -----------------------------------------------------------------------------
// short_segments_dropped
// Obstacle nearly covers the entire swath (x in [0.5, 19.5]) -> residuals < 0.1 m
// must be dropped, so result is empty.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, short_segments_dropped) {
  // Swath: x in [0, 20].  Obstacle covers [0.5, 19.5] -> residuals ~0.5 m each.
  // Add safety_margin=0.45 to shrink residuals to ~0.05 m (< 0.1 threshold).
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 9.5);  // covers x in [0.5, 19.5]

  f2c::obstacle::ObstacleAvoider avoider;
  // With safety_margin=0.45, residuals become ~0.05 m -> all dropped.
  auto result = avoider.avoid(swaths, obstacle, 0.45);

  EXPECT_EQ(result.size(), 0u);
}

// -----------------------------------------------------------------------------
// negative_margin_throws
// A negative safety_margin is invalid -- expect std::invalid_argument.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, negative_margin_throws) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(10.0, 0.0, 2.0);

  f2c::obstacle::ObstacleAvoider avoider;
  EXPECT_THROW(avoider.avoid(swaths, obstacle, -1.0), std::invalid_argument);
}

// -----------------------------------------------------------------------------
// no_obstacle_overlap_returns_full_swath
// Obstacle placed far from swath -> full swath is returned as one segment.
// -----------------------------------------------------------------------------
TEST(fields2cover_obstacle_avoider, no_obstacle_overlap_returns_full_swath) {
  auto swaths = makeStraightSwaths(0.0, 20.0);
  auto obstacle = makeSquareObstacle(100.0, 0.0, 2.0);  // far from swath

  f2c::obstacle::ObstacleAvoider avoider;
  auto result = avoider.avoid(swaths, obstacle, 0.0);

  ASSERT_EQ(result.size(), 1u);
  EXPECT_NEAR(result[0].length(), 20.0, 0.01);
}
