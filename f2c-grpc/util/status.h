// util/status.h — map C++ exceptions to grpc::Status per the phase-10
// decisions table. Every handler wraps its body in:
//
//     try { /* f2c + conversion */ return grpc::Status::OK; }
//     catch (const std::exception& e) { return exceptionToStatus(e); }

#pragma once

#include <exception>

#include <grpcpp/grpcpp.h>

namespace f2c_grpc::util {

// Maps a C++ exception to a grpc::Status:
//   std::invalid_argument -> INVALID_ARGUMENT
//   std::out_of_range     -> INVALID_ARGUMENT
//   std::runtime_error    -> INTERNAL
//   any other std::exception -> INTERNAL
// In all cases the status message is `e.what()`.
grpc::Status exceptionToStatus(const std::exception& e) noexcept;

}  // namespace f2c_grpc::util
