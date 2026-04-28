# f2c-grpc — standalone C++ gRPC server wrapping fields2cover operations.
# Links against the Fields2Cover shared library (built by the top-level
# CMakeLists) and the generated C++ stubs from proto/f2c.proto.

cmake_minimum_required(VERSION 3.16)

# ---- Dependencies ---------------------------------------------------------
#
# IMPORTANT: the top-level build pulls in or-tools via FetchContent, which
# ships its own vendored protobuf v25.x under build/_deps/ortools-src/include/
# and registers global `protobuf::libprotobuf` / `protobuf::protoc` imported
# targets pointing at that vendored copy. That protobuf version is
# incompatible with the *system* gRPC (1.51) + libprotobuf-dev (3.21) that the
# shim needs to link against — so we deliberately avoid the imported targets
# and resolve the system toolchain + libraries by explicit path.
#
# This does mean the shim is currently Linux/Debian-only. Dockerization in a
# later plan will bake in a known-good base image so the toolchain is
# deterministic.

find_program(_protoc_exe NAMES protoc
  PATHS /usr/bin /usr/local/bin
  NO_DEFAULT_PATH
  REQUIRED)
find_program(_grpc_cpp_plugin_exe grpc_cpp_plugin
  PATHS /usr/bin /usr/local/bin
  NO_DEFAULT_PATH
  REQUIRED)

# System libprotobuf + libgrpc++ via pkg-config. pkg-config gives the
# complete transitive link line (grpc, gpr, absl_*, cares, re2, ssl, etc.)
# that the bare imported targets from CMake config files don't expose under
# a plain `find_library`. We also deliberately avoid `protobuf::libprotobuf`
# and `gRPC::grpc++` imported targets because they drag in the
# or-tools-vendored protobuf v25 include dir, which conflicts with system
# libprotobuf-dev v3.21.
find_package(PkgConfig REQUIRED)
pkg_check_modules(SYS_GRPCPP REQUIRED grpc++)
pkg_check_modules(SYS_PROTOBUF REQUIRED protobuf)
set(SYS_PROTO_GRPC_INCLUDE /usr/include)

# ---- Codegen --------------------------------------------------------------

set(F2C_PROTO "${CMAKE_SOURCE_DIR}/proto/f2c.proto")
set(F2C_GEN_DIR "${CMAKE_CURRENT_BINARY_DIR}/generated")
file(MAKE_DIRECTORY "${F2C_GEN_DIR}")

set(F2C_PB_H   "${F2C_GEN_DIR}/f2c.pb.h")
set(F2C_PB_CC  "${F2C_GEN_DIR}/f2c.pb.cc")
set(F2C_GPB_H  "${F2C_GEN_DIR}/f2c.grpc.pb.h")
set(F2C_GPB_CC "${F2C_GEN_DIR}/f2c.grpc.pb.cc")

add_custom_command(
  OUTPUT  "${F2C_PB_H}" "${F2C_PB_CC}" "${F2C_GPB_H}" "${F2C_GPB_CC}"
  COMMAND ${_protoc_exe}
  ARGS    --proto_path=${CMAKE_SOURCE_DIR}/proto
          --cpp_out=${F2C_GEN_DIR}
          --grpc_out=${F2C_GEN_DIR}
          --plugin=protoc-gen-grpc=${_grpc_cpp_plugin_exe}
          ${F2C_PROTO}
  DEPENDS "${F2C_PROTO}"
  COMMENT "Generating C++ stubs from proto/f2c.proto"
  VERBATIM
)

