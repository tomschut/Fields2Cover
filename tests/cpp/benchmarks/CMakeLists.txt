# Phase 7 PRF-01: swath / headland / path-interp benchmarks (Plan 07-07)
# Standalone executables — not part of the unittests binary.

function(f2c_add_bench name src)
  add_executable(${name} ${src})
  target_include_directories(${name} PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
  target_link_libraries(${name} PRIVATE Fields2Cover)
endfunction()

f2c_add_bench(bench_swath        bench_swath.cpp)
f2c_add_bench(bench_headland     bench_headland.cpp)
f2c_add_bench(bench_path_interp  bench_path_interp.cpp)

# Not registered with add_test() — these are benchmarks, not unit tests.
# Run manually via scripts/run-benchmarks.sh.
