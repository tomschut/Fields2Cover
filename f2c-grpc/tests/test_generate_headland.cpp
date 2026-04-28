// test_generate_headland.cpp — integration tests for GenerateHeadland RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;
using f2c_grpc::testing::MakeTestRobot;

TEST_F(F2cGrpcTestFixture, GenerateHeadland_HappyPath_ReturnsRings) {
  auto field = ParseSampleField();

  f2c::v1::GenerateHeadlandRequest req;
  *req.mutable_field() = field;
  *req.mutable_robot() = MakeTestRobot();
  req.set_width_m(2.0);
  req.set_count(3);

  f2c::v1::GenerateHeadlandResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateHeadland(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_FALSE(resp.field_with_headlands().inner_field_wkt().empty());
  EXPECT_FALSE(resp.field_with_headlands().headlands_wkt().empty());
  EXPECT_EQ(resp.field_with_headlands().headland_count(), 3);
}

TEST_F(F2cGrpcTestFixture, GenerateHeadland_ErrorPath_NegativeWidth_InvalidArgument) {
  auto field = ParseSampleField();

  f2c::v1::GenerateHeadlandRequest req;
  *req.mutable_field() = field;
  *req.mutable_robot() = MakeTestRobot();
  req.set_width_m(-1.0);
  req.set_count(3);

  f2c::v1::GenerateHeadlandResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateHeadland(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
