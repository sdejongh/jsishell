# Feature Specification: Shell Interpreter

**Feature Branch**: `001-shell-interpreter`
**Created**: 2025-01-25
**Status**: Active Development (on-demand)
**Input**: User description: "Command line interpreter and programming language for OS interaction with file system management, program execution, consistent syntax, color display, autocompletion, and YAML configuration"

## Development Approach

This project follows an **on-demand development** approach. There is no fixed roadmap. Features and improvements are implemented as requested by the user.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Execute Commands (Priority: P1)

As a user, I want to type commands and have them executed by the shell so that I can interact with my operating system and perform tasks like listing files, navigating directories, and running programs.

**Why this priority**: Command execution is the core functionality of any shell. Without it, no other features are useful. This is the fundamental value proposition.

**Independent Test**: Can be fully tested by launching the shell, typing basic commands (list files, change directory, run a program), and verifying the output matches expected results.

**Acceptance Scenarios**:

1. **Given** the shell is running, **When** user types `list` and presses Enter, **Then** the shell displays the contents of the current directory
2. **Given** the shell is running, **When** user types an external program name with arguments, **Then** the shell executes the program and displays its output
3. **Given** the shell is running, **When** user types an invalid command, **Then** the shell displays a clear error message indicating the command was not found
4. **Given** a command is running, **When** user presses Ctrl+C, **Then** the running command is interrupted and the shell returns to the prompt

---

### User Story 2 - File System Navigation (Priority: P1)

As a user, I want to navigate and manage the file system using intuitive commands so that I can organize my files and directories efficiently.

**Why this priority**: File system operations are essential daily tasks for any shell user. This is a core feature that provides immediate value.

**Independent Test**: Can be tested by navigating directories, creating/deleting files and folders, copying/moving items, and verifying changes in the file system.

**Acceptance Scenarios**:

1. **Given** the shell is running, **When** user types `cd /path/to/directory`, **Then** the current working directory changes to the specified path
2. **Given** the shell is running, **When** user types `mkdir newfolder`, **Then** a new directory named "newfolder" is created in the current location
3. **Given** a file exists, **When** user types `copy source.txt destination.txt`, **Then** the file is copied to the new location
4. **Given** a file exists, **When** user types `remove file.txt`, **Then** the file is deleted from the file system
5. **Given** the shell is running, **When** user types `pwd`, **Then** the shell displays the current working directory path

---

### User Story 3 - Inline Autocompletion (Priority: P2)

As a user, I want the shell to display inline suggestions as I type and allow me to accept them with Tab so that I can type commands faster with fewer keystrokes.

**Why this priority**: Inline autocompletion significantly improves productivity and is a modern shell feature that differentiates this shell from traditional ones.

**Independent Test**: Can be tested by typing partial commands or paths and verifying inline suggestions appear and can be accepted.

**Acceptance Scenarios**:

1. **Given** the shell is running, **When** user starts typing `lis`, **Then** the shell displays `t` as ghost text (suggesting `list`)
2. **Given** an inline suggestion is displayed, **When** user presses Tab, **Then** the suggestion is accepted and inserted into the command line
3. **Given** a partial file path `/home/us` is typed, **When** inline completion is triggered, **Then** the shell suggests the rest of the path (e.g., `er` for `/home/user`)
4. **Given** multiple completions are possible, **When** user presses Tab twice, **Then** all possible completions are displayed as a list
5. **Given** the user continues typing, **When** new characters are entered, **Then** the inline suggestion updates dynamically
6. **Given** an external command exists in PATH (e.g., `vim`), **When** user types `vi`, **Then** the shell suggests `m` as ghost text (completing to `vim`)

---

### User Story 4 - Color Display (Priority: P2)

As a user, I want the shell to display output with colors so that I can quickly distinguish between different types of information (directories, files, errors, etc.).

**Why this priority**: Colors improve readability and user experience but are not essential for functionality.

**Independent Test**: Can be tested by running commands and verifying that output uses appropriate colors for different element types.

**Acceptance Scenarios**:

