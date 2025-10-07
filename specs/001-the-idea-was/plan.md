# Implementation Plan: Script-to-CLI TUI System

**Branch**: `001-the-idea-was` | **Date**: 2025-09-18 | **Last Updated**: 2025-01-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-the-idea-was/spec.md`
**Status**: Implementation Complete

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Script-to-CLI TUI System that automatically discovers scripts in configured directories and presents them through a clean Terminal User Interface for execution, using the Charm Bubble Tea framework for responsive TUI rendering and directory-based script organization.

## Technical Context
**Language/Version**: Go 1.21+ (for Bubble Tea framework)
**Primary Dependencies**:
- Bubble Tea framework (github.com/charmbracelet/bubbletea)
- Lip Gloss for styling (github.com/charmbracelet/lipgloss)
- Cobra for CLI framework (github.com/spf13/cobra)
- Viper for configuration management
**Parser/Lexer System**: Custom-built shell and Python script parsers for metadata extraction
**Storage**: File system-based script discovery and configuration files
**Testing**: Go standard testing package with table-driven tests
**Target Platform**: Cross-platform terminal applications (Linux, macOS, Windows)
**Project Type**: single - CLI/TUI application
**Performance Goals**: Sub-100ms script discovery, responsive UI updates <16ms
**Constraints**: Clean minimal UI, manual refresh mechanism, no argument passing in v1
**Scale/Scope**: Support for hundreds of scripts across nested directory structures

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on the constitution template, evaluating against common constitutional principles:

**Library-First**: ✅ PASS - Core script discovery and TUI components will be modular libraries
**CLI Interface**: ✅ PASS - Primary interface is CLI/TUI with both interactive and non-interactive modes
**Test-First**: ✅ PASS - Will implement TDD for script discovery, UI components, and execution logic
**Integration Testing**: ✅ PASS - Focus on script discovery, TUI navigation, and execution workflows
**Simplicity**: ✅ PASS - Clean, minimal design without unnecessary complexity

No constitutional violations identified.

## Project Structure

### Documentation (this feature)
```
specs/001-the-idea-was/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/
```

**Structure Decision**: Option 1 - Single project CLI/TUI application

## Phase 0: Outline & Research

Research tasks needed:
1. Bubble Tea framework best practices for TUI applications
2. File system watching/discovery patterns in Go
3. Script execution and output capture techniques
4. Cross-platform terminal compatibility considerations
5. Configuration management for CLI tools

**Output**: research.md with all technical decisions documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Script entity with path, type, permissions, metadata
   - Directory structure representation
   - Configuration entity for user settings
   - Execution session for runtime state

2. **Generate API contracts** from functional requirements:
   - Script discovery interface
   - TUI state management contracts
   - Script execution interface
   - Configuration management interface

3. **Generate contract tests** from contracts:
   - Script discovery test scenarios
   - TUI navigation test patterns
   - Execution output capture tests

4. **Extract test scenarios** from user stories:
   - Directory scanning and script detection
   - TUI navigation and selection
   - Script execution and output display
   - Configuration management flows

5. **Update agent file incrementally**:
   - Add Bubble Tea framework context
   - Include Go TUI development patterns
   - Document script discovery approaches

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, CLAUDE.md

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P]
- Each user story → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:
- TDD order: Tests before implementation
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

No constitutional violations identified.

## Progress Tracking
*This checklist is updated during execution flow*
*Last Updated: 2025-01-07*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [x] Phase 3: Tasks generated (/tasks command)
- [x] Phase 4: Implementation complete
- [x] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

**Implementation Milestones**:
- [x] PR #1: Initial TUI system with Bubble Tea
- [x] PR #2: UI overflow and layout fixes
- [x] PR #3: Parser/lexer system with metadata extraction
- [x] PR #4: Breadcrumb navigation component
- [x] PR #5: Sidebar width increase (24→35) and description line break preservation
- [x] Core features (FR-001 to FR-015): Complete
- [x] Enhanced features (FR-016 to FR-019): Complete

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*