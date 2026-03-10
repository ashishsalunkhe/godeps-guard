package golist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// GoListPackage represents the JSON output from go list -deps -json ./...
type GoListPackage struct {
	ImportPath string   `json:"ImportPath"`
	Standard   bool     `json:"Standard"`
	Deps       []string `json:"Deps"`
	Module     *GoModule`json:"Module"`
}

// GoModule represents the Module field in the go list output.
type GoModule struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
	Replace  *struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Replace"`
}

// Run executes go list -deps -json ./... and builds a snapshot of packages and modules.
func Run(dir string) ([]types.PackageNode, []types.ModuleRef, error) {
	cmd := exec.Command("go", "list", "-deps", "-json", "./...")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run go list: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(out))
	var packages []types.PackageNode
	modulesMap := make(map[string]types.ModuleRef)

	for decoder.More() {
		var pkg GoListPackage
		if err := decoder.Decode(&pkg); err != nil {
			return nil, nil, fmt.Errorf("failed to decode go list package: %w", err)
		}

		node := types.PackageNode{
			ImportPath: pkg.ImportPath,
			Standard:   pkg.Standard,
			Deps:       pkg.Deps,
		}

		if pkg.Module != nil {
			node.ModulePath = pkg.Module.Path

			ref := types.ModuleRef{
				Path:     pkg.Module.Path,
				Version:  pkg.Module.Version,
				Indirect: pkg.Module.Indirect,
			}
			if pkg.Module.Replace != nil {
				ref.Replace = fmt.Sprintf("%s@%s", pkg.Module.Replace.Path, pkg.Module.Replace.Version)
				if pkg.Module.Replace.Version == "" {
					ref.Replace = pkg.Module.Replace.Path
				}
			}

			if existing, ok := modulesMap[ref.Path]; ok {
				if !ref.Indirect {
					existing.Indirect = false
					modulesMap[ref.Path] = existing
				}
			} else {
				if ref.Version != "" { 
					modulesMap[ref.Path] = ref
				}
			}
		}

		packages = append(packages, node)
	}

	var modules []types.ModuleRef
	for _, m := range modulesMap {
		modules = append(modules, m)
	}

	return packages, modules, nil
}
