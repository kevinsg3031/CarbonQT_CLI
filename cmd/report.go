package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/repo"
	"carbonqt/internal/report"

	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var (
		format  string
		output  string
		title   string
		topRows int
	)

	cmd := &cobra.Command{
		Use:   "report [duration]",
		Short: "Generate CSV and PDF reports from collected metrics",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration := 30 * time.Second
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
			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()

			data, err := report.CollectReport(ctx, report.ReportConfig{
				Estimator: estimator,
				RepoRoot:  repoRoot,
				Duration:  duration,
				Interval:  1 * time.Second,
			})
			if err != nil {
				return err
			}

			normalized := strings.ToLower(strings.TrimSpace(format))
			if normalized == "" {
				normalized = "both"
			}
			if normalized != "csv" && normalized != "pdf" && normalized != "both" {
				return fmt.Errorf("invalid format: %s (use csv, pdf, or both)", format)
			}

			if output == "" {
				stamp := time.Now().Format("20060102-150405")
				output = fmt.Sprintf("./reports/report-%s", stamp)
			}

			written := make([]string, 0, 4)
			if normalized == "csv" || normalized == "both" {
				csvPaths, err := report.WriteCSV(output, data)
				if err != nil {
					return err
				}
				written = append(written, csvPaths.SummaryPath, csvPaths.ProcessesPath, csvPaths.TrendPath)
			}
			if normalized == "pdf" || normalized == "both" {
				pdfPath, err := report.WritePDF(output, data, report.PDFOptions{
					Title:        title,
					MaxTableRows: topRows,
					TopProcesses: topRows,
				})
				if err != nil {
					return err
				}
				written = append(written, pdfPath)
			}

			sort.Strings(written)
			fmt.Println("Report files:")
			for _, path := range written {
				fmt.Printf("- %s\n", path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "both", "Report format: csv, pdf, or both")
	cmd.Flags().StringVar(&output, "output", "", "Base output path (default ./reports/report-YYYYMMDD-HHMMSS)")
	cmd.Flags().StringVar(&title, "title", "CarbonQT Report", "Title for the PDF report")
	cmd.Flags().IntVar(&topRows, "top", 15, "Top processes to include in the PDF charts and table")

	return cmd
}

func init() {
	rootCmd.AddCommand(newReportCmd())
}
