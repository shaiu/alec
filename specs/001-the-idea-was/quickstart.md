# Quick Start Guide: Script-to-CLI TUI System

**Last Updated**: 2025-01-07
**Implementation Status**: All features documented here are fully implemented

## Installation

### Download and Install
```bash
# Download latest release
curl -L https://github.com/your-org/alec/releases/latest/download/alec-linux-amd64 -o alec
chmod +x alec
sudo mv alec /usr/local/bin/

# Or install via package manager
brew install alec          # macOS
apt install alec           # Ubuntu/Debian
choco install alec         # Windows
```

### Build from Source
```bash
git clone https://github.com/your-org/alec.git
cd alec
go build -o alec ./cmd/alec
./alec --version
```

## First Run

### 1. View and Configure
```bash
# View current configuration
alec config show

# Edit configuration file
alec config edit

# Reset to defaults
alec config reset
```

### 2. Prepare Script Directory
```bash
# Create sample scripts directory
mkdir -p ./scripts/examples

# Create a sample shell script
cat > ./scripts/examples/hello.sh << 'EOF'
#!/bin/bash
echo "Hello from script: $(basename $0)"
echo "Current time: $(date)"
echo "Script path: $(realpath $0)"
EOF

chmod +x ./scripts/examples/hello.sh

# Create a sample Python script
cat > ./scripts/examples/system-info.py << 'EOF'
#!/usr/bin/env python3
import platform
import sys
import os

print(f"Python version: {sys.version}")
print(f"Platform: {platform.system()} {platform.release()}")
print(f"Architecture: {platform.machine()}")
print(f"Current directory: {os.getcwd()}")
EOF

chmod +x ./scripts/examples/system-info.py
```

### 3. Launch Interactive Mode
```bash
# Start the TUI (default command when run without arguments)
alec

# Or launch with specific directories
alec -d ./scripts -d ~/tools
```

## Basic Usage

### Interactive Mode (TUI)

#### Navigation
- **↑/↓ or k/j**: Navigate through directory tree and scripts
- **Enter**:
  - On a directory: Navigate into the directory
  - On a script: Execute the script (app will exit and script runs)
  - On "..": Navigate up one level
- **/** or **Ctrl+F**: Enter search mode (contextual - searches within current directory)
- **Esc**: Exit search mode
- **r**: Refresh script list
- **q** or **Ctrl+C**: Quit application

**UI Features**:
- Breadcrumb navigation row shows current path (e.g., "📁 scripts › database › backups")
- Script descriptions with preserved line breaks displayed in main content
- Script metadata includes interpreter, modification time, and preview
- Sidebar width: 35 characters (wide enough for most script names)

#### Execution
1. Navigate to desired script using ↑/↓ or k/j
2. Press **Enter** to execute
3. The TUI exits and the script runs with full terminal control
4. After script completes, you return to your shell

#### Search and Filter (Contextual)
1. Press **/** or **Ctrl+F** to enter search mode
2. Type query to filter scripts by name (searches within current directory context)
3. Use **↑/↓** or **j/k** to navigate filtered results
4. Press **Enter** to execute selected script
5. Press **Esc** to exit search and return to normal navigation

### Non-Interactive Mode (CLI)

#### List Available Scripts
```bash
# List all scripts
alec list

# List scripts in specific directory
alec list --directory ./scripts/database

# List with details
alec list --details
```

#### Execute Scripts Directly
```bash
# Execute by name (searches configured directories)
alec run hello.sh

# Execute by path
alec run ./scripts/examples/system-info.py

# Execute with dry run (shows what would be executed)
alec run --dry-run backup.sh
```

#### Refresh Script Cache
```bash
# Refresh all configured directories
alec refresh

# Refresh with cache clearing
alec refresh --clear-cache

# Refresh specific directories
alec refresh -d ./scripts -d ~/tools
```

## Configuration

### Configuration File Location
- **Linux**: `~/.config/alec/alec.yaml`
- **macOS**: `~/Library/Application Support/alec/alec.yaml`
- **Windows**: `%APPDATA%/alec/alec.yaml`

### Sample Configuration
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
  ".rb": "ruby"

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
  layout:
    sidebar_ratio: 0.382
    min_terminal_width: 80
  show_hidden: false
  confirm_on_execute: false

# Security settings
security:
  allowed_extensions:
    - ".sh"
    - ".bash"
    - ".py"
    - ".js"
  max_execution_time: "10m"
  max_output_size: 10000

# Logging
logging:
  level: "info"
  file: ""  # Empty for stdout only
```

### Environment Variables
```bash
export ALEC_SCRIPT_DIRS="./scripts:~/.local/bin"
export ALEC_EXECUTION_TIMEOUT="10m"
export ALEC_LOGGING_LEVEL="debug"
export ALEC_UI_THEME_PRIMARY="#FF0000"
```

## Common Workflows

### Development Workflow
1. **Organize scripts by project**:
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

2. **Launch TUI and navigate** to project-specific scripts
3. **Execute scripts** with real-time output monitoring
4. **Switch between projects** using directory navigation

### Operations Workflow
1. **Centralize operational scripts** in `~/.local/bin`
2. **Use non-interactive mode** for automation:
   ```bash
   alec run backup-database.sh
   alec run --timeout 30m long-maintenance.py
   ```
3. **Monitor execution** through TUI when needed

### Team Workflow
1. **Share configuration** via version control:
   ```bash
   # Add to project root
   cat > .alec.yaml << EOF
   script_dirs:
     - "./scripts"
     - "./tools"
   EOF
   ```
2. **Standardize script organization** across team
3. **Use environment variables** for user-specific paths

## Troubleshooting

### Script Not Found
```bash
# Check configuration
alec config show

# Refresh script cache
alec refresh

# List discovered scripts
alec list --details
```

### Execution Permissions
```bash
# Fix permissions for all scripts
find ./scripts -name "*.sh" -exec chmod +x {} \;
find ./scripts -name "*.py" -exec chmod +x {} \;
```

### Path Issues
```bash
# Use absolute paths in config
alec config set script_dirs /home/user/scripts /usr/local/scripts

# Check current working directory
pwd
```

### Performance Issues
```bash
# Reduce scan scope
alec config set security.max_output_size 500

# Check large directories
du -sh ~/.local/bin/*
```

## Getting Help

### Built-in Help
```bash
# General help
alec --help

# Command-specific help
alec run --help
alec config --help
alec list --help
alec refresh --help

# Demo mode (shows system status)
alec demo
```

### Version Information
```bash
alec --version
alec version --detailed
```

### Debug Mode
```bash
# Enable debug logging
ALEC_LOGGING_LEVEL=debug alec

# Save debug output
alec --log-file debug.log
```

This quickstart guide covers the essential usage patterns for the Script-to-CLI TUI System. For advanced configuration and troubleshooting, refer to the full documentation.