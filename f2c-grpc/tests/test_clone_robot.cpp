// test_clone_robot.cpp — integration tests for CloneRobot RPC.
//
// CloneRobot has no natural INVALID_ARGUMENT path — it's a pure value
// copy. The plan explicitly calls this out and allows either "two happy
// tests on different robot configs" OR "one happy + one gRPC-level bad
// call". We pick the former: one rich-config happy path, one
// missing-robot error path (the handler guards has_robot()).

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;
using f2c_grpc::testing::MakeTestRobot;

TEST_F(F2cGrpcTestFixture, CloneRobot_HappyPath_RoundTrips) {
  f2c::v1::CloneRobotRequest req;
  f2c::v1::Robot r = MakeTestRobot();
  r.set_name("r1");
  r.set_width_m(2.5);
  r.set_cov_width_m(2.0);
  r.set_min_turning_radius_m(1.5);
  *req.mutable_robot() = r;

  f2c::v1::CloneRobotResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().CloneRobot(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_EQ(resp.robot().name(), "r1");
  EXPECT_DOUBLE_EQ(resp.robot().width_m(), 2.5);
  EXPECT_DOUBLE_EQ(resp.robot().cov_width_m(), 2.0);
  // min_turning_radius_m round-trips through f2c's radius <-> curvature
  // conversion, which introduces ~1e-7 FP drift; use NEAR instead of
  // DOUBLE_EQ.
  EXPECT_NEAR(resp.robot().min_turning_radius_m(), 1.5, 1e-6);
}

TEST_F(F2cGrpcTestFixture, CloneRobot_ErrorPath_MissingRobot_InvalidArgument) {
  f2c::v1::CloneRobotRequest req;
  // no robot set

  f2c::v1::CloneRobotResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().CloneRobot(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
