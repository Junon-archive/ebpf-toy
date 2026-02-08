# Week 2 Summary — CPU(runqlat) + I/O(iolat) + One-Click Collection

Date: 2026-02-07  
Repo: `~/eBPF/ebpf-toy`  
Goal: Build the “top 3 bottlenecks” tools with a consistent UX/output format and collect them in one shot.

---

## What I built

### 1) `memlat` — Memory access latency histogram
- Measures memory access latency (microseconds) as a log2 histogram.
- Output: `memlat.csv` + `memlat.csv.summary.json`

### 2) `runqlat` — CPU run-queue latency histogram
- Measures run-queue wait time:
  - `sched_wakeup`: task becomes runnable (enters run queue)
  - `sched_switch`: task actually starts running
  - Delta = run-queue latency
- Output: `runqlat.csv` + `runqlat.csv.summary.json`
- Note: file currently saved as `runqlat.csv.csv` due to an output naming bug (planned fix).

### 3) `iolat` — Block I/O request latency histogram
- Measures block I/O latency:
  - `block_rq_issue`: request issued to device
  - `block_rq_complete`: request completed
  - Delta = I/O latency
- Key design: use `struct request *` as a stable request identity (pointer value).
- Output: `iolat.csv` + `iolat.summary.json`

---

## One-click collection (Week 2 MVP)

A single script collects all three modules with the same duration and saves results into one folder.

Command:
```bash
./scripts/collect_all.sh --duration 30s --out results/week2_one_click
