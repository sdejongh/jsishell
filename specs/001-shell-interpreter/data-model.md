# Data Model: Shell Interpreter

**Feature**: 001-shell-interpreter
**Date**: 2025-01-25

## Overview

This document defines the core data structures and their relationships for the JSIShell interpreter.

---

## Entity Diagram

```text
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Shell       │────▶│   LineEditor    │────▶│   Completer     │
│                 │     │                 │     │                 │
│ - config        │     │ - buffer        │     │ - builtins      │
│ - env           │     │ - cursor        │     │ - history       │
│ - history       │     │ - prompt        │     │                 │
│ - executor      │     │ - ghostText     │     └─────────────────┘
└────────┬────────┘     └─────────────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Executor     │────▶│     Lexer       │────▶│     Parser      │
│                 │     │                 │     │                 │
│ - builtins      │     │ - input         │     │ - tokens        │
│ - env           │     │ - pos           │     │                 │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                                                         ▼
                                                ┌─────────────────┐
                                                │    Command      │
                                                │     (AST)       │
                                                │                 │
                                                │ - name          │
                                                │ - args          │
                                                │ - options       │
                                                └─────────────────┘
```

---

## Core Entities

### 1. Shell

The main orchestrator managing the REPL loop and coordinating all components.

```go
// Shell represents the main shell instance
type Shell struct {
    Config     *Config        // User configuration
    Env        *Environment   // Environment variables
    History    *History       // Command history
    Executor   *Executor      // Command executor
    Editor     *LineEditor    // Line editing
    Completer  *Completer     // Autocompletion
    Terminal   *Terminal      // Terminal I/O
    Running    bool           // REPL state
    ExitCode   int            // Last command exit code
}
```

**Lifecycle**:
1. `New()` → Initialize with defaults
2. `LoadConfig()` → Load YAML configuration
3. `Run()` → Start REPL loop
4. `Exit()` → Cleanup and terminate

---

### 2. Config

User preferences loaded from YAML file.

```go
// Config holds shell configuration
type Config struct {
    // Prompt settings
    Prompt          string `yaml:"prompt"`           // Default: "$ "
    ContinuationPS2 string `yaml:"continuation"`    // Default: "> "

    // History settings
    HistoryFile     string `yaml:"history_file"`    // Default: ~/.jsishell_history
    HistorySize     int    `yaml:"history_size"`    // Default: 1000
    HistoryIgnoreDups bool  `yaml:"history_ignore_dups"` // Default: true

    // Display settings
    Colors          bool   `yaml:"colors"`          // Default: true (auto-detect)
    ColorScheme     ColorScheme `yaml:"color_scheme"`

    // Behavior settings
    Abbreviations   bool   `yaml:"abbreviations"`   // Default: true
    CompletionLimit int    `yaml:"completion_limit"` // Default: 100

    // Paths
    ConfigPath      string `yaml:"-"` // Set at runtime
}

// ColorScheme defines colors for different elements
type ColorScheme struct {
    Directory  string `yaml:"directory"`  // Default: "blue"
    File       string `yaml:"file"`       // Default: "white"
    Executable string `yaml:"executable"` // Default: "green"
    Error      string `yaml:"error"`      // Default: "red"
    Warning    string `yaml:"warning"`    // Default: "yellow"
    GhostText  string `yaml:"ghost_text"` // Default: "gray"
}
```

**Validation Rules**:
- `HistorySize` must be > 0
- `CompletionLimit` must be > 0
- Color names must be valid ANSI colors

**File Location**: `~/.config/jsishell/config.yaml`

---

### 3. Environment

Collection of environment variables.

```go
// Environment manages shell environment variables
type Environment struct {
    vars    map[string]string // Current variables
    parent  map[string]string // Inherited from OS
    exports map[string]bool   // Variables to export to children
}
```

**Operations**:
- `Get(key)` → Get variable value
- `Set(key, value)` → Set variable
- `Export(key)` → Mark for export to child processes
- `ToSlice()` → Convert to `KEY=VALUE` slice for exec
- `Expand(input)` → Expand `$VAR` references in string

**Special Variables**:
| Variable | Description |
|----------|-------------|
| `$?` | Exit code of last command |
| `$PWD` | Current working directory |
| `$HOME` | User home directory |
| `$PATH` | Executable search path |

---

### 4. Token

Lexical token from input parsing.

```go
// TokenType identifies the type of token
type TokenType int

const (
    TokenWord       TokenType = iota // Command or argument
    TokenString                      // Quoted string
    TokenOption                      // --flag or -f
    TokenEquals                      // =
    TokenVariable                    // $VAR
    TokenWhitespace                  // Space/tab
    TokenNewline                     // End of line
    TokenEOF                         // End of input
)

// Token represents a lexical token
type Token struct {
    Type    TokenType
    Value   string    // Raw value
    Literal string    // Processed value (quotes removed, escapes applied)
    Pos     Position  // Source position for error reporting
}

// Position in source input
type Position struct {
    Line   int
    Column int
    Offset int
}
```

---

### 5. Command (AST)

Parsed command representation.

