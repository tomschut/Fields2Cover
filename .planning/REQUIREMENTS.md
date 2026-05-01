# Requirements: Fields2Cover v4.0 Multi-Robot Planning

**Defined:** 2026-04-29
**Core Value:** Every API call either produces a directly-usable result or produces a typed handoff that the next call in the journey accepts without adaptation.

## v4.0 Requirements

Requirements for the Multi-Robot Planning milestone. Phases continue from 23.

### C++ Library Extensions

- [x] **F2C-01**: f2c library exposes a multi-robot field partitioning algorithm that divides a field into N zones, where zone sizes are proportional to each robot's work rate (robot_width × robot_speed)
- [x] **F2C-02**: f2c library exposes a follower coordination algorithm that, given a robot's coverage route and follower specs (tank capacity, unload time, travel speed), computes the cart's travel path and the headland rendezvous points where unloading occurs
- [x] **F2C-03**: f2c library exposes SPATIAL_RTREE and LENGTH_BALANCED multi-robot partition strategies as alternatives to the strip-based partitioner
- [x] **F2C-04**: f2c library exposes a graph-based swath traversal optimizer (GraphRouteOptimizer) that minimizes total travel distance using direction-aware Dijkstra scoring
- [x] **F2C-05**: f2c library exposes an obstacle avoidance algorithm (ObstacleAvoider) that fragments swaths around inflated polygon obstacles, dropping segments shorter than 0.1 m

### Multi-Robot Coverage (MRD)

- [ ] **MRD-01**: User can define a fleet of N robots, each with independent specs (working width, speed, headland width, route planner variant, path planner variant)
- [ ] **MRD-02**: System partitions the field into N zones proportional to each robot's work rate; a robot with double the work rate receives double the area
- [ ] **MRD-03**: Each robot zone receives a full independent coverage plan (headlands, swaths, route, path) using that robot's specs
- [ ] **MRD-04**: All robot paths are displayed simultaneously on the map in distinct per-robot colors
- [ ] **MRD-05**: User can trigger the full multi-robot plan from a single "Run" action and see all results in one response

### Follower Coordination (FOL)

- [ ] **FOL-01**: User can configure a follower alongside the primary robot (tank capacity in kg, unload time in seconds, follower travel speed in m/s)
- [ ] **FOL-02**: System computes the cart's travel path based on the robot's planned route and the cart configuration
- [ ] **FOL-03**: Headland rendezvous/unload meeting points are identified and marked on the map
- [ ] **FOL-04**: The cart's full path and the rendezvous markers are displayed as distinct overlays on the map, togglable independently of robot result layers

### API & Infrastructure (API)

- [ ] **API-01**: New gRPC RPC for multi-robot planning: accepts field geometry + N robot specs, returns N zone geometries + N full coverage plans
- [ ] **API-02**: New gRPC RPC for follower coordination: accepts robot path + follower specs, returns follower path + rendezvous point list
- [ ] **API-03**: New Go API endpoint `POST /pipeline/plan-multi-robot` with OpenAPI spec, typed request/response
- [ ] **API-04**: New Go API endpoint `POST /pipeline/plan-follower` with OpenAPI spec, typed request/response

## Future Requirements

### Multi-Robot Enhancements

- **MRD-F01**: User can drag zone boundaries on the map to manually adjust the automatic partition
- **MRD-F02**: Conflict/collision avoidance — system schedules robot start times or headland order to prevent robots meeting on the same headland
- **MRD-F03**: Export multi-robot plan as separate GPX/KML files, one per robot

### Follower Enhancements

- **FOL-F01**: Multiple carts — plan coordination for more than one trailing cart simultaneously
- **FOL-F02**: Configurable unload strategy — headland-only vs. on-the-go (in-field) unloading

### UX

- **UX-F01**: Robot fleet panel with per-robot color picker and name
- **UX-F02**: Animation playback — step through the multi-robot route in time order

## Out of Scope

| Feature | Reason |
|---------|--------|
| Real-time robot tracking / telemetry | Requires external hardware integration |
| Conflict avoidance scheduling | High complexity — deferred to v4.1 |
| Multiple cart coordination | Single cart is the common case; deferred |
| On-the-go (in-field) unloading | Requires different follower path model; deferred |
| Mobile app | Web-first remains |
| User authentication | Demo tool |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| F2C-01 | Phase 24 | Complete |
| F2C-02 | Phase 25 | Complete |
| F2C-03 | Phase 25a | Complete |
| F2C-04 | Phase 25b | Pending |
| F2C-05 | Phase 25c | Complete |
| MRD-01 | Phase 27 | Pending |
| MRD-02 | Phase 27 | Pending |
| MRD-03 | Phase 27 | Pending |
| MRD-04 | Phase 27 | Pending |
| MRD-05 | Phase 27 | Pending |
| FOL-01 | Phase 28 | Pending |
| FOL-02 | Phase 28 | Pending |
| FOL-03 | Phase 28 | Pending |
| FOL-04 | Phase 28 | Pending |
| API-01 | Phase 26 | Pending |
| API-02 | Phase 26 | Pending |
| API-03 | Phase 26 | Pending |
| API-04 | Phase 26 | Pending |

---
*Requirements defined: 2026-04-29*
