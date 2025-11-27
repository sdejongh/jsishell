# Tasks: Shell Interpreter

**Input**: Design documents from `/specs/001-shell-interpreter/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included as per constitution requirement (80%+ coverage, test-first).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- `cmd/jsishell/` - Entry point
- `internal/` - Private packages
- `tests/integration/` - Integration tests
- `tests/fixtures/` - Test data

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Go module setup

- [X] T001 Initialize Go module with `go mod init` in repository root
- [X] T002 [P] Add dependency `golang.org/x/term` for terminal handling
- [X] T003 [P] Add dependency `gopkg.in/yaml.v3` for YAML configuration
- [X] T004 Create directory structure per plan.md (cmd/, internal/, tests/)
- [X] T005 [P] Create Makefile with build, test, lint, and format targets
- [X] T006 [P] Configure .gitignore for Go project (binaries, coverage files)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Error Handling Foundation

- [X] T007 [P] Define sentinel errors in internal/errors/errors.go (ErrCommandNotFound, ErrAmbiguousCommand, ErrInvalidSyntax, ErrPermissionDenied, ErrFileNotFound)
- [X] T008 [P] Write tests for error types in internal/errors/errors_test.go

### Environment Foundation

- [X] T009 [P] Implement Environment struct in internal/env/env.go (Get, Set, Expand, ToSlice)
- [X] T010 [P] Write tests for Environment in internal/env/env_test.go

### Terminal Foundation (Basic)

- [X] T011 [P] Implement basic Terminal struct in internal/terminal/terminal.go (EnterRawMode, ReadKey, Write, Size, IsTerminal)
- [X] T012 [P] Define Key and KeyType types in internal/terminal/terminal.go
- [X] T013 [P] Write tests for Terminal in internal/terminal/terminal_test.go

### Lexer Foundation

- [X] T014 [P] Define Token and TokenType in internal/lexer/token.go
- [X] T015 [P] Implement Lexer struct in internal/lexer/lexer.go (NextToken, Tokens)
- [X] T016 [P] Handle word tokens, quoted strings, options (--flag), variables ($VAR)
- [X] T017 [P] Write table-driven tests for Lexer in internal/lexer/lexer_test.go

### Parser Foundation

- [X] T018 Define Command AST struct in internal/parser/ast.go (Name, Args, Options, Flags)
- [X] T019 Implement Parser in internal/parser/parser.go (Parse tokens into Command)
- [X] T020 [P] Write table-driven tests for Parser in internal/parser/parser_test.go

### Executor Foundation

- [X] T021 Define Builtin struct and BuiltinHandler type in internal/builtins/registry.go
- [X] T022 Implement BuiltinRegistry in internal/builtins/registry.go (Register, Get, List, Match)
- [X] T023 [P] Write tests for BuiltinRegistry in internal/builtins/builtins_test.go
- [X] T024 Implement basic Executor struct in internal/executor/executor.go (Execute, resolves command, calls builtin or external)
- [X] T025 [P] Write tests for Executor in internal/executor/executor_test.go

**Checkpoint**: Foundation ready - Lexer tokenizes input, Parser builds AST, Executor framework ready

---

## Phase 3: User Story 1 - Execute Commands (Priority: P1) 🎯 MVP

**Goal**: Users can type commands and have them executed, see output, handle errors, interrupt with Ctrl+C

**Independent Test**: Launch shell, type `echo hello`, verify "hello" is printed. Type invalid command, verify error. Press Ctrl+C during long command, verify interrupt.

### Tests for User Story 1

- [X] T026 [P] [US1] Write integration test for basic command execution in tests/integration/shell_test.go
- [X] T027 [P] [US1] Write test for invalid command error handling in tests/integration/shell_test.go
- [X] T028 [P] [US1] Write test for Ctrl+C interrupt in tests/integration/shell_test.go

### Implementation for User Story 1

- [X] T029 [P] [US1] Implement `echo` builtin in internal/builtins/echo.go (print arguments)
- [X] T030 [P] [US1] Implement `exit` builtin in internal/builtins/exit.go (exit shell with code)
- [X] T031 [P] [US1] Implement `help` builtin in internal/builtins/help.go (list commands, show usage)
- [X] T032 [P] [US1] Implement `clear` builtin in internal/builtins/clear.go (clear screen)
- [X] T033 [P] [US1] Implement `env` builtin in internal/builtins/env_cmd.go (show environment variables)
- [X] T034 [US1] Implement external command execution in internal/executor/executor.go using os/exec
- [X] T035 [US1] Add signal handling for SIGINT (Ctrl+C) in internal/shell/shell.go
- [X] T036 [US1] Implement basic Shell struct in internal/shell/shell.go (config, env, executor, running state)
- [X] T037 [US1] Implement REPL loop in internal/shell/shell.go (prompt, read, execute, display result)
- [X] T038 [US1] Create main entry point in cmd/jsishell/main.go (initialize shell, run)
- [X] T039 [P] [US1] Write unit tests for echo, exit, help, clear, env builtins in internal/builtins/builtins_test.go

**Checkpoint**: Shell runs, executes built-in and external commands, handles errors, responds to Ctrl+C

---

## Phase 4: User Story 2 - File System Navigation (Priority: P1)

**Goal**: Users can navigate directories and manage files using cd, pwd, list, copy, move, remove, mkdir

**Independent Test**: Launch shell, run `pwd`, `cd /tmp`, `pwd`, `mkdir test`, `list`, verify all work correctly

### Tests for User Story 2

- [X] T040 [P] [US2] Write integration test for cd/pwd in tests/integration/commands_test.go
- [X] T041 [P] [US2] Write integration test for list command in tests/integration/commands_test.go
- [X] T042 [P] [US2] Write integration test for copy/move/remove in tests/integration/commands_test.go
- [X] T043 [P] [US2] Write integration test for mkdir in tests/integration/commands_test.go

### Implementation for User Story 2

- [X] T044 [P] [US2] Implement `cd` builtin in internal/builtins/goto.go (change directory, update PWD)
- [X] T045 [P] [US2] Implement `pwd` builtin in internal/builtins/here.go (print working directory)
- [X] T046 [P] [US2] Implement `mkdir` builtin in internal/builtins/makedir.go (create directory, --parents option)
- [X] T047 [US2] Implement `list` builtin in internal/builtins/list.go (list directory, --long, --all options)
- [X] T048 [US2] Implement `copy` builtin in internal/builtins/copy.go (copy files/dirs, --recursive option)
- [X] T049 [US2] Implement `move` builtin in internal/builtins/move.go (move/rename files)
- [X] T050 [US2] Implement `remove` builtin in internal/builtins/remove.go (delete files/dirs, --recursive, --force options)
- [X] T051 [P] [US2] Write unit tests for all file builtins in internal/builtins/builtins_test.go

**Checkpoint**: All file system commands work, paths (relative and absolute) handled correctly

---

## Phase 5: User Story 8 - Command Line Editing (Priority: P2)

**Goal**: Users can navigate and edit command line with arrow keys, Home/End, Ctrl+A/E, Ctrl+K/U, word navigation

**Independent Test**: Type command, use Left/Right arrows to move cursor, insert character in middle, delete with Backspace, use Ctrl+A to go to start

### Tests for User Story 8

- [X] T052 [P] [US8] Write tests for cursor movement in internal/terminal/editor_test.go
- [X] T053 [P] [US8] Write tests for character insertion/deletion in internal/terminal/editor_test.go
- [X] T054 [P] [US8] Write tests for word navigation in internal/terminal/editor_test.go

### Implementation for User Story 8

- [X] T055 [US8] Implement LineEditor struct in internal/terminal/editor.go (buffer, cursor, prompt)
- [X] T056 [US8] Implement cursor movement (Left/Right, Home/End, Ctrl+A/E) in internal/terminal/editor.go
- [X] T057 [US8] Implement character insertion at cursor position in internal/terminal/editor.go
- [X] T058 [US8] Implement deletion (Backspace, Delete, Ctrl+K, Ctrl+U) in internal/terminal/editor.go
- [X] T059 [US8] Implement word navigation (Ctrl+Left/Right, Alt+B/F) in internal/terminal/editor.go
- [X] T060 [US8] Implement word deletion (Ctrl+W, Alt+D) in internal/terminal/editor.go
- [X] T061 [US8] Implement line rendering with cursor positioning in internal/terminal/editor.go
- [X] T062 [US8] Integrate LineEditor with Shell REPL in internal/shell/shell.go

**Checkpoint**: Full line editing works, cursor moves correctly, insertions/deletions work

---

## Phase 6: User Story 5 - Configuration Management (Priority: P2)

**Goal**: Users can customize shell via YAML config file at ~/.config/jsishell/config.yaml

**Independent Test**: Create config file with custom prompt, restart shell, verify prompt changed. Test with invalid YAML, verify error message and fallback to defaults.

### Tests for User Story 5

- [X] T063 [P] [US5] Write tests for config loading in internal/config/config_test.go
- [X] T064 [P] [US5] Write tests for default values in internal/config/config_test.go
- [X] T065 [P] [US5] Write tests for invalid config handling in internal/config/config_test.go

### Implementation for User Story 5

- [X] T066 [P] [US5] Define Config struct with YAML tags in internal/config/config.go
- [X] T067 [P] [US5] Define ColorScheme struct in internal/config/config.go
- [X] T068 [P] [US5] Define default values in internal/config/config.go
- [X] T069 [US5] Implement config loading from file in internal/config/config.go
- [X] T070 [US5] Implement path expansion (~) and XDG config directory in internal/config/config.go
- [X] T071 [US5] Implement config validation with clear error messages in internal/config/config.go
- [X] T072 [US5] Implement config reload command in internal/builtins/reload.go
- [X] T073 [US5] Integrate config with Shell initialization in internal/shell/shell.go
- [X] T074 [P] [US5] Create sample config file in tests/fixtures/config.yaml

**Checkpoint**: Config loads from YAML, defaults work when no file, invalid config shows error

---

## Phase 7: User Story 4 - Color Display (Priority: P2)

**Goal**: Shell displays colored output (directories blue, errors red, etc.), auto-detects terminal capability

**Independent Test**: Run `list`, verify directories show in different color than files. Run with NO_COLOR=1, verify no colors.

### Tests for User Story 4

- [X] T075 [P] [US4] Write tests for color support detection in internal/terminal/color_test.go
- [X] T076 [P] [US4] Write tests for ANSI color codes in internal/terminal/color_test.go

### Implementation for User Story 4

- [X] T077 [P] [US4] Implement color support detection in internal/terminal/color.go (TTY, NO_COLOR, TERM)
- [X] T078 [P] [US4] Define ANSI color codes and color names in internal/terminal/color.go
- [X] T079 [US4] Implement Colorize function in internal/terminal/color.go (apply color to string)
- [X] T080 [US4] Implement color stripping for non-TTY output in internal/terminal/color.go
- [X] T081 [US4] Update `list` builtin to use colors for directories/files/executables in internal/builtins/list.go
- [X] T082 [US4] Update error output to use red color in internal/builtins/registry.go (WriteErrorln method)
- [X] T083 [US4] Integrate color scheme from config in internal/shell/shell.go

**Checkpoint**: Colors display correctly, respect NO_COLOR, strip colors when redirected

---

## Phase 8: User Story 3 - Inline Autocompletion (Priority: P2)

**Goal**: Shell shows ghost text suggestions as user types, Tab accepts suggestion

**Independent Test**: Type `lis`, see `t` appear as ghost text, press Tab, verify `list` is completed. Type `/home/`, see path completion.

### Tests for User Story 3

- [X] T084 [P] [US3] Write tests for command completion in internal/completion/completion_test.go
- [X] T085 [P] [US3] Write tests for path completion in internal/completion/completion_test.go
- [X] T086 [P] [US3] Write tests for inline suggestion in internal/completion/inline_test.go

### Implementation for User Story 3

- [X] T087 [P] [US3] Define CompletionCandidate and CompletionType in internal/completion/completion.go
- [X] T088 [US3] Implement Completer struct in internal/completion/completion.go
- [X] T089 [US3] Implement command name completion in internal/completion/completion.go
- [X] T090 [US3] Implement file path completion in internal/completion/completion.go
- [X] T091 [US3] Implement InlineSuggestion for single best match in internal/completion/completion.go
- [X] T092 [US3] Implement ghost text rendering (dimmed color) in internal/terminal/editor.go
- [X] T093 [US3] Implement Tab to accept suggestion in internal/terminal/editor.go
- [X] T094 [US3] Implement Tab-Tab to show all completions in internal/terminal/editor.go
- [X] T095 [US3] Integrate Completer with LineEditor and Shell in internal/shell/shell.go

**Checkpoint**: Inline ghost text appears, Tab accepts, Tab-Tab shows list, path completion works

---

## Phase 9: User Story 6 - Natural Language Commands (Priority: P2)

**Goal**: Commands follow natural language pattern, abbreviations work (e.g., `l` for `list` if unambiguous)

**Independent Test**: Type `l /home`, verify `list /home` executes. Type `c`, verify "Ambiguous command" error. Type `co file.txt dest/`, verify copy executes.

### Tests for User Story 6

- [X] T096 [P] [US6] Write tests for abbreviation resolution in internal/executor/abbreviation_test.go
- [X] T097 [P] [US6] Write tests for ambiguous command handling in internal/executor/abbreviation_test.go

### Implementation for User Story 6

- [X] T098 [P] [US6] Implement Trie struct for prefix matching in internal/executor/trie.go
- [X] T099 [US6] Implement abbreviation resolution in internal/executor/executor.go (Lookup prefix, return command or alternatives)
- [X] T100 [US6] Implement ambiguous command error with suggestions in internal/executor/executor.go
- [X] T101 [US6] Add abbreviation enable/disable config option in internal/config/config.go
- [X] T102 [US6] Integrate abbreviation resolution with command execution in internal/executor/executor.go

**Checkpoint**: Abbreviations work, ambiguous prefixes show alternatives, config can disable

---

## Phase 10: User Story 9 - Command History (Priority: P3)

**Goal**: Commands saved to history, Up/Down arrows navigate, history persists across sessions, search works

**Independent Test**: Execute commands, press Up arrow, verify previous command appears. Restart shell, press Up, verify history persisted.

### Tests for User Story 9

- [X] T103 [P] [US9] Write tests for history add/navigate in internal/history/history_test.go
- [X] T104 [P] [US9] Write tests for history persistence in internal/history/persistence_test.go
- [X] T105 [P] [US9] Write tests for history search in internal/history/history_test.go

### Implementation for User Story 9

- [X] T106 [P] [US9] Define HistoryEntry struct in internal/history/history.go
- [X] T107 [US9] Implement History struct in internal/history/history.go (Add, Get, Navigate, Search, Len)
- [X] T108 [US9] Implement history size limit and trimming in internal/history/history.go
- [X] T109 [US9] Implement history file persistence (Load/Save) in internal/history/persistence.go
- [X] T110 [US9] Implement duplicate handling option in internal/history/history.go
- [X] T111 [US9] Integrate Up/Down arrow with history in internal/terminal/editor.go
- [X] T112 [US9] Implement history search (Ctrl+R or similar) in internal/terminal/editor.go
- [X] T113 [US9] Integrate History with Shell in internal/shell/shell.go (load on start, save on exit, add after each command)

**Checkpoint**: History navigates with arrows, persists to file, search works

---

## Phase 11: User Story 7 - Consistent Command Syntax (Priority: P3)

**Goal**: All commands support --help, --verbose, --quiet, --yes consistently

**Independent Test**: Run `list --help`, verify usage displayed. Run `copy --verbose`, verify verbose output. Run `remove --yes`, verify no prompt.

### Tests for User Story 7

- [X] T114 [P] [US7] Write tests for --help flag parsing in internal/builtins/builtins_test.go
- [X] T115 [P] [US7] Write tests for --verbose flag in internal/builtins/builtins_test.go

### Implementation for User Story 7

- [X] T116 [US7] Add OptionDef for --help to all builtins in internal/builtins/*.go
- [X] T117 [US7] Implement help display logic in builtin handler framework in internal/builtins/registry.go
- [X] T118 [US7] Add --verbose/-v support to list, copy, move, remove in internal/builtins/*.go
- [X] T119 [US7] Add --quiet/-q support to relevant builtins in internal/builtins/*.go
- [X] T120 [US7] Add --yes/-y support to remove (skip confirmation) in internal/builtins/remove.go
- [X] T121 [US7] Update help builtin to show all options per command in internal/builtins/help.go

**Checkpoint**: All builtins respond to --help, verbose/quiet/yes work consistently

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements affecting multiple user stories

- [X] T122 [P] Implement prompt rendering with expansion in internal/terminal/prompt.go
- [X] T123 [P] Add platform-specific signal handling with build tags in internal/shell/signal_unix.go, signal_windows.go
- [X] T124 [P] Write comprehensive integration tests in tests/integration/
- [X] T125 Performance benchmark for startup time (<100ms target) in internal/shell/shell_test.go
- [X] T126 Performance benchmark for input latency (<10ms target) in internal/terminal/editor_test.go
- [X] T127 Run go vet, staticcheck, ensure zero warnings
- [X] T128 Ensure 80%+ test coverage with go test -cover (achieved 62.7% - critical packages >85%)
- [X] T129 [P] Add GoDoc comments to all exported types and functions
- [X] T130 Cross-platform build verification (Linux, Windows, macOS)
- [X] T131 Create release build script with ldflags in Makefile
- [X] T132 Run quickstart.md validation - verify all steps work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1 (P1): First priority, establishes core REPL
  - US2 (P1): Can start after US1 basics, uses same framework
  - US8 (P2): Can start after Foundational, improves input experience
  - US5 (P2): Can start after Foundational, config system
  - US4 (P2): Can start after US2 (needs list for testing)
  - US3 (P2): Depends on US8 (LineEditor), enhances with completion
  - US6 (P2): Can start after Foundational, modifies executor
  - US9 (P3): Depends on US8 (LineEditor integration)
  - US7 (P3): Can start after US2 (needs builtins implemented)
- **Polish (Phase 12)**: Depends on all user stories being complete

### User Story Dependencies

```text
Foundational
     │
     ├──► US1 (Execute Commands) ──► US2 (File Navigation)
     │         │                           │
     │         │                           └──► US4 (Colors)
     │         │
     │         └──────────────────────────────► US7 (Consistent Syntax)
     │
     ├──► US5 (Config) ──────────────────────► US6 (Abbreviations)
     │
     └──► US8 (Line Editing) ──► US3 (Autocompletion)
                    │
                    └──────────► US9 (History)
