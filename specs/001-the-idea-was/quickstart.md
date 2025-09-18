# Quick Start Guide: Script-to-CLI TUI System

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

### 1. Initialize Configuration
```bash
# Create default configuration
alec init

# Configure script directories
alec config set script_dirs ./scripts ~/.local/bin ~/tools

# Verify configuration
alec config show
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
# Start the TUI
alec

# Or launch with specific directory
alec --scripts-dir ./scripts
```

## Basic Usage

### Interactive Mode (TUI)

#### Navigation
- **Arrow Keys**: Navigate through script tree
- **Enter**: Execute selected script
- **Tab**: Switch between sidebar and main view
- **Space**: Expand/collapse directories
- **/**: Search scripts
- **Esc**: Go back/cancel
- **q**: Quit application

#### Execution
1. Navigate to desired script using arrow keys
2. Press **Enter** to execute
3. View output in real-time
4. Press **Esc** to return to browser when complete

#### Search and Filter
1. Press **/** to enter search mode
2. Type query to filter scripts by name
3. Press **Enter** to select first match
4. Press **Esc** to clear search

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
# Execute by name
alec run hello.sh

# Execute by path
alec run ./scripts/examples/system-info.py

# Execute with timeout
alec run --timeout 30s slow-script.sh
```

#### Refresh Script Cache
```bash
# Refresh all directories
alec refresh

# Refresh specific directory
alec refresh --directory ./scripts
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

# Interactive help
alec  # Then press '?' in TUI
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