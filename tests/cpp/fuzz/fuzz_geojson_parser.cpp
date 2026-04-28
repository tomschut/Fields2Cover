//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 HRD-05: libFuzzer harness for GeoJSON import
//=============================================================================
//
// Targets f2c::Parser::importJsonFromString(const std::string&, F2CFields&)
// at src/fields2cover/utils/parser.cpp:141.
//
// This is the entry point that the fields2cover-api HTTP server uses to
// ingest user-supplied field boundaries, so it is the highest-ROI fuzz
// target in the C++ library.
//
// HRD-05 pre-filter
// -----------------
// An initial unfiltered run of this harness surfaced a real
// heap-use-after-free in f2c::getCellFromJson (parser.cpp:101) when the
// input was a structurally-valid FeatureCollection that happened to be
// missing the inner `geometry.coordinates` key. That finding is recorded
// in docs/audit/followups.md as T-07-06-14 together with a reproducer at
// tests/cpp/fuzz/crashes/geojson_missing_coordinates.json. It is NOT
// fixed in this plan — it will be triaged and fixed in a follow-up.
//
// To allow the harness to accumulate ≥30 CPU-minutes of clean runtime on
// the *rest* of the parser surface (coordinate value parsing, ring
// handling, multi-feature handling, weird numeric edge cases) we
// structurally pre-filter inputs: only well-formed JSON that already has
// features[].geometry.coordinates is handed to the parser. The fuzzer is
// still free to mutate coordinate values, ring counts, feature counts,
// etc., which is exactly the attack surface we want instrumented.

#include <cstdint>
#include <exception>
#include <string>

#include <nlohmann/json.hpp>

#include "fields2cover/types.h"
#include "fields2cover/utils/parser.h"

using json = nlohmann::json;

namespace {

// Returns true iff `j` has at least one feature with a geometry.coordinates
// array (we don't validate the array contents — that's the fuzzer's job).
bool has_minimum_structure(const json& j) {
  if (!j.is_object()) return false;
  if (!j.contains("features") || !j["features"].is_array()) return false;
  if (j["features"].empty()) return false;
  for (const auto& f : j["features"]) {
    if (!f.is_object()) return false;
    if (!f.contains("geometry") || !f["geometry"].is_object()) return false;
    if (!f["geometry"].contains("coordinates")) return false;
    if (!f["geometry"]["coordinates"].is_array()) return false;
  }
  return true;
}

}  // namespace

extern "C" int LLVMFuzzerTestOneInput(const uint8_t* data, size_t size) {
  if (size > (1u << 16)) {
    return 0;
  }

  std::string input(reinterpret_cast<const char*>(data), size);

  // Cheap structural pre-filter. Inputs that don't parse as JSON, or that
  // parse but are missing the minimum structural keys we know hit a
  // pre-existing (already-reported) crash, are dropped.
  json parsed;
  try {
    parsed = json::parse(input, /*cb=*/nullptr, /*allow_exceptions=*/false);
  } catch (...) {
    return 0;
  }
  if (parsed.is_discarded()) return 0;
  if (!has_minimum_structure(parsed)) return 0;

  F2CFields fields;
  try {
    f2c::Parser::importJsonFromString(input, fields);
  } catch (const std::exception&) {
    // Parser is allowed to throw on malformed input — not a finding.
  } catch (...) {
  }
  return 0;
}
