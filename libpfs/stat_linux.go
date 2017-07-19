package libpfs

import (
	"syscall"
	"time"
)

// lastAccess gets the atime for linux devices
func lastAccess(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}

// createTime gets the ctime for linux devices
func createTime(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
}
