// Package terminal handles low-level terminal I/O operations.
// It provides raw mode input, key reading, and terminal detection.
package terminal

import (
	"io"
	"os"

	"golang.org/x/term"
)

// KeyType identifies special keys.
type KeyType int

const (
	KeyNone KeyType = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyDelete
	KeyBackspace
	KeyTab
	KeyEnter
	KeyEscape
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlK
	KeyCtrlL
	KeyCtrlN
	KeyCtrlP
	KeyCtrlU
	KeyCtrlW
	KeyCtrlR     // Reverse search
	KeyCtrlLeft  // Word navigation
	KeyCtrlRight // Word navigation
)

// Key represents a keyboard input.
type Key struct {
	Rune    rune    // Character (0 for special keys)
	Special KeyType // Special key type
	Alt     bool    // Alt modifier
	Ctrl    bool    // Ctrl modifier
}

// Terminal handles low-level terminal I/O.
type Terminal struct {
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	fd       int // File descriptor for stdin
	oldState *term.State
	inRaw    bool
}

// New creates a new Terminal with standard I/O.
func New() *Terminal {
	return &Terminal{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		fd:     int(os.Stdin.Fd()),
	}
}

// NewWithIO creates a Terminal with custom I/O streams.
// This is useful for testing.
func NewWithIO(stdin io.Reader, stdout, stderr io.Writer, fd int) *Terminal {
	return &Terminal{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		fd:     fd,
	}
}

// EnterRawMode switches the terminal to raw mode for key-by-key input.
// Returns a restore function that must be called to restore normal mode.
func (t *Terminal) EnterRawMode() (restore func(), err error) {
	if !t.IsTerminal() {
		return func() {}, nil // No-op for non-terminals
	}

	oldState, err := term.MakeRaw(t.fd)
	if err != nil {
		return nil, err
	}

	t.oldState = oldState
	t.inRaw = true

	return func() {
		if t.oldState != nil {
			term.Restore(t.fd, t.oldState)
			t.inRaw = false
			t.oldState = nil
		}
	}, nil
}

// IsRawMode returns true if the terminal is in raw mode.
func (t *Terminal) IsRawMode() bool {
	return t.inRaw
}

// ReadKey reads a single key or escape sequence from the terminal.
func (t *Terminal) ReadKey() (Key, error) {
	buf := make([]byte, 6) // Large enough for escape sequences
	n, err := t.stdin.Read(buf[:1])
	if err != nil {
		return Key{}, err
	}
	if n == 0 {
		return Key{}, io.EOF
	}

	b := buf[0]

	// Handle escape sequences FIRST (before other control chars)
	if b == 27 { // ESC
		return t.handleEscapeSequence(buf)
	}

	// Handle control characters
	if b < 32 {
		return t.handleControlChar(b), nil
	}

	// Handle DEL (127)
	if b == 127 {
		return Key{Special: KeyBackspace}, nil
	}

	// Regular character
	return Key{Rune: rune(b)}, nil
}

// handleControlChar converts control characters to Key.
func (t *Terminal) handleControlChar(b byte) Key {
	switch b {
	case 1: // Ctrl+A
		return Key{Special: KeyCtrlA, Ctrl: true}
	case 2: // Ctrl+B
		return Key{Special: KeyCtrlB, Ctrl: true}
	case 3: // Ctrl+C
		return Key{Special: KeyCtrlC, Ctrl: true}
	case 4: // Ctrl+D
		return Key{Special: KeyCtrlD, Ctrl: true}
	case 5: // Ctrl+E
		return Key{Special: KeyCtrlE, Ctrl: true}
	case 6: // Ctrl+F
		return Key{Special: KeyCtrlF, Ctrl: true}
	case 9: // Tab
		return Key{Special: KeyTab}
	case 10, 13: // Enter (LF or CR)
		return Key{Special: KeyEnter}
	case 11: // Ctrl+K
		return Key{Special: KeyCtrlK, Ctrl: true}
	case 12: // Ctrl+L
		return Key{Special: KeyCtrlL, Ctrl: true}
	case 14: // Ctrl+N
		return Key{Special: KeyCtrlN, Ctrl: true}
	case 16: // Ctrl+P
		return Key{Special: KeyCtrlP, Ctrl: true}
	case 18: // Ctrl+R
		return Key{Special: KeyCtrlR, Ctrl: true}
	case 21: // Ctrl+U
		return Key{Special: KeyCtrlU, Ctrl: true}
	case 23: // Ctrl+W
		return Key{Special: KeyCtrlW, Ctrl: true}
	case 27: // Escape
		return Key{Special: KeyEscape}
	default:
		// Unknown control character, return as-is with Ctrl flag
		return Key{Rune: rune(b + 64), Ctrl: true}
	}
}

