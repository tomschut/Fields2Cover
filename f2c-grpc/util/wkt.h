// util/wkt.h — WKT <-> OGRGeometry round-trip helpers.
//
// All geometry on the proto wire is WKT stored in `bytes` fields. Handlers
// use wktToOgr() to parse an inbound request and ogrToWkt() to serialize
// an f2c-produced geometry for the response.

#pragma once

#include <memory>
#include <string>

class OGRGeometry;

namespace f2c_grpc::util {

struct OGRGeometryDeleter {
  void operator()(OGRGeometry* g) const noexcept;
};

using OGRGeometryPtr = std::unique_ptr<OGRGeometry, OGRGeometryDeleter>;

// Parse a WKT string into an owning OGRGeometry.
// Throws std::invalid_argument if the WKT is empty or malformed.
OGRGeometryPtr wktToOgr(const std::string& wkt);

// Serialize an OGRGeometry to a WKT string suitable for assignment to a
// proto `bytes` field.
std::string ogrToWkt(const OGRGeometry& geom);

}  // namespace f2c_grpc::util
