// Package executor executes parsed commands.
package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/errors"
	"github.com/sdejongh/jsishell/internal/parser"
	"github.com/sdejongh/jsishell/internal/terminal"
)

// Executor executes parsed commands.
type Executor struct {
	registry            *builtins.Registry
	env                 *env.Environment
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	workDir             string
	abbreviationsEnable bool
	colors              *terminal.ColorScheme
}

// Option is a functional option for configuring the Executor.
type Option func(*Executor)

// WithRegistry sets the builtin registry.
func WithRegistry(r *builtins.Registry) Option {
	return func(e *Executor) {
		e.registry = r
	}
}

// WithEnv sets the environment.
func WithEnv(env *env.Environment) Option {
	return func(e *Executor) {
		e.env = env
	}
}

// WithStdin sets the standard input.
func WithStdin(r io.Reader) Option {
	return func(e *Executor) {
		e.stdin = r
	}
}

// WithStdout sets the standard output.
func WithStdout(w io.Writer) Option {
	return func(e *Executor) {
		e.stdout = w
	}
}

// WithStderr sets the standard error.
func WithStderr(w io.Writer) Option {
	return func(e *Executor) {
		e.stderr = w
	}
}

// WithWorkDir sets the working directory.
func WithWorkDir(dir string) Option {
	return func(e *Executor) {
		e.workDir = dir
	}
}

// WithAbbreviations enables or disables command abbreviations.
func WithAbbreviations(enable bool) Option {
	return func(e *Executor) {
		e.abbreviationsEnable = enable
	}
}

// WithColors sets the color scheme.
func WithColors(colors *terminal.ColorScheme) Option {
	return func(e *Executor) {
		e.colors = colors
	}
}

// New creates a new Executor with the given options.
func New(opts ...Option) *Executor {
	e := &Executor{
		registry:            builtins.NewRegistry(),
		env:                 env.New(),
		stdin:               os.Stdin,
		stdout:              os.Stdout,
		stderr:              os.Stderr,
		abbreviationsEnable: true,
		colors:              terminal.NewColorScheme(nil), // Default colors
	}

	// Get current working directory
	if wd, err := os.Getwd(); err == nil {
		e.workDir = wd
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Registry returns the builtin registry.
func (e *Executor) Registry() *builtins.Registry {
	return e.registry
}

// Env returns the environment.
func (e *Executor) Env() *env.Environment {
	return e.env
}

// Execute executes a parsed command.
// Returns the exit code and any error that occurred.
func (e *Executor) Execute(ctx context.Context, cmd *parser.Command) (int, error) {
	if cmd == nil {
		return 0, nil // Empty command
	}

	// Handle Windows drive letters (e.g., "c:", "D:") as "cd <drive>:"
	if isWindowsDriveLetter(cmd.Name) {
		// Transform into a cd command
		cdCmd := parser.NewCommand()
		cdCmd.Name = "cd"
		cdCmd.Resolved = "cd"
		cdCmd.Args = []string{cmd.Name}
		cdCmd.ArgsWithInfo = []parser.Arg{{Value: cmd.Name, Quoted: false}}
		cdCmd.RawInput = cmd.RawInput
		cmd = cdCmd
	}

	// Resolve command name (handle abbreviations)
	resolved, alternatives, err := e.ResolveCommand(cmd.Name)
	if err != nil {
		if err == errors.ErrAmbiguousCommand {
			return 1, fmt.Errorf("%w: %s (did you mean: %v?)", err, cmd.Name, alternatives)
		}
		return 127, err // 127 is standard for "command not found"
	}

	cmd.Resolved = resolved

	// Try builtin first
	if def, ok := e.registry.Get(resolved); ok {
		return e.executeBuiltin(ctx, cmd, def)
	}

	// Try external command
	return e.executeExternal(ctx, cmd)
}

// ExecuteInput parses and executes a command string.
func (e *Executor) ExecuteInput(ctx context.Context, input string) (int, error) {
	cmd, err := parser.ParseInputWithEnv(input, e.env)
	if err != nil {
		return 1, err
	}
	return e.Execute(ctx, cmd)
}

// ResolveCommand resolves a command name, handling abbreviations.
// Returns the resolved name, any alternatives (for ambiguous commands), and error.
func (e *Executor) ResolveCommand(name string) (string, []string, error) {
	// Check for exact match first
	if e.registry.Has(name) {
		return name, nil, nil
	}

	// If abbreviations disabled, only check for external command
	if !e.abbreviationsEnable {
		// Check if it's an external command
		if _, err := exec.LookPath(name); err == nil {
			return name, nil, nil
		}
		return "", nil, fmt.Errorf("%w: %s", errors.ErrCommandNotFound, name)
	}

	// Try to match as prefix
	matches := e.registry.Match(name)

	switch len(matches) {
	case 0:
		// No builtin matches, check external command
		if _, err := exec.LookPath(name); err == nil {
			return name, nil, nil
		}
		return "", nil, fmt.Errorf("%w: %s", errors.ErrCommandNotFound, name)

	case 1:
		// Unique match
		return matches[0], nil, nil

	default:
		// Ambiguous
		return "", matches, errors.ErrAmbiguousCommand
	}
}

// executeBuiltin executes a builtin command.
func (e *Executor) executeBuiltin(ctx context.Context, cmd *parser.Command, def builtins.Definition) (int, error) {
	execCtx := &builtins.Context{
		Stdin:   e.stdin,
		Stdout:  e.stdout,
		Stderr:  e.stderr,
		Env:     e.env,
		WorkDir: e.workDir,
		Colors:  e.colors,
	}

	code, err := def.Handler(ctx, cmd, execCtx)

	// Sync workDir from context (in case builtin changed it, e.g., goto/cd)
	if execCtx.WorkDir != e.workDir {
		e.workDir = execCtx.WorkDir
	}

	return code, err
}

// Colors returns the color scheme.
func (e *Executor) Colors() *terminal.ColorScheme {
	return e.colors
}

// SetColors sets the color scheme.
func (e *Executor) SetColors(colors *terminal.ColorScheme) {
	e.colors = colors
}

// executeExternal executes an external command.
func (e *Executor) executeExternal(ctx context.Context, cmd *parser.Command) (int, error) {
	// Look up the command
	path, err := exec.LookPath(cmd.Resolved)
	if err != nil {
		return 127, fmt.Errorf("%w: %s", errors.ErrCommandNotFound, cmd.Name)
	}

	// Build arguments: include flags and options for external commands
	args := cmd.AllArgs()

	// Create the command
	extCmd := exec.CommandContext(ctx, path, args...)
	extCmd.Stdin = e.stdin
	extCmd.Stdout = e.stdout
	extCmd.Stderr = e.stderr
	extCmd.Env = e.env.ToSlice()
	extCmd.Dir = e.workDir

	// Run the command
	err = extCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}

// SetWorkDir sets the working directory.
func (e *Executor) SetWorkDir(dir string) error {
	e.workDir = dir
	e.env.Set("PWD", dir)
	return os.Chdir(dir)
}

// WorkDir returns the current working directory.
func (e *Executor) WorkDir() string {
	return e.workDir
}

// SetAbbreviations enables or disables command abbreviations.
func (e *Executor) SetAbbreviations(enable bool) {
	e.abbreviationsEnable = enable
}

// AbbreviationsEnabled returns whether abbreviations are enabled.
func (e *Executor) AbbreviationsEnabled() bool {
	return e.abbreviationsEnable
}
