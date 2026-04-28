// util/route_conversions.h — proto <-> f2c converters for Route and Path.
//
// These are the terminal converters for the coverage pipeline. GenerateRoute
// calls routeToProto() to serialize its output, PlanPath calls routeFromProto()
// to reconstruct an f2c Route from the incoming request and pathToProto() to
// serialize its output.

#pragma once

#include <generated/f2c.pb.h>

#include "fields2cover/types/Path.h"
#include "fields2cover/types/Route.h"
#include "fields2cover/types/Swaths.h"

namespace f2c_grpc::util {

// Serialize an f2c Route into a proto Route. `parent_swaths` is the
// GenerateRouteRequest's Swaths field — it carries both the FieldWithHeadlands
// envelope and the sorted swath items that we echo forward into the response.
// `parent_swaths.sorted()` is forced to true on the output.
f2c::v1::Route routeToProto(const f2c::types::Route& route,
                            const f2c::v1::Swaths& parent_swaths);

// Reverse direction used by PlanPath. Walks `proto.swaths().items()` into an
// f2c Swaths group and `proto.connections()` into the matching list of
// MultiPoint connections via WKT round-trip. The result carries enough
// structure for f2c::pp::PathPlanning::planPath to produce a valid path.
f2c::types::Route routeFromProto(const f2c::v1::Route& proto);

// Serialize an f2c Path to proto, embedding the supplied parent field (taken
// from the originating PlanPathRequest.route().field()).
f2c::v1::Path pathToProto(const f2c::types::Path& path,
                          const f2c::v1::FieldWithHeadlands& parent);

// swathsFromProto / swathsToProto are provided by util/swath_conversions.h.

}  // namespace f2c_grpc::util
