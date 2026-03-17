package sbom

import (
	"encoding/json"
	"io"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// LicenseEntry represents a license in CycloneDX format.
type LicenseEntry struct {
	ID string `json:"id,omitempty"`
}

// LicenseWrapper wraps a license entry for the CycloneDX licenses array.
type LicenseWrapper struct {
	License LicenseEntry `json:"license"`
}

// Component represents a CycloneDX component.
type Component struct {
	Type     string           `json:"type"`
	Name     string           `json:"name"`
	Version  string           `json:"version,omitempty"`
	Purl     string           `json:"purl,omitempty"`
	Licenses []LicenseWrapper `json:"licenses,omitempty"`
}

// SBOM represents a CycloneDX SBOM document.
type SBOM struct {
	BomFormat   string      `json:"bomFormat"`
	SpecVersion string      `json:"specVersion"`
	Version     int         `json:"version"`
	Components  []Component `json:"components"`
}

// CycloneDX generates a CycloneDX format SBOM JSON, optionally enriched with license data.
func CycloneDX(snap *types.Snapshot, licenses map[string]string, out io.Writer) error {
	doc := SBOM{
		BomFormat:   "CycloneDX",
		SpecVersion: "1.4",
		Version:     1,
	}

	for _, m := range snap.Modules {
		// Build Package URL
		purl := "pkg:golang/" + m.Path
		if m.Version != "" {
			purl += "@" + m.Version
		}

		comp := Component{
			Type:    "library",
			Name:    m.Path,
			Version: m.Version,
			Purl:    purl,
		}

		// Enrich with license if available
		if licenses != nil {
			if lic, ok := licenses[m.Path]; ok && lic != "" && lic != "Unknown" {
				comp.Licenses = []LicenseWrapper{
					{License: LicenseEntry{ID: lic}},
				}
			}
		}

		doc.Components = append(doc.Components, comp)
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}
