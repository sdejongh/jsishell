package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// MoveDefinition returns the move command definition.
func MoveDefinition() Definition {
	return Definition{
		Name:        "move",
		Description: "Move or rename files and directories",
		Usage:       "move [options] source... destination",
		Handler:     moveHandler,
		Options: []OptionDef{
			{Long: "--verbose", Short: "-v", Description: "Print file names as they are moved"},
			{Long: "--force", Short: "-f", Description: "Overwrite existing files without prompting"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func moveHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showMoveHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) < 2 {
		execCtx.WriteErrorln("move: missing file operand")
		return 1, nil
	}

	verbose := cmd.HasFlag("-v", "--verbose")
	force := cmd.HasFlag("-f", "--force")

	// Last argument is destination
	dest := cmd.Args[len(cmd.Args)-1]
	sources := cmd.Args[:len(cmd.Args)-1]

	// Check if destination is a directory (or should be)
	destInfo, destErr := os.Stat(dest)
	destIsDir := destErr == nil && destInfo.IsDir()

	// Multiple sources require directory destination
	if len(sources) > 1 && !destIsDir {
		execCtx.WriteErrorln("move: target '%s' is not a directory", dest)
		return 1, nil
	}

	exitCode := 0

	for _, src := range sources {
		// Check if source exists
		if _, err := os.Stat(src); err != nil {
			execCtx.WriteErrorln("move: cannot stat '%s': %v", src, err)
			exitCode = 1
			continue
		}

		// Determine actual destination path
		actualDest := dest
		if destIsDir {
			actualDest = filepath.Join(dest, filepath.Base(src))
		}

		// Check if destination exists
		if _, err := os.Stat(actualDest); err == nil {
			if !force {
				execCtx.WriteErrorln("move: '%s' already exists", actualDest)
				exitCode = 1
				continue
			}
			// Force: remove destination first
			if err := os.RemoveAll(actualDest); err != nil {
				execCtx.WriteErrorln("move: cannot remove '%s': %v", actualDest, err)
				exitCode = 1
				continue
			}
		}

		// Perform the move
		if err := os.Rename(src, actualDest); err != nil {
			execCtx.WriteErrorln("move: cannot move '%s' to '%s': %v", src, actualDest, err)
			exitCode = 1
			continue
		}

		if verbose {
			fmt.Fprintf(execCtx.Stdout, "'%s' -> '%s'\n", src, actualDest)
		}
	}

	return exitCode, nil
}

func showMoveHelp(execCtx *Context) {
	help := `move - Move or rename files and directories

Usage: move [options] source... destination

Options:
  -v, --verbose   Print file names as they are moved
  -f, --force     Overwrite existing files without prompting
      --help      Show this help message

Examples:
  move file.txt newname.txt        Rename a file
  move file1.txt file2.txt dir/    Move files to directory
  move dir1 dir2                   Rename a directory
  move -vf src.txt dst.txt         Move with verbose, force overwrite
`
	execCtx.Stdout.Write([]byte(help))
}
