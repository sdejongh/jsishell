// Package history provides command history functionality for the shell.
package history

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrOutOfBounds is returned when accessing an invalid history index.
var ErrOutOfBounds = errors.New("history index out of bounds")

// ErrEmptyHistory is returned when accessing an empty history.
var ErrEmptyHistory = errors.New("history is empty")

// HistoryEntry represents a single command in the history.
type HistoryEntry struct {
	Command   string    // The command text
	Timestamp time.Time // When the command was executed
}

// History manages command history with navigation and search capabilities.
type History struct {
	entries     []HistoryEntry
	maxSize     int
	navPosition int    // Current position during navigation (-1 = at end/current line)
	currentLine string // The line user was typing before navigating
	mu          sync.RWMutex

	// Search state
	searchQuery   string
	searchResults []int // Indices of matching entries
	searchPos     int   // Position in search results

	// Options
	ignoreDuplicates  bool
	ignoreSpacePrefix bool
}

// New creates a new History with the specified maximum size.
func New(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 1000 // Default
	}
	return &History{
		entries:     make([]HistoryEntry, 0, maxSize),
		maxSize:     maxSize,
		navPosition: -1,
	}
}

// SetIgnoreDuplicates sets whether to ignore consecutive duplicate commands.
func (h *History) SetIgnoreDuplicates(ignore bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ignoreDuplicates = ignore
}

// SetIgnoreSpacePrefix sets whether to ignore commands starting with space.
func (h *History) SetIgnoreSpacePrefix(ignore bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ignoreSpacePrefix = ignore
}

// Add adds a command to the history.
func (h *History) Add(command string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Skip empty commands
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// Skip commands starting with space if configured
	if h.ignoreSpacePrefix && strings.HasPrefix(command, " ") {
		return
	}

	// Skip consecutive duplicates if configured
	if h.ignoreDuplicates && len(h.entries) > 0 {
		if h.entries[len(h.entries)-1].Command == command {
			return
		}
	}

	entry := HistoryEntry{
		Command:   command,
		Timestamp: time.Now(),
	}

	h.entries = append(h.entries, entry)

	// Trim if exceeds max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}

	// Reset navigation
	h.navPosition = -1
	h.currentLine = ""
}

// Get returns the entry at the specified index.
func (h *History) Get(index int) (HistoryEntry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if index < 0 || index >= len(h.entries) {
		return HistoryEntry{}, ErrOutOfBounds
	}
	return h.entries[index], nil
}

// Len returns the number of entries in the history.
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// MaxSize returns the maximum history size.
func (h *History) MaxSize() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxSize
}

// All returns all history entries.
func (h *History) All() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]HistoryEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// Last returns the most recent entry.
func (h *History) Last() (HistoryEntry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.entries) == 0 {
		return HistoryEntry{}, ErrEmptyHistory
	}
	return h.entries[len(h.entries)-1], nil
}

// Clear removes all entries from the history.
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = h.entries[:0]
	h.navPosition = -1
	h.currentLine = ""
}

// Navigation methods

// SetCurrentLine sets the current line (what user is typing) for navigation.
func (h *History) SetCurrentLine(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentLine = line
}

// CurrentLine returns the current line.
func (h *History) CurrentLine() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentLine
}

// ResetNavigation resets navigation to the end (current line position).
func (h *History) ResetNavigation() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.navPosition = -1
}

// Previous returns the previous (older) command in history.
// Returns false if at the beginning of history.
func (h *History) Previous() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 {
		return "", false
	}

	if h.navPosition == -1 {
		// Start from the end
		h.navPosition = len(h.entries) - 1
	} else if h.navPosition > 0 {
		h.navPosition--
	} else {
		// At the beginning
		return "", false
	}

	return h.entries[h.navPosition].Command, true
}

