package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/internal/license"
	"github.com/spf13/cobra"
)

var (
	licensesFormat      string
	licensesFailUnknown bool
)

var licensesCmd = &cobra.Command{
	Use:   "licenses",
	Short: "Detect licenses of all dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(".godepsguard.yaml")

		dir := "."
		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		results := license.Detect(snap.Modules)

		switch licensesFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(results); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		default:
			fmt.Printf("%-16s  %-55s  %s\n", "LICENSE", "MODULE", "VERSION")
			fmt.Printf("%-16s  %-55s  %s\n", "-------", "------", "-------")

			for _, r := range results {
				fmt.Printf("%-16s  %-55s  %s\n", r.License, r.ModulePath, r.Version)
			}
		}

		if licensesFailUnknown {
			for _, r := range results {
				if r.License == "Unknown" {
					fmt.Fprintf(os.Stderr, "\nError: undetected licenses found. Use --fail-on-unknown=false to ignore.\n")
					os.Exit(1)
				}
			}
		}

		return nil
	},
}

func init() {
	licensesCmd.Flags().StringVar(&licensesFormat, "format", "table", "Output format (table, json)")
	licensesCmd.Flags().BoolVar(&licensesFailUnknown, "fail-on-unknown", false, "Exit with code 1 if any license is undetected")
	rootCmd.AddCommand(licensesCmd)
}
