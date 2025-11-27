package parser

import (
	"testing"

	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/lexer"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Args == nil {
		t.Error("Args should not be nil")
	}
	if cmd.Options == nil {
		t.Error("Options should not be nil")
	}
	if cmd.Flags == nil {
		t.Error("Flags should not be nil")
	}
}

func TestCommandHasFlag(t *testing.T) {
	cmd := NewCommand()
	cmd.Flags["--verbose"] = true
	cmd.Flags["-v"] = true

	if !cmd.HasFlag("--verbose") {
		t.Error("HasFlag(--verbose) should be true")
	}
	if !cmd.HasFlag("-v") {
		t.Error("HasFlag(-v) should be true")
	}
	if !cmd.HasFlag("--verbose", "-v") {
		t.Error("HasFlag(--verbose, -v) should be true")
	}
	if cmd.HasFlag("--quiet") {
		t.Error("HasFlag(--quiet) should be false")
	}
}

func TestCommandGetOption(t *testing.T) {
	cmd := NewCommand()
	cmd.Options["--output"] = "/tmp/file"
	cmd.Options["-o"] = "/tmp/other"

	if got := cmd.GetOption("--output"); got != "/tmp/file" {
		t.Errorf("GetOption(--output) = %q, want %q", got, "/tmp/file")
	}
	if got := cmd.GetOption("-o"); got != "/tmp/other" {
		t.Errorf("GetOption(-o) = %q, want %q", got, "/tmp/other")
	}
	if got := cmd.GetOption("--missing"); got != "" {
		t.Errorf("GetOption(--missing) = %q, want empty", got)
	}
	if got := cmd.GetOption("--output", "-o"); got != "/tmp/file" {
		t.Errorf("GetOption(--output, -o) = %q, want %q", got, "/tmp/file")
	}
}

func TestCommandGetOptionOr(t *testing.T) {
	cmd := NewCommand()
	cmd.Options["--output"] = "/tmp/file"

	if got := cmd.GetOptionOr("default", "--output"); got != "/tmp/file" {
		t.Errorf("GetOptionOr(default, --output) = %q, want %q", got, "/tmp/file")
	}
	if got := cmd.GetOptionOr("default", "--missing"); got != "default" {
		t.Errorf("GetOptionOr(default, --missing) = %q, want %q", got, "default")
	}
}

func TestCommandArg(t *testing.T) {
	cmd := NewCommand()
	cmd.Args = []string{"arg1", "arg2", "arg3"}

	if got := cmd.Arg(0); got != "arg1" {
		t.Errorf("Arg(0) = %q, want %q", got, "arg1")
	}
	if got := cmd.Arg(2); got != "arg3" {
		t.Errorf("Arg(2) = %q, want %q", got, "arg3")
	}
	if got := cmd.Arg(-1); got != "" {
		t.Errorf("Arg(-1) = %q, want empty", got)
	}
	if got := cmd.Arg(10); got != "" {
		t.Errorf("Arg(10) = %q, want empty", got)
	}
}

func TestCommandArgCount(t *testing.T) {
	cmd := NewCommand()
	cmd.Args = []string{"a", "b", "c"}

	if got := cmd.ArgCount(); got != 3 {
		t.Errorf("ArgCount() = %d, want 3", got)
	}
}

func TestParseSimpleCommand(t *testing.T) {
	tests := []struct {
		input string
		name  string
		args  []string
	}{
		{"list", "list", nil},
		{"cd /home", "cd", []string{"/home"}},
		{"copy src dest", "copy", []string{"src", "dest"}},
		{"echo hello world", "echo", []string{"hello", "world"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("ParseInput error: %v", err)
			}

			if cmd.Name != tt.name {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.name)
			}

			if len(tt.args) == 0 && len(cmd.Args) > 0 {
				t.Errorf("Args = %v, want empty", cmd.Args)
			} else if len(tt.args) > 0 {
				if len(cmd.Args) != len(tt.args) {
					t.Fatalf("len(Args) = %d, want %d", len(cmd.Args), len(tt.args))
				}
				for i, arg := range tt.args {
					if cmd.Args[i] != arg {
						t.Errorf("Args[%d] = %q, want %q", i, cmd.Args[i], arg)
					}
				}
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	cmd, err := ParseInput("")
	if err != nil {
		t.Fatalf("ParseInput error: %v", err)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil for empty input", cmd)
	}

	cmd, err = ParseInput("   ")
	if err != nil {
		t.Fatalf("ParseInput error: %v", err)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil for whitespace-only input", cmd)
	}
}

func TestParseWithFlags(t *testing.T) {
	tests := []struct {
		input string
		name  string
		flags []string
	}{
		{"list -a", "list", []string{"-a"}},
		{"list --all", "list", []string{"--all"}},
		{"list -a --long", "list", []string{"-a", "--long"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("ParseInput error: %v", err)
			}

			if cmd.Name != tt.name {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.name)
			}

			for _, flag := range tt.flags {
				if !cmd.Flags[flag] {
					t.Errorf("Flag %q not set", flag)
				}
			}
		})
	}
}

