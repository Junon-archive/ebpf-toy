#!/usr/bin/env bash
set -euo pipefail
DUR="${1:-10}"

# 메모리를 크게 잡고 페이지를 계속 터치해서 fault/메모리 경로를 자극
python3 - <<PY
import time, mmap, os
dur = int("${DUR}")
size = 1024 * 1024 * 1024   # 1GB
page = 4096

mm = mmap.mmap(-1, size, flags=mmap.MAP_PRIVATE | mmap.MAP_ANONYMOUS)
t0 = time.time()
i = 0
while time.time() - t0 < dur:
    off = (i * 7919) % (size - page)
    mm[off:off+1] = b'\x01'
    i += 1
print("touched pages:", i)
PY
