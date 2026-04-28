// handlers/transform_to_utm.cpp — TransformToUTM RPC.
//
// Project a Field from its current CRS to a local UTM zone.
//
// Request:  f2c.v1.TransformToUTMRequest  { Field field }
// Response: f2c.v1.TransformToUTMResponse { Field field }
//
// f2c API used: f2c::Transform::transformToUTM(F2CField&, bool) — mutates
// the field in place to its local UTM projection. The bool selects the
// ETRS89 datum optimization for European fields (default true matches the
// f2c default).

#include "service_impl.h"

#include "util/conversions.h"
#include "util/status.h"

#include "fields2cover/types.h"
#include "fields2cover/utils/transformation.h"

#include <stdexcept>

namespace f2c_grpc {

grpc::Status F2CServiceImpl::TransformToUTM(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::TransformToUTMRequest* req,
    f2c::v1::TransformToUTMResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("null request or response");
    }
    if (!req->has_field()) {
      throw std::invalid_argument("TransformToUTM: field is required");
    }

    f2c::types::Field field = util::fieldFromProto(req->field());
    f2c::Transform::transformToUTM(field);

    *resp->mutable_field() = util::fieldToProto(field);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
