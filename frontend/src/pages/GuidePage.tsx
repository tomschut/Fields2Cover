import { CodeTabs } from '../components/CodeTabs'

const CURL_EXAMPLE = `curl -X POST http://localhost:8080/pipeline/plan-coverage \\
  -H "Content-Type: application/json" \\
  -d '{
    "geojson": {
      "type": "FeatureCollection",
      "features": [{
        "type": "Feature",
        "geometry": {
          "type": "Polygon",
          "coordinates": [[[5.0,52.0],[5.01,52.0],[5.01,52.01],[5.0,52.01],[5.0,52.0]]]
        },
        "properties": null
      }]
    },
    "robot": { "width_m": 3.0, "cov_width_m": 2.5 },
    "headland_width_m": 5.0,
    "swath_width_m": 2.5,
    "turning_algorithm": "DUBINS"
  }'`

const JS_EXAMPLE = `const response = await fetch('/pipeline/plan-coverage', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    geojson: {
      type: 'FeatureCollection',
      features: [{
        type: 'Feature',
        geometry: {
          type: 'Polygon',
          coordinates: [[[5.0,52.0],[5.01,52.0],[5.01,52.01],[5.0,52.01],[5.0,52.0]]]
        },
        properties: null
      }]
    },
    robot: { width_m: 3.0, cov_width_m: 2.5 },
    headland_width_m: 5.0,
    swath_width_m: 2.5,
    turning_algorithm: 'DUBINS'
  })
})
const result = await response.json()`

const GO_EXAMPLE = `body := \`{
  "geojson": {
    "type": "FeatureCollection",
    "features": [{
      "type": "Feature",
      "geometry": {
        "type": "Polygon",
        "coordinates": [[[5.0,52.0],[5.01,52.0],[5.01,52.01],[5.0,52.01],[5.0,52.0]]]
      },
      "properties": null
    }]
  },
  "robot": { "width_m": 3.0, "cov_width_m": 2.5 },
  "headland_width_m": 5.0,
  "swath_width_m": 2.5,
  "turning_algorithm": "DUBINS"
}\`

req, _ := http.NewRequest("POST", "http://localhost:8080/pipeline/plan-coverage",
  strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
resp, _ := http.DefaultClient.Do(req)
defer resp.Body.Close()`

export default function GuidePage() {
  return (
    <div className="docs-page">
      <h1 className="docs-title">Integration Guide</h1>
      <p className="docs-lead">
        Everything you need to call the Fields2Cover API from your own application.
      </p>

      <h2 className="docs-section-heading">Prerequisites</h2>
      <p className="docs-prose">
        The Fields2Cover API server must be running on port 8080. Start it with{' '}
        <code>docker compose up</code> or by running the Go binary directly.
        All endpoints accept and return JSON.
      </p>

      <h2 className="docs-section-heading">Full Pipeline Call</h2>
      <p className="docs-prose">
        The one-shot endpoint <code>POST /pipeline/plan-coverage</code> runs the full
        coverage planning pipeline and returns headlands, swaths, route, and path in a
        single response.
      </p>
      <CodeTabs curl={CURL_EXAMPLE} js={JS_EXAMPLE} go={GO_EXAMPLE} />

      <h2 className="docs-section-heading">Request Parameters</h2>
      <table className="docs-table">
        <thead>
          <tr>
            <th>Field</th>
            <th>Type</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>geojson</code></td>
            <td>GeoJSONFeatureCollection</td>
            <td>Field boundary polygon as a GeoJSON FeatureCollection (required)</td>
          </tr>
          <tr>
            <td><code>robot.width_m</code></td>
            <td>number</td>
            <td>Total robot width in metres (required)</td>
          </tr>
          <tr>
            <td><code>robot.cov_width_m</code></td>
            <td>number</td>
            <td>Effective coverage width in metres (required)</td>
          </tr>
          <tr>
            <td><code>headland_width_m</code></td>
            <td>number</td>
            <td>Headland buffer width in metres (required)</td>
          </tr>
          <tr>
            <td><code>swath_width_m</code></td>
            <td>number</td>
            <td>Swath spacing in metres (required)</td>
          </tr>
          <tr>
            <td><code>turning_algorithm</code></td>
            <td>string</td>
            <td>
              Turn type: <code>DUBINS</code>, <code>DUBINS_CC</code>,{' '}
              <code>REEDS_SHEPP</code>, or <code>REEDS_SHEPP_HC</code> (required)
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}
