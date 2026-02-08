![figure](ebpf-toy/assets/figure.png)

# ebpf-toy — Practical Latency Observability with eBPF

This repository is a hands-on eBPF project focused on **latency observability**
inside the Linux kernel.

It implements three latency probes that share the same UX and output format:

- **memlat**  : memory access latency
- **runqlat** : CPU run-queue waiting latency
- **iolat**   : block I/O request latency

The goal is not just measurement, but **reproducible analysis and interpretation**
that resembles real-world performance debugging.

---

## Features

- Kernel-level latency measurement using eBPF
- Unified histogram-based output (CSV + JSON)
- One-click collection for all probes
- Designed for repeatable experiments and reporting

---

## Tools Overview

### memlat
Measures memory access latency and page-fault–related delays.

### runqlat
Measures how long tasks wait in the CPU run queue
(from wakeup to actual execution).

### iolat
Measures block I/O request latency
(from request issue to completion).

All tools:
- Use log2 histogram buckets (microseconds)
- Output CSV + summary JSON
- Follow the same command-line interface

---

## One-Click Collection

Run all probes with the same duration and store results together:
``` bash
./scripts/collect_all.sh --duration 30s --out results/week2_one_click
```
This generates a timestamped directory containing:
- memlat.csv
- runqlat.csv
- iolat.csv
- *.summary.json files for each tool

## Directory Structure
- `bpf/` : eBPF programs
- `cmd/` : Go user-space loaders
- `scripts/` : workload and collection scripts
- `docs/` : design documents
- `results/` : measurement outputs
---

## Design Philosophy
- Same UX for all probes
- Simple, inspectable data formats
- Focus on tail latency, not averages
- Easy to reproduce and explain

This project is intentionally small and explicit,  
to demonstrate practical eBPF usage rather than abstract frameworks.
