---
phase: 24
slug: c-multi-robot-partitioning
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-29
---

# Phase 24 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GoogleTest (system package) |
| **Config file** | `tests/CMakeLists.txt` + `tests/unittests.cpp` (main()) |
| **Quick run command** | `./build/tests/unittests --gtest_filter="fields2cover_partition*"` |
| **Full suite command** | `./build/tests/unittests` |
| **Estimated runtime** | ~30 seconds (full), ~5 seconds (filter) |

---

## Sampling Rate

- **After every task commit:** Run `make -C build unittests -j$(nproc) && ./build/tests/unittests --gtest_filter="fields2cover_partition*"`
- **After every plan wave:** Run `./build/tests/unittests`
- **Before `/gsd-verify-work`:** Full suite must be green (294 + new tests)
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 24-01-01 | 01 | 1 | F2C-01 | — | N/A | build | `make -C build unittests -j$(nproc) && ./build/tests/unittests` | ⬜ Wave 0 | ⬜ pending |
| 24-01-02 | 01 | 1 | F2C-01 | — | N/A | unit | `./build/tests/unittests --gtest_filter="fields2cover_partition_multi_robot.*"` | ⬜ Wave 0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `include/fields2cover/partition/multi_robot_partition.h` — class declaration (created by Task 1)
- [ ] `src/fields2cover/partition/multi_robot_partition.cpp` — strip-partition implementation (created by Task 1)
- [ ] `tests/cpp/partition/multi_robot_partition_test.cpp` — 2-robot + 3-robot test cases (created by Task 2)

*No new framework or build system changes required — CMake GLOB_RECURSE picks up new files automatically.*

---

## Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| F2C-01 | 2-robot equal-rate: zones have equal area | unit | `./build/tests/unittests --gtest_filter="fields2cover_partition_multi_robot.two_robots_equal_rate"` | ⬜ Wave 0 |
| F2C-01 | 3-robot proportional: zone areas match 2:1:1 ratio | unit | `./build/tests/unittests --gtest_filter="fields2cover_partition_multi_robot.three_robots_proportional"` | ⬜ Wave 0 |
| F2C-01 | Total area of all zones = input field area | unit (embedded in above tests) | same as above | ⬜ Wave 0 |
| F2C-01 | Existing 294 tests still pass | regression | `./build/tests/unittests` | ✅ Yes |

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
