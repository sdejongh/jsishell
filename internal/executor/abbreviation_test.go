package executor

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/sdejongh/jsishell/internal/builtins"
	shellerrors "github.com/sdejongh/jsishell/internal/errors"
	"github.com/sdejongh/jsishell/internal/parser"
)

// TestAbbreviationResolution tests command abbreviation resolution scenarios.
func TestAbbreviationResolution(t *testing.T) {
	tests := []struct {
		name       string
		commands   []string // Commands to register
		input      string   // Abbreviation to resolve
		wantResult string   // Expected resolved command
		wantErr    error    // Expected error (nil if should succeed)
		wantAlts   int      // Expected number of alternatives (for ambiguous)
	}{
		{
			name:       "single letter unique match",
			commands:   []string{"list", "copy", "move"},
			input:      "l",
			wantResult: "list",
			wantErr:    nil,
		},
		{
			name:       "two letter unique match",
			commands:   []string{"list", "copy", "move"},
			input:      "li",
			wantResult: "list",
			wantErr:    nil,
		},
		{
			name:       "full command name",
			commands:   []string{"list", "copy", "move"},
			input:      "list",
			wantResult: "list",
			wantErr:    nil,
		},
		{
			name:       "prefix matches single command",
			commands:   []string{"echo", "exit", "env"},
			input:      "ech",
			wantResult: "echo",
			wantErr:    nil,
		},
		{
			name:     "ambiguous single letter",
			commands: []string{"copy", "cd", "clear"},
			input:    "c",
			wantErr:  shellerrors.ErrAmbiguousCommand,
			wantAlts: 3,
		},
		{
			name:     "ambiguous two letters",
			commands: []string{"copy", "count", "compare"},
			input:    "co",
			wantErr:  shellerrors.ErrAmbiguousCommand,
			wantAlts: 3,
		},
		{
			name:       "resolve ambiguity with more chars",
			commands:   []string{"copy", "count", "compare"},
			input:      "cop",
			wantResult: "copy",
			wantErr:    nil,
		},
		{
			name:     "no match returns command not found",
			commands: []string{"list", "copy"},
			input:    "x",
			wantErr:  shellerrors.ErrCommandNotFound,
		},
		{
			name:     "longer prefix no match",
			commands: []string{"list", "copy"},
			input:    "xyz",
			wantErr:  shellerrors.ErrCommandNotFound,
		},
		{
			name:       "case sensitive matching",
			commands:   []string{"List", "list"},
			input:      "L",
			wantResult: "List",
			wantErr:    nil,
		},
		{
			name:     "empty prefix matches nothing useful",
			commands: []string{"list", "copy"},
			input:    "",
			wantErr:  shellerrors.ErrAmbiguousCommand,
			wantAlts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := builtins.NewRegistry()
			for _, cmd := range tt.commands {
				reg.Register(builtins.Definition{Name: cmd})
			}

			e := New(WithRegistry(reg), WithAbbreviations(true))

			resolved, alts, err := e.ResolveCommand(tt.input)

			// Check error
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if tt.wantAlts > 0 && len(alts) != tt.wantAlts {
					t.Errorf("alternatives = %d, want %d", len(alts), tt.wantAlts)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resolved != tt.wantResult {
				t.Errorf("resolved = %q, want %q", resolved, tt.wantResult)
			}
		})
	}
}

// TestAmbiguousCommandHandling tests how ambiguous commands are handled.
func TestAmbiguousCommandHandling(t *testing.T) {
	tests := []struct {
		name         string
		commands     []string
		input        string
		wantAlts     []string // Expected alternatives in sorted order
		wantContains string   // String that error message should contain
	}{
		{
			name:         "c is ambiguous with cd, clear, copy",
			commands:     []string{"cd", "clear", "copy", "list"},
			input:        "c",
			wantAlts:     []string{"cd", "clear", "copy"},
			wantContains: "did you mean",
		},
		{
			name:         "e is ambiguous with echo, env, exit",
			commands:     []string{"echo", "env", "exit", "list"},
			input:        "e",
			wantAlts:     []string{"echo", "env", "exit"},
			wantContains: "did you mean",
		},
		{
			name:         "ma is ambiguous with make, makedir",
			commands:     []string{"list", "make", "makedir", "move"},
			input:        "ma",
			wantAlts:     []string{"make", "makedir"},
			wantContains: "did you mean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := builtins.NewRegistry()
			for _, cmd := range tt.commands {
				reg.Register(builtins.Definition{Name: cmd})
			}

			e := New(WithRegistry(reg))

			cmd := &parser.Command{Name: tt.input}
			_, err := e.Execute(context.Background(), cmd)

			if err == nil {
				t.Fatal("expected error for ambiguous command")
			}

			// Check error contains alternatives
			errMsg := err.Error()
			for _, alt := range tt.wantAlts {
				if !contains(errMsg, alt) {
					t.Errorf("error message %q should contain alternative %q", errMsg, alt)
				}
			}

			// Check error message format
			if !contains(errMsg, tt.wantContains) {
				t.Errorf("error message %q should contain %q", errMsg, tt.wantContains)
			}
		})
	}
}

