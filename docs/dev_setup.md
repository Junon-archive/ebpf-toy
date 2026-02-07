# Development Setup

This document describes the reproducible development environment and the initial validation steps for the eBPF toy project.

The goal of this setup is to ensure that:
- the kernel environment is compatible with eBPF,
- the toolchain is correctly installed,
- and a minimal eBPF program can be built, attached, and observed from user space.

---

## System Information

The following system information is recorded to ensure kernel-level compatibility
with eBPF programs and tracepoints.

```bash
uname -a
Linux junon-B650M-K 6.8.0-90-generic #91~22.04.1-Ubuntu SMP PREEMPT_DYNAMIC Thu Nov 20 15:20:45 UTC 2024 x86_64 x86_64 x86_64 GNU/Linux

uname -r
6.8.0-90-generic

lsb_release -a || cat /etc/os-release
No LSB modules are available.
Distributor ID: Ubuntu
Description:    Ubuntu 22.04.5 LTS
Release:        22.04
Codename:       jammy
```

---
## Toolchain Overview

The project uses the following toolchain:
- clang / llvm
    Used to compile eBPF C programs into BPF bytecode (-target bpf).
- libbpf headers
    Required for eBPF helper APIs and map/program definitions.
- Go (user-space)
    Used to load, attach, and read data from eBPF programs via cilium/ebpf.

This reflects the standard eBPF architecture:
- kernel-space: high-performance event tracing (eBPF C)
- user-space: control logic and data aggregation (Go)

---
## 0-Week Validation: Tracepoint Counter (Hello eBPF)
### Goal
Verify the complete eBPF workflow:

### build → load → attach → collect → read → output
using the simplest possible example.

---
### Step 1. Build eBPF Object

Compile the eBPF program that attaches to the
sched:sched_switch tracepoint and increments a counter on each event.
```bash
make bpf-runqlat
```

Expected build command:
```bash
clang -O2 -g -target bpf -D__TARGET_ARCH_x86 -I./bpf \
  -c bpf/runqlat.bpf.c -o bpf/runqlat.bpf.o
```

---
### Step 2. Build User-space Binaries

Build Go-based user-space controllers.
``` bash
make build
```

Expected output:
``` bash
mkdir -p bin
go build -o bin/memlat ./cmd/memlat
go build -o bin/runqlat ./cmd/runqlat
go build -o bin/iolat ./cmd/iolat
```

---
### Step 3. Run Hello Example

Attach the eBPF program and collect events for a short duration.
``` bash
sudo ./bin/runqlat --duration 2s
```

Example output:
``` bash
[runqlat-hello] attached. collecting for 2s...
[runqlat-hello] sched_switch count = 9428
```

---
### Interpretation

This confirms that:
- the eBPF program was successfully loaded into the kernel,
- the tracepoint attachment worked as expected,
- kernel events were counted in an eBPF map,
- and the user-space program correctly read and printed the data.
This milestone validates the fundamental capability to observe kernel-level behavior using eBPF.

---