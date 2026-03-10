# godeps-guard

> Prevent dependency sprawl and binary bloat in Go services before it reaches production.

A Go dependency impact analyzer and CI enforcement tool.

## Features

- Inspects dependency graph changes in a project over time.
- Measures built binary size changes.
- Applies configured guardrails to fail CI if:
  - Too many direct deps added
  - Too many transitive deps added
  - Binary size grows by X MB or X %
  - Blocked vendor patterns are matched

## Installation

\`\`\`bash
go install github.com/ashishsalunkhe/godeps-guard/cmd/godeps-guard@latest
\`\`\`

## Usage

### 1. Initialize Configuration

Create a \`.godepsguard.yaml\` file in the root of your project:

\`\`\`yaml
build:
  target: ./cmd/godeps-guard
  output: /tmp/app
  ldflags:
    - "-s"
    - "-w"

thresholds:
  max_direct_deps_added: 2
  max_transitive_packages_added: 40
  max_binary_size_increase_bytes: 5242880 # 5MB
  max_binary_size_increase_percent: 8

policies:
  blocked_modules:
    - github.com/aws/aws-sdk-go
  warn_on_indirect_only_growth: true
  require_reason_for_new_direct_dep: true

report:
  format: markdown
  verbose: true
\`\`\`

### 2. Run in GitHub Actions CI

See `.github/workflows/godeps-guard.yaml` for a working GitHub actions pipeline.

Or manually run:

\`\`\`bash
godeps-guard check --base origin/main --config .godepsguard.yaml
\`\`\`

### Other Commands
- `godeps-guard scan` - Capture dependency snapshot for current checkout
- `godeps-guard diff` - Compare two independent snapshots using locally saved json models.

## Roadmap
- [ ] Dependency Impact Attribution (Blame & Transitive calculation)
- [ ] PR Comment Mode (`--comment` flag)
- [ ] Historical Tracking
- [ ] Risk Scoring Mechanism
- [ ] SBOM and License Generation
- [ ] Dependency Graph Visualization
