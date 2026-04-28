// util/swath_conversions.cpp — implementation of the proto <-> f2c
// converters declared in util/swath_conversions.h.
//
// All geometry crosses the wire as WKT bytes. We lean on f2c's built-in
// importFromWkt / exportToWkt helpers (inherited from
// f2c::types::Geometry<T, R>) so the conversion never has to touch raw
// OGR APIs here.

#include "util/swath_conversions.h"

#include <stdexcept>
#include <string>

#include "fields2cover/types/LineString.h"

namespace f2c_grpc::util {

namespace {

// Parse a WKT MULTIPOLYGON / POLYGON into an f2c Cells. Throws
// std::invalid_argument on empty input or OGR parse error.
f2c::types::Cells cellsFromWkt(const std::string& wkt, const char* context) {
  if (wkt.empty()) {
    throw std::invalid_argument(
        std::string(context) + ": geometry WKT is empty");
  }
  f2c::types::Cells cells;
  try {
    cells.importFromWkt(wkt);
  } catch (const std::exception& e) {
    throw std::invalid_argument(
        std::string(context) + ": failed to parse WKT — " + e.what());
  }
  return cells;
}

}  // namespace

// ---------------------------------------------------------------------
// Local Field converters (stand-in for plan 10-03's util::field*Proto)
// ---------------------------------------------------------------------

f2c::types::Field fieldFromProtoLocal(const f2c::v1::Field& proto) {
  f2c::types::Cells cells =
      cellsFromWkt(proto.geometry_wkt(), "Field.geometry_wkt");
  f2c::types::Field field(cells, proto.id());
  if (proto.has_crs()) {
    const auto& crs = proto.crs();
    if (crs.epsg() != 0) {
      field.setEPSGCoordSystem(crs.epsg());
    } else if (!crs.raw().empty()) {
      field.setCRS(crs.raw());
    }
  }
  f2c::types::Point ref(proto.ref_point().x(), proto.ref_point().y());
  field.setRefPoint(ref);
  return field;
}

f2c::v1::Field fieldToProtoLocal(const f2c::types::Field& field,
                                 const f2c::v1::Field& source) {
  f2c::v1::Field out;
  // Preserve client-supplied metadata verbatim so the RPC is transparent
  // for the id / crs / ref_point fields. We only overwrite geometry_wkt.
  out.set_id(source.id());
  if (source.has_crs()) {
    *out.mutable_crs() = source.crs();
  }
  *out.mutable_ref_point() = source.ref_point();
  out.set_geometry_wkt(field.getField().exportToWkt());
  return out;
}

// ---------------------------------------------------------------------
// FieldWithHeadlands
// ---------------------------------------------------------------------

FieldWithHeadlandsNative fieldWithHeadlandsFromProto(
    const f2c::v1::FieldWithHeadlands& proto) {
  FieldWithHeadlandsNative out;
  out.field = fieldFromProtoLocal(proto.field());

  if (!proto.inner_field_wkt().empty()) {
    out.inner_cells =
        cellsFromWkt(proto.inner_field_wkt(),
                     "FieldWithHeadlands.inner_field_wkt");
  } else {
    // Fall back to the base field's cells — happens when a client sends a
    // FieldWithHeadlands without pre-computed inner area (not the normal
    // flow, but keeps the conversion total).
    out.inner_cells = out.field.getField();
  }

  // Headlands are optional on the from-proto path: they only matter for
  // GenerateRoute downstream (plan 10-05). Skip parsing if empty.
  if (!proto.headlands_wkt().empty()) {
    try {
      out.headlands.importFromWkt(proto.headlands_wkt());
    } catch (const std::exception&) {
      // Headland WKT may legitimately be MULTILINESTRING (ring swaths),
      // which Cells won't accept. Leave out.headlands empty in that
      // case — GenerateSwaths/SortSwaths don't need it.
      out.headlands = f2c::types::Cells();
    }
  }

  out.width_m = proto.headland_width_m();
  out.count = proto.headland_count();
  return out;
}

f2c::v1::FieldWithHeadlands fieldWithHeadlandsToProto(
    const f2c::types::Field& base_field,
    const f2c::types::Cells& inner_cells,
    const f2c::types::Cells& headland_rings,
    double width_m, int count) {
  f2c::v1::FieldWithHeadlands out;

  // Echo the base field's metadata through an empty source shell, so the
  // id / crs / ref_point survive the round-trip.
  f2c::v1::Field source;
  source.set_id(base_field.getId());
  // CRS: publish whatever f2c knows (EPSG if available, else raw string).
  if (base_field.isCoordSystemEPSG()) {
    source.mutable_crs()->set_epsg(base_field.getEPSGCoordSystem());
  }
  source.mutable_crs()->set_raw(base_field.getCRS());
  source.mutable_ref_point()->set_x(base_field.getRefPoint().getX());
  source.mutable_ref_point()->set_y(base_field.getRefPoint().getY());

  *out.mutable_field() = fieldToProtoLocal(base_field, source);
  // Overwrite geometry_wkt with the ORIGINAL (outer) cells — fieldToProto
  // already did that because base_field.getField() is the outer cells.
  if (headland_rings.size() > 0) {
    out.set_headlands_wkt(headland_rings.exportToWkt());
  }
  if (inner_cells.size() > 0) {
    out.set_inner_field_wkt(inner_cells.exportToWkt());
  }
  out.set_headland_width_m(width_m);
  out.set_headland_count(count);
  return out;
}

// ---------------------------------------------------------------------
// Swath / Swaths
// ---------------------------------------------------------------------

f2c::types::Swath swathFromProto(const f2c::v1::Swath& proto) {
  if (proto.path_wkt().empty()) {
    throw std::invalid_argument("Swath.path_wkt is required");
  }
  f2c::types::LineString path;
  try {
    path.importFromWkt(proto.path_wkt());
  } catch (const std::exception& e) {
    throw std::invalid_argument(
        std::string("Swath.path_wkt: invalid WKT — ") + e.what());
  }
  // Map proto SwathType -> f2c enum.
  f2c::types::SwathType type = f2c::types::SwathType::MAINLAND;
  if (proto.type() == f2c::v1::SWATH_TYPE_HEADLAND) {
    type = f2c::types::SwathType::HEADLAND;
  }
  f2c::types::Swath s(path, proto.width_m(), proto.id(), type);
  s.setCreationDir(proto.creation_dir());
  return s;
}

f2c::v1::Swath swathToProto(const f2c::types::Swath& swath) {
  f2c::v1::Swath out;
  out.set_id(swath.getId());
  out.set_path_wkt(swath.getPath().exportToWkt());
  out.set_width_m(swath.getWidth());
  // f2c SwathType -> proto enum.
  switch (swath.getType()) {
    case f2c::types::SwathType::HEADLAND:
      out.set_type(f2c::v1::SWATH_TYPE_HEADLAND);
      break;
    case f2c::types::SwathType::MAINLAND:
    default:
      out.set_type(f2c::v1::SWATH_TYPE_MAINLAND);
      break;
  }
  out.set_creation_dir(swath.getCreationDir());
  return out;
}

f2c::types::Swaths swathsFromProto(const f2c::v1::Swaths& proto) {
  f2c::types::Swaths out;
  for (const auto& p : proto.items()) {
    out.emplace_back(swathFromProto(p));
  }
  return out;
}

f2c::v1::Swaths swathsToProto(const f2c::types::Swaths& swaths,
                              const f2c::v1::FieldWithHeadlands& parent,
                              bool sorted) {
  f2c::v1::Swaths out;
  *out.mutable_field() = parent;
  for (size_t i = 0; i < swaths.size(); ++i) {
    *out.add_items() = swathToProto(swaths[i]);
  }
  out.set_sorted(sorted);
  return out;
}

}  // namespace f2c_grpc::util
