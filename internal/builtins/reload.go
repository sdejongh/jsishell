package builtins

import (
	"context"

	"github.com/sdejongh/jsishell/internal/config"
	"github.com/sdejongh/jsishell/internal/parser"
)

// ReloadCallback is called when configuration is reloaded.
// The shell should set this to apply new configuration.
var ReloadCallback func(*config.Config)

// SetReloadCallback sets the function to call when config is reloaded.
func SetReloadCallback(cb func(*config.Config)) {
	ReloadCallback = cb
}

// ReloadDefinition returns the reload command definition.
func ReloadDefinition() Definition {
	return Definition{
		Name:        "reload",
		Description: "Reload shell configuration",
		Usage:       "reload",
		Handler:     reloadHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func reloadHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showReloadHelp(execCtx)
		return 0, nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		execCtx.Stderr.Write([]byte("Error loading config: " + err.Error() + "\n"))
		return 1, err
	}

	// Call the reload callback if set
	if ReloadCallback != nil {
		ReloadCallback(cfg)
	}

	execCtx.Stdout.Write([]byte("Configuration reloaded\n"))
	return 0, nil
}

func showReloadHelp(execCtx *Context) {
	help := `reload - Reload shell configuration

Usage: reload

Reloads the configuration from the config file at:
  ~/.config/jsishell/config.yaml

Options:
  --help    Show this help message

Examples:
  reload    Reload configuration from disk
`
	execCtx.Stdout.Write([]byte(help))
}
