package completion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T084: Tests for command completion
func TestCommandCompletion(t *testing.T) {
	// Create a completer with known commands
	commands := []string{"list", "list-files", "goto", "goto-home", "help", "here", "exit", "echo", "copy", "clear"}
	c := NewCompleter(commands)

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst string
	}{
		{
			name:      "single char prefix",
			input:     "l",
			wantCount: 2, // list, list-files
			wantFirst: "list",
		},
		{
			name:      "exact match returns itself",
			input:     "list",
			wantCount: 2, // list, list-files
			wantFirst: "list",
		},
		{
			name:      "no matches",
			input:     "xyz",
			wantCount: 0,
			wantFirst: "",
		},
		{
			name:      "multiple prefix matches",
			input:     "h",
			wantCount: 2, // help, here
			wantFirst: "help",
		},
		{
			name:      "unique prefix",
			input:     "ex",
			wantCount: 1,
			wantFirst: "exit",
		},
		{
			name:      "case insensitive",
			input:     "L",
			wantCount: 2,
			wantFirst: "list",
		},
		{
			name:      "empty input returns all",
			input:     "",
			wantCount: 10, // all commands
			wantFirst: "clear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := c.CompleteCommand(tt.input)
			if len(candidates) != tt.wantCount {
				t.Errorf("CompleteCommand(%q) returned %d candidates, want %d", tt.input, len(candidates), tt.wantCount)
			}
			if tt.wantCount > 0 && len(candidates) > 0 && candidates[0].Text != tt.wantFirst {
				t.Errorf("CompleteCommand(%q)[0].Text = %q, want %q", tt.input, candidates[0].Text, tt.wantFirst)
			}
		})
	}
}

func TestCompletionCandidateType(t *testing.T) {
	commands := []string{"list", "exit"}
	c := NewCompleter(commands)

	candidates := c.CompleteCommand("l")
	if len(candidates) == 0 {
		t.Fatal("Expected at least one candidate")
	}

	if candidates[0].Type != TypeCommand {
		t.Errorf("Expected TypeCommand, got %v", candidates[0].Type)
	}
}

// T085: Tests for path completion
func TestPathCompletion(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files and directories
	dirs := []string{"documents", "downloads", "desktop"}
	files := []string{"file1.txt", "file2.txt", "readme.md"}

	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmpDir, d), 0755)
	}
	for _, f := range files {
		os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644)
	}

	// Create a subdirectory with files
	os.MkdirAll(filepath.Join(tmpDir, "documents", "work"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "documents", "work", "report.txt"), []byte("test"), 0644)

	c := NewCompleter(nil)

	tests := []struct {
		name        string
		input       string
		wantCount   int
		wantContain string
	}{
		{
			name:        "list directory contents",
			input:       tmpDir + "/",
			wantCount:   6, // 3 dirs + 3 files
			wantContain: "documents",
		},
		{
			name:        "prefix d",
			input:       tmpDir + "/d",
			wantCount:   3, // documents, downloads, desktop
			wantContain: "documents",
		},
		{
			name:        "prefix do",
			input:       tmpDir + "/do",
			wantCount:   2, // documents, downloads
			wantContain: "documents",
		},
		{
			name:        "exact match directory",
			input:       tmpDir + "/documents",
			wantCount:   1,
			wantContain: "documents",
		},
		{
			name:        "subdirectory",
			input:       tmpDir + "/documents/",
			wantCount:   1, // work subdirectory
			wantContain: "work",
		},
		{
			name:        "file prefix",
			input:       tmpDir + "/file",
			wantCount:   2, // file1.txt, file2.txt
			wantContain: "file1.txt",
		},
		{
			name:        "no matches",
			input:       tmpDir + "/xyz",
			wantCount:   0,
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := c.CompletePath(tt.input)
			if len(candidates) != tt.wantCount {
				names := make([]string, len(candidates))
				for i, c := range candidates {
					names[i] = c.Text
				}
				t.Errorf("CompletePath(%q) returned %d candidates %v, want %d", tt.input, len(candidates), names, tt.wantCount)
			}
			if tt.wantContain != "" {
				found := false
				for _, c := range candidates {
					if c.Text == tt.wantContain || filepath.Base(c.Text) == tt.wantContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("CompletePath(%q) does not contain %q", tt.input, tt.wantContain)
				}
			}
		})
	}
}

