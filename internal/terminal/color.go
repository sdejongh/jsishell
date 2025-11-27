package terminal

import (
	"os"
	"regexp"
	"strings"

	"github.com/sdejongh/jsishell/internal/config"
)

// ANSI color codes
const (
	ResetCode = "\033[0m"
)

// Color codes map
var colorCodes = map[string]string{
	"black":          "\033[30m",
	"red":            "\033[31m",
	"green":          "\033[32m",
	"yellow":         "\033[33m",
	"blue":           "\033[34m",
	"magenta":        "\033[35m",
	"cyan":           "\033[36m",
	"white":          "\033[37m",
	"bright_black":   "\033[90m",
	"bright_red":     "\033[91m",
	"bright_green":   "\033[92m",
	"bright_yellow":  "\033[93m",
	"bright_blue":    "\033[94m",
	"bright_magenta": "\033[95m",
	"bright_cyan":    "\033[96m",
	"bright_white":   "\033[97m",
}

// Regex to match ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\033\[[0-9;]*m`)

// ColorCode returns the ANSI code for a color name.
// Returns empty string for invalid color names.
func ColorCode(color string) string {
	return colorCodes[color]
}

// StripColors removes all ANSI color codes from a string.
func StripColors(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// ColorScheme manages terminal colors with configuration support.
type ColorScheme struct {
	enabled bool
	config  *config.ColorScheme
}

// NewColorScheme creates a new ColorScheme.
// If cfg is nil, default colors are used.
func NewColorScheme(cfg *config.ColorScheme) *ColorScheme {
	cs := &ColorScheme{
		enabled: true,
	}
	if cfg != nil {
		cs.config = cfg
		cs.enabled = cfg.Enabled
	}
	return cs
}

// SetEnabled enables or disables colors.
func (cs *ColorScheme) SetEnabled(enabled bool) {
	cs.enabled = enabled
}

// IsSupported returns true if colors are supported in the current environment.
// Checks for NO_COLOR env var and TERM=dumb.
func (cs *ColorScheme) IsSupported() bool {
	// Check NO_COLOR environment variable (https://no-color.org/)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// Check for dumb terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

// Colorize applies a color to text if colors are enabled and supported.
func (cs *ColorScheme) Colorize(text, color string) string {
	if !cs.enabled || !cs.IsSupported() {
		return text
	}

	code := ColorCode(color)
	if code == "" {
		return text
	}

	return code + text + ResetCode
}

// getColor returns the configured color or a default.
func (cs *ColorScheme) getColor(configColor, defaultColor string) string {
	if configColor != "" {
		return configColor
	}
	return defaultColor
}

// Directory colorizes a directory name.
func (cs *ColorScheme) Directory(text string) string {
	color := "blue"
	if cs.config != nil && cs.config.Directory != "" {
		color = cs.config.Directory
	}
	return cs.Colorize(text, color)
}

// File colorizes a file name.
func (cs *ColorScheme) File(text string) string {
	color := "white"
	if cs.config != nil && cs.config.File != "" {
		color = cs.config.File
	}
	return cs.Colorize(text, color)
}

// Executable colorizes an executable file name.
func (cs *ColorScheme) Executable(text string) string {
	color := "bright_green"
	if cs.config != nil && cs.config.Executable != "" {
		color = cs.config.Executable
	}
	return cs.Colorize(text, color)
}

// Symlink colorizes a symbolic link name.
func (cs *ColorScheme) Symlink(text string) string {
	color := "cyan"
	if cs.config != nil && cs.config.Symlink != "" {
		color = cs.config.Symlink
	}
	return cs.Colorize(text, color)
}

// Error colorizes an error message.
func (cs *ColorScheme) Error(text string) string {
	color := "red"
	if cs.config != nil && cs.config.Error != "" {
		color = cs.config.Error
	}
	return cs.Colorize(text, color)
}

// Warning colorizes a warning message.
func (cs *ColorScheme) Warning(text string) string {
	color := "yellow"
	if cs.config != nil && cs.config.Warning != "" {
		color = cs.config.Warning
	}
	return cs.Colorize(text, color)
}

// Success colorizes a success message.
func (cs *ColorScheme) Success(text string) string {
	color := "green"
	if cs.config != nil && cs.config.Success != "" {
		color = cs.config.Success
	}
	return cs.Colorize(text, color)
}

// Prompt colorizes the prompt.
func (cs *ColorScheme) Prompt(text string) string {
	color := "green"
	if cs.config != nil && cs.config.Prompt != "" {
		color = cs.config.Prompt
	}
	return cs.Colorize(text, color)
}

// GhostText colorizes ghost/suggestion text.
func (cs *ColorScheme) GhostText(text string) string {
	color := "bright_black"
	if cs.config != nil && cs.config.GhostText != "" {
		color = cs.config.GhostText
	}
	return cs.Colorize(text, color)
}

// Bold applies bold formatting to text.
func (cs *ColorScheme) Bold(text string) string {
	if !cs.enabled || !cs.IsSupported() {
		return text
	}
	return "\033[1m" + text + ResetCode
}

// Dim applies dim formatting to text.
func (cs *ColorScheme) Dim(text string) string {
	if !cs.enabled || !cs.IsSupported() {
		return text
	}
	return "\033[2m" + text + ResetCode
}

// Underline applies underline formatting to text.
func (cs *ColorScheme) Underline(text string) string {
	if !cs.enabled || !cs.IsSupported() {
		return text
	}
	return "\033[4m" + text + ResetCode
}

// FormatSize formats a file size with appropriate color.
func (cs *ColorScheme) FormatSize(size int64) string {
	var sizeStr string
	switch {
	case size >= 1<<30:
		sizeStr = strings.TrimRight(strings.TrimRight(
			strings.Replace(string(rune(size/(1<<30)))+"G", ".", ",", 1), "0"), ",") + "G"
	case size >= 1<<20:
		sizeStr = strings.TrimRight(strings.TrimRight(
			strings.Replace(string(rune(size/(1<<20)))+"M", ".", ",", 1), "0"), ",") + "M"
	case size >= 1<<10:
		sizeStr = strings.TrimRight(strings.TrimRight(
			strings.Replace(string(rune(size/(1<<10)))+"K", ".", ",", 1), "0"), ",") + "K"
	default:
		sizeStr = string(rune(size))
	}
	return sizeStr
}
