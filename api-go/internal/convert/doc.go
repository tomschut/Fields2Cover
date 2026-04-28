// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package convert holds the hand-written translation layer between
// the OpenAPI types in internal/api and the protobuf-generated types
// in internal/f2cclient.
//
// Per CON-04 (Phase 9 contract design), the OpenAPI surface and the
// gRPC surface are deliberately disjoint — there is no schema sharing
// and no reflection-based glue. Every paired converter lives in this
// package as plain Go.
//
// File layout (one logical object per file):
//
//	crs.go     CRS
//	robot.go   Robot
//	field.go   Field, FieldWithHeadlands
//	swaths.go  Swath, Swaths
//	route.go   Route, RouteConnection
//	path.go    Path, PathState
//
// Naming convention: XToProto returns *f2cclient.X from an api.X;
// XFromProto returns api.X from a *f2cclient.X. Nil-input pointers
// produce zero-value outputs (callers can treat the conversion as
// total).
//
// Geometry strategy: the OpenAPI side carries inline GeoJSON (as a
// GeoJSONFeatureCollection struct or LineString sub-object), while the
// proto side carries opaque WKT bytes. Phase 11 does NOT semantically
// translate between the two. The pass-through strategy:
//
//   - api → proto: json.Marshal the GeoJSON struct into the *Wkt bytes
//     field as a placeholder. The shim ignores this on requests where
//     it doesn't need geometry input; for the few RPCs that accept
//     geometry, Phase 12 will wire real GeoJSON-to-WKT conversion.
//   - proto → api: try to json.Unmarshal the bytes back into the
//     GeoJSON struct. If unmarshal fails (because the shim returned
//     real WKT), wrap the bytes in a synthetic FeatureCollection with
//     a single Feature whose properties carry the raw WKT string.
//
// The handler tests assert presence and shape, not semantic
// equivalence. Phase 12 integration tests against the real shim will
// catch any genuine geometry mismatch.
package convert
