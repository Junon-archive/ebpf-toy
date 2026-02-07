package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	duration := flag.Duration("duration", 1*time.Second, "how long to run (e.g., 1s, 10s)")
	out := flag.String("out", "", "output directory (optional)")
	flag.Parse()

	fmt.Printf("[iolat] starting (duration=%s, out=%q)\n", duration.String(), *out)
	time.Sleep(*duration)
	fmt.Println("[iolat] done (stub). Next: attach block tracepoints + read histogram.")
}