add_library(f2c_grpc_proto STATIC
  "${F2C_PB_CC}"
  "${F2C_GPB_CC}"
)
# Order matters: put /usr/include FIRST so the system protobuf headers shadow
# the or-tools-vendored ones that get injected by Fields2Cover's transitive
# include dirs. Using BEFORE on an absolute path is a no-op for PUBLIC
# propagation, so we also replicate this on the executable target below.
target_include_directories(f2c_grpc_proto BEFORE PUBLIC
  "${SYS_PROTO_GRPC_INCLUDE}"
  "${CMAKE_CURRENT_BINARY_DIR}"
)
target_link_libraries(f2c_grpc_proto PUBLIC
  ${SYS_GRPCPP_LIBRARIES}
  ${SYS_PROTOBUF_LIBRARIES}
)
target_link_directories(f2c_grpc_proto PUBLIC
  ${SYS_GRPCPP_LIBRARY_DIRS}
  ${SYS_PROTOBUF_LIBRARY_DIRS}
)
# protobuf-generated code uses deprecated-but-still-valid patterns; keep
# warnings from blowing up the build.
target_compile_options(f2c_grpc_proto PRIVATE
  -Wno-unused-parameter -Wno-deprecated-declarations
)

# ---- Utility library (shared with tests) ----------------------------------
#
# Compile the shim's utility modules into an OBJECT library so both the
# server executable and the unit-test binary can consume them without
# rebuilding. Matching the include/link pattern used for f2c_grpc_proto:
# system protobuf headers go FIRST, libs resolved via pkg-config.

add_library(f2c_grpc_util OBJECT
  util/wkt.cpp
  util/status.cpp
  util/conversions.cpp
  util/swath_conversions.cpp
  util/route_conversions.cpp
  server_lifecycle.cpp
)
target_include_directories(f2c_grpc_util BEFORE PUBLIC
  "${SYS_PROTO_GRPC_INCLUDE}"
  "${CMAKE_CURRENT_SOURCE_DIR}"
  "${CMAKE_CURRENT_BINARY_DIR}"
)
target_link_libraries(f2c_grpc_util PUBLIC
  f2c_grpc_proto
  Fields2Cover
  ${SYS_GRPCPP_LIBRARIES}
  ${SYS_PROTOBUF_LIBRARIES}
)
target_link_directories(f2c_grpc_util PUBLIC
  ${SYS_GRPCPP_LIBRARY_DIRS}
  ${SYS_PROTOBUF_LIBRARY_DIRS}
)
target_compile_features(f2c_grpc_util PUBLIC cxx_std_17)

# ---- Server executable ----------------------------------------------------

add_executable(f2c-grpc-server
  main.cpp
  service_impl.cpp
  handlers/parse_geojson.cpp
  handlers/transform_to_utm.cpp
  handlers/clone_field.cpp
  handlers/clone_robot.cpp
  handlers/generate_headland.cpp
  handlers/generate_swaths.cpp
  handlers/sort_swaths.cpp
  handlers/generate_route.cpp
  handlers/plan_path.cpp
  $<TARGET_OBJECTS:f2c_grpc_util>
)

target_include_directories(f2c-grpc-server BEFORE PRIVATE
  "${SYS_PROTO_GRPC_INCLUDE}"
  "${CMAKE_CURRENT_SOURCE_DIR}"
  "${CMAKE_CURRENT_BINARY_DIR}"
)

target_link_libraries(f2c-grpc-server PRIVATE
  f2c_grpc_proto
  Fields2Cover
  ${SYS_GRPCPP_LIBRARIES}
  ${SYS_PROTOBUF_LIBRARIES}
)
target_link_directories(f2c-grpc-server PRIVATE
  ${SYS_GRPCPP_LIBRARY_DIRS}
  ${SYS_PROTOBUF_LIBRARY_DIRS}
)

target_compile_features(f2c-grpc-server PRIVATE cxx_std_17)

# Install target — later phases (dockerization) rely on this path.
install(TARGETS f2c-grpc-server RUNTIME DESTINATION bin)

# ---- Tests ----------------------------------------------------------------
#
# Gated on BUILD_TESTING alone (not on GNUPLOT_FOUND like the parent test
# suite) — the shim's util tests don't need gnuplot.

if(BUILD_TESTING)
  enable_testing()
  add_subdirectory(tests)
endif()
