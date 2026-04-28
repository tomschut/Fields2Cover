#!/usr/bin/env bash
# Phase 7 PRF-01 — Plan 07-07 reproducer.
#
# Runs all Fields2Cover perf benchmarks (swath / headland / path interp /
# OR-Tools TSP) and captures:
#   - CSV output from each harness into $OUT_DIR/<name>.csv
#   - perf flamegraphs into $OUT_DIR/flamegraphs/<name>.svg, IF `perf record`
#     is usable on this host AND $FLAMEGRAPH points at the FlameGraph repo.
#
# If perf is unavailable (e.g. kernel.perf_event_paranoid >= 2, no CAP_SYS_ADMIN,
# no FlameGraph install) the script still emits the CSV files — those are the
# mandatory output. Flamegraphs are best-effort.
#
# Usage:
#   BUILD_DIR=build ./scripts/run-benchmarks.sh
#   OUT_DIR=docs/perf FLAMEGRAPH=/opt/FlameGraph ./scripts/run-benchmarks.sh

set -euo pipefail

BUILD_DIR="${BUILD_DIR:-build}"
OUT_DIR="${OUT_DIR:-docs/perf}"
FG_DIR="$OUT_DIR/flamegraphs"
mkdir -p "$FG_DIR"

BENCHES=(
  "bench_swath:tests/cpp/benchmarks/bench_swath"
  "bench_headland:tests/cpp/benchmarks/bench_headland"
  "bench_path_interp:tests/cpp/benchmarks/bench_path_interp"
  "ortools_benchmark:tests/cpp/route_planning/benchmarks/route_planner_benchmark"
)

PERF_OK=0
if command -v perf >/dev/null 2>&1; then
  # Dry-run a tiny perf record to detect paranoid-level gating.
  if perf record -F 99 -o /tmp/.f2c-perf-probe.data -- /bin/true >/dev/null 2>&1 \
      && [ -s /tmp/.f2c-perf-probe.data ]; then
    PERF_OK=1
  fi
  rm -f /tmp/.f2c-perf-probe.data
fi

for entry in "${BENCHES[@]}"; do
  name="${entry%%:*}"
  bin="$BUILD_DIR/${entry##*:}"
  if [[ ! -x "$bin" ]]; then
    echo "skip: $name (missing $bin)"
    continue
  fi
  echo "=== $name ==="
  "$bin" | tee "$OUT_DIR/${name}.csv"

  if [[ "$PERF_OK" -eq 1 ]]; then
    perf record -F 99 -g --call-graph=dwarf -o "/tmp/perf-$name.data" -- "$bin" || true
    if [[ -s "/tmp/perf-$name.data" && -n "${FLAMEGRAPH:-}" ]]; then
      perf script -i "/tmp/perf-$name.data" \
        | "$FLAMEGRAPH/stackcollapse-perf.pl" \
        | "$FLAMEGRAPH/flamegraph.pl" > "$FG_DIR/${name}.svg" || true
      echo "  flamegraph -> $FG_DIR/${name}.svg"
    fi
  else
    echo "  (perf unavailable — flamegraph skipped)"
  fi
done

echo
echo "Done. CSV outputs in $OUT_DIR/*.csv"
