package builtins

import (
	"context"
	"strconv"

	"github.com/sdejongh/jsishell/internal/parser"
)

// ExitCode is a special error type that signals shell exit.
type ExitCode struct {
	Code int
}

func (e ExitCode) Error() string {
	return "exit"
}

// ExitDefinition returns the exit command definition.
func ExitDefinition() Definition {
	return Definition{
		Name:        "exit",
		Description: "Exit the shell",
		Usage:       "exit [code]",
		Handler:     exitHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func exitHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showExitHelp(execCtx)
		return 0, nil
	}

	// Default exit code is 0
	code := 0

	// If an argument is provided, use it as exit code
	if len(cmd.Args) > 0 {
		var err error
		code, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			execCtx.Stderr.Write([]byte("exit: invalid exit code: " + cmd.Args[0] + "\n"))
			code = 1
		}
	}

	// Return special ExitCode error to signal shell should exit
	return code, ExitCode{Code: code}
}

func showExitHelp(execCtx *Context) {
	help := `exit - Exit the shell

Usage: exit [code]

Arguments:
  code   Exit code (default: 0)

Examples:
  exit      Exit with code 0
  exit 1    Exit with code 1
`
	execCtx.Stdout.Write([]byte(help))
}
