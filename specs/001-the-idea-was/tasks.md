# Tasks: Script-to-CLI TUI System

**Input**: Design documents from `/specs/001-the-idea-was/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → If not found: ERROR "No implementation plan found"
   → Extract: tech stack, libraries, structure
2. Load optional design documents:
   → data-model.md: Extract entities → model tasks
   → contracts/: Each file → contract test task
   → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
   → Setup: project init, dependencies, linting
   → Tests: contract tests, integration tests
   → Core: models, services, CLI commands
   → Integration: DB, middleware, logging
   → Polish: unit tests, performance, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests?
   → All entities have models?
   → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `src/`, `tests/` at repository root
- Paths shown below assume single project based on plan.md structure

## Phase 3.1: Setup
- [ ] T001 Create Go module and project structure with src/, tests/, cmd/ directories
- [ ] T002 Initialize Go module with Bubble Tea, Lip Gloss, and Viper dependencies
- [ ] T003 [P] Configure Go linting tools (golangci-lint) and formatting (gofmt)
- [ ] T004 [P] Create Makefile with build, test, lint, and install targets

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests
- [ ] T005 [P] Contract test for ScriptDiscovery interface in tests/contract/script_discovery_test.go
- [ ] T006 [P] Contract test for ScriptExecutor interface in tests/contract/script_execution_test.go
- [ ] T007 [P] Contract test for TUIManager interface in tests/contract/tui_interface_test.go
- [ ] T008 [P] Contract test for ConfigManager interface in tests/contract/configuration_test.go

### Integration Tests (User Scenarios)
- [ ] T009 [P] Integration test: TUI navigation and script selection in tests/integration/tui_navigation_test.go
- [ ] T010 [P] Integration test: Script execution with output capture in tests/integration/script_execution_test.go
- [ ] T011 [P] Integration test: Directory scanning and refresh in tests/integration/directory_scan_test.go
- [ ] T012 [P] Integration test: Configuration loading and validation in tests/integration/config_loading_test.go
- [ ] T013 [P] Integration test: Non-interactive CLI mode script execution in tests/integration/cli_mode_test.go
- [ ] T014 [P] Integration test: Search and filtering functionality in tests/integration/search_filter_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Data Models
- [ ] T015 [P] Script entity with validation and state transitions in src/models/script.go
- [ ] T016 [P] Directory entity with hierarchical relationships in src/models/directory.go
- [ ] T017 [P] ExecutionSession entity with process management in src/models/execution_session.go
- [ ] T018 [P] Configuration entity with default values in src/models/configuration.go
- [ ] T019 [P] UIState entity with responsive layout calculations in src/models/ui_state.go

### Service Layer
- [ ] T020 [P] ScriptDiscovery service implementing directory scanning in src/services/script_discovery.go
- [ ] T021 [P] ScriptExecutor service implementing cross-platform execution in src/services/script_execution.go
- [ ] T022 [P] ConfigManager service implementing Viper-based configuration in src/services/configuration.go
- [ ] T023 SecurityValidator service for path validation and permission checks in src/services/security.go

### TUI Components
- [ ] T024 [P] Root Bubble Tea model with message routing in src/tui/root_model.go
- [ ] T025 [P] Sidebar model with tree navigation using treeview library in src/tui/sidebar_model.go
- [ ] T026 [P] Main content model with script details and output display in src/tui/main_model.go
- [ ] T027 [P] Header model with breadcrumbs and status in src/tui/header_model.go
- [ ] T028 [P] Footer model with help and key bindings in src/tui/footer_model.go
- [ ] T029 TUI manager implementing responsive layout and focus management in src/tui/manager.go

### CLI Commands
- [ ] T030 [P] Root CLI command with Cobra setup in cmd/alec/main.go
- [ ] T031 [P] Interactive mode command launching TUI in src/cli/interactive.go
- [ ] T032 [P] List command for non-interactive script listing in src/cli/list.go
- [ ] T033 [P] Run command for direct script execution in src/cli/run.go
- [ ] T034 [P] Config command for configuration management in src/cli/config.go
- [ ] T035 [P] Refresh command for manual directory rescan in src/cli/refresh.go

## Phase 3.4: Integration
- [ ] T036 Wire ScriptDiscovery service to TUI sidebar model for script loading
- [ ] T037 Wire ScriptExecutor service to TUI main model for execution display
- [ ] T038 Wire ConfigManager service to CLI commands for configuration access
- [ ] T039 Implement keyboard event handling and message routing in TUI
- [ ] T040 Add terminal size detection and responsive layout updates
- [ ] T041 Implement search functionality with real-time filtering
- [ ] T042 Add error handling with user-friendly error messages
- [ ] T043 Implement process cancellation and cleanup for script execution

## Phase 3.5: Polish
- [ ] T044 [P] Unit tests for Script model validation in tests/unit/script_test.go
- [ ] T045 [P] Unit tests for Directory tree operations in tests/unit/directory_test.go
- [ ] T046 [P] Unit tests for ExecutionSession state management in tests/unit/execution_session_test.go
- [ ] T047 [P] Unit tests for Configuration validation in tests/unit/configuration_test.go
- [ ] T048 [P] Unit tests for SecurityValidator path checks in tests/unit/security_test.go
- [ ] T049 Performance tests: script discovery under 100ms in tests/performance/discovery_test.go
- [ ] T050 Performance tests: UI updates under 16ms in tests/performance/ui_response_test.go
- [ ] T051 [P] Add comprehensive error handling and logging throughout application
- [ ] T052 [P] Create example scripts and configuration for quickstart guide
- [ ] T053 Run manual testing scenarios from quickstart.md guide
- [ ] T054 [P] Add installation script and build automation
- [ ] T055 Code review and refactoring for Go idioms and best practices

## Dependencies

### Setup Dependencies
- T001 blocks T002-T004
- T002 blocks all implementation tasks (T015+)

### Test Dependencies (Must complete before implementation)
- T005-T014 (all test tasks) must complete before T015-T043 (implementation tasks)
- Contract tests (T005-T008) can run in parallel
- Integration tests (T009-T014) can run in parallel

### Implementation Dependencies
- Models (T015-T019) before Services (T020-T023)
- Services (T020-T023) before TUI Components (T024-T029)
- Models and Services before CLI Commands (T030-T035)
- Core implementation (T015-T035) before Integration (T036-T043)
- Integration (T036-T043) before Polish (T044-T055)

### Specific Blocking Relationships
- T015-T019 (models) block T020-T023 (services)
- T020-T023 (services) block T024-T029 (TUI components)
- T024-T029 (TUI components) block T036-T041 (integration)
- T021 (ScriptExecutor) blocks T037 (execution display integration)
- T022 (ConfigManager) blocks T038 (configuration CLI integration)

## Parallel Example
```bash
# Launch contract tests together (T005-T008):
Task: "Contract test for ScriptDiscovery interface in tests/contract/script_discovery_test.go"
Task: "Contract test for ScriptExecutor interface in tests/contract/script_execution_test.go"
Task: "Contract test for TUIManager interface in tests/contract/tui_interface_test.go"
Task: "Contract test for ConfigManager interface in tests/contract/configuration_test.go"

