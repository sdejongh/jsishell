// Package terminal provides the LineEditor for interactive line editing.
package terminal

import (
	"strings"
	"time"
	"unicode"
)

// CompletionProvider provides completion suggestions.
type CompletionProvider interface {
	InlineSuggestion(input string) (string, bool)
	GetCompletionList(input string) []string
	AcceptSuggestion(input string) string
}

// HistoryProvider provides command history navigation.
type HistoryProvider interface {
	SetCurrentLine(line string)
	CurrentLine() string
	ResetNavigation()
	Previous() (string, bool)
	Next() (string, bool)
	// Search methods
	StartSearch(query string)
	PreviousMatch() (string, bool)
	NextMatch() (string, bool)
	EndSearch()
}

// LineEditor handles interactive line input with cursor movement and editing.
type LineEditor struct {
	buffer    []rune             // Current input buffer
	cursor    int                // Cursor position in buffer (0-indexed)
	prompt    string             // Prompt string
	ghostText string             // Inline completion suggestion (dimmed)
	terminal  *Terminal          // Terminal for I/O
	completer CompletionProvider // Completion provider
	history   HistoryProvider    // History provider
	lastTab   time.Time          // Time of last Tab press for Tab-Tab detection
	colors    *ColorScheme       // Color scheme for ghost text

	// Search mode state
	searchMode   bool   // Whether we're in search mode (Ctrl+R)
	searchQuery  string // Current search query
	searchPrompt string // Search prompt (e.g., "(reverse-i-search)`query':")
}

// NewLineEditor creates a new LineEditor.
func NewLineEditor(term *Terminal) *LineEditor {
	return &LineEditor{
		buffer:   make([]rune, 0),
		cursor:   0,
		terminal: term,
	}
}

// SetBuffer sets the buffer content.
func (e *LineEditor) SetBuffer(s string) {
	e.buffer = []rune(s)
	if e.cursor > len(e.buffer) {
		e.cursor = len(e.buffer)
	}
}

// SetCursor sets the cursor position.
func (e *LineEditor) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(e.buffer) {
		pos = len(e.buffer)
	}
	e.cursor = pos
}

// Cursor returns the current cursor position.
func (e *LineEditor) Cursor() int {
	return e.cursor
}

// String returns the buffer as a string.
func (e *LineEditor) String() string {
	return string(e.buffer)
}

// Len returns the length of the buffer in runes.
func (e *LineEditor) Len() int {
	return len(e.buffer)
}

// Prompt returns the current prompt.
func (e *LineEditor) Prompt() string {
	return e.prompt
}

// SetPrompt sets the prompt string.
func (e *LineEditor) SetPrompt(prompt string) {
	e.prompt = prompt
}

// GhostText returns the current ghost text suggestion.
func (e *LineEditor) GhostText() string {
	return e.ghostText
}

// SetGhostText sets the ghost text suggestion.
func (e *LineEditor) SetGhostText(text string) {
	e.ghostText = text
}

// SetCompleter sets the completion provider.
func (e *LineEditor) SetCompleter(c CompletionProvider) {
	e.completer = c
}

// SetHistory sets the history provider.
func (e *LineEditor) SetHistory(h HistoryProvider) {
	e.history = h
}

// SetColors sets the color scheme for ghost text.
func (e *LineEditor) SetColors(colors *ColorScheme) {
	e.colors = colors
}

// Clear clears the buffer and resets cursor.
func (e *LineEditor) Clear() {
	e.buffer = e.buffer[:0]
	e.cursor = 0
	e.ghostText = ""
}

// ============================================================================
// Cursor Movement (T056)
// ============================================================================

// MoveLeft moves the cursor one position to the left.
func (e *LineEditor) MoveLeft() {
	if e.cursor > 0 {
		e.cursor--
	}
}

// MoveRight moves the cursor one position to the right.
func (e *LineEditor) MoveRight() {
	if e.cursor < len(e.buffer) {
		e.cursor++
	}
}

// MoveToStart moves the cursor to the start of the line (Home/Ctrl+A).
func (e *LineEditor) MoveToStart() {
	e.cursor = 0
}

