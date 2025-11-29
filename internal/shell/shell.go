// Package shell provides the main shell REPL and orchestration.
package shell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/completion"
	"github.com/sdejongh/jsishell/internal/config"
	"github.com/sdejongh/jsishell/internal/env"
	shellerrors "github.com/sdejongh/jsishell/internal/errors"
	"github.com/sdejongh/jsishell/internal/executor"
	"github.com/sdejongh/jsishell/internal/history"
	"github.com/sdejongh/jsishell/internal/terminal"
)

// Shell represents the main shell instance.
type Shell struct {
	executor       *executor.Executor
	terminal       *terminal.Terminal
	lineEditor     *terminal.LineEditor
	env            *env.Environment
	config         *config.Config
	history        *history.History
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	promptExpander *terminal.PromptExpander

	promptFormat string // Prompt format string (with %d, %u, etc.)
	running      bool
	exitCode     int
	interactive  bool // true if using LineEditor

	// Signal handling
	sigChan chan os.Signal
	ctx     context.Context
	cancel  context.CancelFunc
}

// Option is a functional option for configuring the Shell.
type Option func(*Shell)

// WithPrompt sets the shell prompt format string.
// Supports variables: %d (cwd), %D (cwd basename), %~ (cwd with ~),
// %u (username), %h (hostname), %t (time), %T (time with seconds), %n (newline), %% (literal %)
func WithPrompt(prompt string) Option {
	return func(s *Shell) {
		s.promptFormat = prompt
	}
}

// WithStdin sets the standard input.
func WithStdin(r io.Reader) Option {
	return func(s *Shell) {
		s.stdin = r
	}
}

// WithStdout sets the standard output.
func WithStdout(w io.Writer) Option {
	return func(s *Shell) {
		s.stdout = w
	}
}

// WithStderr sets the standard error.
func WithStderr(w io.Writer) Option {
	return func(s *Shell) {
		s.stderr = w
	}
}

// WithExecutor sets the executor.
func WithExecutor(e *executor.Executor) Option {
	return func(s *Shell) {
		s.executor = e
	}
}

// WithEnv sets the environment.
func WithEnv(e *env.Environment) Option {
	return func(s *Shell) {
		s.env = e
	}
}

// WithConfig sets the configuration.
func WithConfig(c *config.Config) Option {
	return func(s *Shell) {
		s.config = c
	}
}

// New creates a new Shell with the given options.
func New(opts ...Option) *Shell {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Shell{
		stdin:          os.Stdin,
		stdout:         os.Stdout,
		stderr:         os.Stderr,
		promptFormat:   "", // Will be set by config or options
		promptExpander: terminal.NewPromptExpander(),
		running:        false,
		exitCode:       0,
		sigChan:        make(chan os.Signal, 1),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		// Log error but continue with defaults
		fmt.Fprintf(s.stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.Default()
	}
	s.config = cfg

	// Setup color scheme for prompt expander
	colorScheme := terminal.NewColorScheme(&cfg.Colors)
	s.promptExpander.SetColorScheme(colorScheme)

	// Apply configuration (sets prompt from config)
	s.applyConfig()

	// Apply options AFTER config - options override config
	for _, opt := range opts {
		opt(s)
	}

	// If config was explicitly provided via option, apply it again
	// (WithConfig sets s.config but we need to apply it)
	// But since options run after applyConfig, any WithPrompt overrides config

	// Initialize environment if not provided
	if s.env == nil {
		s.env = env.New()
	}

	// Initialize terminal
	s.terminal = terminal.New()

	// Initialize line editor if terminal is interactive
	if s.terminal.IsTerminal() {
		s.lineEditor = terminal.NewLineEditor(s.terminal)
		s.lineEditor.SetPrompt(s.expandedPrompt())
		s.interactive = true

		// Setup color scheme for ghost text
		if s.config != nil {
			s.lineEditor.SetColors(terminal.NewColorScheme(&s.config.Colors))
		}
	}

	// Initialize executor if not provided
	if s.executor == nil {
		reg := builtins.NewRegistry()
		builtins.RegisterAll(reg)

		// Create color scheme from config
		var colorScheme *terminal.ColorScheme
		if s.config != nil {
			colorScheme = terminal.NewColorScheme(&s.config.Colors)
		} else {
			colorScheme = terminal.NewColorScheme(nil)
		}

		// Get abbreviations setting from config
		abbreviationsEnabled := true
		if s.config != nil {
			abbreviationsEnabled = s.config.Abbreviations.Enabled
		}

		s.executor = executor.New(
			executor.WithRegistry(reg),
			executor.WithEnv(s.env),
			executor.WithStdin(s.stdin),
			executor.WithStdout(s.stdout),
			executor.WithStderr(s.stderr),
			executor.WithColors(colorScheme),
			executor.WithAbbreviations(abbreviationsEnabled),
		)
	}

	// Setup reload callback
	builtins.SetReloadCallback(s.onConfigReload)

	// Setup history provider callback (will be set after history initialization)
	// Deferred to initHistory

	// Setup completion if interactive
	if s.interactive && s.lineEditor != nil {
		completer := s.createCompleter()
		s.lineEditor.SetCompleter(completer)
	}

	// Initialize history
	s.initHistory()

	return s
}

// Run starts the shell REPL loop. Blocks until exit.
func (s *Shell) Run() error {
	s.running = true

	// Setup signal handling
	s.setupSignals()
	defer s.cleanupSignals()

	// Use interactive mode if available
	if s.interactive {
		return s.runInteractive()
	}

	return s.runNonInteractive()
}

// runInteractive runs the shell with line editing support.
func (s *Shell) runInteractive() error {
	// Ensure history is saved on exit
	defer s.saveHistory()

	for s.running {
		// Update prompt before each read (to reflect cwd changes, time, etc.)
		s.lineEditor.SetPrompt(s.expandedPrompt())

		// Read line with editor
		line, err := s.lineEditor.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(s.stdout)
				break
			}
			return err
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Add to history before execution
		if s.history != nil {
			s.history.Add(line)
		}

		// Execute command
		exitCode, err := s.Execute(line)
		s.exitCode = exitCode

		// Check for exit command
		var exitErr builtins.ExitCode
		if errors.As(err, &exitErr) {
			s.exitCode = exitErr.Code
			break
		}

		// Display error if any (but not for exit)
		if err != nil {
			fmt.Fprintf(s.stderr, "error: %v\n", err)
		}
	}

	return nil
}

