package energy

import (
	"time"

	"carbonqt/internal/models"
)

type Estimator struct {
	CPUWatts       float64
	EmissionFactor float64
}

func NewEstimator(cpuTDP, emissionFactor float64) Estimator {
	return Estimator{CPUWatts: cpuTDP, EmissionFactor: emissionFactor}
}

func (e Estimator) EnergyJ(cpuPercent float64, duration time.Duration) float64 {
	return (cpuPercent / 100.0) * e.CPUWatts * duration.Seconds()
}

func (e Estimator) PowerW(cpuPercent float64) float64 {
	return (cpuPercent / 100.0) * e.CPUWatts
}

func (e Estimator) CarbonKg(cpuPercent float64, duration time.Duration) float64 {
	return e.EnergyJ(cpuPercent, duration) * e.EmissionFactor
}

func (e Estimator) ApplyCarbon(processes []models.ProcessMetrics, duration time.Duration) (float64, []models.ProcessMetrics) {
	total := 0.0
	for i := range processes {
		processes[i].CarbonKg = e.CarbonKg(processes[i].CPUPercent, duration)
		processes[i].PowerW = e.PowerW(processes[i].CPUPercent)
		total += processes[i].CarbonKg
	}

	return total, processes
}
