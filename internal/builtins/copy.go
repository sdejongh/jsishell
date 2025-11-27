package builtins

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// CopyDefinition returns the copy command definition.
func CopyDefinition() Definition {
	return Definition{
		Name:        "copy",
		Description: "Copy files and directories",
		Usage:       "copy [options] source... destination",
		Handler:     copyHandler,
		Options: []OptionDef{
			{Long: "--recursive", Short: "-r", Description: "Copy directories recursively"},
			{Long: "--verbose", Short: "-v", Description: "Print file names as they are copied"},
			{Long: "--force", Short: "-f", Description: "Overwrite existing files without prompting"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func copyHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showCopyHelp(execCtx)
		return 0, nil
	}

	if len(cmd.Args) < 2 {
		execCtx.WriteErrorln("copy: missing file operand")
		return 1, nil
	}

	recursive := cmd.HasFlag("-r", "--recursive")
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
		execCtx.WriteErrorln("copy: target '%s' is not a directory", dest)
		return 1, nil
	}

	exitCode := 0

	for _, src := range sources {
		srcInfo, err := os.Stat(src)
		if err != nil {
			execCtx.WriteErrorln("copy: cannot stat '%s': %v", src, err)
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
			if !recursive {
				execCtx.WriteErrorln("copy: omitting directory '%s'", src)
				exitCode = 1
				continue
			}
			if err := copyDir(src, actualDest, verbose, force, execCtx); err != nil {
				execCtx.WriteErrorln("copy: error copying '%s': %v", src, err)
				exitCode = 1
			}
		} else {
			if err := copyFile(src, actualDest, verbose, force, execCtx); err != nil {
				execCtx.WriteErrorln("copy: error copying '%s': %v", src, err)
				exitCode = 1
			}
		}
	}

	return exitCode, nil
}

func copyFile(src, dest string, verbose, force bool, execCtx *Context) error {
	// Check if destination exists
	if _, err := os.Stat(dest); err == nil {
		if !force {
			execCtx.WriteErrorln("copy: '%s' already exists", dest)
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

	if verbose {
		fmt.Fprintf(execCtx.Stdout, "'%s' -> '%s'\n", src, dest)
	}

	return nil
}

func copyDir(src, dest string, verbose, force bool, execCtx *Context) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(execCtx.Stdout, "'%s' -> '%s'\n", src, dest)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath, verbose, force, execCtx); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath, verbose, force, execCtx); err != nil {
				return err
			}
		}
	}

	return nil
}

func showCopyHelp(execCtx *Context) {
	help := `copy - Copy files and directories

Usage: copy [options] source... destination

Options:
  -r, --recursive   Copy directories recursively
  -v, --verbose     Print file names as they are copied
  -f, --force       Overwrite existing files without prompting
      --help        Show this help message

Examples:
  copy file.txt backup.txt         Copy a file
  copy file1.txt file2.txt dir/    Copy files to directory
  copy -r source/ dest/            Copy directory recursively
  copy -vf src.txt dst.txt         Copy with verbose, force overwrite
`
	execCtx.Stdout.Write([]byte(help))
}