# Launch integration tests together (T009-T014):
Task: "Integration test: TUI navigation and script selection in tests/integration/tui_navigation_test.go"
Task: "Integration test: Script execution with output capture in tests/integration/script_execution_test.go"
Task: "Integration test: Directory scanning and refresh in tests/integration/directory_scan_test.go"
Task: "Integration test: Configuration loading and validation in tests/integration/config_loading_test.go"

# Launch model creation together (T015-T019):
Task: "Script entity with validation and state transitions in src/models/script.go"
Task: "Directory entity with hierarchical relationships in src/models/directory.go"
Task: "ExecutionSession entity with process management in src/models/execution_session.go"
```

## Notes
- [P] tasks = different files, no dependencies
- Verify tests fail before implementing
- Follow TDD: Red-Green-Refactor cycle
- Commit after each completed task
- Use Go 1.21+ features (filepath.IsLocal for security)
- Implement responsive design with golden ratio layouts
- Ensure cross-platform compatibility (Linux, macOS, Windows)

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
   - 4 contract files → 4 contract test tasks [P] (T005-T008)
   - Each interface → corresponding service implementation

2. **From Data Model**:
   - 5 entities → 5 model creation tasks [P] (T015-T019)
   - Relationships → service layer integration tasks

3. **From User Stories** (quickstart.md):
   - Navigation scenario → TUI navigation test (T009)
   - Execution scenario → Script execution test (T010)
   - CLI mode → Non-interactive test (T013)
   - Search scenario → Search filtering test (T014)

4. **Ordering**:
   - Setup (T001-T004) → Tests (T005-T014) → Models (T015-T019) → Services (T020-T023) → TUI (T024-T029) → CLI (T030-T035) → Integration (T036-T043) → Polish (T044-T055)

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All contracts have corresponding tests (T005-T008)
- [x] All entities have model tasks (T015-T019)
- [x] All tests come before implementation (T005-T014 before T015+)
- [x] Parallel tasks truly independent ([P] tasks use different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] User scenarios from quickstart.md covered in integration tests
- [x] TDD approach enforced with failing tests requirement
- [x] Go project structure and conventions followed