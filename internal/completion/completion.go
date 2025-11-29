// Package completion provides autocompletion functionality for the shell.
package completion

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CompletionType indicates what kind of completion candidate this is.
type CompletionType int

const (
	// TypeCommand is a shell command completion.
	TypeCommand CompletionType = iota
	// TypeFile is a regular file completion.
	TypeFile
	// TypeDirectory is a directory completion.
	TypeDirectory
	// TypeExecutable is an executable file completion.
	TypeExecutable
	// TypeVariable is an environment variable completion.
	TypeVariable
	// TypeOption is a command option completion.
	TypeOption
)

// CompletionCandidate represents a single completion suggestion.
type CompletionCandidate struct {
	Text        string         // The completion text
	Type        CompletionType // The type of completion
	Description string         // Optional description
}

// OptionDef defines a command option for completion.
type OptionDef struct {
	Long        string // Long form (e.g., "--verbose")
	Short       string // Short form (e.g., "-v")
	Description string // Help text
}

// CommandDef defines a command with its options for completion.
type CommandDef struct {
	Name        string      // Command name
	Description string      // Short description
	Options     []OptionDef // Available options
}

// Completer provides command and path completion functionality.
type Completer struct {
	commands    []string              // Known command names (for backward compatibility)
	commandDefs map[string]CommandDef // Command definitions with options
}

// NewCompleter creates a new Completer with the given command list.
func NewCompleter(commands []string) *Completer {
	return &Completer{
		commands:    commands,
		commandDefs: make(map[string]CommandDef),
	}
}

// NewCompleterWithDefs creates a new Completer with command definitions.
func NewCompleterWithDefs(defs []CommandDef) *Completer {
	commands := make([]string, 0, len(defs))
	commandDefs := make(map[string]CommandDef, len(defs))

	for _, def := range defs {
		commands = append(commands, def.Name)
		commandDefs[def.Name] = def
	}

	return &Completer{
		commands:    commands,
		commandDefs: commandDefs,
	}
}

// SetCommands updates the list of known commands.
func (c *Completer) SetCommands(commands []string) {
	c.commands = commands
}

// SetCommandDefs updates the command definitions.
func (c *Completer) SetCommandDefs(defs []CommandDef) {
	c.commands = make([]string, 0, len(defs))
	c.commandDefs = make(map[string]CommandDef, len(defs))

	for _, def := range defs {
		c.commands = append(c.commands, def.Name)
		c.commandDefs[def.Name] = def
	}
}

// CompleteCommand returns completion candidates for a command name prefix.
func (c *Completer) CompleteCommand(prefix string) []CompletionCandidate {
	var candidates []CompletionCandidate
	lowerPrefix := strings.ToLower(prefix)

	for _, cmd := range c.commands {
		if strings.HasPrefix(strings.ToLower(cmd), lowerPrefix) {
			candidates = append(candidates, CompletionCandidate{
				Text: cmd,
				Type: TypeCommand,
			})
		}
	}

	// Sort by name
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Text < candidates[j].Text
	})

	return candidates
}

