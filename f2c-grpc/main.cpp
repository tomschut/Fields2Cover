// f2c-grpc-server — production entry point.
//
// Listens on a Unix domain socket (default /tmp/f2c.sock, override via
// $F2C_GRPC_SOCKET) and exposes the F2CServiceImpl handlers. Graceful
// shutdown on SIGTERM or SIGINT.
//
// Shutdown flow: the signal handler just sets an atomic flag + writes a
// byte to a self-pipe to wake a dedicated shutdown watcher thread. That
// watcher is the one that calls grpc::Server::Shutdown() — doing so
// directly from the signal handler races with Wait()'s internal absl
// mutex acquisition and triggers a deadlock/abort in grpc 1.51.

#include "server_lifecycle.h"

#include <grpcpp/grpcpp.h>
#include <unistd.h>

#include <atomic>
#include <condition_variable>
#include <csignal>
#include <cstdlib>
#include <iostream>
#include <mutex>
#include <string>
#include <thread>

namespace {

std::atomic<int> g_last_signal{0};
std::atomic<bool> g_shutdown_requested{false};
std::mutex g_shutdown_mu;
std::condition_variable g_shutdown_cv;

void HandleSignal(int signum) {
  g_last_signal.store(signum);
  g_shutdown_requested.store(true);
  g_shutdown_cv.notify_all();
}

}  // namespace

int main(int /*argc*/, char** /*argv*/) {
  const std::string path = f2c_grpc::ResolveSocketPath("/tmp/f2c.sock");
  f2c_grpc::UnlinkStaleSocket(path);
  const std::string uri = f2c_grpc::SocketPathToUri(path);

  try {
    auto handle = f2c_grpc::StartServer(uri);

    std::signal(SIGTERM, HandleSignal);
    std::signal(SIGINT, HandleSignal);

    // Shutdown watcher: waits until a signal handler flags shutdown, then
    // calls Shutdown() from a plain thread (not a signal handler context).
    grpc::Server* server_raw = handle.server.get();
    std::thread shutdown_watcher([server_raw]() {
      std::unique_lock<std::mutex> lk(g_shutdown_mu);
      g_shutdown_cv.wait(lk, [] { return g_shutdown_requested.load(); });
      std::cerr << "f2c-grpc-server: received signal "
                << g_last_signal.load() << ", shutting down..." << std::endl;
      server_raw->Shutdown();
    });

    std::cerr << "f2c-grpc-server: listening on " << uri << std::endl;
    handle.server->Wait();

    // Wait() returned — either a client-requested shutdown or our watcher
    // ran. Ensure the watcher wakes and joins cleanly.
    g_shutdown_requested.store(true);
    g_shutdown_cv.notify_all();
    if (shutdown_watcher.joinable()) {
      shutdown_watcher.join();
    }

    std::cerr << "f2c-grpc-server: server stopped cleanly" << std::endl;
    f2c_grpc::UnlinkStaleSocket(path);
    return 0;
  } catch (const std::exception& e) {
    std::cerr << "f2c-grpc-server: fatal: " << e.what() << std::endl;
    return 1;
  }
}
