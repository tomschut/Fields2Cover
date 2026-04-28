import type { Geometry } from 'geojson'

interface FieldStatusProps {
  fieldGeometry: Geometry | null
}

function countVertices(geometry: Geometry): number {
  if (geometry.type === 'Polygon') {
    return (geometry as GeoJSON.Polygon).coordinates[0]?.length ?? 0
  }
  if (geometry.type === 'MultiPolygon') {
    return (geometry as GeoJSON.MultiPolygon).coordinates.reduce(
      (sum, poly) => sum + (poly[0]?.length ?? 0), 0
    )
  }
  return 0
}

export function FieldStatus({ fieldGeometry }: FieldStatusProps) {
  return (
    <div className="field-status">
      {fieldGeometry
        ? <span className="badge">Field ready &middot; {countVertices(fieldGeometry)} vertices</span>
        : <span>No field defined</span>
      }
    </div>
  )
}
