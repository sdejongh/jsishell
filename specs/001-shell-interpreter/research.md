# Research: Shell Interpreter

**Feature**: 001-shell-interpreter
**Date**: 2025-01-25

## Overview

This document consolidates research findings for implementing a cross-platform interactive shell in Go with no external runtime dependencies.

---

## 1. Terminal Handling in Go

### Decision: Use `golang.org/x/term`

**Rationale**: Official Go sub-repository providing cross-platform terminal operations including raw mode, terminal size detection, and password input. Statically links with the binary.

**Alternatives Considered**:
- `github.com/mattn/go-isatty` - Only detects TTY, doesn't provide raw mode
- `github.com/containerd/console` - Container-focused, unnecessary complexity
- Manual syscalls - Platform-specific, maintenance burden

**Implementation Notes**:
- Use `term.MakeRaw()` for raw mode (key-by-key input)
- Use `term.GetSize()` for terminal dimensions
- Save/restore terminal state on entry/exit
- Handle SIGWINCH for resize events

```go
import "golang.org/x/term"

oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
defer term.Restore(int(os.Stdin.Fd()), oldState)
```

---

## 2. YAML Configuration Parsing

### Decision: Use `gopkg.in/yaml.v3`

**Rationale**: De-facto standard Go YAML library. Stable, well-maintained, supports YAML 1.2. No runtime dependencies - compiles into binary.

**Alternatives Considered**:
- `github.com/goccy/go-yaml` - Faster but less mature
- TOML (`github.com/BurntSushi/toml`) - Different format, user requested YAML
- JSON (standard library) - Less human-readable for config

**Implementation Notes**:
- Use struct tags for mapping: `yaml:"key"`
- Support `~` expansion for paths manually
- Validate config after parsing

```go
type Config struct {
    Prompt      string `yaml:"prompt"`
    HistorySize int    `yaml:"history_size"`
    Colors      bool   `yaml:"colors"`
}
```

---

## 3. Cross-Platform File Operations

### Decision: Use `os` and `path/filepath` packages

**Rationale**: Standard library provides cross-platform file operations. `filepath` handles path separators correctly on all platforms.

**Key Considerations**:
- Use `filepath.Join()` not string concatenation
- Use `os.UserHomeDir()` for home directory
- Use `os.UserConfigDir()` for config directory (follows XDG on Linux)
- Handle Windows drive letters in path parsing

**Implementation Notes**:
```go
// Cross-platform home directory
home, _ := os.UserHomeDir()

// Cross-platform config directory
configDir, _ := os.UserConfigDir() // ~/.config on Linux, AppData on Windows

// Always use filepath.Join
configPath := filepath.Join(configDir, "jsishell", "config.yaml")
```

---

## 4. Signal Handling Across Platforms

### Decision: Use `os/signal` with platform-specific handling

**Rationale**: Standard library provides signal handling. Windows lacks POSIX signals but supports Ctrl+C via SIGINT.

**Cross-Platform Signals**:
| Signal | Linux/macOS | Windows | Action |
|--------|-------------|---------|--------|
| SIGINT | Yes | Yes (Ctrl+C) | Interrupt current command |
| SIGTERM | Yes | No | Graceful shutdown |
| SIGWINCH | Yes | No | Terminal resize |
| SIGTSTP | Yes | No | Suspend (not in MVP) |

**Implementation Notes**:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt) // SIGINT on all platforms

