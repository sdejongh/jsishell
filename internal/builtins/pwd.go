package builtins

import (
	"context"
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/parser"
)

// PwdDefinition returns the pwd command definition.
func PwdDefinition() Definition {
	return Definition{
		Name:        "pwd",
		Description: "Print the current working directory",
		Usage:       "pwd",
		Handler:     pwdHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func pwdHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showPwdHelp(execCtx)
		return 0, nil
	}

	// Get current working directory
	pwd := execCtx.Env.Get("PWD")
	if pwd == "" {
		var err error
		pwd, err = os.Getwd()
		if err != nil {
			execCtx.WriteErrorln("pwd: %v", err)
			return 1, nil
		}
	}

	fmt.Fprintln(execCtx.Stdout, pwd)
	return 0, nil
}

func showPwdHelp(execCtx *Context) {
	help := `pwd - Print the current working directory

Usage: pwd

Description:
  Prints the absolute path of the current working directory.

Examples:
  pwd    Print current directory
`
	execCtx.Stdout.Write([]byte(help))
}
