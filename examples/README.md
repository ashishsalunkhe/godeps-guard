# godeps-guard Examples

This directory contains examples to help you get started with `godeps-guard`.

## 1. Simple Go App (`examples/simple-go-app`)

A minimal Go project with common dependencies (`cobra`, `yaml.v3`) that you can use to test the detection features.

### Detect Licenses
```bash
# From the project root
./godeps-guard licenses
```

### Generate Graph
```bash
# Generate a Mermaid graph for the example app
./godeps-guard graph --format mermaid
```

### Run Policy Check
```bash
# Record current state as a snapshot
./godeps-guard scan --output head.json
# Compare snapshots
./godeps-guard diff --base head.json --head head.json --format terminal
```

## 2. Configuration (`examples/config`)

Contains a sample `.godepsguard.yaml` showing all available thresholds and risk patterns.

To use a specific config file:
```bash
./godeps-guard check --config examples/config/.godepsguard.yaml --base origin/main
```
