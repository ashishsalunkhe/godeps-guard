package graph

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// RenderDOT generates a Graphviz DOT format dependency graph from a snapshot.
// If filter is non-empty, only modules matching the prefix are included.
func RenderDOT(snap *types.Snapshot, filter string, out io.Writer) {
	// Build module-to-module edges from the package import graph
	edges, directMods := buildModuleGraph(snap, filter)

	fmt.Fprintln(out, "digraph godeps {")
	fmt.Fprintln(out, "  rankdir=LR;")
	fmt.Fprintln(out, "  node [shape=box, style=\"rounded,filled\", fontname=\"Helvetica\"];")
	fmt.Fprintln(out, "  edge [color=\"#666666\"];")
	fmt.Fprintln(out)

	// App root node
	appModule := getAppModule(snap)
	fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#4CAF50\", fontcolor=white, label=\"%s\"];\n", appModule, appModule)

	// Render nodes with styling based on direct vs indirect
	rendered := map[string]bool{appModule: true}
	for mod := range edges {
		if rendered[mod] {
			continue
		}
		rendered[mod] = true
		if directMods[mod] {
			fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#BBDEFB\", label=\"%s\"];\n", mod, shortName(mod))
		} else {
			fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#FFF9C4\", style=\"rounded,filled,dashed\", label=\"%s\"];\n", mod, shortName(mod))
		}
	}
	// Also render target nodes that haven't been rendered as source
	for _, targets := range edges {
		for _, t := range targets {
			if rendered[t] {
				continue
			}
			rendered[t] = true
			if directMods[t] {
				fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#BBDEFB\", label=\"%s\"];\n", t, shortName(t))
			} else {
				fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#FFF9C4\", style=\"rounded,filled,dashed\", label=\"%s\"];\n", t, shortName(t))
			}
		}
	}

	fmt.Fprintln(out)

	// Render edges
	for from, tos := range edges {
		for _, to := range tos {
			if directMods[to] && from == appModule {
				fmt.Fprintf(out, "  \"%s\" -> \"%s\" [color=\"#1976D2\", penwidth=2];\n", from, to)
			} else {
				fmt.Fprintf(out, "  \"%s\" -> \"%s\";\n", from, to)
			}
		}
	}

	fmt.Fprintln(out, "}")
}

// RenderMermaid generates a Mermaid flowchart from a snapshot.
func RenderMermaid(snap *types.Snapshot, filter string, out io.Writer) {
	edges, directMods := buildModuleGraph(snap, filter)

	fmt.Fprintln(out, "graph LR")

	appModule := getAppModule(snap)

	// Node IDs must be safe for mermaid (no dots, slashes)
	idOf := func(mod string) string {
		r := strings.NewReplacer("/", "_", ".", "_", "-", "_", "@", "_")
		return r.Replace(mod)
	}

	// Render app node
	fmt.Fprintf(out, "  %s[\"%s\"]:::app\n", idOf(appModule), appModule)

	rendered := map[string]bool{appModule: true}
	allMods := map[string]bool{}
	for from, tos := range edges {
		allMods[from] = true
		for _, to := range tos {
			allMods[to] = true
		}
	}

	// Sort for deterministic output
	sortedMods := make([]string, 0, len(allMods))
	for m := range allMods {
		sortedMods = append(sortedMods, m)
	}
	sort.Strings(sortedMods)

	for _, mod := range sortedMods {
		if rendered[mod] {
			continue
		}
		rendered[mod] = true
		if directMods[mod] {
			fmt.Fprintf(out, "  %s[\"%s\"]:::direct\n", idOf(mod), shortName(mod))
		} else {
			fmt.Fprintf(out, "  %s[\"%s\"]:::indirect\n", idOf(mod), shortName(mod))
		}
	}

	fmt.Fprintln(out)

	// Render edges
	for from, tos := range edges {
		sortedTos := make([]string, len(tos))
		copy(sortedTos, tos)
		sort.Strings(sortedTos)
		for _, to := range sortedTos {
			fmt.Fprintf(out, "  %s --> %s\n", idOf(from), idOf(to))
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "  classDef app fill:#4CAF50,color:white,stroke:#333")
	fmt.Fprintln(out, "  classDef direct fill:#BBDEFB,stroke:#1976D2")
	fmt.Fprintln(out, "  classDef indirect fill:#FFF9C4,stroke:#FBC02D,stroke-dasharray: 5 5")
}

// buildModuleGraph builds module-to-module edges by analyzing the package import graph.
// Returns (edges map, direct modules set).
func buildModuleGraph(snap *types.Snapshot, filter string) (map[string][]string, map[string]bool) {
	// Build a map of package -> module
	pkgToMod := make(map[string]string)
	for _, p := range snap.Packages {
		if p.ModulePath != "" {
			pkgToMod[p.ImportPath] = p.ModulePath
		}
	}

	// Identify direct modules
	directMods := make(map[string]bool)
	for _, m := range snap.Modules {
		if !m.Indirect {
			directMods[m.Path] = true
		}
	}

	appModule := getAppModule(snap)

	// Build module-to-module edges from package imports
	edgeSet := make(map[string]map[string]bool)
	for _, p := range snap.Packages {
		fromMod := pkgToMod[p.ImportPath]
		if fromMod == "" {
			continue
		}

		for _, dep := range p.Deps {
			toMod := pkgToMod[dep]
			if toMod == "" || toMod == fromMod {
				continue
			}

			// Apply filter
			if filter != "" {
				if !strings.HasPrefix(fromMod, filter) && !strings.HasPrefix(toMod, filter) && fromMod != appModule {
					continue
				}
			}

			if edgeSet[fromMod] == nil {
				edgeSet[fromMod] = make(map[string]bool)
			}
			edgeSet[fromMod][toMod] = true
		}
	}

	// Convert set to sorted slices for deterministic output
	edges := make(map[string][]string)
	for from, tos := range edgeSet {
		for to := range tos {
			edges[from] = append(edges[from], to)
		}
		sort.Strings(edges[from])
	}

	return edges, directMods
}

// getAppModule returns the module path of the app itself (first module without a version).
func getAppModule(snap *types.Snapshot) string {
	// The app's own module typically has no version in the snapshot
	// We look for the target's module path in the packages
	if snap.Target != "" {
		for _, p := range snap.Packages {
			if p.ImportPath == snap.Target || strings.HasPrefix(p.ImportPath, snap.Target) {
				if p.ModulePath != "" {
					return p.ModulePath
				}
			}
		}
	}

	// Fallback: find the module that most packages belong to
	modCount := make(map[string]int)
	for _, p := range snap.Packages {
		if p.ModulePath != "" && !p.Standard {
			modCount[p.ModulePath]++
		}
	}
	bestMod := "app"
	bestCount := 0
	for mod, count := range modCount {
		// Prefer modules that are direct (no version = the app itself)
		found := false
		for _, m := range snap.Modules {
			if m.Path == mod && m.Version != "" {
				found = true
				break
			}
		}
		if !found && count > bestCount {
			bestMod = mod
			bestCount = count
		}
	}
	return bestMod
}

// shortName returns the last two path segments of a module path for brevity.
func shortName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	if len(parts) <= 2 {
		return modulePath
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