1. **Given** a directory listing is displayed, **When** the output is rendered, **Then** directories are shown in a distinct color from files
2. **Given** an error occurs, **When** the error message is displayed, **Then** it appears in a warning/error color (e.g., red)
3. **Given** the user has configured custom colors, **When** output is displayed, **Then** the configured colors are used
4. **Given** output is redirected to a file, **When** the command executes, **Then** color codes are stripped from the output

---

### User Story 5 - Configuration Management (Priority: P2)

As a user, I want to customize the shell behavior through a YAML configuration file so that I can personalize my experience and set persistent preferences.

**Why this priority**: Configuration allows personalization but the shell works with sensible defaults.

**Independent Test**: Can be tested by modifying the configuration file and verifying that changes take effect after restart or reload.

**Acceptance Scenarios**:

1. **Given** a configuration file exists at `~/.config/jsishell/config.yaml`, **When** the shell starts, **Then** it loads and applies the settings from the file
2. **Given** no configuration file exists, **When** the shell starts, **Then** it uses sensible default values
3. **Given** the configuration file contains invalid syntax, **When** the shell starts, **Then** it displays a clear error message and falls back to defaults
4. **Given** the shell is running, **When** user executes a reload command, **Then** the configuration is reloaded without restarting the shell

---

### User Story 6 - Natural Language Commands (Priority: P2)

As a user, I want commands to follow natural language patterns and support abbreviations so that I can type commands intuitively and quickly without memorizing cryptic syntax.

**Why this priority**: Natural language syntax is a core differentiator of this shell and directly impacts usability. Abbreviations significantly speed up experienced user workflows.

**Independent Test**: Can be tested by typing natural language commands and abbreviated forms, verifying they execute correctly.

**Acceptance Scenarios**:

1. **Given** the shell is running, **When** user types `move file.txt /backup/`, **Then** the file is moved to the backup directory (natural verb-subject-destination order)
2. **Given** the shell is running, **When** user types `l /home`, **Then** the shell executes `list /home` if `list` is the only command starting with `l`
3. **Given** `copy` and `clear` both exist, **When** user types `c`, **Then** the shell displays an error: "Ambiguous command 'c'. Did you mean: clear, copy?"
4. **Given** `copy` and `clear` both exist, **When** user types `co`, **Then** the shell executes `copy` (unambiguous prefix)
5. **Given** abbreviation is disabled in config, **When** user types `l`, **Then** the shell displays "Command not found: l"

---

### User Story 7 - Consistent Command Syntax (Priority: P3)

As a user, I want all commands to follow the same syntax patterns so that I can learn the shell quickly and predict how commands work.

**Why this priority**: Consistency improves learnability but requires foundational commands to be implemented first.

**Independent Test**: Can be tested by verifying that the same argument (e.g., `--help`, `--verbose`) works consistently across all commands.

**Acceptance Scenarios**:

1. **Given** any built-in command, **When** user adds `--help` argument, **Then** the command displays its usage information
2. **Given** any command that supports verbose output, **When** user adds `--verbose` argument, **Then** the command provides detailed output
3. **Given** any command that produces output, **When** user adds `--quiet` argument, **Then** only essential output is displayed
4. **Given** a command requires confirmation, **When** user adds `--yes` argument, **Then** the command proceeds without prompting

---

### User Story 8 - Command Line Editing (Priority: P2)

As a user, I want to navigate and edit the current command line using keyboard shortcuts so that I can correct mistakes without retyping the entire command.

**Why this priority**: Line editing is essential for usability and user productivity when working with long commands.

**Independent Test**: Can be tested by typing a command, then using arrow keys and keyboard shortcuts to modify it.

**Acceptance Scenarios**:

1. **Given** a command is partially typed, **When** user presses Left/Right arrow, **Then** cursor moves one character in that direction
2. **Given** cursor is in the middle of input, **When** user presses Home or Ctrl+A, **Then** cursor moves to the beginning of the line
3. **Given** cursor is in the middle of input, **When** user presses End or Ctrl+E, **Then** cursor moves to the end of the line
4. **Given** cursor is in the middle of input, **When** user presses Backspace, **Then** character before cursor is deleted
5. **Given** cursor is in the middle of input, **When** user presses Ctrl+Left/Right, **Then** cursor moves to previous/next word boundary
6. **Given** cursor is in the middle of input, **When** user presses Ctrl+K, **Then** text from cursor to end of line is deleted
7. **Given** cursor is in the middle of input, **When** user types a character, **Then** it is inserted at cursor position

