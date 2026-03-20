// Package ai provides an abstraction layer over LLM providers for godeps-guard.
// All features are opt-in and the tool remains fully functional without any API key.
package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Client is the interface implemented by all AI provider backends and the NoopClient.
type Client interface {
	// SummarizeDelta converts a dependency delta + policy result into a plain-English summary.
	SummarizeDelta(ctx context.Context, delta *types.Delta, policy *types.PolicyResult) (string, error)

	// EnhanceRisk takes a module impact and returns an AI-enhanced risk score + reasons.
	// The returned score should be treated as advisory; callers may cap at 10.
	EnhanceRisk(ctx context.Context, impact *types.ModuleImpact) (score int, reasons []string, err error)

	// ValidateReason checks whether the stated justification for a new dependency makes
	// sense given the module name and the PR diff context.
	ValidateReason(ctx context.Context, modulePath, reason, prDiff string) (findings []string, err error)

	// SuggestAlternatives returns a map of modulePath → lighter alternative suggestion
	// for each impact that appears heavy.
	SuggestAlternatives(ctx context.Context, impacts []types.ModuleImpact) (map[string]string, error)

	// AnalyzeTrend takes historical records and returns a narrative anomaly/trend report.
	AnalyzeTrend(ctx context.Context, records []types.HistoryRecord) (string, error)

	// GenerateConfig takes a plain-English policy description and returns a .godepsguard.yaml string.
	GenerateConfig(ctx context.Context, description string) (string, error)

	// IsNoop returns true if this is a no-op client (no API key configured).
	IsNoop() bool
}

// NewClient creates the appropriate Client based on the provider name, API key, and model.
// If apiKey is empty, a NoopClient is returned so the tool degrades gracefully.
func NewClient(provider, apiKey, model string) (Client, error) {
	if apiKey == "" {
		return &NoopClient{}, nil
	}
	switch provider {
	case "gemini", "":
		return newGeminiClient(apiKey, model)
	case "openai":
		return newOpenAIClient(apiKey, model)
	default:
		return nil, fmt.Errorf("unknown AI provider %q; supported: gemini, openai", provider)
	}
}

// NewClientFromEnv reads GODEPS_GUARD_AI_PROVIDER and GODEPS_GUARD_AI_KEY from the
// environment and returns a ready-to-use Client (or NoopClient if no key is set).
func NewClientFromEnv() (Client, error) {
	provider := os.Getenv("GODEPS_GUARD_AI_PROVIDER")
	key := os.Getenv("GODEPS_GUARD_AI_KEY")
	model := os.Getenv("GODEPS_GUARD_AI_MODEL")
	return NewClient(provider, key, model)
}
