package ai

import (
	"context"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// NoopClient is used when no API key is configured.
// All methods return zero values without error so the tool degrades gracefully.
type NoopClient struct{}

func (n *NoopClient) SummarizeDelta(_ context.Context, _ *types.Delta, _ *types.PolicyResult) (string, error) {
	return "", nil
}

func (n *NoopClient) EnhanceRisk(_ context.Context, _ *types.ModuleImpact) (int, []string, error) {
	return 0, nil, nil
}

func (n *NoopClient) ValidateReason(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}

func (n *NoopClient) SuggestAlternatives(_ context.Context, _ []types.ModuleImpact) (map[string]string, error) {
	return nil, nil
}

func (n *NoopClient) AnalyzeTrend(_ context.Context, _ []types.HistoryRecord) (string, error) {
	return "", nil
}

func (n *NoopClient) GenerateConfig(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (n *NoopClient) IsNoop() bool {
	return true
}
