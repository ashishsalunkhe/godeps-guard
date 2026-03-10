package diff

import (
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Compare Snapshots computes the exact difference between a base and head snapshot.
func Compare(base, head *types.Snapshot) *types.Delta {
	delta := &types.Delta{}

	// Map base modules by path
	baseModules := make(map[string]types.ModuleRef)
	for _, m := range base.Modules {
		baseModules[m.Path] = m
	}

	// Map head modules by path
	headModules := make(map[string]types.ModuleRef)
	for _, m := range head.Modules {
		headModules[m.Path] = m
	}

	// Find added and changed modules
	for path, headMod := range headModules {
		if baseMod, exists := baseModules[path]; exists {
			// Exists in both, check for change
			if headMod.Version != baseMod.Version {
				delta.ChangedModules = append(delta.ChangedModules, types.ModuleChange{
					Path:   path,
					Before: baseMod.Version,
					After:  headMod.Version,
				})
			}
		} else {
			// Exists only in head -> Added
			delta.AddedModules = append(delta.AddedModules, headMod)
		}
	}

	// Find removed modules
	for path, baseMod := range baseModules {
		if _, exists := headModules[path]; !exists {
			delta.RemovedModules = append(delta.RemovedModules, baseMod)
		}
	}

	// Calculate package differences simply by counting for now
	delta.AddedPackages = len(head.Packages) - len(base.Packages)
	if delta.AddedPackages < 0 {
		delta.RemovedPackages = -delta.AddedPackages
		delta.AddedPackages = 0
	}

	// Binary size metrics (we'll implement this fully in Milestone 3)
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
