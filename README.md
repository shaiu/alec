# Alec

> Terminal UI for organizing and executing scripts without CLI maintenance overhead

Alec is a Script-to-CLI TUI system that automatically discovers scripts in configured directories and presents them through a clean, navigable Terminal User Interface. Focus on writing scripts—not maintaining CLI infrastructure.

[![Release](https://img.shields.io/github/v/release/shaiu/alec)](https://github.com/shaiu/alec/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🚀 **Automatic Script Discovery** - Scans configured directories and discovers shell/Python scripts
- 📁 **Hierarchical Navigation** - Preserves your folder structure with directory-based navigation
- 🎨 **Beautiful TUI** - Built with Bubble Tea framework for a polished terminal experience
- 🔍 **Contextual Search** - Filter scripts within the current directory
- 📝 **Metadata Extraction** - Displays script descriptions, interpreters, and previews
- 🍞 **Breadcrumb Navigation** - Shows current path hierarchy
- ⚡ **Multiple Modes** - Interactive TUI (default) or non-interactive CLI commands
- 🔄 **Manual Refresh** - Press 'r' to refresh script listings
- 🎯 **Clean Exit** - Scripts run with full terminal control

## Installation

### Homebrew (macOS)

```bash
brew tap shaiu/tap
brew install alec
```

### Direct Download

Download the latest release for your platform:

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/shaiu/alec/releases/latest/download/alec_0.1.0_darwin_aarch64.tar.gz -o alec.tar.gz
tar xzf alec.tar.gz
sudo mv alec /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/shaiu/alec/releases/latest/download/alec_0.1.0_darwin_x86_64.tar.gz -o alec.tar.gz
tar xzf alec.tar.gz
sudo mv alec /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/shaiu/alec.git
cd alec
go build -o alec ./cmd/alec
sudo mv alec /usr/local/bin/
```

## Quick Start

### 1. Configuration

Create a configuration file to specify script directories:

```bash
# Create config directory
mkdir -p ~/.config/alec

# Create config file
cat > ~/.config/alec/alec.yaml << EOF
script_dirs:
  - "./scripts"
  - "~/.local/bin"
EOF
```

Or view/edit configuration:

```bash
alec config show    # View current config
alec config edit    # Edit in default editor
alec config reset   # Reset to defaults
```

### 2. Create Sample Scripts

```bash
mkdir -p ./scripts/examples

# Create a shell script
cat > ./scripts/examples/hello.sh << 'EOF'
#!/bin/bash
# Description: Simple hello world script
echo "Hello from Alec!"
EOF
chmod +x ./scripts/examples/hello.sh

# Create a Python script
cat > ./scripts/examples/info.py << 'EOF'
#!/usr/bin/env python3
"""System information script."""
import platform
print(f"Platform: {platform.system()} {platform.release()}")
EOF
chmod +x ./scripts/examples/info.py
```

### 3. Launch Alec

```bash
alec  # Launch interactive TUI
```

## Usage

### Interactive Mode (TUI)

Launch the TUI by running `alec` without arguments:

```bash
alec
```

**Navigation:**
- `↑/↓` or `j/k` - Navigate through directory tree and scripts
- `Enter` - On directory: navigate into it; On script: execute it
- `..` - Navigate up one level
- `/` or `Ctrl+F` - Search within current directory
- `Esc` - Exit search mode
- `r` - Refresh script list
- `q` or `Ctrl+C` - Quit

**UI Features:**
- Breadcrumb row shows current path (e.g., `📁 scripts › database › backups`)
- Script descriptions with preserved line breaks
- Metadata: interpreter, modification time, preview
- Fixed 35-character sidebar width

### Non-Interactive Mode (CLI)

**List Scripts:**
```bash
alec list                                # List all scripts
alec list --directory ./scripts/database # List in specific directory
alec list --details                      # Show detailed information
```

**Execute Scripts:**
```bash
alec run hello.sh                        # Execute by name
alec run ./scripts/examples/info.py      # Execute by path
alec run --dry-run backup.sh             # Show what would be executed
```

**Configuration:**
```bash
alec config show                         # View configuration
alec config edit                         # Edit configuration
alec config reset                        # Reset to defaults
```

**Refresh:**
```bash
alec refresh                             # Refresh all directories
alec refresh --clear-cache               # Clear cache and refresh
```

**Version:**
```bash
alec --version                           # Show version
alec version --detailed                  # Show detailed version info
```

## Configuration

### Configuration File Location

- **macOS**: `~/Library/Application Support/alec/alec.yaml`
- **Linux**: `~/.config/alec/alec.yaml`
- **Windows**: `%APPDATA%/alec/alec.yaml`

### Configuration Options

```yaml
# Script directories to scan
script_dirs:
  - "./scripts"
  - "~/.local/bin"
  - "~/tools"

# Script type mappings
extensions:
  ".sh": "shell"
  ".bash": "shell"
  ".py": "python"
  ".js": "node"

# Execution settings
execution:
  timeout: "5m"
  max_output_size: 1000
  shell: "/bin/bash"

# UI settings
ui:
  theme:
    primary: "#7D56F4"
    secondary: "#EE6FF8"
    focused: "#00FF00"
  show_hidden: false

# Security settings
security:
  allowed_extensions:
    - ".sh"
    - ".bash"
    - ".py"
    - ".js"
  max_execution_time: "10m"

# Logging
logging:
  level: "info"
  file: ""  # Empty for stdout only
```

### Environment Variables

Override config with environment variables (prefix: `ALEC_`):

```bash
export ALEC_SCRIPT_DIRS="./scripts:~/.local/bin"
export ALEC_EXECUTION_TIMEOUT="10m"
export ALEC_LOGGING_LEVEL="debug"
```

## Script Organization

Organize your scripts in a hierarchical structure:

```
./scripts/
├── database/
│   ├── backup.sh
│   ├── restore.sh
│   └── migrate.py
├── deployment/
│   ├── deploy.sh
│   ├── rollback.sh
│   └── health-check.py
└── utilities/
    ├── clean-logs.sh
    └── system-check.py
```

Alec will discover and display them with the folder structure preserved.

## Development

### Tech Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Configuration**: [Viper](https://github.com/spf13/viper)

### Building

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Build for all platforms
make build-all
```

### Project Structure

```
alec/
├── cmd/alec/           # CLI entry point
├── pkg/
│   ├── models/         # Data models
│   ├── services/       # Business logic
│   ├── parser/         # Script metadata extraction
│   ├── tui/           # TUI components
│   └── contracts/     # Interface definitions
├── tests/
│   ├── unit/          # Unit tests
│   ├── contract/      # Contract tests
│   └── integration/   # Integration tests
└── specs/             # Feature specifications
```

### Running Tests

```bash
go test ./tests/unit ./tests/contract    # Run unit and contract tests
go test ./... -v                          # Run all tests with verbose output
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Built with the amazing [Charm](https://charm.sh) libraries:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## Links

- [Documentation](https://github.com/shaiu/alec/tree/main/specs/001-the-idea-was)
- [Issue Tracker](https://github.com/shaiu/alec/issues)
- [Releases](https://github.com/shaiu/alec/releases)
- [Homebrew Tap](https://github.com/shaiu/homebrew-tap)
