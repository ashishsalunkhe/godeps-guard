package cli

import (
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/internal/license"
	"github.com/ashishsalunkhe/godeps-guard/internal/sbom"
	"github.com/spf13/cobra"
)

var sbomOutput string

var sbomCmd = &cobra.Command{
	Use:   "sbom",
	Short: "Generate Software Bill of Materials (CycloneDX)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(".godepsguard.yaml")

		dir := "."
		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		// Detect licenses to enrich SBOM
		licenses := license.DetectMap(snap.Modules)

		out := os.Stdout
		if sbomOutput != "" {
			f, err := os.Create(sbomOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			out = f
		}

		if err := sbom.CycloneDX(snap, licenses, out); err != nil {
			return fmt.Errorf("failed to encode SBOM: %w", err)
		}

		if sbomOutput != "" {
			fmt.Printf("SBOM written to %s\n", sbomOutput)
		}

		return nil
	},
}

func init() {
	sbomCmd.Flags().StringVar(&sbomOutput, "output", "", "Output file for SBOM (e.g., sbom.json). Prints to stdout if omitted.")
	rootCmd.AddCommand(sbomCmd)
}