// Platform-specific signals via build tags
// +build !windows
signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGTERM)
```

---

## 5. Process Execution

### Decision: Use `os/exec` package

**Rationale**: Standard library provides cross-platform process execution with proper argument handling and environment management.

**Key Features**:
- `exec.Command()` for creating commands
- `cmd.Stdin/Stdout/Stderr` for stream redirection
- `cmd.Env` for environment variables
- `cmd.Run()` blocks, `cmd.Start()` + `cmd.Wait()` for control

**Implementation Notes**:
```go
cmd := exec.Command(name, args...)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Env = os.Environ()
err := cmd.Run()
```

---

## 6. Line Editing Implementation

### Decision: Custom implementation using raw terminal input

**Rationale**: No suitable Go library provides the exact feature set needed (inline autocompletion with ghost text). Custom implementation allows full control.

**Key Components**:
1. **Input Buffer**: Stores current line with cursor position
2. **Key Reader**: Reads raw input, handles escape sequences
3. **Renderer**: Redraws line with cursor, ghost text

**Escape Sequences** (ANSI):
| Sequence | Meaning |
|----------|---------|
| `\x1b[D` | Cursor left |
| `\x1b[C` | Cursor right |
| `\x1b[H` | Cursor home |
| `\x1b[K` | Clear to end of line |
| `\x1b[2m` | Dim text (ghost) |
| `\x1b[0m` | Reset formatting |

**Implementation Notes**:
```go
type LineEditor struct {
    buffer    []rune
    cursor    int
    prompt    string
    ghostText string
}
```

---

## 7. History Persistence

### Decision: Simple line-based text file

**Rationale**: Human-readable, easy to edit manually, compatible with other shells' history format.

**File Format**:
```text
list /home
cd /tmp
copy file.txt backup/
```

**Implementation Notes**:
- One command per line
- Append on each command execution
- Truncate from beginning when limit reached
- Load on startup, keep in memory
- Handle concurrent access (file locking on write)

---

## 8. Autocompletion Strategy

### Decision: Prefix-based completion with multiple sources

**Rationale**: Simple and fast. Sources include built-in commands, file paths, and command history.

**Completion Sources**:
1. **Built-in Commands**: Exact prefix match against registry
2. **File Paths**: Use `filepath.Glob()` with user input as prefix
3. **History**: Prefix match against recent commands

**Inline Ghost Text**:
- Show only when single match exists
- Display in dim/gray color
- Accept with Tab, dismiss with any other key

**Implementation Notes**:
```go
type Completer struct {
    builtins []string
    history  *History
}

func (c *Completer) Complete(input string) []string {
    // Return all matches for given prefix
}

func (c *Completer) InlineSuggestion(input string) string {
    // Return single best match or empty
}
```

---

## 9. Command Abbreviation Algorithm

### Decision: Prefix trie for O(m) lookup

**Rationale**: Efficient prefix matching for command abbreviations. Trie allows quick check for unique vs ambiguous prefixes.

**Algorithm**:
1. Build trie from built-in command names
2. On input, traverse trie
3. If single leaf reachable → execute that command
4. If multiple leaves → "Ambiguous: did you mean X, Y?"
5. If no leaves → "Command not found"

**Implementation Notes**:
```go
type TrieNode struct {
    children map[rune]*TrieNode
    command  string // Non-empty if this node terminates a command
}

func (t *Trie) Lookup(prefix string) (command string, alternatives []string)
```

---

## 10. Color Support Detection

### Decision: Check `TERM` environment variable + TTY detection

**Rationale**: Standard approach used by most CLI tools.

**Detection Logic**:
1. Check if stdout is a TTY (`term.IsTerminal()`)
2. Check `NO_COLOR` environment variable (respect user preference)
3. Check `TERM` for "dumb" (no colors)
4. Check `COLORTERM` for "truecolor" support

**Implementation Notes**:
```go
func SupportsColor() bool {
    if os.Getenv("NO_COLOR") != "" {
        return false
    }
    if !term.IsTerminal(int(os.Stdout.Fd())) {
        return false
    }
    if os.Getenv("TERM") == "dumb" {
        return false
    }
    return true
}
```

---

## 11. Error Handling Strategy

### Decision: Sentinel errors + wrapped context

**Rationale**: Go idiom. Sentinel errors for known conditions, wrapped errors for context.

**Sentinel Errors**:
```go
var (
    ErrCommandNotFound  = errors.New("command not found")
    ErrAmbiguousCommand = errors.New("ambiguous command")
    ErrInvalidSyntax    = errors.New("invalid syntax")
    ErrPermissionDenied = errors.New("permission denied")
    ErrFileNotFound     = errors.New("file not found")
)
```

**Error Wrapping**:
```go
return fmt.Errorf("copy %s: %w", src, ErrPermissionDenied)
```

---

## 12. Build Configuration

### Decision: Standard Go build with ldflags for size optimization

**Rationale**: Produces single static binary, no external dependencies at runtime.

**Build Commands**:
```bash
# Development build
go build -o jsishell ./cmd/jsishell

# Release build (stripped, smaller)
go build -ldflags="-s -w" -o jsishell ./cmd/jsishell

# Cross-compile
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o jsishell.exe ./cmd/jsishell
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o jsishell-darwin ./cmd/jsishell
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o jsishell-linux ./cmd/jsishell
```

**Expected Binary Sizes**:
- Development: ~8-10MB
- Release (stripped): ~5-7MB

---

## Summary

All technical unknowns have been resolved. The implementation will use:

| Component | Solution |
|-----------|----------|
| Terminal | `golang.org/x/term` |
| YAML | `gopkg.in/yaml.v3` |
| File Ops | Standard `os` + `path/filepath` |
| Signals | Standard `os/signal` + build tags |
| Processes | Standard `os/exec` |
| Line Edit | Custom implementation |
| History | Line-based text file |
| Completion | Prefix-based with trie |
| Colors | TTY + env detection |
| Errors | Sentinel + wrapping |

**External Dependencies (compile-time only)**:
1. `golang.org/x/term` - Terminal handling
2. `gopkg.in/yaml.v3` - YAML parsing

Both produce statically-linked code with no runtime dependencies.
