package cli

import (
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/spf13/cobra"
)

var scanOutput string

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Capture dependency snapshot for current checkout",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config or use default
		cfg, _ := config.Load(".godepsguard.yaml")

		dir := "." // scan current dir
		
		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		out := os.Stdout
		if scanOutput != "" {
			f, err := os.Create(scanOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			out = f
		}

		if err := graph.WriteJSON(snap, out); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}

		if scanOutput != "" {
			fmt.Printf("Snapshot written to %s\\n", scanOutput)
		}

		return nil
	},
}

func init() {
	scanCmd.Flags().StringVar(&scanOutput, "output", "", "Output file for snapshot (e.g., snapshot.json). Prints to stdout if omitted.")
}
