//go:build windows

package builtins

import (
	"io/fs"
)

// getOwnerGroup returns placeholder values on Windows.
// Windows uses a different security model (SIDs) that doesn't map directly
// to Unix-style uid/gid. For simplicity, we return placeholder values.
func getOwnerGroup(info fs.FileInfo) (owner, group string) {
	// Windows doesn't have Unix-style owner/group
	// Return placeholder values for ls -l output
	return "-", "-"
}
