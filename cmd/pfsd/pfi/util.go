package pfi

import (
	"syscall"

	"github.com/hanwen/go-fuse/fuse"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
	"github.com/cpssd-students/paranoid/pkg/logger"
)

var (
	// SendOverNetwork determines whether the message should be send over network
	// or locally
	SendOverNetwork bool

	// Log used for pfi
	Log *logger.ParanoidLogger
)

// GetFuseReturnCode from the internal return code
func GetFuseReturnCode(retcode returncodes.Code) fuse.Status {
	switch retcode {
	case returncodes.ENOENT:
		return fuse.ENOENT
	case returncodes.EACCES:
		return fuse.EACCES
	case returncodes.EEXIST:
		return fuse.Status(syscall.EEXIST)
	case returncodes.ENOTEMPTY:
		return fuse.Status(syscall.ENOTEMPTY)
	case returncodes.ENOTDIR:
		return fuse.ENOTDIR
	case returncodes.EISDIR:
		return fuse.Status(syscall.EISDIR)
	case returncodes.EIO:
		return fuse.EIO
	case returncodes.EBUSY:
		return fuse.EBUSY
	default:
		return fuse.OK
	}
}