---

### User Story 9 - Command History (Priority: P3)

As a user, I want to access and search my command history so that I can quickly repeat or modify previous commands.

**Why this priority**: History improves efficiency but is not essential for basic shell operation.

**Independent Test**: Can be tested by executing commands, using arrow keys to navigate history, and using search functionality.

**Acceptance Scenarios**:

1. **Given** commands have been executed, **When** user presses Up arrow, **Then** the previous command is displayed in the input line
2. **Given** commands have been executed, **When** user types a search pattern and activates history search, **Then** matching commands are displayed
3. **Given** the shell restarts, **When** user accesses history, **Then** commands from previous sessions are available
4. **Given** the history limit is configured, **When** new commands exceed the limit, **Then** oldest commands are removed

---

### Edge Cases

- What happens when a command produces very large output (thousands of lines)?
  - Output is streamed progressively; user can interrupt with Ctrl+C
- What happens when the user's home directory does not exist?
  - Shell uses current working directory as fallback and displays a warning
- What happens when the configuration file is locked by another process?
  - Shell displays a warning and uses default configuration
- What happens when Tab completion has thousands of matches?
  - Shell prompts user before displaying more than a configurable threshold (default: 100)
- What happens when a command path contains spaces or special characters?
  - Proper quoting/escaping is supported and documented
- What happens on terminals that don't support colors?
  - Shell detects terminal capabilities and disables colors automatically
- What happens when an abbreviated command matches an external program name exactly?
  - Abbreviation matching applies only to built-in commands; external programs require exact names
- What happens when a new built-in command is added that creates new abbreviation conflicts?
  - Existing single-character abbreviations may become ambiguous; users must type more characters

## Requirements *(mandatory)*

### Functional Requirements

#### Core Shell Operations

- **FR-001**: System MUST read user input from the command line and execute corresponding commands
- **FR-002**: System MUST display command output to standard output and errors to standard error
- **FR-003**: System MUST provide a prompt indicating readiness for input
- **FR-004**: System MUST support interrupting running commands with Ctrl+C
- **FR-005**: System MUST return appropriate exit codes (0 for success, non-zero for failure)

#### File System Management

- **FR-006**: System MUST provide commands for navigating directories (change directory, print working directory)
- **FR-007**: System MUST provide commands for listing directory contents with configurable detail levels
- **FR-008**: System MUST provide commands for creating and removing files and directories
- **FR-009**: System MUST provide commands for copying and moving files and directories
- **FR-010**: System MUST support both relative and absolute paths in all file operations

#### Command Syntax Consistency

- **FR-011**: All built-in commands MUST support `--help` to display usage information
- **FR-012**: All built-in commands supporting verbose output MUST use `--verbose` or `-v` flag
- **FR-013**: All built-in commands supporting quiet mode MUST use `--quiet` or `-q` flag
- **FR-014**: Command and argument names MUST use English words that clearly describe their function
- **FR-015**: System MUST use consistent argument naming across all commands (e.g., `--recursive` not sometimes `--recurse`)

#### Command Structure

- **FR-049**: Commands follow standard Unix naming convention (`cd`, `ls`, `cp`, `mv`, `rm`, `mkdir`, `pwd`)
- **FR-050**: Command syntax follows standard shell conventions: `<command> [options] [arguments]`
- **FR-051**: Arguments follow standard Unix order (source before destination for file operations)

#### Command Abbreviation (Unambiguous Prefix Matching)

- **FR-052**: System MUST accept abbreviated command names when the prefix uniquely identifies a single command
- **FR-053**: If user types `l` and only `ls` starts with `l`, the system MUST execute `ls`
- **FR-054**: If user types an ambiguous prefix (e.g., `c` matching `cd`, `cp`, `clear`), the system MUST display an error listing all matching commands
- **FR-055**: User MUST type enough characters to disambiguate (e.g., `cp` for `cp`, `cl` for `clear`, `cd` for `cd`)
- **FR-056**: Abbreviation matching MUST apply only to built-in commands, not external programs
- **FR-057**: System MUST provide a configuration option to disable abbreviation matching for users who prefer explicit commands

