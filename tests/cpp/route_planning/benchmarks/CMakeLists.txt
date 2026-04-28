# Phase 5 SLV-01: OR-Tools baseline benchmark
# Standalone executable; not part of the unittests binary.

add_executable(route_planner_benchmark
  ortools_benchmark.cpp
)

target_include_directories(route_planner_benchmark
  PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)

target_link_libraries(route_planner_benchmark
  PRIVATE
    Fields2Cover
)

# Not registered with add_test() — this is a benchmark, not a unit test.
# Run manually:  ./tests/cpp/route_planning/benchmarks/route_planner_benchmark
#                [output-csv-path]
