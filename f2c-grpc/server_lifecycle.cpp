// server_lifecycle.cpp — see server_lifecycle.h for rationale.

#include "server_lifecycle.h"

#include <grpcpp/grpcpp.h>
#include <sys/stat.h>
#include <unistd.h>

#include <cstdlib>
#include <stdexcept>
#include <string>

namespace f2c_grpc {

std::string ResolveSocketPath(const char* default_path) {
  const char* env = std::getenv("F2C_GRPC_SOCKET");
  if (env != nullptr && env[0] != '\0') {
    return std::string(env);
  }
  return std::string(default_path);
}

std::string SocketPathToUri(const std::string& path) {
  return "unix:" + path;
}

void UnlinkStaleSocket(const std::string& path) {
  struct stat st;
  if (::stat(path.c_str(), &st) == 0) {
    ::unlink(path.c_str());  // best-effort
  }
}

ServerHandle StartServer(const std::string& socket_uri) {
  ServerHandle handle;
  handle.listen_uri = socket_uri;
  handle.service = std::make_unique<F2CServiceImpl>();

  grpc::ServerBuilder builder;
  builder.AddListeningPort(socket_uri, grpc::InsecureServerCredentials());
  builder.RegisterService(handle.service.get());
  handle.server = builder.BuildAndStart();

  if (handle.server == nullptr) {
    throw std::runtime_error(
        "StartServer: BuildAndStart returned null for " + socket_uri);
  }
  return handle;
}

}  // namespace f2c_grpc
