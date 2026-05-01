import type { FeatureCollection, Polygon } from 'geojson'

const PDOK_BASE = 'https://api.pdok.nl/rvo/gewaspercelen/ogc/v1/collections/brpgewas/items'
const CRS84 = 'http://www.opengis.net/def/crs/OGC/1.3/CRS84'

export interface BBox {
  minLon: number
  minLat: number
  maxLon: number
  maxLat: number
}

// Netherlands rough bounds check
export function isWithinNetherlands(bbox: BBox): boolean {
  return (
    bbox.maxLat > 50.5 && bbox.minLat < 53.7 &&
    bbox.maxLon > 3.3 && bbox.minLon < 7.3
  )
}

export async function fetchParcels(
  bbox: BBox,
  signal?: AbortSignal,
  limit = 200,
): Promise<FeatureCollection<Polygon>> {
  const params = new URLSearchParams({
    limit: String(limit),
    bbox: `${bbox.minLon},${bbox.minLat},${bbox.maxLon},${bbox.maxLat}`,
    'bbox-crs': CRS84,
    crs: CRS84,
    f: 'json',
  })
  const url = `${PDOK_BASE}?${params}`
  const res = await fetch(url, { signal })
  if (!res.ok) {
    throw new Error(`PDOK request failed: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<FeatureCollection<Polygon>>
}
