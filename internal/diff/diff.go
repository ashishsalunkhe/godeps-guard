package diff

import (
	"github.com/ashishsalunkhe/godeps-guard/internal/policy"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Compare Snapshots computes the exact difference between a base and head snapshot.
func Compare(base, head *types.Snapshot) *types.Delta {
	delta := &types.Delta{}

	baseModules := make(map[string]types.ModuleRef)
	for _, m := range base.Modules {
		baseModules[m.Path] = m
	}

	headModules := make(map[string]types.ModuleRef)
	for _, m := range head.Modules {
		headModules[m.Path] = m
	}

	for path, headMod := range headModules {
		if baseMod, exists := baseModules[path]; exists {
			if headMod.Version != baseMod.Version {
				delta.ChangedModules = append(delta.ChangedModules, types.ModuleChange{
					Path:   path,
					Before: baseMod.Version,
					After:  headMod.Version,
				})
			}
		} else {
			delta.AddedModules = append(delta.AddedModules, headMod)
		}
	}

	for path, baseMod := range baseModules {
		if _, exists := headModules[path]; !exists {
			delta.RemovedModules = append(delta.RemovedModules, baseMod)
		}
	}

	basePkgs := make(map[string]bool)
	for _, p := range base.Packages {
		basePkgs[p.ImportPath] = true
	}

	headPkgs := make(map[string]types.PackageNode)
	for _, p := range head.Packages {
		headPkgs[p.ImportPath] = p
	}

	var addedPkgs []types.PackageNode
	for path, p := range headPkgs {
		if !basePkgs[path] {
			addedPkgs = append(addedPkgs, p)
		}
	}

	var removedPkgs []string
	for path := range basePkgs {
		if _, exists := headPkgs[path]; !exists {
			removedPkgs = append(removedPkgs, path)
		}
	}

	delta.AddedPackages = len(addedPkgs)
	delta.RemovedPackages = len(removedPkgs)

	// Dependency attribution (Blame)
	// For every added *direct* module, how many packages in addedPkgs belong to it?
	for _, addedMod := range delta.AddedModules {
		if !addedMod.Indirect {
			impact := types.ModuleImpact{
				Module: addedMod,
			}

			// Simple attribution: How many packages in the graph share this module path?
			// Note: true blame of transitives is complicated (who imported the transitive?)
			// For v1.1, we count packages belonging exactly to this module path
			pkgCount := 0
			for _, p := range addedPkgs {
				if p.ModulePath == addedMod.Path {
					pkgCount++
				}
			}

			impact.AddedPackages = pkgCount

			// Post-v1: Calculate Risk
			policy.CalculateRisk(&impact)

			delta.DirectImpacts = append(delta.DirectImpacts, impact)
		}
	}

	delta.BinarySizeBefore = base.BinarySize
	delta.BinarySizeAfter = head.BinarySize
	delta.BinarySizeDelta = head.BinarySize - base.BinarySize
	if base.BinarySize > 0 {
		delta.BinaryDeltaPercent = float64(delta.BinarySizeDelta) / float64(base.BinarySize) * 100.0
	} else if head.BinarySize > 0 {
		delta.BinaryDeltaPercent = 100.0
	}

	return delta
}
