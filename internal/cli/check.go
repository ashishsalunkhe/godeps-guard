package cli

import (
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/diff"
	"github.com/ashishsalunkhe/godeps-guard/internal/git"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/internal/policy"
	"github.com/ashishsalunkhe/godeps-guard/internal/report"
	"github.com/spf13/cobra"
)

var (
	checkBase   string
	checkFormat string
	checkConfig string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "End-to-end CI command",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkBase == "" {
			return fmt.Errorf("--base flag is required")
		}

		cfg, err := config.Load(checkConfig)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Target current directory
		dir := "."

		// 1. Snapshot HEAD
		headSnap, err := graph.GenerateSnapshot(dir, cfg.Build.Target, cfg.Build.Output, cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to snapshot head: %w", err)
		}

		// 2. Snapshot Base (using a temporary git worktree)
		baseWorktreeDir, cleanup, err := git.CreateTempWorktree(dir, checkBase)
		if err != nil {
			return fmt.Errorf("failed to create base worktree: %w", err)
		}
		defer cleanup()

		baseSnap, err := graph.GenerateSnapshot(baseWorktreeDir, cfg.Build.Target, cfg.Build.Output+"_base", cfg.Build.Ldflags)
		if err != nil {
			return fmt.Errorf("failed to snapshot base: %w", err)
		}

		// 3. Diff Snapshots
		delta := diff.Compare(baseSnap, headSnap)

		// 4. Enforce Policy
		polRes := policy.Evaluate(delta, cfg)

		// 5. Output Report
		err = report.Output(delta, polRes, checkFormat, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to render report: %w", err)
		}

		// 6. Fail CI if policy breached
		if !polRes.Passed {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkBase, "base", "", "Base git ref (e.g. origin/main)")
	checkCmd.Flags().StringVar(&checkConfig, "config", ".godepsguard.yaml", "Path to config file")
	checkCmd.Flags().StringVar(&checkFormat, "format", "markdown", "Output format (markdown, json)")
	
	rootCmd.AddCommand(checkCmd)
}
