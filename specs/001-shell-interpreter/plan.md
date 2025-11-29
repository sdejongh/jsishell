# Implementation Plan: Shell Interpreter

**Branch**: `master` | **Date**: 2025-01-25 | **Spec**: [spec.md](./spec.md)
**Status**: Active Development (on-demand)

## Development Approach

This project follows an **on-demand development** approach. There is no fixed roadmap. Features and improvements are implemented as requested by the user.

## Summary

JSIShell is a cross-platform interactive command line interpreter in Go with standard Unix command syntax, inline autocompletion, configurable behavior via YAML, and modern line editing capabilities. The shell provides 14 built-in commands with extensive options (especially `ls`), foreground execution, and produces a single statically-linked executable with no external runtime dependencies.

## Technical Context

**Language/Version**: Go 1.21+ (statically compiled, no external dependencies)
**Primary Dependencies**: Standard library only + `golang.org/x/term` for terminal handling
**Storage**: File-based (YAML config at `~/.config/jsishell/config.yaml`, history at `~/.jsishell_history`)
**Testing**: `go test` with table-driven tests, `go test -race` for concurrency
**Target Platform**: Linux, Windows, macOS (cross-platform, single binary)
**Project Type**: Single CLI application
**Performance Goals**: <100ms startup, <10ms input latency, <100ms autocompletion
**Constraints**: No external runtime dependencies, <10MB binary size, minimal memory footprint
**Scale/Scope**: Single-user interactive shell, 1000+ history entries, 15 built-in commands

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Language Standard | ✅ PASS | Go 1.21+ required, standard library focus |
| II. Cross-Platform Compatibility | ✅ PASS | Linux/Windows/macOS support required |
| III. Test-First (NON-NEGOTIABLE) | ✅ PASS | 80%+ coverage required, table-driven tests |
| IV. Documentation & Comments | ✅ PASS | GoDoc-compliant comments required |
| V. Go Coding Standards | ✅ PASS | gofmt, go vet, staticcheck compliance |
| VI. English-Only Policy | ✅ PASS | All code/docs in English |
| VII. Performance Optimization | ✅ PASS | <100ms startup, <10ms latency targets |
| VIII. Interactive Shell Standards | ✅ PASS | REPL, history, line editing, completion |

**Gate Status**: ✅ ALL PRINCIPLES SATISFIED

## Project Structure

### Documentation (this feature)

```text
specs/001-shell-interpreter/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (internal APIs)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── jsishell/
    └── main.go          # Entry point

internal/
├── shell/               # Main shell orchestration
│   ├── shell.go         # Shell struct and main loop
│   └── shell_test.go
├── lexer/               # Tokenization
│   ├── lexer.go         # Token types and lexer
│   ├── token.go         # Token definitions
│   └── lexer_test.go
├── parser/              # Command parsing
│   ├── parser.go        # AST construction
│   ├── ast.go           # AST node types
│   └── parser_test.go
├── executor/            # Command execution
│   ├── executor.go      # Execute parsed commands
│   └── executor_test.go
├── builtins/            # Built-in commands
│   ├── registry.go      # Command registry + Context type
│   ├── cd.go            # cd command (change directory)
│   ├── pwd.go           # pwd command (print working directory)
│   ├── ls.go            # ls command with extensive options + exclude
│   ├── cp.go            # cp command (copy) + exclude
│   ├── mv.go            # mv command (move)
│   ├── rm.go            # rm command (remove) + exclude
│   ├── mkdir.go         # mkdir command
│   ├── init.go          # init command (generate default config)
│   ├── exit.go          # exit command
│   ├── help.go          # help command
│   ├── clear.go         # clear command
│   ├── echo.go          # echo command
│   ├── env_cmd.go       # env command
│   ├── history.go       # history command
│   ├── reload.go        # reload config command
│   └── builtins_test.go
├── completion/          # Autocompletion engine
│   ├── completion.go    # Completer, CompletionCandidate, CompletionType
│   ├── completion_test.go
│   ├── inline_test.go   # InlineSuggestion tests
│   └── relative_path_test.go  # Relative path completion tests
├── history/             # Command history (Phase 10)
│   ├── history.go       # History management
│   ├── persistence.go   # File I/O
│   └── history_test.go
├── config/              # Configuration management
│   ├── config.go        # Config struct, ColorConfig, loading
│   └── config_test.go
├── terminal/            # Terminal I/O
│   ├── terminal.go      # Raw mode, key reading
│   ├── editor.go        # Line editing + completion integration
│   ├── color.go         # ColorScheme, ANSI codes, detection
│   ├── color_test.go
│   ├── terminal_test.go
│   └── editor_test.go
├── env/                 # Environment management
│   ├── env.go           # Environment variables
│   └── env_test.go
└── errors/              # Error definitions
    ├── errors.go        # Sentinel errors
    └── errors_test.go

tests/
├── integration/         # Integration tests
│   ├── shell_test.go
│   └── commands_test.go
└── fixtures/            # Test data
    └── config.yaml
```