#### User Interface

- **FR-016**: System MUST support color output for terminal displays that support it
- **FR-017**: System MUST automatically detect terminal color capability and disable colors when not supported
- **FR-018**: System MUST strip color codes when output is redirected to a file or pipe
- **FR-019**: System MUST provide Tab-based autocompletion for commands, file paths, and arguments
- **FR-020**: System MUST display multiple completion options when Tab is pressed with ambiguous input

#### Inline Autocompletion

- **FR-058**: System MUST display inline autocompletion suggestions as the user types (ghost text style)
- **FR-059**: User MUST be able to accept inline suggestion by pressing Tab
- **FR-060**: Inline suggestions MUST update dynamically as the user continues typing
- **FR-061**: System MUST autocomplete file system paths (directories and files) as the user types
- **FR-062**: Inline autocompletion MUST be visually distinct from user input (e.g., dimmed or different color)

#### Command Line Editing

- **FR-063**: User MUST be able to move cursor left/right within the current input line using arrow keys
- **FR-064**: User MUST be able to move cursor to beginning of line (Home or Ctrl+A) and end of line (End or Ctrl+E)
- **FR-065**: User MUST be able to delete character before cursor (Backspace) and at cursor (Delete)
- **FR-066**: User MUST be able to delete from cursor to end of line (Ctrl+K) and from cursor to beginning (Ctrl+U)
- **FR-067**: User MUST be able to move cursor word-by-word (Ctrl+Left/Right or Alt+B/Alt+F)
- **FR-068**: User MUST be able to delete word before cursor (Ctrl+W) and word after cursor (Alt+D)
- **FR-069**: System MUST support inserting characters at cursor position (insert mode by default)

#### Configuration

- **FR-021**: System MUST load configuration from `~/.config/jsishell/config.yaml` by default
- **FR-022**: System MUST support specifying an alternate configuration file path via command-line argument
- **FR-023**: System MUST use sensible defaults when no configuration file is present
- **FR-024**: System MUST validate configuration file syntax and report errors clearly
- **FR-025**: System MUST support reloading configuration without restarting the shell

#### History

- **FR-026**: System MUST persist command history across sessions
- **FR-027**: System MUST support navigating history with Up/Down arrow keys
- **FR-028**: System MUST support searching command history
- **FR-029**: System MUST respect configurable history size limits

#### Program Execution

- **FR-030**: System MUST be able to execute external programs by name or path
- **FR-031**: System MUST support passing arguments to external programs
- **FR-032**: System MUST inherit and manage environment variables for child processes

#### Scripting Scope (Phased)

- **FR-033**: Phase 1 (MVP): System MUST support environment variable substitution in commands (e.g., `$HOME`, `$PATH`)
- **FR-034**: Phase 1 (MVP): System MUST NOT include scripting constructs (conditionals, loops, functions) - deferred to future phases
- **FR-035**: Future Phase 2: Basic scripting with conditionals (if/else) and loops (for/while)
- **FR-036**: Future Phase 3: Full scripting with functions, arrays, and advanced control flow

#### Pipes and Redirection (Phased)

- **FR-037**: Phase 1 (MVP): System MUST NOT support pipes or redirection - deferred to future phase
- **FR-038**: Future Phase: System MUST support pipes (`|`) for chaining command output to input
- **FR-039**: Future Phase: System MUST support output redirection (`>`, `>>`, `2>`, `&>`)
- **FR-040**: Future Phase: System MUST support input redirection (`<`) and here-documents (`<<`)

#### Aliases (Phased)

- **FR-041**: Phase 1 (MVP): System MUST NOT support aliases - deferred to future phase
- **FR-042**: Future Phase: System MUST support alias definition and expansion

#### Built-in Commands (Phased)

- **FR-043**: System provides these built-in commands: `cd`, `pwd`, `ls`, `cp`, `mv`, `rm`, `mkdir`, `exit`, `help`, `clear`, `echo`, `env`, `reload`, `history`
- **FR-044**: Additional commands may be added on-demand as requested

#### Job Control (Phased)

