// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// CRSToProto converts the OpenAPI CRS type to its proto counterpart.
// A zero-value api.CRS (Epsg=0) is preserved as &CRS{Epsg: 0}.
func CRSToProto(in api.CRS) *f2cclient.CRS {
	out := &f2cclient.CRS{
		Epsg: int32(in.Epsg),
	}
	if in.UtmZone != nil {
		out.UtmZone = *in.UtmZone
	}
	return out
}

// CRSFromProto converts a proto CRS to the OpenAPI CRS type. Nil
// input yields a zero-valued api.CRS (Epsg: 0). The proto-only Raw
// field is intentionally dropped — the OpenAPI contract carries only
// epsg + optional utm_zone.
func CRSFromProto(in *f2cclient.CRS) api.CRS {
	if in == nil {
		return api.CRS{}
	}
	out := api.CRS{
		Epsg: int(in.Epsg),
	}
	if in.UtmZone != "" {
		zone := in.UtmZone
		out.UtmZone = &zone
	}
	return out
}
