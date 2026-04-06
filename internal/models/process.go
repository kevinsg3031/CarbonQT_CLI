package models

import "time"

type ProcessMetrics struct {
	PID             int32
	Name            string
	CPUPercent      float64
	MemoryPercent   float64
	ContextSwitches *uint64
	Cwd             string
	ExePath         string
	StartTime       time.Time
	PowerW          float64
	CarbonKg        float64
}

type SystemMetrics struct {
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsedBytes  uint64
	CPUModel         string
	CPUCores         int
	MemoryTotalBytes uint64
	Platform         string
	UptimeSeconds    uint64
}
