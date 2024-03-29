package libpfs

import (
	"syscall"
	"time"
)

// lastAccess gets the atime for darwin devices
func lastAccess(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}

// createTime gets the ctime for darwin devices
func createTime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
}
