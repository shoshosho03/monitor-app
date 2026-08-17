// Command monitor-cli periodically prints CPU and memory usage.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitor-app/internal/monitor"
)

func main() {
	interval := flag.Duration("interval", time.Second, "monitoring interval (for example: 500ms, 2s, 1m)")
	flag.Parse()

	if *interval <= 0 {
		fmt.Fprintln(os.Stderr, "interval must be greater than 0")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	fmt.Printf("Monitoring every %s. Press Ctrl+C to stop.\n", *interval)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Monitoring stopped.")
			return
		case <-ticker.C:
			memoryPercent, cpuPercent, err := monitor.GetSystemStats()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get system stats: %v\n", err)
				continue
			}

			fmt.Printf("Memory: %6.2f%%  CPU: %6.2f%%\n", memoryPercent, cpuPercent)
		}
	}
}
