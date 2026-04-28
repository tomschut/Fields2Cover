// server_lifecycle.h — reusable gRPC server bootstrap.
//
// Used by main.cpp (real server on /tmp/f2c.sock) and by the test fixture
// (temp socket per test file). Keeping the bootstrap in one place avoids
// drift between production and test wiring.

#pragma once

#include <grpcpp/grpcpp.h>

#include <memory>
#include <string>

#include "service_impl.h"

namespace f2c_grpc {

struct ServerHandle {
  std::unique_ptr<grpc::Server> server;
  std::unique_ptr<F2CServiceImpl> service;
  std::string listen_uri;  // e.g. "unix:/tmp/f2c.sock"
};

// Resolve the Unix socket path from $F2C_GRPC_SOCKET with fallback default.
std::string ResolveSocketPath(const char* default_path = "/tmp/f2c.sock");

// Build the gRPC listen URI for a Unix socket path (adds "unix:" prefix).
std::string SocketPathToUri(const std::string& path);

// Unlink a stale socket file at `path` if present (no error if missing).
void UnlinkStaleSocket(const std::string& path);

// Build + start a gRPC server listening on `socket_uri` (must already include
// the "unix:" prefix). Returns the running handle. Throws std::runtime_error
// on BuildAndStart failure.
ServerHandle StartServer(const std::string& socket_uri);

}  // namespace f2c_grpc
