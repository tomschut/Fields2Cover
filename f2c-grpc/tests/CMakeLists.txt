# Unit tests for the f2c-grpc shim utility layer.
#
# Matches the parent f2c-grpc/CMakeLists.txt dependency pattern: system
# protobuf/grpc headers put FIRST via the SYS_PROTO_GRPC_INCLUDE variable
# inherited from the parent scope, libs resolved via pkg-config variables
# SYS_GRPCPP_LIBRARIES / SYS_PROTOBUF_LIBRARIES.

find_package(GTest REQUIRED)
include(GoogleTest)

add_executable(f2c-grpc-utils-tests
  util/test_wkt.cpp
  util/test_status.cpp
  $<TARGET_OBJECTS:f2c_grpc_util>
)

target_include_directories(f2c-grpc-utils-tests BEFORE PRIVATE
  "${SYS_PROTO_GRPC_INCLUDE}"
  "${CMAKE_CURRENT_SOURCE_DIR}/.."
  "${CMAKE_BINARY_DIR}/f2c-grpc"
)

target_link_libraries(f2c-grpc-utils-tests PRIVATE
  f2c_grpc_proto
  Fields2Cover
  ${SYS_GRPCPP_LIBRARIES}
  ${SYS_PROTOBUF_LIBRARIES}
  GTest::gtest
  GTest::gtest_main
)
target_link_directories(f2c-grpc-utils-tests PRIVATE
  ${SYS_GRPCPP_LIBRARY_DIRS}
  ${SYS_PROTOBUF_LIBRARY_DIRS}
)

target_compile_features(f2c-grpc-utils-tests PRIVATE cxx_std_17)

gtest_discover_tests(f2c-grpc-utils-tests
  PROPERTIES LABELS "f2c_grpc_utils"
)

# ---- Integration tests (end-to-end through a live gRPC server) -----------

add_executable(f2c-grpc-integration-tests
  test_parse_geojson.cpp
  test_transform_to_utm.cpp
  test_clone_field.cpp
  test_clone_robot.cpp
  test_generate_headland.cpp
  test_generate_swaths.cpp
  test_sort_swaths.cpp
  test_generate_route.cpp
  test_plan_path.cpp
  ../service_impl.cpp
  ../handlers/parse_geojson.cpp
  ../handlers/transform_to_utm.cpp
  ../handlers/clone_field.cpp
  ../handlers/clone_robot.cpp
  ../handlers/generate_headland.cpp
  ../handlers/generate_swaths.cpp
  ../handlers/sort_swaths.cpp
  ../handlers/generate_route.cpp
  ../handlers/plan_path.cpp
  $<TARGET_OBJECTS:f2c_grpc_util>
)

target_include_directories(f2c-grpc-integration-tests BEFORE PRIVATE
  "${SYS_PROTO_GRPC_INCLUDE}"
  "${CMAKE_CURRENT_SOURCE_DIR}"
  "${CMAKE_CURRENT_SOURCE_DIR}/.."
  "${CMAKE_BINARY_DIR}/f2c-grpc"
)

target_compile_definitions(f2c-grpc-integration-tests PRIVATE
  F2C_GRPC_TEST_DATA_DIR="${CMAKE_CURRENT_SOURCE_DIR}/testdata"
)

target_link_libraries(f2c-grpc-integration-tests PRIVATE
  f2c_grpc_proto
  Fields2Cover
  ${SYS_GRPCPP_LIBRARIES}
  ${SYS_PROTOBUF_LIBRARIES}
  GTest::gtest
  GTest::gtest_main
)
target_link_directories(f2c-grpc-integration-tests PRIVATE
  ${SYS_GRPCPP_LIBRARY_DIRS}
  ${SYS_PROTOBUF_LIBRARY_DIRS}
)

target_compile_features(f2c-grpc-integration-tests PRIVATE cxx_std_17)

gtest_discover_tests(f2c-grpc-integration-tests
  PROPERTIES LABELS "f2c_grpc_integration"
)
