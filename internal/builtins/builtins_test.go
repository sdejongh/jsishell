package builtins

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/parser"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for new registry", r.Count())
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	def := Definition{
		Name:        "test",
		Description: "A test command",
		Usage:       "test [args]",
		Handler: func(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
			return 0, nil
		},
	}

	r.Register(def)

	got, ok := r.Get("test")
	if !ok {
		t.Fatal("Get(test) returned false")
	}
	if got.Name != "test" {
		t.Errorf("got.Name = %q, want %q", got.Name, "test")
	}
	if got.Description != "A test command" {
		t.Errorf("got.Description = %q, want %q", got.Description, "A test command")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true, want false")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()

	// Register in non-alphabetical order
	r.Register(Definition{Name: "zebra"})
	r.Register(Definition{Name: "alpha"})
	r.Register(Definition{Name: "beta"})

	list := r.List()

	if len(list) != 3 {
		t.Fatalf("len(List()) = %d, want 3", len(list))
	}

	// Should be sorted
	expected := []string{"alpha", "beta", "zebra"}
	for i, want := range expected {
		if list[i] != want {
			t.Errorf("List()[%d] = %q, want %q", i, list[i], want)
		}
	}
}

func TestRegistryMatch(t *testing.T) {
	r := NewRegistry()

	r.Register(Definition{Name: "copy"})
	r.Register(Definition{Name: "cd"})
	r.Register(Definition{Name: "clear"})
	r.Register(Definition{Name: "list"})

	tests := []struct {
		prefix string
		want   []string
	}{
		{"c", []string{"cd", "clear", "copy"}},
		{"co", []string{"copy"}},
		{"cl", []string{"clear"}},
		{"l", []string{"list"}},
		{"x", nil},
		{"", []string{"cd", "clear", "copy", "list"}},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := r.Match(tt.prefix)

			if len(got) != len(tt.want) {
				t.Fatalf("Match(%q) = %v, want %v", tt.prefix, got, tt.want)
			}

			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("Match(%q)[%d] = %q, want %q", tt.prefix, i, got[i], want)
				}
			}
		})
	}
}

func TestRegistryAll(t *testing.T) {
	r := NewRegistry()

	r.Register(Definition{Name: "cmd1", Description: "desc1"})
	r.Register(Definition{Name: "cmd2", Description: "desc2"})

	all := r.All()

	if len(all) != 2 {
		t.Fatalf("len(All()) = %d, want 2", len(all))
	}

	if all["cmd1"].Description != "desc1" {
		t.Errorf("All()[cmd1].Description = %q, want %q", all["cmd1"].Description, "desc1")
	}
}

func TestRegistryCount(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}

	r.Register(Definition{Name: "cmd1"})
	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1", r.Count())
	}

	r.Register(Definition{Name: "cmd2"})
	if r.Count() != 2 {
		t.Errorf("Count() = %d, want 2", r.Count())
	}
}

func TestRegistryHas(t *testing.T) {
	r := NewRegistry()
	r.Register(Definition{Name: "exists"})

	if !r.Has("exists") {
		t.Error("Has(exists) = false, want true")
	}
	if r.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}
}

func TestRegistryOverwrite(t *testing.T) {
	r := NewRegistry()

	r.Register(Definition{Name: "cmd", Description: "original"})
	r.Register(Definition{Name: "cmd", Description: "updated"})

	def, _ := r.Get("cmd")
	if def.Description != "updated" {
		t.Errorf("Description = %q, want %q after overwrite", def.Description, "updated")
	}
}

func TestOptionDef(t *testing.T) {
	opt := OptionDef{
		Long:        "--verbose",
		Short:       "-v",
		Description: "Enable verbose output",
		HasValue:    false,
		Default:     "",
	}

	if opt.Long != "--verbose" {
		t.Errorf("Long = %q, want %q", opt.Long, "--verbose")
	}
	if opt.Short != "-v" {
		t.Errorf("Short = %q, want %q", opt.Short, "-v")
	}
}

