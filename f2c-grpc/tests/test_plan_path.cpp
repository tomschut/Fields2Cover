// test_plan_path.cpp — integration tests for PlanPath RPC (terminal step).

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;
using f2c_grpc::testing::MakeTestRobot;

TEST_F(F2cGrpcTestFixture, PlanPath_HappyPath_Dubins_ReturnsPath) {
  auto route = BuildRoute();

  f2c::v1::PlanPathRequest req;
  *req.mutable_route() = route;
  *req.mutable_robot() = MakeTestRobot();
  req.set_turning_algorithm(f2c::v1::TURNING_ALGORITHM_DUBINS);

  f2c::v1::PlanPathResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().PlanPath(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_GT(resp.path().states_size(), 0);
}

TEST_F(F2cGrpcTestFixture, PlanPath_ErrorPath_EmptyRoute_InvalidArgument) {
  // Empty route: no swaths at all.
  f2c::v1::PlanPathRequest req;
  // route left default-constructed (no swaths)
  *req.mutable_robot() = MakeTestRobot();
  req.set_turning_algorithm(f2c::v1::TURNING_ALGORITHM_DUBINS);

  f2c::v1::PlanPathResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().PlanPath(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
