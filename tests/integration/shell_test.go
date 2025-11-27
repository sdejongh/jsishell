// Package integration provides integration tests for JSIShell.
package integration

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/executor"
)

// TestBasicCommandExecution tests that basic commands execute correctly.
// T026: Write integration test for basic command execution
func TestBasicCommandExecution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOut  string
		wantCode int
	}{
		{
			name:     "echo simple",
			input:    "echo hello",
			wantOut:  "hello\n",
			wantCode: 0,
		},
		{
			name:     "echo multiple args",
			input:    "echo hello world",
			wantOut:  "hello world\n",
			wantCode: 0,
		},
		{
			name:     "echo with quotes",
			input:    `echo "hello world"`,
			wantOut:  "hello world\n",
			wantCode: 0,
		},
		{
			name:     "echo empty",
			input:    "echo",
			wantOut:  "\n",
			wantCode: 0,
		},
		{
			name:     "help command",
			input:    "help",
			wantOut:  "", // Just check it runs without error
			wantCode: 0,
		},
		{
			name:     "here command",
			input:    "here",
			wantOut:  "", // Will output current directory
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			reg := builtins.NewRegistry()
			builtins.RegisterAll(reg)

			e := executor.New(
				executor.WithRegistry(reg),
				executor.WithEnv(env.New()),
				executor.WithStdout(stdout),
				executor.WithStderr(stderr),
			)

			ctx := context.Background()
			code, err := e.ExecuteInput(ctx, tt.input)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}

			// Only check output if expected output is specified
			if tt.wantOut != "" {
				if got := stdout.String(); got != tt.wantOut {
					t.Errorf("stdout = %q, want %q", got, tt.wantOut)
				}
			}

			// For successful commands, stderr should be empty
			if tt.wantCode == 0 && stderr.Len() > 0 {
				t.Errorf("unexpected stderr output: %q", stderr.String())
			}

			// No error expected for successful commands
			if tt.wantCode == 0 && err != nil {
				// ExitCode is a special case
				if _, ok := err.(builtins.ExitCode); !ok {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestInvalidCommandErrorHandling tests error handling for invalid commands.
// T027: Write test for invalid command error handling
func TestInvalidCommandErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCode   int
		wantStderr string
	}{
		{
			name:       "command not found",
			input:      "nonexistentcommand12345",
			wantCode:   127,
			wantStderr: "command not found",
		},
		{
			name:       "ambiguous command",
			input:      "c", // Could match clear, copy, etc.
			wantCode:   1,
			wantStderr: "ambiguous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			reg := builtins.NewRegistry()
			builtins.RegisterAll(reg)

			e := executor.New(
				executor.WithRegistry(reg),
				executor.WithEnv(env.New()),
				executor.WithStdout(stdout),
				executor.WithStderr(stderr),
				executor.WithAbbreviations(true),
			)

			ctx := context.Background()
			code, err := e.ExecuteInput(ctx, tt.input)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}

			if err == nil {
				t.Error("expected an error, got nil")
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.wantStderr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantStderr)
			}
		})
	}
}

// TestExitCommand tests the exit builtin.
func TestExitCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCode int
	}{
		{
			name:     "exit default",
			input:    "exit",
			wantCode: 0,
		},
		{
			name:     "exit 0",
			input:    "exit 0",
			wantCode: 0,
		},
		{
			name:     "exit 1",
			input:    "exit 1",
			wantCode: 1,
		},
		{
			name:     "exit 42",
			input:    "exit 42",
			wantCode: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			reg := builtins.NewRegistry()
			builtins.RegisterAll(reg)

			e := executor.New(
				executor.WithRegistry(reg),
				executor.WithEnv(env.New()),
				executor.WithStdout(stdout),
				executor.WithStderr(stderr),
			)

			ctx := context.Background()
			code, err := e.ExecuteInput(ctx, tt.input)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}

			// exit should return ExitCode error
			exitErr, ok := err.(builtins.ExitCode)
			if !ok {
				t.Errorf("expected ExitCode error, got %T", err)
			} else if exitErr.Code != tt.wantCode {
				t.Errorf("ExitCode.Code = %d, want %d", exitErr.Code, tt.wantCode)
			}
		})
	}
}

// TestCommandTimeout tests that commands respect context cancellation.
// T028: Write test for Ctrl+C interrupt (via context cancellation)
func TestCommandTimeout(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	e := executor.New(
		executor.WithRegistry(reg),
		executor.WithEnv(env.New()),
		executor.WithStdout(stdout),
		executor.WithStderr(stderr),
	)

	// Create a context that's already cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(5 * time.Millisecond)

	// Execute a command - it should handle the cancelled context gracefully
	// For builtins, they typically don't check context, so this mainly tests external commands
	code, _ := e.ExecuteInput(ctx, "echo test")

	// The command may or may not complete depending on timing
	// But it should not panic
	_ = code
}

// TestVariableExpansion tests that environment variables are expanded correctly.
func TestVariableExpansion(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	environment := env.New()
	environment.Set("TEST_VAR", "hello")
	environment.Set("ANOTHER", "world")

	e := executor.New(
		executor.WithRegistry(reg),
		executor.WithEnv(environment),
		executor.WithStdout(stdout),
		executor.WithStderr(stderr),
	)

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantOut string
	}{
		{
			name:    "simple variable",
			input:   "echo $TEST_VAR",
			wantOut: "hello\n",
		},
		{
			name:    "variable with braces",
			input:   "echo ${TEST_VAR}",
			wantOut: "hello\n",
		},
		{
			name:    "multiple variables",
			input:   "echo $TEST_VAR $ANOTHER",
			wantOut: "hello world\n",
		},
		{
			name:    "undefined variable",
			input:   "echo $UNDEFINED",
			wantOut: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()

			code, err := e.ExecuteInput(ctx, tt.input)

			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got := stdout.String(); got != tt.wantOut {
				t.Errorf("stdout = %q, want %q", got, tt.wantOut)
			}
		})
	}
}

// TestExternalCommandExecution tests executing external commands.
func TestExternalCommandExecution(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	e := executor.New(
		executor.WithRegistry(reg),
		executor.WithEnv(env.New()),
		executor.WithStdout(stdout),
		executor.WithStderr(stderr),
	)

	ctx := context.Background()

	// Test with 'true' command which exists on all Unix systems
	code, err := e.ExecuteInput(ctx, "true")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	// Test with 'false' command
	stdout.Reset()
	stderr.Reset()

	code, err = e.ExecuteInput(ctx, "false")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

// TestAbbreviations tests command abbreviation resolution.
func TestAbbreviations(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantResolved  string
		wantCode      int
		abbreviations bool
	}{
		{
			name:          "unique prefix",
			input:         "hel", // Should resolve to "help"
			wantResolved:  "help",
			wantCode:      0,
			abbreviations: true,
		},
		{
			name:          "exact match",
			input:         "echo test",
			wantResolved:  "echo",
			wantCode:      0,
			abbreviations: true,
		},
		{
			name:          "abbreviations disabled",
			input:         "ech", // Should fail when abbreviations disabled
			wantResolved:  "",
			wantCode:      127,
			abbreviations: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			reg := builtins.NewRegistry()
			builtins.RegisterAll(reg)

			e := executor.New(
				executor.WithRegistry(reg),
				executor.WithEnv(env.New()),
				executor.WithStdout(stdout),
				executor.WithStderr(stderr),
				executor.WithAbbreviations(tt.abbreviations),
			)

			ctx := context.Background()
			code, _ := e.ExecuteInput(ctx, tt.input)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}
