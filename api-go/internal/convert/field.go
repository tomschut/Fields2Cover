// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"encoding/json"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// geoJSONToWKTBytes serializes an OpenAPI GeoJSONFeatureCollection into
// the opaque []byte field the proto side carries.
//
// Three cases, in priority order:
//
//  1. The FeatureCollection is a "synthesized" one produced by
//     wktBytesToGeoJSON after a round-trip from the shim (real WKT
//     that could not be decoded as JSON). Those FCs carry the raw WKT
//     string in features[0].properties.wkt — return that verbatim so
//     the shim sees the same WKT it originally produced.
//
//  2. Empty/zero FeatureCollection — return nil so downstream callers
//     can treat "no geometry" uniformly.
//
//  3. Otherwise — JSON-encode the whole struct as a placeholder. Phase
//     11 does not translate GeoJSON to real WKT; the shim is the
//     source of truth for real WKT on the other side of the wire. This
//     fallback is preserved for handler unit tests that pre-date the
//     integration round-trip fix.
func geoJSONToWKTBytes(g api.GeoJSONFeatureCollection) []byte {
	if len(g.Features) == 0 && g.Type == "" && len(g.AdditionalProperties) == 0 {
		return nil
	}
	// Case 1: extract properties.wkt from the first feature if present.
	if len(g.Features) >= 1 && g.Features[0].Properties != nil {
		if raw, ok := (*g.Features[0].Properties)["wkt"]; ok {
			if s, ok := raw.(string); ok && s != "" {
				return []byte(s)
			}
		}
	}
	b, err := json.Marshal(g)
	if err != nil {
		return nil
	}
	return b
}

// wktBytesToGeoJSON reverses geoJSONToWKTBytes. If the bytes are valid
// JSON for a GeoJSONFeatureCollection (the placeholder case), they
// round-trip exactly. If unmarshal fails (because the shim returned
// real WKT), wrap the bytes in a synthetic FeatureCollection whose
// single Feature carries the raw WKT string under
// properties.wkt — handler tests assert presence, not semantic
// equivalence.
func wktBytesToGeoJSON(b []byte) api.GeoJSONFeatureCollection {
	if len(b) == 0 {
		return api.GeoJSONFeatureCollection{
			Type: api.FeatureCollection,
		}
	}
	var out api.GeoJSONFeatureCollection
	if err := json.Unmarshal(b, &out); err == nil && out.Type == api.FeatureCollection {
		return out
	}
	// Synthesize: wrap the raw WKT string in properties.wkt so the
	// response is still a valid GeoJSONFeatureCollection.
	props := map[string]interface{}{"wkt": string(b)}
	return api.GeoJSONFeatureCollection{
		Type: api.FeatureCollection,
		Features: []struct {
			Geometry   map[string]interface{}               `json:"geometry"`
			Properties *map[string]interface{}              `json:"properties"`
			Type       api.GeoJSONFeatureCollectionFeaturesType `json:"type"`
		}{
			{
				Geometry:   map[string]interface{}{},
				Properties: &props,
				Type:       api.Feature,
			},
		},
	}
}

// refPointToProto converts the OpenAPI [x, y] slice into a proto Point.
// Returns nil for absent or malformed input.
func refPointToProto(in *[]float64) *f2cclient.Point {
	if in == nil || len(*in) < 2 {
		return nil
	}
	pts := *in
	return &f2cclient.Point{X: pts[0], Y: pts[1]}
}

// refPointFromProto reverses refPointToProto.
func refPointFromProto(in *f2cclient.Point) *[]float64 {
	if in == nil {
		return nil
	}
	out := []float64{in.X, in.Y}
	return &out
}

// FieldToProto converts the OpenAPI Field envelope to a proto Field.
func FieldToProto(in api.Field) *f2cclient.Field {
	return &f2cclient.Field{
		Id:          in.Id,
		Crs:         CRSToProto(in.Crs),
		GeometryWkt: geoJSONToWKTBytes(in.Geometry),
		RefPoint:    refPointToProto(in.RefPoint),
	}
}

// FieldFromProto reverses FieldToProto. Nil input yields a zero-valued
// api.Field.
func FieldFromProto(in *f2cclient.Field) api.Field {
	if in == nil {
		return api.Field{}
	}
	return api.Field{
		Id:       in.Id,
		Crs:      CRSFromProto(in.Crs),
		Geometry: wktBytesToGeoJSON(in.GeometryWkt),
		RefPoint: refPointFromProto(in.RefPoint),
	}
}

// hcMarker is the AdditionalProperties key used to stash the proto
// HeadlandCount on the OpenAPI Headlands FeatureCollection so the
// round-trip Go->shim->Go->shim preserves it (the OpenAPI schema does
// not have a top-level HeadlandCount on FieldWithHeadlands, but the
// shim GenerateRoute RPC requires it).
const hcMarker = "f2c_headland_count"

// FieldWithHeadlandsToProto converts the OpenAPI FieldWithHeadlands
// envelope to its proto counterpart. HeadlandCount is read back from
// the Headlands FeatureCollection AdditionalProperties (see hcMarker)
// if the OpenAPI envelope was produced by FieldWithHeadlandsFromProto;
// otherwise it defaults to zero and the shim will apply its own
// default (3) where applicable.
func FieldWithHeadlandsToProto(in api.FieldWithHeadlands) *f2cclient.FieldWithHeadlands {
	out := &f2cclient.FieldWithHeadlands{
		Field:          FieldToProto(in.Field),
		HeadlandsWkt:   geoJSONToWKTBytes(in.Headlands),
		InnerFieldWkt:  geoJSONToWKTBytes(in.InnerField),
		HeadlandWidthM: in.HeadlandWidthM,
	}
	if v, ok := in.Headlands.AdditionalProperties[hcMarker]; ok {
		switch n := v.(type) {
		case float64:
			out.HeadlandCount = int32(n)
		case int:
			out.HeadlandCount = int32(n)
		case int32:
			out.HeadlandCount = n
		case int64:
			out.HeadlandCount = int32(n)
		}
	}
	return out
}

// FieldWithHeadlandsFromProto reverses FieldWithHeadlandsToProto.
// HeadlandCount from the proto is stashed on the Headlands
// FeatureCollection's AdditionalProperties under hcMarker so the
// downstream handlers can round-trip it back without being exposed on
// the OpenAPI schema surface.
func FieldWithHeadlandsFromProto(in *f2cclient.FieldWithHeadlands) api.FieldWithHeadlands {
	if in == nil {
		return api.FieldWithHeadlands{}
	}
	hl := wktBytesToGeoJSON(in.HeadlandsWkt)
	if in.HeadlandCount > 0 {
		if hl.AdditionalProperties == nil {
			hl.AdditionalProperties = make(map[string]interface{})
		}
		hl.AdditionalProperties[hcMarker] = float64(in.HeadlandCount)
	}
	return api.FieldWithHeadlands{
		Field:          FieldFromProto(in.Field),
		Headlands:      hl,
		InnerField:     wktBytesToGeoJSON(in.InnerFieldWkt),
		HeadlandWidthM: in.HeadlandWidthM,
	}
}
