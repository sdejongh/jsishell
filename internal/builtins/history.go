package builtins

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sdejongh/jsishell/internal/parser"
)

// HistoryEntry represents a single command in the history (for the builtin).
type HistoryEntry struct {
	Command   string
	Timestamp time.Time
}

// HistoryProvider provides access to command history.
// This interface is implemented by history.History.
type HistoryProvider interface {
	Len() int
	All() []HistoryEntry
	Clear()
}

// historyAdapter adapts the actual history.History to HistoryProvider.
// This is set by the shell at initialization.
var historyProviderFunc func() HistoryProvider

// SetHistoryProvider sets the function that returns the history provider.
func SetHistoryProvider(fn func() HistoryProvider) {
	historyProviderFunc = fn
}

// HistoryDefinition returns the history command definition.
func HistoryDefinition() Definition {
	return Definition{
		Name:        "history",
		Description: "Display or manage command history",
		Usage:       "history [count] [-c|--clear]",
		Handler:     historyHandler,
		Options: []OptionDef{
			{Long: "--clear", Short: "-c", Description: "Clear the history"},
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func historyHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showHistoryHelp(execCtx)
		return 0, nil
	}

	// Get history provider
	if historyProviderFunc == nil {
		execCtx.WriteErrorln("history: history not available")
		return 1, nil
	}

	provider := historyProviderFunc()
	if provider == nil {
		execCtx.WriteErrorln("history: history not available")
		return 1, nil
	}

	// Check for --clear or -c
	if cmd.HasFlag("--clear") || cmd.HasFlag("-c") {
		provider.Clear()
		fmt.Fprintln(execCtx.Stdout, "History cleared")
		return 0, nil
	}

	// Get all entries
	entries := provider.All()
	if len(entries) == 0 {
		fmt.Fprintln(execCtx.Stdout, "No history")
		return 0, nil
	}

	// Check if count is specified
	count := len(entries)
	if len(cmd.Args) > 0 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil || n <= 0 {
			execCtx.WriteErrorln("history: invalid count: %s", cmd.Args[0])
			return 1, nil
		}
		if n < count {
			count = n
		}
	}

	// Display entries (last 'count' entries)
	startIdx := len(entries) - count
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(entries); i++ {
		entry := entries[i]
		// Format: number  command
		fmt.Fprintf(execCtx.Stdout, "%5d  %s\n", i+1, entry.Command)
	}

	return 0, nil
}

func showHistoryHelp(execCtx *Context) {
	help := `history - Display or manage command history

Usage: history [count] [options]

Arguments:
  count         Number of recent entries to display (default: all)

Options:
  -c, --clear   Clear the history
  --help        Show this help message

Examples:
  history       Display all history entries
  history 10    Display the last 10 entries
  history -c    Clear the history
`
	execCtx.Stdout.Write([]byte(help))
}