func TestContextFields(t *testing.T) {
	ctx := &Context{
		WorkDir: "/home/user",
	}

	if ctx.WorkDir != "/home/user" {
		t.Errorf("WorkDir = %q, want %q", ctx.WorkDir, "/home/user")
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	done := make(chan bool, 20)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			r.Register(Definition{Name: "concurrent"})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			r.List()
			r.Match("c")
			r.Get("concurrent")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

// ============================================================================
// Builtin Command Handler Tests (T039, T051)
// ============================================================================

// createTestContext creates a Context for testing builtins.
func createTestContext() (*Context, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	environment := env.New()

	ctx := &Context{
		Stdin:   nil,
		Stdout:  stdout,
		Stderr:  stderr,
		Env:     environment,
		WorkDir: "/tmp",
	}
	return ctx, stdout, stderr
}

// TestEchoHandler tests the echo builtin command.
func TestEchoHandler(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]bool
		wantOut  string
		wantCode int
	}{
		{
			name:     "simple echo",
			args:     []string{"hello"},
			wantOut:  "hello\n",
			wantCode: 0,
		},
		{
			name:     "multiple args",
			args:     []string{"hello", "world"},
			wantOut:  "hello world\n",
			wantCode: 0,
		},
		{
			name:     "empty echo",
			args:     []string{},
			wantOut:  "\n",
			wantCode: 0,
		},
		{
			name:     "no newline flag short",
			args:     []string{"test"},
			flags:    map[string]bool{"-n": true},
			wantOut:  "test",
			wantCode: 0,
		},
		{
			name:     "no newline flag long",
			args:     []string{"test"},
			flags:    map[string]bool{"--no-newline": true},
			wantOut:  "test",
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCtx, stdout, _ := createTestContext()

			cmd := &parser.Command{
				Name:  "echo",
				Args:  tt.args,
				Flags: tt.flags,
			}
			if cmd.Flags == nil {
				cmd.Flags = make(map[string]bool)
			}

			code, err := echoHandler(context.Background(), cmd, execCtx)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if stdout.String() != tt.wantOut {
				t.Errorf("stdout = %q, want %q", stdout.String(), tt.wantOut)
			}
		})
	}
}

// TestExitHandler tests the exit builtin command.
func TestExitHandler(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  bool
	}{
		{
			name:     "default exit",
			args:     []string{},
			wantCode: 0,
			wantErr:  true, // ExitCode error expected
		},
		{
			name:     "exit 0",
			args:     []string{"0"},
			wantCode: 0,
			wantErr:  true,
		},
		{
			name:     "exit 1",
			args:     []string{"1"},
			wantCode: 1,
			wantErr:  true,
		},
		{
			name:     "exit 42",
			args:     []string{"42"},
			wantCode: 42,
			wantErr:  true,
		},
		{
			name:     "exit invalid",
			args:     []string{"not_a_number"},
			wantCode: 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCtx, _, stderr := createTestContext()

			cmd := &parser.Command{
				Name:  "exit",
				Args:  tt.args,
				Flags: make(map[string]bool),
			}

			code, err := exitHandler(context.Background(), cmd, execCtx)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}

			if tt.wantErr {
				exitErr, ok := err.(ExitCode)
				if !ok {
					t.Errorf("expected ExitCode error, got %T", err)
				} else if exitErr.Code != tt.wantCode {
					t.Errorf("ExitCode.Code = %d, want %d", exitErr.Code, tt.wantCode)
				}
			}

			// Check stderr for invalid exit code
			if tt.name == "exit invalid" && stderr.Len() == 0 {
				t.Error("expected error message in stderr for invalid exit code")
			}
		})
	}
}

// TestClearHandler tests the clear builtin command.
func TestClearHandler(t *testing.T) {
	execCtx, stdout, _ := createTestContext()

	cmd := &parser.Command{
		Name:  "clear",
		Flags: make(map[string]bool),
	}

	code, err := clearHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should output ANSI clear sequence
	expected := "\033[2J\033[H"
	if stdout.String() != expected {
		t.Errorf("stdout = %q, want %q", stdout.String(), expected)
	}
}

