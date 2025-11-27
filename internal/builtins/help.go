package builtins

import (
	"context"
	"fmt"
	"sort"

	"github.com/sdejongh/jsishell/internal/parser"
)

// helpRegistry holds a reference to the registry for the help command.
// This is set by RegisterAll.
var helpRegistry *Registry

// HelpDefinition returns the help command definition.
func HelpDefinition() Definition {
	return Definition{
		Name:        "help",
		Description: "Show help information",
		Usage:       "help [command]",
		Handler:     helpHandler,
		Options:     []OptionDef{},
	}
}

func helpHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	if helpRegistry == nil {
		execCtx.Stderr.Write([]byte("help: registry not initialized\n"))
		return 1, nil
	}

	// If a command name is provided, show specific help
	if len(cmd.Args) > 0 {
		return showCommandHelp(cmd.Args[0], execCtx)
	}

	// Otherwise, show general help
	return showGeneralHelp(execCtx)
}

func showGeneralHelp(execCtx *Context) (int, error) {
	fmt.Fprintln(execCtx.Stdout, "JSIShell - Interactive Shell")
	fmt.Fprintln(execCtx.Stdout, "")
	fmt.Fprintln(execCtx.Stdout, "Available commands:")
	fmt.Fprintln(execCtx.Stdout, "")

	// Get all commands sorted by name
	all := helpRegistry.All()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	// Find max name length for formatting
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Print each command
	for _, name := range names {
		def := all[name]
		fmt.Fprintf(execCtx.Stdout, "  %-*s  %s\n", maxLen, name, def.Description)
	}

	fmt.Fprintln(execCtx.Stdout, "")
	fmt.Fprintln(execCtx.Stdout, "Type 'help <command>' for detailed information about a command.")
	fmt.Fprintln(execCtx.Stdout, "")

	return 0, nil
}

func showCommandHelp(name string, execCtx *Context) (int, error) {
	// Try exact match first
	def, ok := helpRegistry.Get(name)
	if !ok {
		// Try prefix match
		matches := helpRegistry.Match(name)
		if len(matches) == 0 {
			execCtx.WriteErrorln("help: no command found: %s", name)
			return 1, nil
		}
		if len(matches) > 1 {
			execCtx.WriteErrorln("help: ambiguous command '%s'. Did you mean:", name)
			for _, m := range matches {
				execCtx.WriteErrorln("  %s", m)
			}
			return 1, nil
		}
		def, _ = helpRegistry.Get(matches[0])
	}

	// Print command help
	fmt.Fprintf(execCtx.Stdout, "%s - %s\n\n", def.Name, def.Description)
	fmt.Fprintf(execCtx.Stdout, "Usage: %s\n", def.Usage)

	if len(def.Options) > 0 {
		fmt.Fprintln(execCtx.Stdout, "")
		fmt.Fprintln(execCtx.Stdout, "Options:")

		for _, opt := range def.Options {
			if opt.Short != "" && opt.Long != "" {
				fmt.Fprintf(execCtx.Stdout, "  %s, %-12s  %s\n", opt.Short, opt.Long, opt.Description)
			} else if opt.Long != "" {
				fmt.Fprintf(execCtx.Stdout, "      %-12s  %s\n", opt.Long, opt.Description)
			} else if opt.Short != "" {
				fmt.Fprintf(execCtx.Stdout, "  %-16s  %s\n", opt.Short, opt.Description)
			}
		}
	}

	fmt.Fprintln(execCtx.Stdout, "")
	return 0, nil
}

// SetHelpRegistry sets the registry for the help command.
func SetHelpRegistry(r *Registry) {
	helpRegistry = r
}