```go
// Command represents a parsed command
type Command struct {
    Name      string            // Command name (may be abbreviated)
    Resolved  string            // Resolved full command name
    Args      []string          // Positional arguments
    Options   map[string]string // Named options (--key=value)
    Flags     map[string]bool   // Boolean flags (--verbose)
    RawInput  string            // Original input string
}
```

**Validation Rules**:
- `Name` must not be empty
- `Resolved` set after abbreviation resolution
- Options keys must start with `-` or `--`

---

### 6. HistoryEntry

Single command history record.

```go
// HistoryEntry represents a command in history
type HistoryEntry struct {
    Command   string    // The command string
    Timestamp time.Time // When executed
    ExitCode  int       // Result code
}

// History manages command history
type History struct {
    entries  []HistoryEntry // In-memory entries
    maxSize  int            // Maximum entries to keep
    filepath string         // Persistence file path
    position int            // Current navigation position
}
```

**Operations**:
- `Add(entry)` → Add new entry, trim if over limit
- `Search(prefix)` → Find entries matching prefix
- `Navigate(direction)` → Move through history (Up/Down)
- `Load()` → Load from file on startup
- `Save()` → Persist to file

**File Format**:
```text
# One command per line, newest at end
list /home
cd /tmp
copy file.txt backup/
```

---

### 7. CompletionCandidate

Autocompletion suggestion.

```go
// CompletionCandidate represents a completion option
type CompletionCandidate struct {
    Value       string          // Completion text
    Display     string          // Display text (may include formatting)
    Description string          // Optional description
    Type        CompletionType  // Source type
}

// CompletionType identifies completion source
type CompletionType int

const (
    CompletionBuiltin CompletionType = iota // Built-in command
    CompletionFile                          // File path
    CompletionDir                           // Directory path
    CompletionHistory                       // From history
    CompletionOption                        // Command option
)
```

---

### 8. Builtin

Built-in command definition.

```go
// Builtin defines a built-in shell command
type Builtin struct {
    Name        string                              // Command name
    Description string                              // Short description
    Usage       string                              // Usage pattern
    Handler     func(ctx *Context, cmd *Command) error
    Options     []OptionDef                         // Supported options
}

// OptionDef defines a command option
type OptionDef struct {
    Long        string // --name
    Short       string // -n
    Description string
    HasValue    bool   // Takes a value?
    Default     string // Default value
}

// Context provides execution context to builtins
type Context struct {
    Shell   *Shell
    Stdin   io.Reader
    Stdout  io.Writer
    Stderr  io.Writer
    Env     *Environment
}
```

**MVP Built-in Commands**:
| Command | Description |
|---------|-------------|
| `cd` | Change directory |
| `pwd` | Print working directory |
| `list` | List directory contents |
| `copy` | Copy files/directories |
| `move` | Move/rename files |
| `remove` | Delete files/directories |
| `mkdir` | Create directory |
| `exit` | Exit shell |
| `help` | Show help |
| `clear` | Clear screen |
| `echo` | Print arguments |
| `env` | Show environment |

---

### 9. LineEditor

Interactive line editing state.

```go
// LineEditor handles interactive line input
type LineEditor struct {
    buffer      []rune  // Current input buffer
    cursor      int     // Cursor position in buffer
    prompt      string  // Primary prompt
    ghostText   string  // Inline completion suggestion
    history     *History
    completer   *Completer
    terminal    *Terminal
}
```

**Key Bindings**:
| Key | Action |
|-----|--------|
| Left/Right | Move cursor |
| Home/Ctrl+A | Move to start |
| End/Ctrl+E | Move to end |
| Backspace | Delete before cursor |
| Delete | Delete at cursor |
| Ctrl+K | Delete to end |
| Ctrl+U | Delete to start |
| Ctrl+W | Delete word back |
| Tab | Accept completion |
| Up/Down | Navigate history |
| Enter | Execute command |
| Ctrl+C | Cancel input |

---

## State Transitions

### Shell State

```text
┌─────────┐    LoadConfig    ┌─────────┐    Run     ┌─────────┐
│  Init   │ ───────────────▶ │  Ready  │ ────────▶ │ Running │
└─────────┘                  └─────────┘           └────┬────┘
                                                        │
                                   ┌────────────────────┤
                                   │                    │
                                   ▼                    ▼
                             ┌──────────┐        ┌──────────┐
                             │ Waiting  │◀──────▶│ Executing│
                             │ (prompt) │        │ (command)│
                             └────┬─────┘        └──────────┘
                                  │
                                  │ exit
                                  ▼
                             ┌──────────┐
                             │  Exited  │
                             └──────────┘
```

### Command Lifecycle

```text
Input → Lex → Parse → Resolve → Validate → Execute → Output
                        │
                        └─▶ Abbreviation expansion
```

---

## Relationships Summary

| From | To | Relationship | Cardinality |
|------|-----|--------------|-------------|
| Shell | Config | has | 1:1 |
| Shell | Environment | has | 1:1 |
| Shell | History | has | 1:1 |
| Shell | Executor | has | 1:1 |
| Shell | LineEditor | has | 1:1 |
| Executor | Builtin | uses | 1:N |
| History | HistoryEntry | contains | 1:N |
| Completer | CompletionCandidate | produces | 1:N |
| Lexer | Token | produces | 1:N |
| Parser | Command | produces | 1:1 |
