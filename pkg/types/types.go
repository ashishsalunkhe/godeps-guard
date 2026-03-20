package types

// AIInsights carries AI-generated content optionally appended to reports.
type AIInsights struct {
	Summary      string            `json:"summary,omitempty"`
	Alternatives map[string]string `json:"alternatives,omitempty"`
	TrendReport  string            `json:"trend_report,omitempty"`
	AIWarnings   []string          `json:"ai_warnings,omitempty"`
}

// ModuleRef represents a Go module reference.
type ModuleRef struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	Indirect bool   `json:"indirect"`
	Replace  string `json:"replace,omitempty"`
}

// ModuleChange represents the change in version for a specific module.
type ModuleChange struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// PackageNode represents a package in the dependency graph.
type PackageNode struct {
	ImportPath string   `json:"import_path"`
	ModulePath string   `json:"module_path"`
	Standard   bool     `json:"standard"`
	Deps       []string `json:"deps"`
}

// Snapshot represents a full capturing of dependencies and binary data at a point in time.
type Snapshot struct {
	Modules    []ModuleRef   `json:"modules"`
	Packages   []PackageNode `json:"packages"`
	BinarySize int64         `json:"binary_size"`
	Commit     string        `json:"commit"`
	Target     string        `json:"target"`
}

// ModuleImpact stores blame and transitive size logic for a direct module addition.
type ModuleImpact struct {
	Module            ModuleRef `json:"module"`
	AddedPackages     int       `json:"added_packages"`
	TransitiveModules []string  `json:"transitive_modules,omitempty"`
	RiskScore         int       `json:"risk_score"`
	RiskReasons       []string  `json:"risk_reasons"`
}

// Delta represents the difference between two Snapshots.
type Delta struct {
	AddedModules       []ModuleRef    `json:"added_modules"`
	RemovedModules     []ModuleRef    `json:"removed_modules"`
	ChangedModules     []ModuleChange `json:"changed_modules"`
	AddedPackages      int            `json:"added_packages"`
	RemovedPackages    int            `json:"removed_packages"`
	BinarySizeBefore   int64          `json:"binary_size_before"`
	BinarySizeAfter    int64          `json:"binary_size_after"`
	BinarySizeDelta    int64          `json:"binary_size_delta"`
	BinaryDeltaPercent float64        `json:"binary_delta_percent"`

	// Post-v1 Attribution
	DirectImpacts []ModuleImpact `json:"direct_impacts"`

	// AI-generated insights (populated when --ai flag is used)
	AI AIInsights `json:"ai,omitempty"`
}

// PolicyResult holds the results of evaluating the rules.
type PolicyResult struct {
	Passed   bool     `json:"passed"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`

	// AI-generated warnings (populated when --ai flag is used)
	AIWarnings []string `json:"ai_warnings,omitempty"`
}
