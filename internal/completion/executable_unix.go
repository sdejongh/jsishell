//go:build !windows

package completion

import "os"

// isExecutable checks if a file is executable on Unix systems.
// On Unix, executability is determined by the mode bits.
func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0111 != 0
}

// executableName returns the display name for an executable.
// On Unix, we just return the name as-is.
func executableName(name string) string {
	return name
}
