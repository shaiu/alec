# Alec CLI Constitution

## Core Principles

### I. Go-First Architecture
Every component must follow Go idioms and conventions; Interfaces define contracts before implementations; Error handling via explicit error returns (no exceptions); Use Go 1.21+ features including `filepath.IsLocal` for security; Standard library preferred over external dependencies when suitable; Package organization follows Go project layout standards.

### II. Bubble Tea Framework Adherence
Model-View-Update (MVU) pattern strictly enforced; All UI state changes through Bubble Tea messages; No direct goroutine management - use `tea.Cmd` for async operations; Components must implement `tea.Model` interface; Message passing for inter-component communication; Responsive design with terminal size awareness via `tea.WindowSizeMsg`.

### III. Test-Driven Development (NON-NEGOTIABLE)
Contract tests written first for all interfaces; Integration tests before feature implementation; Unit tests for all business logic; Tests must fail before implementation begins; Table-driven tests for comprehensive coverage; Minimum 80% code coverage maintained; Performance tests for sub-100ms discovery and <16ms UI response times.

### IV. Security-First Design
Path traversal prevention using `filepath.IsLocal`; Input validation for all user-provided paths; Script execution limited to configured directories; No privilege escalation - use current user permissions; File permission verification before execution; Configuration validation prevents malicious inputs; Secure defaults in all configuration options.

### V. CLI/TUI Separation
Clear separation between CLI commands and TUI components; CLI commands for non-interactive automation; TUI for interactive script browsing and execution; Both modes share common service layer; Configuration unified across both interfaces; Error handling appropriate for each mode (JSON vs. user-friendly).

### VI. Performance and Responsiveness
Script discovery must complete under 100ms for typical directories; UI updates must render under 16ms for responsive feel; Memory usage limited to prevent system impact; Large output streams properly buffered and truncated; Background operations use context for cancellation; Graceful degradation for large directory structures.

### VII. Cross-Platform Compatibility
Support Linux, macOS, and Windows consistently; Shell detection adapts to platform (bash/sh vs cmd); File path handling uses `filepath` package exclusively; Terminal capabilities detected and adapted; Installation methods for all major platforms; Testing on multiple operating systems required.

## Technical Constraints

### Dependency Management
**Required Dependencies**: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/spf13/viper`, `github.com/spf13/cobra`; **Optional Dependencies**: `github.com/Digital-Shane/treeview` for tree navigation; **Forbidden Dependencies**: GUI frameworks, database drivers, network libraries (unless explicitly needed); All dependencies must be actively maintained with Go module support; Vendor directory not used - rely on Go module proxy.

### Configuration Standards
YAML primary format with JSON/TOML support; Environment variables with `ALEC_` prefix override file settings; CLI flags override environment variables; XDG Base Directory specification compliance on Unix; OS-specific config locations (Windows: `%APPDATA%`, macOS: `~/Library/Application Support`); Configuration validation on load with helpful error messages; Default configuration provides working system without user input.

### Error Handling Protocol
All errors must include actionable context; User-facing errors must be non-technical and helpful; System errors logged with full technical details; Error types categorized: User, System, Security, Performance; Graceful degradation preferred over application crashes; Configuration errors must suggest corrections; File operation errors must indicate permissions or existence issues.

### Logging and Observability
Structured logging using standard library `log/slog`; Log levels: DEBUG, INFO, WARN, ERROR; Development: DEBUG to stderr; Production: INFO+ to file with rotation; No sensitive information in logs (paths only, no content); Performance metrics for script discovery and execution times; User actions logged for debugging support; Configurable log output destination.

## Development Workflow

### Code Organization
```
cmd/alec/           # CLI entry point
src/
├── models/         # Data structures and entities
├── services/       # Business logic layer
├── tui/           # Bubble Tea components
├── cli/           # Cobra CLI commands
└── lib/           # Shared utilities
tests/
├── contract/      # Interface compliance tests
├── integration/   # End-to-end scenarios
├── unit/         # Component-level tests
└── performance/  # Benchmark tests
```

### Testing Requirements
Contract tests verify interface implementations; Integration tests cover user workflows from quickstart guide; Unit tests achieve 80% minimum coverage; Performance tests validate response time requirements; Tests run on CI for Linux, macOS, Windows; Manual testing checklist for release validation; Benchmarks track performance regression.

### Code Quality Standards
`golangci-lint` configuration enforced in CI; `gofmt` and `goimports` for consistent formatting; Cyclomatic complexity limited to 10 per function; No functions exceeding 50 lines; Package documentation required for all public APIs; Comments explain "why" not "what"; Code review required for all changes.

### Release Process
Semantic versioning (MAJOR.MINOR.PATCH); Feature branches for all development; GitHub releases with binaries for all platforms; Installation via package managers (Homebrew, apt, chocolatey); Backward compatibility maintained for configuration files; Migration guides for breaking changes; Performance regression testing before release.

## Security Requirements

### Script Execution Security
No shell expansion or command injection vulnerabilities; Path validation prevents directory traversal; Script type detection by file extension only; Execution timeout prevents runaway processes; Process cleanup prevents resource leaks; User permission inheritance (no sudo/elevation); Output size limits prevent memory exhaustion.

### Configuration Security
Configuration files created with 0600 permissions; No secrets stored in configuration files; Environment variable validation; Path canonicalization before use; Input sanitization for all user-provided values; Default configuration provides minimal attack surface; Configuration backup/restore with integrity checking.

### File System Access
Read-only access outside configured directories; Write access only for configuration and logs; Symlink following prevented in script directories; Hidden file access controlled by configuration; Temporary file handling with secure cleanup; File locking for concurrent access prevention.

## Governance

### Constitutional Authority
This constitution supersedes all other development practices; Amendments require documented justification and team approval; Violations must be justified in complexity tracking; Code review verifies constitutional compliance; Automated checks enforce core principles where possible.

### Quality Gates
All pull requests must pass contract tests; Integration tests required for new features; Performance tests must not regress; Security review for all file system operations; Documentation updates for user-facing changes; Breaking changes require migration path.

### Exception Process
Constitutional violations require documented justification; Alternative approaches must be evaluated and rejected; Security exceptions require additional review; Performance exceptions need benchmarking evidence; All exceptions tracked in implementation plans.

**Version**: 1.0.0 | **Ratified**: 2025-09-18 | **Last Amended**: 2025-09-18