// MoveToEnd moves the cursor to the end of the line (End/Ctrl+E).
func (e *LineEditor) MoveToEnd() {
	e.cursor = len(e.buffer)
}

// ============================================================================
// Character Insertion/Deletion (T057, T058)
// ============================================================================

// Insert inserts a character at the cursor position.
func (e *LineEditor) Insert(r rune) {
	// Make room for the new character
	e.buffer = append(e.buffer, 0)
	// Shift characters after cursor to the right
	copy(e.buffer[e.cursor+1:], e.buffer[e.cursor:])
	// Insert the character
	e.buffer[e.cursor] = r
	e.cursor++
	// Clear ghost text on any input
	e.ghostText = ""
}

// InsertString inserts a string at the cursor position.
func (e *LineEditor) InsertString(s string) {
	for _, r := range s {
		e.Insert(r)
	}
}

// Backspace deletes the character before the cursor.
func (e *LineEditor) Backspace() {
	if e.cursor > 0 {
		// Remove character at cursor-1
		copy(e.buffer[e.cursor-1:], e.buffer[e.cursor:])
		e.buffer = e.buffer[:len(e.buffer)-1]
		e.cursor--
	}
}

// Delete deletes the character at the cursor position.
func (e *LineEditor) Delete() {
	if e.cursor < len(e.buffer) {
		// Remove character at cursor
		copy(e.buffer[e.cursor:], e.buffer[e.cursor+1:])
		e.buffer = e.buffer[:len(e.buffer)-1]
	}
}

// DeleteToEnd deletes from cursor to end of line (Ctrl+K).
func (e *LineEditor) DeleteToEnd() {
	e.buffer = e.buffer[:e.cursor]
}

// DeleteToStart deletes from start of line to cursor (Ctrl+U).
func (e *LineEditor) DeleteToStart() {
	e.buffer = e.buffer[e.cursor:]
	e.cursor = 0
}

// ============================================================================
// Word Navigation (T059)
// ============================================================================

// MoveWordLeft moves the cursor to the start of the previous word (Ctrl+Left/Alt+B).
func (e *LineEditor) MoveWordLeft() {
	if e.cursor == 0 {
		return
	}

	// Skip any spaces before cursor
	for e.cursor > 0 && unicode.IsSpace(e.buffer[e.cursor-1]) {
		e.cursor--
	}

	// Move to start of word
	for e.cursor > 0 && !unicode.IsSpace(e.buffer[e.cursor-1]) {
		e.cursor--
	}
}

// MoveWordRight moves the cursor to the start of the next word (Ctrl+Right/Alt+F).
func (e *LineEditor) MoveWordRight() {
	if e.cursor >= len(e.buffer) {
		return
	}

	// Skip current word
	for e.cursor < len(e.buffer) && !unicode.IsSpace(e.buffer[e.cursor]) {
		e.cursor++
	}

	// Skip spaces to reach the next word
	for e.cursor < len(e.buffer) && unicode.IsSpace(e.buffer[e.cursor]) {
		e.cursor++
	}
}

// ============================================================================
// Word Deletion (T060)
// ============================================================================

// DeleteWordBackward deletes the word before the cursor (Ctrl+W).
func (e *LineEditor) DeleteWordBackward() {
	if e.cursor == 0 {
		return
	}

	// Find start of word to delete
	end := e.cursor

	// Skip spaces before cursor
	for e.cursor > 0 && unicode.IsSpace(e.buffer[e.cursor-1]) {
		e.cursor--
	}

	// Skip the word
	for e.cursor > 0 && !unicode.IsSpace(e.buffer[e.cursor-1]) {
		e.cursor--
	}

	// Delete from cursor to end
	copy(e.buffer[e.cursor:], e.buffer[end:])
	e.buffer = e.buffer[:len(e.buffer)-(end-e.cursor)]
}

// DeleteWordForward deletes the word after the cursor (Alt+D).
func (e *LineEditor) DeleteWordForward() {
	if e.cursor >= len(e.buffer) {
		return
	}

	start := e.cursor

	// Skip current word only (don't include trailing spaces)
	end := e.cursor
	for end < len(e.buffer) && !unicode.IsSpace(e.buffer[end]) {
		end++
	}

	// Delete from start to end
	copy(e.buffer[start:], e.buffer[end:])
	e.buffer = e.buffer[:len(e.buffer)-(end-start)]
}