func TestPathCompletionType(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)
	// Create executable file
	execPath := filepath.Join(tmpDir, "script.sh")
	os.WriteFile(execPath, []byte("#!/bin/bash"), 0755)

	c := NewCompleter(nil)
	candidates := c.CompletePath(tmpDir + "/")

	// Check we have the right types
	typeMap := make(map[string]CompletionType)
	for _, cand := range candidates {
		typeMap[filepath.Base(cand.Text)] = cand.Type
	}

	if typeMap["subdir"] != TypeDirectory {
		t.Errorf("subdir should be TypeDirectory, got %v", typeMap["subdir"])
	}
	if typeMap["file.txt"] != TypeFile {
		t.Errorf("file.txt should be TypeFile, got %v", typeMap["file.txt"])
	}
	if typeMap["script.sh"] != TypeExecutable {
		t.Errorf("script.sh should be TypeExecutable, got %v", typeMap["script.sh"])
	}
}

func TestTildePathCompletion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	c := NewCompleter(nil)

	// Test completion of ~ alone
	candidates := c.CompletePath("~")
	if len(candidates) == 0 {
		t.Error("~ should return home directory contents")
	}

	// All results should start with ~
	for _, cand := range candidates {
		if !strings.HasPrefix(cand.Text, "~") {
			t.Errorf("completion for ~ should start with ~, got %q", cand.Text)
		}
	}

	// Test completion of ~/
	candidates = c.CompletePath("~/")
	if len(candidates) == 0 {
		t.Error("~/ should return home directory contents")
	}

	// All results should start with ~/
	for _, cand := range candidates {
		if !strings.HasPrefix(cand.Text, "~/") {
			t.Errorf("completion for ~/ should start with ~/, got %q", cand.Text)
		}
	}

	// Test completion with partial path after ~
	// Create a test directory in home if possible
	testDir := filepath.Join(homeDir, ".jsishell_test_completion")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Skip("Cannot create test directory in home")
	}
	defer os.RemoveAll(testDir)

	candidates = c.CompletePath("~/.jsishell_test")
	found := false
	for _, cand := range candidates {
		if strings.Contains(cand.Text, ".jsishell_test_completion") {
			found = true
			if !strings.HasPrefix(cand.Text, "~") {
				t.Errorf("completion should preserve ~, got %q", cand.Text)
			}
			break
		}
	}
	if !found {
		t.Error("should find .jsishell_test_completion directory")
	}
}

func TestCompleteInput(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	commands := []string{"list", "goto", "help"}
	c := NewCompleter(commands)

	tests := []struct {
		name      string
		input     string
		wantType  CompletionType
		wantCount int
	}{
		{
			name:      "command at start",
			input:     "l",
			wantType:  TypeCommand,
			wantCount: 1, // list
		},
		{
			name:      "path after command",
			input:     "list " + tmpDir + "/t",
			wantType:  TypeFile,
			wantCount: 1, // test.txt
		},
		{
			name:      "empty input",
			input:     "",
			wantType:  TypeCommand,
			wantCount: 3, // all commands
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := c.Complete(tt.input)
			if len(candidates) != tt.wantCount {
				t.Errorf("Complete(%q) returned %d candidates, want %d", tt.input, len(candidates), tt.wantCount)
			}
			if tt.wantCount > 0 && candidates[0].Type != tt.wantType {
				t.Errorf("Complete(%q)[0].Type = %v, want %v", tt.input, candidates[0].Type, tt.wantType)
			}
		})
	}
}

// Tests for option completion
func TestOptionCompletion(t *testing.T) {
	// Create command definitions with options
	defs := []CommandDef{
		{
			Name:        "list",
			Description: "List directory contents",
			Options: []OptionDef{
				{Long: "--all", Short: "-a", Description: "Include hidden files"},
				{Long: "--long", Short: "-l", Description: "Long format"},
				{Long: "--help", Description: "Show help"},
			},
		},
		{
			Name:        "copy",
			Description: "Copy files",
			Options: []OptionDef{
				{Long: "--recursive", Short: "-r", Description: "Copy recursively"},
				{Long: "--force", Short: "-f", Description: "Force overwrite"},
				{Long: "--verbose", Short: "-v", Description: "Verbose output"},
			},
		},
		{
			Name:        "echo",
			Description: "Print arguments",
			Options:     nil, // No options
		},
	}

	c := NewCompleterWithDefs(defs)

	tests := []struct {
		name      string
		command   string
		prefix    string
		wantCount int
		wantFirst string
	}{
		{
			name:      "list all options with double dash",
			command:   "list",
			prefix:    "--",
			wantCount: 3, // --all, --long, --help
			wantFirst: "--all",
		},
		{
			name:      "list options starting with --a",
			command:   "list",
			prefix:    "--a",
			wantCount: 1, // --all
			wantFirst: "--all",
		},
		{
			name:      "list short options with single dash",
			command:   "list",
			prefix:    "-",
			wantCount: 5, // --all, --help, --long, -a, -l (sorted alphabetically)
			wantFirst: "--all",
		},
		{
			name:      "list short option -a",
			command:   "list",
			prefix:    "-a",
			wantCount: 1, // -a
			wantFirst: "-a",
		},
		{
			name:      "copy recursive option",
			command:   "copy",
			prefix:    "--r",
			wantCount: 1, // --recursive
			wantFirst: "--recursive",
		},
		{
			name:      "command with no options",
			command:   "echo",
			prefix:    "--",
			wantCount: 0,
			wantFirst: "",
		},
		{
			name:      "unknown command",
			command:   "unknown",
			prefix:    "--",
			wantCount: 0,
			wantFirst: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := c.CompleteOption(tt.command, tt.prefix)
			if len(candidates) != tt.wantCount {
				names := make([]string, len(candidates))
				for i, cand := range candidates {
					names[i] = cand.Text
				}
				t.Errorf("CompleteOption(%q, %q) returned %d candidates %v, want %d",
					tt.command, tt.prefix, len(candidates), names, tt.wantCount)
			}
			if tt.wantCount > 0 && len(candidates) > 0 && candidates[0].Text != tt.wantFirst {
				t.Errorf("CompleteOption(%q, %q)[0].Text = %q, want %q",
					tt.command, tt.prefix, candidates[0].Text, tt.wantFirst)
			}
			// Verify all candidates are TypeOption
			for _, cand := range candidates {
				if cand.Type != TypeOption {
					t.Errorf("candidate %q should be TypeOption, got %v", cand.Text, cand.Type)
				}
			}
		})
	}
}

