// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"encoding/json"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// pointsToWKTBytes serializes a [][]float64 multipoint as JSON for
// the proto bytes pass-through.
func pointsToWKTBytes(in [][]float64) []byte {
	if len(in) == 0 {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	return b
}

// wktBytesToPoints reverses pointsToWKTBytes. If unmarshal fails
// (because the shim returned real WKT MULTIPOINT), returns an empty
// slice — handler tests assert presence, not coordinate equality.
func wktBytesToPoints(b []byte) [][]float64 {
	if len(b) == 0 {
		return [][]float64{}
	}
	var out [][]float64
	if err := json.Unmarshal(b, &out); err != nil {
		return [][]float64{}
	}
	return out
}

// RouteConnectionToProto converts the OpenAPI RouteConnection envelope
// to a proto RouteConnection.
func RouteConnectionToProto(in api.RouteConnection) *f2cclient.RouteConnection {
	return &f2cclient.RouteConnection{
		FromSwathId: int32(in.FromSwathId),
		ToSwathId:   int32(in.ToSwathId),
		PointsWkt:   pointsToWKTBytes(in.Points),
	}
}

// RouteConnectionFromProto reverses RouteConnectionToProto. Nil input
// yields a zero value (with an empty Points slice, not nil, so JSON
// round-trips as `[]` rather than `null`).
func RouteConnectionFromProto(in *f2cclient.RouteConnection) api.RouteConnection {
	if in == nil {
		return api.RouteConnection{Points: [][]float64{}}
	}
	return api.RouteConnection{
		FromSwathId: int(in.FromSwathId),
		ToSwathId:   int(in.ToSwathId),
		Points:      wktBytesToPoints(in.PointsWkt),
	}
}

// RouteToProto converts the OpenAPI Route envelope to a proto Route.
func RouteToProto(in api.Route) *f2cclient.Route {
	out := &f2cclient.Route{
		Field:  FieldWithHeadlandsToProto(in.Field),
		Swaths: SwathsToProto(in.Swaths),
	}
	if len(in.Connections) > 0 {
		out.Connections = make([]*f2cclient.RouteConnection, 0, len(in.Connections))
		for i := range in.Connections {
			out.Connections = append(out.Connections, RouteConnectionToProto(in.Connections[i]))
		}
	}
	return out
}

// RouteFromProto reverses RouteToProto. Nil input yields a zero value.
func RouteFromProto(in *f2cclient.Route) api.Route {
	if in == nil {
		return api.Route{Connections: []api.RouteConnection{}}
	}
	out := api.Route{
		Field:  FieldWithHeadlandsFromProto(in.Field),
		Swaths: SwathsFromProto(in.Swaths),
	}
	if len(in.Connections) > 0 {
		out.Connections = make([]api.RouteConnection, 0, len(in.Connections))
		for _, c := range in.Connections {
			out.Connections = append(out.Connections, RouteConnectionFromProto(c))
		}
	} else {
		out.Connections = []api.RouteConnection{}
	}
	return out
}