**Structure Decision**: Single project structure following Go conventions with `cmd/` for entry point and `internal/` for private packages. This aligns with the constitution's shell architecture requirements (lexer, parser, executor, builtins, history, completion, terminal).

## Complexity Tracking

> No violations - all principles satisfied with standard architecture.

| Consideration | Decision | Rationale |
|---------------|----------|-----------|
| External YAML library | Use `gopkg.in/yaml.v3` | Standard library lacks YAML; this is the de-facto Go YAML library, well-maintained |
| Terminal library | Use `golang.org/x/term` | Official Go sub-repository, cross-platform terminal handling |
| No other external deps | Standard library only | Minimizes binary size, security surface, and maintenance burden |

## Implementation Notes (Lessons Learned)

### Phase 7: Color Display

#### Architecture Decisions

1. **ColorScheme as Method-Based API**: Rather than exposing raw color codes, ColorScheme provides methods like `Directory(text)`, `Error(text)` that return colored strings only if colors are enabled. This centralizes the enable/disable logic.

2. **Color Detection Order**:
   ```go
   func ShouldUseColor() bool {
       if os.Getenv("NO_COLOR") != "" { return false }
       if os.Getenv("TERM") == "dumb" { return false }
       return term.IsTerminal(int(os.Stdout.Fd()))
   }
   ```

3. **Integration Pattern**: ColorScheme is created in Shell from config, then passed to:
   - Executor → for builtins via Context
   - LineEditor → for ghost text rendering
   - Builtins access colors via `ctx.Colors`

4. **Error Coloring**: Added `WriteErrorln(format, args...)` to `builtins.Context` that automatically applies red color. All builtins use this for error output.

#### Key Files
- `internal/terminal/color.go`: ColorScheme struct, ANSI codes, ShouldUseColor()
- `internal/builtins/registry.go`: Context.Colors field, WriteErrorln method
- `internal/builtins/list.go`: colorizeEntry() for directory listings

### Phase 8: Inline Autocompletion

#### Architecture Decisions

1. **Separate Package**: Completion logic in `internal/completion/` is independent of terminal handling. This allows testing without terminal dependencies.

2. **CompletionProvider Interface**: Decouples LineEditor from Completer:
   ```go
   type CompletionProvider interface {
       InlineSuggestion(input string) (suggestion string, has bool)
       GetCompletionList(input string) []string
   }
   ```

3. **Ghost Text Rendering**: Ghost text (suggestion) is rendered in dimmed color after the cursor, then cursor is repositioned back. Uses ANSI escape sequences.

4. **Tab Behavior**:
   - Single Tab: Accept current ghost text (append to input)
   - Double Tab (within 500ms): Display all completion candidates below input line

5. **Path Completion Complexity**:
   - Context detection: No space = command completion, has space = path completion
   - Last word extraction: For "list /home/us", extract "/home/us" for completion
   - Relative paths: Must preserve "./" prefix and trailing "/" for directories
   - `filepath.Join` strips trailing slashes - must add manually for directories

#### Critical Bug Fixes

1. **InlineSuggestion for paths**: Initially compared full candidate path with full input. Fixed to compare only the relevant portion (last word for paths).

2. **Relative path completion**: `filepath.Dir("int")` returns "." but candidates were returned as `./internal/` breaking prefix matching. Fixed by returning `internal/` for implicit relative paths and `./internal/` only when user typed `./`.

3. **Trailing slash preservation**: `filepath.Join(dir, name)` strips trailing `/` from directory names. Now append `/` after joining for directories.

#### Key Files
- `internal/completion/completion.go`: Completer, CompletePath, InlineSuggestion
- `internal/terminal/editor.go`: SetCompleter, updateGhostText, handleTab, showCompletionList
- `internal/shell/shell.go`: Creates Completer with command list, passes to LineEditor

### General Patterns

1. **Test-First Discovery**: Writing tests first (especially `relative_path_test.go`) revealed edge cases that weren't obvious from the spec.

2. **Temporary Directory for Tests**: Path completion tests use `t.TempDir()` and `os.Chdir()` to create controlled test environments.

