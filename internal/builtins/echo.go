package builtins

import (
	"context"
	"strings"

	"github.com/sdejongh/jsishell/internal/parser"
)

// EchoDefinition returns the echo command definition.
func EchoDefinition() Definition {
	return Definition{
		Name:        "echo",
		Description: "Print arguments to standard output",
		Usage:       "echo [options] [arguments...]",
		Handler:     echoHandler,
		Options: []OptionDef{
			{Long: "--no-newline", Short: "-n", Description: "Do not output trailing newline"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func echoHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showEchoHelp(execCtx)
		return 0, nil
	}

	// Check for -n (no newline)
	noNewline := cmd.HasFlag("-n", "--no-newline")

	// Print all arguments separated by spaces
	output := strings.Join(cmd.Args, " ")
	execCtx.Stdout.Write([]byte(output))

	if !noNewline {
		execCtx.Stdout.Write([]byte("\n"))
	}

	return 0, nil
}

func showEchoHelp(execCtx *Context) {
	help := `echo - Print arguments to standard output

Usage: echo [options] [arguments...]

Options:
  -n, --no-newline   Do not output trailing newline
      --help         Show this help message

Examples:
  echo hello world     Prints "hello world"
  echo -n no newline   Prints without trailing newline
`
	execCtx.Stdout.Write([]byte(help))
}
