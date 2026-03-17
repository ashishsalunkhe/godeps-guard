package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/internal/license"
	"github.com/ashishsalunkhe/godeps-guard/internal/util"
	"github.com/spf13/cobra"
)

var (
	licensesFormat      string
	licensesFailUnknown bool
	licensesShort       bool
)

func shortPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

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
			moduleCol := "MODULE"
			if licensesShort {
				moduleCol = "MOD (SHORT)"
			}
			header := fmt.Sprintf("%-16s  %-55s  %s", "LICENSE", moduleCol, "VERSION")
			fmt.Println(util.Bold + header + util.Reset)
			fmt.Printf("%-16s  %-55s  %s\n", "-------", "------", "-------")

			for _, r := range results {
				licStr := r.License
				if r.License == "Unknown" {
					licStr = util.Colorize(r.License, util.Yellow)
				} else if r.License == "GPL-3.0" || r.License == "AGPL-3.0" || r.License == "GPL-2.0" {
					licStr = util.Colorize(r.License, util.Red)
				} else {
					licStr = util.Colorize(r.License, util.Green)
				}
				
				// Manual padding for colorized string to maintain alignment
				padding := ""
				if len(r.License) < 16 {
					padding = strings.Repeat(" ", 16-len(r.License))
				}
				
				modPath := r.ModulePath
				if licensesShort {
					modPath = shortPath(modPath)
				}

				fmt.Printf("%s%s  %-55s  %s\n", licStr, padding, modPath, r.Version)
			}
		}

		if licensesFailUnknown {
			for _, r := range results {
				if r.License == "Unknown" {
					fmt.Fprintf(os.Stderr, "\n"+util.Error("Error: undetected licenses found. Use --fail-on-unknown=false to ignore.")+"\n")
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
	licensesCmd.Flags().BoolVar(&licensesShort, "short", false, "Shorten module paths in table output")
	rootCmd.AddCommand(licensesCmd)
}