3. **Rebuild Required**: After code changes, always rebuild binary: `go build -o jsishell ./cmd/jsishell`

### v1.2.0 Features

#### Prompt Shell Indicator (`%$`)

1. **Implementation**: Added `%$` variable in `internal/terminal/prompt.go`:
   ```go
   func (p *PromptExpander) shellIndicator() string {
       if os.Getuid() == 0 {
           return "#"
       }
       return "$"
   }
   ```

2. **Default Prompt Update**: Changed from literal `$` to `%$` in config.go and init.go to properly show root indicator.

#### Multi-Value Option Parsing (`--exclude`)

1. **Parser Enhancement**: Added `MultiOptions map[string][]string` to `Command` struct and `GetOptions()` method in `ast.go`.

2. **Option Population**: Both `Options` (last value) and `MultiOptions` (all values) are populated during parsing.

3. **Usage Pattern**:
   ```go
   opts.excludePatterns = cmd.GetOptions("-e", "--exclude")
   ```

#### Exclude Pattern Support

1. **Glob Matching**: Uses `filepath.Match()` against base filename only.

2. **rm -r with Exclude**: Custom `rmDirRecursive()` function that:
   - Recursively processes directory contents
   - Skips files matching exclude patterns
   - Only removes directory if empty after filtering

3. **cp with Exclude**: Filters entries during `cpDir()` recursive copy.

#### Tilde Expansion Fix

1. **Problem**: External commands like `vim ~/.config/file` received literal `~` instead of expanded path.

2. **Solution**: Added `expandTilde()` in `parser.go` that:
   - Checks if value starts with `~`
   - Gets HOME from environment or `os.UserHomeDir()`
   - Expands `~` alone or `~/path` format

#### Path Autocompletion Fix (`../`)

1. **Detection**: Added `..` and `../` to `isPathLike()` function.

2. **Prefix Preservation**: Added `hasExplicitDotDot` variable to track when user typed `../`.

3. **Double Slash Fix**: When `pathPrefix` is `..` or ends with `/`, use `strings.TrimSuffix()` to avoid `..//` results.

#### init Command

1. **Purpose**: Generate default configuration file at `~/.config/jsishell/config.yaml`.

2. **Behavior**:
   - Creates directory structure if needed
   - Uses embedded default config template
   - Will not overwrite existing config unless `--force` used

### v1.2.1 Features

#### PATH Executable Completion

1. **Implementation**: Added `EnablePathCompletion()` method to `Completer` struct:
   ```go
   func (c *Completer) EnablePathCompletion(pathEnv string) {
       c.pathExecutable = true
       if pathEnv == "" {
           pathEnv = os.Getenv("PATH")
       }
       c.pathDirs = filepath.SplitList(pathEnv)
   }
   ```

2. **Execution Flow**:
   - `completePathExecutables()` scans each PATH directory
   - Filters by executable permission (mode & 0111)
   - First occurrence wins (earlier PATH directories take priority)
   - Results merged with builtin commands (builtins take priority)

3. **Integration**: `shell.createCompleter()` calls `EnablePathCompletion()` with environment PATH.

4. **Behavior**:
   - Typing `vi` suggests `vim`, `view`, etc.
   - Typing `pyt` suggests `python`, `python3`, etc.
   - Builtins like `cd` are not shadowed by PATH executables

### v1.3.0 Features

#### search Command

1. **Purpose**: Find files and directories matching glob patterns.

2. **Implementation**: `internal/builtins/search.go`
   ```go
   func SearchDefinition() Definition {
       return Definition{
           Name:        "search",
           Description: "Search for files and directories",
           Usage:       "search <directory> <pattern>... [options]",
           Handler:     searchHandler,
           Options: []OptionDef{
               {Long: "--recursive", Short: "-r", Description: "Search recursively"},
               {Long: "--level", Short: "-l", HasValue: true, Description: "Max depth (0=unlimited)"},
               {Long: "--help", Description: "Show help message"},
           },
       }
   }
   ```

3. **Arguments**:
   - First argument: directory to search in (required)
   - Remaining arguments: glob patterns to match (at least one required)

4. **Options**:
   - `-r, --recursive`: Search subdirectories recursively
   - `-l, --level=<n>`: Maximum depth level (0 = unlimited, default)
   - `--help`: Show help message

5. **Behavior**:
   - Non-recursive by default (only searches specified directory)
   - Supports multiple patterns (OR logic)
   - Uses `filepath.Match()` for glob matching
   - Colorized output based on file type (directory, executable, file)
   - Depth level controls how deep to recurse (0 = no limit)

