package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/internal/ai"
	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/internal/diff"
	"github.com/ashishsalunkhe/godeps-guard/internal/git"
	"github.com/ashishsalunkhe/godeps-guard/internal/graph"
	"github.com/ashishsalunkhe/godeps-guard/internal/license"
	"github.com/ashishsalunkhe/godeps-guard/internal/policy"
	"github.com/ashishsalunkhe/godeps-guard/internal/report"
	"github.com/spf13/cobra"
)

var (
	checkBase    string
	checkFormat  string
	checkConfig  string
	checkComment bool
	checkAI      bool
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

		// 3. Detect licenses for risk scoring
		licenses := license.DetectMap(headSnap.Modules)

		// 4. Diff Snapshots
		delta := diff.Compare(baseSnap, headSnap, licenses, cfg.Policies.HeavyVendorPatterns)

		// 5. Enforce Policy
		polRes := policy.Evaluate(delta, cfg, licenses)

		// 6. AI Enrichment (opt-in via --ai flag)
		if checkAI {
			aiClient, aiErr := ai.NewClientFromEnv()
			if aiErr != nil {
				fmt.Fprintf(os.Stderr, "⚠️  AI init failed: %v\n", aiErr)
			} else if aiClient.IsNoop() {
				fmt.Fprintf(os.Stderr, "⚠️  AI key missing; please set GODEPS_GUARD_AI_KEY in your .env or environment\n")
			} else {
				ctx := context.Background()

				// Feature 2: Smart risk scoring per direct dep
				if cfg.AI.Features.SmartRisk {
					for i := range delta.DirectImpacts {
						score, reasons, rErr := aiClient.EnhanceRisk(ctx, &delta.DirectImpacts[i])
						if rErr == nil {
							policy.MergeAIRisk(&delta.DirectImpacts[i], score, reasons)
						}
					}
				}

				// Feature 3: Validate dep reasons from PR diff
				if cfg.AI.Features.ValidateReason && cfg.Policies.RequireReasonForNewDirectDep {
					prDiff := getPRDiff(checkBase)
					for _, impact := range delta.DirectImpacts {
						reason := fmt.Sprintf("New dependency: %s", impact.Module.Path)
						findings, vErr := aiClient.ValidateReason(ctx, impact.Module.Path, reason, prDiff)
						if vErr == nil {
							polRes.AIWarnings = append(polRes.AIWarnings, findings...)
						}
					}
				}

				// Feature 1: AI PR summary
				if cfg.AI.Features.PRSummary && len(delta.AddedModules) > 0 {
					summary, sErr := aiClient.SummarizeDelta(ctx, delta, polRes)
					if sErr == nil {
						delta.AI.Summary = summary
					}
				}

				// Feature 4: Suggest alternatives for heavy deps
				if cfg.AI.Features.SuggestAlternatives && len(delta.DirectImpacts) > 0 {
					alts, aErr := aiClient.SuggestAlternatives(ctx, delta.DirectImpacts)
					if aErr == nil {
						delta.AI.Alternatives = alts
					}
				}
			}
		}

		// 7. Output Report
		if checkComment {
			err = report.OutputComment(delta, polRes, os.Stdout)
		} else {
			err = report.Output(delta, polRes, checkFormat, os.Stdout)
		}

		if err != nil {
			return fmt.Errorf("failed to render report: %w", err)
		}

		// 8. Fail CI if policy breached
		if !polRes.Passed {
			os.Exit(1)
		}

		return nil
	},
}

// getPRDiff returns the git diff between the base ref and HEAD for AI context.
func getPRDiff(base string) string {
	out, err := exec.Command("git", "diff", base+"...HEAD", "--", "go.mod", "go.sum").Output()
	if err != nil {
		return ""
	}
	diff := strings.TrimSpace(string(out))
	if len(diff) > 3000 {
		diff = diff[:3000] + "\n...[truncated]"
	}
	return diff
}

func init() {
	checkCmd.Flags().StringVar(&checkBase, "base", "", "Base git ref (e.g. origin/main)")
	checkCmd.Flags().StringVar(&checkConfig, "config", ".godepsguard.yaml", "Path to config file")
	checkCmd.Flags().StringVar(&checkFormat, "format", "terminal", "Output format (terminal, markdown, json)")
	checkCmd.Flags().BoolVar(&checkComment, "comment", false, "Output a condensed Markdown format ideal for PR comments")
	checkCmd.Flags().BoolVar(&checkAI, "ai", false, "Enable AI-powered analysis (requires GODEPS_GUARD_AI_KEY env var)")

	rootCmd.AddCommand(checkCmd)
}

