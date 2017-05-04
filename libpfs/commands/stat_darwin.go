package commands

import (
	"syscall"
	"time"
)

// lastAccess gets the atime for darwin devices
func lastAccess(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
}

// createTime gets the ctime for darwin devices
func createTime(stat *syscall.Stat_t) time.Time {
	return time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
}
