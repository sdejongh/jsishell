package env

import "os"

// PathEnvName returns the environment variable name for the executable search path.
// On Windows, this is "Path" (case-insensitive but conventionally capitalized this way).
// On Unix systems, this is "PATH".
func PathEnvName() string {
	// Windows environment variables are case-insensitive, but Go's os.Getenv
	// on Windows handles this automatically. However, when iterating through
	// environment variables or using a custom Environment map, we need the
	// correct case. Windows typically uses "Path" while Unix uses "PATH".

	// Check if "Path" exists (Windows convention)
	if path := os.Getenv("Path"); path != "" {
		return "Path"
	}
	// Check if "PATH" exists (Unix convention, also works on some Windows setups)
	if path := os.Getenv("PATH"); path != "" {
		return "PATH"
	}
	// Default to PATH (Unix standard)
	return "PATH"
}

// GetPath returns the value of the PATH/Path environment variable.
// It handles the case difference between Windows (Path) and Unix (PATH).
func GetPath() string {
	// Try Windows convention first
	if path := os.Getenv("Path"); path != "" {
		return path
	}
	// Fall back to Unix convention
	return os.Getenv("PATH")
}

// GetPathFrom returns the PATH value from an Environment, handling case differences.
func GetPathFrom(e *Environment) string {
	// Try Unix convention first (more common in our codebase)
	if path := e.Get("PATH"); path != "" {
		return path
	}
	// Try Windows convention
	return e.Get("Path")
}
