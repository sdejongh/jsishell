//go:build windows

package executor

// isWindowsDriveLetter checks if the command is a Windows drive letter (e.g., "c:", "D:").
// Returns true if it's a drive letter that should be treated as "cd <drive>:".
func isWindowsDriveLetter(name string) bool {
	if len(name) != 2 {
		return false
	}
	letter := name[0]
	if name[1] != ':' {
		return false
	}
	// Check if it's a valid drive letter (a-z or A-Z)
	return (letter >= 'a' && letter <= 'z') || (letter >= 'A' && letter <= 'Z')
}
