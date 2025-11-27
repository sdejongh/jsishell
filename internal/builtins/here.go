package builtins

import (
	"context"
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/parser"
)

// HereDefinition returns the here command definition.
func HereDefinition() Definition {
	return Definition{
		Name:        "here",
		Description: "Print the current working directory",
		Usage:       "here",
		Handler:     hereHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func hereHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showHereHelp(execCtx)
		return 0, nil
	}

	// Get current working directory
	pwd := execCtx.Env.Get("PWD")
	if pwd == "" {
		var err error
		pwd, err = os.Getwd()
		if err != nil {
			execCtx.WriteErrorln("here: %v", err)
			return 1, nil
		}
	}

	fmt.Fprintln(execCtx.Stdout, pwd)
	return 0, nil
}

func showHereHelp(execCtx *Context) {
	help := `here - Print the current working directory

Usage: here

Description:
  Prints the absolute path of the current working directory.

Examples:
  here    Print current directory
`
	execCtx.Stdout.Write([]byte(help))
}
