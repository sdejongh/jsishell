// Package builtins provides built-in shell commands and their registry.
package builtins

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/parser"
	"github.com/sdejongh/jsishell/internal/terminal"
)

// Context provides execution context to builtin commands.
type Context struct {
	Stdin   io.Reader             // Standard input
	Stdout  io.Writer             // Standard output
	Stderr  io.Writer             // Standard error
	Env     *env.Environment      // Environment variables
	WorkDir string                // Current working directory
	Colors  *terminal.ColorScheme // Color scheme for output
}

// WriteError writes an error message to stderr with red color if colors are enabled.
func (c *Context) WriteError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if c.Colors != nil {
		msg = c.Colors.Error(msg)
	}
	fmt.Fprint(c.Stderr, msg)
}

// WriteErrorln writes an error message with newline to stderr with red color if colors are enabled.
func (c *Context) WriteErrorln(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if c.Colors != nil {
		msg = c.Colors.Error(msg)
	}
	fmt.Fprintln(c.Stderr, msg)
}

// Handler is the function signature for builtin command handlers.
// It returns an exit code (0 for success) and an optional error.
type Handler func(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error)

// OptionDef defines a command option.
type OptionDef struct {
	Long        string // Long form (e.g., "--verbose")
	Short       string // Short form (e.g., "-v")
	Description string // Help text
	HasValue    bool   // Whether the option takes a value
	Default     string // Default value if any
}

// Definition defines a builtin command.
type Definition struct {
	Name        string      // Command name
	Description string      // Short description
	Usage       string      // Usage pattern
	Handler     Handler     // Command handler function
	Options     []OptionDef // Supported options
}

// Registry manages builtin commands.
type Registry struct {
	mu       sync.RWMutex
	builtins map[string]Definition
}

// NewRegistry creates a new empty builtin registry.
func NewRegistry() *Registry {
	return &Registry{
		builtins: make(map[string]Definition),
	}
}

// Register adds a builtin command to the registry.
func (r *Registry) Register(def Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builtins[def.Name] = def
}

// Get returns a builtin by name.
// Returns the definition and true if found, zero value and false otherwise.
func (r *Registry) Get(name string) (Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.builtins[name]
	return def, ok
}

// List returns all builtin names in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.builtins))
	for name := range r.builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Match finds builtins whose names start with the given prefix.
// Returns matching names in sorted order.
func (r *Registry) Match(prefix string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []string
	for name := range r.builtins {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}
	sort.Strings(matches)
	return matches
}

// All returns all registered builtins.
func (r *Registry) All() map[string]Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Definition, len(r.builtins))
	for k, v := range r.builtins {
		result[k] = v
	}
	return result
}

// Count returns the number of registered builtins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.builtins)
}

// Has returns true if a builtin with the given name exists.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.builtins[name]
	return ok
}