// TestClearHandlerHelp tests the clear --help flag.
func TestClearHandlerHelp(t *testing.T) {
	execCtx, stdout, _ := createTestContext()

	cmd := &parser.Command{
		Name:  "clear",
		Flags: map[string]bool{"--help": true},
	}

	code, err := clearHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if len(output) == 0 {
		t.Error("expected help output, got empty string")
	}
	if !bytes.Contains([]byte(output), []byte("Usage:")) {
		t.Errorf("help output should contain 'Usage:', got: %s", output)
	}
}

// TestEnvHandler tests the env builtin command.
func TestEnvHandler(t *testing.T) {
	t.Run("display all", func(t *testing.T) {
		execCtx, stdout, _ := createTestContext()
		execCtx.Env.Set("TEST_VAR", "test_value")

		cmd := &parser.Command{
			Name:  "env",
			Args:  []string{},
			Flags: make(map[string]bool),
		}

		code, err := envHandler(context.Background(), cmd, execCtx)

		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := stdout.String()
		if !bytes.Contains([]byte(output), []byte("TEST_VAR=test_value")) {
			t.Errorf("output should contain TEST_VAR=test_value, got: %s", output)
		}
	})

	t.Run("display specific", func(t *testing.T) {
		execCtx, stdout, _ := createTestContext()
		execCtx.Env.Set("MY_VAR", "my_value")

		cmd := &parser.Command{
			Name:  "env",
			Args:  []string{"MY_VAR"},
			Flags: make(map[string]bool),
		}

		code, err := envHandler(context.Background(), cmd, execCtx)

		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := "MY_VAR=my_value\n"
		if stdout.String() != expected {
			t.Errorf("stdout = %q, want %q", stdout.String(), expected)
		}
	})

	t.Run("set variable", func(t *testing.T) {
		execCtx, _, _ := createTestContext()

		cmd := &parser.Command{
			Name:  "env",
			Args:  []string{"NEW_VAR=new_value"},
			Flags: make(map[string]bool),
		}

		code, err := envHandler(context.Background(), cmd, execCtx)

		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if got := execCtx.Env.Get("NEW_VAR"); got != "new_value" {
			t.Errorf("NEW_VAR = %q, want %q", got, "new_value")
		}
	})
}