- **FR-045**: Phase 1 (MVP): System MUST execute commands in foreground only - no background job support
- **FR-046**: Future Phase: System MUST support background execution with `&` suffix
- **FR-047**: Future Phase: System MUST provide job control commands: `jobs`, `fg`, `bg`
- **FR-048**: Future Phase: System MUST support job suspension with Ctrl+Z (SIGTSTP)

### Key Entities

- **Command**: Represents a user input to be parsed and executed; includes command name, arguments, and options
- **Configuration**: User preferences loaded from YAML file; includes prompt style, colors, history settings
- **History Entry**: A previously executed command with timestamp; stored persistently for recall
- **Completion Candidate**: A suggested completion for partial input; includes command names, file paths, argument values
- **Environment**: Collection of environment variables available to the shell and child processes

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can execute their first command within 5 seconds of launching the shell
- **SC-002**: Shell startup time is under 500 milliseconds on standard hardware
- **SC-003**: Autocompletion suggestions appear within 100 milliseconds of pressing Tab
- **SC-004**: 95% of common file operations (list, copy, move, delete) complete without consulting documentation
- **SC-005**: Users can customize at least 10 different aspects of shell behavior through configuration
- **SC-006**: Command history retains at least 1000 entries across sessions
- **SC-007**: Shell operates correctly on Linux, Windows, and macOS without platform-specific user actions
- **SC-008**: Color output is correctly displayed on 95% of modern terminal emulators
- **SC-009**: All built-in commands respond to `--help` with useful documentation
- **SC-010**: Shell gracefully handles and recovers from 100% of invalid user inputs without crashing

## Clarifications

### Session 2025-01-25

- Q: What is the scripting language scope for MVP? → A: Interactive only for phase 1; basic scripting (conditionals, loops) in phase 2; full scripting (functions, arrays) in phase 3
- Q: Should MVP include pipe and redirection support? → A: No pipes/redirection in phase 1; full support (pipes, all redirections, here-documents) in future phase
- Q: Should MVP include alias support? → A: No alias support in MVP - defer to future phase
- Q: What built-in commands are required for MVP? → A: Core set in P1 (`goto`, `here`, `list`, `copy`, `move`, `remove`, `makedir`, `exit`, `help`, `clear`, `echo`, `env`); extended set (`cat`, `touch`, `find`, `grep`, `which`, `type`, `export`, `unset`) in later phases
- Q: Should MVP include background job execution? → A: No background jobs in MVP - foreground execution only; full job control (`&`, `jobs`, `fg`, `bg`, Ctrl+Z) deferred to future phase
- Q: Command syntax structure? → A: Natural language structure (`<verb> <subject> [destination]`) with unambiguous prefix abbreviation support (e.g., `l` for `list` if unique, `co` for `copy` when `c` is ambiguous)

### Session 2025-11-25 (Implementation)

- Q: Command naming convention for navigation? → A: Use natural language names instead of Unix abbreviations:
  - `cd` renamed to `goto` (more intuitive: "go to this directory")
  - `pwd` renamed to `here` (more intuitive: "where am I? here")
  - `mkdir` renamed to `makedir` (consistent with natural language pattern)

### Session 2025-11-26 (Phase 7-8 Implementation)

#### Color Display Implementation Details