// CompletePath returns completion candidates for a file path prefix.
func (c *Completer) CompletePath(pathPrefix string) []CompletionCandidate {
	var candidates []CompletionCandidate

	// Handle empty path
	if pathPrefix == "" {
		pathPrefix = "."
	}

	// Preserve the original prefix style for constructing results
	originalPrefix := pathPrefix

	// Track if we need to use ~ in results
	hasTildePrefix := strings.HasPrefix(pathPrefix, "~")
	var homeDir string

	// Expand ~ to home directory
	if hasTildePrefix {
		var err error
		homeDir, err = os.UserHomeDir()
		if err == nil {
			if pathPrefix == "~" {
				pathPrefix = homeDir
			} else if strings.HasPrefix(pathPrefix, "~/") {
				pathPrefix = filepath.Join(homeDir, pathPrefix[2:])
			} else {
				// ~username style - not supported, treat as literal
				hasTildePrefix = false
			}
		} else {
			hasTildePrefix = false
		}
	}

	// Determine the directory to search and the prefix to match
	dir := filepath.Dir(pathPrefix)
	base := filepath.Base(pathPrefix)

	// Handle "./" and "../" prefixes - filepath.Dir("./foo") returns ".", filepath.Base returns "foo"
	// We need to preserve these prefixes in results
	hasExplicitDot := strings.HasPrefix(originalPrefix, "./") || originalPrefix == "."
	hasExplicitDotDot := strings.HasPrefix(originalPrefix, "../") || originalPrefix == ".."

	// If path ends with separator or is "." or "..", we're listing a directory
	if strings.HasSuffix(pathPrefix, string(os.PathSeparator)) || pathPrefix == "." || pathPrefix == ".." {
		dir = strings.TrimSuffix(pathPrefix, string(os.PathSeparator))
		base = ""
	}

	// Special case: "~" or "~/" should list home directory
	if hasTildePrefix && (originalPrefix == "~" || originalPrefix == "~/") {
		dir = homeDir
		base = ""
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return candidates
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless prefix starts with .
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}

		// Check if entry matches prefix
		if base != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(base)) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		var compType CompletionType
		isDir := entry.IsDir()
		if isDir {
			compType = TypeDirectory
		} else if info.Mode()&0111 != 0 {
			compType = TypeExecutable
		} else {
			compType = TypeFile
		}

		// Build the completion text
		var completionText string
		if dir == "." && !hasExplicitDot {
			// For relative paths without explicit "./", just use the name
			completionText = name
		} else if dir == "." && hasExplicitDot {
			// For "./" prefix, preserve it in the result
			completionText = "./" + name
		} else if (dir == ".." || strings.HasPrefix(dir, "../")) && hasExplicitDotDot {
			// For "../" prefix, preserve it in the result
			completionText = dir + "/" + name
		} else if hasTildePrefix && homeDir != "" {
			// For ~ prefix, reconstruct with ~ instead of absolute home path
			fullPath := filepath.Join(dir, name)
			if strings.HasPrefix(fullPath, homeDir) {
				completionText = "~" + fullPath[len(homeDir):]
			} else {
				completionText = fullPath
			}
		} else {
			// Use filepath.Join for other paths
			completionText = filepath.Join(dir, name)
		}

		// Add trailing slash for directories
		if isDir {
			completionText = completionText + string(os.PathSeparator)
		}

		candidates = append(candidates, CompletionCandidate{
			Text: completionText,
			Type: compType,
		})
	}

	// Sort by name
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Text < candidates[j].Text
	})

	return candidates
}

// CompleteOption returns completion candidates for command options.
func (c *Completer) CompleteOption(commandName, optionPrefix string) []CompletionCandidate {
	var candidates []CompletionCandidate

	// Look up command definition
	def, ok := c.commandDefs[commandName]
	if !ok {
		return candidates
	}

	for _, opt := range def.Options {
		// Match long options (--verbose)
		if opt.Long != "" && strings.HasPrefix(opt.Long, optionPrefix) {
			candidates = append(candidates, CompletionCandidate{
				Text:        opt.Long,
				Type:        TypeOption,
				Description: opt.Description,
			})
		}
		// Match short options (-v)
		if opt.Short != "" && strings.HasPrefix(opt.Short, optionPrefix) {
			candidates = append(candidates, CompletionCandidate{
				Text:        opt.Short,
				Type:        TypeOption,
				Description: opt.Description,
			})
		}
	}

	// Sort by text
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Text < candidates[j].Text
	})

	return candidates
}

