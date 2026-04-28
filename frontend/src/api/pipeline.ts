import ky, { HTTPError } from 'ky'
import type { Feature } from 'geojson'

// --- Enum types ---
export type SortAlgorithm = 'BOUSTROPHEDON' | 'SNAKE' | 'SPIRAL'
export type TurningAlgorithm = 'DUBINS' | 'DUBINS_CC' | 'REEDS_SHEPP' | 'REEDS_SHEPP_HC'

// --- Request ---
export interface Robot {
  width_m: number
  cov_width_m: number
  cruise_vel_mps?: number
}

export interface PlanCoverageRequest {
  geojson: { type: 'FeatureCollection'; features: object[] }
  robot: Robot
  headland_width_m: number
  swath_angle_rad?: number
  swath_width_m: number
  sort_algorithm: SortAlgorithm
  turning_algorithm: TurningAlgorithm
  sort_variant?: number
  sort_start_point?: [number, number]  // WGS84 [lng, lat] — server picks closest variant
}

// --- PlanSwaths types ---
export type PlanSwathsRequest = PlanCoverageRequest

export interface PlanSwathsResponse {
  field_with_headlands: FieldWithHeadlands
  sorted_swaths: Swaths
  journey: object
}

// --- Response ---
export interface PathState {
  point: [number, number]
  angle_rad: number
  length_m: number
  velocity: number
  direction: 'FORWARD' | 'BACKWARD'
  section_type: 'SWATH' | 'TURN'
}

export interface GeoJSONFeatureCollection {
  type: 'FeatureCollection'
  features: Feature[]
}

export interface Field {
  boundary: GeoJSONFeatureCollection
  crs?: object
}

export interface FieldWithHeadlands {
  field: Field
  headlands: GeoJSONFeatureCollection
  inner_field: GeoJSONFeatureCollection
  headland_width_m: number
}

export interface Swath {
  id: number
  path: { type: 'LineString'; coordinates: number[][] }
  width_m: number
  type: 'MAINLAND' | 'HEADLAND'
}

export interface Swaths {
  field: FieldWithHeadlands
  items: Swath[]
  sorted?: boolean
}

export interface RouteConnection {
  from_swath_id: number
  to_swath_id: number
  points: [number, number][]
}

export interface Route {
  field: FieldWithHeadlands
  swaths: Swaths
  connections: RouteConnection[]
}

export interface Path {
  field: FieldWithHeadlands
  states: PathState[]
  length_m: number
  task_time_s: number
}

export interface FieldIntermediates {
  field?: Field | null
  field_with_headlands?: FieldWithHeadlands | null
  swaths?: Swaths | null
  sorted_swaths?: Swaths | null
  route?: Route | null
}

export interface PlanCoverageResponse {
  path: Path
  intermediates: FieldIntermediates
  journey: object
}

// --- API call ---
export async function planCoverage(req: PlanCoverageRequest): Promise<PlanCoverageResponse> {
  try {
    return await ky.post('/pipeline/plan-coverage', { json: req }).json<PlanCoverageResponse>()
  } catch (err) {
    if (err instanceof HTTPError) {
      const status = err.response.status
      if (status >= 400 && status < 500) {
        throw new Error('Invalid parameters. Check that field geometry is a valid polygon and all values are in range.')
      }
      throw new Error('Server error. The pipeline could not complete. Try again or adjust parameters.')
    }
    throw new Error('Could not reach the server. Check your connection and try again.')
  }
}

export async function planSwaths(req: PlanSwathsRequest): Promise<PlanSwathsResponse> {
  try {
    return await ky.post('/pipeline/plan-swaths', { json: req }).json<PlanSwathsResponse>()
  } catch (err) {
    if (err instanceof HTTPError) {
      const status = err.response.status
      if (status >= 400 && status < 500) {
        throw new Error('Invalid parameters. Check that field geometry is a valid polygon and all values are in range.')
      }
      throw new Error('Server error. The pipeline could not complete. Try again or adjust parameters.')
    }
    throw new Error('Could not reach the server. Check your connection and try again.')
  }
}

export function pickBestVariant(
  startPoint: [number, number],
  attempts: Array<{ variant: number; firstSwathStart: [number, number] | null }>,
): number {
  // startPoint is [lat, lng] (Leaflet); firstSwathStart is [lng, lat] (GeoJSON)
  // Convert startPoint to [lng, lat] for consistent comparison
  const [lat, lng] = startPoint
  let bestVariant = 0
  let bestDist = Infinity
  for (const attempt of attempts) {
    if (attempt.firstSwathStart === null) continue
    const [swathLng, swathLat] = attempt.firstSwathStart
    const dist = Math.sqrt((lng - swathLng) ** 2 + (lat - swathLat) ** 2)
    if (dist < bestDist) {
      bestDist = dist
      bestVariant = attempt.variant
    }
  }
  return bestVariant
}
