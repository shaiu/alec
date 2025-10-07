# Feature Specification: Script-to-CLI TUI System

**Feature Branch**: `001-the-idea-was`
**Created**: 2025-09-18
**Last Updated**: 2025-01-07
**Status**: Implemented
**Input**: User description: "The idea was to have a TUI (Terminal UI) that could enable developers to write their own CLI without worrying about writing the actual CLI. This way, developers can focus on writing their usual shell scripts (or Python scripts) and pick it up in the CLI.

Essentially, you should have a dedicated directory organized in a structured hierarchy of folders and scripts to hold all the necessary bash scripts.

Then the CLI picks up the scripts in the folders you choose and incorporates them into the CLI. This means that each script becomes a command that you can run from within the CLI. Once the scripts are detected, they are registered as commands in the CLI. This involves mapping each script to a corresponding command invoked from the CLI.

When you navigate through the CLI and select a specific script, the script is executed. This is done by invoking the underlying script file and capturing its output. The CLI gives developers an accessible way to run their scripts. They just need to add their script to a folder; voila, it's ready to use using the CLI.

Now the developer can focus on what is really important: developing their new features and not maintain [how to add/update various CLIs to use]

This reduced our time planning new features because we knew the usual overhead was gone."

## Execution Flow (main)
```
1. Parse user description from Input
   - If empty: ERROR "No feature description provided"
2. Extract key concepts from description
   - Identify: actors, actions, data, constraints
3. For each unclear aspect:
   - Mark with [NEEDS CLARIFICATION: specific question]
4. Fill User Scenarios & Testing section
   - If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   - Each requirement must be testable
   - Mark ambiguous requirements
6. Identify Key Entities (if data involved)
7. Run Review Checklist
   - If any [NEEDS CLARIFICATION]: WARN "Spec has uncertainties"
   - If implementation details found: ERROR "Remove tech details"
8. Return: SUCCESS (spec ready for planning)
```

---

## Quick Guidelines
-  Focus on WHAT users need and WHY
-  Avoid HOW to implement (no tech stack, APIs, code structure)
-  Written for business stakeholders, not developers

### Section Requirements
- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

### For AI Generation
When creating this spec from a user prompt:
1. **Mark all ambiguities**: Use [NEEDS CLARIFICATION: specific question] for any assumption you'd need to make
2. **Don't guess**: If the prompt doesn't specify something (e.g., "login system" without auth method), mark it
3. **Think like a tester**: Every vague requirement should fail the "testable and unambiguous" checklist item
4. **Common underspecified areas**:
   - User types and permissions
   - Data retention/deletion policies
   - Performance targets and scale
   - Error handling behaviors
   - Integration requirements
   - Security/compliance needs

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
As a developer, I want to organize my scripts in a directory structure and have them automatically available as commands in a Terminal UI, so that I can focus on writing scripts instead of maintaining CLI infrastructure.

### Acceptance Scenarios
1. **Given** I have a directory with organized shell/Python scripts, **When** I run `alec` without arguments, **Then** I see the TUI with all my scripts in a directory tree
2. **Given** I add a new script to my designated script directory, **When** I press 'r' to refresh in the TUI, **Then** the new script appears in the tree
3. **Given** I navigate to a specific script in the TUI, **When** I press Enter to execute it, **Then** the TUI exits and the script runs with full terminal control
4. **Given** I organize scripts in a folder hierarchy, **When** I browse the TUI, **Then** I can navigate into directories with Enter and up with ".." to find scripts
5. **Given** I have both shell scripts and Python scripts in my directory, **When** I use the TUI, **Then** both script types are detected and shown with appropriate icons
6. **Given** I am viewing a directory in the TUI, **When** I press '/' to search, **Then** I can filter scripts within that directory context
7. **Given** I want to run a script without the TUI, **When** I run `alec run script-name.sh`, **Then** the script executes directly in CLI mode

