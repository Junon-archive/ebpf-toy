#!/usr/bin/env bash
set -euo pipefail

DUR="${1:-30}"   # seconds
OUTDIR="${2:-/tmp/ebpf-toy-workload}"
mkdir -p "$OUTDIR"

echo "[workload] duration=${DUR}s out=${OUTDIR}"
echo "[workload] starting CPU+IO+MEM loads..."

# 1) CPU load: spin loops (no dependencies)
CPU_JOBS="${CPU_JOBS:-4}"
pids=()
for i in $(seq 1 "$CPU_JOBS"); do
  ( while :; do :; done ) &
  pids+=("$!")
done
echo "[workload] cpu jobs=${CPU_JOBS}"

# 2) IO load: dd to file (best effort)
IO_FILE="${OUTDIR}/iolat_dd.bin"
( dd if=/dev/zero of="$IO_FILE" bs=4K count=200000 oflag=direct status=none || \
  dd if=/dev/zero of="$IO_FILE" bs=4K count=200000 status=none ) &
pids+=("$!")
echo "[workload] io job=dd (file=${IO_FILE})"

# 3) MEM load: reuse your existing PF workload if present
if [[ -x "./scripts/workload_pf.sh" ]]; then
  ( ./scripts/workload_pf.sh "$DUR" ) &
  pids+=("$!")
  echo "[workload] mem job=workload_pf.sh"
else
  # fallback: touch memory by allocating a big tmp file and reading it
  MEM_FILE="${OUTDIR}/mem_touch.bin"
  ( dd if=/dev/zero of="$MEM_FILE" bs=1M count=2048 status=none && \
    cat "$MEM_FILE" > /dev/null ) &
  pids+=("$!")
  echo "[workload] mem job=fallback (file=${MEM_FILE})"
fi

# run for duration
sleep "$DUR"

echo "[workload] stopping..."
for pid in "${pids[@]}"; do
  kill "$pid" 2>/dev/null || true
done
wait 2>/dev/null || true

sync || true
echo "[workload] done."
