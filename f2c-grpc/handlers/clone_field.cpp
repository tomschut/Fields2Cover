// handlers/clone_field.cpp — CloneField RPC.
//
// Deep-copy a proto Field by round-tripping through f2c::types::Field.
// The fieldFromProto / fieldToProto pair already produces an isolated
// copy via the WKT-backed Cells round-trip — no shared OGR handles.

#include "service_impl.h"

#include "util/conversions.h"
#include "util/status.h"

#include <stdexcept>

namespace f2c_grpc {

grpc::Status F2CServiceImpl::CloneField(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::CloneFieldRequest* req,
    f2c::v1::CloneFieldResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("null request or response");
    }
    if (!req->has_field()) {
      throw std::invalid_argument("CloneField: field is required");
    }
    f2c::types::Field copy = util::fieldFromProto(req->field());
    *resp->mutable_field() = util::fieldToProto(copy);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
