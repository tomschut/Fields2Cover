//=============================================================================
//    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
//                     Author: Gonzalo Mier
//                        BSD-3 License
//=============================================================================

#include <type_traits>
#include <gtest/gtest.h>
#include "fields2cover/types/Swaths.h"
#include "fields2cover/types/SwathsByCells.h"

// T-012 / T-013 regression: ensure Swaths and SwathsByCells advertise
// noexcept moves so std::vector<> reallocation takes the move path instead
// of silently falling back to copy.

static_assert(std::is_nothrow_move_constructible_v<f2c::types::Swaths>,
    "Swaths must be nothrow move constructible (T-012)");
static_assert(std::is_nothrow_move_assignable_v<f2c::types::Swaths>,
    "Swaths must be nothrow move assignable (T-012)");
static_assert(std::is_nothrow_move_constructible_v<f2c::types::SwathsByCells>,
    "SwathsByCells must be nothrow move constructible (T-013)");
static_assert(std::is_nothrow_move_assignable_v<f2c::types::SwathsByCells>,
    "SwathsByCells must be nothrow move assignable (T-013)");

TEST(SwathsMove, NoexceptMoveTraits) {
  EXPECT_TRUE(std::is_nothrow_move_constructible_v<f2c::types::Swaths>);
  EXPECT_TRUE(std::is_nothrow_move_assignable_v<f2c::types::Swaths>);
  EXPECT_TRUE(std::is_nothrow_move_constructible_v<f2c::types::SwathsByCells>);
  EXPECT_TRUE(std::is_nothrow_move_assignable_v<f2c::types::SwathsByCells>);
}