// ============================================================================
// History Navigation (T111)
// ============================================================================

// historyPrevious navigates to the previous (older) command in history.
func (e *LineEditor) historyPrevious() {
	if e.history == nil {
		return
	}

	// Save current line before navigating
	if e.history.CurrentLine() == "" && len(e.buffer) > 0 {
		e.history.SetCurrentLine(string(e.buffer))
	}

	cmd, ok := e.history.Previous()
	if ok {
		e.buffer = []rune(cmd)
		e.cursor = len(e.buffer)
		e.ghostText = ""
	}
}

// historyNext navigates to the next (newer) command in history.
func (e *LineEditor) historyNext() {
	if e.history == nil {
		return
	}

	cmd, ok := e.history.Next()
	if ok {
		e.buffer = []rune(cmd)
		e.cursor = len(e.buffer)
		e.ghostText = ""
	} else {
		// Back to current line
		current := e.history.CurrentLine()
		e.buffer = []rune(current)
		e.cursor = len(e.buffer)
		e.history.SetCurrentLine("")
		e.ghostText = ""
	}
}

// ============================================================================
// History Search (T112)
// ============================================================================

// startHistorySearch enters reverse search mode (Ctrl+R).
func (e *LineEditor) startHistorySearch() {
	if e.history == nil {
		return
	}

	e.searchMode = true
	e.searchQuery = ""
	e.searchPrompt = "(reverse-i-search)`':"
	e.history.StartSearch("")
}

// exitHistorySearch exits search mode.
func (e *LineEditor) exitHistorySearch() {
	e.searchMode = false
	e.searchQuery = ""
	e.searchPrompt = ""
	if e.history != nil {
		e.history.EndSearch()
	}
}

// handleSearchKey handles a key while in search mode.
// Returns true if the line is complete (Enter pressed).
func (e *LineEditor) handleSearchKey(key Key) bool {
	switch key.Special {
	case KeyEnter:
		// Accept current result and exit search
		e.exitHistorySearch()
		return true

	case KeyCtrlR:
		// Search for next match
		if e.history != nil {
			cmd, ok := e.history.PreviousMatch()
			if ok {
				e.buffer = []rune(cmd)
				e.cursor = len(e.buffer)
			}
		}

	case KeyCtrlC, KeyEscape:
		// Cancel search
		e.exitHistorySearch()
		e.Clear()
		if e.terminal != nil {
			e.terminal.WriteString("^C\r\n")
		}

	case KeyBackspace:
		// Remove last char from search query
		if len(e.searchQuery) > 0 {
			e.searchQuery = e.searchQuery[:len(e.searchQuery)-1]
			e.searchPrompt = "(reverse-i-search)`" + e.searchQuery + "':"
			if e.history != nil {
				e.history.StartSearch(e.searchQuery)
				// Get first match
				cmd, ok := e.history.PreviousMatch()
				if ok {
					e.buffer = []rune(cmd)
					e.cursor = len(e.buffer)
				}
			}
		}

	case KeyLeft, KeyRight, KeyHome, KeyEnd:
		// Exit search mode and apply the key
		e.exitHistorySearch()
		return e.HandleKey(key)

	case KeyNone:
		// Add character to search query
		if key.Rune != 0 && !key.Ctrl && !key.Alt {
			e.searchQuery += string(key.Rune)
			e.searchPrompt = "(reverse-i-search)`" + e.searchQuery + "':"
			if e.history != nil {
				e.history.StartSearch(e.searchQuery)
				// Get first match
				cmd, ok := e.history.PreviousMatch()
				if ok {
					e.buffer = []rune(cmd)
					e.cursor = len(e.buffer)
				}
			}
		}
	}

	return false
}

// renderSearch renders the search prompt and buffer.
func (e *LineEditor) renderSearch() {
	if e.terminal == nil {
		return
	}

	// Move to start of line and clear it
	e.terminal.WriteString("\r")
	e.terminal.WriteString("\033[K")

	// Write search prompt
	e.terminal.WriteString(e.searchPrompt)

	// Write buffer content
	e.terminal.WriteString(" " + string(e.buffer))
}

