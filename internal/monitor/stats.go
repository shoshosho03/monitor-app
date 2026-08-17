// Package monitor provides system resource usage statistics.
package monitor

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// GetSystemStats returns memory and CPU usage percentages.
func GetSystemStats() (float64, float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, fmt.Errorf("get memory usage: %w", err)
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, 0, fmt.Errorf("get CPU usage: %w", err)
	}
	if len(cpuPercent) == 0 {
		return 0, 0, fmt.Errorf("get CPU usage: no values returned")
	}

	return vmStat.UsedPercent, cpuPercent[0], nil
}
