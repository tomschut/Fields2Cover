import { useState, useCallback } from 'react'
import type { Geometry } from 'geojson'
import { Sidebar } from '../components/Sidebar'
import { MapPanel } from '../components/MapPanel'
import { planCoverage } from '../api/pipeline'
import type { PlanCoverageRequest, PlanCoverageResponse } from '../api/pipeline'
import type { LayerVisibility } from '../components/ResultLayerToggles'

export default function MapPage() {
  const [fieldGeometry, setFieldGeometry] = useState<Geometry | null>(null)
  const [pipelineResult, setPipelineResult] = useState<PlanCoverageResponse | null>(null)
  const [pipelineError, setPipelineError] = useState<string | null>(null)
  const [pipelineLoading, setPipelineLoading] = useState(false)
  const [layerVisibility, setLayerVisibility] = useState<LayerVisibility>({
    headlands: true, swaths: true, route: true, path: true,
  })
  const [startPoint, setStartPoint] = useState<[number, number] | null>(null)

  const handleGeometry = useCallback((g: Geometry) => {
    setFieldGeometry(g)
  }, [])

  const handleRunPipeline = useCallback(async (req: PlanCoverageRequest) => {
    setPipelineLoading(true)
    setPipelineError(null)
    try {
      // Convert Leaflet [lat, lng] to GeoJSON [lng, lat] for the API
      const sortStartPoint = startPoint !== null
        ? [startPoint[1], startPoint[0]] as [number, number]
        : undefined
      const result = await planCoverage({ ...req, sort_start_point: sortStartPoint })
      setPipelineResult(result)
      setLayerVisibility({ headlands: true, swaths: true, route: true, path: true })
    } catch (err) {
      setPipelineError(err instanceof Error ? err.message : 'Unknown error')
      // CRITICAL: do NOT call setPipelineResult(null) here — previous overlays must stay visible (PIPE-05)
    } finally {
      setPipelineLoading(false)
    }
  }, [startPoint])

  return (
    <div className="map-shell">
      <Sidebar
        fieldGeometry={fieldGeometry}
        onGeometry={handleGeometry}
        pipelineResult={pipelineResult}
        pipelineError={pipelineError}
        pipelineLoading={pipelineLoading}
        layerVisibility={layerVisibility}
        onRun={handleRunPipeline}
        onLayerVisibilityChange={setLayerVisibility}
      />
      <MapPanel
        fieldGeometry={fieldGeometry}
        onGeometryChange={handleGeometry}
        pipelineResult={pipelineResult}
        layerVisibility={layerVisibility}
        startPoint={startPoint}
        onStartPointChange={setStartPoint}
      />
    </div>
  )
}
