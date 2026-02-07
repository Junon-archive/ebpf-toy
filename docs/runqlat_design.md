# CPU Run Queue Latency (runqlat) — Week 2 Design & Validation

This document describes the design and validation plan of the `runqlat` module.
`runqlat` measures **how long a thread waits in the CPU run queue** before it
actually starts running on a CPU.

This is written for interviewers and engineers evaluating practical
observability and measurement design skills.

---

## Measurement Goal

Measure scheduler wait latency:

Δt = t_run - t_ready

cat > docs/runqlat_design.md << 'EOF'
# CPU Run Queue Latency (runqlat) — Week 2 Design & Validation

This document describes the design and validation plan of the `runqlat` module.
`runqlat` measures **how long a thread waits in the CPU run queue** before it
actually starts running on a CPU.

This is written for interviewers and engineers evaluating practical
observability and measurement design skills.

---

## Measurement Goal

Measure scheduler wait latency:

Δt = t_run - t_ready


Rationale:
- Using PID alone can cause collisions in multi-threaded workloads.
- `pid_tgid` uniquely identifies the runnable entity.

---

## Map Design

Two maps, same pattern as Week 1 memlat.

### 1) start_ts (HASH)
Stores ready timestamps.

| Field | Description |
|------|-------------|
| Key  | pid_tgid |
| Value | t_ready (nanoseconds) |

Rules:
- written on wakeup
- read + deleted on switch (to avoid stale entries)

### 2) hist (ARRAY)
Latency histogram.

| Field | Description |
|------|-------------|
| Key  | bucket index |
| Value | count |

Bucketization:

delta_us = (now_ns - start_ns) / 1000
bucket = log2(delta_us)
hist[bucket]++

---

## Kernel-Space Logic (eBPF)

### Wakeup probe (entry)
1) key = pid_tgid
2) start_ts[key] = now_ns

### Switch probe (exit)
1) key = pid_tgid(next task)
2) if start exists:
   - delta = now - start
   - bucketize and hist++
   - delete start_ts[key]
3) if start missing:
   - skip (valid case: task was switched in without a recorded wakeup)

---

## User-Space Responsibilities (Go)

The Go program:
- loads eBPF object
- attaches tracepoints
- sleeps for duration
- reads hist map and prints histogram
- saves:
  - `<module>.csv`
  - `<module>.summary.json`

Output schema is unified across memlat/runqlat/iolat:

- CSV: `bucket,lo_us,hi_us,count`
- Summary JSON includes `module, metric, duration_sec, total_events, tail_events, max_bucket`

---

## Validation Plan

We validate that the metric reacts to CPU pressure.

Two runs:
- OFF: idle (no load)
- ON : CPU load enabled (e.g., `stress-ng --cpu N`)

We compare tail events, e.g.:

tail = Δt >= 8us (bucket >= 3)

Expected:
- Under CPU load, run queue wait tail increases.

---

## Key Takeaways

- `sched_wakeup → sched_switch(next)` is the canonical pattern to measure run queue latency.
- `pid_tgid` is essential for correctness under multi-threading.
- Histogram tail is the most meaningful signal for “CPU feels laggy” scenarios.

---
