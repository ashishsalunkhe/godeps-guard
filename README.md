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
  - Risk score exceeds threshold
  - Blocked licenses are detected

## Installation

```bash
go install github.com/ashishsalunkhe/godeps-guard/cmd/godeps-guard@latest
```

## Usage

### 1. Initialize Configuration

Create a `.godepsguard.yaml` file in the root of your project:

```yaml
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
  max_risk_score: 7

policies:
  blocked_modules:
    - github.com/aws/aws-sdk-go
  blocked_licenses:
    - GPL-3.0
    - AGPL-3.0
  heavy_vendor_patterns:
    - github.com/aws/aws-sdk-go
    - google.golang.org/api
  warn_on_indirect_only_growth: true
  require_reason_for_new_direct_dep: true

report:
  format: markdown
  verbose: true
```

### 2. Run in GitHub Actions CI

See `.github/workflows/godeps-guard.yaml` for a working GitHub actions pipeline.

You can use `godeps-guard` easily via the GitHub Action. Create a `.github/workflows/godeps-guard.yaml`:

```yaml
name: godeps-guard

on:
  pull_request:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run godeps-guard
        uses: ashishsalunkhe/godeps-guard@v1
        with:
          base_ref: origin/${{ github.base_ref }}
          config: .godepsguard.yaml
```

Or manually run:

```bash
godeps-guard check --base origin/main --config .godepsguard.yaml
```

### Commands

| Command | Description |
|---|---|
| `godeps-guard check` | End-to-end CI command: snapshot, diff, policy, report |
| `godeps-guard scan` | Capture dependency snapshot for current checkout |
| `godeps-guard diff` | Compare two independent snapshots |
| `godeps-guard licenses` | Detect licenses of all dependencies |
| `godeps-guard sbom` | Generate CycloneDX SBOM with license data |
| `godeps-guard graph` | Visualize dependency graph (DOT or Mermaid) |
| `godeps-guard history record` | Record a snapshot in history |
| `godeps-guard history report` | Print historical growth report |

### License Detection

```bash
# Table format (default)
godeps-guard licenses

# JSON format
godeps-guard licenses --format json

# Fail CI if any license is undetected
godeps-guard licenses --fail-on-unknown
```

### Graph Visualization

```bash
# Graphviz DOT format (default)
godeps-guard graph --output graph.dot
dot -Tsvg graph.dot -o graph.svg

# Mermaid format (embeddable in GitHub markdown)
godeps-guard graph --format mermaid

# Filter to specific module subtree
godeps-guard graph --filter github.com/spf13
```

### SBOM Generation

```bash
# CycloneDX SBOM with auto-detected licenses
godeps-guard sbom --output sbom.json
```

## Roadmap
- [x] Dependency Impact Attribution (Blame & Transitive calculation)
- [x] PR Comment Mode (`--comment` flag)
- [x] Historical Tracking
- [x] Risk Scoring Mechanism
- [x] SBOM and License Generation
- [x] Dependency Graph Visualization
