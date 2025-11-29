package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// RmDefinition returns the rm command definition.
func RmDefinition() Definition {
	return Definition{
		Name:        "rm",
		Description: "Remove files and directories",
		Usage:       "rm [options] file...",
		Handler:     rmHandler,
		Options: []OptionDef{
			{Long: "--recursive", Short: "-r", Description: "Remove directories and their contents recursively"},
			{Long: "--force", Short: "-f", Description: "Ignore nonexistent files, never prompt"},
			{Long: "--yes", Short: "-y", Description: "Skip confirmation prompts (same as --force)"},
			{Long: "--verbose", Short: "-v", Description: "Print file names as they are removed"},
			{Long: "--quiet", Short: "-q", Description: "Suppress error messages"},
			{Long: "--exclude", Short: "-e", HasValue: true, Description: "Exclude files matching glob pattern (can be repeated)"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// rmOptions holds the options for the rm command.
type rmOptions struct {
	recursive       bool
	force           bool
	verbose         bool
	quiet           bool
	excludePatterns []string
}

func rmHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showRmHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) == 0 {
		execCtx.WriteErrorln("rm: missing operand")
		return 1, nil
	}

	opts := rmOptions{
		recursive:       cmd.HasFlag("-r", "--recursive"),
		force:           cmd.HasFlag("-f", "--force", "-y", "--yes"), // --yes is alias for --force
		verbose:         cmd.HasFlag("-v", "--verbose"),
		quiet:           cmd.HasFlag("-q", "--quiet"),
		excludePatterns: cmd.GetOptions("-e", "--exclude"),
	}

	exitCode := 0

	for _, path := range cmd.Args {
		// Check if path matches exclude pattern
		if matchesRmExcludePattern(filepath.Base(path), opts.excludePatterns) {
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if !opts.force && !opts.quiet {
					execCtx.WriteErrorln("rm: cannot remove '%s': No such file or directory", path)
					exitCode = 1
				}
				continue
			}
			if !opts.quiet {
				execCtx.WriteErrorln("rm: cannot remove '%s': %v", path, err)
			}
			exitCode = 1
			continue
		}

		// Check if it's a directory
		if info.IsDir() {
			if !opts.recursive {
				if !opts.quiet {
					execCtx.WriteErrorln("rm: cannot remove '%s': Is a directory", path)
				}
				exitCode = 1
				continue
			}
			// Recursive remove with exclude support
			if err := rmDirRecursive(path, opts, execCtx); err != nil {
				if !opts.quiet {
					execCtx.WriteErrorln("rm: cannot remove '%s': %v", path, err)
				}
				exitCode = 1
				continue
			}
		} else {
			// Remove file
			if err := os.Remove(path); err != nil {
				if !opts.quiet {
					execCtx.WriteErrorln("rm: cannot remove '%s': %v", path, err)
				}
				exitCode = 1
				continue
			}
			if opts.verbose && !opts.quiet {
				fmt.Fprintf(execCtx.Stdout, "removed '%s'\n", path)
			}
		}
	}

	return exitCode, nil
}

// matchesRmExcludePattern checks if a name matches any of the exclude patterns.
func matchesRmExcludePattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

// rmDirRecursive removes a directory recursively, respecting exclude patterns.
func rmDirRecursive(path string, opts rmOptions, execCtx *Context) error {
	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Remove contents first
	for _, entry := range entries {
		name := entry.Name()
		entryPath := filepath.Join(path, name)

		// Check if entry matches exclude pattern
		if matchesRmExcludePattern(name, opts.excludePatterns) {
			continue
		}

		if entry.IsDir() {
			if err := rmDirRecursive(entryPath, opts, execCtx); err != nil {
				return err
			}
		} else {
			if err := os.Remove(entryPath); err != nil {
				return err
			}
			if opts.verbose && !opts.quiet {
				fmt.Fprintf(execCtx.Stdout, "removed '%s'\n", entryPath)
			}
		}
	}

	// Check if directory is empty before removing it
	remaining, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Only remove directory if it's empty (all contents were removed or excluded)
	if len(remaining) == 0 {
		if err := os.Remove(path); err != nil {
			return err
		}
		if opts.verbose && !opts.quiet {
			fmt.Fprintf(execCtx.Stdout, "removed directory '%s'\n", path)
		}
	}

	return nil
}

func showRmHelp(execCtx *Context) {
	help := `rm - Remove files and directories

Usage: rm [options] file...

Options:
  -r, --recursive        Remove directories and their contents recursively
  -f, --force            Ignore nonexistent files, never prompt
  -y, --yes              Skip confirmation prompts (same as --force)
  -v, --verbose          Print file names as they are removed
  -q, --quiet            Suppress error messages
  -e, --exclude=<glob>   Exclude files matching glob pattern (can be repeated)
      --help             Show this help message

Exclude patterns:
  Standard glob patterns: *, ?, [abc], [a-z]
  Can be specified multiple times to exclude several patterns.
  When using --exclude with -r, directories containing excluded files
  will not be removed (they won't be empty).

Examples:
  rm file.txt                     Remove a file
  rm -r directory/                Remove directory recursively
  rm -f nonexistent               No error for missing files
  rm -rv dir1 dir2                Remove multiple directories verbosely
  rm -q missing                   Silently ignore missing files
  rm -r --exclude=*.log dir/      Remove dir but keep .log files
  rm -r -e=*.txt -e=*.md dir/     Remove dir but keep .txt and .md files
`
	execCtx.Stdout.Write([]byte(help))
}
