import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, act } from '@testing-library/react'
import { MapContainer } from 'react-leaflet'
import { PdokLayer } from './PdokLayer'
import * as pdokApi from '../api/pdok'
import type { FeatureCollection, Polygon } from 'geojson'

// Mock the pdok api module
vi.mock('../api/pdok', () => ({
  fetchParcels: vi.fn(),
  isWithinNetherlands: vi.fn().mockReturnValue(true),
}))

const mockFeatureCollection: FeatureCollection<Polygon> = {
  type: 'FeatureCollection',
  features: [
    {
      type: 'Feature',
      properties: { id: 'parcel-1', gewas: 'Grasland', jaar: 2025 },
      geometry: {
        type: 'Polygon',
        coordinates: [[[5.0, 52.0], [5.1, 52.0], [5.1, 52.1], [5.0, 52.1], [5.0, 52.0]]],
      },
    },
  ],
}

function renderWithMap(onSelect = vi.fn()) {
  return render(
    <MapContainer center={[52.1, 5.3]} zoom={15} style={{ width: 800, height: 600 }}>
      <PdokLayer onSelect={onSelect} />
    </MapContainer>
  )
}

describe('PdokLayer', () => {
  beforeEach(() => {
    vi.mocked(pdokApi.fetchParcels).mockResolvedValue(mockFeatureCollection)
    vi.mocked(pdokApi.isWithinNetherlands).mockReturnValue(true)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing inside a MapContainer', async () => {
    await act(async () => {
      renderWithMap()
    })
    // Component mounts without error
    expect(pdokApi.fetchParcels).toHaveBeenCalled()
  })

  it('calls fetchParcels on mount when zoom >= 14', async () => {
    await act(async () => {
      renderWithMap()
    })
    expect(pdokApi.fetchParcels).toHaveBeenCalledTimes(1)
  })

  it('passes bbox to fetchParcels', async () => {
    await act(async () => {
      renderWithMap()
    })
    const call = vi.mocked(pdokApi.fetchParcels).mock.calls[0]
    const bbox = call[0]
    expect(bbox).toHaveProperty('minLon')
    expect(bbox).toHaveProperty('minLat')
    expect(bbox).toHaveProperty('maxLon')
    expect(bbox).toHaveProperty('maxLat')
  })

  it('does not fetch when isWithinNetherlands returns false', async () => {
    vi.mocked(pdokApi.isWithinNetherlands).mockReturnValue(false)
    await act(async () => {
      renderWithMap()
    })
    expect(pdokApi.fetchParcels).not.toHaveBeenCalled()
  })
})
