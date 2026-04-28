// util/route_conversions.cpp — Route/Path <-> proto converters.
//
// See header for the contract. Implementation notes:
//
//  * The swath path geometry crosses the wire as a WKT LINESTRING per swath
//    item. We round-trip it through f2c::types::LineString::importFromWkt
//    (inherited from Geometry<>) to get the f2c representation back.
//
//  * Route connections cross the wire as WKT MULTIPOINT strings via
//    f2c::types::MultiPoint::exportToWkt / importFromWkt. This matches the
//    v1 proto RouteConnection.points_wkt field.
//
//  * Path uses f2c::types::PathState enums whose underlying values are the
//    f2c numeric constants (FORWARD=1, BACKWARD=-1, SWATH=1, TURN=2,
//    HL_SWATH=3). We map those explicitly to the proto enum values rather
//    than static_cast'ing — the two value spaces are not guaranteed to line
//    up now or in the future.

#include "util/route_conversions.h"

#include <stdexcept>
#include <string>

#include "fields2cover/types/LineString.h"
#include "fields2cover/types/MultiPoint.h"
#include "fields2cover/types/PathState.h"
#include "fields2cover/types/Swath.h"
#include "fields2cover/types/Swaths.h"

namespace f2c_grpc::util {

namespace {

f2c::v1::SwathType swathTypeToProto(f2c::types::SwathType t) {
  switch (t) {
    case f2c::types::SwathType::MAINLAND:
      return f2c::v1::SWATH_TYPE_MAINLAND;
    case f2c::types::SwathType::HEADLAND:
      return f2c::v1::SWATH_TYPE_HEADLAND;
  }
  return f2c::v1::SWATH_TYPE_UNSPECIFIED;
}

f2c::types::SwathType swathTypeFromProto(f2c::v1::SwathType t) {
  switch (t) {
    case f2c::v1::SWATH_TYPE_HEADLAND:
      return f2c::types::SwathType::HEADLAND;
    case f2c::v1::SWATH_TYPE_MAINLAND:
    case f2c::v1::SWATH_TYPE_UNSPECIFIED:
    default:
      return f2c::types::SwathType::MAINLAND;
  }
}

f2c::v1::PathDirection pathDirToProto(f2c::types::PathDirection d) {
  switch (d) {
    case f2c::types::PathDirection::FORWARD:
      return f2c::v1::PATH_DIRECTION_FORWARD;
    case f2c::types::PathDirection::BACKWARD:
      return f2c::v1::PATH_DIRECTION_BACKWARD;
  }
  return f2c::v1::PATH_DIRECTION_UNSPECIFIED;
}

f2c::v1::PathSectionType pathSectionToProto(f2c::types::PathSectionType t) {
  switch (t) {
    case f2c::types::PathSectionType::SWATH:
    case f2c::types::PathSectionType::HL_SWATH:
      // Both SWATH variants map to the proto's single SWATH section type —
      // the v1 proto does not distinguish headland swaths from mainland
      // swaths in PathState.section_type. The Go client can recover
      // mainland/headland from the Swath items if needed.
      return f2c::v1::PATH_SECTION_TYPE_SWATH;
    case f2c::types::PathSectionType::TURN:
      return f2c::v1::PATH_SECTION_TYPE_TURN;
  }
  return f2c::v1::PATH_SECTION_TYPE_UNSPECIFIED;
}

f2c::types::Swath swathFromProto(const f2c::v1::Swath& proto) {
  f2c::types::LineString path;
  if (!proto.path_wkt().empty()) {
    path.importFromWkt(std::string(proto.path_wkt()));
  }
  f2c::types::Swath out(path, proto.width_m(), proto.id(),
                        swathTypeFromProto(proto.type()));
  out.setCreationDir(proto.creation_dir());
  return out;
}

}  // namespace

// swathsFromProto + swathFromProto are provided by util/swath_conversions.*;
// do not redefine them here.

static void swathsToProtoFlatten(const f2c::types::Route& route,
                                 f2c::v1::Swaths* dst) {
  // Walk Route::v_swaths_ (via getVectorSwaths) and append every Swath into
  // dst->items in route order, preserving ids/widths/types.
  for (const auto& group : route.getVectorSwaths()) {
    for (const auto& sw : group) {
      auto* item = dst->add_items();
      item->set_id(sw.getId());
      f2c::types::LineString path = sw.getPath();
      item->set_path_wkt(path.exportToWkt());
      item->set_width_m(sw.getWidth());
      item->set_type(swathTypeToProto(sw.getType()));
      item->set_creation_dir(sw.getCreationDir());
    }
  }
  dst->set_sorted(true);
}

f2c::v1::Route routeToProto(const f2c::types::Route& route,
                            const f2c::v1::Swaths& parent_swaths) {
  f2c::v1::Route out;
  // Carry forward the FieldWithHeadlands envelope from the request so the
  // response is self-contained for PlanPath.
  *out.mutable_field() = parent_swaths.field();

  // Rebuild the flat swath sequence from the planner's output, not the raw
  // request swaths, because genRoute may have re-ordered / redirected them.
  auto* dst_swaths = out.mutable_swaths();
  *dst_swaths->mutable_field() = parent_swaths.field();
  swathsToProtoFlatten(route, dst_swaths);

  // Serialize turn connections as WKT MULTIPOINT.
  // sizeConnections() == sizeVectorSwaths() - 1 when v_swaths_ is populated
  // one group per swath (the route planner's normal layout). We best-effort
  // fill from/to swath ids by walking the flattened swath order.
  const std::vector<f2c::types::Swaths>& groups = route.getVectorSwaths();
  const size_t n_conn = route.sizeConnections();
  // Map each connection to the swath ids on either side by tracking the last
  // swath id we emitted and the first swath id of the next group.
  // Layout from f2c::rp::RoutePlannerBase::transformSolutionToRoute:
  //   connection[i] connects the end of group[i] to the start of group[i+1].
  for (size_t i = 0; i < n_conn; ++i) {
    const f2c::types::MultiPoint& mp = route.getConnection(i);
    auto* c = out.add_connections();
    int32_t from_id = 0;
    int32_t to_id = 0;
    if (i < groups.size() && groups[i].size() > 0) {
      from_id = groups[i].back().getId();
    }
    if ((i + 1) < groups.size() && groups[i + 1].size() > 0) {
      to_id = groups[i + 1].at(0).getId();
    }
    c->set_from_swath_id(from_id);
    c->set_to_swath_id(to_id);
    // exportToWkt() on an empty MultiPoint yields "MULTIPOINT EMPTY" which
    // round-trips fine.
    c->set_points_wkt(
        const_cast<f2c::types::MultiPoint&>(mp).exportToWkt());
  }
  return out;
}

f2c::types::Route routeFromProto(const f2c::v1::Route& proto) {
  // PathPlanning::planPath(robot, route, turn) walks v_swaths_ and
  // connections_ in order. Rebuild the minimal shape it needs:
  //   * one Swaths group per connection boundary
  //   * connections_[i] between group[i] and group[i+1]
  //
  // The simplest faithful round-trip is: one group per swath item (size 1
  // Swaths each), with connections proto <-> MultiPoint preserved positionally.
  // This mirrors how genRoute lays out v_swaths_ after optimization (one
  // swath per group, connected by headland turns).
  f2c::types::Route out;

  const auto& items = proto.swaths().items();
  if (items.empty()) {
    return out;
  }

  // Seed the first group with swath 0.
  out.addSwaths();
  out.getLastSwaths().emplace_back(swathFromProto(items.Get(0)));

  // For each subsequent swath, emit a connection (from proto.connections if
  // present, otherwise an empty MultiPoint) followed by a new swath group.
  for (int i = 1; i < items.size(); ++i) {
    f2c::types::MultiPoint mp;
    const int conn_idx = i - 1;
    if (conn_idx < proto.connections_size()) {
      const auto& wkt = proto.connections(conn_idx).points_wkt();
      if (!wkt.empty()) {
        // "MULTIPOINT EMPTY" is valid WKT but f2c's importFromWkt is picky —
        // only import when the payload is non-empty.
        mp.importFromWkt(std::string(wkt));
      }
    }
    out.addConnectedSwaths(mp, {});
    out.getLastSwaths().emplace_back(swathFromProto(items.Get(i)));
  }
  return out;
}

f2c::v1::Path pathToProto(const f2c::types::Path& path,
                          const f2c::v1::FieldWithHeadlands& parent) {
  f2c::v1::Path out;
  *out.mutable_field() = parent;
  out.set_length_m(path.length());
  out.set_task_time_s(path.getTaskTime());
  for (size_t i = 0; i < path.size(); ++i) {
    const f2c::types::PathState& s = path.getState(i);
    auto* ps = out.add_states();
    ps->mutable_point()->set_x(s.point.getX());
    ps->mutable_point()->set_y(s.point.getY());
    ps->set_angle_rad(s.angle);
    ps->set_length_m(s.len);
    ps->set_velocity(s.velocity);
    ps->set_direction(pathDirToProto(s.dir));
    ps->set_section_type(pathSectionToProto(s.type));
  }
  return out;
}

}  // namespace f2c_grpc::util