// TestHelpHandler tests the help builtin command.
func TestHelpHandler(t *testing.T) {
	// Setup registry with test commands
	reg := NewRegistry()
	reg.Register(Definition{
		Name:        "test",
		Description: "Test command",
		Usage:       "test [options]",
		Options: []OptionDef{
			{Long: "--verbose", Short: "-v", Description: "Verbose output"},
		},
	})
	reg.Register(Definition{
		Name:        "another",
		Description: "Another command",
		Usage:       "another",
	})
	SetHelpRegistry(reg)

	t.Run("general help", func(t *testing.T) {
		execCtx, stdout, _ := createTestContext()

		cmd := &parser.Command{
			Name:  "help",
			Args:  []string{},
			Flags: make(map[string]bool),
		}

		code, err := helpHandler(context.Background(), cmd, execCtx)

		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := stdout.String()
		if !bytes.Contains([]byte(output), []byte("test")) {
			t.Errorf("help output should contain 'test', got: %s", output)
		}
		if !bytes.Contains([]byte(output), []byte("another")) {
			t.Errorf("help output should contain 'another', got: %s", output)
		}
	})

	t.Run("specific command help", func(t *testing.T) {
		execCtx, stdout, _ := createTestContext()

		cmd := &parser.Command{
			Name:  "help",
			Args:  []string{"test"},
			Flags: make(map[string]bool),
		}

		code, err := helpHandler(context.Background(), cmd, execCtx)

		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := stdout.String()
		if !bytes.Contains([]byte(output), []byte("test - Test command")) {
			t.Errorf("help output should contain command description, got: %s", output)
		}
		if !bytes.Contains([]byte(output), []byte("--verbose")) {
			t.Errorf("help output should contain option --verbose, got: %s", output)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		execCtx, _, stderr := createTestContext()

		cmd := &parser.Command{
			Name:  "help",
			Args:  []string{"nonexistent"},
			Flags: make(map[string]bool),
		}

		code, err := helpHandler(context.Background(), cmd, execCtx)

		if code != 1 {
			t.Errorf("exit code = %d, want 1 for unknown command", code)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if stderr.Len() == 0 {
			t.Error("expected error message in stderr for unknown command")
		}
	})
}

// TestRegisterAll tests that RegisterAll registers all expected commands.
func TestRegisterAll(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	expectedCommands := []string{
		"echo", "exit", "help", "clear", "env",
		"goto", "here", "list", "makedir", "copy", "move", "remove",
		"reload", "history",
	}

	for _, cmd := range expectedCommands {
		if !reg.Has(cmd) {
			t.Errorf("RegisterAll should register %q command", cmd)
		}
	}
}

// T114: Test that all builtins support --help flag
func TestAllBuiltinsHaveHelp(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Commands that should have --help (all except 'help' itself)
	commandsWithHelp := []string{
		"echo", "exit", "clear", "env",
		"goto", "here", "list", "makedir", "copy", "move", "remove",
		"reload", "history",
	}

	for _, cmdName := range commandsWithHelp {
		def, ok := reg.Get(cmdName)
		if !ok {
			t.Errorf("Command %q not found", cmdName)
			continue
		}

		// Check that --help is in Options
		hasHelp := false
		for _, opt := range def.Options {
			if opt.Long == "--help" {
				hasHelp = true
				break
			}
		}

		if !hasHelp {
			t.Errorf("Command %q should have --help option defined", cmdName)
		}
	}
}

// T114: Test --help flag produces output for all commands
func TestHelpFlagProducesOutput(t *testing.T) {
	reg := NewRegistry()
	RegisterAll(reg)

	// Test each command's --help
	testCases := []struct {
		name    string
		handler Handler
	}{
		{"echo", echoHandler},
		{"exit", exitHandler},
		{"clear", clearHandler},
		{"env", envHandler},
		{"goto", gotoHandler},
		{"here", hereHandler},
		{"list", listHandler},
		{"makedir", makedirHandler},
		{"copy", copyHandler},
		{"move", moveHandler},
		{"remove", removeHandler},
		{"reload", reloadHandler},
		{"history", historyHandler},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			execCtx := &Context{
				Stdout:  &stdout,
				Stderr:  &stderr,
				Env:     env.New(),
				WorkDir: "/tmp",
			}

			cmd := &parser.Command{
				Name:  tc.name,
				Flags: map[string]bool{"--help": true},
			}

			code, err := tc.handler(context.Background(), cmd, execCtx)

			if code != 0 {
				t.Errorf("%s --help: exit code = %d, want 0", tc.name, code)
			}
			if err != nil {
				t.Errorf("%s --help: unexpected error: %v", tc.name, err)
			}

			output := stdout.String()
			if !bytes.Contains([]byte(output), []byte("Usage:")) && !bytes.Contains([]byte(output), []byte("usage:")) {
				t.Errorf("%s --help: output should contain 'Usage:', got: %s", tc.name, output)
			}
		})
	}
}

// TestEchoHelp tests that echo --help works.
func TestEchoHelp(t *testing.T) {
	execCtx, stdout, _ := createTestContext()

	cmd := &parser.Command{
		Name:  "echo",
		Flags: map[string]bool{"--help": true},
	}

	code, err := echoHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Usage:")) {
		t.Errorf("help output should contain 'Usage:', got: %s", output)
	}
}

// Mock history provider for testing
type mockHistoryProvider struct {
	entries []HistoryEntry
	cleared bool
}

