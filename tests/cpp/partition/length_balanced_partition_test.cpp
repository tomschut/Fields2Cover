//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include <algorithm>
#include <numeric>
#include "fields2cover/types.h"
#include "fields2cover/partition/length_balanced_partition.h"

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

TEST(fields2cover_partition_length_balanced, two_robots) {
  F2CCells field = makeRect(20.0, 20.0);  // area = 400.0

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);

  f2c::partition::LengthBalancedPartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  // Both zones should have similar area (balanced load → balanced coverage area)
  double a0 = zones[0].area();
  double a1 = zones[1].area();
  double total = a0 + a1;
  EXPECT_GT(total, 0.0);
  // Neither zone should dominate — each robot gets a meaningful share
  EXPECT_GT(a0 / total, 0.30);
  EXPECT_GT(a1 / total, 0.30);
}

TEST(fields2cover_partition_length_balanced, three_robots) {
  F2CCells field = makeRect(30.0, 30.0);  // area = 900.0

  F2CRobot r1(3.0), r2(3.0), r3(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  r3.setCruiseVel(1.0);

  f2c::partition::LengthBalancedPartition part;
  auto zones = part.partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  EXPECT_GT(zones[2].area(), 0.0);
}

TEST(fields2cover_partition_length_balanced, empty_robots_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  f2c::partition::LengthBalancedPartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}

TEST(fields2cover_partition_length_balanced, zero_width_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  F2CRobot r_zero(1.0);
  r_zero.setCovWidth(0.0);  // set to zero after construction (constructor requires > 0)
  r_zero.setCruiseVel(1.0);
  f2c::partition::LengthBalancedPartition part;
  EXPECT_THROW(part.partition(field, {r_zero}), std::invalid_argument);
}
