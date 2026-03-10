package sbom

import (
	"encoding/json"
	"io"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// CycloneDX generates a basic CycloneDX format SBOM JSON.
func CycloneDX(snap *types.Snapshot, out io.Writer) error {
	type Component struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Version string `json:"version,omitempty"`
		Purl    string `json:"purl,omitempty"`
	}

	type SBOM struct {
		BomFormat   string      `json:"bomFormat"`
		SpecVersion string      `json:"specVersion"`
		Version     int         `json:"version"`
		Components  []Component `json:"components"`
	}

	doc := SBOM{
		BomFormat:   "CycloneDX",
		SpecVersion: "1.4",
		Version:     1,
	}

	for _, m := range snap.Modules {
		// e.g. pkg:golang/github.com/google/uuid@v1.6.0
		purl := "pkg:golang/" + m.Path
		if m.Version != "" {
			purl += "@" + m.Version
		}

		doc.Components = append(doc.Components, Component{
			Type:    "library",
			Name:    m.Path,
			Version: m.Version,
			Purl:    purl,
		})
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}
