// test_sort_swaths.cpp — integration tests for SortSwaths RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;

TEST_F(F2cGrpcTestFixture, SortSwaths_HappyPath_Boustrophedon) {
  auto swaths = BuildSwaths();
  const int n_in = swaths.items_size();

  f2c::v1::SortSwathsRequest req;
  *req.mutable_swaths() = swaths;
  req.set_algorithm(f2c::v1::SORT_ALGORITHM_BOUSTROPHEDON);
  req.set_variant(0);

  f2c::v1::SortSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().SortSwaths(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_EQ(resp.swaths().items_size(), n_in);
  EXPECT_TRUE(resp.swaths().sorted());
}

TEST_F(F2cGrpcTestFixture, SortSwaths_WithStartPoint_ReturnsSortedSwaths) {
  auto swaths = BuildSwaths();
  const int n_in = swaths.items_size();

  f2c::v1::SortSwathsRequest req;
  *req.mutable_swaths() = swaths;
  req.set_algorithm(f2c::v1::SORT_ALGORITHM_BOUSTROPHEDON);
  req.set_variant(0);
  // Set a start_point near the field centre (arbitrary local-metric coords).
  req.mutable_start_point()->set_x(0.0);
  req.mutable_start_point()->set_y(0.0);

  f2c::v1::SortSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().SortSwaths(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_EQ(resp.swaths().items_size(), n_in);
  EXPECT_TRUE(resp.swaths().sorted());
}

TEST_F(F2cGrpcTestFixture, SortSwaths_Snake_WithStartPoint_ReturnsSortedSwaths) {
  auto swaths = BuildSwaths();

  f2c::v1::SortSwathsRequest req;
  *req.mutable_swaths() = swaths;
  req.set_algorithm(f2c::v1::SORT_ALGORITHM_SNAKE);
  req.set_variant(0);
  req.mutable_start_point()->set_x(100.0);
  req.mutable_start_point()->set_y(50.0);

  f2c::v1::SortSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().SortSwaths(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_TRUE(resp.swaths().sorted());
}

TEST_F(F2cGrpcTestFixture, SortSwaths_ErrorPath_InvalidVariant_InvalidArgument) {
  auto swaths = BuildSwaths();

  f2c::v1::SortSwathsRequest req;
  *req.mutable_swaths() = swaths;
  req.set_algorithm(f2c::v1::SORT_ALGORITHM_BOUSTROPHEDON);
  req.set_variant(99);  // out of [0,3]

  f2c::v1::SortSwathsResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().SortSwaths(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
