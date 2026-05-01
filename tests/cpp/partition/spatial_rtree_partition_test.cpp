//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include <stdexcept>
#include "fields2cover/types.h"
#include "fields2cover/partition/spatial_rtree_partition.h"

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

TEST(fields2cover_partition_spatial_rtree, two_robots) {
  F2CCells field = makeRect(20.0, 20.0);  // area = 400.0

  F2CRobot r1(3.0), r2(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);

  f2c::partition::SpatialRtreePartition part;
  auto zones = part.partition(field, {r1, r2});

  ASSERT_EQ(zones.size(), 2u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  // Total swath coverage should be within 5% of field area
  double total = zones[0].area() + zones[1].area();
  EXPECT_NEAR(total, field.area(), field.area() * 0.05);
}

TEST(fields2cover_partition_spatial_rtree, three_robots) {
  F2CCells field = makeRect(30.0, 30.0);  // area = 900.0

  F2CRobot r1(3.0), r2(3.0), r3(3.0);
  r1.setCruiseVel(1.0);
  r2.setCruiseVel(1.0);
  r3.setCruiseVel(1.0);

  f2c::partition::SpatialRtreePartition part;
  auto zones = part.partition(field, {r1, r2, r3});

  ASSERT_EQ(zones.size(), 3u);
  EXPECT_GT(zones[0].area(), 0.0);
  EXPECT_GT(zones[1].area(), 0.0);
  EXPECT_GT(zones[2].area(), 0.0);
}

TEST(fields2cover_partition_spatial_rtree, empty_robots_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  f2c::partition::SpatialRtreePartition part;
  EXPECT_THROW(part.partition(field, {}), std::invalid_argument);
}

TEST(fields2cover_partition_spatial_rtree, zero_width_throws) {
  F2CCells field = makeRect(20.0, 20.0);
  F2CRobot r_zero(1.0);
  r_zero.setCovWidth(0.0);  // set to zero after construction (constructor requires > 0)
  r_zero.setCruiseVel(1.0);
  f2c::partition::SpatialRtreePartition part;
  EXPECT_THROW(part.partition(field, {r_zero}), std::invalid_argument);
}
