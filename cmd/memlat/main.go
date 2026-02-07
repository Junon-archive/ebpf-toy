package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

const maxSlots = 64

type Summary struct {
	DurationSec int    `json:"duration_sec"`
	Total       uint64 `json:"total"`
	Note        string `json:"note"`
}

func main() {
	var dur time.Duration
	var out string
	flag.DurationVar(&dur, "duration", 10*time.Second, "collection duration (e.g., 10s)")
	flag.StringVar(&out, "out", "", "output path (optional). If set, writes CSV + summary.json")
	flag.Parse()

	fmt.Printf("[memlat] starting (duration=%s, out=%q)\n", dur, out)

	spec, err := ebpf.LoadCollectionSpec("bpf/memlat.bpf.o")
	if err != nil {
		fatalf("load spec: %v", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		fatalf("new collection: %v", err)
	}
	defer coll.Close()

	entry := coll.Programs["memlat_entry"]
	exit := coll.Programs["memlat_exit"]
	if entry == nil || exit == nil {
		fatalf("programs not found (need memlat_entry/memlat_exit)")
	}

	hist := coll.Maps["hist"]
	if hist == nil {
		fatalf("map 'hist' not found")
	}

	// Attach entry/exit probes to the same function.
	// If this fails due to missing symbol, we'll swap to another symbol later.
	kp, err := link.Kprobe("handle_mm_fault", entry, nil)
	if err != nil {
		fatalf("attach kprobe(handle_mm_fault): %v", err)
	}
	defer kp.Close()

	krp, err := link.Kretprobe("handle_mm_fault", exit, nil)
	if err != nil {
		fatalf("attach kretprobe(handle_mm_fault): %v", err)
	}
	defer krp.Close()

	fmt.Printf("[memlat] attached. collecting for %s...\n", dur)
	time.Sleep(dur)

	counts, total := readHist(hist)
	printHist(counts)

	if out != "" {
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			fatalf("mkdir out dir: %v", err)
		}
		if err := saveCSV(out, counts); err != nil {
			fatalf("save csv: %v", err)
		}
		sum := Summary{
			DurationSec: int(dur.Seconds()),
			Total:       total,
			Note:        "bucket i means [2^i, 2^(i+1)) microseconds",
		}
		if err := saveJSON(out+".summary.json", sum); err != nil {
			fatalf("save json: %v", err)
		}
		fmt.Printf("[memlat] saved: %s (+ .summary.json)\n", out)
	}

	fmt.Println("[memlat] done")
}

func readHist(hist *ebpf.Map) ([]uint64, uint64) {
	counts := make([]uint64, maxSlots)
	var total uint64
	for i := uint32(0); i < maxSlots; i++ {
		var v uint64
		if err := hist.Lookup(&i, &v); err != nil {
			fatalf("hist lookup bucket=%d: %v", i, err)
		}
		counts[i] = v
		total += v
	}
	return counts, total
}

func printHist(counts []uint64) {
	fmt.Println("bucket  range(us)            count")
	for i, c := range counts {
		// 너무 뒤는 기본적으로 생략(원하면 조건 제거)
		if c == 0 && i > 25 {
			continue
		}
		lo := uint64(1) << uint(i)
		hi := uint64(1) << uint(i+1)
		fmt.Printf("%2d      [%8d, %8d)   %d\n", i, lo, hi, c)
	}
}

func saveCSV(path string, counts []uint64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "bucket,lo_us,hi_us,count")
	for i, c := range counts {
		lo := uint64(1) << uint(i)
		hi := uint64(1) << uint(i+1)
		fmt.Fprintf(f, "%d,%d,%d,%d\n", i, lo, hi, c)
	}
	return nil
}

func saveJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
