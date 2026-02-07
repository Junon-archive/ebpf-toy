package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

const maxSlots = 64

type Summary struct {
	Module          string `json:"module"`
	Metric          string `json:"metric"`
	Unit            string `json:"unit"`
	DurationSec     int    `json:"duration_sec"`
	TotalEvents     uint64 `json:"total_events"`
	TailThresholdUs uint64 `json:"tail_threshold_us"`
	TailEvents      uint64 `json:"tail_events"`
	MaxBucket       int    `json:"max_bucket"`
	Notes           string `json:"notes"`
}

func main() {
	var dur time.Duration
	var out string
	flag.DurationVar(&dur, "duration", 10*time.Second, "collection duration (e.g., 10s)")
	flag.StringVar(&out, "out", "", "output path prefix (optional). If set, writes CSV + summary.json")
	flag.Parse()

	fmt.Printf("[runqlat] starting (duration=%s, out=%q)\n", dur, out)

	spec, err := ebpf.LoadCollectionSpec("bpf/runqlat.bpf.o")
	if err != nil {
		fatalf("load spec: %v", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		fatalf("new collection: %v", err)
	}
	defer coll.Close()

	wakeupProg := coll.Programs["runqlat_wakeup"]
	switchProg := coll.Programs["runqlat_switch"]
	if wakeupProg == nil || switchProg == nil {
		fatalf("programs not found (need runqlat_wakeup/runqlat_switch)")
	}

	hist := coll.Maps["hist"]
	if hist == nil {
		fatalf("map 'hist' not found")
	}

	tpWake, err := link.Tracepoint("sched", "sched_wakeup", wakeupProg, nil)
	if err != nil {
		fatalf("attach tracepoint sched:sched_wakeup: %v", err)
	}
	defer tpWake.Close()

	tpSw, err := link.Tracepoint("sched", "sched_switch", switchProg, nil)
	if err != nil {
		fatalf("attach tracepoint sched:sched_switch: %v", err)
	}
	defer tpSw.Close()

	fmt.Printf("[runqlat] attached. collecting for %s...\n", dur)
	time.Sleep(dur)

	counts, total := readHist(hist)
	printHist(counts)

	tailThresholdUs := uint64(8) // same as Week1 simple tail definition
	tail := sumTail(counts, tailThresholdUs)

	maxB := maxBucket(counts)

	if out != "" {
		csvPath := normalizeCSVPath(out)
		summaryPath := strings.TrimSuffix(csvPath, ".csv") + ".summary.json"
		if err := os.MkdirAll(filepath.Dir(csvPath), 0o755); err != nil {
			fatalf("mkdir out dir: %v", err)
		}
		if err := saveCSV(csvPath, counts); err != nil {
			fatalf("save csv: %v", err)
		}
		sum := Summary{
			Module:          "runqlat",
			Metric:          "runqueue_wait_latency",
			Unit:            "microseconds",
			DurationSec:     int(dur.Seconds()),
			TotalEvents:     total,
			TailThresholdUs: tailThresholdUs,
			TailEvents:      tail,
			MaxBucket:       maxB,
			Notes:           "bucket i means [2^i, 2^(i+1)) microseconds",
		}
		if err := saveJSON(summaryPath, sum); err != nil {
			fatalf("save json: %v", err)
		}
		fmt.Printf("[runqlat] saved: %s (+ %s)\n", csvPath, summaryPath)
	}

	fmt.Println("[runqlat] done")
}

func readHist(hist *ebpf.Map) ([]uint64, uint64) {
	counts := make([]uint64, maxSlots)
	var total uint64
	for i := uint32(0); i < maxSlots; i++ {
		var perCPU []uint64
		if err := hist.Lookup(&i, &perCPU); err != nil {
			// bucket이 없거나(이론상 없음) 에러 처리
			continue
		}
		var cnt uint64
		for _, v := range perCPU {
			cnt += v
		}
		counts[i] = cnt
		total += cnt
	}
	return counts, total
}

func printHist(counts []uint64) {
	fmt.Println("bucket  range(us)            count")
	for i, c := range counts {
		// print more at low buckets; skip long zeros
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

func sumTail(counts []uint64, thresholdUs uint64) uint64 {
	// thresholdUs=8 => bucket>=3
	var startBucket int
	for startBucket = 0; startBucket < maxSlots; startBucket++ {
		if (uint64(1) << uint(startBucket)) >= thresholdUs {
			break
		}
	}
	var s uint64
	for i := startBucket; i < len(counts); i++ {
		s += counts[i]
	}
	return s
}

func maxBucket(counts []uint64) int {
	maxB := -1
	for i := len(counts) - 1; i >= 0; i-- {
		if counts[i] > 0 {
			maxB = i
			break
		}
	}
	return maxB
}

func normalizeCSVPath(out string) string {
	if out == "" {
		return ""
	}
	if strings.HasSuffix(out, ".csv") {
		return out
	}
	return out + ".csv"
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
