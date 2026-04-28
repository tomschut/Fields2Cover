# libFuzzer harnesses for fields2cover (HRD-05)
#
# These binaries link against the main Fields2Cover library and add
# libFuzzer + ASan + UBSan instrumentation on top. Clang only.

set(F2C_FUZZ_FLAGS
  -fsanitize=fuzzer,address,undefined
  -fno-omit-frame-pointer
  -O1
  -g
)

function(f2c_add_fuzzer name src)
  add_executable(${name} ${src})
  target_link_libraries(${name} PRIVATE Fields2Cover)
  # nlohmann_json is PRIVATE on Fields2Cover, re-expose the include dirs
  # for harnesses that need to parse/validate JSON directly.
  target_include_directories(${name} PRIVATE
    ${CMAKE_BINARY_DIR}/_deps/json-src/include
  )
  target_compile_options(${name} PRIVATE ${F2C_FUZZ_FLAGS})
  target_link_options(${name} PRIVATE ${F2C_FUZZ_FLAGS})
endfunction()

f2c_add_fuzzer(fuzz_geojson_parser fuzz_geojson_parser.cpp)
f2c_add_fuzzer(fuzz_wkt_parser     fuzz_wkt_parser.cpp)
f2c_add_fuzzer(fuzz_route_gen      fuzz_route_gen.cpp)
