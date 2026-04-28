// util/conversions.cpp — proto <-> f2c native type conversion.
//
// Implementation notes (DEVIATION from plan pseudocode):
//   * f2c::types::Field has private members; the plan's
//     `field.id_` / `field.coord_sys_` / `field.field_` direct access is
//     not legal. We use the public getters/setters (setId/getId,
//     setCRS/getCRS, getField/setField, setRefPoint/getRefPoint).
//   * The plan's three-arg Field constructor `Field(cells, id, ref_point)`
//     does not exist; only `Field()` and `Field(const Cells&, const std::string&)`
//     are declared. We construct with two args and call setRefPoint after.
//   * Robot has no `min_turning_radius_m`-as-default sentinel; we always
//     copy the value the proto carries. setMinTurningRadius accepts any
//     non-negative double in f2c.

#include "util/conversions.h"

#include "fields2cover/types/Cells.h"
#include "fields2cover/types/Point.h"

#include <stdexcept>
#include <string>

namespace f2c_grpc::util {

f2c::types::Field fieldFromProto(const f2c::v1::Field& proto) {
  if (proto.geometry_wkt().empty()) {
    throw std::invalid_argument("Field.geometry_wkt is required");
  }
  f2c::types::Cells cells;
  // Geometry::importFromWkt takes a std::string by const ref. proto's
  // bytes accessor returns a std::string (under the hood) so this is a
  // straightforward copy.
  cells.importFromWkt(std::string(proto.geometry_wkt()));

  f2c::types::Field field(cells, proto.id());
  field.setRefPoint(
      f2c::types::Point(proto.ref_point().x(), proto.ref_point().y()));

  // CRS: prefer explicit EPSG; fall back to the raw CRS string for
  // round-trip fidelity if the client supplied one.
  if (proto.crs().epsg() > 0) {
    field.setEPSGCoordSystem(proto.crs().epsg());
  } else if (!proto.crs().raw().empty()) {
    field.setCRS(proto.crs().raw());
  }
  return field;
}

// Parse a CRS string produced by f2c::Transform::transformToUTM such as
// "UTM:31N datum:WGS84" or "UTM:23S datum:ETRS89" into the matching
// standard EPSG code. Returns 0 for anything that isn't a recognised
// UTM CRS string — callers fall back to whatever f2c::Field reports
// directly. Pure-string parse, no f2c state mutation.
static int utmCrsStringToEpsg(const std::string& crs) {
  const auto prefix = std::string("UTM:");
  const auto pos = crs.find(prefix);
  if (pos == std::string::npos) {
    return 0;
  }
  size_t i = pos + prefix.size();
  int zone = 0;
  while (i < crs.size() && crs[i] >= '0' && crs[i] <= '9') {
    zone = zone * 10 + (crs[i] - '0');
    ++i;
  }
  if (zone < 1 || zone > 60 || i >= crs.size()) {
    return 0;
  }
  const char hemi = crs[i];
  if (hemi != 'N' && hemi != 'S') {
    return 0;
  }
  const bool etrs89 = crs.find("ETRS89") != std::string::npos;
  if (etrs89) {
    // ETRS89 / UTM: 25828..25838 (north only — ETRS89 is a European
    // datum, south-hemisphere zones aren't defined).
    return hemi == 'N' ? 25800 + zone : 0;
  }
  // WGS84 / UTM: 326xx (N) or 327xx (S).
  return (hemi == 'N' ? 32600 : 32700) + zone;
}

f2c::v1::Field fieldToProto(const f2c::types::Field& field) {
  f2c::v1::Field out;
  out.set_id(field.getId());

  auto* crs = out.mutable_crs();
  // getEPSGCoordSystem() returns a non-positive value when the CRS is
  // not an EPSG (e.g. unset, or a raw PROJ string). f2c::Transform::
  // transformToUTM sets the CRS to "UTM:<zone><N|S> datum:<WGS84|ETRS89>"
  // which getEPSGCoordSystem does not recognise, so we parse it here.
  int epsg = field.getEPSGCoordSystem();
  if (epsg <= 0) {
    epsg = utmCrsStringToEpsg(field.getCRS());
  }
  if (epsg > 0) {
    crs->set_epsg(epsg);
  }
  crs->set_raw(field.getCRS());

  // Geometry as WKT bytes via the OGR-backed exportToWkt path.
  out.set_geometry_wkt(field.getField().exportToWkt());

  auto* ref = out.mutable_ref_point();
  const auto& rp = field.getRefPoint();
  ref->set_x(rp.getX());
  ref->set_y(rp.getY());
  return out;
}

f2c::types::Robot robotFromProto(const f2c::v1::Robot& proto) {
  // Defaults match f2c::types::Robot's own constructor defaults
  // (max_curv=1.0, max_diff_curv=0.3) when the proto carries 0.
  f2c::types::Robot robot(
      proto.width_m(),
      proto.cov_width_m(),
      proto.max_curv() > 0 ? proto.max_curv() : 1.0,
      proto.max_diff_curv() > 0 ? proto.max_diff_curv() : 0.3);
  if (!proto.name().empty()) {
    robot.setName(proto.name());
  }
  if (proto.min_turning_radius_m() > 0) {
    robot.setMinTurningRadius(proto.min_turning_radius_m());
  }
  if (proto.cruise_vel_mps() > 0) {
    robot.setCruiseVel(proto.cruise_vel_mps());
  }
  if (proto.turn_vel_mps() > 0) {
    robot.setTurnVel(proto.turn_vel_mps());
  }
  return robot;
}

f2c::v1::Robot robotToProto(const f2c::types::Robot& robot) {
  f2c::v1::Robot out;
  out.set_name(robot.getName());
  out.set_width_m(robot.getWidth());
  out.set_cov_width_m(robot.getCovWidth());
  out.set_min_turning_radius_m(robot.getMinTurningRadius());
  out.set_max_curv(robot.getMaxCurv());
  out.set_max_diff_curv(robot.getMaxDiffCurv());
  out.set_cruise_vel_mps(robot.getCruiseVel());
  out.set_turn_vel_mps(robot.getTurnVel());
  return out;
}

}  // namespace f2c_grpc::util
