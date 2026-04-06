package report

import (
	"context"
	"errors"
	"sort"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/models"
	"carbonqt/internal/monitor"
)

type ReportConfig struct {
	Estimator energy.Estimator
	RepoRoot  string
	Duration  time.Duration
	Interval  time.Duration
}

type ReportData struct {
	GeneratedAt   time.Time
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Interval      time.Duration
	RepoRoot      string
	Estimator     energy.Estimator
	System        models.SystemMetrics
	Processes     []models.ProcessMetrics
	TotalCarbonKg float64
	Trend         []float64
}

type aggregate struct {
	metrics models.ProcessMetrics
	cpuSum  float64
	samples int
}

func CollectReport(ctx context.Context, cfg ReportConfig) (ReportData, error) {
	if cfg.Duration <= 0 {
		return ReportData{}, errors.New("duration must be greater than zero")
	}
	if cfg.Interval <= 0 {
		return ReportData{}, errors.New("interval must be greater than zero")
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, cfg.Duration)
	defer cancel()

	agg := make(map[int32]*aggregate)
	trend := make([]float64, 0, int(cfg.Duration/cfg.Interval)+1)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			system, err := monitor.GetSystemMetrics()
			if err != nil {
				return ReportData{}, err
			}

			results := make([]models.ProcessMetrics, 0, len(agg))
			total := 0.0
			for _, item := range agg {
				if item.samples == 0 {
					continue
				}
				avgCPU := item.cpuSum / float64(item.samples)
				item.metrics.CPUPercent = avgCPU
				item.metrics.PowerW = cfg.Estimator.PowerW(avgCPU)
				item.metrics.CarbonKg = cfg.Estimator.CarbonKg(avgCPU, cfg.Duration)
				total += item.metrics.CarbonKg
				results = append(results, item.metrics)
			}
			sort.Slice(results, func(i, j int) bool { return results[i].CarbonKg > results[j].CarbonKg })

			return ReportData{
				GeneratedAt:   time.Now(),
				StartTime:     start,
				EndTime:       time.Now(),
				Duration:      cfg.Duration,
				Interval:      cfg.Interval,
				RepoRoot:      cfg.RepoRoot,
				Estimator:     cfg.Estimator,
				System:        system,
				Processes:     results,
				TotalCarbonKg: total,
				Trend:         trend,
			}, nil
		case <-ticker.C:
			procs, err := monitor.ListProcesses(cfg.RepoRoot)
			if err != nil {
				return ReportData{}, err
			}
			stepTotal := 0.0
			for _, proc := range procs {
				item, ok := agg[proc.PID]
				if !ok {
					item = &aggregate{}
					agg[proc.PID] = item
				}
				item.metrics = proc
				item.cpuSum += proc.CPUPercent
				item.samples++
				stepTotal += cfg.Estimator.CarbonKg(proc.CPUPercent, cfg.Interval)
			}
			trend = append(trend, stepTotal)
		}
	}
}
