# Memory Latency Measurement (memlat) — Week 1 Design & Validation

This document describes the design, implementation, and validation of the
`memlat` module, which measures memory-related latency using eBPF.

The purpose of this module is to:
- learn and solidify the **entry/exit latency measurement pattern**,
- build a latency histogram that reflects real kernel behavior,
- and validate the metric through controlled workload on/off experiments.

This document is intended for readers with systems or kernel interest,
including interviewers evaluating practical observability skills.

---

## Measurement Goal

Measure the execution latency of a kernel memory-fault handling path.

The latency is defined as:

Δt = t_exit - t_entry

where:
- `t_entry` is the timestamp when the kernel function is entered,
- `t_exit`  is the timestamp when the function returns.

The resulting Δt is aggregated into a histogram to observe
both common-case performance and rare high-latency tail behavior.

---

## Probe Selection

### Target Function
handle_mm_fault()

This function is part of the Linux memory management subsystem and is
invoked when a process accesses a virtual memory address that requires
kernel intervention (e.g., page faults).

### Probe Types
- **kprobe**     → captures function entry
- **kretprobe**  → captures function return

This pairing enables precise measurement of function execution time
without modifying kernel code.

---

## Key Design

### Why pid_tgid?

Each latency measurement must correctly match an entry event
with its corresponding exit event.

The key used is:
pid_tgid = (PID << 32) | TID

Rationale:
- Multiple threads within the same process may fault concurrently.
- Using PID alone would cause collisions and incorrect matching.
- `pid_tgid` uniquely identifies a thread-level execution context.

This ensures correctness even under multi-threaded workloads.

---

## Map Design

Two eBPF maps are used.

### 1. start_ts (HASH)
Stores entry timestamps.

| Field | Description |
|------|-------------|
| Key  | pid_tgid |
| Value | entry timestamp (nanoseconds) |

Usage:
- Written at function entry
- Read and deleted at function exit

This prevents stale state accumulation and ensures bounded memory usage.

---

### 2. hist (ARRAY)
Aggregates latency distribution.

| Field | Description |
|------|-------------|
| Key  | bucket index (log2 scale) |
| Value | count |

Bucket index is computed as:
bucket = log2(delta_us)

This logarithmic scale emphasizes tail behavior while remaining compact.

---

## Kernel-Space Logic (eBPF)

High-level logic:

1. **Entry probe**
   - read current timestamp
   - store it in `start_ts[pid_tgid]`

2. **Exit probe**
   - look up entry timestamp
   - compute `delta = now - start`
   - convert to microseconds
   - bucketize using `log2`
   - increment histogram
   - delete `start_ts[pid_tgid]`

The eBPF program performs only minimal computation:
no formatting, no aggregation beyond histogram increment.

---

## User-Space Responsibilities (Go)

The Go user-space program is responsible for:
- loading the eBPF object,
- attaching kprobe/kretprobe,
- controlling collection duration,
- reading histogram data from maps,
- printing human-readable output,
- optionally saving results to disk.

This separation keeps kernel logic simple and safe,
while allowing flexible presentation and analysis in user space.

---

## Validation Strategy

### Why Validation Is Necessary

Latency histograms can look “reasonable” even when the metric is wrong.
Therefore, explicit validation is required to prove that:

> the measured metric reacts meaningfully to system stress.

---

### Validation Workload

A synthetic workload was used to stress the memory subsystem by
continuously touching a large virtual memory region.

This increases memory fault activity and exercises the measured path.

---

### Experimental Setup

Two runs were compared:

- **OFF**: idle system, no workload
- **ON** : memory page-touch workload enabled

Each run collected data for 10 seconds.

---

## Results Summary

Tail latency was defined as:
Δt >= 8 microseconds (bucket >= 3)

Observed counts:

| Case | Tail Events |
|-----|-------------|
| OFF | 136 |
| ON  | 186 |

Additionally, under workload, rare high-latency buckets
(e.g., 128–256 microseconds) appeared more frequently.

---

## Interpretation

- The histogram responds clearly to memory-related workload.
- While most events remain fast, stress increases the frequency of
  slow-path executions.
- Tail behavior, rather than average latency, provides the most
  meaningful signal for diagnosing latency-sensitive issues.

This validates that the metric reflects real kernel behavior
and is suitable for further analysis.

---

## Key Takeaways

- Entry/exit probe pairing is a robust pattern for latency measurement.
- Thread-level keys (`pid_tgid`) are essential for correctness.
- Histogram-based analysis exposes tail latency that averages hide.
- Validation through workload comparison is critical for trustworthiness.

---

## Next Steps

- Apply the same design pattern to:
  - CPU scheduling latency (`runqlat`)
  - I/O request latency (`iolat`)
- Unify output formats and collection workflows across modules.
- Extend analysis to correlate latency with process context.

---
