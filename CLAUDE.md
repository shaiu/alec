# Alec - Script-to-CLI TUI System

A Terminal User Interface (TUI) application built with Go and Bubble Tea that automatically discovers scripts in configured directories and presents them through a clean, navigable interface for execution.

## Project Overview

**Purpose**: Enable developers to organize scripts in directory structures and have them automatically available as commands in a TUI, eliminating CLI maintenance overhead.

**Core Value**: Focus on writing scripts instead of maintaining CLI infrastructure.

## Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lip Gloss (github.com/charmbracelet/lipgloss)
- **Tree Navigation**: github.com/Digital-Shane/treeview
- **Configuration**: Viper + Cobra
- **Testing**: Go standard testing with table-driven tests
- **Target**: Cross-platform terminal applications

## Architecture

### Component Structure
```
RootModel (Router/Compositor)
├── HeaderModel (Status, breadcrumbs)
├── SidebarModel (Script tree navigation)
├── MainModel (Content area)
│   ├── ScriptListModel (Filtered results)
│   ├── PreviewModel (Script preview)
│   └── ExecutorModel (Output display)
└── FooterModel (Help, status)
```

### Core Entities
- **Script**: Executable file with metadata (path, type, permissions, execution status)
- **Directory**: Hierarchical organization with parent-child relationships
- **ExecutionSession**: Single script execution with output, status, duration
- **Configuration**: User settings (directories, timeouts, UI preferences)
- **UIState**: Current TUI state (view, focus, dimensions, selections)

### Key Interfaces
- **ScriptDiscovery**: Directory scanning, script validation, filtering
- **ScriptExecutor**: Execution management, output streaming, cancellation
- **TUIManager**: State management, navigation, responsive layout
- **ConfigManager**: Configuration loading, validation, environment integration

## Development Patterns

### Bubble Tea Architecture
- **Model Tree**: Hierarchical models with message routing
- **State Isolation**: Each model manages its own state
- **Message-Driven**: Use Bubble Tea's message system for communication
- **Focus Management**: Clear focus states and visual indicators

### Script Execution
- **Context-Based**: Use `exec.CommandContext` for cancellation/timeout
- **Streaming Output**: Real-time output with unbuffered streaming
- **Security**: Path validation, permission checks, directory restrictions
- **Cross-Platform**: Shell detection (bash/sh on Unix, cmd on Windows)

### Configuration
- **Hierarchy**: CLI flags > env vars > config file > defaults
- **Locations**: OS-specific config directories
- **Validation**: Path existence, permission verification
- **Environment**: ALEC_ prefix for environment variables

## Key Features

### Functional Requirements (from spec)
- **FR-001**: Automatic script discovery in configured directories
- **FR-002**: Support shell scripts and Python scripts
- **FR-003**: Terminal User Interface for navigation/selection
- **FR-004**: Preserve hierarchical folder structure
- **FR-005**: Execute scripts and capture output
- **FR-006**: Configurable directory scanning
- **FR-007**: Manual refresh mechanism
- **FR-008**: Graceful error handling with meaningful messages
- **FR-009**: Basic execution without argument passing (v1)
- **FR-010**: Execute with current user permissions
- **FR-011**: Interactive (TUI) and non-interactive (CLI) modes
- **FR-012**: Responsive rendering based on terminal size
- **FR-013**: Clean, minimal UI for daily developer use
- **FR-014**: Easy installation and simple update process

### Performance Targets
- Sub-100ms script discovery
- <16ms UI updates for responsive feel
- Support hundreds of scripts in nested structures
- Golden ratio layouts (38% sidebar, 62% main content)

### Security
- Path traversal prevention using `filepath.IsLocal`
- Directory restriction enforcement
- File extension validation
- No privilege escalation (user permissions only)

## File Structure

```
/
├── src/
│   ├── models/          # Data models and entities
│   ├── services/        # Business logic services
│   ├── cli/            # CLI command handlers
│   └── lib/            # Shared libraries
├── tests/
│   ├── contract/       # Contract tests
│   ├── integration/    # Integration tests
│   └── unit/          # Unit tests
├── specs/001-the-idea-was/
│   ├── spec.md         # Feature specification
│   ├── plan.md         # Implementation plan
│   ├── research.md     # Technical research
│   ├── data-model.md   # Data model design
│   ├── quickstart.md   # User guide
│   └── contracts/      # Go interface contracts
└── CLAUDE.md          # This file
```

## Current Development Status

**Branch**: `001-the-idea-was`
**Phase**: Implementation Planning Complete
**Next**: Task generation and implementation

### Completed
- ✅ Feature specification with 14 functional requirements
- ✅ Technical research (Bubble Tea, Go patterns, security)
- ✅ Data model design (5 core entities with relationships)
- ✅ API contracts (4 interfaces with comprehensive requirements)
- ✅ Quickstart guide with usage examples
- ✅ Constitutional review (no violations)

### Next Steps
1. Generate implementation tasks with `/tasks` command
2. Implement core interfaces following TDD approach
3. Build TUI components with responsive layouts
4. Integrate script discovery and execution
5. Add configuration management
6. Testing and validation

## Development Guidelines

### Code Style
- Follow Go conventions and idioms
- Use table-driven tests for comprehensive coverage
- Implement interfaces before concrete types
- Leverage Go 1.21+ features (filepath.IsLocal, context patterns)

### Testing Approach
- TDD mandatory: Tests → Implementation
- Contract tests for interface compliance
- Integration tests for end-to-end workflows
- Unit tests for business logic

### UI Principles
- Clean, minimal design (no excessive icons/emojis)
- Responsive to terminal size changes
- Clear navigation with breadcrumbs
- Real-time feedback for long operations

### Security Requirements
- Validate all file paths against traversal attacks
- Restrict execution to configured directories
- Enforce file permission checks
- No command injection vulnerabilities

This context provides everything needed to continue development of the Script-to-CLI TUI System following established patterns and requirements.