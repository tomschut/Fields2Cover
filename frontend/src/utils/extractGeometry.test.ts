import { describe, it, expect } from 'vitest'
import { extractGeometry } from './extractGeometry'

const VALID_POLYGON = {
  type: 'Polygon',
  coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
}

const VALID_MULTIPOLYGON = {
  type: 'MultiPolygon',
  coordinates: [[[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]]],
}

describe('extractGeometry', () => {
  it('returns Polygon geometry directly', () => {
    const result = extractGeometry(VALID_POLYGON)
    expect(result.type).toBe('Polygon')
  })

  it('returns MultiPolygon geometry directly', () => {
    const result = extractGeometry(VALID_MULTIPOLYGON)
    expect(result.type).toBe('MultiPolygon')
  })

  it('extracts geometry from a Feature', () => {
    const feature = { type: 'Feature', properties: {}, geometry: VALID_POLYGON }
    const result = extractGeometry(feature)
    expect(result.type).toBe('Polygon')
  })

  it('extracts geometry from the first feature in a FeatureCollection', () => {
    const fc = {
      type: 'FeatureCollection',
      features: [{ type: 'Feature', properties: {}, geometry: VALID_MULTIPOLYGON }],
    }
    const result = extractGeometry(fc)
    expect(result.type).toBe('MultiPolygon')
  })

  it('throws on null input', () => {
    expect(() => extractGeometry(null)).toThrow('Not a JSON object')
  })

  it('throws on non-object input', () => {
    expect(() => extractGeometry('hello')).toThrow('Not a JSON object')
  })

  it('throws on unsupported GeoJSON type', () => {
    expect(() => extractGeometry({ type: 'Point', coordinates: [5, 52] })).toThrow('Unsupported GeoJSON type')
  })

  it('throws on Feature with non-polygon geometry', () => {
    const feature = { type: 'Feature', properties: {}, geometry: { type: 'Point', coordinates: [5, 52] } }
    expect(() => extractGeometry(feature)).toThrow('Feature geometry is not Polygon or MultiPolygon')
  })

  it('throws on empty FeatureCollection', () => {
    expect(() => extractGeometry({ type: 'FeatureCollection', features: [] })).toThrow('Empty FeatureCollection')
  })
})