- Q: How to implement color support detection? → A: Multi-factor detection:
  1. Check `NO_COLOR` environment variable (standard: https://no-color.org/)
  2. Check `TERM` variable (disable for "dumb" terminals)
  3. Check if output is a TTY using `term.IsTerminal()`
  4. All three conditions must pass for colors to be enabled

- Q: How to structure color configuration? → A: ColorScheme struct with method-based API:
  ```go
  type ColorScheme struct {
      enabled     bool
      Directory   string  // ANSI color name (e.g., "blue", "cyan")
      Executable  string
      File        string
      Error       string
      Suggestion  string  // For inline autocompletion ghost text
  }
  ```
  Methods like `Directory(text)`, `Error(text)` apply colors if enabled.

- Q: How to handle error messages in color? → A: Added `WriteErrorln` method to builtins.Context that automatically applies red color to error messages when colors are enabled.

#### Inline Autocompletion Implementation Details

- Q: Where does completion logic live? → A: Separated into `internal/completion/` package with:
  - `completion.go`: Core Completer struct, CompletionCandidate, CompletionType
  - `completion_test.go`: Command and path completion tests
  - `inline_test.go`: InlineSuggestion tests
  - `relative_path_test.go`: Relative path completion tests

- Q: How does inline suggestion work? → A:
  1. `InlineSuggestion(input)` returns the text to append (not the full completion)
  2. For single candidate: returns suffix to complete the word
  3. For multiple candidates: returns common prefix suffix (if any)
  4. Ghost text displayed in dimmed color after cursor

- Q: How to handle path completion after a command? → A:
  - `Complete(input)` detects context: command completion if no space, path completion if space exists
  - `InlineSuggestion` extracts the last word (after last space) for path completion comparison
  - Candidates include full path, comparison done against typed prefix

- Q: How to handle relative path completion? → A: Critical implementation details:
  1. `filepath.Dir("int")` returns "." and `filepath.Base("int")` returns "int"
  2. For paths without explicit "./" prefix: return just the filename (e.g., "internal/")
  3. For paths with "./" prefix: preserve it in the result (e.g., "./internal/")
  4. Always append trailing "/" for directories (filepath.Join strips it, so add manually)

- Q: How is Tab completion integrated with LineEditor? → A:
  - `CompletionProvider` interface with `InlineSuggestion()` and `GetCompletionList()` methods
  - Single Tab: accepts current ghost text suggestion
  - Double Tab (within 500ms): displays all completion candidates
  - Ghost text updated after each keystroke via `updateGhostText()`

- Q: How are completion candidates rendered? → A:
  - Displayed below the input line in columns
  - Different colors for TypeCommand, TypeDirectory, TypeFile, TypeExecutable
  - After display, cursor returns to input line for continued editing

### Session 2025-11-29 (v1.2.0 Features)

#### Prompt Shell Indicator

- Q: How to show user privilege level in prompt? → A: Added `%$` prompt variable that displays:
  - `$` for regular users (uid != 0)
  - `#` for root user (uid == 0)
  - Default prompt now uses `%$` instead of literal `$`

#### Exclude Option for File Commands

- Q: How to support repeated options like `--exclude`? → A:
  - Added `MultiOptions map[string][]string` to Command struct
  - Added `GetOptions(names ...string) []string` method
  - Parser populates both `Options` (last value) and `MultiOptions` (all values)

- Q: Which commands support `--exclude`? → A: `ls`, `cp`, and `rm`
  - Uses `filepath.Match()` for glob pattern matching
  - Pattern matched against base filename only
  - `rm -r` with exclude: directories containing excluded files are not removed

#### Tilde Expansion

- Q: Was tilde expanded for external commands? → A: No, fixed in parser.go
  - Added `expandTilde()` function in parser
  - Now `vim ~/.config/file` works correctly with external programs
  - Expands `~` to home directory, `~/path` to home + path

#### Path Autocompletion

- Q: How to handle `../` path completion? → A:
  - Added `..` and `../` detection in `isPathLike()`
  - Added `hasExplicitDotDot` tracking for prefix preservation
  - Fixed double slash issue when completing `../` alone

### Session 2025-11-29 (v1.2.1 Features)

#### PATH Executable Completion

- Q: How to autocomplete external commands from PATH? → A:
  - Added `EnablePathCompletion(pathEnv string)` to Completer
  - Scans each directory in PATH for executable files (mode & 0111)
  - First occurrence in PATH wins (earlier directories have priority)
  - Builtin commands take priority over PATH executables with same name
  - Integration: `shell.createCompleter()` calls `EnablePathCompletion()` automatically

- Q: What is the completion behavior? → A:
  - Typing `vi` suggests `vim`, `view`, etc. from PATH
  - Typing `pyt` suggests `python`, `python3`, etc.
  - Builtins like `cd`, `ls` are never shadowed by PATH executables
  - Results sorted alphabetically, builtins listed first

## Assumptions

- Users have basic familiarity with command-line interfaces
- Target terminals support UTF-8 encoding
- YAML is an acceptable format for configuration (widely supported, human-readable)
- Default configuration location follows XDG Base Directory specification (`~/.config/`)
- English is the primary language for command syntax (as specified in requirements)
- Standard terminal key bindings (Ctrl+C for interrupt, Tab for completion) are expected

---

## Implementation Summary

### Packages Implemented

| Package | Purpose |
|---------|---------|
| `internal/lexer` | Tokenization of command input |
| `internal/parser` | AST generation from tokens |
| `internal/executor` | Command execution engine |
| `internal/builtins` | Built-in command implementations |
| `internal/completion` | Inline autocompletion |
| `internal/history` | Command history management |
| `internal/config` | YAML configuration loading |
| `internal/terminal` | Terminal I/O and line editing |
| `internal/shell` | Main REPL orchestration |
| `internal/env` | Environment variable handling |
| `internal/errors` | Custom error types |

### Built-in Commands

| Command | Description | Options |
|---------|-------------|---------|
| `echo` | Print arguments | `-n`, `--help` |
| `exit` | Exit the shell | `--help` |
| `help` | Show help | `--help` |
| `clear` | Clear screen | `--help` |
| `env` | Show/set environment variables | `--help` |
| `cd` | Change directory | `--help` |
| `pwd` | Print working directory | `--help` |
| `ls` | List directory contents | `-a`, `-d`, `-l`, `-R`, `-v`, `-q`, `-s/--sort`, `-e/--exclude`, `--help` |
| `mkdir` | Create directories | `-p`, `-v`, `--help` |
| `cp` | Copy files/directories | `-r`, `-f`, `-v`, `-e/--exclude`, `--help` |
| `mv` | Move/rename files | `-f`, `-v`, `--help` |
| `rm` | Delete files/directories | `-r`, `-f`, `-y`, `-q`, `-v`, `-e/--exclude`, `--help` |
| `init` | Generate default config | `--help` |
| `reload` | Reload configuration | `--help` |
| `history` | Show/clear command history | `-c`, `--help` |

### ls Command Details

The `ls` command supports extensive options:
- `-a, --all` - Include hidden files (starting with .)
- `-d, --directory` - List directories only
- `-l, --long` - Long format (permissions, owner, group, size, date, name)
- `-R, --recursive` - List subdirectories recursively
- `-v, --verbose` - Verbose mode (implies -l, adds file type indicator)
- `-q, --quiet` - Only show file names
- `-s, --sort=<spec>` - Sort specification (name, size, time, dir; prefix with ! to reverse)
- `-e, --exclude=<glob>` - Exclude files matching glob pattern (can be repeated)

Sort examples:
- `--sort=name` - Sort by name (default)
- `--sort=!time` - Sort by modification time, newest first
- `--sort=dir,name` - Directories first, then by name
- `--sort=!size,name` - By size descending, then by name

### Exclude Pattern Support (ls, cp, rm)

The `--exclude` option uses standard glob patterns and can be repeated:
- `ls -e=*.log` - List files excluding .log files
- `cp -r -e=*.tmp -e=*.bak src/ dest/` - Copy excluding .tmp and .bak files
- `rm -r --exclude=*.txt dir/` - Remove directory keeping .txt files

### Performance Results

| Benchmark | Target | Achieved |
|-----------|--------|----------|
| Shell startup time | < 100ms | **0.14ms** |
| Input latency | < 10ms | **< 0.01ms** |

### Platform Support

| Platform | Architecture |
|----------|--------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 (Apple Silicon) |
| Windows | amd64 |

### User Stories Status

| Story | Status |
|-------|--------|
| US1 - Execute Commands | ✅ Complete |
| US2 - File System Navigation | ✅ Complete |
| US3 - Inline Autocompletion | ✅ Complete |
| US4 - Color Display | ✅ Complete |
| US5 - Configuration Management | ✅ Complete |
| US6 - Standard Unix Commands | ✅ Complete |
| US7 - Consistent Command Syntax | ✅ Complete |
| US8 - Command Line Editing | ✅ Complete |
| US9 - Command History | ✅ Complete |

### Potential Future Features (On-Demand)

The following features may be implemented if requested:
- Pipes and redirections (`|`, `>`, `<`, `>>`)
- Shell scripting (conditionals, loops, functions)
- Aliases
- Job control (background processes, `&`, `fg`, `bg`)
- Additional builtins (`cat`, `touch`, `find`, `grep`, etc.)
