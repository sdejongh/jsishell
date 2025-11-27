package builtins

import (
	"context"

	"github.com/sdejongh/jsishell/internal/parser"
)

// ClearDefinition returns the clear command definition.
func ClearDefinition() Definition {
	return Definition{
		Name:        "clear",
		Description: "Clear the terminal screen",
		Usage:       "clear",
		Handler:     clearHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func clearHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showClearHelp(execCtx)
		return 0, nil
	}

	// ANSI escape sequence to clear screen and move cursor to home
	// ESC[2J clears the entire screen
	// ESC[H moves cursor to home position (1,1)
	execCtx.Stdout.Write([]byte("\033[2J\033[H"))

	return 0, nil
}

func showClearHelp(execCtx *Context) {
	help := `clear - Clear the terminal screen

Usage: clear

Description:
  Clears the terminal screen and moves the cursor to the top-left corner.

Examples:
  clear    Clear the screen
`
	execCtx.Stdout.Write([]byte(help))
}
