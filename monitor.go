package main

import "monitor-app/internal/monitor"

// GetSystemStats keeps the desktop application's existing API while sharing
// the implementation with the CLI application.
func GetSystemStats() (float64, float64, error) {
	return monitor.GetSystemStats()
}
