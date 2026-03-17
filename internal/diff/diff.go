package diff

import (
	"github.com/ashishsalunkhe/godeps-guard/internal/policy"
	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Compare Snapshots computes the exact difference between a base and head snapshot.
// licenses is an optional map of modulePath -> license identifier for risk scoring.
// heavyPatterns is an optional list of module path prefixes considered heavy vendors.
func Compare(base, head *types.Snapshot, licenses map[string]string, heavyPatterns []string) *types.Delta {
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

	// Build a set of added module paths for transitive tracking
	addedModSet := make(map[string]bool)
	for _, m := range delta.AddedModules {
		addedModSet[m.Path] = true
	}

	// Build package import graph for transitive walking
	pkgToMod := make(map[string]string)
	for _, p := range head.Packages {
		if p.ModulePath != "" {
			pkgToMod[p.ImportPath] = p.ModulePath
		}
	}

	// Dependency attribution (Blame)
	// For every added *direct* module, trace which transitive modules it pulls in.
	for _, addedMod := range delta.AddedModules {
		if !addedMod.Indirect {
			impact := types.ModuleImpact{
				Module: addedMod,
			}

			// Count packages belonging to this module
			pkgCount := 0
			for _, p := range addedPkgs {
				if p.ModulePath == addedMod.Path {
					pkgCount++
				}
			}
			impact.AddedPackages = pkgCount

			// Deep transitive blame: walk the import graph from this module's packages
			// to find which other newly-added modules are reachable
			transitiveModules := traceTransitiveModules(addedMod.Path, head.Packages, pkgToMod, addedModSet)
			impact.TransitiveModules = transitiveModules

			// Calculate Risk
			policy.CalculateRisk(&impact, licenses, heavyPatterns)

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

// traceTransitiveModules walks the import graph starting from packages belonging to
// directModPath, and returns a list of other newly-added modules reachable transitively.
func traceTransitiveModules(directModPath string, packages []types.PackageNode, pkgToMod map[string]string, addedMods map[string]bool) []string {
	// Build adjacency list
	depGraph := make(map[string][]string)
	for _, p := range packages {
		depGraph[p.ImportPath] = p.Deps
	}

	// BFS from all packages of the direct module
	visited := make(map[string]bool)
	queue := []string{}
	for _, p := range packages {
		if pkgToMod[p.ImportPath] == directModPath {
			queue = append(queue, p.ImportPath)
			visited[p.ImportPath] = true
		}
	}

	transitiveModSet := make(map[string]bool)
	for len(queue) > 0 {
		pkg := queue[0]
		queue = queue[1:]

		for _, dep := range depGraph[pkg] {
			if visited[dep] {
				continue
			}
			visited[dep] = true

			depMod := pkgToMod[dep]
			if depMod != "" && depMod != directModPath && addedMods[depMod] {
				transitiveModSet[depMod] = true
			}

			queue = append(queue, dep)
		}
	}

	result := make([]string, 0, len(transitiveModSet))
	for mod := range transitiveModSet {
		result = append(result, mod)
	}
	return result
}
