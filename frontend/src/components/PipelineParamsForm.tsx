import { useState } from 'react'
import type { Geometry } from 'geojson'
import type { PlanCoverageRequest, SortAlgorithm, TurningAlgorithm } from '../api/pipeline'

interface PipelineParamsFormProps {
  fieldGeometry: Geometry | null
  onRun: (req: PlanCoverageRequest) => void
  loading: boolean
}

type FieldKey = 'robotWidth' | 'coverageWidth' | 'speed' | 'headlandWidth' | 'sortAlgorithm' | 'turningAlgorithm'

function validateField(key: FieldKey, value: string): string | null {
  if (key === 'speed') {
    if (value.trim() === '') return null // optional
    const n = parseFloat(value)
    if (isNaN(n) || n <= 0) return 'Must be greater than 0'
    return null
  }
  if (key === 'sortAlgorithm' || key === 'turningAlgorithm') {
    return null
  }
  if (!value.trim()) return 'Required'
  const n = parseFloat(value)
  if (isNaN(n) || n <= 0) return 'Must be greater than 0'
  return null
}

export function PipelineParamsForm({ fieldGeometry, onRun, loading }: PipelineParamsFormProps) {
  const [robotWidth, setRobotWidth] = useState('3.0')
  const [coverageWidth, setCoverageWidth] = useState('3.0')
  const [speed, setSpeed] = useState('')
  const [headlandWidth, setHeadlandWidth] = useState('6.0')
  const [sortAlgorithm, setSortAlgorithm] = useState<SortAlgorithm>('BOUSTROPHEDON')
  const [turningAlgorithm, setTurningAlgorithm] = useState<TurningAlgorithm>('DUBINS')
  const [errors, setErrors] = useState<Partial<Record<FieldKey, string>>>({})

  const handleBlur = (key: FieldKey, value: string) => {
    const err = validateField(key, value)
    setErrors(prev => ({ ...prev, [key]: err ?? undefined }))
  }

  const hasErrors = () => {
    const keys: FieldKey[] = ['robotWidth', 'coverageWidth', 'headlandWidth']
    return keys.some(k => {
      const val = k === 'robotWidth' ? robotWidth : k === 'coverageWidth' ? coverageWidth : headlandWidth
      return validateField(k, val) !== null
    })
  }

  const isDisabled = fieldGeometry === null || loading || hasErrors()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    // Validate all fields on submit
    const newErrors: Partial<Record<FieldKey, string>> = {}
    const fieldValues: Record<FieldKey, string> = {
      robotWidth,
      coverageWidth,
      speed,
      headlandWidth,
      sortAlgorithm,
      turningAlgorithm,
    }
    let hasAnyError = false
    for (const [key, val] of Object.entries(fieldValues) as [FieldKey, string][]) {
      const err = validateField(key, val)
      if (err) {
        newErrors[key] = err
        hasAnyError = true
      }
    }

    if (hasAnyError) {
      setErrors(newErrors)
      return
    }

    if (!fieldGeometry) return

    const req: PlanCoverageRequest = {
      geojson: {
        type: 'FeatureCollection',
        features: [{
          type: 'Feature',
          geometry: fieldGeometry,
          properties: {},
        }],
      },
      robot: {
        width_m: parseFloat(robotWidth),
        cov_width_m: parseFloat(coverageWidth),
        ...(speed.trim() ? { cruise_vel_mps: parseFloat(speed) } : {}),
      },
      headland_width_m: parseFloat(headlandWidth),
      // swath_angle_rad omitted — backend auto-selects the heading that minimises swath count
      swath_width_m: parseFloat(coverageWidth),
      sort_algorithm: sortAlgorithm,
      turning_algorithm: turningAlgorithm,
      sort_variant: 0, // overridden by MapPage.handleRunPipeline when startPoint is set
    }

    onRun(req)
  }

  return (
    <form className="param-form" onSubmit={handleSubmit} noValidate>
      <h2 className="section-heading">Pipeline parameters</h2>

      <div className="param-row">
        <label className="param-label">Robot width (m)</label>
        <input
          className={`param-input${errors.robotWidth ? ' param-input--error' : ''}`}
          type="number"
          min="0.1"
          step="0.1"
          value={robotWidth}
          onChange={e => setRobotWidth(e.target.value)}
          onBlur={() => handleBlur('robotWidth', robotWidth)}
          aria-label="Robot width (m)"
        />
        {errors.robotWidth && <span className="param-error">{errors.robotWidth}</span>}
      </div>

      <div className="param-row">
        <label className="param-label">Coverage width (m)</label>
        <input
          className={`param-input${errors.coverageWidth ? ' param-input--error' : ''}`}
          type="number"
          min="0.1"
          step="0.1"
          value={coverageWidth}
          onChange={e => setCoverageWidth(e.target.value)}
          onBlur={() => handleBlur('coverageWidth', coverageWidth)}
          aria-label="Coverage width (m)"
        />
        {errors.coverageWidth && <span className="param-error">{errors.coverageWidth}</span>}
      </div>

      <div className="param-row">
        <label className="param-label">Speed (m/s)</label>
        <input
          className={`param-input${errors.speed ? ' param-input--error' : ''}`}
          type="number"
          min="0.1"
          step="0.1"
          value={speed}
          onChange={e => setSpeed(e.target.value)}
          onBlur={() => handleBlur('speed', speed)}
          aria-label="Speed (m/s)"
        />
        {errors.speed && <span className="param-error">{errors.speed}</span>}
      </div>

      <div className="param-row">
        <label className="param-label">Headland width (m)</label>
        <input
          className={`param-input${errors.headlandWidth ? ' param-input--error' : ''}`}
          type="number"
          min="0.1"
          step="0.1"
          value={headlandWidth}
          onChange={e => setHeadlandWidth(e.target.value)}
          onBlur={() => handleBlur('headlandWidth', headlandWidth)}
          aria-label="Headland width (m)"
        />
        {errors.headlandWidth && <span className="param-error">{errors.headlandWidth}</span>}
      </div>

      <div className="param-row">
        <label className="param-label">Sort algorithm</label>
        <select
          className="param-select"
          value={sortAlgorithm}
          onChange={e => setSortAlgorithm(e.target.value as SortAlgorithm)}
          aria-label="Sort algorithm"
        >
          <option value="BOUSTROPHEDON">Boustrophedon</option>
          <option value="SNAKE">Snake</option>
          <option value="SPIRAL">Spiral</option>
        </select>
      </div>

      <div className="param-row">
        <label className="param-label">Turning algorithm</label>
        <select
          className="param-select"
          value={turningAlgorithm}
          onChange={e => setTurningAlgorithm(e.target.value as TurningAlgorithm)}
          aria-label="Turning algorithm"
        >
          <option value="DUBINS">Dubins</option>
          <option value="DUBINS_CC">Dubins CC</option>
          <option value="REEDS_SHEPP">Reeds-Shepp</option>
          <option value="REEDS_SHEPP_HC">Reeds-Shepp HC</option>
        </select>
      </div>

      <button
        type="submit"
        className={`run-btn${loading ? ' run-btn--loading' : ''}`}
        disabled={isDisabled}
        title={fieldGeometry === null ? 'Draw or load a field first' : undefined}
      >
        {loading ? 'Running\u2026' : 'Run pipeline'}
      </button>
    </form>
  )
}
