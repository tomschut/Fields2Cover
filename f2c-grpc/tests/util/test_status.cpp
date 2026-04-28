// Unit tests for util/status.{h,cpp} — exception -> grpc::Status mapping.

#include "util/status.h"

#include <gtest/gtest.h>

#include <exception>
#include <stdexcept>
#include <string>

using f2c_grpc::util::exceptionToStatus;

namespace {

class Custom : public std::exception {
 public:
  const char* what() const noexcept override { return "custom-kind"; }
};

}  // namespace

TEST(StatusMapping, InvalidArgumentMapsToInvalidArgument) {
  const auto s = exceptionToStatus(std::invalid_argument("bad field"));
  EXPECT_EQ(s.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
  EXPECT_EQ(s.error_message(), "bad field");
}

TEST(StatusMapping, OutOfRangeMapsToInvalidArgument) {
  const auto s = exceptionToStatus(std::out_of_range("missing key"));
  EXPECT_EQ(s.error_code(), grpc::StatusCode::INVALID_ARGUMENT);
  EXPECT_EQ(s.error_message(), "missing key");
}

TEST(StatusMapping, RuntimeErrorMapsToInternal) {
  const auto s = exceptionToStatus(std::runtime_error("f2c blew up"));
  EXPECT_EQ(s.error_code(), grpc::StatusCode::INTERNAL);
  EXPECT_EQ(s.error_message(), "f2c blew up");
}

TEST(StatusMapping, UnknownExceptionMapsToInternal) {
  const auto s = exceptionToStatus(Custom{});
  EXPECT_EQ(s.error_code(), grpc::StatusCode::INTERNAL);
  EXPECT_EQ(s.error_message(), "custom-kind");
}
