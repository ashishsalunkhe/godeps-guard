package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the .godepsguard.yaml configuration file format.
type Config struct {
	Build struct {
		Target  string   `yaml:"target"`
		Output  string   `yaml:"output"`
		Ldflags []string `yaml:"ldflags"`
	} `yaml:"build"`

	Thresholds struct {
		MaxDirectDepsAdded            int   `yaml:"max_direct_deps_added"`
		MaxTransitivePackagesAdded    int   `yaml:"max_transitive_packages_added"`
		MaxBinarySizeIncreaseBytes    int64 `yaml:"max_binary_size_increase_bytes"`
		MaxBinarySizeIncreasePercent  int   `yaml:"max_binary_size_increase_percent"`
	} `yaml:"thresholds"`

	Policies struct {
		BlockedModules               []string `yaml:"blocked_modules"`
		WarnOnIndirectOnlyGrowth     bool     `yaml:"warn_on_indirect_only_growth"`
		RequireReasonForNewDirectDep bool     `yaml:"require_reason_for_new_direct_dep"`
	} `yaml:"policies"`

	Report struct {
		Format  string `yaml:"format"`
		Verbose bool   `yaml:"verbose"`
	} `yaml:"report"`
}

// Default returns a Config with reasonable defaults.
func Default() *Config {
	cfg := &Config{}
	cfg.Build.Target = "./..."
	cfg.Build.Output = "/tmp/godepsguard_app"
	cfg.Report.Format = "markdown"
	cfg.Report.Verbose = false
	return cfg
}

// Load reads and parses a configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return cfg, nil
}