// Next returns the next (newer) command in history.
// Returns false if at the end (current line position).
func (h *History) Next() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.navPosition == -1 {
		// Already at current line
		return "", false
	}

	if h.navPosition < len(h.entries)-1 {
		h.navPosition++
		return h.entries[h.navPosition].Command, true
	}

	// Back to current line
	h.navPosition = -1
	return "", false
}

// Search methods

// Search returns all entries containing the query (case-insensitive).
// Results are in reverse chronological order (most recent first).
func (h *History) Search(query string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	query = strings.ToLower(query)
	var results []HistoryEntry

	// Iterate in reverse for most recent first
	for i := len(h.entries) - 1; i >= 0; i-- {
		if query == "" || strings.Contains(strings.ToLower(h.entries[i].Command), query) {
			results = append(results, h.entries[i])
		}
	}

	return results
}

// SearchPrefix returns all entries starting with the prefix (case-insensitive).
// Results are in reverse chronological order.
func (h *History) SearchPrefix(prefix string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	prefix = strings.ToLower(prefix)
	var results []HistoryEntry

	for i := len(h.entries) - 1; i >= 0; i-- {
		if strings.HasPrefix(strings.ToLower(h.entries[i].Command), prefix) {
			results = append(results, h.entries[i])
		}
	}

	return results
}

// StartSearch initiates an interactive search with the given query.
func (h *History) StartSearch(query string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.searchQuery = strings.ToLower(query)
	h.searchResults = nil
	h.searchPos = -1

	// Build list of matching indices (in reverse order for most recent first)
	for i := len(h.entries) - 1; i >= 0; i-- {
		if h.searchQuery == "" || strings.Contains(strings.ToLower(h.entries[i].Command), h.searchQuery) {
			h.searchResults = append(h.searchResults, i)
		}
	}
}

// PreviousMatch returns the previous matching command during search.
func (h *History) PreviousMatch() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.searchResults) == 0 {
		return "", false
	}

	if h.searchPos < len(h.searchResults)-1 {
		h.searchPos++
		idx := h.searchResults[h.searchPos]
		return h.entries[idx].Command, true
	}

	return "", false
}

// NextMatch returns the next matching command during search.
func (h *History) NextMatch() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.searchResults) == 0 || h.searchPos <= 0 {
		return "", false
	}

	h.searchPos--
	idx := h.searchResults[h.searchPos]
	return h.entries[idx].Command, true
}

// EndSearch ends the interactive search.
func (h *History) EndSearch() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.searchQuery = ""
	h.searchResults = nil
	h.searchPos = -1
}

// Persistence methods

// Save saves the history to a file.
func (h *History) Save(path string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating history directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write entries in format: timestamp:command
	for _, entry := range h.entries {
		ts := entry.Timestamp.Unix()
		// Escape newlines in command
		cmd := strings.ReplaceAll(entry.Command, "\n", "\\n")
		_, err := fmt.Fprintf(writer, "%d:%s\n", ts, cmd)
		if err != nil {
			return fmt.Errorf("writing history entry: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flushing history file: %w", err)
	}

	return nil
}

// Load loads history from a file.
func (h *History) Load(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No history file yet
		}
		return fmt.Errorf("opening history file: %w", err)
	}
	defer file.Close()

	h.entries = h.entries[:0] // Clear existing entries

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse format: timestamp:command
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			// Legacy format without timestamp - just the command
			h.entries = append(h.entries, HistoryEntry{
				Command:   line,
				Timestamp: time.Now(),
			})
			continue
		}

		tsStr := line[:colonIdx]
		cmd := line[colonIdx+1:]

		// Unescape newlines
		cmd = strings.ReplaceAll(cmd, "\\n", "\n")

		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			// Invalid timestamp, use current time
			ts = time.Now().Unix()
		}

		h.entries = append(h.entries, HistoryEntry{
			Command:   cmd,
			Timestamp: time.Unix(ts, 0),
		})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	// Trim to max size if necessary
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}

	return nil
}
