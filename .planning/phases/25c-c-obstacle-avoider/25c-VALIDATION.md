---
phase: 25c
slug: c-obstacle-avoider
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-29
---

# Phase 25c — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GoogleTest (gtest) |
| **Config file** | tests/CMakeLists.txt — GLOB_RECURSE auto-discovers test files |
| **Quick run command** | `cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure` |
| **Full suite command** | `cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure` |
| **Estimated runtime** | ~30 seconds (full suite ~60s) |

---

## Sampling Rate

- **After every task commit:** Run `cd /home/tom/devenv/fields2cover/build && ctest -R "obstacle" --output-on-failure`
- **After every plan wave:** Run `cd /home/tom/devenv/fields2cover/build && ctest --output-on-failure`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 25c-01-01 | 01 | 1 | F2C-05 | — | N/A | unit | `ctest -R obstacle --output-on-failure` | ❌ W0 | ⬜ pending |
| 25c-01-02 | 01 | 1 | F2C-05 | — | N/A | unit | `ctest -R obstacle --output-on-failure` | ❌ W0 | ⬜ pending |
| 25c-01-03 | 01 | 1 | F2C-05 | — | N/A | regression | `ctest --output-on-failure` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `include/fields2cover/obstacle/obstacle_avoider.h` — ObstacleAvoider header (needed before test compiles)
- [ ] `src/fields2cover/obstacle/obstacle_avoider.cpp` — implementation
- [ ] `tests/cpp/obstacle/obstacle_avoider_test.cpp` — stubs for F2C-05-a, F2C-05-b, F2C-05-c
- [ ] cmake re-run required after new `obstacle/` directory is created

*All three files are created in a single plan wave; cmake re-run is the first build step.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
