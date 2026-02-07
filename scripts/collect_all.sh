#!/usr/bin/env bash
set -euo pipefail

DURATION="30s"
OUTDIR="results/week2_one_click"
TAG="$(date +%y%m%d-%H%M%S)"

usage() {
  echo "Usage: $0 [--duration 30s] [--out results/week2_one_click]"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --duration) DURATION="$2"; shift 2 ;;
    --out) OUTDIR="$2"; shift 2 ;;
    -h|--help) usage ;;
    *) echo "Unknown arg: $1"; usage ;;
  esac
done

RUN_DIR="${OUTDIR}/${TAG}"
mkdir -p "${RUN_DIR}"

echo "[collect_all] duration=${DURATION}"
echo "[collect_all] out=${RUN_DIR}"

# 1) memlat
echo "[collect_all] memlat..."
sudo ./bin/memlat --duration "${DURATION}" --out "${RUN_DIR}/memlat.csv"

# 2) runqlat
echo "[collect_all] runqlat..."
sudo ./bin/runqlat --duration "${DURATION}" --out "${RUN_DIR}/runqlat.csv"

# 3) iolat
echo "[collect_all] iolat..."
sudo ./bin/iolat --duration "${DURATION}" --out "${RUN_DIR}/iolat.csv"

echo "[collect_all] done: ${RUN_DIR}"
ls -lh "${RUN_DIR}"
