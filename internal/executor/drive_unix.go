//go:build !windows

package executor

// isWindowsDriveLetter always returns false on non-Windows systems.
func isWindowsDriveLetter(name string) bool {
	return false
}