// ============================================================================
// Line Rendering (T061)
// ============================================================================

// Render renders the current line to the terminal.
func (e *LineEditor) Render() {
	if e.terminal == nil {
		return
	}

	// Move to start of line and clear it
	e.terminal.WriteString("\r")     // Move to column 0
	e.terminal.WriteString("\033[K") // Clear from cursor to end of line

	// Write prompt
	e.terminal.WriteString(e.prompt)

	// Write buffer content
	e.terminal.WriteString(string(e.buffer))

	// Write ghost text if any (using color scheme or default dim)
	if e.ghostText != "" {
		if e.colors != nil {
			e.terminal.WriteString(e.colors.GhostText(e.ghostText))
		} else {
			e.terminal.WriteString("\033[2m") // Dim
			e.terminal.WriteString(e.ghostText)
			e.terminal.WriteString("\033[0m") // Reset
		}
	}

	// Move cursor to correct position
	// Calculate how far back we need to move from the end
	moveBack := len(e.buffer) - e.cursor + len([]rune(e.ghostText))
	if moveBack > 0 {
		e.terminal.MoveCursorLeft(moveBack)
	}
}

// RenderNewLine renders a newline (after command submission).
func (e *LineEditor) RenderNewLine() {
	if e.terminal != nil {
		e.terminal.WriteString("\r\n")
	}
}

// updateGhostText updates the ghost text from the completer.
func (e *LineEditor) updateGhostText() {
	if e.completer == nil {
		e.ghostText = ""
		return
	}

	input := string(e.buffer)
	// Only show ghost text if cursor is at the end
	if e.cursor != len(e.buffer) {
		e.ghostText = ""
		return
	}

	suggestion, has := e.completer.InlineSuggestion(input)
	if has {
		e.ghostText = suggestion
	} else {
		e.ghostText = ""
	}
}

// handleTab handles the Tab key for completion.
func (e *LineEditor) handleTab() {
	now := time.Now()
	tabTabThreshold := 300 * time.Millisecond

	// Check for Tab-Tab (double tab within threshold)
	if !e.lastTab.IsZero() && now.Sub(e.lastTab) < tabTabThreshold {
		// Tab-Tab: show all completions
		e.showCompletionList()
		e.lastTab = time.Time{} // Reset to prevent triple-tab issues
		return
	}

	e.lastTab = now

	// Single Tab: accept ghost text if any
	if e.ghostText != "" {
		e.InsertString(e.ghostText)
		e.ghostText = ""
	} else if e.completer != nil {
		// Try to complete from completer
		input := string(e.buffer)
		completed := e.completer.AcceptSuggestion(input)
		if completed != input {
			e.buffer = []rune(completed)
			e.cursor = len(e.buffer)
			e.updateGhostText()
		}
	}
}

// showCompletionList displays all completion options.
func (e *LineEditor) showCompletionList() {
	if e.completer == nil || e.terminal == nil {
		return
	}

	input := string(e.buffer)
	completions := e.completer.GetCompletionList(input)

	if len(completions) == 0 {
		return
	}

	// Print newline, then all completions
	e.terminal.WriteString("\r\n")

	// Display completions in columns
	e.renderCompletions(completions)

	// Re-render the prompt and buffer
	e.Render()
}

// renderCompletions renders a list of completions in columns.
func (e *LineEditor) renderCompletions(completions []string) {
	if len(completions) == 0 {
		return
	}

	// Get terminal width
	width, _, _ := e.terminal.Size()
	if width == 0 {
		width = 80 // Default
	}

	// Find max completion length
	maxLen := 0
	for _, c := range completions {
		if len(c) > maxLen {
			maxLen = len(c)
		}
	}

	// Add padding
	colWidth := maxLen + 2
	numCols := width / colWidth
	if numCols < 1 {
		numCols = 1
	}

	// Render completions in columns
	for i, c := range completions {
		// Truncate to just the base name for paths
		display := c
		if idx := strings.LastIndex(c, "/"); idx >= 0 && idx < len(c)-1 {
			display = c[idx+1:]
		}

		e.terminal.WriteString(display)

		// Add padding or newline
		if (i+1)%numCols == 0 || i == len(completions)-1 {
			e.terminal.WriteString("\r\n")
		} else {
			padding := colWidth - len(display)
			for j := 0; j < padding; j++ {
				e.terminal.WriteString(" ")
			}
		}
	}
}

