#!/usr/bin/env bash
set -euo pipefail

DUR="${1:-30s}"
BASE="${2:-results/week3_runs}"
mkdir -p "$BASE"

for i in 1 2 3; do
  OUT="${BASE}/run${i}"
  TS="$(date +%y%m%d-%H%M%S)"
  OUT="${OUT}/${TS}"
  mkdir -p "$OUT"

  echo "=============================="
  echo "[week3] run${i} out=${OUT}"
  echo "=============================="

  ./scripts/save_meta.sh "$OUT"

  # workload + collection
  # - workload runs in background for duration (seconds)
  dur_s="${DUR%s}"
  ./scripts/run_workload.sh "$dur_s" "$OUT/workload" &
  WL_PID=$!

  ./scripts/collect_all.sh --duration "$DUR" --out "$OUT"

  wait "$WL_PID" || true

  echo "[week3] done run${i}: $OUT"
done
