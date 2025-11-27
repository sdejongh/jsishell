package builtins

import (
	"context"
	"fmt"
	"sort"

	"github.com/sdejongh/jsishell/internal/parser"
)

// EnvDefinition returns the env command definition.
func EnvDefinition() Definition {
	return Definition{
		Name:        "env",
		Description: "Display or set environment variables",
		Usage:       "env [name[=value]...]",
		Handler:     envHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func envHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showEnvHelp(execCtx)
		return 0, nil
	}

	// If no arguments, display all environment variables
	if len(cmd.Args) == 0 {
		return displayEnv(execCtx)
	}

	// Process each argument (could be NAME=VALUE or just NAME)
	for _, arg := range cmd.Args {
		processEnvArg(arg, execCtx)
	}

	return 0, nil
}

// processEnvArg processes a single env argument (NAME=VALUE or NAME).
func processEnvArg(arg string, execCtx *Context) {
	// Check if it's an assignment
	for i := 0; i < len(arg); i++ {
		if arg[i] == '=' {
			// It's an assignment: NAME=VALUE
			name := arg[:i]
			value := arg[i+1:]
			if name != "" {
				execCtx.Env.Set(name, value)
			}
			return
		}
	}

	// It's just a name - display that variable
	value := execCtx.Env.Get(arg)
	if value != "" {
		fmt.Fprintf(execCtx.Stdout, "%s=%s\n", arg, value)
	}
}

func displayEnv(execCtx *Context) (int, error) {
	all := execCtx.Env.All()

	// Sort keys for consistent output
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(execCtx.Stdout, "%s=%s\n", key, all[key])
	}

	return 0, nil
}

func showEnvHelp(execCtx *Context) {
	help := `env - Display or set environment variables

Usage: env [name[=value]...]

Description:
  Without arguments, displays all environment variables.
  With NAME arguments, displays those specific variables.
  With NAME=VALUE arguments, sets those variables.

Examples:
  env              Display all environment variables
  env HOME         Display the HOME variable
  env FOO=bar      Set FOO to "bar"
  env FOO=bar BAZ  Set FOO and display BAZ
`
	execCtx.Stdout.Write([]byte(help))
}
