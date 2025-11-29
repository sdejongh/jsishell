package builtins

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/sdejongh/jsishell/internal/parser"
)

// LsDefinition returns the ls command definition.
func LsDefinition() Definition {
	return Definition{
		Name:        "ls",
		Description: "List directory contents",
		Usage:       "ls [options] [path...]",
		Handler:     lsHandler,
		Options: []OptionDef{
			{Long: "--all", Short: "-a", Description: "Include hidden files (starting with .)"},
			{Long: "--directory", Short: "-d", Description: "List directories only"},
			{Long: "--long", Short: "-l", Description: "Use long listing format"},
			{Long: "--recursive", Short: "-R", Description: "List subdirectories recursively"},
			{Long: "--verbose", Short: "-v", Description: "Show detailed information about each entry"},
			{Long: "--quiet", Short: "-q", Description: "Only show file names, suppress other output"},
			{Long: "--sort", Short: "-s", HasValue: true, Description: "Sort by: name, size, time, dir (comma-separated, prefix with - to reverse)"},
			{Long: "--exclude", Short: "-e", HasValue: true, Description: "Exclude files matching glob pattern (can be used multiple times)"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

// sortCriterion represents a single sort criterion.
type sortCriterion struct {
	field    string // "name", "size", "time", "dir"
	reversed bool   // true if prefixed with -
}

// lsOptions holds the options for the ls command.
type lsOptions struct {
	showAll         bool
	directoriesOnly bool
	longFormat      bool
	recursive       bool
	verbose         bool
	quiet           bool
	sortBy          []sortCriterion // sort criteria in order of priority
	excludePatterns []string        // glob patterns to exclude
}

func lsHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showLsHelp(execCtx)
		return 0, nil
	}

	opts := lsOptions{
		showAll:         cmd.HasFlag("-a", "--all"),
		directoriesOnly: cmd.HasFlag("-d", "--directory"),
		longFormat:      cmd.HasFlag("-l", "--long"),
		recursive:       cmd.HasFlag("-R", "--recursive"),
		verbose:         cmd.HasFlag("-v", "--verbose"),
		quiet:           cmd.HasFlag("-q", "--quiet"),
		sortBy:          []sortCriterion{{field: "name", reversed: false}}, // default sort
	}

	// Parse --sort option
	if sortValue := cmd.GetOption("-s", "--sort"); sortValue != "" {
		criteria, err := parseSortCriteria(sortValue)
		if err != nil {
			execCtx.WriteErrorln("ls: %v", err)
			return 1, nil
		}
		opts.sortBy = criteria
	}

	// Parse --exclude options (can be used multiple times)
	opts.excludePatterns = cmd.GetOptions("-e", "--exclude")

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

		if err := lsPath(path, opts, execCtx); err != nil {
			if !opts.quiet {
				execCtx.WriteErrorln("ls: %s: %v", path, err)
			}
			exitCode = 1
		}
	}

	return exitCode, nil
}

// parseSortCriteria parses a comma-separated list of sort criteria.
// Each criterion can be prefixed with - to reverse the sort order.
// Valid fields: name, size, time, dir
func parseSortCriteria(value string) ([]sortCriterion, error) {
	parts := strings.Split(value, ",")
	criteria := make([]sortCriterion, 0, len(parts))

	validFields := map[string]bool{
		"name": true,
		"size": true,
		"time": true,
		"dir":  true,
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		criterion := sortCriterion{}

		// Check for reverse prefix (! or - followed by valid field)
		if len(part) > 1 && (part[0] == '-' || part[0] == '!') {
			// Check if rest is a valid field
			rest := part[1:]
			if validFields[rest] {
				criterion.reversed = true
				part = rest
			}
			// Otherwise treat the whole thing as the field name (might be invalid)
		}

		// Validate field name
		if !validFields[part] {
			return nil, fmt.Errorf("invalid sort field: %s (valid: name, size, time, dir)", part)
		}

		criterion.field = part
		criteria = append(criteria, criterion)
	}

	// If no valid criteria, default to name
	if len(criteria) == 0 {
		criteria = append(criteria, sortCriterion{field: "name", reversed: false})
	}

	// Always add name as final tiebreaker if not already present
	hasName := false
	for _, c := range criteria {
		if c.field == "name" {
			hasName = true
			break
		}
	}
	if !hasName {
		criteria = append(criteria, sortCriterion{field: "name", reversed: false})
	}

	return criteria, nil
}

// lsEntryInfo holds cached info for sorting.
type lsEntryInfo struct {
	entry   os.DirEntry
	info    fs.FileInfo
	name    string
	nameLow string
}

// sortEntries sorts directory entries according to the given criteria.
func sortEntries(entries []os.DirEntry, criteria []sortCriterion) []os.DirEntry {
	// Build entry info cache
	infos := make([]lsEntryInfo, len(entries))
	for i, entry := range entries {
		info, _ := entry.Info()
		infos[i] = lsEntryInfo{
			entry:   entry,
			info:    info,
			name:    entry.Name(),
			nameLow: strings.ToLower(entry.Name()),
		}
	}

	// Sort using all criteria
	sort.Slice(infos, func(i, j int) bool {
		for _, criterion := range criteria {
			cmp := compareEntries(&infos[i], &infos[j], criterion.field)
			if cmp != 0 {
				if criterion.reversed {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false // equal
	})

	// Extract sorted entries
	result := make([]os.DirEntry, len(entries))
	for i, info := range infos {
		result[i] = info.entry
	}
	return result
}

// compareEntries compares two entries by the given field.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareEntries(a, b *lsEntryInfo, field string) int {
	switch field {
	case "name":
		if a.nameLow < b.nameLow {
			return -1
		}
		if a.nameLow > b.nameLow {
			return 1
		}
		return 0

	case "size":
		var sizeA, sizeB int64
		if a.info != nil {
			sizeA = a.info.Size()
		}
		if b.info != nil {
			sizeB = b.info.Size()
		}
		if sizeA < sizeB {
			return -1
		}
		if sizeA > sizeB {
			return 1
		}
		return 0

	case "time":
		var timeA, timeB int64
		if a.info != nil {
			timeA = a.info.ModTime().UnixNano()
		}
		if b.info != nil {
			timeB = b.info.ModTime().UnixNano()
		}
		if timeA < timeB {
			return -1
		}
		if timeA > timeB {
			return 1
		}
		return 0

	case "dir":
		// Directories come first (are "less than" files)
		aIsDir := a.entry.IsDir()
		bIsDir := b.entry.IsDir()
		if aIsDir && !bIsDir {
			return -1
		}
		if !aIsDir && bIsDir {
			return 1
		}
		return 0

	default:
		return 0
	}
}

// matchesExcludePattern checks if a name matches any of the exclude patterns.
func matchesExcludePattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func lsPath(path string, opts lsOptions, execCtx *Context) error {
	return lsPathInternal(path, opts, execCtx, false)
}

func lsPathInternal(path string, opts lsOptions, execCtx *Context, isRecursiveCall bool) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no such file or directory")
		}
		return err
	}

	// If it's a file, just show the file (unless -d is set, then skip files)
	if !info.IsDir() {
		if opts.directoriesOnly {
			return nil // Skip files when -d is set
		}
		if opts.quiet {
			fmt.Fprintln(execCtx.Stdout, filepath.Base(path))
		} else if opts.longFormat {
			printLsLongEntry(info, path, opts.verbose, execCtx)
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

	// Filter entries first (before sorting for efficiency)
	var filteredEntries []os.DirEntry
	var subDirs []string // For recursive listing
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless -a
		if !opts.showAll && strings.HasPrefix(name, ".") {
			continue
		}

		// Check exclude patterns
		if matchesExcludePattern(name, opts.excludePatterns) {
			continue
		}

		// Collect subdirectories for recursive listing
		if opts.recursive && entry.IsDir() {
			subDirs = append(subDirs, filepath.Join(path, name))
		}

		// Skip non-directories if -d is set
		if opts.directoriesOnly && !entry.IsDir() {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	// Sort entries according to criteria
	filteredEntries = sortEntries(filteredEntries, opts.sortBy)

	// In recursive mode, print the directory path header
	if opts.recursive {
		if isRecursiveCall {
			fmt.Fprintln(execCtx.Stdout) // Blank line before subdirectory
		}
		fmt.Fprintf(execCtx.Stdout, "%s:\n", path)
	}

	// Display based on mode
	if opts.quiet {
		// Quiet mode: just print names, one per line
		for _, entry := range filteredEntries {
			fmt.Fprintln(execCtx.Stdout, entry.Name())
		}
	} else if opts.longFormat {
		// Long format: one entry per line with details
		for _, entry := range filteredEntries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			printLsLongEntry(info, filepath.Join(path, entry.Name()), opts.verbose, execCtx)
		}
	} else {
		// Default: columnar output like bash
		printColumnar(filteredEntries, path, execCtx)
	}

	// Recursively list subdirectories
	if opts.recursive {
		// Sort subdirs for consistent output
		sort.Strings(subDirs)
		for _, subDir := range subDirs {
			if err := lsPathInternal(subDir, opts, execCtx, true); err != nil {
				// Report error but continue with other directories
				if !opts.quiet {
					execCtx.WriteErrorln("ls: %s: %v", subDir, err)
				}
			}
		}
	}

	return nil
}

// printColumnar displays entries in columns, similar to bash ls.
func printColumnar(entries []os.DirEntry, basePath string, execCtx *Context) {
	if len(entries) == 0 {
		return
	}

	// Build display names and calculate max width
	type displayEntry struct {
		name        string
		coloredName string
	}

	displayEntries := make([]displayEntry, 0, len(entries))
	maxWidth := 0

	for _, entry := range entries {
		name := entry.Name()
		info, err := entry.Info()

		var coloredName string
		if err != nil {
			coloredName = name
		} else {
			coloredName = colorizeEntry(name, info.Mode(), execCtx)
			if entry.IsDir() {
				name += "/"
				coloredName += "/"
			}
		}

		displayEntries = append(displayEntries, displayEntry{
			name:        name,
			coloredName: coloredName,
		})

		if len(name) > maxWidth {
			maxWidth = len(name)
		}
	}

	// Terminal width (default 80 if we can't determine)
	termWidth := 80

	// Calculate number of columns
	// Add 2 for spacing between columns
	colWidth := maxWidth + 2
	if colWidth < 1 {
		colWidth = 1
	}
	numCols := termWidth / colWidth
	if numCols < 1 {
		numCols = 1
	}

	// Print entries in rows
	for i, de := range displayEntries {
		// Calculate padding needed (based on raw name length, not colored)
		padding := colWidth - len(de.name)
		if padding < 0 {
			padding = 0
		}

		// Print the colored name
		fmt.Fprint(execCtx.Stdout, de.coloredName)

		// Add padding or newline
		if (i+1)%numCols == 0 || i == len(displayEntries)-1 {
			// End of row or last entry
			fmt.Fprintln(execCtx.Stdout)
		} else {
			// Add spacing to next column
			fmt.Fprint(execCtx.Stdout, strings.Repeat(" ", padding))
		}
	}
}

func printLsLongEntry(info fs.FileInfo, path string, verbose bool, execCtx *Context) {
	mode := info.Mode()
	size := info.Size()
	modTime := info.ModTime()
	name := filepath.Base(path)

	// Format: drwxr-xr-x  owner group  4096 Jan  2 15:04 name
	modeStr := mode.String()
	timeStr := modTime.Format("Jan _2 15:04")

	// Get owner and group
	owner, group := getOwnerGroup(info)

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
		fmt.Fprintf(execCtx.Stdout, "%s %-8s %-8s %10d %s %-6s %s%s\n", modeStr, owner, group, size, timeStr, fileType, coloredName, suffix)
	} else {
		fmt.Fprintf(execCtx.Stdout, "%s %-8s %-8s %10d %s %s%s\n", modeStr, owner, group, size, timeStr, coloredName, suffix)
	}
}

// getOwnerGroup returns the owner and group names for a file.
// Falls back to numeric IDs if names cannot be resolved.
func getOwnerGroup(info fs.FileInfo) (owner, group string) {
	owner = "?"
	group = "?"

	// Get the underlying syscall.Stat_t
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}

	// Get owner name
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	if u, err := user.LookupId(uid); err == nil {
		owner = u.Username
	} else {
		owner = uid
	}

	// Get group name
	gid := strconv.FormatUint(uint64(stat.Gid), 10)
	if g, err := user.LookupGroupId(gid); err == nil {
		group = g.Name
	} else {
		group = gid
	}

	return
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

func showLsHelp(execCtx *Context) {
	help := `ls - List directory contents

Usage: ls [options] [path...]

Options:
  -a, --all              Include hidden files (starting with .)
  -d, --directory        List directories only
  -l, --long             Use long listing format
  -R, --recursive        List subdirectories recursively
  -v, --verbose          Show detailed information (implies --long, adds file type)
  -q, --quiet            Only show file names, suppress other output
  -s, --sort=<spec>      Sort entries (see below)
  -e, --exclude=<glob>   Exclude files matching glob pattern (can be repeated)
      --help             Show this help message

Sort specification:
  Comma-separated list of fields. Prefix with ! to reverse order.
  Fields: name (default), size, time (modification), dir (directories first)

  Examples:
    --sort=name       Sort by name ascending (default)
    --sort=!name      Sort by name descending
    --sort=size       Sort by size ascending (smallest first)
    --sort=!size      Sort by size descending (largest first)
    --sort=time       Sort by modification time (oldest first)
    --sort=!time      Sort by modification time (newest first)
    --sort=dir        Directories first, then files
    --sort=dir,name   Directories first, then sort by name
    --sort=!size,name Sort by size descending, then by name for equal sizes

Exclude patterns:
  Standard glob patterns: *, ?, [abc], [a-z]
  Can be specified multiple times to exclude several patterns.

Long format shows: permissions, owner, group, size, date, name
Verbose format adds: file type indicator ([dir], [file], [exe], [lnk])

Examples:
  ls                           List current directory
  ls -a                        List including hidden files
  ls -d                        List directories only
  ls -l                        Long format listing
  ls -R                        List recursively
  ls -lR                       Long format, recursive
  ls -l --sort=!time           Long format, newest first
  ls --sort=dir,!size          Directories first, then by size descending
  ls -v                        Verbose listing with file types
  ls -q                        Quiet mode, names only
  ls --exclude=*.log           Exclude log files
  ls --exclude=*.tmp -e=*.bak  Exclude .tmp and .bak files
  ls /home                     List /home directory
  ls dir1 dir2                 List multiple directories
`
	execCtx.Stdout.Write([]byte(help))
}
