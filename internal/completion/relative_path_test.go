package completion

import (
	"os"
	"path/filepath"
	"testing"
)

// T087: Tests for relative path completion
func TestRelativePathCompletion(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test directories and files
	os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "images"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "index.txt"), []byte("test"), 0644)

	// Change to temp directory for relative path testing
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	c := NewCompleter([]string{"list", "goto"})

	t.Run("relative path without prefix", func(t *testing.T) {
		pathCandidates := c.CompletePath("int")
		if len(pathCandidates) != 1 {
			t.Errorf("CompletePath(\"int\") returned %d candidates, want 1", len(pathCandidates))
		} else if pathCandidates[0].Text != "internal/" {
			t.Errorf("CompletePath(\"int\") returned %q, want \"internal/\"", pathCandidates[0].Text)
		}
	})

	t.Run("relative path with ./ prefix", func(t *testing.T) {
		pathCandidates := c.CompletePath("./int")
		if len(pathCandidates) != 1 {
			t.Errorf("CompletePath(\"./int\") returned %d candidates, want 1", len(pathCandidates))
		} else if pathCandidates[0].Text != "./internal/" {
			t.Errorf("CompletePath(\"./int\") returned %q, want \"./internal/\"", pathCandidates[0].Text)
		}
	})

	t.Run("inline suggestion for relative path after command", func(t *testing.T) {
		input := "list int"
		suggestion, has := c.InlineSuggestion(input)
		if !has || suggestion != "ernal/" {
			t.Errorf("InlineSuggestion(%q) = %q, %v; want \"ernal/\", true", input, suggestion, has)
		}
	})

	t.Run("inline suggestion for ./ relative path after command", func(t *testing.T) {
		input := "list ./int"
		suggestion, has := c.InlineSuggestion(input)
		if !has || suggestion != "ernal/" {
			t.Errorf("InlineSuggestion(%q) = %q, %v; want \"ernal/\", true", input, suggestion, has)
		}
	})

	t.Run("multiple matches with common prefix", func(t *testing.T) {
		pathCandidates := c.CompletePath("i")
		if len(pathCandidates) != 3 {
			t.Errorf("CompletePath(\"i\") returned %d candidates, want 3", len(pathCandidates))
		}
	})

	// Tests for path completion at start of line (no command yet)
	t.Run("path at start of line with ./", func(t *testing.T) {
		candidates := c.Complete("./int")
		if len(candidates) != 1 {
			t.Errorf("Complete(\"./int\") returned %d candidates, want 1", len(candidates))
		} else if candidates[0].Text != "./internal/" {
			t.Errorf("Complete(\"./int\") returned %q, want \"./internal/\"", candidates[0].Text)
		}
	})

	t.Run("inline suggestion for path at start of line", func(t *testing.T) {
		suggestion, has := c.InlineSuggestion("./int")
		if !has || suggestion != "ernal/" {
			t.Errorf("InlineSuggestion(\"./int\") = %q, %v; want \"ernal/\", true", suggestion, has)
		}
	})

	t.Run("absolute path at start of line", func(t *testing.T) {
		// Create a test file in tmpDir and use absolute path
		absPath := filepath.Join(tmpDir, "int")
		candidates := c.Complete(absPath)
		if len(candidates) != 1 {
			t.Errorf("Complete(%q) returned %d candidates, want 1", absPath, len(candidates))
		}
	})

	t.Run("../path at start of line", func(t *testing.T) {
		// Create subdirectory and test ../ completion
		subDir := filepath.Join(tmpDir, "subdir")
		os.MkdirAll(subDir, 0755)
		os.Chdir(subDir)
		defer os.Chdir(tmpDir)

		candidates := c.Complete("../int")
		if len(candidates) != 1 {
			t.Errorf("Complete(\"../int\") returned %d candidates, want 1", len(candidates))
		} else if candidates[0].Text != "../internal/" {
			t.Errorf("Complete(\"../int\") returned %q, want \"../internal/\"", candidates[0].Text)
		}
	})
}