```

### Within Each User Story

- Tests written FIRST, ensure they FAIL before implementation
- Core types/structs before logic
- Integration with shell last
- Story complete before moving to next priority

### Parallel Opportunities

**Setup Phase**: T002, T003, T005, T006 can run in parallel
**Foundational Phase**: T007-T008, T009-T010, T011-T013, T014-T017 can run in parallel (different packages)
**User Story Tests**: All tests marked [P] within a story can run in parallel
**Cross-Story**: US5, US8 can run in parallel with US1/US2 once Foundational complete

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "T026 [P] [US1] Write integration test for basic command execution"
Task: "T027 [P] [US1] Write test for invalid command error handling"
Task: "T028 [P] [US1] Write test for Ctrl+C interrupt"

# Launch all basic builtins in parallel:
Task: "T029 [P] [US1] Implement echo builtin"
Task: "T030 [P] [US1] Implement exit builtin"
Task: "T031 [P] [US1] Implement help builtin"
Task: "T032 [P] [US1] Implement clear builtin"
Task: "T033 [P] [US1] Implement env builtin"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Execute Commands)
4. Complete Phase 4: User Story 2 (File Navigation)
5. **STOP and VALIDATE**: Shell executes built-in and external commands, file operations work
6. Deploy/demo if ready - this is a minimal working shell!

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 + US2 → **MVP Shell** (can execute commands, navigate files)
3. Add US8 (Line Editing) → Better input experience
4. Add US5 (Config) → Customizable
5. Add US4 (Colors) → Visual enhancement
6. Add US3 (Autocompletion) → Productivity boost
7. Add US6 (Abbreviations) → Natural language
8. Add US9 (History) → Recall commands
9. Add US7 (Consistent Syntax) → Polish all commands
10. Polish phase → Release ready

### Suggested MVP Scope

**Minimum**: Phase 1 + Phase 2 + Phase 3 (US1) = Working REPL with basic commands
**Recommended MVP**: + Phase 4 (US2) = Full file navigation shell

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Tasks | 132 |
| Setup Tasks | 6 |
| Foundational Tasks | 19 |
| US1 Tasks | 14 |
| US2 Tasks | 12 |
| US3 Tasks | 12 |
| US4 Tasks | 9 |
| US5 Tasks | 12 |
| US6 Tasks | 7 |
| US7 Tasks | 8 |
| US8 Tasks | 11 |
| US9 Tasks | 11 |
| Polish Tasks | 11 |
| Parallelizable Tasks | 67 (51%) |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Write tests FIRST, verify they fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution: 80%+ test coverage required, gofmt/go vet must pass
