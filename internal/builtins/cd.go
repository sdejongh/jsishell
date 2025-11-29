package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdejongh/jsishell/internal/parser"
)

// CdDefinition returns the cd command definition.
func CdDefinition() Definition {
	return Definition{
		Name:        "cd",
		Description: "Change the current directory",
		Usage:       "cd [directory]",
		Handler:     cdHandler,
		Options: []OptionDef{
			{Long: "--help", Description: "Show help message"},
		},
	}
}

func cdHandler(ctx context.Context, cmd *parser.Command, execCtx *Context) (int, error) {
	// Check for --help
	if cmd.HasFlag("--help") {
		showCdHelp(execCtx)
		return 0, nil
	}

	var targetDir string

	if len(cmd.Args) == 0 {
		// No argument - go to HOME
		home := execCtx.Env.Get("HOME")
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		if home == "" {
			execCtx.WriteErrorln("cd: HOME not set")
			return 1, nil
		}
		targetDir = home
	} else {
		targetDir = cmd.Args[0]
	}

	// Handle ~ expansion
	if len(targetDir) > 0 && targetDir[0] == '~' {
		home := execCtx.Env.Get("HOME")
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		if len(targetDir) == 1 {
			targetDir = home
		} else {
			targetDir = filepath.Join(home, targetDir[1:])
		}
	}

	// Handle - (previous directory)
	if targetDir == "-" {
		oldPwd := execCtx.Env.Get("OLDPWD")
		if oldPwd == "" {
			execCtx.WriteErrorln("cd: OLDPWD not set")
			return 1, nil
		}
		targetDir = oldPwd
		fmt.Fprintln(execCtx.Stdout, targetDir)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		execCtx.WriteErrorln("cd: %s: %v", targetDir, err)
		return 1, nil
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			execCtx.WriteErrorln("cd: %s: No such file or directory", targetDir)
		} else {
			execCtx.WriteErrorln("cd: %s: %v", targetDir, err)
		}
		return 1, nil
	}

	if !info.IsDir() {
		execCtx.WriteErrorln("cd: %s: Not a directory", targetDir)
		return 1, nil
	}

	// Save current directory as OLDPWD
	currentPwd := execCtx.Env.Get("PWD")
	if currentPwd == "" {
		currentPwd, _ = os.Getwd()
	}

	// Change directory
	if err := os.Chdir(absPath); err != nil {
		execCtx.WriteErrorln("cd: %s: %v", targetDir, err)
		return 1, nil
	}

	// Update environment variables
	execCtx.Env.Set("OLDPWD", currentPwd)
	execCtx.Env.Set("PWD", absPath)
	execCtx.WorkDir = absPath

	return 0, nil
}

func showCdHelp(execCtx *Context) {
	help := `cd - Change the current directory

Usage: cd [directory]

Arguments:
  directory   Target directory (default: $HOME)
              Use ~ for home directory
              Use - for previous directory

Examples:
  cd           Go to home directory
  cd /tmp      Go to /tmp
  cd ~         Go to home directory
  cd ~/docs    Go to docs in home directory
  cd -         Go to previous directory
`
	execCtx.Stdout.Write([]byte(help))
}
