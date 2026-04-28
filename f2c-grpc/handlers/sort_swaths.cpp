// handlers/sort_swaths.cpp — F2CServiceImpl::SortSwaths.
//
// Ports `openapi_server.controllers.route_controller.sort_swaths` to C++.
// Dispatches to one of the three single-cell-swath order planners
// (BoustrophedonOrder, SnakeOrder, SpiralOrder) depending on the
// requested SortAlgorithm.
//
// When start_point is set (and algorithm is not SPIRAL), iterates variants
// 0-3 and picks the variant whose first swath's first point is closest
// (Euclidean distance squared) to the requested start_point.

#include "service_impl.h"

#include <cstdio>
#include <cstring>
#include <limits>
#include <stdexcept>
#include <string>

#include "util/status.h"
#include "util/swath_conversions.h"

#include "fields2cover/route_planning/boustrophedon_order.h"
#include "fields2cover/route_planning/snake_order.h"
#include "fields2cover/route_planning/spiral_order.h"
#include "fields2cover/types.h"

namespace f2c_grpc {

grpc::Status F2CServiceImpl::SortSwaths(
    grpc::ServerContext* /*ctx*/,
    const f2c::v1::SortSwathsRequest* req,
    f2c::v1::SortSwathsResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument(
          "SortSwaths: null request or response");
    }
    if (req->swaths().items_size() == 0) {
      throw std::invalid_argument(
          "SortSwaths: at least one swath is required");
    }
    const int variant = req->variant();
    if (variant < 0 || variant > 3) {
      throw std::invalid_argument(
          "SortSwaths: variant must be in [0,3], got " +
          std::to_string(variant));
    }

    F2CSwaths in = util::swathsFromProto(req->swaths());

    // Lambda: sort with given algorithm and variant v.
    auto do_sort = [&](int v) -> F2CSwaths {
      switch (req->algorithm()) {
        case f2c::v1::SORT_ALGORITHM_SNAKE: {
          f2c::rp::SnakeOrder sorter;
          return sorter.genSortedSwaths(in, static_cast<uint32_t>(v));
        }
        case f2c::v1::SORT_ALGORITHM_SPIRAL: {
          f2c::rp::SpiralOrder sorter;
          return sorter.genSortedSwaths(in);
        }
        default: {  // BOUSTROPHEDON and UNSPECIFIED
          f2c::rp::BoustrophedonOrder sorter;
          return sorter.genSortedSwaths(in, static_cast<uint32_t>(v));
        }
      }
    };

    F2CSwaths sorted;

    if (req->has_start_point() &&
        req->algorithm() != f2c::v1::SORT_ALGORITHM_SPIRAL) {
      const double spx = req->start_point().x();
      const double spy = req->start_point().y();
      double best_dist_sq = std::numeric_limits<double>::max();
      bool found = false;

      for (int v = 0; v <= 3; ++v) {
        F2CSwaths candidate = do_sort(v);
        if (candidate.size() == 0) continue;

        // Extract first point of first swath via WKT: "LINESTRING (x y, ...)"
        std::string wkt = candidate.at(0).getPath().exportToWkt();
        const char* s = std::strstr(wkt.c_str(), "(");
        if (!s) continue;
        double px = 0.0, py = 0.0;
        if (std::sscanf(s + 1, "%lf %lf", &px, &py) != 2) continue;

        double dx = px - spx, dy = py - spy;
        double dist_sq = dx * dx + dy * dy;
        if (!found || dist_sq < best_dist_sq) {
          best_dist_sq = dist_sq;
          sorted = candidate;
          found = true;
        }
      }

      if (!found) {
        sorted = do_sort(0);
      }
    } else {
      sorted = do_sort(variant);
    }

    *resp->mutable_swaths() = util::swathsToProto(
        sorted, req->swaths().field(), /*sorted=*/true);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
