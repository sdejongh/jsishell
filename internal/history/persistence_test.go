package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// T104: Tests for history persistence

func TestHistorySave(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "test_history")

	h := New(100)
	h.Add("first command")
	h.Add("second command")
	h.Add("third command")

	err := h.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(histFile); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Read file content
	content, err := os.ReadFile(histFile)
	if err != nil {
		t.Fatalf("Reading history file: %v", err)
	}

	// Should contain all commands
	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("History file is empty")
	}
}

func TestHistoryLoad(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "test_history")

	// Create and save history
	h1 := New(100)
	h1.Add("cmd1")
	h1.Add("cmd2")
	h1.Add("cmd3")
	err := h1.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load into new history
	h2 := New(100)
	err = h2.Load(histFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if h2.Len() != 3 {
		t.Errorf("Len() = %d, want 3", h2.Len())
	}

	// Verify commands are in order
	entries := h2.All()
	expected := []string{"cmd1", "cmd2", "cmd3"}
	for i, e := range expected {
		if entries[i].Command != e {
			t.Errorf("entries[%d].Command = %q, want %q", i, entries[i].Command, e)
		}
	}
}

func TestHistoryLoadNonExistent(t *testing.T) {
	h := New(100)
	err := h.Load("/nonexistent/path/history")

	// Should not return error for non-existent file (just start with empty history)
	if err != nil {
		t.Errorf("Load() should not error for non-existent file: %v", err)
	}
	if h.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for non-existent file", h.Len())
	}
}

func TestHistoryLoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "empty_history")

	// Create empty file
	err := os.WriteFile(histFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Creating empty file: %v", err)
	}

	h := New(100)
	err = h.Load(histFile)
	if err != nil {
		t.Errorf("Load() error on empty file: %v", err)
	}
	if h.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for empty file", h.Len())
	}
}

func TestHistoryPersistenceRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "roundtrip_history")

	commands := []string{
		"ls -la",
		"cd /home/user",
		"echo 'hello world'",
		"cat file.txt | grep pattern",
		"command with\ttab",
	}

	// Save
	h1 := New(100)
	for _, cmd := range commands {
		h1.Add(cmd)
	}
	err := h1.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load
	h2 := New(100)
	err = h2.Load(histFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify
	if h2.Len() != len(commands) {
		t.Errorf("Len() = %d, want %d", h2.Len(), len(commands))
	}

	entries := h2.All()
	for i, cmd := range commands {
		if entries[i].Command != cmd {
			t.Errorf("entries[%d].Command = %q, want %q", i, entries[i].Command, cmd)
		}
	}
}

func TestHistorySaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "subdir", "nested", "history")

	h := New(100)
	h.Add("test command")

	err := h.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(histFile); os.IsNotExist(err) {
		t.Fatal("History file was not created with nested directories")
	}
}

func TestHistoryLoadWithTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "ts_history")

	// Create history with specific timestamp format
	h1 := New(100)
	h1.Add("cmd1")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	h1.Add("cmd2")

	err := h1.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load and verify timestamps are preserved (or at least reasonable)
	h2 := New(100)
	err = h2.Load(histFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	entries := h2.All()
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Timestamps should be non-zero
	if entries[0].Timestamp.IsZero() {
		t.Error("First entry has zero timestamp")
	}
	if entries[1].Timestamp.IsZero() {
		t.Error("Second entry has zero timestamp")
	}
}

func TestHistoryAppend(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "append_history")

	// First save
	h1 := New(100)
	h1.Add("cmd1")
	h1.Add("cmd2")
	err := h1.Save(histFile)
	if err != nil {
		t.Fatalf("First Save() error: %v", err)
	}

	// Append new command
	h2 := New(100)
	err = h2.Load(histFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	h2.Add("cmd3")
	err = h2.Save(histFile)
	if err != nil {
		t.Fatalf("Second Save() error: %v", err)
	}

	// Load and verify
	h3 := New(100)
	err = h3.Load(histFile)
	if err != nil {
		t.Fatalf("Final Load() error: %v", err)
	}

	if h3.Len() != 3 {
		t.Errorf("Len() = %d, want 3", h3.Len())
	}

	entries := h3.All()
	expected := []string{"cmd1", "cmd2", "cmd3"}
	for i, cmd := range expected {
		if entries[i].Command != cmd {
			t.Errorf("entries[%d].Command = %q, want %q", i, entries[i].Command, cmd)
		}
	}
}

func TestHistorySaveTrimToMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "trim_history")

	// Create history with maxSize 3
	h := New(3)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")
	h.Add("cmd4") // This should cause cmd1 to be removed
	h.Add("cmd5") // This should cause cmd2 to be removed

	err := h.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load and verify only last 3 commands
	h2 := New(3)
	err = h2.Load(histFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if h2.Len() != 3 {
		t.Errorf("Len() = %d, want 3", h2.Len())
	}

	entries := h2.All()
	expected := []string{"cmd3", "cmd4", "cmd5"}
	for i, cmd := range expected {
		if entries[i].Command != cmd {
			t.Errorf("entries[%d].Command = %q, want %q", i, entries[i].Command, cmd)
		}
	}
}

func TestHistoryFileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "format_history")

	h := New(100)
	h.Add("simple command")
	h.Add("command with spaces")

	err := h.Save(histFile)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read raw file content
	content, err := os.ReadFile(histFile)
	if err != nil {
		t.Fatalf("Reading file: %v", err)
	}

	// File should be readable text
	if len(content) == 0 {
		t.Error("File is empty")
	}
}
