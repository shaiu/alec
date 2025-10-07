# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Automated release mechanism with GoReleaser
- GitHub Actions workflow for releases
- Homebrew tap support for easy installation
- Makefile targets for release management

### Changed
- Updated module path from `github.com/your-org/alec` to `github.com/shaiu/alec`
- Updated all import paths throughout the codebase

### Fixed

### Removed

## [0.1.0] - 2025-01-07

### Added
- Initial implementation of Script-to-CLI TUI System
- Terminal User Interface with Bubble Tea framework
- Automatic script discovery in configured directories
- Support for shell scripts (.sh, .bash) and Python scripts (.py)
- Hierarchical directory navigation
- Script metadata extraction with parser/lexer system
- Breadcrumb navigation component
- Contextual search and filtering
- CLI commands: list, run, config, refresh, version, demo
- Configuration management with YAML support
- Security validation for script paths and permissions
- Responsive layout with fixed 35-character sidebar width
- Line break preservation in multi-line script descriptions
- Cross-platform support (macOS focus)

### Core Features
- FR-001 to FR-015: Core functionality complete
- FR-016 to FR-019: Enhanced features (metadata, breadcrumbs, line breaks)

---

## Release Process

To create a new release:

1. Update the `[Unreleased]` section with your changes
2. Create a new version section with the date
3. Create and push a git tag:
   ```bash
   make tag VERSION=x.y.z
   git push origin vx.y.z
   ```
4. GitHub Actions will automatically build and publish the release
5. Homebrew tap will be updated automatically

## Installation

### Homebrew (macOS)
```bash
brew tap shaiu/alec
brew install alec
```

### From Release
Download binaries from [GitHub Releases](https://github.com/shaiu/alec/releases)

### From Source
```bash
git clone https://github.com/shaiu/alec.git
cd alec
make build
```