func TestParseWithOptions(t *testing.T) {
	tests := []struct {
		input   string
		name    string
		options map[string]string
	}{
		{
			"copy --dest=/tmp src",
			"copy",
			map[string]string{"--dest": "/tmp"},
		},
		{
			"list --format=long",
			"list",
			map[string]string{"--format": "long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("ParseInput error: %v", err)
			}

			if cmd.Name != tt.name {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.name)
			}

			for key, want := range tt.options {
				got := cmd.Options[key]
				if got != want {
					t.Errorf("Options[%q] = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestParseQuotedStrings(t *testing.T) {
	tests := []struct {
		input string
		args  []string
	}{
		{`echo "hello world"`, []string{"hello world"}},
		{`echo 'single quotes'`, []string{"single quotes"}},
		{`copy "src file" "dest file"`, []string{"src file", "dest file"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("ParseInput error: %v", err)
			}

			if len(cmd.Args) != len(tt.args) {
				t.Fatalf("len(Args) = %d, want %d", len(cmd.Args), len(tt.args))
			}

			for i, want := range tt.args {
				if cmd.Args[i] != want {
					t.Errorf("Args[%d] = %q, want %q", i, cmd.Args[i], want)
				}
			}
		})
	}
}

func TestParseWithVariables(t *testing.T) {
	e := env.New()
	e.Set("HOME", "/home/user")
	e.Set("NAME", "John")

	tests := []struct {
		input string
		name  string
		args  []string
	}{
		{"cd $HOME", "cd", []string{"/home/user"}},
		{"echo $NAME", "echo", []string{"John"}},
		{"echo ${HOME}", "echo", []string{"/home/user"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseInputWithEnv(tt.input, e)
			if err != nil {
				t.Fatalf("ParseInputWithEnv error: %v", err)
			}

			if cmd.Name != tt.name {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.name)
			}

			if len(cmd.Args) != len(tt.args) {
				t.Fatalf("len(Args) = %d, want %d", len(cmd.Args), len(tt.args))
			}

			for i, want := range tt.args {
				if cmd.Args[i] != want {
					t.Errorf("Args[%d] = %q, want %q", i, cmd.Args[i], want)
				}
			}
		})
	}
}

func TestParseRawInputPreserved(t *testing.T) {
	input := "list --all /home"
	cmd, err := ParseInput(input)
	if err != nil {
		t.Fatalf("ParseInput error: %v", err)
	}

	if cmd.RawInput != input {
		t.Errorf("RawInput = %q, want %q", cmd.RawInput, input)
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	// Create tokens with an error token
	tokens := []lexer.Token{
		{Type: lexer.TokenError, Value: "bad", Literal: "unterminated string"},
	}

	parser := New(tokens)
	_, err := parser.Parse()

	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestParseCommandWithMixedArgs(t *testing.T) {
	input := `copy --recursive -v "source dir" /dest --force`

	cmd, err := ParseInput(input)
	if err != nil {
		t.Fatalf("ParseInput error: %v", err)
	}

	if cmd.Name != "copy" {
		t.Errorf("Name = %q, want %q", cmd.Name, "copy")
	}

	// Check args
	expectedArgs := []string{"source dir", "/dest"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("len(Args) = %d, want %d", len(cmd.Args), len(expectedArgs))
	}
	for i, want := range expectedArgs {
		if cmd.Args[i] != want {
			t.Errorf("Args[%d] = %q, want %q", i, cmd.Args[i], want)
		}
	}

	// Check flags
	expectedFlags := []string{"--recursive", "-v", "--force"}
	for _, flag := range expectedFlags {
		if !cmd.Flags[flag] {
			t.Errorf("Flag %q not set", flag)
		}
	}
}

func TestParserSkipsWhitespace(t *testing.T) {
	input := "  list   -a    /home  "

	cmd, err := ParseInput(input)
	if err != nil {
		t.Fatalf("ParseInput error: %v", err)
	}

	if cmd.Name != "list" {
		t.Errorf("Name = %q, want %q", cmd.Name, "list")
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "/home" {
		t.Errorf("Args = %v, want [/home]", cmd.Args)
	}
	if !cmd.Flags["-a"] {
		t.Error("Flag -a not set")
	}
}
