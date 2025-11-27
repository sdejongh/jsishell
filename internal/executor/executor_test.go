package executor

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/sdejongh/jsishell/internal/builtins"
	"github.com/sdejongh/jsishell/internal/env"
	shellerrors "github.com/sdejongh/jsishell/internal/errors"
	"github.com/sdejongh/jsishell/internal/parser"
)

func TestNewExecutor(t *testing.T) {
	e := New()

	if e.registry == nil {
		t.Error("registry should not be nil")
	}
	if e.env == nil {
		t.Error("env should not be nil")
	}
	if e.stdin == nil {
		t.Error("stdin should not be nil")
	}
	if e.stdout == nil {
		t.Error("stdout should not be nil")
	}
	if e.stderr == nil {
		t.Error("stderr should not be nil")
	}
}

func TestExecutorWithOptions(t *testing.T) {
	reg := builtins.NewRegistry()
	environ := env.New()
	var stdin, stdout, stderr bytes.Buffer

	e := New(
		WithRegistry(reg),
		WithEnv(environ),
		WithStdin(&stdin),
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithWorkDir("/tmp"),
		WithAbbreviations(false),
	)

	if e.Registry() != reg {
		t.Error("registry not set correctly")
	}
	if e.Env() != environ {
		t.Error("env not set correctly")
	}
	if e.WorkDir() != "/tmp" {
		t.Errorf("workDir = %q, want %q", e.WorkDir(), "/tmp")
	}
}

func TestExecuteBuiltin(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name:        "test",
		Description: "Test command",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			execCtx.Stdout.Write([]byte("executed"))
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
	)

	cmd := &parser.Command{Name: "test"}
	exitCode, err := e.Execute(context.Background(), cmd)

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

func TestExecuteNilCommand(t *testing.T) {
	e := New()
	exitCode, err := e.Execute(context.Background(), nil)

	if err != nil {
		t.Errorf("Execute(nil) error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}

func TestExecuteCommandNotFound(t *testing.T) {
	e := New(WithAbbreviations(false))
	cmd := &parser.Command{Name: "nonexistentcommand12345"}

	exitCode, err := e.Execute(context.Background(), cmd)

	if exitCode != 127 {
		t.Errorf("exitCode = %d, want 127", exitCode)
	}
	if !errors.Is(err, shellerrors.ErrCommandNotFound) {
		t.Errorf("error = %v, want ErrCommandNotFound", err)
	}
}

func TestResolveCommandExactMatch(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "list"})
	reg.Register(builtins.Definition{Name: "copy"})

	e := New(WithRegistry(reg))

	resolved, alts, err := e.ResolveCommand("list")
	if err != nil {
		t.Errorf("ResolveCommand error: %v", err)
	}
	if resolved != "list" {
		t.Errorf("resolved = %q, want %q", resolved, "list")
	}
	if alts != nil {
		t.Errorf("alternatives = %v, want nil", alts)
	}
}

func TestResolveCommandAbbreviation(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "list"})
	reg.Register(builtins.Definition{Name: "copy"})

	e := New(WithRegistry(reg))

	// "l" should resolve to "list" (only match)
	resolved, _, err := e.ResolveCommand("l")
	if err != nil {
		t.Errorf("ResolveCommand error: %v", err)
	}
	if resolved != "list" {
		t.Errorf("resolved = %q, want %q", resolved, "list")
	}
}

func TestResolveCommandAmbiguous(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "copy"})
	reg.Register(builtins.Definition{Name: "cd"})
	reg.Register(builtins.Definition{Name: "clear"})

	e := New(WithRegistry(reg))

	// "c" should be ambiguous
	_, alts, err := e.ResolveCommand("c")
	if !errors.Is(err, shellerrors.ErrAmbiguousCommand) {
		t.Errorf("error = %v, want ErrAmbiguousCommand", err)
	}
	if len(alts) != 3 {
		t.Errorf("len(alternatives) = %d, want 3", len(alts))
	}
}

func TestResolveCommandDisabledAbbreviations(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "list"})

	e := New(WithRegistry(reg), WithAbbreviations(false))

	// "l" should not resolve when abbreviations disabled
	_, _, err := e.ResolveCommand("l")
	if !errors.Is(err, shellerrors.ErrCommandNotFound) {
		t.Errorf("error = %v, want ErrCommandNotFound", err)
	}

	// "list" should still work
	resolved, _, err := e.ResolveCommand("list")
	if err != nil {
		t.Errorf("ResolveCommand error: %v", err)
	}
	if resolved != "list" {
		t.Errorf("resolved = %q, want %q", resolved, "list")
	}
}

