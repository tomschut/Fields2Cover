// test_parse_geojson.cpp — integration tests for ParseGeoJSON RPC.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;

TEST_F(F2cGrpcTestFixture, ParseGeoJSON_HappyPath_ReturnsField) {
  f2c::v1::ParseGeoJSONRequest req;
  req.set_geojson(F2cGrpcTestFixture::SampleFieldGeoJson());
  req.set_id("test-1");

  f2c::v1::ParseGeoJSONResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().ParseGeoJSON(&ctx, req, &resp);

  ASSERT_TRUE(status.ok()) << status.error_message();
  EXPECT_EQ(resp.field().id(), "test-1");
  EXPECT_FALSE(resp.field().geometry_wkt().empty());
}

TEST_F(F2cGrpcTestFixture, ParseGeoJSON_ErrorPath_EmptyGeoJson_InvalidArgument) {
  f2c::v1::ParseGeoJSONRequest req;
  req.set_geojson("");

  f2c::v1::ParseGeoJSONResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().ParseGeoJSON(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
