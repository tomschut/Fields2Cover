// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import "math"

// WGS84 ellipsoid (used by all UTM projections here).
const (
	utmA  = 6378137.0
	utmF  = 1.0 / 298.257223563
	utmK0 = 0.9996
	utmFE = 500000.0
)

var (
	utmE2  = 2*utmF - utmF*utmF  // first eccentricity squared
	utmEP2 = utmE2 / (1 - utmE2) // second eccentricity squared
	utmB   = utmA * (1 - utmF)   // semi-minor axis
)

// utmCentralMeridian returns the central meridian in radians for a given
// UTM zone number (1-60).
func utmCentralMeridian(zone int) float64 {
	return (float64(6*zone-183) * math.Pi / 180)
}

// meridionalArc computes the meridional arc M from the equator to latitude
// phi (radians).
func meridionalArc(phi float64) float64 {
	e2 := utmE2
	e4 := e2 * e2
	e6 := e4 * e2
	return utmA * ((1-e2/4-3*e4/64-5*e6/256)*phi -
		(3*e2/8+3*e4/32+45*e6/1024)*math.Sin(2*phi) +
		(15*e4/256+45*e6/1024)*math.Sin(4*phi) -
		(35*e6/3072)*math.Sin(6*phi))
}

// UTMFromWGS84 converts WGS84 lng/lat (degrees) to UTM easting/northing (metres)
// for the given zone number and hemisphere (north=true).
func UTMFromWGS84(lng, lat float64, zone int, north bool) (easting, northing float64) {
	phi := lat * math.Pi / 180
	lambda := lng * math.Pi / 180
	lambda0 := utmCentralMeridian(zone)

	N := utmA / math.Sqrt(1-utmE2*math.Sin(phi)*math.Sin(phi))
	T := math.Tan(phi) * math.Tan(phi)
	C := utmEP2 * math.Cos(phi) * math.Cos(phi)
	A := (lambda - lambda0) * math.Cos(phi)
	M := meridionalArc(phi)

	easting = utmK0*N*(A+(1-T+C)*A*A*A/6+
		(5-18*T+T*T+72*C-58*utmEP2)*A*A*A*A*A/120) + utmFE

	northing = utmK0 * (M + N*math.Tan(phi)*(A*A/2+
		(5-T+9*C+4*C*C)*A*A*A*A/24+
		(61-58*T+T*T+600*C-330*utmEP2)*A*A*A*A*A*A/720))
	if !north {
		northing += 10000000.0
	}
	return
}

// WGS84FromUTM converts UTM easting/northing (metres) to WGS84 lng/lat
// (degrees) for the given zone number and hemisphere (north=true).
func WGS84FromUTM(easting, northing float64, zone int, north bool) (lng, lat float64) {
	fn := northing
	if !north {
		fn -= 10000000.0
	}
	x := easting - utmFE
	lambda0 := utmCentralMeridian(zone)

	M := fn / utmK0
	mu := M / (utmA * (1 - utmE2/4 - 3*utmE2*utmE2/64 - 5*utmE2*utmE2*utmE2/256))

	e1 := (1 - math.Sqrt(1-utmE2)) / (1 + math.Sqrt(1-utmE2))
	phi1 := mu +
		(3*e1/2-27*e1*e1*e1/32)*math.Sin(2*mu) +
		(21*e1*e1/16-55*e1*e1*e1*e1/32)*math.Sin(4*mu) +
		(151*e1*e1*e1/96)*math.Sin(6*mu) +
		(1097*e1*e1*e1*e1/512)*math.Sin(8*mu)

	N1 := utmA / math.Sqrt(1-utmE2*math.Sin(phi1)*math.Sin(phi1))
	T1 := math.Tan(phi1) * math.Tan(phi1)
	C1 := utmEP2 * math.Cos(phi1) * math.Cos(phi1)
	R1 := utmA * (1 - utmE2) / math.Pow(1-utmE2*math.Sin(phi1)*math.Sin(phi1), 1.5)
	D := x / (N1 * utmK0)

	phi := phi1 - (N1*math.Tan(phi1)/R1)*(D*D/2-
		(5+3*T1+10*C1-4*C1*C1-9*utmEP2)*D*D*D*D/24+
		(61+90*T1+298*C1+45*T1*T1-252*utmEP2-3*C1*C1)*D*D*D*D*D*D/720)
	lambda := lambda0 + (D-(1+2*T1+C1)*D*D*D/6+
		(5-2*C1+(28*T1)-3*C1*C1+8*utmEP2+24*T1*T1)*D*D*D*D*D/120)/math.Cos(phi1)

	return lambda * 180 / math.Pi, phi * 180 / math.Pi
}

// epsgToUTMZone decodes an EPSG code into a UTM zone number and north flag.
// Supports EPSG 258xx (ETRS89/UTM north), 326xx (WGS84/UTM north),
// 327xx (WGS84/UTM south). Returns (0, false, false) if unrecognised.
func epsgToUTMZone(epsg int) (zone int, north bool, ok bool) {
	switch {
	case epsg >= 25828 && epsg <= 25838:
		return epsg - 25800, true, true
	case epsg >= 32601 && epsg <= 32660:
		return epsg - 32600, true, true
	case epsg >= 32701 && epsg <= 32760:
		return epsg - 32700, false, true
	default:
		return 0, false, false
	}
}

// LocalToWGS84 converts a local-metric point (lx, ly) in a UTM CRS
// (identified by epsg) back to WGS84 lng/lat. refE/refN are the UTM
// easting/northing of the local origin (the ref_point as stored by the
// shim — absolute UTM metres, not WGS84 degrees). Returns the input
// unchanged if epsg is unrecognised (safe fall-through).
func LocalToWGS84(lx, ly float64, refE, refN float64, epsg int) (outLng, outLat float64) {
	zone, north, ok := epsgToUTMZone(epsg)
	if !ok {
		return lx, ly // unknown CRS — pass through
	}
	return WGS84FromUTM(refE+lx, refN+ly, zone, north)
}

// WGS84ToLocal converts a WGS84 lng/lat point to local-metric coordinates
// relative to the field origin. refE/refN are the UTM easting/northing of
// the local origin (as stored in field.ref_point). Returns (0, 0, false) if
// the EPSG code is unrecognised.
func WGS84ToLocal(lng, lat float64, refE, refN float64, epsg int) (lx, ly float64, ok bool) {
	zone, north, ok := epsgToUTMZone(epsg)
	if !ok {
		return 0, 0, false
	}
	e, n := UTMFromWGS84(lng, lat, zone, north)
	return e - refE, n - refN, true
}

// Ensure utmB is referenced to avoid "declared and not used" compile error.
var _ = utmB
