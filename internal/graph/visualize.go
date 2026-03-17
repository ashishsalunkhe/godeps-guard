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
func RenderDOT(snap *types.Snapshot, filter string, risks map[string]int, out io.Writer) {
	// Build module-to-module edges from the package import graph
	edges, directMods := buildModuleGraph(snap, filter)

	fmt.Fprintln(out, "digraph godeps {")
	fmt.Fprintln(out, "  rankdir=LR;")
	fmt.Fprintln(out, "  node [shape=box, style=\"rounded,filled\", fontname=\"Helvetica\"];")
	fmt.Fprintln(out, "  edge [color=\"#666666\"];")
	fmt.Fprintln(out)

	appModule := getAppModule(snap)
	allModsSorted := getSortedMods(edges)

	// Map modules to safe, anonymous IDs
	idMap := make(map[string]string)
	idMap[appModule] = "root"
	nodeCounter := 1
	for _, mod := range allModsSorted {
		if mod != appModule {
			idMap[mod] = fmt.Sprintf("node_%d", nodeCounter)
			nodeCounter++
		}
	}

	// App root node
	fmt.Fprintf(out, "  \"%s\" [fillcolor=\"#4CAF50\", fontcolor=white, label=\"%s\"];\n", idMap[appModule], "Project Root")

	// Render nodes with styling based on direct vs indirect and risk
	rendered := map[string]bool{appModule: true}
	for _, mod := range allModsSorted {
		if rendered[mod] {
			continue
		}
		rendered[mod] = true
		
		color := getNodeColor(mod, directMods[mod], risks[mod])
		style := "rounded,filled"
		if !directMods[mod] {
			style += ",dashed"
		}
		
		fmt.Fprintf(out, "  \"%s\" [fillcolor=\"%s\", style=\"%s\", label=\"%s\"];\n", idMap[mod], color, style, shortName(mod))
	}

	fmt.Fprintln(out)

	// Render edges
	for _, from := range allModsSorted {
		tos := edges[from]
		for _, to := range tos {
			if directMods[to] && from == appModule {
				fmt.Fprintf(out, "  \"%s\" -> \"%s\" [color=\"#1976D2\", penwidth=2];\n", idMap[from], idMap[to])
			} else {
				fmt.Fprintf(out, "  \"%s\" -> \"%s\";\n", idMap[from], idMap[to])
			}
		}
	}

	fmt.Fprintln(out, "}")
}

// RenderMermaid generates a Mermaid flowchart from a snapshot.
func RenderMermaid(snap *types.Snapshot, filter string, risks map[string]int, out io.Writer) {
	edges, directMods := buildModuleGraph(snap, filter)

	fmt.Fprintln(out, "graph LR")

	appModule := getAppModule(snap)
	allModsSorted := getSortedMods(edges)

	// Map modules to safe, anonymous IDs
	idMap := make(map[string]string)
	idMap[appModule] = "root"
	nodeCounter := 1
	for _, mod := range allModsSorted {
		if mod != appModule {
			idMap[mod] = fmt.Sprintf("node_%d", nodeCounter)
			nodeCounter++
		}
	}

	// Render app node
	fmt.Fprintf(out, "  %s[\"%s\"]:::app\n", idMap[appModule], "Project Root")

	rendered := map[string]bool{appModule: true}
	for _, mod := range allModsSorted {
		if rendered[mod] {
			continue
		}
		rendered[mod] = true
		
		class := "indirect"
		if directMods[mod] {
			class = "direct"
		}
		
		// If high risk, override class
		risk := risks[mod]
		if risk >= 7 {
			class = "highrisk"
		} else if risk >= 4 {
			class = "medrisk"
		}

		fmt.Fprintf(out, "  %s[\"%s\"]:::%s\n", idMap[mod], shortName(mod), class)
	}

	fmt.Fprintln(out)

	// Render edges
	for _, from := range allModsSorted {
		tos := edges[from]
		for _, to := range tos {
			fmt.Fprintf(out, "  %s --> %s\n", idMap[from], idMap[to])
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "  classDef app fill:#4CAF50,color:white,stroke:#333")
	fmt.Fprintln(out, "  classDef direct fill:#BBDEFB,stroke:#1976D2")
	fmt.Fprintln(out, "  classDef indirect fill:#FFF9C4,stroke:#FBC02D,stroke-dasharray: 5 5")
	fmt.Fprintln(out, "  classDef medrisk fill:#FFE0B2,stroke:#FB8C00")
	fmt.Fprintln(out, "  classDef highrisk fill:#FFCDD2,stroke:#E53935")
}

func getNodeColor(mod string, direct bool, risk int) string {
	if risk >= 7 {
		return "#FFCDD2" // Red for high risk
	}
	if risk >= 4 {
		return "#FFE0B2" // Orange for med risk
	}
	if direct {
		return "#BBDEFB" // Blue for direct
	}
	return "#FFF9C4" // Yellow for indirect
}

func getSortedMods(edges map[string][]string) []string {
	modSet := make(map[string]bool)
	for from, tos := range edges {
		modSet[from] = true
		for _, to := range tos {
			modSet[to] = true
		}
	}
	mods := make([]string, 0, len(modSet))
	for m := range modSet {
		mods = append(mods, m)
	}
	sort.Strings(mods)
	return mods
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

// getAppModule returns the module path of the app itself.
func getAppModule(snap *types.Snapshot) string {
	if snap.Target != "" {
		for _, p := range snap.Packages {
			if p.ImportPath == snap.Target || strings.HasPrefix(p.ImportPath, snap.Target) {
				if p.ModulePath != "" {
					return p.ModulePath
				}
			}
		}
	}
	return "app"
}

// shortName returns the last segment of a module path for brevity and anonymity.
func shortName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	return parts[len(parts)-1]
}
