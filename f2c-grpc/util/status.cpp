// util/status.cpp — implementation of exceptionToStatus.

#include "util/status.h"

#include <stdexcept>

namespace f2c_grpc::util {

grpc::Status exceptionToStatus(const std::exception& e) noexcept {
  // Most-specific first: invalid_argument and out_of_range both derive from
  // std::logic_error but carry request-validity semantics, so we map them to
  // INVALID_ARGUMENT. runtime_error -> INTERNAL. Everything else -> INTERNAL.
  if (dynamic_cast<const std::invalid_argument*>(&e) != nullptr) {
    return grpc::Status(grpc::StatusCode::INVALID_ARGUMENT, e.what());
  }
  if (dynamic_cast<const std::out_of_range*>(&e) != nullptr) {
    return grpc::Status(grpc::StatusCode::INVALID_ARGUMENT, e.what());
  }
  if (dynamic_cast<const std::runtime_error*>(&e) != nullptr) {
    return grpc::Status(grpc::StatusCode::INTERNAL, e.what());
  }
  return grpc::Status(grpc::StatusCode::INTERNAL, e.what());
}

}  // namespace f2c_grpc::util