// handleEscapeSequence handles escape sequences (arrow keys, etc.)
func (t *Terminal) handleEscapeSequence(buf []byte) (Key, error) {
	// Read second byte
	n, err := t.stdin.Read(buf[1:2])
	if err != nil || n == 0 {
		// Just ESC pressed
		return Key{Special: KeyEscape}, nil
	}

	// Alt+key: ESC followed by a printable character (but not '[' or 'O')
	if buf[1] >= 32 && buf[1] < 127 && buf[1] != '[' && buf[1] != 'O' {
		return Key{Rune: rune(buf[1]), Alt: true}, nil
	}

	// CSI sequence: ESC [ ... or SS3 sequence: ESC O ...
	if buf[1] == '[' || buf[1] == 'O' {
		// Read the next character
		n, err = t.stdin.Read(buf[2:3])
		if err != nil || n == 0 {
			return Key{Special: KeyEscape}, nil
		}

		// Handle SS3 sequences (ESC O ...)
		if buf[1] == 'O' {
			switch buf[2] {
			case 'A':
				return Key{Special: KeyUp}, nil
			case 'B':
				return Key{Special: KeyDown}, nil
			case 'C':
				return Key{Special: KeyRight}, nil
			case 'D':
				return Key{Special: KeyLeft}, nil
			case 'H':
				return Key{Special: KeyHome}, nil
			case 'F':
				return Key{Special: KeyEnd}, nil
			}
			return Key{Special: KeyEscape}, nil
		}

		// Handle CSI sequences (ESC [ ...)
		switch buf[2] {
		case 'A':
			return Key{Special: KeyUp}, nil
		case 'B':
			return Key{Special: KeyDown}, nil
		case 'C':
			return Key{Special: KeyRight}, nil
		case 'D':
			return Key{Special: KeyLeft}, nil
		case 'H':
			return Key{Special: KeyHome}, nil
		case 'F':
			return Key{Special: KeyEnd}, nil
		case '1':
			// Could be Home (ESC [ 1 ~) or modifier sequences (ESC [ 1 ; ...)
			n, _ = t.stdin.Read(buf[3:4])
			if n > 0 {
				if buf[3] == '~' {
					return Key{Special: KeyHome}, nil
				}
				if buf[3] == ';' {
					// Modifier sequence: ESC [ 1 ; mod X
					// Read modifier and key
					t.stdin.Read(buf[4:5]) // modifier (2=Shift, 3=Alt, 5=Ctrl, 7=Ctrl+Alt)
					t.stdin.Read(buf[5:6]) // key
					modifier := buf[4]
					switch buf[5] {
					case 'A':
						return Key{Special: KeyUp}, nil
					case 'B':
						return Key{Special: KeyDown}, nil
					case 'C':
						// Ctrl+Right for word navigation
						if modifier == '5' {
							return Key{Special: KeyCtrlRight, Ctrl: true}, nil
						}
						return Key{Special: KeyRight}, nil
					case 'D':
						// Ctrl+Left for word navigation
						if modifier == '5' {
							return Key{Special: KeyCtrlLeft, Ctrl: true}, nil
						}
						return Key{Special: KeyLeft}, nil
					case 'H':
						return Key{Special: KeyHome}, nil
					case 'F':
						return Key{Special: KeyEnd}, nil
					}
				}
			}
		case '3':
			// Delete (ESC [ 3 ~)
			n, _ = t.stdin.Read(buf[3:4])
			if n > 0 && buf[3] == '~' {
				return Key{Special: KeyDelete}, nil
			}
		case '4':
			// End (ESC [ 4 ~)
			n, _ = t.stdin.Read(buf[3:4])
			if n > 0 && buf[3] == '~' {
				return Key{Special: KeyEnd}, nil
			}
		case '5':
			// Page Up (ESC [ 5 ~) - ignore
			t.stdin.Read(buf[3:4])
		case '6':
			// Page Down (ESC [ 6 ~) - ignore
			t.stdin.Read(buf[3:4])
		case '7':
			// Home (ESC [ 7 ~) - rxvt
			n, _ = t.stdin.Read(buf[3:4])
			if n > 0 && buf[3] == '~' {
				return Key{Special: KeyHome}, nil
			}
		case '8':
			// End (ESC [ 8 ~) - rxvt
			n, _ = t.stdin.Read(buf[3:4])
			if n > 0 && buf[3] == '~' {
				return Key{Special: KeyEnd}, nil
			}
		}
	}

	// Unknown escape sequence
	return Key{Special: KeyEscape}, nil
}

