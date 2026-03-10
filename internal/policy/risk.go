package policy

import (
	"fmt"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// CalculateRisk evaluates a 0-10 risk score for a single direct dependency impact.
func CalculateRisk(impact *types.ModuleImpact) {
	score := 0
	var reasons []string

	// Factor 1: Transitive Fanout (High Weight)
	if impact.AddedPackages >= 50 {
		score += 5
		reasons = append(reasons, fmt.Sprintf("Massive dependency fanout (%d packages)", impact.AddedPackages))
	} else if impact.AddedPackages >= 20 {
		score += 3
		reasons = append(reasons, fmt.Sprintf("Large dependency fanout (%d packages)", impact.AddedPackages))
	} else if impact.AddedPackages >= 5 {
		score += 1
		reasons = append(reasons, fmt.Sprintf("Moderate fanout (%d packages)", impact.AddedPackages))
	} else if impact.AddedPackages == 0 {
		reasons = append(reasons, "No transitive imports")
	}

	// Factor 2: Version Volatility (Low Weight, e.g. v0.x.x)
	// Very simple check for pre-1.0
	if len(impact.Module.Version) > 2 && impact.Module.Version[:2] == "v0" {
		score += 2
		reasons = append(reasons, "Pre-v1 release implies volatility")
	}

	// Factor 3: Known patterns (Medium Weight)
	// Hardcoded example for v1.2, could be moved to config
	if impact.Module.Path == "github.com/aws/aws-sdk-go" {
		score += 3
		reasons = append(reasons, "Known heavy vendor pattern")
	}

	// Cap at 10
	if score > 10 {
		score = 10
	}
	if score == 0 && impact.AddedPackages <= 2 {
		score = 1
		reasons = append(reasons, "Small, contained dependency")
	}

	impact.RiskScore = score
	impact.RiskReasons = reasons
}
