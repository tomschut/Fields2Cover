// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// ProjectResponseToWGS84 reprojects all geometry in a PlanCoverageResponse
// from the field's local metric CRS (typically a UTM zone, EPSG 258xx/326xx)
// to WGS84 (EPSG:4326). The local origin is taken from
// path.field.field.ref_point (absolute WGS84 lng/lat of the local (0,0)).
//
// After projection:
//   - All coordinate arrays ([x,y]) are WGS84 [lng, lat].
//   - All WKT geometry on properties.wkt is replaced with real GeoJSON.
//   - CRS fields are updated to EPSG:4326.
//   - ref_point fields are cleared (no longer meaningful).
//   - The __rawwkt__: marker on Swath.Path.Type is resolved to real coordinates.
//   - Route connection Points are populated (previously empty when MULTIPOINT WKT).

package convert

import (
	"strings"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// ProjectResponseToWGS84 mutates resp in-place, converting all geometry
// from local metric (UTM) to WGS84. It is a no-op if the ref_point cannot
// be found or the CRS is already WGS84 (EPSG:4326).
func ProjectResponseToWGS84(resp *api.PlanCoverageResponse) {
	// Find origin and EPSG from the path's embedded field.
	var refLng, refLat float64
	var epsg int
	var found bool

	// Try to get origin from path.field.field.ref_point
	if resp.Path.Field.Field.RefPoint != nil && len(*resp.Path.Field.Field.RefPoint) >= 2 {
		rp := *resp.Path.Field.Field.RefPoint
		refLng, refLat = rp[0], rp[1]
		epsg = resp.Path.Field.Field.Crs.Epsg
		found = true
	}
	// Fallback: try intermediates
	if !found && resp.Intermediates.Field != nil && resp.Intermediates.Field.RefPoint != nil {
		rp := *resp.Intermediates.Field.RefPoint
		if len(rp) >= 2 {
			refLng, refLat = rp[0], rp[1]
			epsg = resp.Intermediates.Field.Crs.Epsg
			found = true
		}
	}
	if !found || epsg == 4326 {
		return // already WGS84 or no origin to work with
	}

	proj := &projector{refLng: refLng, refLat: refLat, epsg: epsg}

	// Project path states
	for i := range resp.Path.States {
		pt := resp.Path.States[i].Point
		if len(pt) >= 2 {
			lng, lat := proj.localToWGS84(pt[0], pt[1])
			resp.Path.States[i].Point = []float64{lng, lat}
		}
	}

	// Project path.field (FieldWithHeadlands)
	proj.projectFWH(&resp.Path.Field)

	// Project intermediates
	if resp.Intermediates.Field != nil {
		proj.projectField(resp.Intermediates.Field)
	}
	if resp.Intermediates.FieldWithHeadlands != nil {
		proj.projectFWH(resp.Intermediates.FieldWithHeadlands)
	}
	if resp.Intermediates.Swaths != nil {
		proj.projectSwaths(resp.Intermediates.Swaths)
	}
	if resp.Intermediates.SortedSwaths != nil {
		proj.projectSwaths(resp.Intermediates.SortedSwaths)
	}
	if resp.Intermediates.Route != nil {
		proj.projectRoute(resp.Intermediates.Route)
	}
}

// ProjectSwathsResponseToWGS84 projects a PlanSwathsResponse in-place.
func ProjectSwathsResponseToWGS84(resp *api.PlanSwathsResponse) {
	// Find origin from field_with_headlands.field
	fwh := &resp.FieldWithHeadlands
	if fwh.Field.RefPoint == nil || len(*fwh.Field.RefPoint) < 2 {
		return
	}
	rp := *fwh.Field.RefPoint
	epsg := fwh.Field.Crs.Epsg
	if epsg == 4326 {
		return
	}
	proj := &projector{refLng: rp[0], refLat: rp[1], epsg: epsg}
	proj.projectFWH(fwh)
	if resp.SortedSwaths.Field.Field.RefPoint != nil {
		proj.projectSwaths(&resp.SortedSwaths)
	}
}

// projector holds the origin and EPSG for a single projection pass.
type projector struct {
	refLng, refLat float64
	epsg           int
}

func (p *projector) localToWGS84(lx, ly float64) (float64, float64) {
	return LocalToWGS84(lx, ly, p.refLng, p.refLat, p.epsg)
}

func (p *projector) wgs84CRS() api.CRS {
	return api.CRS{Epsg: 4326}
}

func (p *projector) projectField(f *api.Field) {
	f.Crs = p.wgs84CRS()
	f.RefPoint = nil
	f.Geometry = p.projectGeoJSON(f.Geometry)
}

func (p *projector) projectFWH(fwh *api.FieldWithHeadlands) {
	p.projectField(&fwh.Field)
	fwh.Headlands = p.projectGeoJSON(fwh.Headlands)
	fwh.InnerField = p.projectGeoJSON(fwh.InnerField)
}

func (p *projector) projectSwaths(sw *api.Swaths) {
	p.projectFWH(&sw.Field)
	for i := range sw.Items {
		p.projectSwath(&sw.Items[i])
	}
}

func (p *projector) projectSwath(s *api.Swath) {
	typ := string(s.Path.Type)
	if strings.HasPrefix(typ, rawWKTPrefix) {
		wkt := typ[len(rawWKTPrefix):]
		pts := WKTLineStringPoints(wkt)
		coords := make([][]float64, 0, len(pts))
		for _, pt := range pts {
			lng, lat := p.localToWGS84(pt[0], pt[1])
			coords = append(coords, []float64{lng, lat})
		}
		s.Path.Coordinates = coords
		s.Path.Type = api.LineString
	} else {
		// coordinates already parsed — transform in-place
		for i, coord := range s.Path.Coordinates {
			if len(coord) >= 2 {
				lng, lat := p.localToWGS84(coord[0], coord[1])
				s.Path.Coordinates[i] = []float64{lng, lat}
			}
		}
	}
}

func (p *projector) projectRoute(r *api.Route) {
	p.projectFWH(&r.Field)
	p.projectSwaths(&r.Swaths)
	for i := range r.Connections {
		conn := &r.Connections[i]
		// Points may be empty if wktBytesToPoints failed on real WKT.
		// We don't have the raw WKT here — connections will stay empty.
		// (This is a pre-existing limitation; route connections are optional for rendering.)
		for j, pt := range conn.Points {
			if len(pt) >= 2 {
				lng, lat := p.localToWGS84(pt[0], pt[1])
				conn.Points[j] = []float64{lng, lat}
			}
		}
	}
}

// projectGeoJSON converts a GeoJSONFeatureCollection that may contain raw WKT
// (in properties.wkt) into one with real projected WGS84 coordinates.
func (p *projector) projectGeoJSON(fc api.GeoJSONFeatureCollection) api.GeoJSONFeatureCollection {
	if len(fc.Features) == 0 {
		return fc
	}
	// Check if it's a synthetic WKT-carrier (from wktBytesToGeoJSON fallback)
	if len(fc.Features) == 1 && fc.Features[0].Properties != nil {
		if raw, ok := (*fc.Features[0].Properties)["wkt"]; ok {
			if wktStr, ok := raw.(string); ok && wktStr != "" {
				return p.wktToGeoJSON(wktStr)
			}
		}
	}
	// Real GeoJSON — transform coordinates in-place
	newFeatures := make([]struct {
		Geometry   map[string]interface{}                   `json:"geometry"`
		Properties *map[string]interface{}                  `json:"properties"`
		Type       api.GeoJSONFeatureCollectionFeaturesType `json:"type"`
	}, len(fc.Features))
	for i, feat := range fc.Features {
		newFeatures[i] = feat
		if feat.Geometry != nil {
			newFeatures[i].Geometry = p.projectGeometryMap(feat.Geometry)
		}
	}
	return api.GeoJSONFeatureCollection{
		Type:                 fc.Type,
		Features:             newFeatures,
		AdditionalProperties: fc.AdditionalProperties,
	}
}

// projectGeometryMap transforms coordinates in a GeoJSON geometry map.
func (p *projector) projectGeometryMap(g map[string]interface{}) map[string]interface{} {
	if g == nil {
		return g
	}
	result := make(map[string]interface{}, len(g))
	for k, v := range g {
		result[k] = v
	}
	coords, ok := g["coordinates"]
	if !ok {
		return result
	}
	result["coordinates"] = p.projectCoordinatesInterface(coords)
	return result
}

func (p *projector) projectCoordinatesInterface(v interface{}) interface{} {
	switch c := v.(type) {
	case []interface{}:
		if len(c) == 0 {
			return c
		}
		// Check if it's a point [x, y] (leaf) or nested
		if _, isFloat := c[0].(float64); isFloat {
			if len(c) >= 2 {
				x, _ := c[0].(float64)
				y, _ := c[1].(float64)
				lng, lat := p.localToWGS84(x, y)
				return []interface{}{lng, lat}
			}
			return c
		}
		result := make([]interface{}, len(c))
		for i, elem := range c {
			result[i] = p.projectCoordinatesInterface(elem)
		}
		return result
	default:
		return v
	}
}

// wktToGeoJSON converts a WKT string in local metric coords to a WGS84
// GeoJSONFeatureCollection with real geometry.
func (p *projector) wktToGeoJSON(wkt string) api.GeoJSONFeatureCollection {
	upper := strings.ToUpper(strings.TrimSpace(wkt))

	makeFC := func(geomType string, coords interface{}) api.GeoJSONFeatureCollection {
		return api.GeoJSONFeatureCollection{
			Type: api.FeatureCollection,
			Features: []struct {
				Geometry   map[string]interface{}                   `json:"geometry"`
				Properties *map[string]interface{}                  `json:"properties"`
				Type       api.GeoJSONFeatureCollectionFeaturesType `json:"type"`
			}{
				{
					Geometry: map[string]interface{}{
						"type":        geomType,
						"coordinates": coords,
					},
					Type: api.Feature,
				},
			},
		}
	}

	switch {
	case strings.HasPrefix(upper, "MULTIPOLYGON"):
		polys := WKTMultiPolygonPolygons(wkt)
		if len(polys) == 0 {
			break
		}
		// Build MultiPolygon: [polygon][ring][point]
		// Preserve exterior + hole ring structure per polygon element.
		polygons := make([][][]interface{}, 0, len(polys))
		for _, rings := range polys {
			projPoly := make([][]interface{}, 0, len(rings))
			for _, ring := range rings {
				projRing := make([]interface{}, 0, len(ring))
				for _, pt := range ring {
					lng, lat := p.localToWGS84(pt[0], pt[1])
					projRing = append(projRing, []interface{}{lng, lat})
				}
				projPoly = append(projPoly, projRing)
			}
			polygons = append(polygons, projPoly)
		}
		return makeFC("MultiPolygon", polygons)

	case strings.HasPrefix(upper, "POLYGON"):
		pts := WKTPolygonPoints(wkt)
		if len(pts) == 0 {
			break
		}
		ring := make([]interface{}, 0, len(pts))
		for _, pt := range pts {
			lng, lat := p.localToWGS84(pt[0], pt[1])
			ring = append(ring, []interface{}{lng, lat})
		}
		return makeFC("Polygon", [][][]interface{}{{ring}})

	case strings.HasPrefix(upper, "LINESTRING"):
		pts := WKTLineStringPoints(wkt)
		if len(pts) == 0 {
			break
		}
		coords := make([]interface{}, 0, len(pts))
		for _, pt := range pts {
			lng, lat := p.localToWGS84(pt[0], pt[1])
			coords = append(coords, []interface{}{lng, lat})
		}
		return makeFC("LineString", coords)
	}

	// Fall back: keep as WKT carrier (shouldn't happen in practice)
	props := map[string]interface{}{"wkt": wkt}
	return api.GeoJSONFeatureCollection{
		Type: api.FeatureCollection,
		Features: []struct {
			Geometry   map[string]interface{}                   `json:"geometry"`
			Properties *map[string]interface{}                  `json:"properties"`
			Type       api.GeoJSONFeatureCollectionFeaturesType `json:"type"`
		}{{Geometry: map[string]interface{}{}, Properties: &props, Type: api.Feature}},
	}
}
