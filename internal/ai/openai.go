package ai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

const defaultOpenAIModel = "gpt-4o-mini"

type openaiClient struct {
	client *openai.Client
	model  string
}

func newOpenAIClient(apiKey, model string) (Client, error) {
	if model == "" {
		model = defaultOpenAIModel
	}
	c := openai.NewClient(apiKey)
	return &openaiClient{client: c, model: model}, nil
}

func (o *openaiClient) generate(ctx context.Context, prompt string) (string, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("openai generate: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

func (o *openaiClient) SummarizeDelta(ctx context.Context, delta *types.Delta, policy *types.PolicyResult) (string, error) {
	data, _ := json.MarshalIndent(map[string]any{"delta": delta, "policy": policy}, "", "  ")
	return o.generate(ctx, fmt.Sprintf(promptSummarize, string(data)))
}

func (o *openaiClient) EnhanceRisk(ctx context.Context, impact *types.ModuleImpact) (int, []string, error) {
	prompt := fmt.Sprintf(promptEnhanceRisk,
		impact.Module.Path,
		impact.Module.Version,
		impact.AddedPackages,
		len(impact.TransitiveModules),
		impact.RiskScore,
		impact.RiskReasons,
	)
	raw, err := o.generate(ctx, prompt)
	if err != nil {
		return 0, nil, err
	}
	return parseRiskResponse(raw)
}

func (o *openaiClient) ValidateReason(ctx context.Context, modulePath, reason, prDiff string) ([]string, error) {
	truncDiff := prDiff
	if len(truncDiff) > 2000 {
		truncDiff = truncDiff[:2000] + "\n...[truncated]"
	}
	raw, err := o.generate(ctx, fmt.Sprintf(promptValidateReason, modulePath, reason, truncDiff))
	if err != nil {
		return nil, err
	}
	return parseStringSlice(raw)
}

func (o *openaiClient) SuggestAlternatives(ctx context.Context, impacts []types.ModuleImpact) (map[string]string, error) {
	data, _ := json.MarshalIndent(impacts, "", "  ")
	raw, err := o.generate(ctx, fmt.Sprintf(promptSuggestAlternatives, string(data)))
	if err != nil {
		return nil, err
	}
	return parseStringMap(raw)
}

func (o *openaiClient) AnalyzeTrend(ctx context.Context, records []types.HistoryRecord) (string, error) {
	data, _ := json.MarshalIndent(records, "", "  ")
	return o.generate(ctx, fmt.Sprintf(promptAnalyzeTrend, string(data)))
}

func (o *openaiClient) GenerateConfig(ctx context.Context, description string) (string, error) {
	return o.generate(ctx, fmt.Sprintf(promptGenerateConfig, description))
}

func (o *openaiClient) IsNoop() bool {
	return false
}
