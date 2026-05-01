import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { fetchParcels, isWithinNetherlands } from './pdok'
import type { BBox } from './pdok'

const NL_BBOX: BBox = { minLon: 5.0, minLat: 52.0, maxLon: 5.1, maxLat: 52.1 }

describe('isWithinNetherlands', () => {
  it('returns true for a bbox clearly inside the Netherlands', () => {
    expect(isWithinNetherlands(NL_BBOX)).toBe(true)
  })

  it('returns false for a bbox completely outside the Netherlands', () => {
    expect(isWithinNetherlands({ minLon: 10, minLat: 48, maxLon: 11, maxLat: 49 })).toBe(false)
  })

  it('returns true for a bbox overlapping the Netherlands', () => {
    expect(isWithinNetherlands({ minLon: 6, minLat: 51, maxLon: 8, maxLat: 54 })).toBe(true)
  })
})

describe('fetchParcels', () => {
  const mockFeatureCollection = {
    type: 'FeatureCollection',
    features: [
      {
        type: 'Feature',
        properties: { gewas: 'Grasland', jaar: 2025 },
        geometry: {
          type: 'Polygon',
          coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
        },
      },
    ],
  }

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('fetches parcels and returns FeatureCollection', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockFeatureCollection),
    } as Response)

    const result = await fetchParcels(NL_BBOX)
    expect(result.type).toBe('FeatureCollection')
    expect(result.features).toHaveLength(1)
  })

  it('builds URL with correct CRS and bbox parameters', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockFeatureCollection),
    } as Response)

    await fetchParcels(NL_BBOX)

    const calledUrl = mockFetch.mock.calls[0][0] as string
    expect(calledUrl).toContain('bbox=5%2C52%2C5.1%2C52.1')
    expect(calledUrl).toContain('crs=http')
    expect(calledUrl).toContain('CRS84')
    expect(calledUrl).toContain('brpgewas/items')
  })

  it('throws on non-ok response', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    } as Response)

    await expect(fetchParcels(NL_BBOX)).rejects.toThrow('PDOK request failed: 500')
  })

  it('passes AbortSignal to fetch', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockFeatureCollection),
    } as Response)

    const controller = new AbortController()
    await fetchParcels(NL_BBOX, controller.signal)

    const callArgs = mockFetch.mock.calls[0]
    expect((callArgs[1] as RequestInit)?.signal).toBe(controller.signal)
  })

  it('respects custom limit parameter', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockFeatureCollection),
    } as Response)

    await fetchParcels(NL_BBOX, undefined, 50)

    const calledUrl = mockFetch.mock.calls[0][0] as string
    expect(calledUrl).toContain('limit=50')
  })
})
