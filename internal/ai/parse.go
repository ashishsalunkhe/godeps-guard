package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

// parseRiskResponse extracts {"score": N, "reasons": [...]} from LLM output,
// tolerating markdown code fences and minor formatting variations.
func parseRiskResponse(raw string) (int, []string, error) {
	raw = extractJSON(raw)
	var result struct {
		Score   int      `json:"score"`
		Reasons []string `json:"reasons"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return 0, nil, nil // soft failure — fall back to static score
	}
	if result.Score < 1 {
		result.Score = 1
	}
	if result.Score > 10 {
		result.Score = 10
	}
	return result.Score, result.Reasons, nil
}

// parseStringSlice extracts a JSON string array from LLM output.
func parseStringSlice(raw string) ([]string, error) {
	raw = extractJSON(raw)
	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, nil // soft failure
	}
	return result, nil
}

// parseStringMap extracts a JSON string→string map from LLM output.
func parseStringMap(raw string) (map[string]string, error) {
	raw = extractJSON(raw)
	var result map[string]string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, nil // soft failure
	}
	return result, nil
}

// extractJSON strips markdown code fences and isolates the first JSON object or array.
var jsonObjectRe = regexp.MustCompile(`(?s)(\{.*\}|\[.*\])`)

func extractJSON(s string) string {
	// Remove markdown code fences
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.TrimSpace(s)
	if m := jsonObjectRe.FindString(s); m != "" {
		return m
	}
	return s
}
