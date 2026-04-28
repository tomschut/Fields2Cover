// test_fixture.h — RAII GoogleTest fixture that starts an F2CServiceImpl
// on a unique temp Unix socket and hands out a ready-to-use client stub.
//
// One fixture instance per TEST_F — SetUp creates a fresh socket path using
// getpid() + test-name, StartServer binds, a channel + stub are created,
// TearDown shuts the server down and unlinks the socket.

#pragma once

#include <generated/f2c.grpc.pb.h>
#include <generated/f2c.pb.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>
#include <unistd.h>

#include <cstdio>
#include <fstream>
#include <memory>
#include <sstream>
#include <string>

#include "server_lifecycle.h"

namespace f2c_grpc::testing {

inline std::string MakeTempSocketPath(const ::testing::TestInfo* info) {
  std::string name = info != nullptr
      ? std::string(info->test_suite_name()) + "_" + info->name()
      : std::string("f2c_grpc_test");
  for (char& c : name) {
    if (c == '/' || c == '.') c = '_';
  }
  std::string path = "/tmp/f2c-test-" + std::to_string(::getpid()) + "-" +
                     name + ".sock";
  ::unlink(path.c_str());
  return path;
}

inline std::string ReadFile(const std::string& path) {
  std::ifstream f(path);
  std::stringstream ss;
  ss << f.rdbuf();
  return ss.str();
}

// Populate a Robot with reasonable values for happy-path tests.
inline f2c::v1::Robot MakeTestRobot() {
  f2c::v1::Robot r;
  r.set_name("test-robot");
  r.set_width_m(2.0);
  r.set_cov_width_m(1.8);
  r.set_min_turning_radius_m(1.0);
  r.set_max_curv(1.0);
  r.set_max_diff_curv(0.3);
  r.set_cruise_vel_mps(1.0);
  return r;
}

class F2cGrpcTestFixture : public ::testing::Test {
 public:
  void SetUp() override {
    const auto* info = ::testing::UnitTest::GetInstance()->current_test_info();
    socket_path_ = MakeTempSocketPath(info);
    const std::string uri = f2c_grpc::SocketPathToUri(socket_path_);
    handle_ = f2c_grpc::StartServer(uri);
    auto channel = grpc::CreateChannel(uri, grpc::InsecureChannelCredentials());
    stub_ = f2c::v1::F2C::NewStub(channel);
  }

  void TearDown() override {
    if (handle_.server != nullptr) {
      handle_.server->Shutdown();
    }
    stub_.reset();
    ::unlink(socket_path_.c_str());
  }

  f2c::v1::F2C::Stub& stub() { return *stub_; }

  // Convenience: read the canonical sample GeoJSON shipped in testdata/.
  static std::string SampleFieldGeoJson() {
    return ReadFile(std::string(F2C_GRPC_TEST_DATA_DIR) +
                    "/sample_field.geojson");
  }

  // Happy-path helper: parse the sample field. Tests that need a base
  // f2c::v1::Field to feed downstream RPCs call this in their setup.
  f2c::v1::Field ParseSampleField() {
    f2c::v1::ParseGeoJSONRequest req;
    req.set_geojson(SampleFieldGeoJson());
    req.set_id("sample");
    f2c::v1::ParseGeoJSONResponse resp;
    grpc::ClientContext ctx;
    auto s = stub().ParseGeoJSON(&ctx, req, &resp);
    EXPECT_TRUE(s.ok()) << "ParseSampleField failed: " << s.error_message();
    return resp.field();
  }

  // Happy-path helper: parse + headland.
  f2c::v1::FieldWithHeadlands BuildHeadlands(double width_m = 2.0,
                                             int count = 3) {
    auto field = ParseSampleField();
    f2c::v1::GenerateHeadlandRequest req;
    *req.mutable_field() = field;
    *req.mutable_robot() = MakeTestRobot();
    req.set_width_m(width_m);
    req.set_count(count);
    f2c::v1::GenerateHeadlandResponse resp;
    grpc::ClientContext ctx;
    auto s = stub().GenerateHeadland(&ctx, req, &resp);
    EXPECT_TRUE(s.ok()) << "BuildHeadlands failed: " << s.error_message();
    return resp.field_with_headlands();
  }

  // Happy-path helper: parse + headland + swaths.
  f2c::v1::Swaths BuildSwaths() {
    auto fwh = BuildHeadlands();
    f2c::v1::GenerateSwathsRequest req;
    *req.mutable_field_with_headlands() = fwh;
    *req.mutable_robot() = MakeTestRobot();
    req.set_angle_rad(0.0);
    req.set_width_m(2.0);
    f2c::v1::GenerateSwathsResponse resp;
    grpc::ClientContext ctx;
    auto s = stub().GenerateSwaths(&ctx, req, &resp);
    EXPECT_TRUE(s.ok()) << "BuildSwaths failed: " << s.error_message();
    return resp.swaths();
  }

  // Happy-path helper: parse + headland + swaths + sort.
  f2c::v1::Swaths BuildSortedSwaths() {
    auto swaths = BuildSwaths();
    f2c::v1::SortSwathsRequest req;
    *req.mutable_swaths() = swaths;
    req.set_algorithm(f2c::v1::SORT_ALGORITHM_BOUSTROPHEDON);
    req.set_variant(0);
    f2c::v1::SortSwathsResponse resp;
    grpc::ClientContext ctx;
    auto s = stub().SortSwaths(&ctx, req, &resp);
    EXPECT_TRUE(s.ok()) << "BuildSortedSwaths failed: " << s.error_message();
    return resp.swaths();
  }

  // Full pipeline up through GenerateRoute.
  f2c::v1::Route BuildRoute() {
    auto sorted = BuildSortedSwaths();
    f2c::v1::GenerateRouteRequest req;
    *req.mutable_swaths() = sorted;
    f2c::v1::GenerateRouteResponse resp;
    grpc::ClientContext ctx;
    auto s = stub().GenerateRoute(&ctx, req, &resp);
    EXPECT_TRUE(s.ok()) << "BuildRoute failed: " << s.error_message();
    return resp.route();
  }

 private:
  std::string socket_path_;
  f2c_grpc::ServerHandle handle_;
  std::unique_ptr<f2c::v1::F2C::Stub> stub_;
};

}  // namespace f2c_grpc::testing
