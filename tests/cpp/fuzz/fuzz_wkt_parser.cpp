//=============================================================================
//    Copyright (C) 2026 Fields2Cover contributors - BSD-3
//    Phase 7 HRD-05: libFuzzer harness for WKT parsing
//=============================================================================
//
// Fields2Cover uses OGRGeometryFactory::createFromWkt (GDAL) to parse WKT
// strings internally — see src/fields2cover/utils/parser.cpp:73. GDAL's WKT
// parser is the de-facto WKT surface for the library, and any input-driven
// bug there is also a bug in f2c.

#include <cstdint>
#include <exception>
#include <string>

#include <ogr_geometry.h>

extern "C" int LLVMFuzzerTestOneInput(const uint8_t* data, size_t size) {
  if (size == 0 || size > (1u << 16)) {
    return 0;
  }

  // createFromWkt wants a null-terminated buffer it can advance a pointer
  // through. Build a local std::string so we get the NUL for free.
  std::string wkt(reinterpret_cast<const char*>(data), size);

  OGRGeometry* geom = nullptr;
  const char* wkt_cstr = wkt.c_str();
  try {
    OGRGeometryFactory::createFromWkt(&wkt_cstr, nullptr, &geom);
  } catch (const std::exception&) {
    // GDAL rarely throws, but tolerate it.
  } catch (...) {
  }
  if (geom != nullptr) {
    OGRGeometryFactory::destroyGeometry(geom);
  }
  return 0;
}