func TestCompleteWithOptions(t *testing.T) {
	defs := []CommandDef{
		{
			Name:        "list",
			Description: "List directory contents",
			Options: []OptionDef{
				{Long: "--all", Short: "-a", Description: "Include hidden files"},
				{Long: "--long", Short: "-l", Description: "Long format"},
			},
		},
	}

	c := NewCompleterWithDefs(defs)

	tests := []struct {
		name      string
		input     string
		wantType  CompletionType
		wantCount int
	}{
		{
			name:      "option completion after command",
			input:     "list --",
			wantType:  TypeOption,
			wantCount: 2, // --all, --long
		},
		{
			name:      "option completion with prefix",
			input:     "list --a",
			wantType:  TypeOption,
			wantCount: 1, // --all
		},
		{
			name:      "short option completion",
			input:     "list -a",
			wantType:  TypeOption,
			wantCount: 1, // -a
		},
		{
			name:      "command completion still works",
			input:     "li",
			wantType:  TypeCommand,
			wantCount: 1, // list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := c.Complete(tt.input)
			if len(candidates) != tt.wantCount {
				names := make([]string, len(candidates))
				for i, cand := range candidates {
					names[i] = cand.Text
				}
				t.Errorf("Complete(%q) returned %d candidates %v, want %d",
					tt.input, len(candidates), names, tt.wantCount)
			}
			if tt.wantCount > 0 && candidates[0].Type != tt.wantType {
				t.Errorf("Complete(%q)[0].Type = %v, want %v", tt.input, candidates[0].Type, tt.wantType)
			}
		})
	}
}

func TestNewCompleterWithDefs(t *testing.T) {
	defs := []CommandDef{
		{Name: "list", Description: "List files"},
		{Name: "copy", Description: "Copy files"},
		{Name: "move", Description: "Move files"},
	}

	c := NewCompleterWithDefs(defs)

	// Verify commands are set
	if len(c.commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(c.commands))
	}

	// Verify command completion works
	candidates := c.CompleteCommand("l")
	if len(candidates) != 1 {
		t.Errorf("Expected 1 candidate for 'l', got %d", len(candidates))
	}
	if candidates[0].Text != "list" {
		t.Errorf("Expected 'list', got %q", candidates[0].Text)
	}
}

func TestSetCommandDefs(t *testing.T) {
	c := NewCompleter([]string{"old"})

	defs := []CommandDef{
		{Name: "new1"},
		{Name: "new2"},
	}

	c.SetCommandDefs(defs)

	// Old commands should be replaced
	candidates := c.CompleteCommand("old")
	if len(candidates) != 0 {
		t.Errorf("Old command should not exist, got %d candidates", len(candidates))
	}

	// New commands should exist
	candidates = c.CompleteCommand("new")
	if len(candidates) != 2 {
		t.Errorf("Expected 2 new commands, got %d", len(candidates))
	}
}

func TestOptionCompletionDescription(t *testing.T) {
	defs := []CommandDef{
		{
			Name: "test",
			Options: []OptionDef{
				{Long: "--verbose", Description: "Enable verbose output"},
			},
		},
	}

	c := NewCompleterWithDefs(defs)
	candidates := c.CompleteOption("test", "--v")

	if len(candidates) != 1 {
		t.Fatalf("Expected 1 candidate, got %d", len(candidates))
	}

	if candidates[0].Description != "Enable verbose output" {
		t.Errorf("Expected description 'Enable verbose output', got %q", candidates[0].Description)
	}
}
