# Phase 1: Data Model

**Last Updated**: 2025-01-07

## Core Entities

### Script
Represents an executable script file with associated metadata.

**Fields**:
- `ID`: Unique identifier (hash of path + modified time)
- `Name`: Display name (filename without extension)
- `Path`: Absolute file system path
- `Type`: Script type (shell, python, node, etc.)
- `Size`: File size in bytes
- `ModifiedTime`: Last modification timestamp
- `Permissions`: File permissions (readable, executable)
- `IsExecutable`: Boolean flag for execution permission
- `Description`: Optional description from script comments
- `Tags`: Array of user-defined tags
- `Metadata`: Reference to ScriptMetadata entity (if parsed)

**Validation Rules**:
- Path must be within allowed directories (FR-006)
- File must have executable permissions (FR-002)
- Type must be in supported script types list
- Path must pass security validation (no traversal)

**State Transitions**:
```
Discovered → Validated → Ready
Ready → Executing → Completed/Failed
Ready → Refreshing → Ready/Removed
```

### Directory
Represents a directory node in the hierarchical script organization.

**Fields**:
- `Path`: Absolute directory path
- `Name`: Directory name
- `Parent`: Reference to parent directory (nil for root)
- `Children`: Array of child directories
- `Scripts`: Array of scripts directly in this directory
- `IsRoot`: Boolean flag for root-level directories
- `IsExpanded`: UI state for tree expansion
- `ScriptCount`: Total number of executable scripts (including subdirs)
- `LastScan`: Timestamp of last directory scan

**Validation Rules**:
- Path must be within configured script directories
- Must have read permissions
- Children must maintain parent references
- ScriptCount must be consistent with actual script count

**Relationships**:
- Parent-child hierarchy with Directory entities
- Contains multiple Script entities
- Root directories configured in Configuration entity

### ExecutionSession
Represents a single script execution instance. Note: In current implementation, the TUI exits when a script is executed, so execution sessions are primarily used in CLI mode.

**Fields**:
- `SessionID`: Unique session identifier (UUID)
- `Script`: Reference to executed script
- `Status`: Current execution status (pending, running, completed, failed, timeout)
- `StartTime`: Execution start timestamp
- `EndTime`: Execution completion timestamp (nil if running)
- `Duration`: Total execution time
- `ExitCode`: Process exit code (nil if running)
- `Output`: Array of output lines (stdout + stderr combined)
- `ErrorMessage`: Error message if execution failed
- `Context`: Cancellation context for process control

**Validation Rules**:
- SessionID must be unique across all sessions
- Script reference must be valid and executable
- Status transitions must follow valid state machine
- StartTime must be set before EndTime

**State Transitions**:
```
Pending → Running → [Completed | Failed | Timeout]
Running → Cancelled → Failed
```

**Note on TUI Execution**: When executing from the TUI, the application exits and the script runs with full terminal control, so session tracking is not maintained. ExecutionSession is primarily used for CLI mode (`alec run`) execution.

### Configuration
Represents user and system configuration settings.

**Fields**:
- `ScriptDirectories`: Array of directory paths to scan
- `ScriptExtensions`: Map of file extensions to script types
- `ExecutionTimeout`: Maximum script execution time
- `MaxOutputLines`: Maximum output buffer size per execution
- `RefreshInterval`: Manual refresh capability (no auto-refresh in v1)
- `DefaultShell`: Default shell for script execution
- `UITheme`: Theme settings for TUI appearance
- `KeyBindings`: Custom key binding configurations
- `SecuritySettings`: Security policy settings
- `LogLevel`: Logging verbosity level

**Validation Rules**:
- ScriptDirectories must contain valid, accessible paths
- ScriptExtensions must contain valid file extensions
- ExecutionTimeout must be positive duration
- MaxOutputLines must be positive integer
- SecuritySettings must enforce path restrictions

**Default Values**:
- ExecutionTimeout: 5 minutes
- MaxOutputLines: 1000
- ScriptExtensions: {".sh": "shell", ".py": "python", ".js": "node"}
- LogLevel: "info"

### UIState
Represents the current state of the Terminal User Interface.

**Fields**:
- `CurrentView`: Active view (welcome, script details)
- `SelectedScript`: Currently selected script
- `CurrentPath`: Current directory being viewed
- `TerminalWidth`: Current terminal width
- `TerminalHeight`: Current terminal height
- `SidebarWidth`: Fixed at 35 characters to prevent layout shifts (updated from 24 to accommodate longer script names)
- `MainWidth`: Calculated as terminal width minus sidebar width
- `BreadcrumbPath`: Current breadcrumb trail for navigation context
- `FocusedComponent`: Currently focused UI component (sidebar always focused in v1)
- `SearchMode`: Boolean indicating if search mode is active
- `SearchQuery`: Current search/filter query string
- `FilteredScripts`: Scripts matching current search within directory context

