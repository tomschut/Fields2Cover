//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#include <gtest/gtest.h>
#include "fields2cover/types/Cell.h"
#include "fields2cover/types/Cells.h"
#include "fields2cover/types/LinearRing.h"
#include "fields2cover/types/LineString.h"
#include "fields2cover/types/MultiLineString.h"
#include "fields2cover/types/MultiPoint.h"
#include "fields2cover/types/Point.h"

// T-002 (M-001) regression: verify that out-parameter getGeometry overloads
// deep-copy so the out-parameter outlives its parent container. Before the
// fix, these accessors installed an EmptyDestructor view over the parent's
// OGR memory, and destroying the parent turned the view into a dangle.

namespace {

f2c::types::Cell makeUnitCell() {
  f2c::types::LinearRing ring;
  ring.addPoint(0.0, 0.0);
  ring.addPoint(1.0, 0.0);
  ring.addPoint(1.0, 1.0);
  ring.addPoint(0.0, 1.0);
  ring.addPoint(0.0, 0.0);
  f2c::types::Cell cell;
  cell.addRing(ring);
  return cell;
}

}  // namespace

TEST(CellsLifetime, GetGeometryOutParamOutlivesParent) {
  f2c::types::Cell out;
  {
    f2c::types::Cells cells_local;
    cells_local.addGeometry(makeUnitCell());
    cells_local.getGeometry(0, out);
  }  // parent destroyed here
  // If T-002 is fixed, `out` is independently owning and these operations
  // do not touch freed memory.
  EXPECT_NO_THROW(out.area());
  EXPECT_NEAR(out.area(), 1.0, 1e-9);
}

TEST(CellLifetime, GetGeometryRingOutlivesParent) {
  f2c::types::LinearRing out;
  {
    f2c::types::Cell cell_local = makeUnitCell();
    cell_local.getGeometry(0, out);
  }  // parent destroyed here
  EXPECT_NO_THROW(out.size());
  EXPECT_GT(out.size(), 0u);
}

TEST(MultiLineStringLifetime, GetGeometryLineOutlivesParent) {
  f2c::types::LineString out;
  {
    f2c::types::MultiLineString lines_local;
    f2c::types::LineString l;
    l.addPoint(0.0, 0.0);
    l.addPoint(1.0, 1.0);
    lines_local.addGeometry(l);
    lines_local.getGeometry(0, out);
  }  // parent destroyed here
  EXPECT_NO_THROW(out.size());
  EXPECT_EQ(out.size(), 2u);
}

TEST(MultiPointLifetime, GetGeometryPointOutlivesParent) {
  f2c::types::Point out;
  {
    f2c::types::MultiPoint mp_local;
    mp_local.addPoint(f2c::types::Point(3.0, 4.0));
    mp_local.getGeometry(0, out);
  }  // parent destroyed here
  EXPECT_NO_THROW(out.getX());
  EXPECT_DOUBLE_EQ(out.getX(), 3.0);
  EXPECT_DOUBLE_EQ(out.getY(), 4.0);
}
