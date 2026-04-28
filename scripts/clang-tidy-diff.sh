#!/usr/bin/env bash
# scripts/clang-tidy-diff.sh
#
# Runs clang-tidy across src/fields2cover and include/fields2cover, diffs the
# findings against build/audit/clang-tidy-baseline.txt, and exits non-zero if
# any NEW high-severity finding is introduced (HRD-02 CI gate).
#
# High-severity check families (blocking):
#   - bugprone-*
#   - clang-analyzer-*
#   - cppcoreguidelines-pro-type-*
#
# All other families (performance-*, modernize-*, readability-*) are surfaced
# as informational new-finding counts but do NOT fail the build. This is the
# deliberate T-07-03-02 acceptance in the plan's threat register.
#
# Environment overrides:
#   BASELINE   — path to baseline file (default: build/audit/clang-tidy-baseline.txt)
#   BUILD_DIR  — compile_commands.json build directory (default: build)

set -euo pipefail

BASELINE="${BASELINE:-build/audit/clang-tidy-baseline.txt}"
BUILD_DIR="${BUILD_DIR:-build}"

if [[ ! -f "$BASELINE" ]]; then
  echo "::error::baseline $BASELINE missing" >&2
  exit 2
fi

if [[ ! -f "$BUILD_DIR/compile_commands.json" ]]; then
  echo "::error::$BUILD_DIR/compile_commands.json missing — configure cmake with -DCMAKE_EXPORT_COMPILE_COMMANDS=ON" >&2
  exit 2
fi

tmp_raw=$(mktemp)
tmp_current=$(mktemp)
tmp_new=$(mktemp)
trap 'rm -f "$tmp_raw" "$tmp_current" "$tmp_new"' EXIT

# Collect source files (same set as baseline generation)
mapfile -t SOURCES < <(find src/fields2cover include/fields2cover \
  \( -name '*.cpp' -o -name '*.hpp' -o -name '*.h' \))

run-clang-tidy -p "$BUILD_DIR" -quiet "${SOURCES[@]}" > "$tmp_raw" 2>&1 || true

sed -E "s|$(pwd)/||g" "$tmp_raw" \
  | grep -E ": warning: " \
  | sort -u > "$tmp_current"

# Findings present in current but not in baseline
comm -23 "$tmp_current" "$BASELINE" > "$tmp_new"

# High-severity filter
high_severity_pattern='\[(bugprone-|clang-analyzer-|cppcoreguidelines-pro-type-)'
new_high=$(grep -E "$high_severity_pattern" "$tmp_new" || true)

new_count=$(wc -l < "$tmp_new")

if [[ -n "$new_high" ]]; then
  echo "::error::new high-severity clang-tidy findings detected:" >&2
  echo "$new_high" >&2
  echo "" >&2
  echo "Total new findings: ${new_count} (high-severity entries above block the build)" >&2
  exit 1
fi

echo "clang-tidy-diff: ${new_count} new findings, 0 high-severity — OK"
exit 0
