import type { Geometry } from 'geojson'

export function extractGeometry(parsed: unknown): Geometry {
  if (typeof parsed !== 'object' || parsed === null) {
    throw new Error('Not a JSON object')
  }
  const obj = parsed as Record<string, unknown>
  const type = obj['type']

  if (type === 'Polygon' || type === 'MultiPolygon') {
    return obj as unknown as Geometry
  }
  if (type === 'Feature') {
    const geom = obj['geometry']
    if (geom && typeof geom === 'object') {
      const g = geom as Record<string, unknown>
      if (g['type'] === 'Polygon' || g['type'] === 'MultiPolygon') {
        return g as unknown as Geometry
      }
    }
    throw new Error('Feature geometry is not Polygon or MultiPolygon')
  }
  if (type === 'FeatureCollection') {
    const features = obj['features']
    if (Array.isArray(features) && features.length > 0) {
      return extractGeometry(features[0])
    }
    throw new Error('Empty FeatureCollection')
  }
  throw new Error(`Unsupported GeoJSON type: ${type}`)
}
