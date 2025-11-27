package builtins

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sdejongh/jsishell/internal/parser"
)

// ListDefinition returns the list command definition.
func ListDefinition() Definition {
	return Definition{
		Name:        "list",
		Description: "List directory contents",
		Usage:       "list [options] [path...]",
		Handler:     listHandler,
		Options: []OptionDef{
			{Long: "--all", Short: "-a", Description: "Include hidden files (starting with .)"},
			{Long: "--long", Short: "-l", Description: "Use long listing format"},
			{Long: "--verbose", Short: "-v", Description: "Show detailed information about each entry"},
			{Long: "--quiet", Short: "-q", Description: "Only show file names, suppress other output"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// listOptions holds the options for the list command.
type listOptions struct {
	showAll    bool
	longFormat bool
	verbose    bool
	quiet      bool
}

func listHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showListHelp(execCtx)
		return 0, nil
	}

	opts := listOptions{
		showAll:    cmd.HasFlag("-a", "--all"),
		longFormat: cmd.HasFlag("-l", "--long"),
		verbose:    cmd.HasFlag("-v", "--verbose"),
		quiet:      cmd.HasFlag("-q", "--quiet"),
	}

	// --verbose implies --long
	if opts.verbose {
		opts.longFormat = true
	}

	// Default to current directory
	paths := cmd.Args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	exitCode := 0
	multiple := len(paths) > 1

	for i, path := range paths {
		if multiple && !opts.quiet {
			if i > 0 {
				fmt.Fprintln(execCtx.Stdout)
			}
			fmt.Fprintf(execCtx.Stdout, "%s:\n", path)
		}

		if err := listPath(path, opts, execCtx); err != nil {
			if !opts.quiet {
				execCtx.WriteErrorln("list: %s: %v", path, err)
			}
			exitCode = 1
		}
	}

	return exitCode, nil
}

func listPath(path string, opts listOptions, execCtx *Context) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no such file or directory")
		}
		return err
	}

	// If it's a file, just show the file
	if !info.IsDir() {
		if opts.quiet {
			fmt.Fprintln(execCtx.Stdout, filepath.Base(path))
		} else if opts.longFormat {
			printLongEntry(info, path, opts.verbose, execCtx)
		} else {
			fmt.Fprintln(execCtx.Stdout, colorizeEntry(path, info.Mode(), execCtx))
		}
		return nil
	}

	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	// Filter and display
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless -a
		if !opts.showAll && strings.HasPrefix(name, ".") {
			continue
		}

		if opts.quiet {
			// Quiet mode: just print names
			fmt.Fprintln(execCtx.Stdout, name)
		} else if opts.longFormat {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			printLongEntry(info, filepath.Join(path, name), opts.verbose, execCtx)
		} else {
			// Simple format: just names with colors
			info, err := entry.Info()
			if err != nil {
				fmt.Fprintln(execCtx.Stdout, name)
				continue
			}
			displayName := colorizeEntry(name, info.Mode(), execCtx)
			if entry.IsDir() {
				displayName += "/"
			}
			fmt.Fprintln(execCtx.Stdout, displayName)
		}
	}

	return nil
}

func printLongEntry(info fs.FileInfo, path string, verbose bool, execCtx *Context) {
	mode := info.Mode()
	size := info.Size()
	modTime := info.ModTime()
	name := filepath.Base(path)

	// Format: drwxr-xr-x  4096 Jan  2 15:04 name
	modeStr := mode.String()
	timeStr := modTime.Format("Jan _2 15:04")

	suffix := ""
	fileType := ""
	if mode.IsDir() {
		suffix = "/"
		fileType = "[dir]"
	} else if mode&0111 != 0 {
		suffix = "*"
		fileType = "[exe]"
	} else if mode&fs.ModeSymlink != 0 {
		fileType = "[lnk]"
	} else {
		fileType = "[file]"
	}

	// Colorize the name
	coloredName := colorizeEntry(name, mode, execCtx)

	if verbose {
		// Verbose format includes file type indicator
		fmt.Fprintf(execCtx.Stdout, "%s %10d %s %-6s %s%s\n", modeStr, size, timeStr, fileType, coloredName, suffix)
	} else {
		fmt.Fprintf(execCtx.Stdout, "%s %10d %s %s%s\n", modeStr, size, timeStr, coloredName, suffix)
	}
}

// colorizeEntry applies color to a file/directory name based on its mode.
func colorizeEntry(name string, mode fs.FileMode, execCtx *Context) string {
	if execCtx.Colors == nil {
		return name
	}
	if mode.IsDir() {
		return execCtx.Colors.Directory(name)
	}
	if mode&0111 != 0 {
		return execCtx.Colors.Executable(name)
	}
	if mode&fs.ModeSymlink != 0 {
		return execCtx.Colors.Symlink(name)
	}
	return execCtx.Colors.File(name)
}

func showListHelp(execCtx *Context) {
	help := `list - List directory contents

Usage: list [options] [path...]

Options:
  -a, --all      Include hidden files (starting with .)
  -l, --long     Use long listing format
  -v, --verbose  Show detailed information (implies --long, adds file type)
  -q, --quiet    Only show file names, suppress other output
      --help     Show this help message

Long format shows: permissions, size, date, name
Verbose format adds: file type indicator ([dir], [file], [exe], [lnk])

Examples:
  list              List current directory
  list -a           List including hidden files
  list -l           Long format listing
  list -v           Verbose listing with file types
  list -q           Quiet mode, names only
  list /home        List /home directory
  list dir1 dir2    List multiple directories
`
	execCtx.Stdout.Write([]byte(help))
}
