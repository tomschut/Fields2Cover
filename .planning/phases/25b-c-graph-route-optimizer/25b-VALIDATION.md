---
phase: 25b
slug: c-graph-route-optimizer
status: ready
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-29
---

# Phase 25b — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | GoogleTest (CMake + ctest) |
| **Config file** | CMakeLists.txt (GLOB_RECURSE auto-discovers new test files) |
| **Quick run command** | `cd build && ctest -R graph_route_optimizer --output-on-failure` |
| **Full suite command** | `cd build && cmake .. && make -j$(nproc) && ctest --output-on-failure` |
| **Estimated runtime** | ~60 seconds (full rebuild + 313 tests) |

---

## Sampling Rate

- **After every task commit:** Run `cd build && ctest -R graph_route_optimizer --output-on-failure`
- **After every plan wave:** Run full suite `cd build && cmake .. && make -j$(nproc) && ctest --output-on-failure`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 25b-01-01 | 01 | 1 | F2C-04 | unit | `cd build && cmake .. && make graph_route_optimizer_test && ctest -R graph_route_optimizer` | ⬜ pending |
| 25b-01-02 | 01 | 1 | F2C-04 | regression | `cd build && ctest --output-on-failure` | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — GoogleTest + CMake GLOB_RECURSE auto-discovers new test files. No new framework installation needed.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