// Complete returns completion candidates based on the current input.
// It determines whether to complete a command, option, or path based on context.
func (c *Completer) Complete(input string) []CompletionCandidate {
	input = strings.TrimLeft(input, " \t")

	// If input has no space, it could be a command or a path
	if !strings.Contains(input, " ") {
		// If it looks like a path (starts with /, ./, ../, or ~), complete as path
		if isPathLike(input) {
			return c.CompletePath(input)
		}
		return c.CompleteCommand(input)
	}

	// Parse the input to find command and current word
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return c.CompleteCommand("")
	}

	commandName := parts[0]

	// Get the last word
	lastWord := ""
	lastSpaceIdx := strings.LastIndex(input, " ")
	if lastSpaceIdx >= 0 && lastSpaceIdx < len(input)-1 {
		lastWord = input[lastSpaceIdx+1:]
	}

	// If last word starts with "-", complete as option
	if strings.HasPrefix(lastWord, "-") {
		return c.CompleteOption(commandName, lastWord)
	}

	// If we're right after a space and no word yet, check if we should suggest options
	// This happens when input ends with a space
	if lastWord == "" && strings.HasSuffix(input, " ") {
		// Could suggest both options and paths here
		// For now, prioritize path completion for usability
		return c.CompletePath(lastWord)
	}

	// Otherwise, complete as path
	return c.CompletePath(lastWord)
}

// InlineSuggestion returns the suggested completion text to show inline (ghost text).
// Returns the text to append and whether there is a suggestion.
func (c *Completer) InlineSuggestion(input string) (string, bool) {
	if input == "" {
		return "", false
	}

	trimmedInput := strings.TrimLeft(input, " \t")

	// Determine if we're completing a command or a path
	// Path completion happens when there's a space (argument) or when input looks like a path
	isPathCompletion := strings.Contains(trimmedInput, " ") || isPathLike(trimmedInput)

	candidates := c.Complete(input)
	if len(candidates) == 0 {
		return "", false
	}

	// For path completion, we need to compare against the last word only
	compareWith := trimmedInput
	if strings.Contains(trimmedInput, " ") {
		lastSpaceIdx := strings.LastIndex(input, " ")
		if lastSpaceIdx >= 0 {
			compareWith = input[lastSpaceIdx+1:]
		}
	} else if isPathCompletion {
		// Path at start of line - compareWith is already the full trimmedInput
		compareWith = trimmedInput
	}

	// If only one candidate, return the completion suffix
	if len(candidates) == 1 {
		candidate := candidates[0].Text
		if strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(compareWith)) {
			suffix := candidate[len(compareWith):]
			if suffix == "" {
				return "", false // Exact match, nothing to suggest
			}
			return suffix, true
		}
	}

	// Find common prefix among all candidates
	common := findCommonPrefix(candidates)
	if len(common) > len(compareWith) && strings.HasPrefix(strings.ToLower(common), strings.ToLower(compareWith)) {
		return common[len(compareWith):], true
	}

	return "", false
}

// AcceptSuggestion returns the input with the inline suggestion applied.
func (c *Completer) AcceptSuggestion(input string) string {
	suggestion, has := c.InlineSuggestion(input)
	if !has {
		return input
	}
	return input + suggestion
}

// GetCompletionList returns a list of all possible completions as strings.
// This is used for Tab-Tab display.
func (c *Completer) GetCompletionList(input string) []string {
	candidates := c.Complete(input)
	result := make([]string, len(candidates))
	for i, cand := range candidates {
		result[i] = cand.Text
	}
	return result
}

// findCommonPrefix finds the longest common prefix among all candidates.
func findCommonPrefix(candidates []CompletionCandidate) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0].Text
	}

	prefix := candidates[0].Text
	for _, cand := range candidates[1:] {
		prefix = commonPrefix(prefix, cand.Text)
		if prefix == "" {
			break
		}
	}
	return prefix
}

// commonPrefix returns the common prefix of two strings.
func commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if strings.ToLower(string(a[i])) != strings.ToLower(string(b[i])) {
			return a[:i]
		}
	}
	return a[:minLen]
}

// isPathLike returns true if the input looks like a path rather than a command name.
// This includes absolute paths (/...), relative paths (./..., ../...), and home paths (~...).
func isPathLike(input string) bool {
	if input == "" {
		return false
	}
	return strings.HasPrefix(input, "/") ||
		strings.HasPrefix(input, "./") ||
		strings.HasPrefix(input, "../") ||
		input == "." ||
		input == ".." ||
		strings.HasPrefix(input, "~")
}
