//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/follower/follower_coordination.h"

namespace {

using FC = f2c::follower::FollowerCoordination;

/// Build a synthetic path with a repeating pattern of:
///   SWATH(swath_len) -> TURN(3.0m) -> HL_SWATH(2.0m)
/// The HL_SWATH start point is placed at (swath_x, 3.0) to make it
/// distinguishable from TURN points (placed at (swath_x, 0.0)).
F2CPath buildCyclicPath(double swath_len, int n_cycles,
                        double start_x = 0.0) {
  F2CPath p;
  double x = start_x;
  for (int i = 0; i < n_cycles; ++i) {
    // SWATH: robot covers the field
    p.addState(F2CPoint(x, 0.0), 0.0, swath_len,
               f2c::types::PathDirection::FORWARD,
               f2c::types::PathSectionType::SWATH, 1.0);
    x += swath_len;
    // TURN: robot turns on the headland
    p.addState(F2CPoint(x, 0.0), M_PI_2, 3.0,
               f2c::types::PathDirection::FORWARD,
               f2c::types::PathSectionType::TURN, 0.5);
    // HL_SWATH: headland swath -- valid rendezvous location
    p.addState(F2CPoint(x, 3.0), M_PI, 2.0,
               f2c::types::PathDirection::FORWARD,
               f2c::types::PathSectionType::HL_SWATH, 0.8);
    x += 2.0;
  }
  return p;
}

}  // namespace

TEST(fields2cover_follower_coordination, empty_path_returns_empty_result) {
  F2CPath empty_path;
  FC::FollowerSpec spec{10.0, 60.0, 3.0};
  FC coord;
  auto result = coord.compute(empty_path, spec);
  EXPECT_TRUE(result.rendezvous_pts.empty());
  EXPECT_EQ(result.path.size(), 0u);
}

TEST(fields2cover_follower_coordination, zero_capacity_throws) {
  F2CPath path = buildCyclicPath(10.0, 2);
  FC::FollowerSpec spec{0.0, 60.0, 3.0};
  FC coord;
  EXPECT_THROW(coord.compute(path, spec), std::invalid_argument);
}

TEST(fields2cover_follower_coordination, negative_capacity_throws) {
  F2CPath path = buildCyclicPath(10.0, 2);
  FC::FollowerSpec spec{-5.0, 60.0, 3.0};
  FC coord;
  EXPECT_THROW(coord.compute(path, spec), std::invalid_argument);
}

TEST(fields2cover_follower_coordination, large_tank_no_rendezvous) {
  // 3 cycles -> total SWATH = 3 x 10m = 30m; tank = 100m -> no rendezvous
  F2CPath path = buildCyclicPath(10.0, 3);
  FC::FollowerSpec spec{100.0, 60.0, 3.0};
  FC coord;
  auto result = coord.compute(path, spec);
  EXPECT_EQ(result.rendezvous_pts.size(), 0u);
}

TEST(fields2cover_follower_coordination, small_tank_triggers_rendezvous) {
  // 4 cycles -> total SWATH = 4 x 10m = 40m; tank = 15m
  // After swath1 (10m): accumulated=10 < 15 -> no trigger
  // After swath2 (10m): accumulated=20 >= 15 -> rendezvous, reset
  // After swath3 (10m): accumulated=10 < 15 -> no trigger
  // After swath4 (10m): accumulated=20 >= 15 -> rendezvous, reset
  // Expected: 2 rendezvous points
  F2CPath path = buildCyclicPath(10.0, 4);
  FC::FollowerSpec spec{15.0, 60.0, 3.0};
  FC coord;
  auto result = coord.compute(path, spec);
  EXPECT_EQ(result.rendezvous_pts.size(), 2u);
  // Follower path LineString has one point per rendezvous
  EXPECT_EQ(result.path.size(), 2u);
}

TEST(fields2cover_follower_coordination, capacity_scaling) {
  // Same path; smaller capacity must produce more rendezvous
  F2CPath path = buildCyclicPath(10.0, 4);
  FC coord;
  auto result_large = coord.compute(path, FC::FollowerSpec{15.0, 60.0, 3.0});
  auto result_small = coord.compute(path, FC::FollowerSpec{5.0, 60.0, 3.0});
  EXPECT_GT(result_small.rendezvous_pts.size(),
            result_large.rendezvous_pts.size());
}

TEST(fields2cover_follower_coordination, rendezvous_on_headland_not_turn) {
  // SWATH(10m) -> TURN -> HL_SWATH; tank=8m -> rendezvous after first swath
  // HL_SWATH point is at (x, 3.0); TURN point is at (x, 0.0) -- must pick HL_SWATH
  F2CPath path = buildCyclicPath(10.0, 2);
  FC::FollowerSpec spec{8.0, 60.0, 3.0};
  FC coord;
  auto result = coord.compute(path, spec);
  ASSERT_GE(result.rendezvous_pts.size(), 1u);
  // HL_SWATH start point Y == 3.0; TURN start point Y == 0.0
  EXPECT_NEAR(result.rendezvous_pts[0].Y(), 3.0, 1e-9);
}
