# Phase 25c: C++ Obstacle Avoider - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped via workflow.skip_discuss)

<domain>
## Phase Boundary

The f2c library can fragment swaths around polygon obstacles, producing trimmed swath segments that avoid inflated obstacle boundaries.

A call to `ObstacleAvoider::avoid()` with a list of swaths and a polygon obstacle returns swath segments with no overlap with the inflated obstacle polygon (inflation = safety margin parameter). Swath segments shorter than a minimum threshold (0.1 m) are dropped rather than returned.

This follows the same module pattern as Phase 24 (MultiRobotPartition), 25a (SpatialRtreePartition/LengthBalancedPartition), and 25b (GraphRouteOptimizer) — new header in `include/fields2cover/obstacle/`, source in `src/fields2cover/obstacle/`, tests in `tests/cpp/obstacle/`.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — discuss phase was skipped per user setting. Use ROADMAP phase goal, success criteria, and codebase conventions to guide decisions.

Key decisions for Claude to make:
- Obstacle inflation strategy (buffer polygon by safety_margin, use GEOS/Boost.Geometry)
- Swath fragmentation: clip each swath LineString against the inflated obstacle polygon, collect residual segments
- Minimum segment threshold: drop segments < 0.1 m (per success criteria)
- Unit test: place known obstacle across a swath so it splits into 2 segments

</decisions>

<code_context>
## Existing Code Insights

Codebase context will be gathered during plan-phase research.

</code_context>

<specifics>
## Specific Ideas

**Requirements reference:** F2C-05

**Success Criteria:**
1. `ObstacleAvoider::avoid()` with swaths + polygon obstacle returns segments with no overlap with inflated obstacle polygon (inflation = safety_margin parameter)
2. Swath segments shorter than 0.1 m are dropped
3. All existing GoogleTest unit tests continue to pass
4. Unit test with known obstacle placement splits at least one swath into two segments

**Inspired by:** farmtrax ObstacleAvoider pattern — fragment swaths around inflated polygon obstacles.

</specifics>

<deferred>
## Deferred Ideas

None — discuss phase skipped.

</deferred>
