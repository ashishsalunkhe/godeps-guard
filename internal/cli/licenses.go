package cli

import (
	"fmt"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/spf13/cobra"
)

var licensesCmd = &cobra.Command{
	Use:   "licenses",
	Short: "Detect licenses of all dependencies (Heuristic MVP)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(".godepsguard.yaml")

		dir := "."
		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		fmt.Println("Dependency Licenses (Placeholder MVP)")
		fmt.Println("-----------------------------------")

		for _, m := range snap.Modules {
			if !m.Indirect {
				// In a full implementation, we would regex search GOMODCACHE or use go-license-detector
				// For the MVP, we print the module path demonstrating the CLI structure
				fmt.Printf("Unknown  %s\\n", m.Path)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(licensesCmd)
}