// runNonInteractive runs the shell without line editing (pipe/script mode).
func (s *Shell) runNonInteractive() error {
	// Create a line reader
	reader := bufio.NewReader(s.stdin)

	for s.running {
		// Display prompt
		fmt.Fprint(s.stdout, s.expandedPrompt())

		// Read line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// End of input (e.g., Ctrl+D)
				fmt.Fprintln(s.stdout)
				break
			}
			// Other error
			if errors.Is(err, shellerrors.ErrInterrupted) {
				fmt.Fprintln(s.stdout)
				continue
			}
			return err
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Execute command
		exitCode, err := s.Execute(line)
		s.exitCode = exitCode

		// Check for exit command
		var exitErr builtins.ExitCode
		if errors.As(err, &exitErr) {
			s.exitCode = exitErr.Code
			break
		}

		// Display error if any (but not for exit)
		if err != nil {
			fmt.Fprintf(s.stderr, "error: %v\n", err)
		}
	}

	return nil
}

// Execute runs a single command string and returns exit code.
func (s *Shell) Execute(input string) (int, error) {
	return s.executor.ExecuteInput(s.ctx, input)
}

// Exit terminates the shell with the given exit code.
func (s *Shell) Exit(code int) {
	s.exitCode = code
	s.running = false
	s.cancel()
}

// ExitCode returns the exit code of the last command.
func (s *Shell) ExitCode() int {
	return s.exitCode
}

// IsRunning returns true if the shell is running.
func (s *Shell) IsRunning() bool {
	return s.running
}

// Executor returns the shell's executor.
func (s *Shell) Executor() *executor.Executor {
	return s.executor
}

// Env returns the shell's environment.
func (s *Shell) Env() *env.Environment {
	return s.env
}

// SetPrompt sets the shell prompt format string.
func (s *Shell) SetPrompt(prompt string) {
	s.promptFormat = prompt
	if s.lineEditor != nil {
		s.lineEditor.SetPrompt(s.expandedPrompt())
	}
}

// setupSignals sets up signal handling for the shell.
func (s *Shell) setupSignals() {
	setupPlatformSignals(s.sigChan)

	go func() {
		for sig := range s.sigChan {
			shouldContinue := s.handlePlatformSignal(sig)
			if shouldContinue {
				// Print newline and prompt after interrupt
				fmt.Fprintln(s.stdout)
				fmt.Fprint(s.stdout, s.expandedPrompt())
			}
		}
	}()
}

// cleanupSignals stops signal handling.
func (s *Shell) cleanupSignals() {
	signal.Stop(s.sigChan)
	close(s.sigChan)
}

// applyConfig applies the current configuration to the shell.
func (s *Shell) applyConfig() {
	if s.config == nil {
		return
	}

	// Apply prompt format
	if s.config.Prompt != "" {
		s.promptFormat = s.config.Prompt
	}
}

