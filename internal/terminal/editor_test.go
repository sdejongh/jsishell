// Package terminal provides tests for the LineEditor.
package terminal

import (
	"bytes"
	"testing"
)

// T052: Write tests for cursor movement
func TestLineEditorCursorMovement(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     int
		action     func(*LineEditor)
		wantCursor int
		wantBuffer string
	}{
		{
			name:       "move left from middle",
			initial:    "hello",
			cursor:     3,
			action:     func(e *LineEditor) { e.MoveLeft() },
			wantCursor: 2,
			wantBuffer: "hello",
		},
		{
			name:       "move left from start (no change)",
			initial:    "hello",
			cursor:     0,
			action:     func(e *LineEditor) { e.MoveLeft() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "move right from middle",
			initial:    "hello",
			cursor:     2,
			action:     func(e *LineEditor) { e.MoveRight() },
			wantCursor: 3,
			wantBuffer: "hello",
		},
		{
			name:       "move right from end (no change)",
			initial:    "hello",
			cursor:     5,
			action:     func(e *LineEditor) { e.MoveRight() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "move to start",
			initial:    "hello",
			cursor:     3,
			action:     func(e *LineEditor) { e.MoveToStart() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "move to end",
			initial:    "hello",
			cursor:     2,
			action:     func(e *LineEditor) { e.MoveToEnd() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "move to start already at start",
			initial:    "hello",
			cursor:     0,
			action:     func(e *LineEditor) { e.MoveToStart() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "move to end already at end",
			initial:    "hello",
			cursor:     5,
			action:     func(e *LineEditor) { e.MoveToEnd() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLineEditor(nil)
			e.SetBuffer(tt.initial)
			e.SetCursor(tt.cursor)

			tt.action(e)

			if e.Cursor() != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", e.Cursor(), tt.wantCursor)
			}
			if e.String() != tt.wantBuffer {
				t.Errorf("buffer = %q, want %q", e.String(), tt.wantBuffer)
			}
		})
	}
}

// T053: Write tests for character insertion/deletion
func TestLineEditorInsertionDeletion(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     int
		action     func(*LineEditor)
		wantCursor int
		wantBuffer string
	}{
		{
			name:       "insert at start",
			initial:    "ello",
			cursor:     0,
			action:     func(e *LineEditor) { e.Insert('h') },
			wantCursor: 1,
			wantBuffer: "hello",
		},
		{
			name:       "insert at end",
			initial:    "hell",
			cursor:     4,
			action:     func(e *LineEditor) { e.Insert('o') },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "insert in middle",
			initial:    "helo",
			cursor:     3,
			action:     func(e *LineEditor) { e.Insert('l') },
			wantCursor: 4,
			wantBuffer: "hello",
		},
		{
			name:       "insert into empty",
			initial:    "",
			cursor:     0,
			action:     func(e *LineEditor) { e.Insert('a') },
			wantCursor: 1,
			wantBuffer: "a",
		},
		{
			name:       "backspace from middle",
			initial:    "helllo",
			cursor:     4,
			action:     func(e *LineEditor) { e.Backspace() },
			wantCursor: 3,
			wantBuffer: "hello",
		},
		{
			name:       "backspace from start (no change)",
			initial:    "hello",
			cursor:     0,
			action:     func(e *LineEditor) { e.Backspace() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "backspace from end",
			initial:    "helloo",
			cursor:     6,
			action:     func(e *LineEditor) { e.Backspace() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "delete at start",
			initial:    "xhello",
			cursor:     0,
			action:     func(e *LineEditor) { e.Delete() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "delete at end (no change)",
			initial:    "hello",
			cursor:     5,
			action:     func(e *LineEditor) { e.Delete() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "delete in middle",
			initial:    "helxlo",
			cursor:     3,
			action:     func(e *LineEditor) { e.Delete() },
			wantCursor: 3,
			wantBuffer: "hello",
		},
		{
			name:       "delete to end (Ctrl+K)",
			initial:    "hello world",
			cursor:     5,
			action:     func(e *LineEditor) { e.DeleteToEnd() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "delete to start (Ctrl+U)",
			initial:    "hello world",
			cursor:     6,
			action:     func(e *LineEditor) { e.DeleteToStart() },
			wantCursor: 0,
			wantBuffer: "world",
		},
		{
			name:       "delete to end at end (no change)",
			initial:    "hello",
			cursor:     5,
			action:     func(e *LineEditor) { e.DeleteToEnd() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
		{
			name:       "delete to start at start (no change)",
			initial:    "hello",
			cursor:     0,
			action:     func(e *LineEditor) { e.DeleteToStart() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLineEditor(nil)
			e.SetBuffer(tt.initial)
			e.SetCursor(tt.cursor)

			tt.action(e)

			if e.Cursor() != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", e.Cursor(), tt.wantCursor)
			}
			if e.String() != tt.wantBuffer {
				t.Errorf("buffer = %q, want %q", e.String(), tt.wantBuffer)
			}
		})
	}
}

// T054: Write tests for word navigation
func TestLineEditorWordNavigation(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     int
		action     func(*LineEditor)
		wantCursor int
	}{
		{
			name:       "move word left from middle of word",
			initial:    "hello world",
			cursor:     8, // in "world"
			action:     func(e *LineEditor) { e.MoveWordLeft() },
			wantCursor: 6, // start of "world"
		},
		{
			name:       "move word left from start of word",
			initial:    "hello world",
			cursor:     6, // start of "world"
			action:     func(e *LineEditor) { e.MoveWordLeft() },
			wantCursor: 0, // start of "hello"
		},
		{
			name:       "move word left from start (no change)",
			initial:    "hello world",
			cursor:     0,
			action:     func(e *LineEditor) { e.MoveWordLeft() },
			wantCursor: 0,
		},
		{
			name:       "move word right from start",
			initial:    "hello world",
			cursor:     0,
			action:     func(e *LineEditor) { e.MoveWordRight() },
			wantCursor: 6, // start of "world" (skips word and spaces)
		},
		{
			name:       "move word right from middle",
			initial:    "hello world",
			cursor:     6, // start of "world"
			action:     func(e *LineEditor) { e.MoveWordRight() },
			wantCursor: 11, // end of "world"
		},
		{
			name:       "move word right from end (no change)",
			initial:    "hello world",
			cursor:     11,
			action:     func(e *LineEditor) { e.MoveWordRight() },
			wantCursor: 11,
		},
		{
			name:       "move word left with multiple spaces",
			initial:    "hello   world",
			cursor:     13, // end
			action:     func(e *LineEditor) { e.MoveWordLeft() },
			wantCursor: 8, // start of "world"
		},
		{
			name:       "move word right with multiple spaces",
			initial:    "hello   world",
			cursor:     5, // end of "hello"
			action:     func(e *LineEditor) { e.MoveWordRight() },
			wantCursor: 8, // start of "world" (skips spaces)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLineEditor(nil)
			e.SetBuffer(tt.initial)
			e.SetCursor(tt.cursor)

			tt.action(e)

			if e.Cursor() != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", e.Cursor(), tt.wantCursor)
			}
		})
	}
}

// Test word deletion (Ctrl+W, Alt+D)
func TestLineEditorWordDeletion(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     int
		action     func(*LineEditor)
		wantCursor int
		wantBuffer string
	}{
		{
			name:       "delete word backward (Ctrl+W)",
			initial:    "hello world",
			cursor:     11, // end
			action:     func(e *LineEditor) { e.DeleteWordBackward() },
			wantCursor: 6,
			wantBuffer: "hello ",
		},
		{
			name:       "delete word backward from middle",
			initial:    "hello world test",
			cursor:     11, // end of "world"
			action:     func(e *LineEditor) { e.DeleteWordBackward() },
			wantCursor: 6,
			wantBuffer: "hello  test",
		},
		{
			name:       "delete word backward at start (no change)",
			initial:    "hello",
			cursor:     0,
			action:     func(e *LineEditor) { e.DeleteWordBackward() },
			wantCursor: 0,
			wantBuffer: "hello",
		},
		{
			name:       "delete word forward (Alt+D)",
			initial:    "hello world",
			cursor:     0,
			action:     func(e *LineEditor) { e.DeleteWordForward() },
			wantCursor: 0,
			wantBuffer: " world", // deletes "hello", keeps space
		},
		{
			name:       "delete word forward from middle",
			initial:    "hello world test",
			cursor:     6, // start of "world"
			action:     func(e *LineEditor) { e.DeleteWordForward() },
			wantCursor: 6,
			wantBuffer: "hello  test", // deletes "world", keeps spaces
		},
		{
			name:       "delete word forward at end (no change)",
			initial:    "hello",
			cursor:     5,
			action:     func(e *LineEditor) { e.DeleteWordForward() },
			wantCursor: 5,
			wantBuffer: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLineEditor(nil)
			e.SetBuffer(tt.initial)
			e.SetCursor(tt.cursor)

			tt.action(e)

			if e.Cursor() != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", e.Cursor(), tt.wantCursor)
			}
			if e.String() != tt.wantBuffer {
				t.Errorf("buffer = %q, want %q", e.String(), tt.wantBuffer)
			}
		})
	}
}

// Test clear buffer
func TestLineEditorClear(t *testing.T) {
	e := NewLineEditor(nil)
	e.SetBuffer("hello world")
	e.SetCursor(5)

	e.Clear()

	if e.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 after clear", e.Cursor())
	}
	if e.String() != "" {
		t.Errorf("buffer = %q, want empty after clear", e.String())
	}
	if e.Len() != 0 {
		t.Errorf("len = %d, want 0 after clear", e.Len())
	}
}

// Test insert string
func TestLineEditorInsertString(t *testing.T) {
	e := NewLineEditor(nil)
	e.SetBuffer("hello")
	e.SetCursor(5)

	e.InsertString(" world")

	if e.String() != "hello world" {
		t.Errorf("buffer = %q, want %q", e.String(), "hello world")
	}
	if e.Cursor() != 11 {
		t.Errorf("cursor = %d, want 11", e.Cursor())
	}
}

// Test buffer length
func TestLineEditorLen(t *testing.T) {
	e := NewLineEditor(nil)

	if e.Len() != 0 {
		t.Errorf("len = %d, want 0 for empty buffer", e.Len())
	}

	e.SetBuffer("hello")
	if e.Len() != 5 {
		t.Errorf("len = %d, want 5", e.Len())
	}

	e.SetBuffer("héllo") // Unicode
	if e.Len() != 5 {
		t.Errorf("len = %d, want 5 for unicode string", e.Len())
	}
}

// Test prompt
func TestLineEditorPrompt(t *testing.T) {
	e := NewLineEditor(nil)

	if e.Prompt() != "" {
		t.Errorf("prompt = %q, want empty by default", e.Prompt())
	}

	e.SetPrompt("$ ")
	if e.Prompt() != "$ " {
		t.Errorf("prompt = %q, want %q", e.Prompt(), "$ ")
	}
}

// Test Render output
func TestLineEditorRender(t *testing.T) {
	stdout := &bytes.Buffer{}
	term := NewWithIO(nil, stdout, nil, 0)
	e := NewLineEditor(term)
	e.SetPrompt("$ ")
	e.SetBuffer("hello")
	e.SetCursor(5)

	e.Render()

	output := stdout.String()
	// Should contain the prompt and buffer
	if len(output) == 0 {
		t.Error("render output should not be empty")
	}
}

// Test HandleKey for various key inputs
func TestLineEditorHandleKey(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     int
		key        Key
		wantCursor int
		wantBuffer string
		wantDone   bool
	}{
		{
			name:       "Enter key",
			initial:    "hello",
			cursor:     5,
			key:        Key{Special: KeyEnter},
			wantCursor: 5,
			wantBuffer: "hello",
			wantDone:   true,
		},
		{
			name:       "Backspace",
			initial:    "hello",
			cursor:     5,
			key:        Key{Special: KeyBackspace},
			wantCursor: 4,
			wantBuffer: "hell",
			wantDone:   false,
		},
		{
			name:       "Delete",
			initial:    "hello",
			cursor:     0,
			key:        Key{Special: KeyDelete},
			wantCursor: 0,
			wantBuffer: "ello",
			wantDone:   false,
		},
		{
			name:       "Left arrow",
			initial:    "hello",
			cursor:     3,
			key:        Key{Special: KeyLeft},
			wantCursor: 2,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Right arrow",
			initial:    "hello",
			cursor:     2,
			key:        Key{Special: KeyRight},
			wantCursor: 3,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Home key",
			initial:    "hello",
			cursor:     3,
			key:        Key{Special: KeyHome},
			wantCursor: 0,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "End key",
			initial:    "hello",
			cursor:     2,
			key:        Key{Special: KeyEnd},
			wantCursor: 5,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+A",
			initial:    "hello",
			cursor:     3,
			key:        Key{Special: KeyCtrlA, Ctrl: true},
			wantCursor: 0,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+E",
			initial:    "hello",
			cursor:     2,
			key:        Key{Special: KeyCtrlE, Ctrl: true},
			wantCursor: 5,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+B (move left)",
			initial:    "hello",
			cursor:     3,
			key:        Key{Special: KeyCtrlB, Ctrl: true},
			wantCursor: 2,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+F (move right)",
			initial:    "hello",
			cursor:     2,
			key:        Key{Special: KeyCtrlF, Ctrl: true},
			wantCursor: 3,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+K (delete to end)",
			initial:    "hello world",
			cursor:     5,
			key:        Key{Special: KeyCtrlK, Ctrl: true},
			wantCursor: 5,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Ctrl+U (delete to start)",
			initial:    "hello world",
			cursor:     6,
			key:        Key{Special: KeyCtrlU, Ctrl: true},
			wantCursor: 0,
			wantBuffer: "world",
			wantDone:   false,
		},
		{
			name:       "Ctrl+W (delete word backward)",
			initial:    "hello world",
			cursor:     11,
			key:        Key{Special: KeyCtrlW, Ctrl: true},
			wantCursor: 6,
			wantBuffer: "hello ",
			wantDone:   false,
		},
		{
			name:       "Ctrl+D on empty buffer",
			initial:    "",
			cursor:     0,
			key:        Key{Special: KeyCtrlD, Ctrl: true},
			wantCursor: 0,
			wantBuffer: "",
			wantDone:   true,
		},
		{
			name:       "Ctrl+D with content (delete char)",
			initial:    "hello",
			cursor:     0,
			key:        Key{Special: KeyCtrlD, Ctrl: true},
			wantCursor: 0,
			wantBuffer: "ello",
			wantDone:   false,
		},
		{
			name:       "Regular character",
			initial:    "hllo",
			cursor:     1,
			key:        Key{Rune: 'e', Special: KeyNone},
			wantCursor: 2,
			wantBuffer: "hello",
			wantDone:   false,
		},
		{
			name:       "Alt+b (word left)",
			initial:    "hello world",
			cursor:     11,
			key:        Key{Rune: 'b', Alt: true, Special: KeyNone},
			wantCursor: 6,
			wantBuffer: "hello world",
			wantDone:   false,
		},
		{
			name:       "Alt+f (word right)",
			initial:    "hello world",
			cursor:     0,
			key:        Key{Rune: 'f', Alt: true, Special: KeyNone},
			wantCursor: 6, // Moves to start of next word
			wantBuffer: "hello world",
			wantDone:   false,
		},
		{
			name:       "Alt+d (delete word forward)",
			initial:    "hello world",
			cursor:     0,
			key:        Key{Rune: 'd', Alt: true, Special: KeyNone},
			wantCursor: 0,
			wantBuffer: " world",
			wantDone:   false,
		},
		{
			name:       "Ctrl+Left (word left)",
			initial:    "hello world",
			cursor:     11, // end
			key:        Key{Special: KeyCtrlLeft, Ctrl: true},
			wantCursor: 6, // start of "world"
			wantBuffer: "hello world",
			wantDone:   false,
		},
		{
			name:       "Ctrl+Right (word right)",
			initial:    "hello world",
			cursor:     0,
			key:        Key{Special: KeyCtrlRight, Ctrl: true},
			wantCursor: 6, // start of "world"
			wantBuffer: "hello world",
			wantDone:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLineEditor(nil)
			e.SetBuffer(tt.initial)
			e.SetCursor(tt.cursor)

			done := e.HandleKey(tt.key)

			if done != tt.wantDone {
				t.Errorf("done = %v, want %v", done, tt.wantDone)
			}
			if e.Cursor() != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", e.Cursor(), tt.wantCursor)
			}
			if e.String() != tt.wantBuffer {
				t.Errorf("buffer = %q, want %q", e.String(), tt.wantBuffer)
			}
		})
	}
}

