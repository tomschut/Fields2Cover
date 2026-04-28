// test_generate_swaths.cpp — integration tests for GenerateSwaths RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;
using f2c_grpc::testing::MakeTestRobot;

TEST_F(F2cGrpcTestFixture, GenerateSwaths_HappyPath_ReturnsSwaths) {
  auto fwh = BuildHeadlands();

  f2c::v1::GenerateSwathsRequest req;
  *req.mutable_field_with_headlands() = fwh;
  *req.mutable_robot() = MakeTestRobot();
  req.set_angle_rad(0.0);
  req.set_width_m(2.0);

  f2c::v1::GenerateSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateSwaths(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_GT(resp.swaths().items_size(), 0);
  EXPECT_FALSE(resp.swaths().sorted());
}

TEST_F(F2cGrpcTestFixture, GenerateSwaths_ErrorPath_ZeroWidth_InvalidArgument) {
  auto fwh = BuildHeadlands();

  f2c::v1::GenerateSwathsRequest req;
  *req.mutable_field_with_headlands() = fwh;
  *req.mutable_robot() = MakeTestRobot();
  req.set_angle_rad(0.0);
  req.set_width_m(0.0);  // invalid

  f2c::v1::GenerateSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateSwaths(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
