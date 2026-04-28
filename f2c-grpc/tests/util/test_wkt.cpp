// Unit tests for util/wkt.{h,cpp} — WKT <-> OGRGeometry round-trip helpers.

#include "util/wkt.h"

#include <gdal/ogr_geometry.h>
#include <gtest/gtest.h>

#include <stdexcept>
#include <string>

using f2c_grpc::util::ogrToWkt;
using f2c_grpc::util::wktToOgr;

namespace {

void ExpectRoundTrip(const std::string& wkt) {
  auto geom = wktToOgr(wkt);
  ASSERT_NE(geom.get(), nullptr);
  const std::string reencoded = ogrToWkt(*geom);
  // Reparse and compare geometrically instead of textually: GDAL may
  // normalize whitespace / numeric precision between encodings.
  auto roundtrip = wktToOgr(reencoded);
  EXPECT_TRUE(geom->Equals(roundtrip.get()))
      << "original: " << wkt << "\nreencoded: " << reencoded;
}

}  // namespace

TEST(WktRoundTrip, Polygon) {
  ExpectRoundTrip("POLYGON ((0 0,10 0,10 10,0 10,0 0))");
}

TEST(WktRoundTrip, LineString) {
  ExpectRoundTrip("LINESTRING (0 0,5 5,10 0)");
}

TEST(WktRoundTrip, MultiPoint) {
  ExpectRoundTrip("MULTIPOINT ((0 0),(1 1),(2 2))");
}

TEST(WktParse, EmptyStringThrowsInvalidArgument) {
  EXPECT_THROW(wktToOgr(""), std::invalid_argument);
}

TEST(WktParse, GarbageThrowsInvalidArgument) {
  EXPECT_THROW(wktToOgr("NOT_A_WKT_GEOMETRY"), std::invalid_argument);
}
