// Package main is the entry point for JSIShell.
package main

import (
	"fmt"
	"os"

	"github.com/sdejongh/jsishell/internal/shell"
)

const version = "0.1.0"

func main() {
	// Check for --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("JSIShell version %s\n", version)
		os.Exit(0)
	}

	// Check for --help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		os.Exit(0)
	}

	// Create and run the shell
	// Configuration is loaded automatically from ~/.config/jsishell/config.yaml
	sh := shell.New()

	if err := sh.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(sh.ExitCode())
}

func printUsage() {
	fmt.Println("JSIShell - Interactive Shell")
	fmt.Println("")
	fmt.Println("Usage: jsishell [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("")
	fmt.Println("Once in the shell, type 'help' to see available commands.")
}
