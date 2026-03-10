package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/diff"
	"github.com/ashishsalunkhe/godeps-guard/internal/report"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
	"github.com/spf13/cobra"
)

var (
	diffBase   string
	diffHead   string
	diffFormat string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		if diffBase == "" || diffHead == "" {
			return fmt.Errorf("--base and --head flags are required")
		}

		baseSnap, err := loadSnapshot(diffBase)
		if err != nil {
			return fmt.Errorf("failed to load base snapshot: %w", err)
		}

		headSnap, err := loadSnapshot(diffHead)
		if err != nil {
			return fmt.Errorf("failed to load head snapshot: %w", err)
		}

		delta := diff.Compare(baseSnap, headSnap)

		return report.Output(delta, nil, diffFormat, os.Stdout)
	},
}

func loadSnapshot(path string) (*types.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap types.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func init() {
	diffCmd.Flags().StringVar(&diffBase, "base", "", "Path to base snapshot.json")
	diffCmd.Flags().StringVar(&diffHead, "head", "", "Path to head snapshot.json")
	diffCmd.Flags().StringVar(&diffFormat, "format", "markdown", "Output format (markdown, json)")
	
	rootCmd.AddCommand(diffCmd)
}
