package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	emissionFactor float64
	cpuTDP         float64
	repoOnly       bool
)

var rootCmd = &cobra.Command{
	Use:   "carbonqt",
	Short: "carbonqt monitors process energy and carbon emissions",
	Long:  "carbonqt is a cross-platform CLI for estimating process energy use and carbon emissions.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Float64Var(&emissionFactor, "emission-factor", 2e-10, "Emission factor in kg CO2 per Joule")
	rootCmd.PersistentFlags().Float64Var(&cpuTDP, "cpu-tdp", 65, "CPU TDP in Watts")
	rootCmd.PersistentFlags().BoolVar(&repoOnly, "repo-only", false, "Filter processes to the current repository")
}
