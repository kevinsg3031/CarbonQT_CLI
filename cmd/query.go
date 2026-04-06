package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/models"
	"carbonqt/internal/monitor"
	"carbonqt/internal/repo"
	"carbonqt/internal/ui"

	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [process_name]",
		Short: "Search running processes by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot := ""
			if repoOnly {
				root, err := repo.DetectRepoRoot()
				if err == nil {
					repoRoot = root
				}
			}

			procs, err := monitor.ListProcesses(repoRoot)
			if err != nil {
				return err
			}

			needle := strings.ToLower(args[0])
			filtered := make([]models.ProcessMetrics, 0)
			for _, proc := range procs {
				if strings.Contains(strings.ToLower(proc.Name), needle) {
					filtered = append(filtered, proc)
				}
			}

			estimator := energy.NewEstimator(cpuTDP, emissionFactor)
			_, filtered = estimator.ApplyCarbon(filtered, time.Second)
			sort.Slice(filtered, func(i, j int) bool { return filtered[i].CarbonKg > filtered[j].CarbonKg })

			if len(filtered) == 0 {
				fmt.Println("No matching processes found.")
				return nil
			}

			fmt.Println(ui.RenderProcessTable(filtered))
			return nil
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newQueryCmd())
}
