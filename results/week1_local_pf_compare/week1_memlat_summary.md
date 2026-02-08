# Week1 memlat summary

## Measurement
- Metric: Δt = t_exit - t_entry (microseconds)
- Probe: kprobe/kretprobe on handle_mm_fault()
- Key: pid_tgid (thread-level matching)
- Maps:
  - start_ts: HASH[pid_tgid] = entry timestamp (ns)
  - hist: ARRAY[log2(delta_us)] = count

## Experiment (validation)
- OFF: idle (no workload)
- ON : page-touch workload (workload_pf.sh)

## Observation
- Tail events (>= 8us, bucket >= 3) increased:
  - OFF: 136
  - ON : 186
- Under workload, rare but noticeable high-latency buckets (e.g., 128~256us) appeared.

## Interpretation
- The metric reacts to memory-path stress, supporting that the histogram reflects real latency changes.
- Tail behavior is a useful signal for diagnosing latency-sensitive memory bottlenecks.
