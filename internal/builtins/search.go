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
			{Long: "--absolute", Short: "-a", Description: "Display absolute paths"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// searchOptions holds the options for the search command.
type searchOptions struct {
	recursive bool
	maxLevel  int  // 0 means unlimited
	absolute  bool // display absolute paths
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
		absolute:  cmd.HasFlag("-a", "--absolute"),
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

	// Normalize Windows drive letters: "d:" -> "d:\" for proper path joining
	// On Windows, "d:" means current directory on drive D, while "d:\" means root
	if len(searchDir) == 2 && searchDir[1] == ':' {
		searchDir = searchDir + string(filepath.Separator)
	}

	// Build expression arguments with quoting information
	var exprArgs []ExprArg
	if len(cmd.ArgsWithInfo) >= 2 {
		// Use ArgsWithInfo if available (has quoting information)
		exprArgs = make([]ExprArg, len(cmd.ArgsWithInfo)-1)
		for i, arg := range cmd.ArgsWithInfo[1:] {
			exprArgs[i] = ExprArg{Value: arg.Value, Quoted: arg.Quoted}
		}
	} else {
		// Fallback to Args without quoting info (for backward compatibility with tests)
		exprArgs = make([]ExprArg, len(cmd.Args)-1)
		for i, arg := range cmd.Args[1:] {
			exprArgs[i] = ExprArg{Value: arg, Quoted: false}
		}
	}

	// Parse the search expression with quoting information
	expr, err := ParseSearchExpressionWithQuoting(exprArgs)
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
		// Silently skip permission errors (common on Windows for system folders)
		if os.IsPermission(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Get file info for type predicates
		info, infoErr := entry.Info()
		if infoErr != nil {
			execCtx.WriteErrorln("search: %s: %v", fullPath, infoErr)
			continue
		}

		// Build FileInfo struct for expression evaluation
		fileInfo := &FileInfo{
			Name:   name,
			Mode:   info.Mode(),
			IsDir:  entry.IsDir(),
			IsLink: info.Mode()&os.ModeSymlink != 0,
		}

		// If it's a symlink, try to get the target
		if fileInfo.IsLink {
			if target, err := os.Readlink(fullPath); err == nil {
				fileInfo.LinkTarget = target
			}
		}

		// Evaluate the expression against the file info
		if expr.Evaluate(fileInfo) {
			*found = true
			// Determine display path
			displayPath := fullPath
			if opts.absolute {
				if absPath, err := filepath.Abs(fullPath); err == nil {
					displayPath = absPath
				}
			}
			// Colorize output based on type
			if execCtx.Colors != nil {
				displayPath = colorizeSearchResult(displayPath, info.Mode(), execCtx)
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
				// Silently skip permission errors, report other errors
				if !os.IsPermission(err) {
					execCtx.WriteErrorln("search: %s: %v", fullPath, err)
				}
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
  expression    Pattern(s) with optional logical operators and type predicates

Options:
  -r, --recursive      Search recursively in subdirectories
  -l, --level=<n>      Maximum depth level (0 = unlimited, default)
  -a, --absolute       Display absolute paths
      --help           Show this help message

Pattern syntax:
  *             Matches any sequence of characters
  ?             Matches any single character
  [abc]         Matches any character in the brackets
  [a-z]         Matches any character in the range

Type predicates (case-insensitive):
  isFile        Match regular files only
  isDir         Match directories only
  isLink        Match symbolic links
  isSymlink     Match symbolic links (alias for isLink)
  isHardlink    Match regular files (hard links cannot be detected)
  isExec        Match executable files

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

Quoting:
  Quoted arguments ("..." or '...') are always treated as patterns,
  never as operators or predicates. Use this to search for files
  named AND, OR, isFile, etc.

Simple examples (backward compatible):
  search . "*.go"                    Find Go files in current directory
  search . "*.go" "*.md" -r          Find Go OR Markdown files recursively

Type predicate examples:
  search . "*.go" AND isFile -r
      Find Go files only (exclude directories named *.go)

  search . isDir -r
      Find all directories

  search . isExec -r
      Find all executable files

  search . "*config*" AND isFile -r
      Find files with "config" in the name

  search . isLink -r
      Find all symbolic links

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

  search . "(" "*.go" OR "*.md" ")" AND NOT "*_test*" AND isFile -r
      Find Go or Markdown files, excluding test files

Quoting examples (for files named like operators/predicates):
  search . "AND"                     Find a file named 'AND'
  search . "isFile"                  Find a file named 'isFile'
  search . "OR" "NOT"                Find files named 'OR' or 'NOT'

Note: When using parentheses, they must be separate arguments or quoted.
      Shell may require escaping: \( \) or '(' ')'
`
	execCtx.Stdout.Write([]byte(help))
}
