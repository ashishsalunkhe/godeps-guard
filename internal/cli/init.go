package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/internal/ai"
	"github.com/spf13/cobra"
)

var initAI bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .godepsguard.yaml config file",
	Long: `Create a .godepsguard.yaml in the current directory.
With --ai, describe your policy in plain English and let AI generate the YAML for you.
Without --ai, a default template is written.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath := ".godepsguard.yaml"

		// Check if file already exists
		if _, err := os.Stat(outPath); err == nil {
			fmt.Printf("⚠️  %s already exists. Overwrite? [y/N]: ", outPath)
			reader := bufio.NewReader(os.Stdin)
			ans, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(ans)), "y") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		var yamlContent string

		if initAI {
			// Initialize AI if requested
			aiClient, err := ai.NewClientFromEnv()
			if err != nil {
				return fmt.Errorf("AI init failed: %w", err)
			}
			if aiClient.IsNoop() {
				return fmt.Errorf("AI key missing; please set GODEPS_GUARD_AI_KEY in your .env or environment")
			}

			fmt.Println("Describe your dependency policy in plain English.")
			fmt.Println("Example: \"Block GPL licenses, warn if binary grows more than 5MB, flag AWS SDKs\"")
			fmt.Print("\n> ")

			reader := bufio.NewReader(os.Stdin)
			description, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(description) == "" {
				return fmt.Errorf("no description provided")
			}

			fmt.Println("\n⏳ Generating config with AI...")
			yamlContent, err = aiClient.GenerateConfig(context.Background(), strings.TrimSpace(description))
			if err != nil {
				return fmt.Errorf("AI config generation failed: %w", err)
			}
			if strings.TrimSpace(yamlContent) == "" {
				return fmt.Errorf("AI returned an empty config; please try again")
			}
		} else {
			yamlContent = defaultConfigTemplate
		}

		if err := os.WriteFile(outPath, []byte(yamlContent), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("✅ Config written to %s\n", outPath)
		return nil
	},
}

const defaultConfigTemplate = `build:
  target: ./...
  output: /tmp/godepsguard_app
  ldflags:
    - "-s"
    - "-w"

thresholds:
  max_direct_deps_added: 3
  max_transitive_packages_added: 50
  max_binary_size_increase_bytes: 5242880  # 5MB
  max_binary_size_increase_percent: 10
  max_risk_score: 7

policies:
  blocked_licenses:
    - GPL-3.0
    - AGPL-3.0
  warn_on_indirect_only_growth: true
  require_reason_for_new_direct_dep: false

report:
  format: markdown
  verbose: false

ai:
  provider: gemini
  features:
    pr_summary: true
    smart_risk: true
    validate_reason: true
    suggest_alternatives: true
    trend_analysis: true
`

func init() {
	initCmd.Flags().BoolVar(&initAI, "ai", false, "Generate config from a plain-English description (requires GODEPS_GUARD_AI_KEY)")
	rootCmd.AddCommand(initCmd)
}
