package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		// Command execution errors
		{"ErrCommandNotFound", ErrCommandNotFound, "command not found"},
		{"ErrAmbiguousCommand", ErrAmbiguousCommand, "ambiguous command"},
		{"ErrInvalidSyntax", ErrInvalidSyntax, "invalid syntax"},

		// File operation errors
		{"ErrPermissionDenied", ErrPermissionDenied, "permission denied"},
		{"ErrFileNotFound", ErrFileNotFound, "file not found"},
		{"ErrDirectoryNotFound", ErrDirectoryNotFound, "directory not found"},
		{"ErrNotADirectory", ErrNotADirectory, "not a directory"},
		{"ErrIsADirectory", ErrIsADirectory, "is a directory"},
		{"ErrFileExists", ErrFileExists, "file exists"},

		// Configuration errors
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
		{"ErrConfigNotFound", ErrConfigNotFound, "configuration file not found"},

		// Terminal errors
		{"ErrNotATerminal", ErrNotATerminal, "not a terminal"},
		{"ErrInterrupted", ErrInterrupted, "interrupted"},
		{"ErrEOF", ErrEOF, "end of input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrCommandNotFound", ErrCommandNotFound},
		{"ErrFileNotFound", ErrFileNotFound},
		{"ErrPermissionDenied", ErrPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap the error with context
			wrapped := fmt.Errorf("operation failed: %w", tt.sentinel)

			// errors.Is should detect the wrapped sentinel
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("errors.Is(wrapped, %s) = false, want true", tt.name)
			}

			// The original sentinel should not match a different error
			if errors.Is(wrapped, ErrInvalidSyntax) && tt.sentinel != ErrInvalidSyntax {
				t.Errorf("errors.Is(wrapped, ErrInvalidSyntax) = true, want false for %s", tt.name)
			}
		})
	}
}

func TestErrorUniqueness(t *testing.T) {
	allErrors := []error{
		ErrCommandNotFound,
		ErrAmbiguousCommand,
		ErrInvalidSyntax,
		ErrPermissionDenied,
		ErrFileNotFound,
		ErrDirectoryNotFound,
		ErrNotADirectory,
		ErrIsADirectory,
		ErrFileExists,
		ErrInvalidConfig,
		ErrConfigNotFound,
		ErrNotATerminal,
		ErrInterrupted,
		ErrEOF,
	}

	// Ensure all errors are unique (different pointers)
	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i != j && err1 == err2 {
				t.Errorf("errors at index %d and %d are the same: %v", i, j, err1)
			}
		}
	}
}