func TestExecuteInput(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "echo",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			for i, arg := range cmd.Args {
				if i > 0 {
					execCtx.Stdout.Write([]byte(" "))
				}
				execCtx.Stdout.Write([]byte(arg))
			}
			execCtx.Stdout.Write([]byte("\n"))
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
	)

	exitCode, err := e.ExecuteInput(context.Background(), "echo hello world")

	if err != nil {
		t.Errorf("ExecuteInput error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if stdout.String() != "hello world\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "hello world\n")
	}
}

func TestExecuteBuiltinWithArgs(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "greet",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			name := "World"
			if len(cmd.Args) > 0 {
				name = cmd.Args[0]
			}
			execCtx.Stdout.Write([]byte("Hello, " + name + "!\n"))
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
	)

	exitCode, err := e.ExecuteInput(context.Background(), "greet Alice")

	if err != nil {
		t.Errorf("ExecuteInput error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
	if stdout.String() != "Hello, Alice!\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "Hello, Alice!\n")
	}
}

func TestExecuteBuiltinWithFlags(t *testing.T) {
	var stdout bytes.Buffer

	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "verbose",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			if cmd.HasFlag("-v", "--verbose") {
				execCtx.Stdout.Write([]byte("verbose mode\n"))
			} else {
				execCtx.Stdout.Write([]byte("normal mode\n"))
			}
			return 0, nil
		},
	})

	e := New(
		WithRegistry(reg),
		WithStdout(&stdout),
	)

	// With flag
	e.ExecuteInput(context.Background(), "verbose -v")
	if stdout.String() != "verbose mode\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "verbose mode\n")
	}

	// Without flag
	stdout.Reset()
	e.ExecuteInput(context.Background(), "verbose")
	if stdout.String() != "normal mode\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "normal mode\n")
	}
}

func TestExecuteBuiltinError(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{
		Name: "fail",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *builtins.Context) (int, error) {
			return 1, errors.New("intentional error")
		},
	})

	e := New(WithRegistry(reg))

	exitCode, err := e.ExecuteInput(context.Background(), "fail")

	if exitCode != 1 {
		t.Errorf("exitCode = %d, want 1", exitCode)
	}
	if err == nil || err.Error() != "intentional error" {
		t.Errorf("error = %v, want 'intentional error'", err)
	}
}

func TestSetWorkDir(t *testing.T) {
	e := New()

	err := e.SetWorkDir("/tmp")
	if err != nil {
		t.Errorf("SetWorkDir error: %v", err)
	}

	if e.WorkDir() != "/tmp" {
		t.Errorf("WorkDir() = %q, want %q", e.WorkDir(), "/tmp")
	}

	// Check PWD env var was set
	if e.Env().Get("PWD") != "/tmp" {
		t.Errorf("PWD env = %q, want %q", e.Env().Get("PWD"), "/tmp")
	}
}

func TestExecuteExternal(t *testing.T) {
	var stdout bytes.Buffer

	e := New(WithStdout(&stdout))

	// Test with a common external command (echo or true)
	exitCode, err := e.ExecuteInput(context.Background(), "true")

	if err != nil {
		t.Errorf("ExecuteInput error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}

func TestExecuteExternalNotFound(t *testing.T) {
	e := New()

	exitCode, err := e.ExecuteInput(context.Background(), "thiscommanddoesnotexist12345")

	if exitCode != 127 {
		t.Errorf("exitCode = %d, want 127", exitCode)
	}
	if !errors.Is(err, shellerrors.ErrCommandNotFound) {
		t.Errorf("error = %v, want ErrCommandNotFound", err)
	}
}

func TestResolveCommandPrefersBuiltin(t *testing.T) {
	reg := builtins.NewRegistry()
	reg.Register(builtins.Definition{Name: "echo"}) // Override external echo

	e := New(WithRegistry(reg))

	resolved, _, err := e.ResolveCommand("echo")
	if err != nil {
		t.Errorf("ResolveCommand error: %v", err)
	}
	if resolved != "echo" {
		t.Errorf("resolved = %q, want %q", resolved, "echo")
	}

	// Verify it's the builtin, not external
	if !e.registry.Has(resolved) {
		t.Error("should resolve to builtin, not external")
	}
}
