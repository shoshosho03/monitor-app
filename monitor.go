package main

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// 純粋なロジック（再利用可能）
func GetSystemStats() (float64, float64, error) {

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, 0, err
	}

	return vmStat.UsedPercent, cpuPercent[0], nil
}
