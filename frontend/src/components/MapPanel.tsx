import { useState } from 'react'
import { MapContainer, TileLayer, GeoJSON, Polyline, CircleMarker, useMapEvents } from 'react-leaflet'
import type { Geometry, Polygon } from 'geojson'
import { GeomanInit } from './GeomanInit'
import { PdokLayer } from './PdokLayer'
import type { PlanCoverageResponse } from '../api/pipeline'
import type { LayerVisibility } from './ResultLayerToggles'

const TILE_LAYERS = {
  osm: {
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 19,
  },
  satellite: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: 'Tiles &copy; Esri &mdash; Source: Esri, Maxar, GeoEye, Earthstar Geographics',
    maxZoom: 18,
  },
} as const

type TileKey = keyof typeof TILE_LAYERS

interface MapPanelProps {
  fieldGeometry: Geometry | null
  onGeometryChange: (g: Geometry) => void
  pipelineResult: PlanCoverageResponse | null
  layerVisibility: LayerVisibility
  startPoint: [number, number] | null
  onStartPointChange: (point: [number, number] | null) => void
}

function StartPointClickHandler({
  active,
  onPlace,
}: {
  active: boolean
  onPlace: (latlng: [number, number]) => void
}) {
  useMapEvents({
    click(e) {
      if (active) {
        onPlace([e.latlng.lat, e.latlng.lng])
      }
    },
  })
  return null
}

export function MapPanel({ fieldGeometry, onGeometryChange, pipelineResult, layerVisibility, startPoint, onStartPointChange }: MapPanelProps) {
  const [activeLayer, setActiveLayer] = useState<TileKey>('osm')
  const [startPointMode, setStartPointMode] = useState(false)

  return (
    <div className="map-panel">
      <MapContainer
        center={[51.997, 5.659]}
        zoom={15}
        style={{ width: '100%', height: '100%' }}
      >
        <TileLayer key={activeLayer} {...TILE_LAYERS[activeLayer]} />
        <PdokLayer onSelect={(geom: Polygon) => onGeometryChange(geom)} />
        <GeomanInit
          onGeometryCreated={onGeometryChange}
          onGeometryEdited={onGeometryChange}
        />
        {fieldGeometry && (
          <GeoJSON
            key={JSON.stringify(fieldGeometry)}
            data={fieldGeometry}
            style={{ color: '#166534', weight: 2, fillColor: '#166534', fillOpacity: 0.25 }}
          />
        )}

        {/* Headland ring — WGS84 GeoJSON geometry from API */}
        {pipelineResult && layerVisibility.headlands && (() => {
          const features = pipelineResult.intermediates.field_with_headlands?.headlands?.features ?? []
          return features.map((f, i) => {
            if (!f.geometry?.type) return null
            const fc: GeoJSON.FeatureCollection = {
              type: 'FeatureCollection',
              features: [{ type: 'Feature', geometry: f.geometry as GeoJSON.Geometry, properties: f.properties ?? {} }],
            }
            return (
              <GeoJSON key={`hl-${i}`} data={fc}
                style={{ color: '#0369a1', weight: 2, fillOpacity: 0.15 }} />
            )
          })
        })()}

        {/* Swath lines — WGS84 [lng, lat] coordinates from API */}
        {pipelineResult && layerVisibility.swaths &&
          (pipelineResult.intermediates.sorted_swaths?.items ?? []).map((swath, i) => {
            const coords = (swath.path?.coordinates ?? []) as [number, number][]
            // Leaflet expects [lat, lng]; API returns [lng, lat]
            const positions: [number, number][] = coords.map(([lng, lat]) => [lat, lng])
            if (positions.length < 2) return null
            return <Polyline key={`sw-${i}`} positions={positions} color="#b45309" weight={1.5} />
          })
        }

        {/* Route connections — WGS84 [lng, lat] points from API */}
        {pipelineResult && layerVisibility.route &&
          (pipelineResult.intermediates.route?.connections ?? []).map((conn, i) => {
            const positions: [number, number][] = (conn.points ?? []).map(([lng, lat]: number[]) => [lat, lng])
            if (positions.length < 2) return null
            return <Polyline key={`rt-${i}`} positions={positions} color="#7c3aed" weight={1} dashArray="4" />
          })
        }

        {/* Path — WGS84 [lng, lat] state points from API */}
        {pipelineResult && layerVisibility.path && (() => {
          const positions: [number, number][] = (pipelineResult.path?.states ?? []).map(
            s => [s.point[1], s.point[0]] as [number, number]
          )
          if (positions.length < 2) return null
          return <Polyline key="path" positions={positions} color="#be185d" weight={2} />
        })()}

        <StartPointClickHandler
          active={startPointMode}
          onPlace={(latlng) => {
            onStartPointChange(latlng)
            setStartPointMode(false)
          }}
        />
        {startPoint && (
          <CircleMarker
            center={startPoint}
            radius={8}
            pathOptions={{ color: '#15803d', fillColor: '#16a34a', fillOpacity: 0.9, weight: 2 }}
          />
        )}
      </MapContainer>
      <div className="tile-toggle">
        {(['osm', 'satellite'] as TileKey[]).map((key) => (
          <button
            key={key}
            className={`tile-toggle__btn${activeLayer === key ? ' tile-toggle__btn--active' : ''}`}
            onClick={() => setActiveLayer(key)}
          >
            {key === 'osm' ? 'Map' : 'Satellite'}
          </button>
        ))}
      </div>
      <div className="start-point-toggle">
        <button
          className={`tile-toggle__btn${startPointMode ? ' tile-toggle__btn--active' : ''}`}
          onClick={() => setStartPointMode((prev) => !prev)}
          title={startPointMode ? 'Cancel start point placement' : 'Set start point'}
        >
          {startPointMode ? 'Cancel' : 'Set start'}
        </button>
        {startPoint && (
          <button
            className="tile-toggle__btn"
            onClick={() => onStartPointChange(null)}
            title="Clear start point"
          >
            Clear
          </button>
        )}
      </div>
    </div>
  )
}