// ============================================================================
// Key Handling
// ============================================================================

// HandleKey processes a key input and returns true if the line is complete (Enter pressed).
func (e *LineEditor) HandleKey(key Key) bool {
	switch key.Special {
	case KeyEnter:
		return true

	case KeyBackspace:
		e.Backspace()

	case KeyDelete:
		e.Delete()

	case KeyLeft:
		e.MoveLeft()

	case KeyRight:
		e.MoveRight()

	case KeyUp:
		e.historyPrevious()

	case KeyDown:
		e.historyNext()

	case KeyHome:
		e.MoveToStart()

	case KeyEnd:
		e.MoveToEnd()

	case KeyCtrlA:
		e.MoveToStart()

	case KeyCtrlE:
		e.MoveToEnd()

	case KeyCtrlB:
		e.MoveLeft()

	case KeyCtrlF:
		e.MoveRight()

	case KeyCtrlK:
		e.DeleteToEnd()

	case KeyCtrlU:
		e.DeleteToStart()

	case KeyCtrlW:
		e.DeleteWordBackward()

	case KeyCtrlLeft:
		e.MoveWordLeft()

	case KeyCtrlRight:
		e.MoveWordRight()

	case KeyCtrlC:
		// Cancel - clear buffer and signal newline
		e.Clear()
		if e.terminal != nil {
			e.terminal.WriteString("^C\r\n")
		}
		return false // Don't return the empty line, continue editing

	case KeyCtrlD:
		// EOF if buffer is empty
		if len(e.buffer) == 0 {
			return true
		}
		// Otherwise delete char at cursor
		e.Delete()

	case KeyCtrlL:
		// Clear screen and redraw
		if e.terminal != nil {
			e.terminal.Clear()
			e.Render()
		}

	case KeyTab:
		e.handleTab()

	case KeyCtrlR:
		e.startHistorySearch()

	case KeyCtrlN:
		// Same as Down arrow
		e.historyNext()

	case KeyCtrlP:
		// Same as Up arrow
		e.historyPrevious()

	case KeyEscape:
		// Ignore escape key alone

	case KeyNone:
		// Regular character or Alt+key
		if key.Rune != 0 {
			if key.Alt {
				// Alt+key combinations
				switch key.Rune {
				case 'b', 'B':
					e.MoveWordLeft()
				case 'f', 'F':
					e.MoveWordRight()
				case 'd', 'D':
					e.DeleteWordForward()
				}
			} else if !key.Ctrl {
				e.Insert(key.Rune)
			}
		}

	default:
		// Handle regular character if present
		if key.Rune != 0 && !key.Ctrl && !key.Alt {
			e.Insert(key.Rune)
		}
	}

	return false
}

// ReadLine reads a line of input interactively.
// Returns the line content and any error.
func (e *LineEditor) ReadLine() (string, error) {
	if e.terminal == nil {
		return "", nil
	}

	// Enter raw mode
	restore, err := e.terminal.EnterRawMode()
	if err != nil {
		return "", err
	}
	defer restore()

	// Clear buffer for new input
	e.Clear()

	// Reset history navigation
	if e.history != nil {
		e.history.ResetNavigation()
	}

	// Initial ghost text update
	e.updateGhostText()

	// Initial render
	e.Render()

	for {
		key, err := e.terminal.ReadKey()
		if err != nil {
			return "", err
		}

		var done bool

		// Handle search mode separately
		if e.searchMode {
			done = e.handleSearchKey(key)
			if done {
				e.RenderNewLine()
				return e.String(), nil
			}
			e.renderSearch()
			continue
		}

		done = e.HandleKey(key)

		if done {
			e.RenderNewLine()
			return e.String(), nil
		}

		// Update ghost text after each key (unless it was Tab which handles it)
		if key.Special != KeyTab {
			e.updateGhostText()
		}

		e.Render()
	}
}
