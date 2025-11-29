package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// CpDefinition returns the cp command definition.
func CpDefinition() Definition {
	return Definition{
		Name:        "cp",
		Description: "Copy files and directories",
		Usage:       "cp [options] source... destination",
		Handler:     cpHandler,
		Options: []OptionDef{
			{Long: "--recursive", Short: "-r", Description: "Copy directories recursively"},
			{Long: "--verbose", Short: "-v", Description: "Print file names as they are copied"},
			{Long: "--force", Short: "-f", Description: "Overwrite existing files without prompting"},
			{Long: "--exclude", Short: "-e", HasValue: true, Description: "Exclude files matching glob pattern (can be repeated)"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// cpOptions holds the options for the cp command.
type cpOptions struct {
	recursive       bool
	verbose         bool
	force           bool
	excludePatterns []string
}

func cpHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showCpHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) < 2 {
		execCtx.WriteErrorln("cp: missing file operand")
		return 1, nil
	}

	opts := cpOptions{
		recursive:       cmd.HasFlag("-r", "--recursive"),
		verbose:         cmd.HasFlag("-v", "--verbose"),
		force:           cmd.HasFlag("-f", "--force"),
		excludePatterns: cmd.GetOptions("-e", "--exclude"),
	}

	// Last argument is destination
	dest := cmd.Args[len(cmd.Args)-1]
	sources := cmd.Args[:len(cmd.Args)-1]

	// Check if destination is a directory (or should be)
	destInfo, destErr := os.Stat(dest)
	destIsDir := destErr == nil && destInfo.IsDir()

	// Multiple sources require directory destination
	if len(sources) > 1 && !destIsDir {
		execCtx.WriteErrorln("cp: target '%s' is not a directory", dest)
		return 1, nil
	}

	exitCode := 0

	for _, src := range sources {
		// Check if source matches exclude pattern
		if matchesCpExcludePattern(filepath.Base(src), opts.excludePatterns) {
			continue
		}

		srcInfo, err := os.Stat(src)
		if err != nil {
			execCtx.WriteErrorln("cp: cannot stat '%s': %v", src, err)
			exitCode = 1
			continue
		}

		// Determine actual destination path
		actualDest := dest
		if destIsDir {
			actualDest = filepath.Join(dest, filepath.Base(src))
		}

		// Check if source is a directory
		if srcInfo.IsDir() {
			if !opts.recursive {
				execCtx.WriteErrorln("cp: omitting directory '%s'", src)
				exitCode = 1
				continue
			}
			if err := cpDir(src, actualDest, opts, execCtx); err != nil {
				execCtx.WriteErrorln("cp: error copying '%s': %v", src, err)
				exitCode = 1
			}
		} else {
			if err := cpFile(src, actualDest, opts, execCtx); err != nil {
				execCtx.WriteErrorln("cp: error copying '%s': %v", src, err)
				exitCode = 1
			}
		}
	}

	return exitCode, nil
}

// matchesCpExcludePattern checks if a name matches any of the exclude patterns.
func matchesCpExcludePattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func cpFile(src, dest string, opts cpOptions, execCtx *Context) error {
	// Check if destination exists
	if _, err := os.Stat(dest); err == nil {
		if !opts.force {
			execCtx.WriteErrorln("cp: '%s' already exists", dest)
			return nil
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	if opts.verbose {
		fmt.Fprintf(execCtx.Stdout, "'%s' -> '%s'\n", src, dest)
	}

	return nil
}

func cpDir(src, dest string, opts cpOptions, execCtx *Context) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	if opts.verbose {
		fmt.Fprintf(execCtx.Stdout, "'%s' -> '%s'\n", src, dest)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		// Check if entry matches exclude pattern
		if matchesCpExcludePattern(name, opts.excludePatterns) {
			continue
		}

		srcPath := filepath.Join(src, name)
		destPath := filepath.Join(dest, name)

		if entry.IsDir() {
			if err := cpDir(srcPath, destPath, opts, execCtx); err != nil {
				return err
			}
		} else {
			if err := cpFile(srcPath, destPath, opts, execCtx); err != nil {
				return err
			}
		}
	}

	return nil
}

func showCpHelp(execCtx *Context) {
	help := `cp - Copy files and directories

Usage: cp [options] source... destination

Options:
  -r, --recursive        Copy directories recursively
  -v, --verbose          Print file names as they are copied
  -f, --force            Overwrite existing files without prompting
  -e, --exclude=<glob>   Exclude files matching glob pattern (can be repeated)
      --help             Show this help message

Exclude patterns:
  Standard glob patterns: *, ?, [abc], [a-z]
  Can be specified multiple times to exclude several patterns.

Examples:
  cp file.txt backup.txt             Copy a file
  cp file1.txt file2.txt dir/        Copy files to directory
  cp -r source/ dest/                Copy directory recursively
  cp -vf src.txt dst.txt             Copy with verbose, force overwrite
  cp -r --exclude=*.log src/ dest/   Copy excluding log files
  cp -r -e=*.tmp -e=*.bak src/ dst/  Copy excluding .tmp and .bak files
`
	execCtx.Stdout.Write([]byte(help))
}
