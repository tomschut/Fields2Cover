import { useState, useRef } from 'react'
import type { Geometry } from 'geojson'
import { extractGeometry } from '../utils/extractGeometry'

interface UploadPanelProps {
  onGeometry: (g: Geometry) => void
}

export function UploadPanel({ onGeometry }: UploadPanelProps) {
  const [dragOver, setDragOver] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const readFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const parsed: unknown = JSON.parse(e.target?.result as string)
        onGeometry(extractGeometry(parsed))
        setError(null)
      } catch {
        setError('Could not read this file. Check that it is valid GeoJSON.')
      }
    }
    reader.readAsText(file)
  }

  return (
    <div>
      <div
        className={`drop-zone${dragOver ? ' drop-zone--active' : ''}`}
        onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragOver(false)
          const f = e.dataTransfer.files[0]
          if (f) readFile(f)
        }}
        onClick={() => inputRef.current?.click()}
      >
        <input
          ref={inputRef}
          type="file"
          accept=".geojson,application/geo+json"
          style={{ display: 'none' }}
          onChange={(e) => {
            const f = e.target.files?.[0]
            if (f) readFile(f)
          }}
        />
        <span>{dragOver ? 'Drop to load field' : 'Drop a .geojson file here, or click to browse'}</span>
      </div>
      {error && <p className="error-text">{error}</p>}
    </div>
  )
}
