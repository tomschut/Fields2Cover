// F2CServiceImpl — implementation of the f2c.v1.F2C gRPC service.
//
// Each RPC override lives in its own handler file (f2c-grpc/handlers/*.cpp)
// added by plans 10-03 / 10-04 / 10-05. This header declares the full
// interface up front so the build is stable as handlers land incrementally.

#pragma once

#include <generated/f2c.grpc.pb.h>
#include <grpcpp/grpcpp.h>

namespace f2c_grpc {

class F2CServiceImpl final : public f2c::v1::F2C::Service {
 public:
  F2CServiceImpl() = default;
  ~F2CServiceImpl() override = default;

  grpc::Status ParseGeoJSON(
      grpc::ServerContext* ctx,
      const f2c::v1::ParseGeoJSONRequest* req,
      f2c::v1::ParseGeoJSONResponse* resp) override;

  grpc::Status TransformToUTM(
      grpc::ServerContext* ctx,
      const f2c::v1::TransformToUTMRequest* req,
      f2c::v1::TransformToUTMResponse* resp) override;

  grpc::Status CloneField(
      grpc::ServerContext* ctx,
      const f2c::v1::CloneFieldRequest* req,
      f2c::v1::CloneFieldResponse* resp) override;

  grpc::Status CloneRobot(
      grpc::ServerContext* ctx,
      const f2c::v1::CloneRobotRequest* req,
      f2c::v1::CloneRobotResponse* resp) override;

  grpc::Status GenerateHeadland(
      grpc::ServerContext* ctx,
      const f2c::v1::GenerateHeadlandRequest* req,
      f2c::v1::GenerateHeadlandResponse* resp) override;

  grpc::Status GenerateSwaths(
      grpc::ServerContext* ctx,
      const f2c::v1::GenerateSwathsRequest* req,
      f2c::v1::GenerateSwathsResponse* resp) override;

  grpc::Status SortSwaths(
      grpc::ServerContext* ctx,
      const f2c::v1::SortSwathsRequest* req,
      f2c::v1::SortSwathsResponse* resp) override;

  grpc::Status GenerateRoute(
      grpc::ServerContext* ctx,
      const f2c::v1::GenerateRouteRequest* req,
      f2c::v1::GenerateRouteResponse* resp) override;

  grpc::Status PlanPath(
      grpc::ServerContext* ctx,
      const f2c::v1::PlanPathRequest* req,
      f2c::v1::PlanPathResponse* resp) override;
};

}  // namespace f2c_grpc
