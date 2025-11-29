//go:build windows

package completion

import (
	"os"
	"strings"
)

// windowsExecutableExts contains the extensions that are executable on Windows.
var windowsExecutableExts = map[string]bool{
	".exe": true,
	".cmd": true,
	".bat": true,
	".com": true,
	".ps1": true,
}

// isExecutable checks if a file is executable on Windows.
// On Windows, executability is determined by file extension, not mode bits.
func isExecutable(info os.FileInfo) bool {
	name := strings.ToLower(info.Name())
	for ext := range windowsExecutableExts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// executableName returns the display name for an executable.
// On Windows, we keep the full name including extension.
func executableName(name string) string {
	return name
}
