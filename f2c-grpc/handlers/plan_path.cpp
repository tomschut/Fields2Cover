// handlers/plan_path.cpp — PlanPath RPC implementation.
//
// Wraps f2c::pp::PathPlanning::planPath with a runtime dispatch on the
// proto's TurningAlgorithm enum. Terminal RPC: no further operations
// follow it on the coverage pipeline.

#include "service_impl.h"
#include "util/route_conversions.h"
#include "util/status.h"

#include "fields2cover/path_planning/dubins_curves.h"
#include "fields2cover/path_planning/dubins_curves_cc.h"
#include "fields2cover/path_planning/path_planning.h"
#include "fields2cover/path_planning/reeds_shepp_curves.h"
#include "fields2cover/path_planning/reeds_shepp_curves_hc.h"
#include "fields2cover/path_planning/turning_base.h"
#include "fields2cover/types/Path.h"
#include "fields2cover/types/Robot.h"
#include "fields2cover/types/Route.h"

#include <memory>
#include <stdexcept>

namespace f2c_grpc {

namespace {

// Minimal proto->f2c Robot converter. This duplicates what plan 10-03's
// util/conversions.h will export — kept local here so this plan builds
// without depending on parallel plans 10-03/10-04. At merge time replace
// the inline call with `util::robotFromProto`.
f2c::types::Robot robotFromProto(const f2c::v1::Robot& proto) {
  f2c::types::Robot out;
  out.setName(proto.name());
  out.setWidth(proto.width_m());
  out.setCovWidth(proto.cov_width_m());
  if (proto.min_turning_radius_m() > 0.0) {
    out.setMinTurningRadius(proto.min_turning_radius_m());
  }
  if (proto.max_curv() > 0.0) {
    out.setMaxCurv(proto.max_curv());
  }
  if (proto.max_diff_curv() > 0.0) {
    out.setMaxDiffCurv(proto.max_diff_curv());
  }
  if (proto.cruise_vel_mps() > 0.0) {
    out.setCruiseVel(proto.cruise_vel_mps());
  }
  if (proto.turn_vel_mps() > 0.0) {
    out.setTurnVel(proto.turn_vel_mps());
  }
  return out;
}

std::unique_ptr<f2c::pp::TurningBase> makeTurnPlanner(
    f2c::v1::TurningAlgorithm alg) {
  switch (alg) {
    case f2c::v1::TURNING_ALGORITHM_UNSPECIFIED:
    case f2c::v1::TURNING_ALGORITHM_DUBINS:
      return std::make_unique<f2c::pp::DubinsCurves>();
    case f2c::v1::TURNING_ALGORITHM_DUBINS_CC:
      return std::make_unique<f2c::pp::DubinsCurvesCC>();
    case f2c::v1::TURNING_ALGORITHM_REEDS_SHEPP:
      return std::make_unique<f2c::pp::ReedsSheppCurves>();
    case f2c::v1::TURNING_ALGORITHM_REEDS_SHEPP_HC:
      return std::make_unique<f2c::pp::ReedsSheppCurvesHC>();
    default:
      throw std::invalid_argument(
          "PlanPath: unknown turning_algorithm enum value");
  }
}

}  // namespace

grpc::Status F2CServiceImpl::PlanPath(
    grpc::ServerContext*, const f2c::v1::PlanPathRequest* req,
    f2c::v1::PlanPathResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("PlanPath: null request or response");
    }
    if (req->route().swaths().items_size() == 0) {
      throw std::invalid_argument("PlanPath: route.swaths.items is empty");
    }

    f2c::types::Robot robot = robotFromProto(req->robot());
    f2c::types::Route route = util::routeFromProto(req->route());

    std::unique_ptr<f2c::pp::TurningBase> turn_planner =
        makeTurnPlanner(req->turning_algorithm());

    // PathPlanning methods are static — no instance needed. Pass route by
    // value-const-ref as the header specifies.
    f2c::types::Path path =
        f2c::pp::PathPlanning::planPath(robot, route, *turn_planner);

    *resp->mutable_path() =
        util::pathToProto(path, req->route().field());
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
