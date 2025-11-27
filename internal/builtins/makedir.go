package builtins

import (
	"context"
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/parser"
)

// MakedirDefinition returns the makedir command definition.
func MakedirDefinition() Definition {
	return Definition{
		Name:        "makedir",
		Description: "Create directories",
		Usage:       "makedir [options] directory...",
		Handler:     makedirHandler,
		Options: []OptionDef{
			{Long: "--parents", Short: "-p", Description: "Create parent directories as needed"},
			{Long: "--verbose", Short: "-v", Description: "Print a message for each created directory"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func makedirHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showMakedirHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) == 0 {
		execCtx.WriteErrorln("makedir: missing operand")
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
			execCtx.WriteErrorln("makedir: cannot create directory '%s': %v", dir, err)
			exitCode = 1
		} else if verbose {
			fmt.Fprintf(execCtx.Stdout, "makedir: created directory '%s'\n", dir)
		}
	}

	return exitCode, nil
}

func showMakedirHelp(execCtx *Context) {
	help := `makedir - Create directories

Usage: makedir [options] directory...

Options:
  -p, --parents   Create parent directories as needed
  -v, --verbose   Print a message for each created directory
      --help      Show this help message

Examples:
  makedir docs           Create 'docs' directory
  makedir -p a/b/c       Create nested directories
  makedir -v new_folder  Create with verbose output
`
	execCtx.Stdout.Write([]byte(help))
}
