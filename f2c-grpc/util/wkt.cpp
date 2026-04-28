// util/wkt.cpp — implementation of WKT <-> OGRGeometry helpers.

#include "util/wkt.h"

#include <gdal/cpl_conv.h>
#include <gdal/ogr_geometry.h>

#include <stdexcept>
#include <string>

namespace f2c_grpc::util {

void OGRGeometryDeleter::operator()(OGRGeometry* g) const noexcept {
  if (g != nullptr) {
    OGRGeometryFactory::destroyGeometry(g);
  }
}

OGRGeometryPtr wktToOgr(const std::string& wkt) {
  if (wkt.empty()) {
    throw std::invalid_argument("wktToOgr: empty WKT string");
  }
  OGRGeometry* raw = nullptr;
  // Note: OGRGeometryFactory::createFromWkt takes a `char**` cursor on GDAL
  // 2.x (as used by the Fields2Cover build). Passing a pointer to the
  // string's internal buffer is safe because createFromWkt only advances
  // the cursor — it does not mutate the characters.
  const char* cursor = wkt.c_str();
  const OGRErr err =
      OGRGeometryFactory::createFromWkt(&cursor, nullptr, &raw);
  if (err != OGRERR_NONE || raw == nullptr) {
    if (raw != nullptr) {
      OGRGeometryFactory::destroyGeometry(raw);
    }
    throw std::invalid_argument("wktToOgr: malformed WKT: " + wkt);
  }
  return OGRGeometryPtr(raw);
}

std::string ogrToWkt(const OGRGeometry& geom) {
  char* out = nullptr;
  const OGRErr err = const_cast<OGRGeometry&>(geom).exportToWkt(&out);
  if (err != OGRERR_NONE || out == nullptr) {
    if (out != nullptr) {
      CPLFree(out);
    }
    throw std::runtime_error("ogrToWkt: exportToWkt failed");
  }
  std::string result(out);
  CPLFree(out);
  return result;
}

}  // namespace f2c_grpc::util
