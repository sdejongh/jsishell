//go:build !windows

package builtins

import (
	"io/fs"
	"os/user"
	"strconv"
	"syscall"
)

// getOwnerGroup returns the owner and group names for a file.
// Falls back to numeric IDs if names cannot be resolved.
func getOwnerGroup(info fs.FileInfo) (owner, group string) {
	owner = "?"
	group = "?"

	// Get the underlying syscall.Stat_t
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}

	// Get owner name
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	if u, err := user.LookupId(uid); err == nil {
		owner = u.Username
	} else {
		owner = uid
	}

	// Get group name
	gid := strconv.FormatUint(uint64(stat.Gid), 10)
	if g, err := user.LookupGroupId(gid); err == nil {
		group = g.Name
	} else {
		group = gid
	}

	return
}
