package builtins

import (
	"context"
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/parser"
)

// MkdirDefinition returns the mkdir command definition.
func MkdirDefinition() Definition {
	return Definition{
		Name:        "mkdir",
		Description: "Create directories",
		Usage:       "mkdir [options] directory...",
		Handler:     mkdirHandler,
		Options: []OptionDef{
			{Long: "--parents", Short: "-p", Description: "Create parent directories as needed"},
			{Long: "--verbose", Short: "-v", Description: "Print a message for each created directory"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func mkdirHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showMkdirHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) == 0 {
		execCtx.WriteErrorln("mkdir: missing operand")
		return 1, nil
	}

	parents := cmd.HasFlag("-p", "--parents")
	verbose := cmd.HasFlag("-v", "--verbose")

	exitCode := 0

	for _, dir := range cmd.Args {
		var err error
		if parents {
			err = os.MkdirAll(dir, 0755)
		} else {
			err = os.Mkdir(dir, 0755)
		}

		if err != nil {
			if os.IsExist(err) && parents {
				// With -p, existing directories are OK
				continue
			}
			execCtx.WriteErrorln("mkdir: cannot create directory '%s': %v", dir, err)
			exitCode = 1
		} else if verbose {
			fmt.Fprintf(execCtx.Stdout, "mkdir: created directory '%s'\n", dir)
		}
	}

	return exitCode, nil
}

func showMkdirHelp(execCtx *Context) {
	help := `mkdir - Create directories

Usage: mkdir [options] directory...

Options:
  -p, --parents   Create parent directories as needed
  -v, --verbose   Print a message for each created directory
      --help      Show this help message

Examples:
  mkdir docs           Create 'docs' directory
  mkdir -p a/b/c       Create nested directories
  mkdir -v new_folder  Create with verbose output
`
	execCtx.Stdout.Write([]byte(help))
}
