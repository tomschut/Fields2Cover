// util/conversions.h — proto <-> f2c native type conversion.
//
// Each handler calls these helpers at its boundaries. Geometry fields cross
// the wire as WKT bytes (see util/wkt.h) and enter/leave f2c as Cells via
// the OGR-backed importFromWkt / exportToWkt round-trip already provided by
// f2c::types::Geometry.

#pragma once

#include <generated/f2c.pb.h>

#include "fields2cover/types/Field.h"
#include "fields2cover/types/Robot.h"

namespace f2c_grpc::util {

// Copy a proto Field message into an f2c::types::Field. Throws
// std::invalid_argument if `geometry_wkt` is empty or malformed.
f2c::types::Field fieldFromProto(const f2c::v1::Field& proto);

// Serialize an f2c::types::Field into a proto Field. The returned message
// has geometry_wkt populated via the Cells WKT export.
f2c::v1::Field fieldToProto(const f2c::types::Field& field);

// Copy a proto Robot scalar-by-scalar into an f2c::types::Robot.
f2c::types::Robot robotFromProto(const f2c::v1::Robot& proto);

// Reverse direction.
f2c::v1::Robot robotToProto(const f2c::types::Robot& robot);

}  // namespace f2c_grpc::util
