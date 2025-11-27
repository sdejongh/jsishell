// Package errors defines sentinel errors for JSIShell.
// These errors represent common error conditions throughout the shell.
package errors

import "errors"

// Sentinel errors for command execution.
var (
	// ErrCommandNotFound indicates the command does not exist.
	ErrCommandNotFound = errors.New("command not found")

	// ErrAmbiguousCommand indicates multiple commands match the prefix.
	ErrAmbiguousCommand = errors.New("ambiguous command")

	// ErrInvalidSyntax indicates the input could not be parsed.
	ErrInvalidSyntax = errors.New("invalid syntax")
)

// Sentinel errors for file operations.
var (
	// ErrPermissionDenied indicates insufficient permissions for the operation.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrFileNotFound indicates the specified file does not exist.
	ErrFileNotFound = errors.New("file not found")

	// ErrDirectoryNotFound indicates the specified directory does not exist.
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrNotADirectory indicates an expected directory is not a directory.
	ErrNotADirectory = errors.New("not a directory")

	// ErrIsADirectory indicates an unexpected directory where a file was expected.
	ErrIsADirectory = errors.New("is a directory")

	// ErrFileExists indicates the target file already exists.
	ErrFileExists = errors.New("file exists")
)

// Sentinel errors for configuration.
var (
	// ErrInvalidConfig indicates the configuration file is malformed.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrConfigNotFound indicates the configuration file does not exist.
	ErrConfigNotFound = errors.New("configuration file not found")
)

// Sentinel errors for terminal operations.
var (
	// ErrNotATerminal indicates stdin/stdout is not connected to a terminal.
	ErrNotATerminal = errors.New("not a terminal")

	// ErrInterrupted indicates the operation was interrupted (e.g., Ctrl+C).
	ErrInterrupted = errors.New("interrupted")

	// ErrEOF indicates end of input was reached.
	ErrEOF = errors.New("end of input")
)
