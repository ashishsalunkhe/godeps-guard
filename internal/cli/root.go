package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "godeps-guard",
	Short:   "A dependency impact analyzer to prevent binary bloat",
	Long:    "godeps-guard inspects dependency changes, measures binary bloat, and fails CI if configured thresholds are exceeded.",
	Version: "dev",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(scanCmd)
	// difference, check, report commands will be added in subsequent milestones
}
