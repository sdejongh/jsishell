// Package history provides command history functionality for the shell.
package history

import (
	"testing"
	"time"
)

// T103: Tests for history add/navigate

func TestNewHistory(t *testing.T) {
	h := New(100)

	if h == nil {
		t.Fatal("New() returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("Len() = %d, want 0", h.Len())
	}
	if h.MaxSize() != 100 {
		t.Errorf("MaxSize() = %d, want 100", h.MaxSize())
	}
}

func TestHistoryAdd(t *testing.T) {
	h := New(100)

	h.Add("first command")
	if h.Len() != 1 {
		t.Errorf("Len() = %d, want 1", h.Len())
	}

	h.Add("second command")
	if h.Len() != 2 {
		t.Errorf("Len() = %d, want 2", h.Len())
	}

	// Verify entries
	entries := h.All()
	if len(entries) != 2 {
		t.Fatalf("All() returned %d entries, want 2", len(entries))
	}
	if entries[0].Command != "first command" {
		t.Errorf("entries[0].Command = %q, want %q", entries[0].Command, "first command")
	}
	if entries[1].Command != "second command" {
		t.Errorf("entries[1].Command = %q, want %q", entries[1].Command, "second command")
	}
}

func TestHistoryAddWithTimestamp(t *testing.T) {
	h := New(100)

	before := time.Now()
	h.Add("test command")
	after := time.Now()

	entries := h.All()
	if len(entries) != 1 {
		t.Fatalf("All() returned %d entries, want 1", len(entries))
	}

	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestHistoryGet(t *testing.T) {
	h := New(100)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	tests := []struct {
		name    string
		index   int
		want    string
		wantErr bool
	}{
		{"first entry", 0, "cmd1", false},
		{"second entry", 1, "cmd2", false},
		{"third entry", 2, "cmd3", false},
		{"negative index", -1, "", true},
		{"out of bounds", 3, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := h.Get(tt.index)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Get(%d) should return error", tt.index)
				}
				return
			}
			if err != nil {
				t.Errorf("Get(%d) error: %v", tt.index, err)
				return
			}
			if entry.Command != tt.want {
				t.Errorf("Get(%d).Command = %q, want %q", tt.index, entry.Command, tt.want)
			}
		})
	}
}

func TestHistoryNavigate(t *testing.T) {
	h := New(100)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	// Reset navigation position
	h.ResetNavigation()

	// Navigate backward (Up arrow - older commands)
	cmd, ok := h.Previous()
	if !ok || cmd != "cmd3" {
		t.Errorf("Previous() = %q, %v, want %q, true", cmd, ok, "cmd3")
	}

	cmd, ok = h.Previous()
	if !ok || cmd != "cmd2" {
		t.Errorf("Previous() = %q, %v, want %q, true", cmd, ok, "cmd2")
	}

	cmd, ok = h.Previous()
	if !ok || cmd != "cmd1" {
		t.Errorf("Previous() = %q, %v, want %q, true", cmd, ok, "cmd1")
	}

	// At beginning, should return false
	cmd, ok = h.Previous()
	if ok {
		t.Errorf("Previous() at beginning should return false, got %q", cmd)
	}

	// Navigate forward (Down arrow - newer commands)
	cmd, ok = h.Next()
	if !ok || cmd != "cmd2" {
		t.Errorf("Next() = %q, %v, want %q, true", cmd, ok, "cmd2")
	}

	cmd, ok = h.Next()
	if !ok || cmd != "cmd3" {
		t.Errorf("Next() = %q, %v, want %q, true", cmd, ok, "cmd3")
	}

	// At end, should return false (or empty line)
	cmd, ok = h.Next()
	if ok {
		t.Errorf("Next() at end should return false, got %q", cmd)
	}
}

func TestHistoryNavigateEmpty(t *testing.T) {
	h := New(100)
	h.ResetNavigation()

	cmd, ok := h.Previous()
	if ok {
		t.Errorf("Previous() on empty history should return false, got %q", cmd)
	}

	cmd, ok = h.Next()
	if ok {
		t.Errorf("Next() on empty history should return false, got %q", cmd)
	}
}

func TestHistoryNavigateWithCurrentLine(t *testing.T) {
	h := New(100)
	h.Add("cmd1")
	h.Add("cmd2")

	// Set current line (what user is typing)
	h.SetCurrentLine("partial input")
	h.ResetNavigation()

	// Go up
	cmd, ok := h.Previous()
	if !ok || cmd != "cmd2" {
		t.Errorf("Previous() = %q, %v, want %q, true", cmd, ok, "cmd2")
	}

	// Go down - should get back to current line
	cmd, ok = h.Next()
	if ok {
		t.Errorf("Next() should return false when at current line position")
	}

	// Get current line
	current := h.CurrentLine()
	if current != "partial input" {
		t.Errorf("CurrentLine() = %q, want %q", current, "partial input")
	}
}

