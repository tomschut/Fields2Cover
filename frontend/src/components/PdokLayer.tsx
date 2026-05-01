import { useEffect, useRef, useState } from 'react'
import { GeoJSON, useMapEvents } from 'react-leaflet'
import type { Layer, LeafletMouseEvent } from 'leaflet'
import type { FeatureCollection, Polygon, Feature } from 'geojson'
import { fetchParcels, isWithinNetherlands } from '../api/pdok'

const MIN_ZOOM = 14

const PARCEL_STYLE = {
  color: '#166534',
  weight: 1.5,
  fillColor: '#166534',
  fillOpacity: 0.15,
}

const SELECTED_STYLE = {
  color: '#166534',
  weight: 2.5,
  fillColor: '#166534',
  fillOpacity: 0.35,
}

interface PdokLayerProps {
  onSelect: (geometry: Polygon) => void
}

export function PdokLayer({ onSelect }: PdokLayerProps) {
  const [parcels, setParcels] = useState<FeatureCollection<Polygon> | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const abortRef = useRef<AbortController | null>(null)

  const map = useMapEvents({
    moveend: () => loadParcels(),
    zoomend: () => loadParcels(),
  })

  function loadParcels() {
    const zoom = map.getZoom()
    if (zoom < MIN_ZOOM) {
      setParcels(null)
      return
    }

    const bounds = map.getBounds()
    const bbox = {
      minLon: bounds.getWest(),
      minLat: bounds.getSouth(),
      maxLon: bounds.getEast(),
      maxLat: bounds.getNorth(),
    }

    if (!isWithinNetherlands(bbox)) {
      setParcels(null)
      return
    }

    // Cancel previous request
    if (abortRef.current) {
      abortRef.current.abort()
    }
    const controller = new AbortController()
    abortRef.current = controller

    fetchParcels(bbox, controller.signal)
      .then(data => setParcels(data))
      .catch(err => {
        if (err.name !== 'AbortError') {
          console.error('PDOK fetch error:', err)
        }
      })
  }

  useEffect(() => {
    // Initial load on mount
    loadParcels()
    return () => {
      abortRef.current?.abort()
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (!parcels) return null

  return (
    <GeoJSON
      key={JSON.stringify(parcels.features.length + '-' + selectedId)}
      data={parcels}
      style={(feature) => {
        const id = (feature as Feature)?.properties?.id ?? JSON.stringify((feature as Feature)?.geometry)
        return id === selectedId ? SELECTED_STYLE : PARCEL_STYLE
      }}
      onEachFeature={(feature: Feature<Polygon>, layer: Layer) => {
        layer.on('click', (e: LeafletMouseEvent) => {
          // Stop propagation so Geoman doesn't intercept
          e.originalEvent?.stopPropagation()
          const id = feature.properties?.id ?? JSON.stringify(feature.geometry)
          setSelectedId(id)
          onSelect(feature.geometry)
        })
      }}
    />
  )
}
