#!/usr/bin/env bash
set -euo pipefail
OUTDIR="${1:?usage: save_meta.sh <outdir>}"
mkdir -p "$OUTDIR"

{
  echo "timestamp=$(date -Is)"
  echo "hostname=$(hostname)"
  echo "kernel=$(uname -a)"
  echo "uname_r=$(uname -r)"
  echo "os=$(cat /etc/os-release 2>/dev/null | tr '\n' ' ' || true)"
  echo "cpu_model=$(lscpu | awk -F: '/Model name/ {gsub(/^[ \t]+/, "", $2); print $2; exit}')"
  echo "cpu_cores=$(nproc)"
  echo "mem_total=$(awk '/MemTotal/ {print $2" "$3}' /proc/meminfo)"
  echo "git_rev=$(git rev-parse --short HEAD 2>/dev/null || echo NA)"
} > "${OUTDIR}/meta.env"
