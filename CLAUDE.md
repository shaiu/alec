# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go CLI application built with Charm libraries:
- **Bubble Tea** for the TUI framework
- **Lipgloss** for styling
- **Cobra** for CLI structure

## Build Commands

```bash
# Build the application
go build -o alec

# Run the application
go run .

# Install dependencies
go mod tidy

# Format code
go fmt ./...

# Run tests (when added)
go test ./...
```

## Project Structure

- `main.go` - CLI entry point and Cobra setup
- `model.go` - Bubble Tea model and TUI logic
- `config.go` - Configuration management
- `filesystem.go` - Directory reading and file type detection
- `executor.go` - Shell script execution
- `go.mod` - Go module dependencies

## Usage

The application navigates directories and executes shell scripts:
- Use `--dir` flag to specify custom root directory
- Directories are shown in green with folder icons
- Shell scripts (.sh, .bash, .zsh) are shown in red with rocket icons
- Hidden files are ignored