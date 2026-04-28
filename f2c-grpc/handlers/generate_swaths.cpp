// handlers/generate_swaths.cpp — F2CServiceImpl::GenerateSwaths.
//
// Ports `openapi_server.controllers.swath_controller.generate_swaths` to
// C++. Iterates EVERY cell of the inner cultivable area (not just the
// first) and merges per-cell swaths into a single Swaths collection with
// globally monotonic ids — matches PROJECT.md's "Multi-cell swath
// generation" active requirement.

#include "service_impl.h"

#include <stdexcept>
#include <string>

#include "util/status.h"
#include "util/swath_conversions.h"

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/types.h"

namespace f2c_grpc {

grpc::Status F2CServiceImpl::GenerateSwaths(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::GenerateSwathsRequest* req,
    f2c::v1::GenerateSwathsResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument(
          "GenerateSwaths: null request or response");
    }
    if (req->width_m() <= 0.0) {
      throw std::invalid_argument(
          "GenerateSwaths: width_m must be > 0");
    }

    auto fwh =
        util::fieldWithHeadlandsFromProto(req->field_with_headlands());
    if (fwh.inner_cells.size() == 0) {
      throw std::invalid_argument(
          "GenerateSwaths: inner cells are empty");
    }

    f2c::sg::BruteForce bf;
    F2CSwaths merged;
    int global_id = 0;

    // Multi-cell iteration: run BruteForce per-cell and assign globally
    // monotonic ids as we flatten results. Using the (angle, width, Cell)
    // overload of generateSwaths that lives on SwathGeneratorBase.
    for (size_t i = 0; i < fwh.inner_cells.size(); ++i) {
      const F2CCell cell = fwh.inner_cells.getGeometry(i);
      F2CSwaths cell_swaths =
          bf.generateSwaths(req->angle_rad(), req->width_m(), cell);
      for (size_t j = 0; j < cell_swaths.size(); ++j) {
        F2CSwath sw = cell_swaths[j];
        sw.setId(global_id++);
        merged.emplace_back(sw);
      }
    }

    if (merged.size() == 0) {
      throw std::invalid_argument(
          "GenerateSwaths: no swaths generated "
          "(field too small for the given width?)");
    }

    *resp->mutable_swaths() = util::swathsToProto(
        merged, req->field_with_headlands(), /*sorted=*/false);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
