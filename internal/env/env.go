// Package env manages shell environment variables.
// It provides variable storage, expansion, and export functionality.
package env

import (
	"os"
	"regexp"
	"strings"
	"sync"
)

// Environment manages shell environment variables.
type Environment struct {
	mu      sync.RWMutex
	vars    map[string]string // Current shell variables
	exports map[string]bool   // Variables marked for export to child processes
}

// New creates a new Environment initialized with current OS environment.
func New() *Environment {
	env := &Environment{
		vars:    make(map[string]string),
		exports: make(map[string]bool),
	}

	// Import current OS environment
	for _, e := range os.Environ() {
		if idx := strings.Index(e, "="); idx != -1 {
			key := e[:idx]
			value := e[idx+1:]
			env.vars[key] = value
			env.exports[key] = true // OS vars are already exported
		}
	}

	return env
}

// Get returns the value of an environment variable.
// Returns empty string if the variable is not set.
func (e *Environment) Get(key string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vars[key]
}

// Set sets an environment variable.
func (e *Environment) Set(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vars[key] = value
}

// Unset removes an environment variable.
func (e *Environment) Unset(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.vars, key)
	delete(e.exports, key)
}

// Export marks a variable for export to child processes.
func (e *Environment) Export(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.vars[key]; exists {
		e.exports[key] = true
	}
}

// IsExported returns true if the variable is marked for export.
func (e *Environment) IsExported(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.exports[key]
}

// varPattern matches $VAR and ${VAR} patterns.
var varPattern = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}|\$([a-zA-Z_][a-zA-Z0-9_]*)`)

// Expand expands $VAR and ${VAR} references in a string.
// Unknown variables are replaced with empty string.
func (e *Environment) Expand(input string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name
		var name string
		if strings.HasPrefix(match, "${") {
			// ${VAR} form
			name = match[2 : len(match)-1]
		} else {
			// $VAR form
			name = match[1:]
		}

		if val, exists := e.vars[name]; exists {
			return val
		}
		return ""
	})
}

// ToSlice returns all exported variables as a KEY=VALUE slice.
// This is suitable for passing to exec.Cmd.Env.
func (e *Environment) ToSlice() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]string, 0, len(e.exports))
	for key := range e.exports {
		if val, exists := e.vars[key]; exists {
			result = append(result, key+"="+val)
		}
	}
	return result
}

// All returns a copy of all variables as a map.
func (e *Environment) All() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]string, len(e.vars))
	for k, v := range e.vars {
		result[k] = v
	}
	return result
}

// Exported returns a copy of all exported variable names.
func (e *Environment) Exported() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]string, 0, len(e.exports))
	for k := range e.exports {
		result = append(result, k)
	}
	return result
}

// Clone creates a deep copy of the environment.
func (e *Environment) Clone() *Environment {
	e.mu.RLock()
	defer e.mu.RUnlock()

	clone := &Environment{
		vars:    make(map[string]string, len(e.vars)),
		exports: make(map[string]bool, len(e.exports)),
	}

	for k, v := range e.vars {
		clone.vars[k] = v
	}
	for k, v := range e.exports {
		clone.exports[k] = v
	}

	return clone
}
