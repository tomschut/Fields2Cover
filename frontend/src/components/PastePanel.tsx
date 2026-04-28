import { useState } from 'react'
import type { Geometry } from 'geojson'
import { extractGeometry } from '../utils/extractGeometry'

interface PastePanelProps {
  onGeometry: (g: Geometry) => void
}

export function PastePanel({ onGeometry }: PastePanelProps) {
  const [value, setValue] = useState('')
  const [error, setError] = useState<string | null>(null)

  const handleLoad = () => {
    if (!value.trim()) {
      setError('Paste GeoJSON before clicking Load.')
      return
    }
    try {
      const parsed: unknown = JSON.parse(value)
      onGeometry(extractGeometry(parsed))
      setError(null)
    } catch {
      setError(
        'Invalid GeoJSON. Check that the input is a valid Feature, FeatureCollection, Polygon, or MultiPolygon.'
      )
    }
  }

  return (
    <div>
      <textarea
        className={`paste-textarea${error ? ' paste-textarea--error' : ''}`}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder='{"type": "Polygon", "coordinates": [[[...]]]}'
      />
      {error && <p className="error-text">{error}</p>}
      <button className="load-field-btn" onClick={handleLoad}>
        Load field
      </button>
    </div>
  )
}
