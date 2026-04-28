// handlers/clone_robot.cpp — CloneRobot RPC.
//
// Deep-copy a proto Robot scalar-by-scalar via f2c::types::Robot.

#include "service_impl.h"

#include "util/conversions.h"
#include "util/status.h"

#include <stdexcept>

namespace f2c_grpc {

grpc::Status F2CServiceImpl::CloneRobot(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::CloneRobotRequest* req,
    f2c::v1::CloneRobotResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("null request or response");
    }
    if (!req->has_robot()) {
      throw std::invalid_argument("CloneRobot: robot is required");
    }
    f2c::types::Robot copy = util::robotFromProto(req->robot());
    *resp->mutable_robot() = util::robotToProto(copy);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
