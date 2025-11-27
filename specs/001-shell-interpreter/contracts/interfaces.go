// Package contracts defines the internal API contracts for JSIShell components.
// This file serves as documentation for the interfaces between components.
// It is not compiled into the final binary.

package contracts

import (
	"context"
	"io"
)

// =============================================================================
// Shell Core Interface
// =============================================================================

// Shell is the main shell interface providing the REPL loop.
type Shell interface {
	// Run starts the shell REPL loop. Blocks until exit.
	Run() error

	// Execute runs a single command string and returns exit code.
	Execute(input string) (exitCode int, err error)

	// Exit terminates the shell with given exit code.
	Exit(code int)

	// Reload reloads configuration without restarting.
	Reload() error
}

// =============================================================================
// Lexer Interface
// =============================================================================

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Value   string // Raw value
	Literal string // Processed value (escapes applied)
	Line    int
	Column  int
}

// TokenType identifies token types.
type TokenType int

const (
	TokenWord TokenType = iota
	TokenString
	TokenOption
	TokenVariable
	TokenWhitespace
	TokenNewline
	TokenEOF
	TokenError
)

// Lexer tokenizes input strings.
type Lexer interface {
	// NextToken returns the next token from input.
	NextToken() Token

	// Tokens returns all tokens from input.
	Tokens() []Token
}

// =============================================================================
// Parser Interface
// =============================================================================

// Command represents a parsed command.
type Command struct {
	Name     string            // Command name (possibly abbreviated)
	Resolved string            // Full resolved command name
	Args     []string          // Positional arguments
	Options  map[string]string // --key=value options
	Flags    map[string]bool   // Boolean flags
	Raw      string            // Original input
}

// Parser parses tokens into commands.
type Parser interface {
	// Parse parses tokens into a Command.
	Parse(tokens []Token) (*Command, error)
}

// =============================================================================
// Executor Interface
// =============================================================================

// ExecutionContext provides context for command execution.
type ExecutionContext struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Env     Environment
	WorkDir string
}

// Executor executes parsed commands.
type Executor interface {
	// Execute runs a command and returns exit code.
	Execute(ctx context.Context, cmd *Command, execCtx *ExecutionContext) (int, error)

	// ResolveCommand resolves abbreviated command names.
	ResolveCommand(name string) (resolved string, alternatives []string, err error)
}

// =============================================================================
// Builtin Interface
// =============================================================================

// BuiltinHandler is the function signature for builtin command handlers.
type BuiltinHandler func(ctx context.Context, cmd *Command, execCtx *ExecutionContext) (int, error)

// BuiltinDef defines a builtin command.
type BuiltinDef struct {
	Name        string
	Description string
	Usage       string
	Handler     BuiltinHandler
	Options     []OptionDef
}

// OptionDef defines a command option.
type OptionDef struct {
	Long        string
	Short       string
	Description string
	HasValue    bool
	Default     string
}

// BuiltinRegistry manages builtin commands.
type BuiltinRegistry interface {
	// Register adds a builtin command.
	Register(def BuiltinDef)

	// Get returns a builtin by name.
	Get(name string) (BuiltinDef, bool)

	// List returns all builtin names.
	List() []string

	// Match finds builtins matching a prefix.
	Match(prefix string) []string
}

// =============================================================================
// Completion Interface
// =============================================================================

// CompletionCandidate represents a completion suggestion.
type CompletionCandidate struct {
	Value       string
	Display     string
	Description string
	Type        CompletionType
}

// CompletionType identifies completion source.
type CompletionType int

const (
	CompletionBuiltin CompletionType = iota
	CompletionFile
	CompletionDirectory
	CompletionHistory
	CompletionOption
)

// Completer provides autocompletion suggestions.
type Completer interface {
	// Complete returns all matching completions for input.
	Complete(input string, cursorPos int) []CompletionCandidate

	// InlineSuggestion returns single best suggestion for ghost text.
	InlineSuggestion(input string) string
}

// =============================================================================
// History Interface
// =============================================================================

// HistoryEntry represents a command in history.
type HistoryEntry struct {
	Command   string
	Timestamp int64
	ExitCode  int
}

// History manages command history.
type History interface {
	// Add adds a command to history.
	Add(command string, exitCode int)

	// Search finds entries matching prefix.
	Search(prefix string) []HistoryEntry

	// Get returns entry at index (0 = most recent).
	Get(index int) (HistoryEntry, bool)

	// Len returns number of entries.
	Len() int

	// Load loads history from persistent storage.
	Load() error

	// Save persists history to storage.
	Save() error
}

// =============================================================================
// Environment Interface
// =============================================================================

// Environment manages shell environment variables.
type Environment interface {
	// Get returns variable value.
	Get(key string) string

	// Set sets variable value.
	Set(key, value string)

	// Unset removes a variable.
	Unset(key string)

	// Export marks variable for export to child processes.
	Export(key string)

	// IsExported returns true if variable is exported.
	IsExported(key string) bool

	// Expand expands $VAR references in string.
	Expand(input string) string

	// ToSlice returns KEY=VALUE slice for exec.
	ToSlice() []string

	// All returns all variables as map.
	All() map[string]string
}

// =============================================================================
// Config Interface
// =============================================================================

// Config holds shell configuration.
type Config interface {
	// Get returns config value by key path (e.g., "colors.error").
	Get(key string) interface{}

	// GetString returns string value or default.
	GetString(key string, defaultVal string) string

	// GetInt returns int value or default.
	GetInt(key string, defaultVal int) int

	// GetBool returns bool value or default.
	GetBool(key string, defaultVal bool) bool

	// Set sets config value.
	Set(key string, value interface{})

	// Reload reloads configuration from file.
	Reload() error

	// Path returns config file path.
	Path() string
}

// =============================================================================
// Terminal Interface
// =============================================================================

// Terminal handles low-level terminal I/O.
type Terminal interface {
	// EnterRawMode switches to raw mode, returns restore function.
	EnterRawMode() (restore func(), err error)

	// ReadKey reads a single key/escape sequence.
	ReadKey() (Key, error)

	// Write writes bytes to terminal.
	Write(p []byte) (n int, err error)

	// WriteString writes string to terminal.
	WriteString(s string) (n int, err error)

	// Size returns terminal dimensions.
	Size() (width, height int, err error)

	// IsTerminal returns true if stdout is a terminal.
	IsTerminal() bool

	// SupportsColor returns true if terminal supports colors.
	SupportsColor() bool

	// Clear clears the screen.
	Clear() error
}

// Key represents a keyboard input.
type Key struct {
	Rune    rune    // Character (0 for special keys)
	Special KeyType // Special key type
	Alt     bool    // Alt modifier
	Ctrl    bool    // Ctrl modifier
}

// KeyType identifies special keys.
type KeyType int

const (
	KeyNone KeyType = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyDelete
	KeyBackspace
	KeyTab
	KeyEnter
	KeyEscape
)

// =============================================================================
// LineEditor Interface
// =============================================================================

// LineEditor handles interactive line input.
type LineEditor interface {
	// ReadLine reads a line of input with editing support.
	ReadLine(prompt string) (string, error)

	// SetHistory sets the history for navigation.
	SetHistory(h History)

	// SetCompleter sets the completer for autocompletion.
	SetCompleter(c Completer)
}
