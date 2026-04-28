# libFuzzer Results — Phase 7 Plan 07-06

**Run:** 2026-04-11 (partial — agent hit rate limit during run)
**Harnesses:** 3 (GeoJSON, WKT, route planner `genRoute`)
**Commits:** `829ba75` (CMake wiring) + `e030f4d` (harnesses + corpora)
**Status:** Harnesses committed; fuzz runs partial. 3 crashes saved under `tests/cpp/fuzz/crashes/`.

## Harnesses

| # | File | Target | Seed |
|---|------|--------|------|
| 1 | `tests/cpp/fuzz/fuzz_geojson_parser.cpp` | `f2c::Parser::importJsonFromString(const std::string&, F2CFields&)` | `corpora/geojson/seed_01.json` (minimal valid FeatureCollection) |
| 2 | `tests/cpp/fuzz/fuzz_wkt_parser.cpp` | WKT → f2c geometry (exact entry point TBD in next run) | `corpora/wkt/seed_01.wkt` |
| 3 | `tests/cpp/fuzz/fuzz_route_gen.cpp` | `RoutePlannerBase::genRoute` | `corpora/route/seed_01.bin` |

## Crashes Found

### C-001 — GeoJSON parser: missing `coordinates` field crashes [FIXED commit d5223c1]

**Harness:** `fuzz_geojson_parser.cpp`
**Crash input:** `tests/cpp/fuzz/crashes/geojson_missing_coordinates.json`

The fuzzer mutated the seed's `"coordinates"` key into a garbled `"c1],[0,ates"` (and renamed `"Name"` → `"Naee"` for good measure). The parser walks `feature["geometry"]["coordinates"]` in `getCellFromJson()` without checking for key presence — `nlohmann::json::operator[]` on a missing key on an object is undefined / throws, and whatever path f2c hits produces a crash.

**Severity:** HIGH — reachable through the new `/parser/import-field-geojson` endpoint. Any client POST with a malformed FeatureCollection can crash the server.

**Proposed fix:** In `src/fields2cover/utils/parser.cpp::getCellFromJson`, check `.contains("geometry")` and `["geometry"].contains("coordinates")` before indexing. On missing key, throw `std::invalid_argument("GeoJSON feature missing geometry.coordinates")`. Then `import_field_geojson` controller catches and returns HTTP 400 `INVALID_GEOJSON_SCHEMA`.

**Tracked as:** T-025 in `docs/audit/followups.md`

---

### C-002 — Route planner: SEGV on 1-byte random input [FIXED commit a264674]

**Harness:** `fuzz_route_gen.cpp`
**Crash input:** `tests/cpp/fuzz/crashes/route_ortools_segv_1byte.bin` (1 byte: `0xa6`)

Single byte of random input, fed through the route planner fuzz entry point, produces a SEGV inside OR-Tools. Almost certainly the pre-existing OR-Tools brittleness flagged in 07-01/07-04 notes (empty/nonsense pipeline → routing.SolveWithParameters gives `Check failed: start >= 0 (-1 vs. 0)` or similar fatal).

**Severity:** HIGH for fuzzing CI but MEDIUM for production API — the route planner is only reached via `/route/generate`, where the body schema validation ensures at least a well-formed field + swaths before the controller runs. Fuzzing bypasses that layer.

**Proposed fix:** The `genRoute` harness itself should do structural pre-validation before calling into OR-Tools (reject empty/tiny inputs early). Alternatively, the upstream `RoutePlannerBase::computeBestRoute` could be hardened with additional pre-conditions (matching the `cov_graph.numNodes() == 0` guard added in the merge). Investigate whether pre-validation in the harness is enough, or whether library-level hardening is warranted.

**Tracked as:** T-026 in `docs/audit/followups.md`

---

### C-003 — Route planner: SEGV on 31-byte input (pattern `0x00 f0 00…fd ff…de`) [FIXED commit a264674]

**Harness:** `fuzz_route_gen.cpp`
**Crash input:** `tests/cpp/fuzz/crashes/route_ortools_segv_31bytes.bin` (31 bytes)

Same class as C-002 but with slightly larger input that gets past the initial byte check and into the graph-construction path before OR-Tools explodes. Bit pattern suggests an integer overflow / signed underflow path (`0xfd ffff ffff` is -3 as a signed 32-bit int).

**Severity:** HIGH — same reasoning as C-002. Same proposed fix.

**Tracked as:** T-027 in `docs/audit/followups.md`

---

## Run Statistics

**Budget:** `-max_total_time=1800` (30 CPU-minutes per harness planned).
**Actual:** Partial — agent hit rate limit mid-execution after ~45 minutes total. Exact per-harness run time not recorded because the runtime reporting was lost when the agent died.

**Next run (scheduled):** Re-execute all 3 harnesses with the same budget after fixing C-001 through C-003. Expect deeper coverage once the immediate crashers are resolved.

## How to Re-run

```bash
cmake -B build-fuzz -DENABLE_FUZZERS=ON -DCMAKE_CXX_COMPILER=clang++
cmake --build build-fuzz -j$(nproc)

# One at a time (fuzzers saturate a CPU):
./build-fuzz/tests/cpp/fuzz/fuzz_geojson_parser tests/cpp/fuzz/corpora/geojson/ -max_total_time=1800
./build-fuzz/tests/cpp/fuzz/fuzz_wkt_parser tests/cpp/fuzz/corpora/wkt/ -max_total_time=1800
./build-fuzz/tests/cpp/fuzz/fuzz_route_gen tests/cpp/fuzz/corpora/route/ -max_total_time=1800
```

New crashes get dropped into `tests/cpp/fuzz/crashes/` (not gitignored — they're regression fixtures).
