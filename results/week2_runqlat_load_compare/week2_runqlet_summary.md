# Week 2 Summary — CPU Run Queue Latency (runqlat)

## 1. What Is Measured (Metric Definition)
- **Target**: CPU run queue waiting latency
- **Definition**:  
  Δt = time between `sched_wakeup` and `sched_switch(next_pid)`
- This measures how long a runnable task waits before actually getting CPU time.

---

## 2. Why This Metric (Design Rationale)
- CPU performance issues often appear first as **queueing delay**, not execution time.
- User-perceived stalls and jitter are dominated by **tail latency**, not averages.
- Measuring the gap between wakeup and actual execution directly captures this effect.

---

## 3. Implementation Overview
- **Entry event**: `sched:sched_wakeup`
  - `start_ts[pid] = now`
- **Exit event**: `sched:sched_switch`
  - `delta = now - start_ts[next_pid]`
  - `bucket = log2(delta_us)` → histogram update
- **Key**: PID
- **Maps**
  - `start_ts`: hash (pid → timestamp)
  - `hist`: array (log2 latency buckets)

---

## 4. Experiment Setup
- Comparison:
  - **OFF**: no artificial CPU load
  - **ON**: CPU-intensive workload applied
- Same collection duration for both cases
- Output format unified with memlat:
  - CSV + summary.json

---

## 5. Observations

### Tail Latency (≥ 8192 µs)
- OFF: **2166**
- ON : **4613**
- Increase: **2.13×**

### Tail Ratio (tail / total)
- OFF: 2166 / 14844 ≈ **14.6%**
- ON : 4613 / 21405 ≈ **21.5%**

---

## 6. Interpretation
- Under CPU load, long run queue waits increase significantly.
- The growth is more pronounced in the **tail**, not in the average.
- This explains intermittent stalls and responsiveness degradation observed by users.

---

## 7. Implications and Next Steps
- `runqlat` provides a quantitative view of CPU contention.
- Tail latency is a more reliable indicator of real performance degradation than mean latency.
- Next steps:
  - Apply the same measurement pattern to **I/O latency (iolat)**
  - Correlate CPU, memory, and I/O latency to narrow down bottleneck sources