func (m *mockHistoryProvider) Len() int {
	return len(m.entries)
}

func (m *mockHistoryProvider) All() []HistoryEntry {
	return m.entries
}

func (m *mockHistoryProvider) Clear() {
	m.entries = nil
	m.cleared = true
}

func TestHistoryCommand(t *testing.T) {
	// Create mock provider
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{
			{Command: "echo hello"},
			{Command: "ls -la"},
			{Command: "cd /home"},
			{Command: "pwd"},
			{Command: "history"},
		},
	}

	// Set provider
	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	// Test displaying all history
	cmd := &parser.Command{
		Name: "history",
	}

	code, err := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("echo hello")) {
		t.Errorf("output should contain 'echo hello', got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("ls -la")) {
		t.Errorf("output should contain 'ls -la', got: %s", output)
	}
}

func TestHistoryCommandWithCount(t *testing.T) {
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{
			{Command: "cmd1"},
			{Command: "cmd2"},
			{Command: "cmd3"},
			{Command: "cmd4"},
			{Command: "cmd5"},
		},
	}

	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	// Test displaying last 2 entries
	cmd := &parser.Command{
		Name: "history",
		Args: []string{"2"},
	}

	code, err := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should contain cmd4 and cmd5, but not cmd1, cmd2, cmd3
	if !bytes.Contains([]byte(output), []byte("cmd4")) {
		t.Errorf("output should contain 'cmd4', got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("cmd5")) {
		t.Errorf("output should contain 'cmd5', got: %s", output)
	}
	if bytes.Contains([]byte(output), []byte("cmd1")) {
		t.Errorf("output should NOT contain 'cmd1', got: %s", output)
	}
}

func TestHistoryCommandClear(t *testing.T) {
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{
			{Command: "cmd1"},
			{Command: "cmd2"},
		},
	}

	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	// Test clearing with -c
	cmd := &parser.Command{
		Name:  "history",
		Flags: map[string]bool{"-c": true},
	}

	code, err := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mock.cleared {
		t.Error("history should have been cleared")
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("cleared")) {
		t.Errorf("output should contain 'cleared', got: %s", output)
	}
}

func TestHistoryCommandClearLongFlag(t *testing.T) {
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{
			{Command: "cmd1"},
		},
	}

	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	// Test clearing with --clear
	cmd := &parser.Command{
		Name:  "history",
		Flags: map[string]bool{"--clear": true},
	}

	code, _ := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	if !mock.cleared {
		t.Error("history should have been cleared with --clear")
	}
}

func TestHistoryCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	cmd := &parser.Command{
		Name:  "history",
		Flags: map[string]bool{"--help": true},
	}

	code, err := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Usage:")) {
		t.Errorf("help output should contain 'Usage:', got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("--clear")) {
		t.Errorf("help output should contain '--clear', got: %s", output)
	}
}

func TestHistoryCommandNoProvider(t *testing.T) {
	// Clear any existing provider
	SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	cmd := &parser.Command{
		Name: "history",
	}

	code, _ := historyHandler(context.Background(), cmd, execCtx)

	if code != 1 {
		t.Errorf("exit code = %d, want 1 when no provider", code)
	}

	errOutput := stderr.String()
	if !bytes.Contains([]byte(errOutput), []byte("not available")) {
		t.Errorf("stderr should contain 'not available', got: %s", errOutput)
	}
}

func TestHistoryCommandInvalidCount(t *testing.T) {
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{
			{Command: "cmd1"},
		},
	}

	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	cmd := &parser.Command{
		Name: "history",
		Args: []string{"invalid"},
	}

	code, _ := historyHandler(context.Background(), cmd, execCtx)

	if code != 1 {
		t.Errorf("exit code = %d, want 1 for invalid count", code)
	}

	errOutput := stderr.String()
	if !bytes.Contains([]byte(errOutput), []byte("invalid count")) {
		t.Errorf("stderr should contain 'invalid count', got: %s", errOutput)
	}
}

