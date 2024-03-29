package pfi

import (
	"log"
	"os"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

//ParanoidFile is a custom file struct with read and write functions
type ParanoidFile struct {
	Name string
	nodefs.File
}

//newParanoidFile returns a new object of ParanoidFile
func newParanoidFile(name string) nodefs.File {
	return &ParanoidFile{
		Name: name,
		File: nodefs.NewDefaultFile(),
	}
}

//Read reads a file and returns an array of bytes
func (f *ParanoidFile) Read(buf []byte, off int64) (fuse.ReadResult, fuse.Status) {
	log.Printf("Read called on file: %s", f.Name)
	code, data, err := libpfs.ReadCommand(globals.ParanoidDir, f.Name, off, int64(len(buf)))
	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running read command: %v", err)
	}

	if err != nil {
		log.Printf("Error running read command: %v", err)
	}

	copy(buf, data)
	if code != returncodes.OK {
		return nil, GetFuseReturnCode(code)
	}
	return fuse.ReadResultData(data), fuse.OK
}

//Write writes to a file
func (f *ParanoidFile) Write(content []byte, off int64) (uint32, fuse.Status) {
	log.Printf("Write called on file %s", f.Name)
	var (
		code         returncodes.Code
		err          error
		bytesWritten int
	)
	if SendOverNetwork {
		code, bytesWritten, err = globals.RaftNetworkServer.RequestWriteCommand(
			f.Name, off, int64(len(content)), content,
		)
	} else {
		code, bytesWritten, err = libpfs.WriteCommand(
			globals.ParanoidDir, f.Name, off, int64(len(content)), content,
		)
	}

	if code == returncodes.EUNEXPECTED {
		log.Printf("Error running write command: %v", err)
	}

	if err != nil {
		log.Printf("Error running write command: %v", err)
	}

	if code != returncodes.OK {
		return 0, GetFuseReturnCode(code)
	}

	return uint32(bytesWritten), fuse.OK
}

//Truncate is called when a file is to be reduced in length to size.
func (f *ParanoidFile) Truncate(size uint64) fuse.Status {
	log.Printf("Truncate called on file %s", f.Name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestTruncateCommand(f.Name, int64(size))
	} else {
		code, err = libpfs.TruncateCommand(globals.ParanoidDir, f.Name, int64(size))
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running truncate command: %v", err)
	}

	if err != nil {
		log.Printf("Error running truncate command: %v", err)
	}

	return GetFuseReturnCode(code)
}

//Utimens updates the access and mofication time of the file.
func (f *ParanoidFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {
	log.Printf("Utimens called on file %s", f.Name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestUtimesCommand(f.Name, atime, mtime)
	} else {
		code, err = libpfs.UtimesCommand(globals.ParanoidDir, f.Name, atime, mtime)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running utimes command: %v", err)
	}

	if err != nil {
		log.Printf("Error running utimes command: %v", err)
	}
	return GetFuseReturnCode(code)
}

//Chmod changes the permission flags of the file
func (f *ParanoidFile) Chmod(perms uint32) fuse.Status {
	log.Printf("Chmod called on file %s", f.Name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestChmodCommand(f.Name, perms)
	} else {
		code, err = libpfs.ChmodCommand(globals.ParanoidDir, f.Name, os.FileMode(perms))
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running chmod command: %v", err)
	}

	if err != nil {
		log.Printf("Error running chmod command: %v", err)
	}
	return GetFuseReturnCode(code)
}
