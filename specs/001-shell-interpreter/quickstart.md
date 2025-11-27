# Quickstart: JSIShell Development

**Feature**: 001-shell-interpreter
**Date**: 2025-01-25

## Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for convenience)

## Setup

### 1. Clone and Initialize

```bash
# Clone the repository
git clone <repository-url>
cd jsishell

# Initialize Go module (if not already done)
go mod init github.com/yourusername/jsishell

# Add dependencies
go get golang.org/x/term
go get gopkg.in/yaml.v3

# Verify setup
go mod tidy
```

### 2. Project Structure

Create the directory structure:

```bash
mkdir -p cmd/jsishell
mkdir -p internal/{shell,lexer,parser,executor,builtins,completion,history,config,terminal,env,errors}
mkdir -p tests/{integration,fixtures}
```

### 3. Verify Build

```bash
# Create minimal main.go
cat > cmd/jsishell/main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("JSIShell - Coming Soon")
}
EOF

# Build
go build -o jsishell ./cmd/jsishell

# Run
./jsishell
```

## Development Workflow

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/lexer/...

# Verbose
go test -v ./internal/parser/...
```

### Building

```bash
# Development build
go build -o jsishell ./cmd/jsishell

# Release build (smaller binary)
go build -ldflags="-s -w" -o jsishell ./cmd/jsishell

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o jsishell.exe ./cmd/jsishell

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o jsishell-darwin ./cmd/jsishell
```

### Code Quality

```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Lint (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck ./...
```

## Quick Implementation Guide

### Step 1: Start with the Lexer

The lexer is the foundation. Start here:

```go
// internal/lexer/lexer.go
package lexer

type Lexer struct {
    input   string
    pos     int
    readPos int
    ch      byte
}

func New(input string) *Lexer {
    l := &Lexer{input: input}
    l.readChar()
    return l
}

func (l *Lexer) NextToken() Token {
    // Implementation
}
```

**Test first**:
```go
// internal/lexer/lexer_test.go
func TestLexer(t *testing.T) {
    tests := []struct {
        input    string
        expected []Token
    }{
        {"list", []Token{{Type: TokenWord, Value: "list"}}},
        {"cd /home", []Token{...}},
    }
    // ...
}
```

### Step 2: Build the Parser

Parser consumes tokens and produces Command AST:

```go
// internal/parser/parser.go
package parser

func Parse(tokens []lexer.Token) (*Command, error) {
    // Build Command from tokens
}
```

### Step 3: Implement Builtins

Each builtin is a function:

```go
// internal/builtins/cd.go
package builtins

func CdCommand(ctx *Context, cmd *Command) error {
    if len(cmd.Args) == 0 {
        return os.Chdir(os.Getenv("HOME"))
    }
    return os.Chdir(cmd.Args[0])
}
```

### Step 4: Wire Up the Executor

Executor connects parser to builtins:

```go
// internal/executor/executor.go
package executor

func (e *Executor) Execute(input string) (int, error) {
    tokens := lexer.New(input).Tokens()
    cmd, err := parser.Parse(tokens)
    if err != nil {
        return 1, err
    }
    return e.run(cmd)
}
```

### Step 5: Add Terminal I/O

Raw mode for key-by-key input:

```go
// internal/terminal/terminal.go
package terminal

import "golang.org/x/term"

func (t *Terminal) EnterRawMode() (func(), error) {
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return nil, err
    }
    return func() { term.Restore(int(os.Stdin.Fd()), oldState) }, nil
}
```

### Step 6: Build the REPL

Main shell loop:

```go
// internal/shell/shell.go
package shell

func (s *Shell) Run() error {
    for s.Running {
        line, err := s.editor.ReadLine(s.config.Prompt)
        if err != nil {
            continue
        }
        exitCode, _ := s.executor.Execute(line)
        s.ExitCode = exitCode
    }
    return nil
}
```

## Testing Strategy

### Unit Tests (80%+ coverage required)

Each package has `*_test.go` files:

```text
internal/
├── lexer/
│   ├── lexer.go
│   └── lexer_test.go      # Table-driven tests for tokenization
├── parser/
│   ├── parser.go
│   └── parser_test.go     # Test parsing edge cases
├── builtins/
│   ├── cd.go
│   └── builtins_test.go   # Test each builtin
```

### Integration Tests

Test full command execution:

```go
// tests/integration/shell_test.go
func TestShellExecution(t *testing.T) {
    shell := NewTestShell()

    exitCode, err := shell.Execute("echo hello")
    assert.NoError(t, err)
    assert.Equal(t, 0, exitCode)
}
```

### Test Fixtures

```yaml
# tests/fixtures/config.yaml
prompt: "test> "
history_size: 10
colors: false
```

## Configuration File

Default location: `~/.config/jsishell/config.yaml`

```yaml
# Shell prompt
prompt: "$ "

# History settings
history_file: ~/.jsishell_history
history_size: 1000
history_ignore_dups: true

# Display
colors: true
color_scheme:
  directory: blue
  file: white
  executable: green
  error: red
  warning: yellow
  ghost_text: gray

# Behavior
abbreviations: true
completion_limit: 100
```

## Key Files to Create First

1. `cmd/jsishell/main.go` - Entry point
2. `internal/errors/errors.go` - Sentinel errors
3. `internal/lexer/lexer.go` - Tokenizer
4. `internal/parser/parser.go` - Parser
5. `internal/builtins/registry.go` - Builtin registry
6. `internal/executor/executor.go` - Command executor
7. `internal/terminal/terminal.go` - Terminal I/O
8. `internal/shell/shell.go` - Main shell struct

## Common Commands During Development

```bash
# Run the shell
go run ./cmd/jsishell

# Run with race detector
go run -race ./cmd/jsishell

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark
go test -bench=. ./internal/lexer/

# Generate documentation
go doc ./internal/lexer
```

## Troubleshooting

### Terminal Not Responding

The shell uses raw mode. If it crashes:
```bash
# Reset terminal
reset
# or
stty sane
```

### Build Errors on Windows

Ensure CGO is disabled for pure Go build:
```bash
CGO_ENABLED=0 go build ./cmd/jsishell
```

### Tests Failing with TTY Errors

Some tests require a TTY. Mock the terminal:
```go
// Use a mock terminal in tests
type MockTerminal struct {
    input  *bytes.Buffer
    output *bytes.Buffer
}
```
