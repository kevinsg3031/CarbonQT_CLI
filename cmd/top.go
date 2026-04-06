package cmd

import (
	"fmt"
	"sort"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/monitor"
	"carbonqt/internal/repo"
	"carbonqt/internal/ui"

	"github.com/spf13/cobra"
)

func newTopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Display processes sorted by highest carbon emissions",
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

			estimator := energy.NewEstimator(cpuTDP, emissionFactor)
			_, procs = estimator.ApplyCarbon(procs, time.Second)
			sort.Slice(procs, func(i, j int) bool { return procs[i].CarbonKg > procs[j].CarbonKg })
			if len(procs) > 20 {
				procs = procs[:20]
			}

			fmt.Println(ui.RenderProcessTable(procs))
			return nil
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newTopCmd())
}
