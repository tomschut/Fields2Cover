//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/partition/spatial_rtree_partition.h"

#include <algorithm>
#include <limits>
#include <numeric>
#include <vector>

// boost::geometry R-tree — included only here to avoid header contamination.
#include <boost/geometry.hpp>
#include <boost/geometry/index/rtree.hpp>

#include "fields2cover/swath_generator/brute_force.h"
#include "fields2cover/objectives/sg_obj/swath_length.h"

namespace f2c::partition {

namespace bg  = boost::geometry;
namespace bgi = boost::geometry::index;

using BgPoint = bg::model::point<double, 2, bg::cs::cartesian>;
using Value   = std::pair<BgPoint, std::size_t>;  // (midpoint, swath_index)

std::vector<F2CCells> SpatialRtreePartition::partition(
    const F2CCells& field,
    const std::vector<F2CRobot>& robots) const {
  if (robots.empty()) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: robots must not be empty");
  }
  if (robots[0].getCovWidth() <= 0.0) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: robots[0] must have positive coverage width");
  }
  if (field.size() != 1) {
    throw std::invalid_argument(
        "SpatialRtreePartition::partition: field must be a single-cell geometry (field.size() == 1)");
  }

  const std::size_t N = robots.size();

  // 1. Generate swaths from the single-cell field.
  f2c::sg::BruteForce sw_gen;
  f2c::obj::SwathLength obj;
  F2CSwaths swaths = sw_gen.generateBestSwaths(obj, robots[0].getCovWidth(), field.getCell(0));

  // 2. Build R-tree of all swath midpoints.
  bgi::rtree<Value, bgi::quadratic<16>> rtree;
  for (std::size_t i = 0; i < swaths.size(); ++i) {
    F2CPoint s = swaths[i].startPoint();
    F2CPoint e = swaths[i].endPoint();
    BgPoint mid((s.getX() + e.getX()) * 0.5, (s.getY() + e.getY()) * 0.5);
    rtree.insert({mid, i});
  }

  // 3. Select N evenly-spaced seed swaths — one per robot.
  std::vector<std::vector<std::size_t>> assigned(N);
  std::vector<bool> is_seed(swaths.size(), false);

  for (std::size_t r = 0; r < N; ++r) {
    std::size_t seed_idx = (swaths.size() == 0)
        ? 0
        : (r * swaths.size()) / N;
    if (seed_idx < swaths.size()) {
      assigned[r].push_back(seed_idx);
      is_seed[seed_idx] = true;
    }
  }

  // 4. Compute per-robot cluster centroid from assigned swaths.
  auto centroid = [&](std::size_t robot_idx) -> BgPoint {
    double cx = 0.0, cy = 0.0;
    std::size_t cnt = 0;
    for (std::size_t idx : assigned[robot_idx]) {
      F2CPoint s = swaths[idx].startPoint();
      F2CPoint e = swaths[idx].endPoint();
      cx += (s.getX() + e.getX()) * 0.5;
      cy += (s.getY() + e.getY()) * 0.5;
      ++cnt;
    }
    if (cnt == 0) {
      return BgPoint(0.0, 0.0);
    }
    return BgPoint(cx / cnt, cy / cnt);
  };

  // 5. Assign all non-seed swaths to the robot with the nearest cluster centroid.
  for (std::size_t i = 0; i < swaths.size(); ++i) {
    if (is_seed[i]) continue;

    F2CPoint s = swaths[i].startPoint();
    F2CPoint e = swaths[i].endPoint();
    BgPoint mid((s.getX() + e.getX()) * 0.5, (s.getY() + e.getY()) * 0.5);

    double best_dist = std::numeric_limits<double>::max();
    std::size_t best_robot = 0;

    for (std::size_t r = 0; r < N; ++r) {
      BgPoint c = centroid(r);
      double dx = bg::get<0>(mid) - bg::get<0>(c);
      double dy = bg::get<1>(mid) - bg::get<1>(c);
      double d2 = dx * dx + dy * dy;
      if (d2 < best_dist) {
        best_dist = d2;
        best_robot = r;
      }
    }
    assigned[best_robot].push_back(i);
  }

  // 6. Build zone geometry: union each robot's swaths' areaCovered().
  std::vector<F2CCells> zones;
  zones.reserve(N);
  for (std::size_t r = 0; r < N; ++r) {
    F2CCells zone;
    for (std::size_t idx : assigned[r]) {
      F2CCells covered = swaths[idx].areaCovered();
      zone = zone.unionOp(covered);
    }
    zones.push_back(zone);
  }

  return zones;
}

}  // namespace f2c::partition
