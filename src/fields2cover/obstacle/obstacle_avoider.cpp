//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Wageningen University
//                        BSD-3 License
//=============================================================================

#include "fields2cover/obstacle/obstacle_avoider.h"

#include <stdexcept>

namespace f2c::obstacle {

F2CSwaths ObstacleAvoider::avoid(
    const F2CSwaths& swaths,
    const F2CCell& obstacle,
    double safety_margin) const {
  if (safety_margin < 0.0) {
    throw std::invalid_argument(
        "ObstacleAvoider::avoid: safety_margin must be >= 0");
  }

  // 1. Inflate the obstacle by the safety margin.
  F2CCell inflated = F2CCell::buffer(obstacle, safety_margin);

  F2CSwaths result;
  int out_id = 0;

  for (size_t i = 0; i < swaths.size(); ++i) {
    const F2CSwath& sw = swaths[i];
    F2CLineString path = sw.getPath();

    // 2. Build a bounding-box Cells that covers the swath + padding,
    //    then subtract the inflated obstacle to obtain the safe region.
    //    Padding = safety_margin + 1.0 ensures the bbox is never swallowed
    //    by a large obstacle (minimum 1 m floor when safety_margin == 0).
    const double padding = safety_margin + 1.0;
    const double x0 = path.getDimMinX() - padding;
    const double x1 = path.getDimMaxX() + padding;
    const double y0 = path.getDimMinY() - padding;
    const double y1 = path.getDimMaxY() + padding;
    F2CLinearRing ring{
        F2CPoint(x0, y0), F2CPoint(x1, y0),
        F2CPoint(x1, y1), F2CPoint(x0, y1),
        F2CPoint(x0, y0)};
    F2CCells bbox_cells(F2CCell{ring});
    F2CCells safe = bbox_cells.difference(inflated);

    // 3. Clip the swath path against the safe region.
    //    getLinesInside returns only the portions of `path` inside `safe`.
    F2CMultiLineString residuals = safe.getLinesInside(path);

    // 4. Filter by minimum length and emit new swaths.
    //    IDs are reassigned sequentially so every output swath has a unique id.
    //    Original width and type are preserved.
    for (size_t j = 0; j < residuals.size(); ++j) {
      F2CLineString seg = residuals.getGeometry(j);
      if (seg.length() >= kMinSegmentLength) {
        result.emplace_back(seg, sw.getWidth(), out_id++, sw.getType());
      }
    }
  }

  return result;
}

}  // namespace f2c::obstacle
