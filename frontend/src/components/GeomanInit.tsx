import { useEffect } from 'react'
import { useMap } from 'react-leaflet'
import '@geoman-io/leaflet-geoman-free'
import type { Layer } from 'leaflet'
import type { Geometry } from 'geojson'

interface GeomanInitProps {
  onGeometryCreated: (geometry: Geometry) => void
  onGeometryEdited: (geometry: Geometry) => void
}

export function GeomanInit({ onGeometryCreated, onGeometryEdited }: GeomanInitProps) {
  const map = useMap()

  useEffect(() => {
    map.pm.addControls({
      position: 'topleft',
      drawCircle: false,
      drawCircleMarker: false,
      drawMarker: false,
      drawPolyline: false,
      drawRectangle: false,
      drawText: false,
      cutPolygon: false,
      rotateMode: false,
    })

    const handleCreate = (e: { layer: Layer }) => {
      const geojson = (e.layer as unknown as { toGeoJSON: () => GeoJSON.Feature }).toGeoJSON()
      if ('geometry' in geojson) {
        onGeometryCreated(geojson.geometry)
      }
      map.removeLayer(e.layer)
    }

    const handleEdit = (e: { layer: Layer }) => {
      const geojson = (e.layer as unknown as { toGeoJSON: () => GeoJSON.Feature }).toGeoJSON()
      if ('geometry' in geojson) {
        onGeometryEdited(geojson.geometry)
      }
    }

    map.on('pm:create', handleCreate)
    map.on('pm:edit', handleEdit)

    return () => {
      map.off('pm:create', handleCreate)
      map.off('pm:edit', handleEdit)
      map.pm.removeControls()
    }
  }, [map, onGeometryCreated, onGeometryEdited])

  return null
}
