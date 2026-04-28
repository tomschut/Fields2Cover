// test_generate_route.cpp — integration tests for GenerateRoute RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;

TEST_F(F2cGrpcTestFixture, GenerateRoute_HappyPath_ReturnsRoute) {
  auto sorted = BuildSortedSwaths();

  f2c::v1::GenerateRouteRequest req;
  *req.mutable_swaths() = sorted;

  f2c::v1::GenerateRouteResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateRoute(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  // The route may have 0 connections for a single-cell field; the
  // presence of a route + its embedded swaths is enough to prove the
  // RPC plumbing works.
  EXPECT_GT(resp.route().swaths().items_size(), 0);
  EXPECT_GE(resp.route().connections_size(), 0);
}

TEST_F(F2cGrpcTestFixture, GenerateRoute_ErrorPath_UnsortedSwaths_InvalidArgument) {
  // BuildSwaths returns unsorted swaths — the handler should reject.
  auto swaths = BuildSwaths();

  f2c::v1::GenerateRouteRequest req;
  *req.mutable_swaths() = swaths;

  f2c::v1::GenerateRouteResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().GenerateRoute(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