func TestHistoryCommandEmptyHistory(t *testing.T) {
	mock := &mockHistoryProvider{
		entries: []HistoryEntry{},
	}

	SetHistoryProvider(func() HistoryProvider {
		return mock
	})
	defer SetHistoryProvider(nil)

	var stdout, stderr bytes.Buffer
	execCtx := &Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Env:    env.New(),
	}

	cmd := &parser.Command{
		Name: "history",
	}

	code, err := historyHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("No history")) {
		t.Errorf("output should contain 'No history', got: %s", output)
	}
}

// ============================================================================
// T115: Tests for --verbose and --quiet flags
// ============================================================================

// TestListVerboseOption tests list --verbose option.
func TestListVerboseOption(t *testing.T) {
	// Create a temp directory with files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	execCtx, stdout, _ := createTestContext()
	execCtx.WorkDir = tmpDir

	cmd := &parser.Command{
		Name:  "list",
		Args:  []string{tmpDir},
		Flags: map[string]bool{"--verbose": true},
	}

	code, err := listHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Verbose mode should include file type indicator
	if !bytes.Contains([]byte(output), []byte("[file]")) {
		t.Errorf("verbose output should contain '[file]', got: %s", output)
	}
}

// TestListQuietOption tests list --quiet option.
func TestListQuietOption(t *testing.T) {
	// Create a temp directory with files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	execCtx, stdout, _ := createTestContext()
	execCtx.WorkDir = tmpDir

	cmd := &parser.Command{
		Name:  "list",
		Args:  []string{tmpDir},
		Flags: map[string]bool{"--quiet": true},
	}

	code, err := listHandler(context.Background(), cmd, execCtx)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Quiet mode should only show file names without additional output
	expected := "test.txt\n"
	if output != expected {
		t.Errorf("quiet output = %q, want %q", output, expected)
	}
}

// TestRemoveQuietOption tests remove --quiet option.
func TestRemoveQuietOption(t *testing.T) {
	execCtx, stdout, stderr := createTestContext()

	cmd := &parser.Command{
		Name:  "remove",
		Args:  []string{"/nonexistent/path/file.txt"},
		Flags: map[string]bool{"--quiet": true},
	}

	code, err := removeHandler(context.Background(), cmd, execCtx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// With --quiet, no error output should be produced
	if stderr.String() != "" {
		t.Errorf("quiet mode should suppress errors, got stderr: %s", stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("quiet mode should produce no stdout, got: %s", stdout.String())
	}
	// Exit code should still be 0 with --quiet (missing file is ignored)
	if code != 0 {
		t.Errorf("exit code = %d, want 0 with --quiet", code)
	}
}

// TestRemoveYesAlias tests that --yes/-y works as alias for --force.
func TestRemoveYesAlias(t *testing.T) {
	execCtx, _, stderr := createTestContext()

	// Test with --yes flag
	cmd := &parser.Command{
		Name:  "remove",
		Args:  []string{"/nonexistent/path/file.txt"},
		Flags: map[string]bool{"--yes": true},
	}

	code, err := removeHandler(context.Background(), cmd, execCtx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// With --yes (alias for --force), no error for missing file
	if stderr.String() != "" {
		t.Errorf("--yes should suppress error for missing file, got stderr: %s", stderr.String())
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0 with --yes for missing file", code)
	}

	// Test with -y short flag
	execCtx2, _, stderr2 := createTestContext()
	cmd2 := &parser.Command{
		Name:  "remove",
		Args:  []string{"/another/nonexistent/file.txt"},
		Flags: map[string]bool{"-y": true},
	}

	code2, err2 := removeHandler(context.Background(), cmd2, execCtx2)

	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}
	if stderr2.String() != "" {
		t.Errorf("-y should suppress error for missing file, got stderr: %s", stderr2.String())
	}
	if code2 != 0 {
		t.Errorf("exit code = %d, want 0 with -y for missing file", code2)
	}
}
