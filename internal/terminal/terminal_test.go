package terminal

import (
	"bytes"
	"io"
	"testing"
)

// mockReader allows simulating terminal input.
type mockReader struct {
	data []byte
	pos  int
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func TestKeyType(t *testing.T) {
	// Verify key type constants are unique
	keyTypes := []KeyType{
		KeyNone, KeyUp, KeyDown, KeyLeft, KeyRight,
		KeyHome, KeyEnd, KeyDelete, KeyBackspace,
		KeyTab, KeyEnter, KeyEscape,
		KeyCtrlA, KeyCtrlB, KeyCtrlC, KeyCtrlD,
		KeyCtrlE, KeyCtrlF, KeyCtrlK, KeyCtrlL,
		KeyCtrlN, KeyCtrlP, KeyCtrlU, KeyCtrlW,
	}

	seen := make(map[KeyType]bool)
	for _, kt := range keyTypes {
		if seen[kt] {
			t.Errorf("Duplicate KeyType value: %d", kt)
		}
		seen[kt] = true
	}
}

func TestTerminalWrite(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdin := &mockReader{}

	term := NewWithIO(stdin, &stdout, &stderr, -1)

	// Test Write
	n, err := term.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}
	if stdout.String() != "hello" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "hello")
	}

	// Test WriteString
	stdout.Reset()
	n, err = term.WriteString("world")
	if err != nil {
		t.Errorf("WriteString error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteString returned %d, want 5", n)
	}
	if stdout.String() != "world" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "world")
	}

	// Test WriteError
	n, err = term.WriteError([]byte("error"))
	if err != nil {
		t.Errorf("WriteError error: %v", err)
	}
	if stderr.String() != "error" {
		t.Errorf("stderr = %q, want %q", stderr.String(), "error")
	}
}

func TestReadKeyRegularChar(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Key
	}{
		{"letter a", []byte{'a'}, Key{Rune: 'a'}},
		{"letter Z", []byte{'Z'}, Key{Rune: 'Z'}},
		{"digit 5", []byte{'5'}, Key{Rune: '5'}},
		{"space", []byte{' '}, Key{Rune: ' '}},
		{"tilde", []byte{'~'}, Key{Rune: '~'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := &mockReader{data: tt.input}
			term := NewWithIO(stdin, &stdout, &stdout, -1)

			key, err := term.ReadKey()
			if err != nil {
				t.Errorf("ReadKey error: %v", err)
			}
			if key.Rune != tt.want.Rune {
				t.Errorf("ReadKey().Rune = %q, want %q", key.Rune, tt.want.Rune)
			}
			if key.Special != tt.want.Special {
				t.Errorf("ReadKey().Special = %d, want %d", key.Special, tt.want.Special)
			}
		})
	}
}

func TestReadKeyControlChars(t *testing.T) {
	tests := []struct {
		name  string
		input byte
		want  Key
	}{
		{"Ctrl+A", 1, Key{Special: KeyCtrlA, Ctrl: true}},
		{"Ctrl+C", 3, Key{Special: KeyCtrlC, Ctrl: true}},
		{"Ctrl+D", 4, Key{Special: KeyCtrlD, Ctrl: true}},
		{"Tab", 9, Key{Special: KeyTab}},
		{"Enter (LF)", 10, Key{Special: KeyEnter}},
		{"Enter (CR)", 13, Key{Special: KeyEnter}},
		{"Ctrl+K", 11, Key{Special: KeyCtrlK, Ctrl: true}},
		{"Ctrl+U", 21, Key{Special: KeyCtrlU, Ctrl: true}},
		{"Ctrl+W", 23, Key{Special: KeyCtrlW, Ctrl: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := &mockReader{data: []byte{tt.input}}
			term := NewWithIO(stdin, &stdout, &stdout, -1)

			key, err := term.ReadKey()
			if err != nil {
				t.Errorf("ReadKey error: %v", err)
			}
			if key.Special != tt.want.Special {
				t.Errorf("ReadKey().Special = %d, want %d", key.Special, tt.want.Special)
			}
			if key.Ctrl != tt.want.Ctrl {
				t.Errorf("ReadKey().Ctrl = %v, want %v", key.Ctrl, tt.want.Ctrl)
			}
		})
	}
}

func TestReadKeyBackspace(t *testing.T) {
	var stdout bytes.Buffer
	stdin := &mockReader{data: []byte{127}} // DEL
	term := NewWithIO(stdin, &stdout, &stdout, -1)

	key, err := term.ReadKey()
	if err != nil {
		t.Errorf("ReadKey error: %v", err)
	}
	if key.Special != KeyBackspace {
		t.Errorf("ReadKey().Special = %d, want %d (KeyBackspace)", key.Special, KeyBackspace)
	}
}

func TestReadKeyArrows(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  KeyType
	}{
		{"Up", []byte{27, '[', 'A'}, KeyUp},
		{"Down", []byte{27, '[', 'B'}, KeyDown},
		{"Right", []byte{27, '[', 'C'}, KeyRight},
		{"Left", []byte{27, '[', 'D'}, KeyLeft},
		{"Home", []byte{27, '[', 'H'}, KeyHome},
		{"End", []byte{27, '[', 'F'}, KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := &mockReader{data: tt.input}
			term := NewWithIO(stdin, &stdout, &stdout, -1)

			key, err := term.ReadKey()
			if err != nil {
				t.Errorf("ReadKey error: %v", err)
			}
			if key.Special != tt.want {
				t.Errorf("ReadKey().Special = %d, want %d", key.Special, tt.want)
			}
		})
	}
}

