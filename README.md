<div align="center">
<p align="center">
  <img src="GreenTrace.png" alt="GreenTrace Banner" width="100%">
</p>
	
[![CI](https://img.shields.io/github/actions/workflow/status/AppajiDheeraj/GreenTrace/ci.yml?branch=main&logo=github)](https://github.com/AppajiDheeraj/GreenTrace/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AppajiDheeraj/GreenTrace?logo=go)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/AppajiDheeraj/GreenTrace?logo=github)](https://github.com/AppajiDheeraj/GreenTrace/releases)

**GreenTrace is a cross-platform terminal UI that estimates per-process energy and carbon impact, highlights top emitters, and lets you act fast with keyboard-driven controls.**

</div>

## Overview

GreenTrace continuously samples system processes and presents a live dashboard that combines CPU, memory, power, and carbon estimates. It is built for quick inspection during development and ops workflows, especially when you want a fast signal on the most expensive processes running on your machine.

## Key Features

- Live system overview (CPU, RAM, platform, uptime)
- Top carbon process summary
- Process table with CPU, memory, power, carbon, runtime, and path
- Keyboard navigation with a kill action
- Repo-aware filtering to focus on current workspace

## Architecture

GreenTrace is organized as a small CLI with composable internal packages:

- `cmd/` for CLI commands and flags
- `internal/monitor/` for system and process sampling
- `internal/energy/` for power and carbon estimation
- `internal/ui/` for dashboard rendering and interactions
- `internal/repo/` for repo detection and filtering

```mermaid
flowchart LR
	A[CLI Commands] --> B[Monitor]
	B --> C[Energy Estimator]
	B --> D[Repo Filter]
	C --> E[UI Dashboard]
	D --> E
```

## Quick Start

```bash
go build -o GreenTrace
./GreenTrace dashboard
```

## Install and Run

### Option 1: Download the latest release

- Grab the newest build from the GitHub Releases page:
	https://github.com/AppajiDheeraj/GreenTrace/releases/latest

After downloading the binary for your OS, run:

```bash
./GreenTrace dashboard
```

#### One-line auto-download (replace the asset name)

Windows PowerShell:

```powershell
$asset = "GreenTrace_windows_amd64.zip"; $url = "https://github.com/AppajiDheeraj/GreenTrace/releases/latest/download/$asset"; Invoke-WebRequest -Uri $url -OutFile $asset
```

macOS/Linux (curl):

```bash
asset="GreenTrace_darwin_amd64.tar.gz"; curl -L "https://github.com/AppajiDheeraj/GreenTrace/releases/latest/download/$asset" -o "$asset"
```

### Option 2: Clone and run locally

```bash
git clone https://github.com/AppajiDheeraj/GreenTrace.git
cd GreenTrace
go build -o GreenTrace
./GreenTrace dashboard
```

### Option 3: Build release artifacts locally

- Unix/macOS: [scripts/build-release.sh](scripts/build-release.sh)
- Windows PowerShell: [scripts/build-release.ps1](scripts/build-release.ps1)

## Commands

- `GreenTrace dashboard` - launch the interactive dashboard
- `GreenTrace run 10s` - monitor for a fixed duration and print a report
- `GreenTrace top` - show top processes by carbon emissions
- `GreenTrace query chrome` - search running processes by name
- `GreenTrace completion powershell` - generate shell completion script

## Flags

- `--repo-only` (default: false) - restrict process list to the current repository
- `--cpu-tdp` - CPU TDP in watts (default: 65)
- `--emission-factor` - kg CO2 per joule (default: 2e-10)

## Controls (Dashboard)

- Up/Down - navigate processes
- Space - select a process
- `k` - kill selected process
- `q` - quit

## Notes

- On some systems, killing processes may require elevated permissions.
- Process paths can be long and are truncated in the table for readability.

## Development

```bash
make fmt
make test
make build
```

## Release and Tagging

Build release artifacts locally:

```bash
./scripts/build-release.sh
```

On Windows PowerShell:

```powershell
./scripts/build-release.ps1
```

For release notes, use GitHub Releases and summarize key changes.