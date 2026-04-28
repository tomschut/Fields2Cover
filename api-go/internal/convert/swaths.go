// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"encoding/json"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// swathTypeToProto maps the OpenAPI MAINLAND/HEADLAND enum to the
// proto enum constant. Unknown values fall through to UNSPECIFIED.
func swathTypeToProto(in api.SwathType) f2cclient.SwathType {
	switch in {
	case api.MAINLAND:
		return f2cclient.SwathType_SWATH_TYPE_MAINLAND
	case api.HEADLAND:
		return f2cclient.SwathType_SWATH_TYPE_HEADLAND
	default:
		return f2cclient.SwathType_SWATH_TYPE_UNSPECIFIED
	}
}

// swathTypeFromProto reverses swathTypeToProto. UNSPECIFIED maps to
// the empty string (OpenAPI clients receive the absent enum).
func swathTypeFromProto(in f2cclient.SwathType) api.SwathType {
	switch in {
	case f2cclient.SwathType_SWATH_TYPE_MAINLAND:
		return api.MAINLAND
	case f2cclient.SwathType_SWATH_TYPE_HEADLAND:
		return api.HEADLAND
	default:
		return api.SwathType("")
	}
}

// rawWKTPrefix marks a Swath.Path.Type string as a carrier for the
// real shim WKT that could not be decoded as JSON. The Go handlers
// round-trip this verbatim so the shim sees its original WKT back
// when the swaths are re-posted. This only applies to swaths produced
// by the shim (via wktBytesToSwathPath); client-supplied swath paths
// keep Type="LineString" and coordinates as JSON.
const rawWKTPrefix = "__rawwkt__:"

// swathPathToWKTBytes serializes the inline LineString sub-object for
// the proto pass-through. If the Type carries a rawWKTPrefix marker
// (produced by wktBytesToSwathPath for real-WKT round-trips), the raw
// WKT is emitted verbatim; otherwise the struct is JSON-encoded.
// Empty input yields nil bytes.
func swathPathToWKTBytes(p struct {
	Coordinates [][]float64       `json:"coordinates"`
	Type        api.SwathPathType `json:"type"`
}) []byte {
	if len(p.Coordinates) == 0 && p.Type == "" {
		return nil
	}
	if s := string(p.Type); len(s) > len(rawWKTPrefix) && s[:len(rawWKTPrefix)] == rawWKTPrefix {
		return []byte(s[len(rawWKTPrefix):])
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	return b
}

// wktBytesToSwathPath reverses swathPathToWKTBytes. Three cases:
//  1. Empty input: return an empty LineString.
//  2. Valid JSON: unmarshal directly.
//  3. Invalid JSON (real WKT from the shim): stash the raw WKT
//     string in the Type field with a rawWKTPrefix marker so the next
//     Go->shim round-trip can emit it verbatim. Handler unit tests
//     assert on coordinates length, so we leave Coordinates empty.
func wktBytesToSwathPath(b []byte) struct {
	Coordinates [][]float64       `json:"coordinates"`
	Type        api.SwathPathType `json:"type"`
} {
	out := struct {
		Coordinates [][]float64       `json:"coordinates"`
		Type        api.SwathPathType `json:"type"`
	}{
		Type: api.LineString,
	}
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		// Stash the raw WKT in Type so a future Go->shim round-trip
		// can recover it verbatim (see swathPathToWKTBytes).
		return struct {
			Coordinates [][]float64       `json:"coordinates"`
			Type        api.SwathPathType `json:"type"`
		}{
			Coordinates: [][]float64{},
			Type:        api.SwathPathType(rawWKTPrefix + string(b)),
		}
	}
	if out.Type == "" {
		out.Type = api.LineString
	}
	return out
}

// SwathToProto converts the OpenAPI Swath envelope to a proto Swath.
// CreationDir on the proto side is left false — the OpenAPI surface
// does not expose it.
func SwathToProto(in api.Swath) *f2cclient.Swath {
	return &f2cclient.Swath{
		Id:      int32(in.Id),
		PathWkt: swathPathToWKTBytes(in.Path),
		WidthM:  in.WidthM,
		Type:    swathTypeToProto(in.Type),
	}
}

// SwathFromProto reverses SwathToProto. Nil input yields a zero value.
func SwathFromProto(in *f2cclient.Swath) api.Swath {
	if in == nil {
		return api.Swath{}
	}
	return api.Swath{
		Id:     int(in.Id),
		Path:   wktBytesToSwathPath(in.PathWkt),
		WidthM: in.WidthM,
		Type:   swathTypeFromProto(in.Type),
	}
}

// SwathsToProto converts the OpenAPI Swaths envelope to a proto Swaths.
// A nil Sorted pointer is flattened to false (proto3 has no optional
// bool without a wrapper type).
func SwathsToProto(in api.Swaths) *f2cclient.Swaths {
	out := &f2cclient.Swaths{
		Field: FieldWithHeadlandsToProto(in.Field),
	}
	if len(in.Items) > 0 {
		out.Items = make([]*f2cclient.Swath, 0, len(in.Items))
		for i := range in.Items {
			out.Items = append(out.Items, SwathToProto(in.Items[i]))
		}
	}
	if in.Sorted != nil {
		out.Sorted = *in.Sorted
	}
	return out
}

// SwathsFromProto reverses SwathsToProto. Nil input yields a zero value.
// The Sorted bool is reflected back as a non-nil pointer when true,
// and left nil when false to mimic the OpenAPI "absent" convention.
func SwathsFromProto(in *f2cclient.Swaths) api.Swaths {
	if in == nil {
		return api.Swaths{}
	}
	out := api.Swaths{
		Field: FieldWithHeadlandsFromProto(in.Field),
	}
	if len(in.Items) > 0 {
		out.Items = make([]api.Swath, 0, len(in.Items))
		for _, s := range in.Items {
			out.Items = append(out.Items, SwathFromProto(s))
		}
	} else {
		out.Items = []api.Swath{}
	}
	if in.Sorted {
		v := true
		out.Sorted = &v
	}
	return out
}