### Edge Cases
- What happens when a script has execution permission issues?
- How does the system handle scripts that require interactive input?
- What occurs when a script has long execution times?
- How are script errors and failures displayed to the user?
- What happens when the script directory structure changes while the TUI is running?
- How are multi-line descriptions with line breaks displayed in the preview?
- What happens when script names exceed the sidebar width?
- How does the breadcrumb display handle deeply nested directory structures?

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: System MUST automatically discover scripts in a configured directory structure
- **FR-002**: System MUST support both shell scripts and Python scripts as executable commands (extensible to other types)
- **FR-003**: System MUST provide a Terminal User Interface for script navigation and selection, launched by running the command without arguments
- **FR-004**: System MUST preserve the hierarchical folder structure when displaying scripts in the TUI with directory-based navigation
- **FR-005**: System MUST execute selected scripts and exit the application, passing control to the script
- **FR-006**: System MUST allow users to configure which directories to scan for scripts
- **FR-007**: System MUST refresh script listings when directory contents change via manual refresh mechanism (press 'r' in TUI or use refresh command)
- **FR-008**: System MUST handle script execution errors gracefully and display meaningful error messages
- **FR-009**: System MUST support basic script execution without argument passing (argument support to be added in future iterations)
- **FR-010**: System MUST execute scripts using the current user's permissions without additional authentication
- **FR-011**: System MUST support both interactive mode (TUI launched by default) and non-interactive mode (CLI commands like list, run, config)
- **FR-012**: System MUST render the TUI responsively based on terminal window size with fixed sidebar width to prevent layout shifts
- **FR-013**: System MUST provide a clean, minimal UI design suitable for daily developer use with appropriate icons for script types
- **FR-014**: System MUST provide an easy installation mechanism via binary distribution or source build
- **FR-015**: System MUST provide contextual search functionality that filters scripts within the current directory context
- **FR-016**: System MUST maintain a fixed sidebar width (35 characters) to prevent layout changes during navigation and accommodate longer script names
- **FR-017**: System MUST extract and display script metadata including descriptions, interpreters, and content previews using dedicated parser/lexer components
- **FR-018**: System MUST display breadcrumb navigation showing the current directory path hierarchy
- **FR-019**: System MUST preserve line breaks in multi-line script descriptions extracted from comments or docstrings

### Key Entities *(include if feature involves data)*
- **Script**: Represents an executable file (shell/Python) with metadata like name, path, permissions, and execution status
- **Directory Structure**: Represents the hierarchical organization of scripts, maintaining parent-child relationships between folders
- **Command Mapping**: Represents the association between a discovered script and its corresponding CLI command interface
- **Execution Session**: Represents a single script execution instance with input, output, error state, and duration
- **Configuration**: Represents user settings including watched directories, script type preferences, and TUI display options

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---

## Implementation Status
*Updated: 2025-01-07*

### Core Features (FR-001 to FR-015)
- ✅ **FR-001**: Script discovery - Fully implemented with recursive directory scanning
- ✅ **FR-002**: Multi-language support - Shell (.sh), Python (.py), Node (.js) supported
- ✅ **FR-003**: TUI interface - Complete with Bubble Tea framework
- ✅ **FR-004**: Hierarchical navigation - Directory tree with enter/back navigation
- ✅ **FR-005**: Script execution - TUI exits and passes control to script
- ✅ **FR-006**: Configurable directories - YAML configuration with script_dirs
- ✅ **FR-007**: Manual refresh - 'r' key in TUI refreshes script list
- ✅ **FR-008**: Error handling - Graceful error messages throughout
- ✅ **FR-009**: Basic execution - No argument passing (as designed for v1)
- ✅ **FR-010**: User permissions - Scripts execute with current user permissions
- ✅ **FR-011**: Interactive & CLI modes - TUI (default) + CLI commands (list, run, config)
- ✅ **FR-012**: Responsive rendering - Terminal size detection and layout adjustment
- ✅ **FR-013**: Clean UI design - Minimal design with type-specific icons
- ✅ **FR-014**: Easy installation - Go build system, binary distribution ready
- ✅ **FR-015**: Contextual search - '/' activates search within current directory

### Enhanced Features (FR-016 to FR-019)
- ✅ **FR-016**: Fixed sidebar width - 35 characters (updated from 24 in PR #5)
- ✅ **FR-017**: Metadata extraction - Complete parser/lexer system for shell and Python
  - Shell script parser extracts shebang, description from comments, markers
  - Python parser extracts docstrings, comment descriptions, interpreter
  - Metadata includes: Interpreter, Description, FullContent, PreviewLines, LineCount, IsTruncated
- ✅ **FR-018**: Breadcrumb navigation - Dedicated component showing path hierarchy
- ✅ **FR-019**: Line break preservation - Multi-line descriptions display with formatting (PR #5)

### Implementation Components
**Core Models** (`pkg/models/`):
- Script, Directory, ExecutionSession, Configuration, UIState

**Services** (`pkg/services/`):
- ScriptDiscovery, ScriptExecution, ConfigManager, SecurityValidator, ServiceRegistry

**Parser System** (`pkg/parser/`):
- Lexer interface, ShellLexer, PythonLexer, Metadata extraction

**TUI Components** (`pkg/tui/`):
- RootModel (compositor), SidebarModel, MainContentModel, HeaderModel, FooterModel, BreadcrumbModel, TUIManager

**CLI Commands** (`cmd/alec/`):
- list, run, config (show/edit/reset), refresh, version, demo

### Recent Enhancements
- **PR #5 (2025-01-07)**: Increased sidebar width to 35 chars, preserved description line breaks
- **PR #4**: Added dedicated breadcrumb row for navigation context
- **PR #3**: Implemented parser/lexer system with metadata extraction
- **PR #2**: Fixed UI overflow and layout issues
- **PR #1**: Initial TUI system implementation

### Known Limitations (By Design - v1)
- No argument passing to scripts (planned for v2)
- Manual refresh only (no auto-watch)
- TUI exits on script execution (by design for full terminal control)

---