6. **Simple Examples** (backward compatible):
   ```
   search . "*.go"                    # Find Go files in current dir
   search /home "*.txt"               # Find text files in /home
   search . "*.go" "*.md" -r          # Find Go and Markdown recursively (implicit OR)
   search . "test_*" -r               # Find files starting with test_
   search . "*.log" -r --level=2      # Find logs up to 2 levels deep
   ```

#### search Logical Expressions (v1.3.0)

1. **Logical Operators**: Full boolean expression support with:
   - `AND`, `&&` - Both patterns must match
   - `OR`, `||` - Either pattern must match
   - `XOR`, `^` - Exactly one pattern must match
   - `NOT`, `!` - Negate the following pattern
   - `( )` - Group expressions

2. **Operator Precedence** (highest to lowest):
   - NOT (unary)
   - AND
   - XOR
   - OR

3. **Implementation** (`internal/builtins/searchexpr.go`):
   - `SearchExpr` interface with `Evaluate(info *FileInfo) bool` method
   - `FileInfo` struct with Name, Mode, IsDir, IsLink, LinkTarget fields
   - AST nodes: `PatternExpr`, `AndExpr`, `OrExpr`, `XorExpr`, `NotExpr`, `TypeExpr`
   - Recursive descent parser for expression parsing
   - Tokenizer handles patterns, operators, type predicates, and parentheses

4. **Parser Architecture**:
   ```go
   // Grammar (recursive descent):
   // expr     = orExpr
   // orExpr   = xorExpr (("OR" | "||") xorExpr)*
   // xorExpr  = andExpr (("XOR" | "^") andExpr)*
   // andExpr  = unaryExpr (("AND" | "&&") unaryExpr)*
   // unaryExpr = ("NOT" | "!") unaryExpr | primary
   // primary  = pattern | typePredicate | "(" expr ")"
   ```

5. **Backward Compatibility**:
   - Multiple patterns without operators use implicit OR
   - Single pattern works exactly as before
   - Operators only activated when detected in arguments

6. **Logical Expression Examples**:
   ```
   search . "*.go" AND NOT "*_test.go" -r     # Go files excluding tests
   search . "test_*" AND "*.py" -r            # Python test files
   search . "*.go" XOR "*.mod" -r             # .go OR .mod but not both
   search . NOT "*.log" -r                    # All files except .log
   search . "(" "*.go" OR "*.md" ")" AND NOT "*_test*" -r
       # Go or Markdown files, excluding test files
   ```

#### search Type Predicates

1. **Type Predicates** (case-insensitive):
   - `isFile` - Match regular files only (not directories or symlinks)
   - `isDir` - Match directories only
   - `isLink` - Match symbolic links
   - `isSymlink` - Match symbolic links (alias for isLink)
   - `isHardlink` - Match regular files (hard links cannot be reliably detected)
   - `isExec` - Match executable files (mode & 0111 != 0)

2. **Implementation**:
   - `TypePredicate` enum in `internal/builtins/searchexpr.go`
   - `TypeExpr` struct implementing `SearchExpr` interface
   - `isTypePredicate()` function for detecting predicates (case-insensitive)
   - `FileInfo` struct passed to `Evaluate()` with file metadata

3. **Type Predicate Examples**:
   ```
   search . "*.go" AND isFile -r       # Find Go files only (exclude directories)
   search . isDir -r                   # Find all directories
   search . isExec -r                  # Find all executable files
   search . isLink -r                  # Find all symbolic links
   search . "*config*" AND isDir -r    # Find directories with "config" in name
   search . "*.sh" AND isExec -r       # Find executable shell scripts
   search . NOT isDir -r               # Find everything except directories
   ```

#### Quoting for Literal Patterns

1. **Problem**: Files named like operators (AND, OR, NOT, XOR) or predicates (isFile, isDir) would be interpreted as operators/predicates instead of patterns.

2. **Solution**: Quoted arguments are always treated as literal patterns, never as operators or predicates.

3. **Implementation**:
   - `parser.Arg` struct with `Value` and `Quoted` fields in `internal/parser/ast.go`
   - Parser populates `ArgsWithInfo` with quoting information from lexer tokens
   - `ExprArg` struct in `searchexpr.go` carries quoting info
   - `ParseSearchExpressionWithQuoting()` respects quoting when tokenizing
   - Quoted args bypass operator/predicate detection in lexer

