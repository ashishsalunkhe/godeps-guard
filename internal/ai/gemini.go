package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

const defaultGeminiModel = "gemini-2.0-flash"

type geminiClient struct {
	model *genai.GenerativeModel
	gc    *genai.Client
}

func newGeminiClient(apiKey, model string) (Client, error) {
	if model == "" {
		model = defaultGeminiModel
	}
	ctx := context.Background()
	gc, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	m := gc.GenerativeModel(model)
	m.SetTemperature(0.3) // low temperature for consistent, factual output
	return &geminiClient{model: m, gc: gc}, nil
}

func (g *geminiClient) generate(ctx context.Context, prompt string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}
	part := resp.Candidates[0].Content.Parts[0]
	if txt, ok := part.(genai.Text); ok {
		return string(txt), nil
	}
	return "", fmt.Errorf("unexpected response part type from Gemini")
}

func (g *geminiClient) SummarizeDelta(ctx context.Context, delta *types.Delta, policy *types.PolicyResult) (string, error) {
	data, _ := json.MarshalIndent(map[string]any{"delta": delta, "policy": policy}, "", "  ")
	return g.generate(ctx, fmt.Sprintf(promptSummarize, string(data)))
}

func (g *geminiClient) EnhanceRisk(ctx context.Context, impact *types.ModuleImpact) (int, []string, error) {
	prompt := fmt.Sprintf(promptEnhanceRisk,
		impact.Module.Path,
		impact.Module.Version,
		impact.AddedPackages,
		len(impact.TransitiveModules),
		impact.RiskScore,
		impact.RiskReasons,
	)
	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return 0, nil, err
	}
	return parseRiskResponse(raw)
}

func (g *geminiClient) ValidateReason(ctx context.Context, modulePath, reason, prDiff string) ([]string, error) {
	truncDiff := prDiff
	if len(truncDiff) > 2000 {
		truncDiff = truncDiff[:2000] + "\n...[truncated]"
	}
	raw, err := g.generate(ctx, fmt.Sprintf(promptValidateReason, modulePath, reason, truncDiff))
	if err != nil {
		return nil, err
	}
	return parseStringSlice(raw)
}

func (g *geminiClient) SuggestAlternatives(ctx context.Context, impacts []types.ModuleImpact) (map[string]string, error) {
	data, _ := json.MarshalIndent(impacts, "", "  ")
	raw, err := g.generate(ctx, fmt.Sprintf(promptSuggestAlternatives, string(data)))
	if err != nil {
		return nil, err
	}
	return parseStringMap(raw)
}

func (g *geminiClient) AnalyzeTrend(ctx context.Context, records []types.HistoryRecord) (string, error) {
	data, _ := json.MarshalIndent(records, "", "  ")
	return g.generate(ctx, fmt.Sprintf(promptAnalyzeTrend, string(data)))
}

func (g *geminiClient) GenerateConfig(ctx context.Context, description string) (string, error) {
	return g.generate(ctx, fmt.Sprintf(promptGenerateConfig, description))
}

func (g *geminiClient) IsNoop() bool {
	return false
}
