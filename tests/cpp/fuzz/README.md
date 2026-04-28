# Fields2Cover Fuzzers (HRD-05)

libFuzzer harnesses for the highest-value external-input entry points.

## Targets

| Harness                  | Target API                                   | Trust boundary           |
| ------------------------ | -------------------------------------------- | ------------------------ |
| `fuzz_geojson_parser`    | `f2c::Parser::importJsonFromString`          | External user input      |
| `fuzz_wkt_parser`        | `OGRGeometryFactory::createFromWkt` (GDAL)   | External user input      |
| `fuzz_route_gen`         | `f2c::rp::RoutePlannerBase::genRoute`        | Solver on weird geometry |

> Note: the plan originally named a GML parser harness. After Phase 8 the HTTP
> API no longer exposes GML ingest, so the equivalent JSON (GeoJSON) entry
> point — which IS user-reachable through the API — was used instead. The
> `f2c::Parser::importFieldGml` C++ function still exists but is not reachable
> from the public API and has low fuzzing ROI.

## Build

Clang only:

```bash
cmake -S . -B build-fuzz \
  -DENABLE_FUZZERS=ON \
  -DBUILD_TESTS=OFF \
  -DBUILD_TESTING=OFF \
  -DBUILD_PYTHON=OFF \
  -DBUILD_TUTORIALS=OFF \
  -DCMAKE_C_COMPILER=clang \
  -DCMAKE_CXX_COMPILER=clang++
cmake --build build-fuzz -j
```

## Run

Each invocation is 1800 seconds (30 minutes) — the HRD-05 gate.

```bash
./build-fuzz/tests/cpp/fuzz/fuzz_geojson_parser \
  tests/cpp/fuzz/corpora/geojson -max_total_time=1800 -print_final_stats=1

./build-fuzz/tests/cpp/fuzz/fuzz_wkt_parser \
  tests/cpp/fuzz/corpora/wkt -max_total_time=1800 -print_final_stats=1

./build-fuzz/tests/cpp/fuzz/fuzz_route_gen \
  tests/cpp/fuzz/corpora/route -max_total_time=1800 -print_final_stats=1
```

Any crash reproducer is saved to the current working directory as
`crash-<sha1>`. Leaks are reported by LSan in the tail of the run.