// TestAbbreviationsDisabled tests behavior when abbreviations are disabled.
func TestAbbreviationsDisabled(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "list"})
	reg.Register(builtins.Definition{Name: "copy"})
	reg.Register(builtins.Definition{Name: "move"})

	e := New(WithRegistry(reg), WithAbbreviations(false))

	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantCmd string
	}{
		{
			name:    "exact match still works",
			input:   "list",
			wantOk:  true,
			wantCmd: "list",
		},
		{
			name:   "abbreviation l fails",
			input:  "l",
			wantOk: false,
		},
		{
			name:   "abbreviation li fails",
			input:  "li",
			wantOk: false,
		},
		{
			name:   "abbreviation cop fails",
			input:  "cop",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, _, err := e.ResolveCommand(tt.input)

			if tt.wantOk {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resolved != tt.wantCmd {
					t.Errorf("resolved = %q, want %q", resolved, tt.wantCmd)
				}
			} else {
				if !errors.Is(err, shellerrors.ErrCommandNotFound) {
					t.Errorf("error = %v, want ErrCommandNotFound", err)
				}
			}
		})
	}
}

// TestAbbreviationWithExternalCommands tests that external commands are checked.
func TestAbbreviationWithExternalCommands(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "list"})

	e := New(WithRegistry(reg))

	// "true" is an external command that exists on most systems
	resolved, _, err := e.ResolveCommand("true")
	if err != nil {
		t.Skipf("skipping: external command 'true' not found: %v", err)
	}
	if resolved != "true" {
		t.Errorf("resolved = %q, want %q", resolved, "true")
	}

	// Verify builtin takes precedence over external
	// "echo" exists both as builtin (registered) and external
	reg.Register(builtins.Definition{Name: "echo"})
	resolved, _, err = e.ResolveCommand("echo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resolved != "echo" {
		t.Errorf("resolved = %q, want %q", resolved, "echo")
	}
}

// TestAbbreviationExecution tests that abbreviated commands execute correctly.
func TestAbbreviationExecution(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name:        "list",
		Description: "List files",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			execCtx.Stdout.Write([]byte("list executed"))
			return 0, nil
		},
	})
	reg.Register(builtins.Definition{
		Name:        "copy",
		Description: "Copy files",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			execCtx.Stdout.Write([]byte("copy executed"))
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
		WithAbbreviations(true),
	)

	// Test abbreviated command execution
	tests := []struct {
		input      string
		wantOutput string
	}{
		{"l", "list executed"},
		{"li", "list executed"},
		{"lis", "list executed"},
		{"list", "list executed"},
		{"cop", "copy executed"},
		{"copy", "copy executed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			stdout.Reset()
			exitCode, err := e.ExecuteInput(context.Background(), tt.input)

			if err != nil {
				t.Errorf("ExecuteInput(%q) error: %v", tt.input, err)
			}
			if exitCode != 0 {
				t.Errorf("exitCode = %d, want 0", exitCode)
			}
			if stdout.String() != tt.wantOutput {
				t.Errorf("output = %q, want %q", stdout.String(), tt.wantOutput)
			}
		})
	}
}

// TestAbbreviationPreservesArgs tests that arguments are passed correctly to abbreviated commands.
func TestAbbreviationPreservesArgs(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "list",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			for i, arg := range cmd.Args {
				if i > 0 {
					execCtx.Stdout.Write([]byte(" "))
				}
				execCtx.Stdout.Write([]byte(arg))
			}
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
	)

	// Use abbreviation with arguments
	exitCode, err := e.ExecuteInput(context.Background(), "l /home /tmp")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if stdout.String() != "/home /tmp" {
		t.Errorf("output = %q, want %q", stdout.String(), "/home /tmp")
	}
}

// TestAbbreviationPreservesFlags tests that flags are passed correctly to abbreviated commands.
func TestAbbreviationPreservesFlags(t *testing.T) {
	var receivedFlags map[string]bool

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "list",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			receivedFlags = cmd.Flags
			return 0, nil
		},
	})

	e := New(WithRegistry(reg))

	_, err := e.ExecuteInput(context.Background(), "l -a --long")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(receivedFlags) != 2 {
		t.Errorf("received %d flags, want 2", len(receivedFlags))
	}
	if !receivedFlags["-a"] {
		t.Error("flag -a should be set")
	}
	if !receivedFlags["--long"] {
		t.Error("flag --long should be set")
	}
}

// TestResolvedFieldSet tests that the Resolved field is set on the command.
func TestResolvedFieldSet(t *testing.T) {
	reg := builtins.NewRegistry()
	var capturedCmd *parser.Command

	reg.Register(builtins.Definition{
		Name: "list",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			capturedCmd = cmd
			return 0, nil
		},
	})

	e := New(WithRegistry(reg))

	_, err := e.ExecuteInput(context.Background(), "l /tmp")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if capturedCmd == nil {
		t.Fatal("command not captured")
	}

	if capturedCmd.Name != "l" {
		t.Errorf("Name = %q, want %q", capturedCmd.Name, "l")
	}
	if capturedCmd.Resolved != "list" {
		t.Errorf("Resolved = %q, want %q", capturedCmd.Resolved, "list")
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
