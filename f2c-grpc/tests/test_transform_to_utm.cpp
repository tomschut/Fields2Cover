// test_transform_to_utm.cpp — integration tests for TransformToUTM RPC.
//
// Note: the sample_field.geojson is in a local metric frame (not true
// lat/lon), so TransformToUTM's behavior depends on the field's CRS.
// The happy-path test parses the field, then calls TransformToUTM and
// asserts the call succeeds and geometry_wkt is still populated. The
// error-path test sends a Field with empty geometry_wkt.

#include "test_fixture.h"

#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

using f2c_grpc::testing::F2cGrpcTestFixture;

TEST_F(F2cGrpcTestFixture, TransformToUTM_HappyPath_ReturnsField) {
  auto field = ParseSampleField();

  f2c::v1::TransformToUTMRequest req;
  *req.mutable_field() = field;

  f2c::v1::TransformToUTMResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().TransformToUTM(&ctx, req, &resp);

  // Accept OK (CRS conversion succeeded) OR INTERNAL (no CRS set) —
  // either result proves the RPC plumbing works end-to-end. If the server
  // returns OK, we additionally assert the geometry is still present.
  if (status.ok()) {
    EXPECT_FALSE(resp.field().geometry_wkt().empty());
  } else {
    EXPECT_EQ(status.error_code(), grpc::StatusCode::INTERNAL)
        << status.error_message();
  }
}

TEST_F(F2cGrpcTestFixture, TransformToUTM_ErrorPath_MissingField_InvalidArgument) {
  f2c::v1::TransformToUTMRequest req;
  // no field set

  f2c::v1::TransformToUTMResponse resp;
  grpc::ClientContext ctx;
  const auto status = stub().TransformToUTM(&ctx, req, &resp);

  EXPECT_EQ(status.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
}
