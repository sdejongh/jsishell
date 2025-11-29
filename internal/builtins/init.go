package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// InitDefinition returns the init command definition.
func InitDefinition() Definition {
	return Definition{
		Name:        "init",
		Description: "Generate default configuration file",
		Usage:       "init [options]",
		Handler:     initHandler,
		Options: []OptionDef{
			{Long: "--force", Short: "-f", Description: "Overwrite existing configuration file"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// defaultConfigContent returns the default configuration file content.
func defaultConfigContent() string {
	return `# JSIShell Configuration
# ======================
# Configuration file for JSIShell
# Location: ~/.config/jsishell/config.yaml

# Shell prompt
# Supported variables:
#   %d  - Current working directory (full path)
#   %D  - Current working directory (basename only)
#   %~  - Current working directory with ~ for home
#   %u  - Username
#   %h  - Hostname (short, without domain)
#   %H  - Hostname (full, with domain)
#   %t  - Time (HH:MM)
#   %T  - Time with seconds (HH:MM:SS)
#   %n  - Newline
#   %$  - Shell indicator ($ for user, # for root)
#   %%  - Literal %
#
# Color codes (use %{color} and %{reset} or %{/}):
#   Colors: black, red, green, yellow, blue, magenta, cyan, white
#   Bright: bright_black, bright_red, bright_green, bright_yellow,
#           bright_blue, bright_magenta, bright_cyan, bright_white
#   Styles: bold, dim, underline
#
# Examples:
#   prompt: "$ "                                    # Simple prompt
#   prompt: "%{green}%u@%h%{/}:%{blue}%~%{/}%$ "    # Colored bash-style
#   prompt: "[%t] %D> "                             # Time and directory
#   prompt: "%{bold}%{cyan}%~%{/} λ "              # Bold cyan path with lambda
#   prompt: "%{green}%u@%h%{/}:%{blue}%~%{/}%$ "  # Bash-style with root indicator
#
prompt: "%{green}%u@%h%{/}:%{blue}%~%{/}%$ "

# History settings
history:
  # Maximum number of commands to keep in history
  max_size: 1000

  # History file location (supports ~ for home directory)
  file: "~/.config/jsishell/.jsishell_history"

  # Don't save duplicate consecutive commands
  ignore_duplicates: true

  # Don't save commands starting with a space
  # Useful for sensitive commands you don't want in history
  ignore_space_prefix: true

# Color scheme settings
colors:
  # Enable/disable colors globally
  # Also respects the NO_COLOR environment variable
  enabled: true

  # Available colors:
  #   black, red, green, yellow, blue, magenta, cyan, white
  #   bright_black, bright_red, bright_green, bright_yellow
  #   bright_blue, bright_magenta, bright_cyan, bright_white

  # Directory listing colors
  directory: "blue"
  file: "white"
  executable: "bright_green"
  symlink: "cyan"

  # Message colors
  error: "red"
  warning: "yellow"
  success: "green"

  # Autocompletion ghost text color
  ghost_text: "bright_black"

# Command abbreviations
abbreviations:
  # Enable command abbreviations (e.g., 'l' for 'ls' if unambiguous)
  enabled: true

# Line editor settings
editor:
  # Tab width for indentation
  tab_width: 4
`
}

func initHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showInitHelp(execCtx)
		return 0, nil
	}

	force := cmd.HasFlag("-f", "--force")

	// Get config directory
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		execCtx.WriteErrorln("Configuration file already exists: %s", configPath)
		execCtx.WriteErrorln("Use --force to overwrite")
		return 1, nil
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		execCtx.WriteErrorln("Failed to create config directory: %v", err)
		return 1, nil
	}

	// Write config file
	content := defaultConfigContent()
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		execCtx.WriteErrorln("Failed to write config file: %v", err)
		return 1, nil
	}

	fmt.Fprintf(execCtx.Stdout, "Configuration file created: %s\n", configPath)
	fmt.Fprintln(execCtx.Stdout, "Run 'reload' to apply the new configuration.")

	return 0, nil
}

// getConfigDir returns the configuration directory path.
func getConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "jsishell")
	}

	// Fall back to ~/.config/jsishell
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".jsishell"
	}

	return filepath.Join(homeDir, ".config", "jsishell")
}

func showInitHelp(execCtx *Context) {
	help := `init - Generate default configuration file

Usage: init [options]

Options:
  -f, --force    Overwrite existing configuration file
      --help     Show this help message

Description:
  Creates a default configuration file at ~/.config/jsishell/config.yaml
  (or $XDG_CONFIG_HOME/jsishell/config.yaml if XDG_CONFIG_HOME is set).

  The configuration file includes:
  - Prompt customization with colors and variables
  - History settings
  - Color scheme for output
  - Command abbreviation settings

  If a configuration file already exists, use --force to overwrite it.

Examples:
  init            Create config if it doesn't exist
  init --force    Overwrite existing config
`
	execCtx.Stdout.Write([]byte(help))
}
