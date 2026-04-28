// handlers/generate_route.cpp — GenerateRoute RPC implementation.
//
// Wraps f2c::rp::RoutePlannerBase::genRoute using the upstream-merged
// signature (time_limit_seconds + search_for_optimum present at the call
// site, even when passing defaults).
//
// Input contract (proto):
//   GenerateRouteRequest {
//     Swaths swaths = 1;  // must be sorted=true, field must carry
//                         //  FieldWithHeadlands with headland_width_m,
//                         //  headland_count, and the original field cells
//                         //  WKT in .field.geometry_wkt
//   }
//
// Why we re-run generateHeadlandSwaths instead of parsing headlands_wkt:
//   The proto's FieldWithHeadlands.headlands_wkt is a flat WKT blob (one
//   MULTILINESTRING / MULTIPOLYGON for display), but the route planner
//   needs the LAYERED F2CCells the constant-headland generator produces
//   (one cell group per ring, indexed). The layered structure is NOT
//   round-trippable from a single WKT. The canonical Python reference
//   (route_controller.generate_route) re-runs generateHeadlandSwaths on
//   the field cells, picks ring index min(1, count-1), and hands that to
//   genRoute. We mirror that exactly.

#include "service_impl.h"
#include "util/route_conversions.h"
#include "util/status.h"
#include "util/swath_conversions.h"

#include "fields2cover/headland_generator/constant_headland.h"
#include "fields2cover/route_planning/route_planner_base.h"
#include "fields2cover/types/Cells.h"
#include "fields2cover/types/Swaths.h"
#include "fields2cover/types/SwathsByCells.h"

#include <algorithm>
#include <stdexcept>
#include <string>
#include <vector>

namespace f2c_grpc {

grpc::Status F2CServiceImpl::GenerateRoute(
    grpc::ServerContext*, const f2c::v1::GenerateRouteRequest* req,
    f2c::v1::GenerateRouteResponse* resp) {
  try {
    if (req == nullptr || resp == nullptr) {
      throw std::invalid_argument("GenerateRoute: null request or response");
    }
    const auto& proto_swaths = req->swaths();
    if (proto_swaths.items_size() == 0) {
      throw std::invalid_argument("GenerateRoute: swaths.items is empty");
    }
    if (!proto_swaths.sorted()) {
      throw std::invalid_argument(
          "GenerateRoute: swaths must be sorted (call SortSwaths first)");
    }
    const auto& fwh = proto_swaths.field();
    const auto& geom_wkt = fwh.field().geometry_wkt();
    if (geom_wkt.empty()) {
      throw std::invalid_argument(
          "GenerateRoute: Field.geometry_wkt is required to build headland "
          "rings");
    }
    const double headland_width = fwh.headland_width_m();
    if (headland_width <= 0.0) {
      throw std::invalid_argument(
          "GenerateRoute: FieldWithHeadlands.headland_width_m must be > 0 "
          "(run GenerateHeadland first)");
    }
    const int headland_count = fwh.headland_count();
    if (headland_count < 1) {
      throw std::invalid_argument(
          "GenerateRoute: FieldWithHeadlands.headland_count must be >= 1");
    }

    // Parse the field cells from WKT.
    f2c::types::Cells field_cells;
    field_cells.importFromWkt(std::string(geom_wkt));

    // Re-run the constant-headland generator to get the layered ring cells.
    // This matches the Python reference in route_controller.generate_route.
    f2c::hg::ConstHL const_hl;
    std::vector<f2c::types::Cells> hl_swath_cells =
        const_hl.generateHeadlandSwaths(
            field_cells, headland_width, headland_count, /*dir_out2in=*/false);
    if (hl_swath_cells.empty()) {
      throw std::runtime_error(
          "GenerateRoute: generateHeadlandSwaths returned no ring layers");
    }
    // Canonical picker: min(1, count-1) — ring 1 when at least 2 exist,
    // else ring 0.
    const size_t hl_idx =
        static_cast<size_t>(std::min(1, headland_count - 1));
    if (hl_idx >= hl_swath_cells.size()) {
      throw std::runtime_error(
          "GenerateRoute: headland ring index out of bounds");
    }
    const f2c::types::Cells& hl_rings = hl_swath_cells[hl_idx];

    // Defensive: f2c::rp::RoutePlannerBase::genRoute segfaults if the
    // chosen ring layer has zero cells (the connector routing tries to
    // build turns through an empty headland strip and dereferences a
    // null pointer inside or-tools VRP). This happens on degenerate
    // fields where headland_width_m * headland_count is large relative
    // to the field bbox — the constant-headland generator silently
    // produces empty ring layers instead of erroring. Catch the empty
    // case here and surface a clean InvalidArgument so clients see a
    // meaningful 400 instead of a 503 Unavailable / EOF crash.
    if (hl_rings.size() == 0) {
      throw std::invalid_argument(
          "GenerateRoute: chosen headland ring layer is empty — "
          "headland_width_m * headland_count is too large for the field "
          "(reduce headland_width_m, reduce headland_count, or use a "
          "larger field)");
    }

    // Wrap swaths as one-group SwathsByCells.
    f2c::types::Swaths swaths = util::swathsFromProto(proto_swaths);
    f2c::types::SwathsByCells sbc;
    sbc.push_back(swaths);

    // Upstream-merged genRoute signature — all seven arguments named
    // explicitly so a future header bump can't silently drop one.
    f2c::rp::RoutePlannerBase planner;
    const bool show_log = false;
    // d_tol = 0.5 not the 1e-4 default — the Python reference bumps it up
    // to absorb floating-point drift between separately-projected field +
    // swaths. See route_controller.generate_route for rationale.
    const double d_tol = 0.5;
    const bool redirect_swaths = true;
    const long int time_limit_seconds = 1;
    const bool search_for_optimum = false;

    f2c::types::Route route = planner.genRoute(
        hl_rings, sbc, show_log, d_tol, redirect_swaths,
        time_limit_seconds, search_for_optimum);

    *resp->mutable_route() = util::routeToProto(route, proto_swaths);
    return grpc::Status::OK;
  } catch (const std::exception& e) {
    return util::exceptionToStatus(e);
  }
}

}  // namespace f2c_grpc
