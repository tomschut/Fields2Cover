// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"strconv"
	"strings"
)

// parseWKTCoordPair parses "x y" or "x y z" into [x, y].
func parseWKTCoordPair(s string) (float64, float64, bool) {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) < 2 {
		return 0, 0, false
	}
	x, err1 := strconv.ParseFloat(parts[0], 64)
	y, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return x, y, true
}

// parseWKTPointList parses a comma-separated list of "x y" pairs into [][2]float64.
func parseWKTPointList(s string) [][2]float64 {
	var out [][2]float64
	for _, tok := range strings.Split(s, ",") {
		if x, y, ok := parseWKTCoordPair(tok); ok {
			out = append(out, [2]float64{x, y})
		}
	}
	return out
}

// WKTLineStringPoints parses a LINESTRING WKT and returns its points.
// Handles "LINESTRING (x y, x y, ...)" and "LINESTRING Z (x y z, ...)".
// Returns nil on parse failure.
func WKTLineStringPoints(wkt string) [][2]float64 {
	s := strings.TrimSpace(wkt)
	// Strip type keyword and optional Z/M/ZM modifier
	upper := strings.ToUpper(s)
	for _, prefix := range []string{"LINESTRING ZM", "LINESTRING Z", "LINESTRING M", "LINESTRING"} {
		if strings.HasPrefix(upper, prefix) {
			s = strings.TrimSpace(s[len(prefix):])
			break
		}
	}
	s = strings.Trim(s, "()")
	return parseWKTPointList(s)
}

// WKTMultiPointPoints parses a MULTIPOINT WKT.
// Handles both "MULTIPOINT ((x y), (x y))" and "MULTIPOINT (x y, x y)".
func WKTMultiPointPoints(wkt string) [][2]float64 {
	s := strings.TrimSpace(wkt)
	upper := strings.ToUpper(s)
	for _, prefix := range []string{"MULTIPOINT ZM", "MULTIPOINT Z", "MULTIPOINT M", "MULTIPOINT"} {
		if strings.HasPrefix(upper, prefix) {
			s = strings.TrimSpace(s[len(prefix):])
			break
		}
	}
	// Remove outer parens
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}
	// Remove per-point parens if present: "(x y)" → "x y"
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	return parseWKTPointList(s)
}

// WKTPolygonRings parses a WKT ring body (the part inside POLYGON or
// one polygon element of MULTIPOLYGON): "((x y,...),(hole...))" and
// returns each ring as a slice of points. Only exterior + first hole are
// meaningful for GeoJSON rendering.
func WKTPolygonRings(body string) [][][2]float64 {
	// body looks like: ((x y,x y,...),(x y,...))
	body = strings.TrimSpace(body)
	body = strings.TrimPrefix(body, "(")
	body = strings.TrimSuffix(body, ")")

	var rings [][][2]float64
	depth := 0
	start := 0
	for i, c := range body {
		switch c {
		case '(':
			if depth == 0 {
				start = i + 1
			}
			depth++
		case ')':
			depth--
			if depth == 0 {
				ring := parseWKTPointList(body[start:i])
				if len(ring) > 0 {
					rings = append(rings, ring)
				}
			}
		}
	}
	return rings
}

// WKTMultiPolygonPolygons parses a MULTIPOLYGON WKT and returns each polygon
// as a slice of rings (first ring = exterior, remaining = holes).
// Returns nil on parse failure.
func WKTMultiPolygonPolygons(wkt string) [][][][2]float64 {
	s := strings.TrimSpace(wkt)
	upper := strings.ToUpper(s)
	for _, prefix := range []string{"MULTIPOLYGON ZM", "MULTIPOLYGON Z", "MULTIPOLYGON M", "MULTIPOLYGON"} {
		if strings.HasPrefix(upper, prefix) {
			s = strings.TrimSpace(s[len(prefix):])
			break
		}
	}
	// Strip outer parens of MULTIPOLYGON
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}

	// Split into individual polygon elements "((ring,...),(ring,...))"
	var polygons [][][][2]float64
	depth := 0
	start := 0
	for i, c := range s {
		switch c {
		case '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ')':
			depth--
			if depth == 0 {
				elem := strings.TrimSpace(s[start : i+1])
				rings := WKTPolygonRings(elem)
				if len(rings) > 0 {
					polygons = append(polygons, rings)
				}
			}
		}
	}
	return polygons
}

// WKTMultiPolygonRings parses a MULTIPOLYGON WKT and returns all rings
// from all polygon members flattened into one slice. Returns nil on parse failure.
//
// Deprecated: use WKTMultiPolygonPolygons to preserve exterior/hole ring
// grouping per polygon element.
func WKTMultiPolygonRings(wkt string) [][][2]float64 {
	polygons := WKTMultiPolygonPolygons(wkt)
	var allRings [][][2]float64
	for _, rings := range polygons {
		allRings = append(allRings, rings...)
	}
	return allRings
}

// WKTPolygonPoints parses a POLYGON WKT (single polygon) and returns the
// exterior ring's points.
func WKTPolygonPoints(wkt string) [][2]float64 {
	s := strings.TrimSpace(wkt)
	upper := strings.ToUpper(s)
	for _, prefix := range []string{"POLYGON ZM", "POLYGON Z", "POLYGON M", "POLYGON"} {
		if strings.HasPrefix(upper, prefix) {
			s = strings.TrimSpace(s[len(prefix):])
			break
		}
	}
	rings := WKTPolygonRings(s)
	if len(rings) == 0 {
		return nil
	}
	return rings[0] // exterior ring
}
