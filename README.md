# godeps-guard

> Prevent dependency sprawl and binary bloat in Go services before it reaches production.

A Go dependency impact analyzer and CI enforcement tool.

## Installation

```bash
go install github.com/ashishsalunkhe/godeps-guard/cmd/godeps-guard@latest
```

## Quick Start & Examples

We provide a set of [examples](./examples) to help you explore the features immediately after installation.

1. **Try License Detection**: `godeps-guard licenses`
2. **Visualize your Graph**: `godeps-guard graph --format mermaid`
3. **Check your Policy**: `godeps-guard check --base origin/main`

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
- [x] AI-Powered Analysis (`--ai` flag)

## AI Features

`godeps-guard` has optional AI-powered analysis. All AI features are **opt-in** — the tool works identically without an API key.

### Setup

You can set these environment variables in your shell or use a `.env` file in your project root (see `.env.example`).

```bash
export GODEPS_GUARD_AI_PROVIDER=gemini   # gemini (default) | openai
export GODEPS_GUARD_AI_KEY=<your-key>
export GODEPS_GUARD_AI_MODEL=gemini-2.0-flash  # optional
```

### Feature 1: AI PR Review Summary
Narrates the dependency impact in plain English, replacing raw numbers with a 3-5 sentence assessment and an overall recommendation.

```bash
godeps-guard check --base origin/main --ai
```

### Feature 2: Smart Risk Scoring
Enhances the static risk score for each new dependency using the LLM's knowledge of CVE history, ecosystem reputation, and maintenance status. The static score is the floor; AI can only raise it.

```bash
godeps-guard check --base origin/main --ai  # risk scores automatically enhanced
```

### Feature 3: Requirement Validator
When `require_reason_for_new_direct_dep: true` is set, AI validates whether the implicit justification for each new dependency actually makes sense (checks for overkill, stdlib alternatives, etc.).

```bash
# Enable in .godepsguard.yaml:
# policies:
#   require_reason_for_new_direct_dep: true

godeps-guard check --base origin/main --ai
```

### Feature 4: Alternative Suggestion Engine
Suggests lighter alternatives for heavy new dependencies and appends a "💡 Suggested Lighter Alternatives" section to the report.

```bash
godeps-guard check --base origin/main --ai  # alternatives appear in report
```

### Feature 5: Historical Trend Anomaly Detection
Analyzes your recorded dependency history and produces a narrative anomaly report identifying unusual growth periods.

```bash
godeps-guard history report --ai
```

### Feature 6: Natural Language Policy Config
Describe your policy in plain English and have AI generate a `.godepsguard.yaml` for you.

```bash
godeps-guard init --ai
# > Block GPL licenses, warn if binary grows more than 5MB, flag AWS SDKs
```

### AI config block in `.godepsguard.yaml`

```yaml
ai:
  provider: gemini          # gemini | openai
  model: gemini-2.0-flash   # optional
  features:
    pr_summary: true
    smart_risk: true
    validate_reason: true
    suggest_alternatives: true
    trend_analysis: true
```
