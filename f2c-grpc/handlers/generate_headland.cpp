// handlers/generate_headland.cpp — F2CServiceImpl::GenerateHeadland.
//
// Ports `openapi_server.controllers.headland_controller.generate_headlands`
// to C++. The handler calls BOTH
//   * `f2c::hg::ConstHL::generateHeadlands`       — inner cultivable area
//   * `f2c::hg::ConstHL::generateHeadlandSwaths`  — headland swath rings
// and folds both outputs into the FieldWithHeadlands response so downstream
// RPCs (GenerateSwaths, GenerateRoute in plan 10-05) have everything they
// need without having to re-run the headland generator.

#include "service_impl.h"

#include <stdexcept>
#include <string>
#include <vector>

#include "util/status.h"
#include "util/swath_conversions.h"

#include "fields2cover/headland_generator/constant_headland.h"
#include "fields2cover/types.h"

namespace f2c_grpc {

namespace {

// Merge every Cells ring in `rings` into a single F2CCells so the
// composite can be serialized as one MULTIPOLYGON WKT. Uses addGeometry
// to flatten cell-by-cell.
f2c::types::Cells mergeHeadlandRings(
    const std::vector<f2c::types::Cells>& rings) {
  f2c::types::Cells merged;
  for (const auto& ring : rings) {
    for (size_t i = 0; i < ring.size(); ++i) {
      merged.addGeometry(ring.getGeometry(i));
    }
  }
  return merged;
}

}  // namespace

grpc::Status F2CServiceImpl::GenerateHeadland(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::GenerateHeadlandRequest* req,
    f2c::v1::GenerateHeadlandResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument(
          "GenerateHeadland: null request or response");
    }
    if (req->width_m() <= 0.0) {
      throw std::invalid_argument(
          "GenerateHeadland: width_m must be > 0");
    }
    // Default to 3 headland swaths — matches the Python pipeline/controller
    // default in fields2cover-api/route_controller.generate_route.
    const int count = req->count() > 0 ? req->count() : 3;

    f2c::types::Field field = util::fieldFromProtoLocal(req->field());
    if (field.getField().size() == 0) {
      throw std::invalid_argument(
          "GenerateHeadland: field geometry is empty");
    }

    f2c::hg::ConstHL hl_gen;
    F2CCells inner =
        hl_gen.generateHeadlands(field.getField(), req->width_m());
    std::vector<F2CCells> rings_per_swath =
        hl_gen.generateHeadlandSwaths(
            field.getField(), req->width_m(), count, /*dir_out2in=*/false);

    F2CCells headland_rings = mergeHeadlandRings(rings_per_swath);

    *resp->mutable_field_with_headlands() = util::fieldWithHeadlandsToProto(
        field, inner, headland_rings, req->width_m(), count);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