4. **Quoting Examples**:
   ```
   search . "AND"                      # Find file named 'AND' (quoted = pattern)
   search . "isFile"                   # Find file named 'isFile' (quoted = pattern)
   search . "OR" "NOT"                 # Find files named 'OR' or 'NOT'
   search . "*.go" AND "isDir" -r      # Pattern *.go AND pattern 'isDir'
   ```

### v1.4.0 Features

#### Glob Expansion in Parser

1. **Purpose**: Expand wildcard patterns like `*.go` in command arguments.

2. **Implementation** (`internal/parser/parser.go`):
   ```go
   // containsGlobPattern checks if a string contains glob wildcards.
   func containsGlobPattern(s string) bool {
       for _, c := range s {
           switch c {
           case '*', '?', '[':
               return true
           }
       }
       return false
   }

   // expandGlob expands a glob pattern to matching file paths.
   func expandGlob(pattern string) []string {
       matches, err := filepath.Glob(pattern)
       if err != nil || len(matches) == 0 {
           return []string{pattern}  // No matches: return original
       }
       return matches
   }
   ```

3. **Behavior**:
   - Unquoted arguments containing `*`, `?`, or `[` are expanded
   - Quoted arguments (`"*.go"`) are never expanded
   - If no files match, original pattern is preserved (bash behavior)
   - Multiple matches result in multiple arguments

#### Windows Cross-Platform Support

1. **Windows Path Handling** (`internal/lexer/lexer.go`):
   - Backslash is NOT an escape character in unquoted words
   - Paths like `C:\Users\name` and `d:\pictures\` work correctly
   - For filenames with spaces, use quotes: `"file name"`

2. **Drive Letter Navigation** (`internal/executor/drive_windows.go`, `drive_unix.go`):
   ```go
   // Windows: isWindowsDriveLetter returns true for "c:", "D:", etc.
   func isWindowsDriveLetter(name string) bool {
       if len(name) != 2 || name[1] != ':' {
           return false
       }
       letter := name[0]
       return (letter >= 'a' && letter <= 'z') || (letter >= 'A' && letter <= 'Z')
   }
   ```
   - Typing `c:` or `D:` is converted to `cd c:` or `cd D:`
   - Unix version always returns false

3. **PATH Environment Variable** (`internal/env/path.go`):
   ```go
   // GetPath returns PATH value, handling Windows "Path" vs Unix "PATH"
   func GetPath() string {
       if path := os.Getenv("Path"); path != "" {
           return path
       }
       return os.Getenv("PATH")
   }
   ```

4. **Executable Detection** (`internal/completion/executable_windows.go`, `executable_unix.go`):
   - Unix: checks mode bits (mode & 0111)
   - Windows: checks extensions (.exe, .cmd, .bat, .com, .ps1)

5. **Owner/Group Display** (`internal/builtins/ls_windows.go`, `ls_unix.go`):
   - Unix: uses syscall.Stat_t for uid/gid lookup
   - Windows: returns `-` placeholders (Windows uses SIDs)

#### PATH Executable Completion Caching

1. **Problem**: Scanning PATH directories on every keystroke caused latency.

2. **Solution**: Cache executables at startup.
   ```go
   type Completer struct {
       // ... existing fields ...
       pathExecCache    []string        // Cached executable names
       pathExecCacheSet map[string]bool // For deduplication
   }

   func (c *Completer) EnablePathCompletion(pathEnv string) {
       c.pathExecutable = true
       c.pathDirs = filepath.SplitList(pathEnv)
       c.scanPathExecutables()  // Scan once at startup
   }

   func (c *Completer) scanPathExecutables() {
       c.pathExecCache = nil
       c.pathExecCacheSet = make(map[string]bool)
       for _, dir := range c.pathDirs {
           entries, _ := os.ReadDir(dir)
           for _, entry := range entries {
               if !entry.IsDir() && isExecutable(entry.Info()) {
                   name := executableName(entry.Name())
                   if !c.pathExecCacheSet[name] {
                       c.pathExecCacheSet[name] = true
                       c.pathExecCache = append(c.pathExecCache, name)
                   }
               }
           }
       }
   }
   ```

3. **Completion now O(n) string filtering** instead of O(dirs × files) filesystem scan.

#### search Command Enhancements

1. **Absolute Path Option**:
   - Added `-a, --absolute` flag
   - Uses `filepath.Abs()` to convert paths

2. **Windows Drive Letter Normalization**:
   - `d:` normalized to `d:\` for proper path joining
   - Prevents `filepath.Join("d:", "2023")` producing `d:2023`

3. **Permission Error Handling**:
   - System folders like "System Volume Information" are silently skipped
   - `os.IsPermission(err)` check before reporting errors
