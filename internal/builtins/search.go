package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sdejongh/jsishell/internal/parser"
)

// SearchDefinition returns the search command definition.
func SearchDefinition() Definition {
	return Definition{
		Name:        "search",
		Description: "Search for files and directories",
		Usage:       "search <directory> <expression> [options]",
		Handler:     searchHandler,
		Options: []OptionDef{
			{Long: "--recursive", Short: "-r", Description: "Search recursively in subdirectories"},
			{Long: "--level", Short: "-l", HasValue: true, Description: "Maximum depth level (0 = unlimited, default)"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// searchOptions holds the options for the search command.
type searchOptions struct {
	recursive bool
	maxLevel  int // 0 means unlimited
}

func searchHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showSearchHelp(execCtx)
		return 0, nil
	}

	// Parse options
	opts := searchOptions{
		recursive: cmd.HasFlag("-r", "--recursive"),
		maxLevel:  0, // default: unlimited
	}

	// Parse --level option
	if levelValue := cmd.GetOption("-l", "--level"); levelValue != "" {
		level, err := strconv.Atoi(levelValue)
		if err != nil || level < 0 {
			execCtx.WriteErrorln("search: invalid level value: %s", levelValue)
			return 1, nil
		}
		opts.maxLevel = level
	}

	// Need at least directory and one pattern
	if len(cmd.Args) < 2 {
		execCtx.WriteErrorln("search: missing arguments")
		execCtx.WriteErrorln("Usage: search <directory> <expression> [options]")
		return 1, nil
	}

	// First argument is the directory
	searchDir := cmd.Args[0]

	// Remaining arguments form the search expression
	exprArgs := cmd.Args[1:]

	// Parse the search expression
	expr, err := ParseSearchExpression(exprArgs)
	if err != nil {
		execCtx.WriteErrorln("search: invalid expression: %v", err)
		return 1, nil
	}

	// Verify directory exists
	info, err := os.Stat(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			execCtx.WriteErrorln("search: %s: no such directory", searchDir)
		} else {
			execCtx.WriteErrorln("search: %s: %v", searchDir, err)
		}
		return 1, nil
	}

	if !info.IsDir() {
		execCtx.WriteErrorln("search: %s: not a directory", searchDir)
		return 1, nil
	}

	// Perform the search
	found := false
	err = searchDirectoryWithExpr(searchDir, expr, opts, execCtx, 0, &found)
	if err != nil {
		execCtx.WriteErrorln("search: %v", err)
		return 1, nil
	}

	return 0, nil
}

// searchDirectoryWithExpr searches for files matching the expression in the given directory.
func searchDirectoryWithExpr(dir string, expr SearchExpr, opts searchOptions, execCtx *Context, currentLevel int, found *bool) error {
	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Evaluate the expression against the filename
		if expr.Evaluate(name) {
			*found = true
			// Colorize output based on type
			displayPath := fullPath
			if execCtx.Colors != nil {
				info, infoErr := entry.Info()
				if infoErr == nil {
					displayPath = colorizeSearchResult(fullPath, info.Mode(), execCtx)
				}
			}
			fmt.Fprintln(execCtx.Stdout, displayPath)
		}

		// Recurse into subdirectories if recursive mode is enabled
		if opts.recursive && entry.IsDir() {
			// Check depth limit
			if opts.maxLevel > 0 && currentLevel >= opts.maxLevel {
				continue
			}
			// Recurse
			if err := searchDirectoryWithExpr(fullPath, expr, opts, execCtx, currentLevel+1, found); err != nil {
				// Report error but continue with other directories
				execCtx.WriteErrorln("search: %s: %v", fullPath, err)
			}
		}
	}

	return nil
}

// colorizeSearchResult applies color to a search result based on its mode.
func colorizeSearchResult(path string, mode os.FileMode, execCtx *Context) string {
	if execCtx.Colors == nil {
		return path
	}

	// Get directory and base name
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	// Colorize just the name portion
	var coloredName string
	if mode.IsDir() {
		coloredName = execCtx.Colors.Directory(name)
	} else if mode&0111 != 0 {
		coloredName = execCtx.Colors.Executable(name)
	} else if mode&os.ModeSymlink != 0 {
		coloredName = execCtx.Colors.Symlink(name)
	} else {
		coloredName = execCtx.Colors.File(name)
	}

	// Return full path with colored name
	if dir == "." {
		return coloredName
	}
	return filepath.Join(dir, coloredName)
}

func showSearchHelp(execCtx *Context) {
	help := `search - Search for files and directories

Usage: search <directory> <expression> [options]

Arguments:
  directory     The directory to search in
  expression    Pattern(s) with optional logical operators

Options:
  -r, --recursive      Search recursively in subdirectories
  -l, --level=<n>      Maximum depth level (0 = unlimited, default)
      --help           Show this help message

Pattern syntax:
  *             Matches any sequence of characters
  ?             Matches any single character
  [abc]         Matches any character in the brackets
  [a-z]         Matches any character in the range

Logical operators (case-insensitive):
  AND, &&       Both patterns must match
  OR, ||        Either pattern must match (default when multiple patterns)
  XOR, ^        Exactly one pattern must match
  NOT, !        Negate the following pattern
  ( )           Group expressions

Operator precedence (highest to lowest):
  1. NOT
  2. AND
  3. XOR
  4. OR

Simple examples (backward compatible):
  search . "*.go"                    Find Go files in current directory
  search . "*.go" "*.md" -r          Find Go OR Markdown files recursively

Logical expression examples:
  search . "*.go" AND NOT "*_test.go" -r
      Find Go files excluding test files

  search . "*.js" OR "*.ts" -r
      Find JavaScript OR TypeScript files

  search . "test_*" AND "*.py" -r
      Find Python test files (name starts with test_ AND ends with .py)

  search . "*.go" XOR "*.mod" -r
      Find files that are .go OR .mod but not both (XOR)

  search . NOT "*.log" -r
      Find all files except .log files

  search . "(" "*.go" OR "*.md" ")" AND NOT "*_test*" -r
      Find Go or Markdown files, excluding test files

  search . "*.txt" AND "(" "test*" OR "spec*" ")" -r
      Find .txt files starting with test or spec

Note: When using parentheses, they must be separate arguments or quoted.
      Shell may require escaping: \( \) or '(' ')'
`
	execCtx.Stdout.Write([]byte(help))
}
