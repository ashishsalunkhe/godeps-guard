package policy

import (
	"fmt"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// copyleftLicenses are licenses that carry strong copyleft obligations.
var copyleftLicenses = map[string]bool{
	"GPL-2.0":  true,
	"GPL-3.0":  true,
	"AGPL-3.0": true,
	"LGPL-3.0": true,
}

// CalculateRisk evaluates a 0-10 risk score for a single direct dependency impact.
// licenses is an optional map of modulePath -> license identifier.
// heavyPatterns is an optional list of module path prefixes considered heavy vendors.
func CalculateRisk(impact *types.ModuleImpact, licenses map[string]string, heavyPatterns []string) {
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
	if len(impact.Module.Version) > 2 && impact.Module.Version[:2] == "v0" {
		score += 2
		reasons = append(reasons, "Pre-v1 release implies volatility")
	}

	// Factor 3: Heavy Vendor Patterns (Medium Weight, config-driven)
	if len(heavyPatterns) > 0 {
		for _, pattern := range heavyPatterns {
			if strings.HasPrefix(impact.Module.Path, pattern) {
				score += 3
				reasons = append(reasons, fmt.Sprintf("Matches heavy vendor pattern: %s", pattern))
				break
			}
		}
	} else {
		// Default known heavy pattern if none configured
		if impact.Module.Path == "github.com/aws/aws-sdk-go" {
			score += 3
			reasons = append(reasons, "Known heavy vendor pattern")
		}
	}

	// Factor 4: License Risk (Medium Weight)
	if licenses != nil {
		if lic, ok := licenses[impact.Module.Path]; ok {
			if copyleftLicenses[lic] {
				score += 3
				reasons = append(reasons, fmt.Sprintf("Copyleft license detected (%s)", lic))
			} else if lic == "Unknown" {
				score += 1
				reasons = append(reasons, "License could not be detected")
			}
		}
	}

	// Factor 5: Transitive Module Count
	if len(impact.TransitiveModules) > 5 {
		score += 2
		reasons = append(reasons, fmt.Sprintf("Pulls in %d transitive modules", len(impact.TransitiveModules)))
	} else if len(impact.TransitiveModules) > 2 {
		score += 1
		reasons = append(reasons, fmt.Sprintf("Pulls in %d transitive modules", len(impact.TransitiveModules)))
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
