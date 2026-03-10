package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// Output writes the delta and policy results in the specified format.
func Output(delta *types.Delta, policy *types.PolicyResult, format string, out io.Writer) error {
	switch format {
	case "json":
		return writeJSON(delta, policy, out)
	case "markdown":
		return writeMarkdown(delta, policy, out)
	default:
		// default to markdown
		return writeMarkdown(delta, policy, out)
	}
}

// OutputComment writes a highly condensed markdown report specifically for GitHub/Gitlab comments.
func OutputComment(delta *types.Delta, policy *types.PolicyResult, out io.Writer) error {
	fmt.Fprintln(out, "## Dependency Impact Report")
	fmt.Fprintln(out)

	if len(delta.DirectImpacts) > 0 {
		fmt.Fprintln(out, "### New Direct Dependencies")
		for _, impact := range delta.DirectImpacts {
			fmt.Fprintf(out, "- **%s**\\n", impact.Module.Path)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, "### Graph Growth")
	if delta.AddedPackages > 0 {
		fmt.Fprintf(out, "+%d packages\\n", delta.AddedPackages)
	} else {
		fmt.Fprintln(out, "No new packages.")
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "### Binary Change")
	if delta.BinarySizeDelta != 0 {
		fmt.Fprintf(out, "%+d bytes (%+.2f%%)\\n", delta.BinarySizeDelta, delta.BinaryDeltaPercent)
	} else {
		fmt.Fprintln(out, "No change.")
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "### Result")
	if policy != nil && !policy.Passed {
		fmt.Fprintln(out, "❌ **Fails policy**")
		for _, err := range policy.Errors {
			fmt.Fprintf(out, "- %s\\n", err)
		}
	} else {
		fmt.Fprintln(out, "✅ **Passes policy**")
	}

	return nil
}

func writeJSON(delta *types.Delta, policy *types.PolicyResult, out io.Writer) error {
	result := struct {
		Delta  *types.Delta        `json:"delta"`
		Policy *types.PolicyResult `json:"policy,omitempty"`
	}{
		Delta:  delta,
		Policy: policy,
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func writeMarkdown(delta *types.Delta, policy *types.PolicyResult, out io.Writer) error {
	fmt.Fprintln(out, "# Dependency Guard Report")

	if policy != nil {
		if policy.Passed {
			fmt.Fprintln(out, "## 🟢 Pass")
		} else {
			fmt.Fprintln(out, "## 🔴 Fail")
		}

		if len(policy.Errors) > 0 {
			fmt.Fprintln(out, "### Errors")
			for _, err := range policy.Errors {
				fmt.Fprintf(out, "- %s\\n", err)
			}
		}

		if len(policy.Warnings) > 0 {
			fmt.Fprintln(out, "### Warnings")
			for _, warn := range policy.Warnings {
				fmt.Fprintf(out, "- %s\\n", warn)
			}
		}
	}

	fmt.Fprintln(out, "## Summary")
	fmt.Fprintln(out, "| Metric | Value |")
	fmt.Fprintln(out, "|---|---|")
	fmt.Fprintf(out, "| Added Modules | %d |\\n", len(delta.AddedModules))
	fmt.Fprintf(out, "| Removed Modules | %d |\\n", len(delta.RemovedModules))
	fmt.Fprintf(out, "| Changed Modules | %d |\\n", len(delta.ChangedModules))
	fmt.Fprintf(out, "| Added Packages | %d |\\n", delta.AddedPackages)
	fmt.Fprintf(out, "| Removed Packages | %d |\\n", delta.RemovedPackages)

	if delta.BinarySizeBefore > 0 || delta.BinarySizeAfter > 0 {
		fmt.Fprintf(out, "| Binary Size Delta | %d bytes (%.2f%%) |\\n", delta.BinarySizeDelta, delta.BinaryDeltaPercent)
	}

	// Dependency Impact Reports
	if len(delta.DirectImpacts) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "### New Direct Dependencies")
		for _, impact := range delta.DirectImpacts {
			fmt.Fprintf(out, "- **%s** `%s`\\n", impact.Module.Path, impact.Module.Version)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "### Transitive Growth Attribution")
		for _, impact := range delta.DirectImpacts {
			if impact.AddedPackages > 0 {
				fmt.Fprintf(out, "- **%s**\\n", impact.Module.Path)
				fmt.Fprintf(out, "  → added %d packages\\n", impact.AddedPackages)
			}
		}
	} else if len(delta.AddedModules) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "### Added Modules")
		for _, m := range delta.AddedModules {
			fmt.Fprintf(out, "- %s@%s\\n", m.Path, m.Version)
		}
	}

	if len(delta.RemovedModules) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "### Removed Modules")
		for _, m := range delta.RemovedModules {
			fmt.Fprintf(out, "- %s@%s\\n", m.Path, m.Version)
		}
	}

	if len(delta.ChangedModules) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "### Changed Modules")
		for _, m := range delta.ChangedModules {
			fmt.Fprintf(out, "- %s (`%s` -> `%s`)\\n", m.Path, m.Before, m.After)
		}
	}

	return nil
}
