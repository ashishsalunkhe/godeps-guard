package cli

import (
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/spf13/cobra"
)

var graphOutput string

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate dependency visualization (Graphviz DOT format)",
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

		fmt.Fprintln(out, "digraph godeps {")
		fmt.Fprintln(out, "  node [shape=box, style=rounded];")

		fmt.Fprintf(out, "  \"%s\";\n", "app")

		for _, m := range snap.Modules {
			if !m.Indirect {
				fmt.Fprintf(out, "  \"%s\" -> \"%s\";\n", "app", m.Path)
			}
		}

		// In a fully developed visualization, you would iterate the actual package import graph
		// and graph every individual node here. For MVP, we map the direct module graph.

		fmt.Fprintln(out, "}")

		if graphOutput != "" {
			fmt.Printf("Graphviz DOT written to %s. Render with: dot -Tsvg %s -o graph.svg\n", graphOutput, graphOutput)
		}

		return nil
	},
}

func init() {
	graphCmd.Flags().StringVar(&graphOutput, "output", "", "Output file for dot format (e.g., graph.dot).")
	rootCmd.AddCommand(graphCmd)
}
