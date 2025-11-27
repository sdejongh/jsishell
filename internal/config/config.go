// Package config handles shell configuration loading and management.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Default configuration values.
const (
	// DefaultPrompt uses prompt variables: %D (dir basename), supports %u (user), %h (host), %~ (cwd with ~), etc.
	DefaultPrompt      = "jsi:%D> "
	DefaultHistorySize = 1000
	DefaultHistoryFile = ".jsishell_history"
)

// Valid color names for terminal output.
var validColors = map[string]bool{
	"black":          true,
	"red":            true,
	"green":          true,
	"yellow":         true,
	"blue":           true,
	"magenta":        true,
	"cyan":           true,
	"white":          true,
	"bright_black":   true,
	"bright_red":     true,
	"bright_green":   true,
	"bright_yellow":  true,
	"bright_blue":    true,
	"bright_magenta": true,
	"bright_cyan":    true,
	"bright_white":   true,
}

// Config represents the shell configuration.
type Config struct {
	Prompt        string              `yaml:"prompt"`
	History       HistoryConfig       `yaml:"history"`
	Colors        ColorScheme         `yaml:"colors"`
	Abbreviations AbbreviationsConfig `yaml:"abbreviations"`
	Editor        EditorConfig        `yaml:"editor"`
}

// HistoryConfig holds history-related settings.
type HistoryConfig struct {
	MaxSize           int    `yaml:"max_size"`
	File              string `yaml:"file"`
	IgnoreDuplicates  bool   `yaml:"ignore_duplicates"`
	IgnoreSpacePrefix bool   `yaml:"ignore_space_prefix"`
}

// ColorScheme defines colors for different output types.
type ColorScheme struct {
	Enabled    bool   `yaml:"enabled"`
	Prompt     string `yaml:"prompt"`
	Directory  string `yaml:"directory"`
	File       string `yaml:"file"`
	Executable string `yaml:"executable"`
	Symlink    string `yaml:"symlink"`
	Error      string `yaml:"error"`
	Warning    string `yaml:"warning"`
	Success    string `yaml:"success"`
	GhostText  string `yaml:"ghost_text"`
}

// AbbreviationsConfig holds abbreviation settings.
type AbbreviationsConfig struct {
	Enabled bool `yaml:"enabled"`
}

// EditorConfig holds line editor settings.
type EditorConfig struct {
	TabWidth int `yaml:"tab_width"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Prompt: DefaultPrompt,
		History: HistoryConfig{
			MaxSize:           DefaultHistorySize,
			File:              filepath.Join(ConfigDir(), DefaultHistoryFile),
			IgnoreDuplicates:  true,
			IgnoreSpacePrefix: true,
		},
		Colors: ColorScheme{
			Enabled:    true,
			Prompt:     "green",
			Directory:  "blue",
			File:       "white",
			Executable: "bright_green",
			Symlink:    "cyan",
			Error:      "red",
			Warning:    "yellow",
			Success:    "green",
			GhostText:  "bright_black",
		},
		Abbreviations: AbbreviationsConfig{
			Enabled: true,
		},
		Editor: EditorConfig{
			TabWidth: 4,
		},
	}
}

// ConfigDir returns the configuration directory path.
func ConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "jsishell")
	}

	// Fall back to ~/.config/jsishell
	home, err := os.UserHomeDir()
	if err != nil {
		return ".jsishell"
	}

	return filepath.Join(home, ".config", "jsishell")
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}

// LoadFromFile loads configuration from a YAML file.
// Returns default config if file doesn't exist.
func LoadFromFile(path string) (*Config, error) {
	path = ExpandPath(path)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Start with defaults
	cfg := Default()

	// Parse YAML into a temporary struct to get user values
	var userCfg Config
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Merge user config with defaults
	cfg = cfg.Merge(&userCfg)

	// Validate the merged config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Load loads configuration from the default path.
func Load() (*Config, error) {
	return LoadFromFile(ConfigPath())
}

// Merge merges another config into this one.
// Non-zero values from other override values in this config.
func (c *Config) Merge(other *Config) *Config {
	result := *c // Copy base config

	if other.Prompt != "" {
		result.Prompt = other.Prompt
	}

	// Merge history
	if other.History.MaxSize != 0 {
		result.History.MaxSize = other.History.MaxSize
	}
	if other.History.File != "" {
		result.History.File = other.History.File
	}
	// Booleans are tricky - we can't distinguish false from unset
	// For now, we only override if the YAML explicitly sets them
	// This is handled by the YAML parser

	// Merge colors
	if other.Colors.Prompt != "" {
		result.Colors.Prompt = other.Colors.Prompt
	}
	if other.Colors.Directory != "" {
		result.Colors.Directory = other.Colors.Directory
	}
	if other.Colors.File != "" {
		result.Colors.File = other.Colors.File
	}
	if other.Colors.Executable != "" {
		result.Colors.Executable = other.Colors.Executable
	}
	if other.Colors.Symlink != "" {
		result.Colors.Symlink = other.Colors.Symlink
	}
	if other.Colors.Error != "" {
		result.Colors.Error = other.Colors.Error
	}
	if other.Colors.Warning != "" {
		result.Colors.Warning = other.Colors.Warning
	}
	if other.Colors.Success != "" {
		result.Colors.Success = other.Colors.Success
	}
	if other.Colors.GhostText != "" {
		result.Colors.GhostText = other.Colors.GhostText
	}

	// Merge editor
	if other.Editor.TabWidth != 0 {
		result.Editor.TabWidth = other.Editor.TabWidth
	}

	return &result
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate history size
	if c.History.MaxSize < 0 {
		return errors.New("history.max_size cannot be negative")
	}

	// Validate colors
	colorFields := map[string]string{
		"prompt":     c.Colors.Prompt,
		"directory":  c.Colors.Directory,
		"file":       c.Colors.File,
		"executable": c.Colors.Executable,
		"symlink":    c.Colors.Symlink,
		"error":      c.Colors.Error,
		"warning":    c.Colors.Warning,
		"success":    c.Colors.Success,
		"ghost_text": c.Colors.GhostText,
	}

	for name, color := range colorFields {
		if color != "" && !IsValidColor(color) {
			return fmt.Errorf("invalid color %q for %s", color, name)
		}
	}

	// Validate editor
	if c.Editor.TabWidth < 1 {
		c.Editor.TabWidth = 4 // Reset to default
	}

	return nil
}

// IsValidColor checks if a color name is valid.
func IsValidColor(color string) bool {
	return validColors[color]
}

// Save saves the configuration to a file.
func (c *Config) Save(path string) error {
	path = ExpandPath(path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
