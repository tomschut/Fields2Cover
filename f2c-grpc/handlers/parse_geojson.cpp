// handlers/parse_geojson.cpp — ParseGeoJSON RPC.
//
// REFERENCE PATTERN for every handler in plans 10-03/04/05:
//
//   try {
//     // 1. Validate inputs (throw std::invalid_argument on missing fields)
//     // 2. Convert proto request -> f2c native types (util::fromProto)
//     // 3. Call the f2c operation
//     // 4. Convert result -> proto response (util::toProto)
//     // 5. return grpc::Status::OK
//   } catch (const std::exception& e) {
//     return util::exceptionToStatus(e);
//   }
//
// Request:  f2c.v1.ParseGeoJSONRequest  { string geojson, CRS target_crs,
//                                          string id }
// Response: f2c.v1.ParseGeoJSONResponse { Field field }

#include "service_impl.h"

#include "util/conversions.h"
#include "util/status.h"

#include "fields2cover/types.h"
#include "fields2cover/utils/parser.h"
#include "fields2cover/utils/transformation.h"

#include <stdexcept>
#include <string>

namespace f2c_grpc {

grpc::Status F2CServiceImpl::ParseGeoJSON(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::ParseGeoJSONRequest* req,
    f2c::v1::ParseGeoJSONResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("null request or response");
    }
    if (req->geojson().empty()) {
      throw std::invalid_argument("ParseGeoJSON: geojson is required");
    }

    F2CFields fields;
    const int rc = f2c::Parser::importJsonFromString(req->geojson(), fields);
    if (rc != 0) {
      throw std::invalid_argument(
          "ParseGeoJSON: failed to parse GeoJSON (Parser rc=" +
          std::to_string(rc) + ")");
    }
    if (fields.empty()) {
      throw std::invalid_argument(
          "ParseGeoJSON: FeatureCollection contained zero fields");
    }

    f2c::types::Field field = fields[0];
    if (!req->id().empty()) {
      field.setId(req->id());
    }
    if (req->target_crs().epsg() > 0) {
      // Explicit target CRS: tag the field and transform if needed.
      field.setEPSGCoordSystem(req->target_crs().epsg());
    } else {
      // No explicit target: auto-pick a local UTM zone from the input
      // centroid. transformToUTM accepts fields with CRS in {empty,
      // EPSG:4326, EPSG:4258} and no-ops geographic→UTM when the parser
      // already assigned a UTM CRS. This is what real-world clients
      // (wodan, farmmaps) rely on — they send WGS84 and expect the
      // server to project for the coverage math.
      f2c::Transform::transformToUTM(field);
    }

    *resp->mutable_field() = util::fieldToProto(field);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
