export interface LayerVisibility {
  headlands: boolean
  swaths: boolean
  route: boolean
  path: boolean
}

interface ResultLayerTogglesProps {
  layerVisibility: LayerVisibility
  onVisibilityChange: (v: LayerVisibility) => void
}

const LAYER_CONFIG: { key: keyof LayerVisibility; label: string; color: string }[] = [
  { key: 'headlands', label: 'Headlands', color: '#0369a1' },
  { key: 'swaths', label: 'Swaths', color: '#b45309' },
  { key: 'route', label: 'Route', color: '#7c3aed' },
  { key: 'path', label: 'Path', color: '#be185d' },
]

export function ResultLayerToggles({ layerVisibility, onVisibilityChange }: ResultLayerTogglesProps) {
  return (
    <div>
      <div className="section-heading">Result layers</div>
      <div className="layer-toggles">
        {LAYER_CONFIG.map(({ key, label, color }) => (
          <label key={key} className="layer-toggle" htmlFor={`layer-toggle-${key}`}>
            <span className="layer-toggle__swatch" style={{ backgroundColor: color }} />
            <input
              id={`layer-toggle-${key}`}
              type="checkbox"
              checked={layerVisibility[key]}
              onChange={() => onVisibilityChange({ ...layerVisibility, [key]: !layerVisibility[key] })}
            />
            {label}
          </label>
        ))}
      </div>
    </div>
  )
}
