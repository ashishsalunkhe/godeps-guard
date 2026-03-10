package policy

import (
	"fmt"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/internal/config"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Evaluate checks the delta against the provided configuration.
func Evaluate(delta *types.Delta, cfg *config.Config) *types.PolicyResult {
	result := &types.PolicyResult{
		Passed: true,
	}

	// Helper to add error
	addError := func(msg string) {
		result.Passed = false
		result.Errors = append(result.Errors, msg)
	}

	// Direct dependencies added
	if cfg.Thresholds.MaxDirectDepsAdded > 0 {
		directAdded := 0
		for _, m := range delta.AddedModules {
			if !m.Indirect {
				directAdded++
			}
		}
		if directAdded > cfg.Thresholds.MaxDirectDepsAdded {
			addError(fmt.Sprintf("Added %d direct dependencies, which exceeds the limit of %d", directAdded, cfg.Thresholds.MaxDirectDepsAdded))
		}
	}

	// Transitive packages added
	if cfg.Thresholds.MaxTransitivePackagesAdded > 0 && delta.AddedPackages > cfg.Thresholds.MaxTransitivePackagesAdded {
		addError(fmt.Sprintf("Added %d transitive packages, which exceeds the limit of %d", delta.AddedPackages, cfg.Thresholds.MaxTransitivePackagesAdded))
	}

	// Binary size byte increase
	if cfg.Thresholds.MaxBinarySizeIncreaseBytes > 0 && delta.BinarySizeDelta > cfg.Thresholds.MaxBinarySizeIncreaseBytes {
		addError(fmt.Sprintf("Binary size increased by %d bytes, which exceeds the limit of %d bytes", delta.BinarySizeDelta, cfg.Thresholds.MaxBinarySizeIncreaseBytes))
	}

	// Binary size percent increase
	if cfg.Thresholds.MaxBinarySizeIncreasePercent > 0 && int(delta.BinaryDeltaPercent) > cfg.Thresholds.MaxBinarySizeIncreasePercent {
		addError(fmt.Sprintf("Binary size increased by %.2f%%, which exceeds the limit of %d%%", delta.BinaryDeltaPercent, cfg.Thresholds.MaxBinarySizeIncreasePercent))
	}

	// Blocked modules
	if len(cfg.Policies.BlockedModules) > 0 {
		blockedMap := make(map[string]bool)
		for _, b := range cfg.Policies.BlockedModules {
			blockedMap[b] = true
		}

		for _, m := range delta.AddedModules {
			if blockedMap[m.Path] {
				addError(fmt.Sprintf("Added module %s which is on the blocked list", m.Path))
			} else {
				for _, b := range cfg.Policies.BlockedModules {
					if strings.HasPrefix(m.Path, b) {
						addError(fmt.Sprintf("Added module %s which matches blocked prefix %s", m.Path, b))
					}
				}
			}
		}
	}

	// Allowed modules (Strict Allowlist Mode)
	if len(cfg.Policies.AllowedModules) > 0 {
		allowedMap := make(map[string]bool)
		for _, a := range cfg.Policies.AllowedModules {
			allowedMap[a] = true
		}

		for _, m := range delta.AddedModules {
			isAllowed := false
			if allowedMap[m.Path] {
				isAllowed = true
			} else {
				for _, a := range cfg.Policies.AllowedModules {
					if strings.HasPrefix(m.Path, a) {
						isAllowed = true
						break
					}
				}
			}
			if !isAllowed && !m.Indirect {
				addError(fmt.Sprintf("Added module %s is not on the allowed list. Strict allowlist is enabled.", m.Path))
			}
		}
	}

	// Warnings
	if cfg.Policies.WarnOnIndirectOnlyGrowth {
		if delta.AddedPackages > 10 && len(delta.AddedModules) > 0 {
			directAdded := 0
			for _, m := range delta.AddedModules {
				if !m.Indirect {
					directAdded++
				}
			}
			if directAdded == 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Graph grew by %d packages with only indirect module additions.", delta.AddedPackages))
			}
		}
	}

	return result
}