// T105: Tests for history search

func TestHistorySearch(t *testing.T) {
	h := New(100)
	h.Add("ls -la")
	h.Add("cd /home")
	h.Add("ls Documents")
	h.Add("cat file.txt")
	h.Add("ls -l")

	// Search for "ls"
	results := h.Search("ls")
	if len(results) != 3 {
		t.Errorf("Search(\"ls\") returned %d results, want 3", len(results))
	}

	// Verify results are in reverse chronological order (most recent first)
	expected := []string{"ls -l", "ls Documents", "ls -la"}
	for i, result := range results {
		if result.Command != expected[i] {
			t.Errorf("results[%d].Command = %q, want %q", i, result.Command, expected[i])
		}
	}
}

func TestHistorySearchCaseInsensitive(t *testing.T) {
	h := New(100)
	h.Add("LS -la")
	h.Add("ls Documents")
	h.Add("Ls -L")

	results := h.Search("ls")
	if len(results) != 3 {
		t.Errorf("Search(\"ls\") returned %d results, want 3", len(results))
	}
}

func TestHistorySearchEmpty(t *testing.T) {
	h := New(100)
	h.Add("ls -la")
	h.Add("cd /home")

	// Empty search returns all
	results := h.Search("")
	if len(results) != 2 {
		t.Errorf("Search(\"\") returned %d results, want 2", len(results))
	}
}

func TestHistorySearchNoMatch(t *testing.T) {
	h := New(100)
	h.Add("ls -la")
	h.Add("cd /home")

	results := h.Search("xyz")
	if len(results) != 0 {
		t.Errorf("Search(\"xyz\") returned %d results, want 0", len(results))
	}
}

func TestHistorySearchPrefix(t *testing.T) {
	h := New(100)
	h.Add("list files")
	h.Add("list -la")
	h.Add("goto /home")
	h.Add("list Documents")

	// Search with prefix match only
	results := h.SearchPrefix("list")
	if len(results) != 3 {
		t.Errorf("SearchPrefix(\"list\") returned %d results, want 3", len(results))
	}

	// "goto" should not match prefix "list"
	for _, r := range results {
		if r.Command == "goto /home" {
			t.Error("SearchPrefix should not match non-prefix commands")
		}
	}
}

func TestHistorySearchNavigate(t *testing.T) {
	h := New(100)
	h.Add("ls -la")
	h.Add("cd /home")
	h.Add("ls Documents")
	h.Add("cat file.txt")
	h.Add("ls -l")

	// Start search
	h.StartSearch("ls")

	// Navigate through search results
	cmd, ok := h.PreviousMatch()
	if !ok || cmd != "ls -l" {
		t.Errorf("PreviousMatch() = %q, %v, want %q, true", cmd, ok, "ls -l")
	}

	cmd, ok = h.PreviousMatch()
	if !ok || cmd != "ls Documents" {
		t.Errorf("PreviousMatch() = %q, %v, want %q, true", cmd, ok, "ls Documents")
	}

	cmd, ok = h.PreviousMatch()
	if !ok || cmd != "ls -la" {
		t.Errorf("PreviousMatch() = %q, %v, want %q, true", cmd, ok, "ls -la")
	}

	// No more matches
	cmd, ok = h.PreviousMatch()
	if ok {
		t.Errorf("PreviousMatch() should return false at end, got %q", cmd)
	}

	// Go forward
	cmd, ok = h.NextMatch()
	if !ok || cmd != "ls Documents" {
		t.Errorf("NextMatch() = %q, %v, want %q, true", cmd, ok, "ls Documents")
	}

	// End search
	h.EndSearch()
}

func TestHistoryClear(t *testing.T) {
	h := New(100)
	h.Add("cmd1")
	h.Add("cmd2")

	h.Clear()

	if h.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", h.Len())
	}
}

func TestHistoryLast(t *testing.T) {
	h := New(100)

	// Empty history
	_, err := h.Last()
	if err == nil {
		t.Error("Last() on empty history should return error")
	}

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	entry, err := h.Last()
	if err != nil {
		t.Errorf("Last() error: %v", err)
	}
	if entry.Command != "cmd3" {
		t.Errorf("Last().Command = %q, want %q", entry.Command, "cmd3")
	}
}
