#==============================================================================
#     Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                      Author: Gonzalo Mier
#                         BSD-3 License
#==============================================================================

import pytest
import fields2cover as f2c

def near(a, b):
  assert pytest.approx(a) == pytest.approx(b)


def _make_simple_example_inputs():
  """Build the 100x100 field with two inner rings shared by simple_example
  and redirect_flag tests (mirrors the C++ fixture)."""
  robot = f2c.Robot(2.0, 2.0);
  robot.setMinTurningRadius(2);
  robot.setMaxDiffCurv(0.1);

  const_hl = f2c.HG_Const_gen();
  ring1 = f2c.LinearRing()
  ring2 = f2c.LinearRing()
  ring3 = f2c.LinearRing()
  ring1.addPoint(0,0); ring1.addPoint(100,0); ring1.addPoint(100,100); ring1.addPoint(0,100); ring1.addPoint(0,0)
  ring2.addPoint(20,20); ring2.addPoint(20,30); ring2.addPoint(30,30); ring2.addPoint(30,20); ring2.addPoint(20,20)
  ring3.addPoint(60,60); ring3.addPoint(80,60); ring3.addPoint(80,80); ring3.addPoint(60,80); ring3.addPoint(60,60)

  cells = f2c.Cells(f2c.Cell(ring1))
  cells.addRing(0, ring2)
  cells.addRing(0, ring3)

  no_hl = const_hl.generateHeadlandArea(cells, robot.getCovWidth(), 3);
  hl_swaths = const_hl.generateHeadlandSwaths(cells, robot.getCovWidth(), 3, False);

  bf = f2c.SG_BruteForce();
  swaths = bf.generateSwaths(3.1416, robot.getCovWidth(), no_hl);

  return cells, hl_swaths, swaths


def test_fields2cover_route_planner_base_simple_example():
  cells, hl_swaths, swaths = _make_simple_example_inputs()

  route_planner = f2c.RP_RoutePlannerBase()
  route = route_planner.genRoute(hl_swaths[1], swaths)

  # CI-safe assertions mirroring the C++ simple_example test.
  # The visualization block from the original test was removed to avoid
  # blocking visualizer display calls on headless CI runners.
  assert not route.isEmpty()
  assert route.sizeVectorSwaths() > 1
  assert route.sizeVectorSwaths() == route.sizeConnections()


def test_fields2cover_route_planner_base_redirect_flag():
  # Mirrors the C++ redirect_flag test: when redirect_swaths=False, the
  # resulting route must preserve each swath's original direction.
  cells, hl_swaths, swaths = _make_simple_example_inputs()

  route_planner = f2c.RP_RoutePlannerBase()
  # Positional form: (cells, swaths, show_log, d_tol, redirect_swaths)
  route = route_planner.genRoute(hl_swaths[1], swaths, False, 1e-4, False)

  old_swaths = swaths.flatten()
  new_swaths = f2c.Swaths()
  for sbc in range(route.sizeVectorSwaths()):
    route_sw = route.getSwaths(sbc)
    for i in range(route_sw.size()):
      new_swaths.push_back(route_sw.at(i))

  assert new_swaths.size() == old_swaths.size()

  for s in range(new_swaths.size()):
    old_swath = old_swaths.at(s)
    new_swath = new_swaths.at(s)
    assert new_swath.hasSameDir(old_swath)

  assert not route.isEmpty()
  assert route.sizeVectorSwaths() > 1
  assert route.sizeVectorSwaths() == route.sizeConnections()


def test_fields2cover_route_planner_base_start_and_end_points():
  # Regression test for PR #177 crash: r_start coincident with a swath
  # endpoint on the border. See .planning/research/route-planner-crash.md.
  # The pre-fix route planner threw (OR-Tools infeasibility / self-loop in
  # shortest graph). With plan 04-01's fixes, genRoute must succeed.
  ring = f2c.LinearRing()
  ring.addPoint(0, 0)
  ring.addPoint(10, 0)
  ring.addPoint(10, 10)
  ring.addPoint(0, 10)
  ring.addPoint(0, 0)
  cells = f2c.Cells(f2c.Cell(ring))

  swaths = f2c.Swaths()
  swaths.push_back(f2c.Swath(f2c.LineString(f2c.VectorPoint(
      [f2c.Point(0, 2), f2c.Point(10, 2)])), 2.0))
  swaths.push_back(f2c.Swath(f2c.LineString(f2c.VectorPoint(
      [f2c.Point(0, 5), f2c.Point(10, 5)])), 2.0))
  swaths.push_back(f2c.Swath(f2c.LineString(f2c.VectorPoint(
      [f2c.Point(0, 8), f2c.Point(10, 8)])), 2.0))

  sbc = f2c.SwathsByCells()
  sbc.push_back(swaths)

  r_start = swaths.at(0).startPoint()  # (0, 2) — on the border
  r_end   = swaths.at(2).endPoint()    # (10, 8) — on the border

  planner = f2c.RP_RoutePlannerBase()
  planner.setStartAndEndPoint(r_start, r_end)

  # Any exception here (pre-fix: OR-Tools nullptr deref / infeasibility)
  # will fail the test via pytest's default uncaught-exception handling.
  route = planner.genRoute(cells, sbc)
  assert not route.isEmpty()
