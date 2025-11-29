// Package parser parses lexer tokens into a command AST.
package parser

// Command represents a parsed shell command.
type Command struct {
	Name         string              // Command name (may be abbreviated)
	Resolved     string              // Resolved full command name (set by executor)
	Args         []string            // Positional arguments
	Options      map[string]string   // Named options (--key=value or --key value)
	MultiOptions map[string][]string // Options that can be specified multiple times
	Flags        map[string]bool     // Boolean flags (--verbose, -v)
	RawInput     string              // Original input string
}

// NewCommand creates a new empty Command.
func NewCommand() *Command {
	return &Command{
		Args:         make([]string, 0),
		Options:      make(map[string]string),
		MultiOptions: make(map[string][]string),
		Flags:        make(map[string]bool),
	}
}

// HasFlag returns true if the flag is set.
func (c *Command) HasFlag(names ...string) bool {
	for _, name := range names {
		if c.Flags[name] {
			return true
		}
	}
	return false
}

// GetOption returns the value of an option, or empty string if not set.
func (c *Command) GetOption(names ...string) string {
	for _, name := range names {
		if val, ok := c.Options[name]; ok {
			return val
		}
	}
	return ""
}

// GetOptionOr returns the value of an option, or the default if not set.
func (c *Command) GetOptionOr(defaultVal string, names ...string) string {
	for _, name := range names {
		if val, ok := c.Options[name]; ok {
			return val
		}
	}
	return defaultVal
}

// GetOptions returns all values for an option that can be specified multiple times.
// Returns nil if the option was not specified.
func (c *Command) GetOptions(names ...string) []string {
	var result []string
	for _, name := range names {
		if vals, ok := c.MultiOptions[name]; ok {
			result = append(result, vals...)
		}
	}
	return result
}

// ArgCount returns the number of positional arguments.
func (c *Command) ArgCount() int {
	return len(c.Args)
}

// Arg returns the argument at index i, or empty string if out of bounds.
func (c *Command) Arg(i int) string {
	if i < 0 || i >= len(c.Args) {
		return ""
	}
	return c.Args[i]
}

// AllArgs returns all arguments including flags and options.
// This is useful for passing to external commands.
// Short flags are combined (e.g., -a -l becomes -al).
// The order is: combined short flags, long flags, options with values, positional args.
func (c *Command) AllArgs() []string {
	var result []string

	// Separate short and long flags
	var shortFlags []rune
	for flag := range c.Flags {
		if len(flag) == 2 && flag[0] == '-' && flag[1] != '-' {
			// Short flag like -a
			shortFlags = append(shortFlags, rune(flag[1]))
		} else {
			// Long flag like --verbose
			result = append(result, flag)
		}
	}

	// Combine short flags into one argument (e.g., -al)
	if len(shortFlags) > 0 {
		result = append(result, "-"+string(shortFlags))
	}

	// Add options with values (e.g., --output=file, --count 5)
	for opt, val := range c.Options {
		if val != "" {
			result = append(result, opt+"="+val)
		} else {
			result = append(result, opt)
		}
	}

	// Add positional arguments
	result = append(result, c.Args...)

	return result
}