// Write writes bytes to the terminal stdout.
func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.stdout.Write(p)
}

// WriteString writes a string to the terminal stdout.
func (t *Terminal) WriteString(s string) (n int, err error) {
	return io.WriteString(t.stdout, s)
}

// WriteError writes to stderr.
func (t *Terminal) WriteError(p []byte) (n int, err error) {
	return t.stderr.Write(p)
}

// Size returns the terminal dimensions (columns, rows).
func (t *Terminal) Size() (width, height int, err error) {
	return term.GetSize(t.fd)
}

// IsTerminal returns true if stdin is connected to a terminal.
func (t *Terminal) IsTerminal() bool {
	return term.IsTerminal(t.fd)
}

// Clear clears the screen.
func (t *Terminal) Clear() error {
	_, err := t.WriteString("\033[2J\033[H")
	return err
}

// ClearLine clears the current line.
func (t *Terminal) ClearLine() error {
	_, err := t.WriteString("\033[2K\r")
	return err
}

// ClearToEnd clears from cursor to end of line.
func (t *Terminal) ClearToEnd() error {
	_, err := t.WriteString("\033[K")
	return err
}

// MoveCursor moves the cursor to the specified position (1-indexed).
func (t *Terminal) MoveCursor(row, col int) error {
	_, err := t.WriteString("\033[" + itoa(row) + ";" + itoa(col) + "H")
	return err
}

// MoveCursorLeft moves the cursor left by n positions.
func (t *Terminal) MoveCursorLeft(n int) error {
	if n <= 0 {
		return nil
	}
	_, err := t.WriteString("\033[" + itoa(n) + "D")
	return err
}

// MoveCursorRight moves the cursor right by n positions.
func (t *Terminal) MoveCursorRight(n int) error {
	if n <= 0 {
		return nil
	}
	_, err := t.WriteString("\033[" + itoa(n) + "C")
	return err
}

// SaveCursor saves the current cursor position.
func (t *Terminal) SaveCursor() error {
	_, err := t.WriteString("\033[s")
	return err
}

// RestoreCursor restores the saved cursor position.
func (t *Terminal) RestoreCursor() error {
	_, err := t.WriteString("\033[u")
	return err
}

// HideCursor hides the cursor.
func (t *Terminal) HideCursor() error {
	_, err := t.WriteString("\033[?25l")
	return err
}

// ShowCursor shows the cursor.
func (t *Terminal) ShowCursor() error {
	_, err := t.WriteString("\033[?25h")
	return err
}

// Bell produces a terminal bell sound.
func (t *Terminal) Bell() error {
	_, err := t.WriteString("\a")
	return err
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
