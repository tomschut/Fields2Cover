// test_clone_field.cpp — integration tests for CloneField RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;

TEST_F(F2cGrpcTestFixture, CloneField_HappyPath_RoundTrips) {
  auto field = ParseSampleField();

  f2c::v1::CloneFieldRequest req;
  *req.mutable_field() = field;

  f2c::v1::CloneFieldResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().CloneField(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_EQ(resp.field().id(), field.id());
  EXPECT_FALSE(resp.field().geometry_wkt().empty());
}

TEST_F(F2cGrpcTestFixture, CloneField_ErrorPath_MissingField_InvalidArgument) {
  f2c::v1::CloneFieldRequest req;
  // no field set

  f2c::v1::CloneFieldResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().CloneField(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
