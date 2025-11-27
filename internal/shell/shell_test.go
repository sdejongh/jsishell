package shell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/executor"
	"github.com/sdejongh/jsishell/internal/parser"
)

func TestNewShell(t *testing.T) {
	s := New()

	if s == nil {
		t.Fatal("New() returned nil")
	}
	// Prompt comes from config file or defaults - just check it's not empty
	if s.promptFormat == "" {
		t.Error("promptFormat should not be empty")
	}
	if s.executor == nil {
		t.Error("executor should not be nil")
	}
	if s.env == nil {
		t.Error("env should not be nil")
	}
}

func TestShellWithOptions(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	environ := env.New()

	s := New(
		WithPrompt("test> "),
		WithStdin(&stdin),
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithEnv(environ),
	)

	if s.promptFormat != "test> " {
		t.Errorf("promptFormat = %q, want %q", s.promptFormat, "test> ")
	}
	if s.env != environ {
		t.Error("env not set correctly")
	}
}

func TestShellExecute(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "test",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			execCtx.Stdout.Write([]byte("executed"))
			return 0, nil
		},
	})

	exec := executor.New(
		executor.WithRegistry(reg),
		executor.WithStdout(&stdout),
	)

	s := New(WithExecutor(exec))

	exitCode, err := s.Execute("test")
	if err != nil {
		t.Errorf("Execute error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if stdout.String() != "executed" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "executed")
	}
}

func TestShellExitCode(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "fail",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			return 42, nil
		},
	})

	exec := executor.New(
		executor.WithRegistry(reg),
		executor.WithStdout(&stdout),
	)

	s := New(WithExecutor(exec))

	exitCode, _ := s.Execute("fail")
	if exitCode != 42 {
		t.Errorf("exitCode = %d, want 42", exitCode)
	}
}

func TestShellSetPrompt(t *testing.T) {
	s := New()

	s.SetPrompt("new> ")
	if s.promptFormat != "new> " {
		t.Errorf("promptFormat = %q, want %q", s.promptFormat, "new> ")
	}
}

func TestShellPromptExpansion(t *testing.T) {
	s := New()

	// Test that prompt variables get expanded
	s.SetPrompt("%D> ")
	expanded := s.expandedPrompt()

	// Should not contain the literal %D
	if strings.Contains(expanded, "%D") {
		t.Errorf("expanded prompt should not contain %%D, got %q", expanded)
	}

	// Should end with "> "
	if !strings.HasSuffix(expanded, "> ") {
		t.Errorf("expanded prompt should end with '> ', got %q", expanded)
	}
}

func TestShellIsRunning(t *testing.T) {
	s := New()

	if s.IsRunning() {
		t.Error("IsRunning() should be false initially")
	}
}

func TestShellExit(t *testing.T) {
	s := New()
	s.running = true

	s.Exit(5)

	if s.IsRunning() {
		t.Error("IsRunning() should be false after Exit")
	}
	if s.ExitCode() != 5 {
		t.Errorf("ExitCode() = %d, want 5", s.ExitCode())
	}
}

func TestShellExecutor(t *testing.T) {
	s := New()

	if s.Executor() == nil {
		t.Error("Executor() should not be nil")
	}
}

func TestShellEnv(t *testing.T) {
	s := New()

	if s.Env() == nil {
		t.Error("Env() should not be nil")
	}
}

func TestShellRunWithExit(t *testing.T) {
	// Create a shell that reads "exit" command
	input := "exit 0\n"
	var stdout, stderr bytes.Buffer

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	exec := executor.New(
		executor.WithRegistry(reg),
		executor.WithStdout(&stdout),
		executor.WithStderr(&stderr),
	)

	s := New(
		WithStdin(strings.NewReader(input)),
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithExecutor(exec),
	)

	err := s.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

func TestShellRunWithEOF(t *testing.T) {
	// Empty input should cause EOF
	var stdout, stderr bytes.Buffer

	s := New(
		WithStdin(strings.NewReader("")),
		WithStdout(&stdout),
		WithStderr(&stderr),
	)

	err := s.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

func TestShellRunMultipleCommands(t *testing.T) {
	input := "echo hello\necho world\nexit\n"
	var stdout, stderr bytes.Buffer

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	exec := executor.New(
		executor.WithRegistry(reg),
		executor.WithStdout(&stdout),
		executor.WithStderr(&stderr),
	)

	s := New(
		WithStdin(strings.NewReader(input)),
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithExecutor(exec),
		WithPrompt(""),
	)

	err := s.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}

	// Check output contains both echo results
	if !strings.Contains(stdout.String(), "hello") {
		t.Error("stdout should contain 'hello'")
	}
	if !strings.Contains(stdout.String(), "world") {
		t.Error("stdout should contain 'world'")
	}
}

func TestShellSkipsEmptyLines(t *testing.T) {
	input := "\n\necho test\n\nexit\n"
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	builtins.RegisterAll(reg)

	exec := executor.New(
		executor.WithRegistry(reg),
		executor.WithStdout(&stdout),
	)

	s := New(
		WithStdin(strings.NewReader(input)),
		WithStdout(&stdout),
		WithExecutor(exec),
		WithPrompt(""),
	)

	err := s.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}

	// Should still execute the echo
	if !strings.Contains(stdout.String(), "test") {
		t.Error("stdout should contain 'test'")
	}
}

// ============================================================================
// T125: Performance benchmark for startup time (<100ms target)
// ============================================================================

// BenchmarkShellStartup measures the time to create and initialize a shell.
func BenchmarkShellStartup(b *testing.B) {
	var stdout, stderr bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := New(
			WithStdin(strings.NewReader("")),
			WithStdout(&stdout),
			WithStderr(&stderr),
		)
		_ = s
		stdout.Reset()
		stderr.Reset()
	}
}

// BenchmarkShellStartupWithAllBuiltins measures startup with all builtins registered.
func BenchmarkShellStartupWithAllBuiltins(b *testing.B) {
	var stdout, stderr bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg := builtins.NewRegistry()
		builtins.RegisterAll(reg)

		exec := executor.New(
			executor.WithRegistry(reg),
			executor.WithStdout(&stdout),
			executor.WithStderr(&stderr),
		)

		s := New(
			WithStdin(strings.NewReader("")),
			WithStdout(&stdout),
			WithStderr(&stderr),
			WithExecutor(exec),
		)
		_ = s
		stdout.Reset()
		stderr.Reset()
	}
}

// TestStartupTimeMeetsTarget verifies startup time is under 100ms.
func TestStartupTimeMeetsTarget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := New(
				WithStdin(strings.NewReader("")),
				WithStdout(&stdout),
				WithStderr(&stderr),
			)
			_ = s
			stdout.Reset()
			stderr.Reset()
		}
	})

	// Calculate average time per operation in milliseconds
	avgMs := float64(start.T.Nanoseconds()) / float64(start.N) / 1e6

	// Target: startup should be under 100ms
	if avgMs > 100 {
		t.Errorf("Shell startup time %.2fms exceeds 100ms target", avgMs)
	}
	t.Logf("Shell startup time: %.2fms (target: <100ms)", avgMs)
}