// onConfigReload is called when the configuration is reloaded.
func (s *Shell) onConfigReload(cfg *config.Config) {
	s.config = cfg
	s.applyConfig()

	// Update color scheme
	var colorScheme *terminal.ColorScheme
	if cfg != nil {
		colorScheme = terminal.NewColorScheme(&cfg.Colors)
	} else {
		colorScheme = terminal.NewColorScheme(nil)
	}

	// Update prompt expander color scheme
	if s.promptExpander != nil {
		s.promptExpander.SetColorScheme(colorScheme)
	}

	// Update line editor prompt if interactive
	if s.lineEditor != nil {
		s.lineEditor.SetPrompt(s.expandedPrompt())
	}

	// Update executor settings
	if s.executor != nil {
		s.executor.SetColors(colorScheme)

		// Update abbreviations setting
		if cfg != nil {
			s.executor.SetAbbreviations(cfg.Abbreviations.Enabled)
		}
	}
}

// expandedPrompt returns the prompt with all variables expanded.
func (s *Shell) expandedPrompt() string {
	if s.promptExpander == nil {
		return s.promptFormat
	}
	// Update the working directory in the expander
	if s.executor != nil {
		s.promptExpander.SetWorkDir(s.executor.WorkDir())
	}
	return s.promptExpander.Expand(s.promptFormat)
}

// Config returns the shell's configuration.
func (s *Shell) Config() *config.Config {
	return s.config
}

// initHistory initializes the command history.
func (s *Shell) initHistory() {
	// Get history settings from config
	maxSize := config.DefaultHistorySize
	histFile := ""
	ignoreDuplicates := true
	ignoreSpacePrefix := true

	if s.config != nil {
		if s.config.History.MaxSize > 0 {
			maxSize = s.config.History.MaxSize
		}
		histFile = s.config.History.File
		ignoreDuplicates = s.config.History.IgnoreDuplicates
		ignoreSpacePrefix = s.config.History.IgnoreSpacePrefix
	}

	// Create history
	s.history = history.New(maxSize)
	s.history.SetIgnoreDuplicates(ignoreDuplicates)
	s.history.SetIgnoreSpacePrefix(ignoreSpacePrefix)

	// Load history from file
	if histFile != "" {
		histFile = config.ExpandPath(histFile)
		if err := s.history.Load(histFile); err != nil {
			fmt.Fprintf(s.stderr, "Warning: failed to load history: %v\n", err)
		}
	}

	// Connect history to line editor
	if s.lineEditor != nil {
		s.lineEditor.SetHistory(s.history)
	}

	// Setup history provider for the history builtin command
	builtins.SetHistoryProvider(func() builtins.HistoryProvider {
		return &historyAdapter{h: s.history}
	})
}

// historyAdapter adapts history.History to builtins.HistoryProvider.
type historyAdapter struct {
	h *history.History
}

func (a *historyAdapter) Len() int {
	if a.h == nil {
		return 0
	}
	return a.h.Len()
}

func (a *historyAdapter) All() []builtins.HistoryEntry {
	if a.h == nil {
		return nil
	}
	entries := a.h.All()
	result := make([]builtins.HistoryEntry, len(entries))
	for i, e := range entries {
		result[i] = builtins.HistoryEntry{
			Command:   e.Command,
			Timestamp: e.Timestamp,
		}
	}
	return result
}

func (a *historyAdapter) Clear() {
	if a.h != nil {
		a.h.Clear()
	}
}

// saveHistory saves the command history to file.
func (s *Shell) saveHistory() {
	if s.history == nil || s.config == nil {
		return
	}

	histFile := s.config.History.File
	if histFile == "" {
		return
	}

	histFile = config.ExpandPath(histFile)
	if err := s.history.Save(histFile); err != nil {
		fmt.Fprintf(s.stderr, "Warning: failed to save history: %v\n", err)
	}
}

// createCompleter creates a completer with command definitions including options.
func (s *Shell) createCompleter() *completion.Completer {
	allDefs := s.executor.Registry().All()

	// Convert builtin definitions to completion definitions
	var defs []completion.CommandDef
	for _, def := range allDefs {
		cmdDef := completion.CommandDef{
			Name:        def.Name,
			Description: def.Description,
			Options:     make([]completion.OptionDef, 0, len(def.Options)),
		}

		for _, opt := range def.Options {
			cmdDef.Options = append(cmdDef.Options, completion.OptionDef{
				Long:        opt.Long,
				Short:       opt.Short,
				Description: opt.Description,
			})
		}

		defs = append(defs, cmdDef)
	}

	return completion.NewCompleterWithDefs(defs)
}
