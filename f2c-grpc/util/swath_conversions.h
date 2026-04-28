// util/swath_conversions.h — proto <-> f2c converters for Swath, Swaths,
// and the FieldWithHeadlands composite. Used by handlers/generate_headland,
// handlers/generate_swaths, handlers/sort_swaths and will be reused by
// plan 10-05 (generate_route / plan_path).
//
// Note: this file does NOT depend on util/conversions.{h,cpp} (the Field
// and Robot converters from plan 10-03), because plan 10-04 runs in
// parallel with 10-03. Field <-> f2c conversion needed by GenerateHeadland
// is implemented inline in swath_conversions.cpp as a local WKT-based
// helper, and will be unified with plan 10-03's util/conversions.h at
// merge time.

#pragma once

#include <cstdint>
#include <string>

#include <generated/f2c.pb.h>

#include "fields2cover/types/Cells.h"
#include "fields2cover/types/Field.h"
#include "fields2cover/types/Swath.h"
#include "fields2cover/types/Swaths.h"

namespace f2c_grpc::util {

// Result of unpacking a proto FieldWithHeadlands into its native f2c
// counterparts. Carries every piece a downstream f2c call may need:
//   * `field`           — the base field (with its outer Cells from
//                         Field.geometry_wkt)
//   * `inner_cells`     — the inner cultivable Cells (from
//                         FieldWithHeadlands.inner_field_wkt). Falls back
//                         to `field.getField()` if the proto omits it.
//   * `headlands`       — the headland swath rings (from
//                         FieldWithHeadlands.headlands_wkt) expressed as
//                         Cells. May be empty.
//   * `width_m`/`count` — echo of the scalar fields (useful for downstream
//                         RPCs such as GenerateRoute).
struct FieldWithHeadlandsNative {
  f2c::types::Field field;
  f2c::types::Cells inner_cells;
  f2c::types::Cells headlands;
  double width_m = 0.0;
  int count = 0;
};

// Decode a proto FieldWithHeadlands into its native pieces. Throws
// std::invalid_argument if the embedded Field carries no geometry_wkt.
FieldWithHeadlandsNative fieldWithHeadlandsFromProto(
    const f2c::v1::FieldWithHeadlands& proto);

// Encode a FieldWithHeadlands composite back into the proto. `base_field`
// supplies Field.id / CRS / ref_point metadata; `inner_cells` is written
// to `inner_field_wkt`; `headland_rings` is written to `headlands_wkt`.
f2c::v1::FieldWithHeadlands fieldWithHeadlandsToProto(
    const f2c::types::Field& base_field,
    const f2c::types::Cells& inner_cells,
    const f2c::types::Cells& headland_rings,
    double width_m, int count);

// Swath <-> proto. swathFromProto throws std::invalid_argument if the
// proto Swath has an empty path_wkt. swathToProto maps the internal
// SwathType enum to the proto SwathType.
f2c::types::Swath swathFromProto(const f2c::v1::Swath& proto);
f2c::v1::Swath    swathToProto(const f2c::types::Swath& swath);

// Swaths collection <-> proto. `swathsFromProto` only decodes the `items`
// repeated field — the parent FieldWithHeadlands context must be handled
// by the caller. `swathsToProto` copies the supplied `parent` into the
// returned Swaths.field so the wire message is self-contained.
f2c::types::Swaths swathsFromProto(const f2c::v1::Swaths& proto);
f2c::v1::Swaths    swathsToProto(const f2c::types::Swaths& swaths,
                                 const f2c::v1::FieldWithHeadlands& parent,
                                 bool sorted);

// Convenience: decode a bare proto Field into an f2c Field by round-
// tripping its geometry_wkt through Cells.importFromWkt. Local helper
// that is a stand-in for plan 10-03's util::fieldFromProto.
f2c::types::Field fieldFromProtoLocal(const f2c::v1::Field& proto);

// Convenience: encode an f2c Field as a proto Field, carrying id /
// crs / ref_point across along with the exported WKT.
f2c::v1::Field fieldToProtoLocal(const f2c::types::Field& field,
                                 const f2c::v1::Field& source);

}  // namespace f2c_grpc::util
