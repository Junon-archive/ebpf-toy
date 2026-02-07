package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const maxSlots = 26

type Summary struct {
	Tool        string  `json:"tool"`
	DurationSec float64 `json:"duration_sec"`
	Total       uint64  `json:"total"`
	GeneratedAt string  `json:"generated_at"`
	Out         string  `json:"out,omitempty"`
}

func normalizeOut(out string) (csvPath string, summaryPath string) {
	if out == "" {
		return "", ""
	}
	// IMPORTANT: don't create .csv.csv
	if strings.HasSuffix(out, ".csv") {
		csvPath = out
	} else {
		csvPath = out + ".csv"
	}
	summaryPath = strings.TrimSuffix(csvPath, ".csv") + ".summary.json"
	return
}

func bucketRangeUs(b int) (lo, hi uint64) {
	// bucket 0: [1,2), bucket 1: [2,4), ...
	lo = 1 << uint(b)
	hi = 1 << uint(b+1)
	return
}

func writeCSV(path string, counts []uint64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "bucket,lo_us,hi_us,count")
	for i := 0; i < len(counts); i++ {
		lo, hi := bucketRangeUs(i)
		fmt.Fprintf(f, "%d,%d,%d,%d\n", i, lo, hi, counts[i])
	}
	return nil
}

func writeSummary(path string, s Summary) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func main() {
	var (
		durationStr string
		out         string
	)
	flag.StringVar(&durationStr, "duration", "2s", "collection duration (e.g. 2s, 10s)")
	flag.StringVar(&out, "out", "", "output csv path (e.g. results/week2_iolat_smoke.csv)")
	flag.Parse()

	dur, err := time.ParseDuration(durationStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --duration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[iolat] starting (duration=%s, out=%q)\n", dur, out)

	if err := rlimit.RemoveMemlock(); err != nil {
		fmt.Fprintf(os.Stderr, "rlimit: %v\n", err)
		os.Exit(1)
	}

	// ===== Load eBPF objects =====
	var objs iolatObjects
	if err := loadIolatObjects(&objs, nil); err != nil {
		fmt.Fprintf(os.Stderr, "load objects: %v\n", err)
		os.Exit(1)
	}
	defer objs.Close()

	// ===== Attach tp_btf programs =====
	lnkIssue, err := link.AttachTracing(link.TracingOptions{
		Program: objs.IolatIssue,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "attach issue: %v\n", err)
		os.Exit(1)
	}
	defer lnkIssue.Close()

	lnkComplete, err := link.AttachTracing(link.TracingOptions{
		Program: objs.IolatComplete,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "attach complete: %v\n", err)
		os.Exit(1)
	}
	defer lnkComplete.Close()

	fmt.Printf("[iolat] attached. collecting for %s...\n", dur)

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	<-ctx.Done()

	// ===== Read histogram map =====
	counts := make([]uint64, maxSlots)
	for i := 0; i < maxSlots; i++ {
		key := uint32(i)
		var val uint64
		if err := objs.Hist.Lookup(&key, &val); err != nil {
			fmt.Fprintf(os.Stderr, "map lookup hist[%d]: %v\n", i, err)
			os.Exit(1)
		}
		counts[i] = val
	}

	// print table
	fmt.Printf("bucket  range(us)            count\n")
	var total uint64
	for i := 0; i < maxSlots; i++ {
		lo, hi := bucketRangeUs(i)
		fmt.Printf("%2d      [%8d, %8d)   %d\n", i, lo, hi, counts[i])
		total += counts[i]
	}

	// save outputs
	csvPath, summaryPath := normalizeOut(out)
	if csvPath != "" {
		if err := writeCSV(csvPath, counts); err != nil {
			fmt.Fprintf(os.Stderr, "write csv: %v\n", err)
			os.Exit(1)
		}
		s := Summary{
			Tool:        "iolat",
			DurationSec: dur.Seconds(),
			Total:       total,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Out:         csvPath,
		}
		if err := writeSummary(summaryPath, s); err != nil {
			fmt.Fprintf(os.Stderr, "write summary: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[iolat] saved: %s (+ %s)\n", csvPath, summaryPath)
	}

	fmt.Printf("[iolat] done\n")
}
