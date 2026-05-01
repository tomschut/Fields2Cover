//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/partition/multi_robot_partition.h"

namespace {

/// Helper: create a rectangular F2CCells with corners (0,0)-(w,h)
F2CCells makeRect(double w, double h) {
  F2CLinearRing ring{
    F2CPoint(0, 0), F2CPoint(w, 0),
    F2CPoint(w, h), F2CPoint(0, h),
    F2CPoint(0, 0)};
  return F2CCells{F2CCell{ring}};
}

}  // namespace

TEST(fields2cover_partition_multi_robot, two_robots_equal_rate) {
  F2CCells field = makeRect(10.0, 10.0);

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);  // equal work rates: 3.0 × 1.0 each

  f2c::partition::MultiRobotPartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_NEAR(zones[0].area() + zones[1].area(), field.area(), 1e-3);
  EXPECT_NEAR(zones[0].area(), zones[1].area(), 1e-3);
}

TEST(fields2cover_partition_multi_robot, three_robots_proportional) {
  F2CCells field = makeRect(12.0, 12.0);  // area = 144.0

  // work rates 2:1:1 → expected areas 72, 36, 36
  F2CRobot r1(2.0), r2(1.0), r3(1.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  r3.setCruiseVel(1.0);

  auto zones = f2c::partition::MultiRobotPartition().partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3u);
  EXPECT_NEAR(
      zones[0].area() + zones[1].area() + zones[2].area(),
      field.area(), 1e-3);
  EXPECT_NEAR(zones[0].area(), 2.0 * zones[1].area(), 1e-3);
  EXPECT_NEAR(zones[1].area(), zones[2].area(), 1e-3);
}

TEST(fields2cover_partition_multi_robot, empty_robots_throws) {
  F2CCells field = makeRect(10.0, 10.0);
  f2c::partition::MultiRobotPartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}

TEST(fields2cover_partition_multi_robot, zero_work_rate_throws) {
  F2CCells field = makeRect(10.0, 10.0);
  // Construct a valid robot, then force cov_width to 0 via setCovWidth.
  // getCovWidth() == 0 → work rate = 0 * cruise_vel == 0 → partition() throws.
  F2CRobot r_zero(1.0);
  r_zero.setCovWidth(0.0);
  r_zero.setCruiseVel(1.0);
  f2c::partition::MultiRobotPartition part;
  EXPECT_THROW(part.partition(field, {r_zero}), std::invalid_argument);
}