func TestReadKeyDelete(t *testing.T) {
	var stdout bytes.Buffer
	stdin := &mockReader{data: []byte{27, '[', '3', '~'}}
	term := NewWithIO(stdin, &stdout, &stdout, -1)

	key, err := term.ReadKey()
	if err != nil {
		t.Errorf("ReadKey error: %v", err)
	}
	if key.Special != KeyDelete {
		t.Errorf("ReadKey().Special = %d, want %d (KeyDelete)", key.Special, KeyDelete)
	}
}

func TestReadKeyAltModifier(t *testing.T) {
	var stdout bytes.Buffer
	stdin := &mockReader{data: []byte{27, 'a'}} // Alt+a
	term := NewWithIO(stdin, &stdout, &stdout, -1)

	key, err := term.ReadKey()
	if err != nil {
		t.Errorf("ReadKey error: %v", err)
	}
	if key.Rune != 'a' {
		t.Errorf("ReadKey().Rune = %q, want 'a'", key.Rune)
	}
	if !key.Alt {
		t.Error("ReadKey().Alt = false, want true")
	}
}

func TestReadKeyCtrlArrows(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  KeyType
	}{
		// Ctrl+Right: ESC [ 1 ; 5 C
		{"Ctrl+Right", []byte{27, '[', '1', ';', '5', 'C'}, KeyCtrlRight},
		// Ctrl+Left: ESC [ 1 ; 5 D
		{"Ctrl+Left", []byte{27, '[', '1', ';', '5', 'D'}, KeyCtrlLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := &mockReader{data: tt.input}
			term := NewWithIO(stdin, &stdout, &stdout, -1)

			key, err := term.ReadKey()
			if err != nil {
				t.Errorf("ReadKey error: %v", err)
			}
			if key.Special != tt.want {
				t.Errorf("ReadKey().Special = %d, want %d", key.Special, tt.want)
			}
			if !key.Ctrl {
				t.Errorf("ReadKey().Ctrl = false, want true")
			}
		})
	}
}

func TestClear(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	err := term.Clear()
	if err != nil {
		t.Errorf("Clear error: %v", err)
	}
	// Clear screen sequence: ESC[2J ESC[H
	expected := "\033[2J\033[H"
	if stdout.String() != expected {
		t.Errorf("Clear wrote %q, want %q", stdout.String(), expected)
	}
}

func TestClearLine(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	err := term.ClearLine()
	if err != nil {
		t.Errorf("ClearLine error: %v", err)
	}
	// Clear line sequence: ESC[2K CR
	expected := "\033[2K\r"
	if stdout.String() != expected {
		t.Errorf("ClearLine wrote %q, want %q", stdout.String(), expected)
	}
}

func TestClearToEnd(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	err := term.ClearToEnd()
	if err != nil {
		t.Errorf("ClearToEnd error: %v", err)
	}
	expected := "\033[K"
	if stdout.String() != expected {
		t.Errorf("ClearToEnd wrote %q, want %q", stdout.String(), expected)
	}
}

func TestMoveCursor(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	err := term.MoveCursor(5, 10)
	if err != nil {
		t.Errorf("MoveCursor error: %v", err)
	}
	expected := "\033[5;10H"
	if stdout.String() != expected {
		t.Errorf("MoveCursor wrote %q, want %q", stdout.String(), expected)
	}
}

func TestMoveCursorLeftRight(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Terminal, int) error
		n        int
		expected string
	}{
		{"left 3", (*Terminal).MoveCursorLeft, 3, "\033[3D"},
		{"right 5", (*Terminal).MoveCursorRight, 5, "\033[5C"},
		{"left 0", (*Terminal).MoveCursorLeft, 0, ""},
		{"right 0", (*Terminal).MoveCursorRight, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

			err := tt.method(term, tt.n)
			if err != nil {
				t.Errorf("error: %v", err)
			}
			if stdout.String() != tt.expected {
				t.Errorf("wrote %q, want %q", stdout.String(), tt.expected)
			}
		})
	}
}

func TestCursorSaveRestore(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	term.SaveCursor()
	term.RestoreCursor()

	expected := "\033[s\033[u"
	if stdout.String() != expected {
		t.Errorf("Save/Restore wrote %q, want %q", stdout.String(), expected)
	}
}

func TestCursorVisibility(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	term.HideCursor()
	term.ShowCursor()

	expected := "\033[?25l\033[?25h"
	if stdout.String() != expected {
		t.Errorf("Hide/Show wrote %q, want %q", stdout.String(), expected)
	}
}

func TestBell(t *testing.T) {
	var stdout bytes.Buffer
	term := NewWithIO(&mockReader{}, &stdout, &stdout, -1)

	err := term.Bell()
	if err != nil {
		t.Errorf("Bell error: %v", err)
	}
	if stdout.String() != "\a" {
		t.Errorf("Bell wrote %q, want %q", stdout.String(), "\a")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
		{-1, "-1"},
		{-42, "-42"},
	}

	for _, tt := range tests {
		got := itoa(tt.n)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestIsRawMode(t *testing.T) {
	term := NewWithIO(&mockReader{}, &bytes.Buffer{}, &bytes.Buffer{}, -1)

	// Initially not in raw mode
	if term.IsRawMode() {
		t.Error("IsRawMode() = true, want false initially")
	}
}

func TestIsTerminalNonTTY(t *testing.T) {
	// With fd=-1, this should not be a terminal
	term := NewWithIO(&mockReader{}, &bytes.Buffer{}, &bytes.Buffer{}, -1)

	if term.IsTerminal() {
		t.Error("IsTerminal() = true for fd=-1, want false")
	}
}
