package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/models"
	"carbonqt/internal/monitor"
	"carbonqt/internal/repo"
	"carbonqt/internal/ui"

	"github.com/spf13/cobra"
)

type runAggregate struct {
	metrics models.ProcessMetrics
	cpuSum  float64
	samples int
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [duration]",
		Short: "Monitor processes for a duration and report carbon emissions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration := 5 * time.Second
			if len(args) > 0 {
				parsed, err := time.ParseDuration(args[0])
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				duration = parsed
			}

			repoRoot := ""
			if repoOnly {
				root, err := repo.DetectRepoRoot()
				if err == nil {
					repoRoot = root
				}
			}

			estimator := energy.NewEstimator(cpuTDP, emissionFactor)
			if duration < time.Second {
				procs, err := monitor.ListProcesses(repoRoot)
				if err != nil {
					return err
				}
				_, procs = estimator.ApplyCarbon(procs, duration)
				sort.Slice(procs, func(i, j int) bool { return procs[i].CarbonKg > procs[j].CarbonKg })
				fmt.Println(ui.RenderProcessTable(procs))
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			agg := make(map[int32]*runAggregate)
			trend := make([]float64, 0, int(duration.Seconds()))
			for {
				select {
				case <-ctx.Done():
					results := make([]models.ProcessMetrics, 0, len(agg))
					for _, item := range agg {
						if item.samples == 0 {
							continue
						}
						avgCPU := item.cpuSum / float64(item.samples)
						item.metrics.CPUPercent = avgCPU
						item.metrics.PowerW = estimator.PowerW(avgCPU)
						item.metrics.CarbonKg = estimator.CarbonKg(avgCPU, duration)
						results = append(results, item.metrics)
					}
					sort.Slice(results, func(i, j int) bool { return results[i].CarbonKg > results[j].CarbonKg })
					fmt.Println(ui.RenderProcessTable(results))
					if len(trend) > 0 {
						fmt.Println()
						fmt.Println(ui.RenderCarbonTrend(trend))
					}
					return nil
				case <-ticker.C:
					procs, err := monitor.ListProcesses(repoRoot)
					if err != nil {
						return err
					}
					stepTotal := 0.0
					for _, proc := range procs {
						item, ok := agg[proc.PID]
						if !ok {
							item = &runAggregate{}
							agg[proc.PID] = item
						}
						item.metrics = proc
						item.cpuSum += proc.CPUPercent
						item.samples++
						stepTotal += estimator.CarbonKg(proc.CPUPercent, time.Second)
					}
					trend = append(trend, stepTotal)
				}
			}
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newRunCmd())
}
