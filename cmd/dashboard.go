package cmd

import (
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/repo"
	"carbonqt/internal/ui"

	"github.com/spf13/cobra"
)

func newDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Launch the interactive carbonqt dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot := ""
			if repoOnly {
				root, err := repo.DetectRepoRoot()
				if err == nil {
					repoRoot = root
				}
			}

			cfg := ui.DashboardConfig{
				Estimator: energy.NewEstimator(cpuTDP, emissionFactor),
				RepoRoot:  repoRoot,
				Refresh:   1 * time.Second,
			}
			return ui.StartDashboard(cfg)
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newDashboardCmd())
}
