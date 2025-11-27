package builtins

import (
	"context"
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/parser"
)

// RemoveDefinition returns the remove command definition.
func RemoveDefinition() Definition {
	return Definition{
		Name:        "remove",
		Description: "Remove files and directories",
		Usage:       "remove [options] file...",
		Handler:     removeHandler,
		Options: []OptionDef{
			{Long: "--recursive", Short: "-r", Description: "Remove directories and their contents recursively"},
			{Long: "--force", Short: "-f", Description: "Ignore nonexistent files, never prompt"},
			{Long: "--yes", Short: "-y", Description: "Skip confirmation prompts (same as --force)"},
			{Long: "--verbose", Short: "-v", Description: "Print file names as they are removed"},
			{Long: "--quiet", Short: "-q", Description: "Suppress error messages"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func removeHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showRemoveHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) == 0 {
		execCtx.WriteErrorln("remove: missing operand")
		return 1, nil
	}

	recursive := cmd.HasFlag("-r", "--recursive")
	force := cmd.HasFlag("-f", "--force", "-y", "--yes") // --yes is alias for --force
	verbose := cmd.HasFlag("-v", "--verbose")
	quiet := cmd.HasFlag("-q", "--quiet")

	exitCode := 0

	for _, path := range cmd.Args {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if !force && !quiet {
					execCtx.WriteErrorln("remove: cannot remove '%s': No such file or directory", path)
					exitCode = 1
				}
				continue
			}
			if !quiet {
				execCtx.WriteErrorln("remove: cannot remove '%s': %v", path, err)
			}
			exitCode = 1
			continue
		}

		// Check if it's a directory
		if info.IsDir() {
			if !recursive {
				if !quiet {
					execCtx.WriteErrorln("remove: cannot remove '%s': Is a directory", path)
				}
				exitCode = 1
				continue
			}
			// Recursive remove
			if err := os.RemoveAll(path); err != nil {
				if !quiet {
					execCtx.WriteErrorln("remove: cannot remove '%s': %v", path, err)
				}
				exitCode = 1
				continue
			}
		} else {
			// Remove file
			if err := os.Remove(path); err != nil {
				if !quiet {
					execCtx.WriteErrorln("remove: cannot remove '%s': %v", path, err)
				}
				exitCode = 1
				continue
			}
		}

		if verbose && !quiet {
			fmt.Fprintf(execCtx.Stdout, "removed '%s'\n", path)
		}
	}

	return exitCode, nil
}

func showRemoveHelp(execCtx *Context) {
	help := `remove - Remove files and directories

Usage: remove [options] file...

Options:
  -r, --recursive   Remove directories and their contents recursively
  -f, --force       Ignore nonexistent files, never prompt
  -y, --yes         Skip confirmation prompts (same as --force)
  -v, --verbose     Print file names as they are removed
  -q, --quiet       Suppress error messages
      --help        Show this help message

Examples:
  remove file.txt            Remove a file
  remove -r directory/       Remove directory recursively
  remove -f nonexistent      No error for missing files
  remove -rv dir1 dir2       Remove multiple directories verbosely
  remove -q missing          Silently ignore missing files
`
	execCtx.Stdout.Write([]byte(help))
}
