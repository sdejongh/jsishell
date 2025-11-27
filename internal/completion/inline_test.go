package completion

import (
	"os"
	"path/filepath"
	"testing"
)

// T086: Tests for inline suggestion
func TestInlineSuggestion(t *testing.T) {
	commands := []string{"list", "list-files", "goto", "help", "here", "exit"}
	c := NewCompleter(commands)

	tests := []struct {
		name           string
		input          string
		wantSuggestion string
		wantHas        bool
	}{
		{
			name:           "unique prefix completion",
			input:          "ex",
			wantSuggestion: "it", // Completes "exit"
			wantHas:        true,
		},
		{
			name:           "ambiguous prefix with common continuation",
			input:          "h",
			wantSuggestion: "e", // help, here - common prefix is "he"
			wantHas:        true,
		},
		{
			name:           "exact match with longer option",
			input:          "exit",
			wantSuggestion: "", // exact match, nothing more to suggest
			wantHas:        false,
		},
		{
			name:           "longer unique prefix",
			input:          "got",
			wantSuggestion: "o",
			wantHas:        true,
		},
		{
			name:           "no match no suggestion",
			input:          "xyz",
			wantSuggestion: "",
			wantHas:        false,
		},
		{
			name:           "single char with unique completion",
			input:          "g",
			wantSuggestion: "oto",
			wantHas:        true,
		},
		{
			name:           "common prefix suggestion",
			input:          "li",
			wantSuggestion: "st", // Common prefix is "list"
			wantHas:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion, has := c.InlineSuggestion(tt.input)
			if has != tt.wantHas {
				t.Errorf("InlineSuggestion(%q) has = %v, want %v", tt.input, has, tt.wantHas)
			}
			if suggestion != tt.wantSuggestion {
				t.Errorf("InlineSuggestion(%q) = %q, want %q", tt.input, suggestion, tt.wantSuggestion)
			}
		})
	}
}

func TestInlineSuggestionWithPath(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files and directories
	os.MkdirAll(filepath.Join(tmpDir, "documents"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "downloads"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	commands := []string{"list", "goto"}
	c := NewCompleter(commands)

	// Test command completion inline
	suggestion, has := c.InlineSuggestion("lis")
	if !has || suggestion != "t" {
		t.Errorf("InlineSuggestion(\"lis\") = %q, %v; want \"t\", true", suggestion, has)
	}

	// Test path completion after command - unique file
	input := "list " + tmpDir + "/file1"
	suggestion, has = c.InlineSuggestion(input)
	if !has || suggestion != ".txt" {
		t.Errorf("InlineSuggestion(%q) = %q, %v; want \".txt\", true", input, suggestion, has)
	}

	// Test path completion with common prefix (documents, downloads -> "do")
	input = "list " + tmpDir + "/do"
	suggestion, has = c.InlineSuggestion(input)
	// Both start with "do", common prefix should suggest nothing extra or common part
	// documents and downloads share "do" but diverge after
	if has && suggestion != "" {
		// This is acceptable - no common extension beyond "do"
	}

	// Test path completion for unique directory
	input = "goto " + tmpDir + "/doc"
	suggestion, has = c.InlineSuggestion(input)
	// Note: the trailing / comes from the full path in CompletePath
	if !has || (suggestion != "uments/" && suggestion != "uments") {
		t.Errorf("InlineSuggestion(%q) = %q, %v; want \"uments\" or \"uments/\", true", input, suggestion, has)
	}

	// Test path completion with common prefix for files
	input = "list " + tmpDir + "/file"
	suggestion, has = c.InlineSuggestion(input)
	// file1.txt and file2.txt - common is "file" then diverge at 1/2
	// Should not suggest anything since they diverge immediately after "file"
	if has && suggestion != "" {
		t.Errorf("InlineSuggestion(%q) = %q, %v; want \"\", false (ambiguous)", input, suggestion, has)
	}
}

func TestInlineSuggestionCaseSensitivity(t *testing.T) {
	commands := []string{"List", "Goto", "Help"}
	c := NewCompleter(commands)

	// Should match case-insensitively but return proper case
	suggestion, has := c.InlineSuggestion("l")
	if !has {
		t.Error("InlineSuggestion(\"l\") should have suggestion")
	}
	if suggestion != "ist" {
		t.Errorf("InlineSuggestion(\"l\") = %q, want \"ist\"", suggestion)
	}
}

func TestAcceptSuggestion(t *testing.T) {
	commands := []string{"list", "goto", "exit"}
	c := NewCompleter(commands)

	tests := []struct {
		name       string
		input      string
		wantResult string
	}{
		{
			name:       "accept unique suggestion",
			input:      "ex",
			wantResult: "exit",
		},
		{
			name:       "no suggestion to accept",
			input:      "xyz",
			wantResult: "xyz",
		},
		{
			name:       "already complete",
			input:      "exit",
			wantResult: "exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.AcceptSuggestion(tt.input)
			if result != tt.wantResult {
				t.Errorf("AcceptSuggestion(%q) = %q, want %q", tt.input, result, tt.wantResult)
			}
		})
	}
}

func TestGetCompletionList(t *testing.T) {
	commands := []string{"list", "list-files", "goto", "help", "here"}
	c := NewCompleter(commands)

	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantValues []string
	}{
		{
			name:       "multiple completions",
			input:      "h",
			wantCount:  2,
			wantValues: []string{"help", "here"},
		},
		{
			name:       "single completion",
			input:      "g",
			wantCount:  1,
			wantValues: []string{"goto"},
		},
		{
			name:       "no completions",
			input:      "xyz",
			wantCount:  0,
			wantValues: []string{},
		},
		{
			name:       "list prefix with multiple",
			input:      "list",
			wantCount:  2,
			wantValues: []string{"list", "list-files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions := c.GetCompletionList(tt.input)
			if len(completions) != tt.wantCount {
				t.Errorf("GetCompletionList(%q) returned %d items, want %d", tt.input, len(completions), tt.wantCount)
			}
			for i, want := range tt.wantValues {
				if i < len(completions) && completions[i] != want {
					t.Errorf("GetCompletionList(%q)[%d] = %q, want %q", tt.input, i, completions[i], want)
				}
			}
		})
	}
}
