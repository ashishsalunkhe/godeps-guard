package license

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ashishsalunkhe/godeps-guard/pkg/types"
)

// LicenseInfo holds the detected license for a module.
type LicenseInfo struct {
	ModulePath string `json:"module_path"`
	Version    string `json:"version"`
	License    string `json:"license"`
}

// licenseFiles are filenames we look for in a module's directory.
var licenseFiles = []string{
	"LICENSE",
	"LICENSE.txt",
	"LICENSE.md",
	"LICENCE",
	"LICENCE.txt",
	"LICENCE.md",
	"COPYING",
	"COPYING.txt",
	"COPYING.md",
}

// classifierRules maps substring patterns to SPDX identifiers.
// Order matters: more specific patterns should come first.
var classifierRules = []struct {
	patterns []string
	license  string
}{
	{patterns: []string{"GNU AFFERO GENERAL PUBLIC LICENSE", "AGPL"}, license: "AGPL-3.0"},
	{patterns: []string{"GNU LESSER GENERAL PUBLIC LICENSE", "LGPL"}, license: "LGPL-3.0"},
	{patterns: []string{"GNU GENERAL PUBLIC LICENSE", "Version 3"}, license: "GPL-3.0"},
	{patterns: []string{"GNU GENERAL PUBLIC LICENSE", "Version 2"}, license: "GPL-2.0"},
	{patterns: []string{"MOZILLA PUBLIC LICENSE", "Version 2.0"}, license: "MPL-2.0"},
	{patterns: []string{"APACHE LICENSE", "VERSION 2.0"}, license: "Apache-2.0"},
	{patterns: []string{"APACHE LICENSE"}, license: "Apache-2.0"},
	{patterns: []string{"MIT LICENSE"}, license: "MIT"},
	{patterns: []string{"PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"}, license: "MIT"},
	{patterns: []string{"ISC LICENSE"}, license: "ISC"},
	{patterns: []string{"PERMISSION TO USE, COPY, MODIFY, AND/OR DISTRIBUTE"}, license: "ISC"},
	{patterns: []string{"THE UNLICENSE"}, license: "Unlicense"},
	{patterns: []string{"THIS IS FREE AND UNENCUMBERED SOFTWARE"}, license: "Unlicense"},
	{patterns: []string{"REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS", "3 CONDITIONS"}, license: "BSD-3-Clause"},
	{patterns: []string{"REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS", "THREE CONDITIONS"}, license: "BSD-3-Clause"},
	{patterns: []string{"REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS", "2 CONDITIONS"}, license: "BSD-2-Clause"},
	{patterns: []string{"REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS", "TWO CONDITIONS"}, license: "BSD-2-Clause"},
	{patterns: []string{"REDISTRIBUTION AND USE IN SOURCE AND BINARY FORMS"}, license: "BSD-3-Clause"},
	{patterns: []string{"BOOST SOFTWARE LICENSE"}, license: "BSL-1.0"},
	{patterns: []string{"CREATIVE COMMONS"}, license: "CC"},
	{patterns: []string{"ECLIPSE PUBLIC LICENSE"}, license: "EPL-2.0"},
}

// Detect scans GOMODCACHE and returns license info for each module.
func Detect(modules []types.ModuleRef) []LicenseInfo {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	modCache := filepath.Join(gopath, "pkg", "mod")

	// Also check GOMODCACHE env override
	if override := os.Getenv("GOMODCACHE"); override != "" {
		modCache = override
	}

	var results []LicenseInfo

	for _, m := range modules {
		lic := detectForModule(modCache, m.Path, m.Version)
		results = append(results, LicenseInfo{
			ModulePath: m.Path,
			Version:    m.Version,
			License:    lic,
		})
	}

	return results
}

// DetectMap is a convenience wrapper that returns a map of modulePath → license.
func DetectMap(modules []types.ModuleRef) map[string]string {
	infos := Detect(modules)
	result := make(map[string]string, len(infos))
	for _, info := range infos {
		result[info.ModulePath] = info.License
	}
	return result
}

func detectForModule(modCache, modulePath, version string) string {
	// Module cache stores modules at paths like:
	//   $GOMODCACHE/github.com/spf13/cobra@v1.10.2/LICENSE.txt
	// The first letter of each path element is case-encoded with '!' prefix for uppercase.
	encodedPath := encodePath(modulePath)

	var dir string
	if version != "" {
		dir = filepath.Join(modCache, encodedPath+"@"+version)
	} else {
		dir = filepath.Join(modCache, encodedPath)
	}

	for _, name := range licenseFiles {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		if lic := classify(string(content)); lic != "" {
			return lic
		}
	}

	return "Unknown"
}

// classify matches the content of a license file against known patterns.
func classify(content string) string {
	upper := strings.ToUpper(content)

	for _, rule := range classifierRules {
		allMatch := true
		for _, pattern := range rule.patterns {
			if !strings.Contains(upper, pattern) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return rule.license
		}
	}

	return ""
}

// encodePath converts a module path to the case-encoded form used in the module cache.
// In the module cache, uppercase letters are replaced with '!' followed by the lowercase letter.
func encodePath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if 'A' <= r && r <= 'Z' {
			b.WriteByte('!')
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
