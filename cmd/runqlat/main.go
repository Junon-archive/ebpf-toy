package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

func main() {
	duration := flag.Duration("duration", 2*time.Second, "how long to run (e.g., 1s, 10s)")
	objPath := flag.String("obj", "bpf/runqlat.bpf.o", "path to BPF object file")
	flag.Parse()

	spec, err := ebpf.LoadCollectionSpec(*objPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load spec failed: %v\n", err)
		os.Exit(1)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new collection failed: %v\n", err)
		os.Exit(1)
	}
	defer coll.Close()

	prog := coll.Programs["tp_sched_switch"]
	if prog == nil {
		fmt.Fprintf(os.Stderr, "program tp_sched_switch not found in obj\n")
		os.Exit(1)
	}

	lnk, err := link.Tracepoint("sched", "sched_switch", prog, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attach tracepoint failed: %v\n", err)
		os.Exit(1)
	}
	defer lnk.Close()

	fmt.Printf("[runqlat-hello] attached. collecting for %s...\n", duration.String())
	time.Sleep(*duration)

	m := coll.Maps["counter"]
	if m == nil {
		fmt.Fprintf(os.Stderr, "map counter not found in obj\n")
		os.Exit(1)
	}

	var key uint32 = 0
	var val uint64
	if err := m.Lookup(&key, &val); err != nil {
		fmt.Fprintf(os.Stderr, "map lookup failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[runqlat-hello] sched_switch count = %d\n", val)
}
