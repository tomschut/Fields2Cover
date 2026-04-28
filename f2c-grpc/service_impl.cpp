#include "service_impl.h"

namespace f2c_grpc {

// All 9 RPC methods are implemented in f2c-grpc/handlers/*.cpp:
//   ParseGeoJSON        → handlers/parse_geojson.cpp        (plan 10-03)
//   TransformToUTM      → handlers/transform_to_utm.cpp     (plan 10-03)
//   CloneField          → handlers/clone_field.cpp          (plan 10-03)
//   CloneRobot          → handlers/clone_robot.cpp          (plan 10-03)
//   GenerateHeadland    → handlers/generate_headland.cpp    (plan 10-04)
//   GenerateSwaths      → handlers/generate_swaths.cpp      (plan 10-04)
//   SortSwaths          → handlers/sort_swaths.cpp          (plan 10-04)
//   GenerateRoute       → handlers/generate_route.cpp       (plan 10-05)
//   PlanPath            → handlers/plan_path.cpp            (plan 10-05)

}  // namespace f2c_grpc
