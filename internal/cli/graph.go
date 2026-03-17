package cli

import (
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/spf13/cobra"
)

var (
	graphOutput string
	graphFormat string
	graphFilter string
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate dependency graph visualization",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(".godepsguard.yaml")

		dir := "."
		snap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		out := os.Stdout
		if graphOutput != "" {
			f, err := os.Create(graphOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			out = f
		}

		switch graphFormat {
		case "mermaid":
			graph.RenderMermaid(snap, graphFilter, out)
		default:
			graph.RenderDOT(snap, graphFilter, out)
		}

		if graphOutput != "" {
			switch graphFormat {
			case "mermaid":
				fmt.Printf("Mermaid graph written to %s\n", graphOutput)
			default:
				fmt.Printf("Graphviz DOT written to %s. Render with: dot -Tsvg %s -o graph.svg\n", graphOutput, graphOutput)
			}
		}

		return nil
	},
}

func init() {
	graphCmd.Flags().StringVar(&graphOutput, "output", "", "Output file (e.g., graph.dot or graph.md)")
	graphCmd.Flags().StringVar(&graphFormat, "format", "dot", "Output format (dot, mermaid)")
	graphCmd.Flags().StringVar(&graphFilter, "filter", "", "Filter to modules matching this prefix (e.g., github.com/aws)")
	rootCmd.AddCommand(graphCmd)
}