// ============================================================================
// T126: Performance benchmark for input latency (<10ms target)
// ============================================================================

// BenchmarkLineEditorInsert measures the time to insert a character.
func BenchmarkLineEditorInsert(b *testing.B) {
	e := NewLineEditor(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Insert('a')
	}
}

// BenchmarkLineEditorCursorMovement measures cursor movement latency.
func BenchmarkLineEditorCursorMovement(b *testing.B) {
	e := NewLineEditor(nil)
	e.SetBuffer("hello world this is a test line for benchmarking")
	e.SetCursor(len("hello world this is a test line for benchmarking") / 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.MoveLeft()
		e.MoveRight()
	}
}

// BenchmarkLineEditorBackspace measures backspace latency.
func BenchmarkLineEditorBackspace(b *testing.B) {
	e := NewLineEditor(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Insert then delete to keep buffer size stable
		e.Insert('x')
		e.Backspace()
	}
}

// BenchmarkLineEditorRender measures render latency.
func BenchmarkLineEditorRender(b *testing.B) {
	var output bytes.Buffer
	e := NewLineEditor(nil)
	e.SetPrompt("user@host:~/projects> ")
	e.SetBuffer("echo 'hello world' | grep hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
		output.Reset()
	}
}

// TestInputLatencyMeetsTarget verifies input processing is under 10ms.
func TestInputLatencyMeetsTarget(t *testing.T) {
	// Benchmark character insertion
	insertResult := testing.Benchmark(func(b *testing.B) {
		e := NewLineEditor(nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			e.Insert('a')
			if e.Len() > 100 {
				e.SetBuffer("")
				e.SetCursor(0)
			}
		}
	})

	// Calculate average time per operation in milliseconds
	insertMs := float64(insertResult.T.Nanoseconds()) / float64(insertResult.N) / 1e6

	// Target: input latency should be under 10ms
	if insertMs > 10 {
		t.Errorf("Input latency %.4fms exceeds 10ms target", insertMs)
	}
	t.Logf("Input latency: %.4fms (target: <10ms)", insertMs)
}
