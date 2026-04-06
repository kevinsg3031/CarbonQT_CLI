package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ReportCSVPaths struct {
	SummaryPath   string
	ProcessesPath string
	TrendPath     string
}

func WriteCSV(basePath string, data ReportData) (ReportCSVPaths, error) {
	base := strings.TrimSpace(basePath)
	if base == "" {
		return ReportCSVPaths{}, fmt.Errorf("output path is required")
	}
	if strings.HasSuffix(strings.ToLower(base), ".csv") {
		base = strings.TrimSuffix(base, filepath.Ext(base))
	}

	summaryPath := base + "-summary.csv"
	processesPath := base + "-processes.csv"
	trendPath := base + "-trend.csv"

	if err := writeSummaryCSV(summaryPath, data); err != nil {
		return ReportCSVPaths{}, err
	}
	if err := writeProcessesCSV(processesPath, data); err != nil {
		return ReportCSVPaths{}, err
	}
	if err := writeTrendCSV(trendPath, data); err != nil {
		return ReportCSVPaths{}, err
	}

	return ReportCSVPaths{
		SummaryPath:   summaryPath,
		ProcessesPath: processesPath,
		TrendPath:     trendPath,
	}, nil
}

func writeSummaryCSV(path string, data ReportData) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rows := [][]string{
		{"generated_at", data.GeneratedAt.Format(time.RFC3339)},
		{"duration", data.Duration.String()},
		{"interval", data.Interval.String()},
		{"repo_root", data.RepoRoot},
		{"total_carbon_kg", fmt.Sprintf("%.8f", data.TotalCarbonKg)},
		{"total_carbon_mg", formatCarbonMg(data.TotalCarbonKg)},
		{"cpu_model", data.System.CPUModel},
		{"cpu_cores", fmt.Sprintf("%d", data.System.CPUCores)},
		{"cpu_usage_percent", fmt.Sprintf("%.2f", data.System.CPUPercent)},
		{"memory_usage_percent", fmt.Sprintf("%.2f", data.System.MemoryPercent)},
		{"memory_used", formatBytes(data.System.MemoryUsedBytes)},
		{"memory_total", formatBytes(data.System.MemoryTotalBytes)},
		{"platform", data.System.Platform},
		{"uptime", formatDuration(time.Duration(data.System.UptimeSeconds) * time.Second)},
		{"cpu_tdp_watts", fmt.Sprintf("%.2f", data.Estimator.CPUWatts)},
		{"emission_factor", fmt.Sprintf("%.8g", data.Estimator.EmissionFactor)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

func writeProcessesCSV(path string, data ReportData) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"pid", "name", "cpu_percent", "mem_percent", "power_w", "carbon_kg", "carbon_mg", "runtime", "exe_path"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, proc := range data.Processes {
		runtime := ""
		if !proc.StartTime.IsZero() {
			runtime = formatDuration(time.Since(proc.StartTime))
		}
		row := []string{
			fmt.Sprintf("%d", proc.PID),
			proc.Name,
			fmt.Sprintf("%.2f", proc.CPUPercent),
			fmt.Sprintf("%.2f", proc.MemoryPercent),
			fmt.Sprintf("%.2f", proc.PowerW),
			fmt.Sprintf("%.8f", proc.CarbonKg),
			formatCarbonMg(proc.CarbonKg),
			runtime,
			proc.ExePath,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

func writeTrendCSV(path string, data ReportData) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"second", "carbon_kg", "carbon_mg"}); err != nil {
		return err
	}

	for i, value := range data.Trend {
		row := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%.8f", value),
			formatCarbonMg(value),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
