import type { Geometry } from 'geojson'
import type { PlanCoverageRequest, PlanCoverageResponse } from '../api/pipeline'
import type { LayerVisibility } from './ResultLayerToggles'
import { InputMethodTabs } from './InputMethodTabs'
import { FieldStatus } from './FieldStatus'
import { PipelineParamsForm } from './PipelineParamsForm'
import { PipelineErrorBanner } from './PipelineErrorBanner'
import { ResultLayerToggles } from './ResultLayerToggles'
import { ResultStatsRow } from './ResultStatsRow'

interface SidebarProps {
  fieldGeometry: Geometry | null
  onGeometry: (g: Geometry) => void
  pipelineResult: PlanCoverageResponse | null
  pipelineError: string | null
  pipelineLoading: boolean
  layerVisibility: LayerVisibility
  onRun: (req: PlanCoverageRequest) => void
  onLayerVisibilityChange: (v: LayerVisibility) => void
}

export function Sidebar({
  fieldGeometry,
  onGeometry,
  pipelineResult,
  pipelineError,
  pipelineLoading,
  layerVisibility,
  onRun,
  onLayerVisibilityChange,
}: SidebarProps) {
  return (
    <aside className="sidebar">
      <InputMethodTabs onGeometry={onGeometry} />
      <PipelineParamsForm
        fieldGeometry={fieldGeometry}
        onRun={onRun}
        loading={pipelineLoading}
      />
      {pipelineError && <PipelineErrorBanner error={pipelineError} />}
      {pipelineResult && (
        <>
          <ResultLayerToggles
            layerVisibility={layerVisibility}
            onVisibilityChange={onLayerVisibilityChange}
          />
          <ResultStatsRow result={pipelineResult} />
        </>
      )}
      <FieldStatus fieldGeometry={fieldGeometry} />
    </aside>
  )
}