**Validation Rules**:
- Terminal dimensions must be positive
- Width calculations must sum to terminal width
- SelectedScript must be valid if not nil
- SearchQuery must be valid string for case-insensitive filtering

**Calculated Fields**:
- SidebarWidth: Fixed at 35 characters (not calculated, updated from 24 in implementation)
- MainWidth: TerminalWidth - SidebarWidth - borders/padding
- IsResponsive: Boolean indicating if terminal is large enough (minimum 80x24)

### ScriptMetadata
Represents parsed metadata extracted from script file content using dedicated lexer/parser components.

**Fields**:
- `Interpreter`: Detected interpreter from shebang line (e.g., "/bin/bash", "/usr/bin/env python3")
- `Description`: Extracted description from comments or docstrings (preserves line breaks)
- `FullContent`: Complete script content or truncated preview
- `PreviewLines`: Number of lines in preview
- `LineCount`: Total line count in full script
- `IsTruncated`: Boolean flag indicating if preview is truncated
- `DescriptionMaxChars`: Maximum characters for description (default: 300)
- `MaxPreviewLines`: Maximum lines for preview display (default: 50)
- `FullScriptThreshold`: Line count threshold for full vs preview (default: 100)

**Validation Rules**:
- Description must preserve original line breaks (newlines, not spaces)
- FullContent must not exceed memory limits
- LineCount must match actual script lines
- IsTruncated correctly reflects preview state

**Parser Rules** (Shell Scripts):
- Extract description from header comment block
- Skip shebang line (#!)
- Support custom markers: "# Description:", "# @desc", "# Summary:"
- Join multi-line comments with newlines (not spaces)

**Parser Rules** (Python Scripts):
- Prioritize module docstrings (""" or ''')
- Fallback to header comment block if no docstring
- Extract from first 20 lines only
- Join docstring lines with newlines

### BreadcrumbModel
Represents the navigation breadcrumb display component.

**Fields**:
- `Breadcrumbs`: Formatted breadcrumb trail string
- `Width`: Available width for breadcrumb display
- `Height`: Fixed height (typically 2 lines: content + border)

**Validation Rules**:
- Must fit within terminal width
- Path parts separated by " › " character
- Root directory shows base script directory name

**Display Format**:
```
📁 scripts › database › backups
```

## Entity Relationships

```
Configuration 1:N Directory (script directories)
Directory 1:N Directory (parent-child hierarchy)
Directory 1:N Script (contains scripts)
Script 1:1 ScriptMetadata (parsed metadata)
Script 1:N ExecutionSession (execution history)
UIState N:1 Script (current selection)
UIState N:1 Directory (current selection)
UIState N:1 ExecutionSession (current execution)
UIState 1:1 BreadcrumbModel (navigation display)
```

## Data Flow Patterns

### Script Discovery Flow
```
Configuration.ScriptDirectories → Directory.Scan() → Script.Validate() → Lexer.ExtractMetadata() → Script.Ready
```

### Metadata Extraction Flow
```
Script.Path → Lexer.DetectType() → ShellLexer/PythonLexer.ExtractMetadata() → ScriptMetadata → Script.Metadata
```

### Execution Flow
```
UIState.SelectedScript → ExecutionSession.Create() → ExecutionSession.Execute() → ExecutionSession.Complete()
```

### Navigation Flow
```
UIState.CurrentPath → Directory.BuildItems() → UIState.NavigationItems → User Selection → UIState.SelectedScript
```

### Search Flow
```
User presses '/' → UIState.SearchMode = true → User types query → UIState.SearchQuery → Filter scripts in CurrentPath → UIState.FilteredScripts → Display filtered results
```

### Refresh Flow
```
UIState.RefreshTrigger → Directory.Rescan() → Script.UpdateModifiedTime() → Script.Revalidate()
```

## Validation and Constraints

### Cross-Entity Validation
- Selected entities in UIState must exist in their respective collections
- ExecutionSession.Script must reference valid Script entity
- Directory parent-child relationships must be acyclic
- Script paths must be unique within the system

### Performance Constraints
- Directory scanning limited to configured depths
- Script collection limited by available memory
- Output buffering limited by MaxOutputLines
- UI updates throttled to maintain responsiveness

### Security Constraints
- All paths validated against path traversal attacks
- Script execution restricted to configured directories
- File permissions verified before execution attempts
- Configuration changes require validation

## Data Persistence

### In-Memory Storage
- UIState: Ephemeral, not persisted
- ExecutionSession: Recent sessions cached in memory
- Directory/Script trees: Rebuilt on startup

### File-Based Storage
- Configuration: YAML file in user config directory
- Execution history: Optional JSON log file
- Cache data: Optional for large directory structures

### No Database Required
All data is file-system based or ephemeral, eliminating database dependency and simplifying deployment.