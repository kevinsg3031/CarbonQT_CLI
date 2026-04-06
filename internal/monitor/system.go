package monitor

import (
	"carbonqt/internal/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func GetSystemMetrics() (models.SystemMetrics, error) {
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return models.SystemMetrics{}, err
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return models.SystemMetrics{}, err
	}

	info, err := host.Info()
	if err != nil {
		return models.SystemMetrics{}, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return models.SystemMetrics{}, err
	}

	cores, err := cpu.Counts(false)
	if err != nil {
		cores = 0
	}

	cpuPercent := 0.0
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	cpuModel := ""
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return models.SystemMetrics{
		CPUPercent:       cpuPercent,
		MemoryPercent:    memStats.UsedPercent,
		MemoryUsedBytes:  memStats.Used,
		CPUModel:         cpuModel,
		CPUCores:         cores,
		MemoryTotalBytes: memStats.Total,
		Platform:         info.Platform,
		UptimeSeconds:    info.Uptime,
	}, nil
}
